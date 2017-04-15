package atom

import (
	"encoding/xml"

	"github.com/jaytaylor/html2text"
	"github.com/kennygrant/sanitize"
	"github.com/pkg/errors"
	"gopkg.in/kyokomi/emoji.v1"

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
	commonxml.HasLinks

	XMLName xml.Name `xml:"http://www.w3.org/2005/Atom feed" json:"-"`
	Title   string   `xml:"title" json:"title,omitempty"`
	ID      string   `xml:"id" json:"id,omitempty"`
	Updated string   `xml:"updated" json:"updated,omitempty"`
	Author  *Author  `xml:"author" json:"author,omitempty"`
	Entry   []Entry  `xml:"entry" json:"entry,omitempty"`
}

type Author struct {
	ID      string `xml:"id,omitempty" json:"id,omitempty"`
	Name    string `xml:"name" json:"name,omitempty"`
	URI     string `xml:"uri,omitempty" json:"uri,omitempty"`
	Email   string `xml:"email,omitempty" json:"email,omitempty"`
	Summary string `xml:"summary,omitempty" json:"summary,omitempty"`

	InnerXML string `xml:",innerxml" json:"-"`
}

func (a *Author) String() string {
	if a == nil {
		return "NIL"
	}

	s := a.Name

	if a.Email != "" {
		s += " <" + a.Email + ">"
	}

	return s
}

type Entry struct {
	commonxml.HasLinks

	ID        string   `xml:"id" json:"id,omitempty"`
	Title     string   `xml:"title" json:"title,omitempty"`
	Published string   `xml:"published" json:"published,omitempty"`
	Updated   string   `xml:"updated" json:"updated,omitempty"`
	Author    *Author  `xml:"author" json:"author,omitempty"`
	Summary   *Content `xml:"summary" json:"summary,omitempty"`
	Content   *Content `xml:"content" json:"content,omitempty"`

	InnerXML string `xml:",innerxml" json:"-"`
}

type Content struct {
	Type string `xml:"type,attr" json:"type,omitempty"`
	Body string `xml:",chardata" json:"body,omitempty"`
}

func (c *Content) Text() string {
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

func (c *Content) HTML() string {
	if c == nil {
		return ""
	}

	if c.Type != "html" {
		return sanitize.HTML(emoji.Sprint(c.Body))
	}

	t, err := sanitize.HTMLAllowing(emoji.Sprint(c.Body))
	if err != nil {
		return sanitize.HTML(emoji.Sprint(c.Body))
	}

	return t
}
