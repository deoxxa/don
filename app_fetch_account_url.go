package main

import (
	"net/url"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"fknsrs.biz/p/don/acct"
	"fknsrs.biz/p/don/activitystreams"
	"fknsrs.biz/p/don/webfinger"
)

func fetchAccountURL(key string, v interface{}) ([]byte, error) {
	a, ok := v.(*activitystreams.Author)
	if !ok {
		return nil, errors.Errorf("expected second argument to be *activitystreams.Author; was %T", v)
	}

	var uri, host string
	for _, l := range a.GetLinks("alternate") {
		if l.Type == "text/html" {
			uri = l.Href
		}
	}

	if uri == "" {
		uri = a.URI
	}

	if uri == "" {
		if _, err := url.Parse(a.ID); err == nil {
			uri = a.ID
		}
	}

	if uri != "" {
		u, err := url.Parse(uri)
		if err != nil {
			return nil, errors.Wrap(err, "fetchAccountURL: couldn't parse uri")
		}

		host = u.Hostname()
	}

	if uri == "" && a.Email == "" {
		return nil, errors.Errorf("fetchAccountURL: uri and email were empty")
	}

	if host == "" {
		return nil, errors.Errorf("fetchAccountURL: couldn't determine host for user")
	}

	var ids []string
	for _, e := range guessAccountURL(a) {
		ids = append(ids, e.a)
	}

	ids = append(ids, a.ID, a.URI, uri)

	for _, id := range ids {
		if id == "" {
			continue
		}

		wu := webfinger.MakeURL(host, id, nil)

		r, err := webfinger.Fetch(wu)
		if err != nil {
			continue
		}

		if strings.HasPrefix(r.Subject, "acct:") {
			if _, err := acct.FromString(string(r.Subject)); err == nil {
				return []byte(r.Subject), nil
			}
		}

		for _, s := range r.Aliases {
			if strings.HasPrefix(s, "acct:") {
				if _, err := acct.FromString(string(s)); err == nil {
					return []byte(s), nil
				}
			}
		}
	}

	return nil, errors.Errorf("fetchAccountURL: couldn't determine account uri for user")
}

type accountURLGuess struct {
	a string
	c int
}

type accountURLGuesses []accountURLGuess

func (l accountURLGuesses) Len() int           { return len(l) }
func (l accountURLGuesses) Less(a, b int) bool { return l[a].c < l[b].c }
func (l accountURLGuesses) Swap(a, b int)      { l[a], l[b] = l[b], l[a] }

func guessAccountURL(a *activitystreams.Author) accountURLGuesses {
	accts := make(map[string]int)
	hosts := make(map[string]int)
	names := make(map[string]int)

	if a.PreferredUsername != "" {
		names[a.PreferredUsername] = names[a.PreferredUsername] + 1
	}

	if a.Email != "" {
		if bits := strings.Split(a.Email, "@"); len(bits) == 2 {
			hosts[bits[1]] = hosts[bits[1]] + 1
			names[bits[0]] = names[bits[0]] + 1
			accts[a.Email] = accts[a.Email] + 1
		}
	}

	for _, s := range []string{a.URI, a.GetPermalink()} {
		if s != "" {
			if u, err := url.Parse(s); err == nil {
				if bits := strings.Split(u.Path, "/"); len(bits) > 1 {
					names[bits[len(bits)-1]] = names[bits[len(bits)-1]] + 1
				}

				hosts[u.Hostname()] = hosts[u.Hostname()] + 1
			}
		}
	}

	for host := range hosts {
		for name := range names {
			accts[name+"@"+host] = accts[name+"@"+host] + 1
		}
	}

	var l accountURLGuesses
	for a, c := range accts {
		l = append(l, accountURLGuess{a: a, c: c})
	}

	sort.Sort(l)

	return l
}
