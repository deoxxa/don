package activitystreams

import (
	"github.com/pkg/errors"
	"net/url"

	"fknsrs.biz/p/don/atom"
	"fknsrs.biz/p/don/commonxml"
)

func Fetch(u string) (*Feed, error) {
	var v Feed
	if err := commonxml.Fetch(u, &v); err != nil {
		return nil, errors.Wrap(err, "Fetch")
	}

	return &v, nil
}

type Feed struct {
	atom.Feed

	Author *Author  `xml:"author" json:"author,omitempty"`
	Entry  []Object `xml:"entry" json:"entry,omitempty"`
}

type Author struct {
	atom.Author

	ObjectType        string `xml:"http://activitystrea.ms/spec/1.0/ object-type" json:"objectType,omitempty"`
	PreferredUsername string `xml:"http://portablecontacts.net/spec/1.0 preferredUsername" json:"preferredUsername,omitempty"`
	DisplayName       string `xml:"http://portablecontacts.net/spec/1.0 displayName" json:"displayName,omitempty"`
	Note              string `xml:"http://portablecontacts.net/spec/1.0 note" json:"note,omitempty"`
	Scope             string `xml:"http://mastodon.social/schema/1.0 scope" json:"scope,omitempty"`
}

func (a *Author) Acct() string {
	if a.Email != "" {
		return a.Email
	}

	if a.URI != "" && a.PreferredUsername != "" {
		if u, err := url.Parse(a.URI); err == nil {
			return a.PreferredUsername + "@" + u.Host
		}
	}

	return ""
}

type Object struct {
	atom.Entry

	Author     *Author         `xml:"author" json:"author,omitempty"`
	Verb       string          `xml:"http://activitystrea.ms/spec/1.0/ verb" json:"verb,omitempty"`
	ObjectType string          `xml:"http://activitystrea.ms/spec/1.0/ object-type" json:"objectType,omitempty"`
	Object     *Object         `xml:"http://activitystrea.ms/spec/1.0/ object" json:"object,omitempty"`
	InReplyTo  *commonxml.Link `xml:"http://purl.org/syndication/thread/1.0 in-reply-to" json:"inReplyTo,omitempty"`
}
