package activitystreams2

import (
	"encoding/xml"
	"html"
	"strings"
	"time"

	"fknsrs.biz/p/don/commonxml"
)

type baseActivity struct {
	GenericObject

	XMLName   xml.Name           `xml:"http://activitystrea.ms/spec/1.0/ object" json:"-"`
	Content   []Content          `xml:"http://www.w3.org/2005/Atom content,omitempty" json:"content,omitempty"`
	Published time.Time          `xml:"http://www.w3.org/2005/Atom published" json:"published,omitempty"`
	Updated   time.Time          `xml:"http://www.w3.org/2005/Atom updated" json:"updated,omitempty"`
	Author    *Author            `xml:"http://www.w3.org/2005/Atom author" json:"author,omitempty"`
	Verb      string             `xml:"http://activitystrea.ms/spec/1.0/ verb" json:"verb,omitempty"`
	Object    *commonxml.DOMNode `xml:"http://activitystrea.ms/spec/1.0/ object" json:"object,omitempty"`
	InReplyTo *InReplyTo         `xml:"http://purl.org/syndication/thread/1.0 in-reply-to" json:"inReplyTo,omitempty"`
}

type Activity struct {
	baseActivity

	Object ObjectLike `xml:"http://activitystrea.ms/spec/1.0/ object" json:"object,omitempty"`
}

func (a *Activity) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if err := d.DecodeElement(&a.baseActivity, &start); err != nil {
		return err
	}

	if n := a.baseActivity.Object; n != nil {
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

func (a *Activity) GetSummary() string {
	if a.Summary != "" {
		return a.Summary
	}

	for _, c := range a.Content {
		if c.Type == "html" {
			return c.Body
		}
	}

	for _, c := range a.Content {
		return strings.Replace(html.EscapeString(c.Body), "\n", "<br>", -1)
	}

	return ""
}

func (a *Activity) GetActor() *Author {
	return a.Author
}

func (a *Activity) GetObject() ObjectLike {
	return a.Object
}

func (a *Activity) GetObjectType() string {
	if a.ObjectType != "" {
		return a.ObjectType
	}

	return "http://activitystrea.ms/schema/1.0/activityxxx"
}

func (a *Activity) GetVerb() string {
	return a.Verb
}

func (a *Activity) GetTime() time.Time {
	return a.Published
}

func (a *Activity) GetTitle() string {
	return a.Title
}

func (a *Activity) GetInReplyTo() *InReplyTo {
	return a.InReplyTo
}
