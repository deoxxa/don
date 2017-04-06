package main

import (
	"encoding/xml"
	"html/template"
	"net/http"

	"github.com/jaytaylor/html2text"
	"github.com/kennygrant/sanitize"
	"github.com/pkg/errors"
	"gopkg.in/kyokomi/emoji.v1"
)

func AtomFetch(u string) (*AtomFeed, error) {
	res, err := http.Get(u)
	if err != nil {
		return nil, errors.Wrap(err, "AtomFetch")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.Errorf("AtomFetch: invalid status code; expected 200 but got %d", res.StatusCode)
	}

	var v AtomFeed
	if err := xml.NewDecoder(res.Body).Decode(&v); err != nil {
		return nil, errors.Wrap(err, "AtomFetch")
	}

	return &v, nil
}

type AtomFeed struct {
	XMLName xml.Name    `xml:"http://www.w3.org/2005/Atom feed" json:"-"`
	Title   string      `xml:"title" json:"title,omitempty"`
	ID      string      `xml:"id" json:"id,omitempty"`
	Updated string      `xml:"updated" json:"updated,omitempty"`
	Author  *AtomAuthor `xml:"author" json:"author,omitempty"`
	Entry   []AtomEntry `xml:"entry" json:"entry,omitempty"`
	Link    []AtomLink  `xml:"link" json:"link,omitempty"`
}

func (f *AtomFeed) GetLink(rel string) *AtomLink {
	for _, l := range f.Link {
		if l.Rel == rel {
			return &l
		}
	}

	return nil
}

type AtomLink struct {
	Rel      string `xml:"rel,attr,omitempty" json:"rel,omitempty"`
	Type     string `xml:"type,attr,omitempty" json:"type,omitempty"`
	Href     string `xml:"href,attr" json:"href,omitempty"`
	HrefLang string `xml:"hreflang,attr,omitempty" json:"hrefLang,omitempty"`
	Title    string `xml:"title,attr,omitempty" json:"title,omitempty"`
	Length   uint   `xml:"length,attr,omitempty" json:"length,omitempty"`
}

type AtomAuthor struct {
	ID      string `xml:"id,omitempty" json:"id,omitempty"`
	Name    string `xml:"name" json:"name,omitempty"`
	URI     string `xml:"uri,omitempty" json:"uri,omitempty"`
	Email   string `xml:"email,omitempty" json:"email,omitempty"`
	Summary string `xml:"summary,omitempty" json:"summary,omitempty"`

	ObjectType        string `xml:"http://activitystrea.ms/spec/1.0/ object-type" json:"objectType,omitempty"`
	PreferredUsername string `xml:"http://portablecontacts.net/spec/1.0 preferredUsername" json:"preferredUsername,omitempty"`
	DisplayName       string `xml:"http://portablecontacts.net/spec/1.0 displayName" json:"displayName,omitempty"`
	Note              string `xml:"http://portablecontacts.net/spec/1.0 note" json:"note,omitempty"`
	Scope             string `xml:"http://mastodon.social/schema/1.0 scope" json:"scope,omitempty"`

	InnerXML string `xml:",innerxml" json:"-"`
}

func (a *AtomAuthor) String() string {
	if a == nil {
		return "NIL"
	}

	s := a.Name

	if a.Email != "" {
		s += " <" + a.Email + ">"
	}

	return s
}

type AtomEntry struct {
	ID        string       `xml:"id" json:"id,omitempty"`
	Title     string       `xml:"title" json:"title,omitempty"`
	Published string       `xml:"published" json:"published,omitempty"`
	Updated   string       `xml:"updated" json:"updated,omitempty"`
	Author    *AtomAuthor  `xml:"author" json:"author,omitempty"`
	Summary   *AtomContent `xml:"summary" json:"summary,omitempty"`
	Content   *AtomContent `xml:"content" json:"content,omitempty"`
	Link      []AtomLink   `xml:"link" json:"link,omitempty"`

	Verb       string     `xml:"http://activitystrea.ms/spec/1.0/ verb" json:"verb,omitempty"`
	ObjectType string     `xml:"http://activitystrea.ms/spec/1.0/ object-type" json:"objectType,omitempty"`
	Object     *AtomEntry `xml:"http://activitystrea.ms/spec/1.0/ object" json:"object,omitempty"`

	InReplyTo *AtomLink `xml:"http://purl.org/syndication/thread/1.0 in-reply-to" json:"inReplyTo,omitempty"`
}

func (e *AtomEntry) GetLinks(rel string) []AtomLink {
	var a []AtomLink

	for _, l := range e.Link {
		if l.Rel == rel {
			a = append(a, l)
		}
	}

	return a
}

func (e *AtomEntry) GetLink(rel string) *AtomLink {
	if a := e.GetLinks(rel); len(a) > 0 {
		return &a[0]
	}

	return nil
}

type AtomContent struct {
	Type string `xml:"type,attr" json:"type,omitempty"`
	Body string `xml:",chardata" json:"body,omitempty"`
}

func (c *AtomContent) Text() string {
	if c == nil {
		return "NIL"
	}

	switch c.Type {
	case "html":
		if t, err := html2text.FromString(c.Body); err == nil {
			return t
		}
	}

	return c.Body
}

func (c *AtomContent) HTML() template.HTML {
	if c == nil {
		return ""
	}

	if c.Type != "html" {
		return template.HTML(sanitize.HTML(emoji.Sprint(c.Body)))
	}

	t, err := sanitize.HTMLAllowing(emoji.Sprint(c.Body))
	if err != nil {
		return template.HTML(sanitize.HTML(emoji.Sprint(c.Body)))
	}

	return template.HTML(t)
}
