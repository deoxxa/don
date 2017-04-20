package activitystreams2

import (
	"encoding/xml"
	"html"
	"strings"
	"time"

	"fknsrs.biz/p/don/commonxml"
)

type baseEntry struct {
	commonxml.HasLinks

	feed *Feed `xml:"-" json:"-"`

	XMLName    xml.Name           `xml:"http://www.w3.org/2005/Atom entry" json:"-"`
	ID         string             `xml:"http://www.w3.org/2005/Atom id" json:"id,omitempty"`
	Title      string             `xml:"http://www.w3.org/2005/Atom title" json:"title,omitempty"`
	Summary    string             `xml:"http://www.w3.org/2005/Atom summary,omitempty" json:"summary,omitempty"`
	Content    []Content          `xml:"http://www.w3.org/2005/Atom content,omitempty" json:"content,omitempty"`
	Published  time.Time          `xml:"http://www.w3.org/2005/Atom published" json:"published,omitempty"`
	Updated    time.Time          `xml:"http://www.w3.org/2005/Atom updated" json:"updated,omitempty"`
	Author     *Author            `xml:"http://www.w3.org/2005/Atom author" json:"author,omitempty"`
	Verb       string             `xml:"http://activitystrea.ms/spec/1.0/ verb" json:"verb,omitempty"`
	ObjectType string             `xml:"http://activitystrea.ms/spec/1.0/ object-type" json:"objectType,omitempty"`
	Object     *commonxml.DOMNode `xml:"http://activitystrea.ms/spec/1.0/ object" json:"object,omitempty"`
	InReplyTo  *InReplyTo         `xml:"http://purl.org/syndication/thread/1.0 in-reply-to" json:"inReplyTo,omitempty"`
}

type Entry struct {
	baseEntry

	Object ObjectLike `json:"object"`
}

func (a *Entry) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if err := d.DecodeElement(&a.baseEntry, &start); err != nil {
		return err
	}

	if n := a.baseEntry.Object; n != nil {
		if c := n.GetChildByTagName(xml.Name{Space: "http://activitystrea.ms/spec/1.0/", Local: "object-type"}); c != nil {
			switch c.Text() {
			case "http://activitystrea.ms/schema/1.0/activity":
				a.Object = &Activity{}
			case "http://activitystrea.ms/schema/1.0/note":
				a.Object = &Note{}
			case "http://activitystrea.ms/schema/1.0/person":
				a.Object = &Author{}
			}
		}

		if a.Object == nil {
			a.Object = &GenericObject{}
		}

		if err := n.UnmarshalInto(a.Object); err != nil {
			return err
		}
	}

	return nil
}

func (a *Entry) GetID() string {
	return a.ID
}

func (a *Entry) GetName() string {
	return a.Title
}

func (a *Entry) GetSummary() string {
	if a.Object != nil && a.ObjectType != "" {
		if a.Summary != "" {
			return a.Summary
		}

		for _, c := range a.Content {
			if c.Type == "text/html" {
				return c.Body
			}
		}

		for _, c := range a.Content {
			return strings.Replace(html.EscapeString(c.Body), "\n", "<br>", -1)
		}
	}

	return ""
}

func (a *Entry) GetRepresentativeImage() string {
	return ""
}

func (a *Entry) GetPermalink() string {
	for _, l := range a.GetLinks("alternate") {
		if l.Type == "text/html" {
			return l.Href
		}
	}

	return ""
}

func (a *Entry) GetActor() *Author {
	if a.feed != nil {
		return a.feed.Author
	}

	return a.Author
}

func (a *Entry) GetObject() ObjectLike {
	if a.Object != nil {
		return a.Object
	}

	return a
}

func (a *Entry) GetObjectType() string {
	return a.ObjectType
}

func (a *Entry) GetVerb() string {
	if a.Verb == "" {
		return "http://activitystrea.ms/schema/1.0/post"
	}

	return a.Verb
}

func (a *Entry) GetTime() time.Time {
	return a.Published
}

func (a *Entry) GetTitle() string {
	return a.Title
}

func (e *Entry) GetContent() string {
	for _, c := range e.Content {
		if c.Type == "html" {
			return c.Body
		}
	}

	for _, c := range e.Content {
		return strings.Replace(html.EscapeString(c.Body), "\n", "<br>", -1)
	}

	return ""
}

func (e *Entry) GetInReplyTo() *InReplyTo {
	return e.InReplyTo
}
