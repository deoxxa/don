package activitystreams2

import (
	"encoding/xml"
	"time"

	"fknsrs.biz/p/don/commonxml"
)

type baseFeed struct {
	commonxml.HasLinks

	XMLName    xml.Name  `xml:"http://www.w3.org/2005/Atom feed" json:"-"`
	Title      string    `xml:"http://www.w3.org/2005/Atom title" json:"title,omitempty"`
	ID         string    `xml:"http://www.w3.org/2005/Atom id" json:"id,omitempty"`
	Updated    time.Time `xml:"http://www.w3.org/2005/Atom updated" json:"updated,omitempty"`
	Author     *Author   `xml:"http://www.w3.org/2005/Atom author" json:"author,omitempty"`
	Activities []Entry   `xml:"http://www.w3.org/2005/Atom entry" json:"entry,omitempty"`
}

type Feed struct{ baseFeed }

func (f *Feed) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if err := d.DecodeElement(&f.baseFeed, &start); err != nil {
		return err
	}

	for i := range f.Activities {
		f.Activities[i].feed = f
	}

	return nil
}

func (f *Feed) GetActivities() []ActivityLike {
	a := make([]ActivityLike, len(f.Activities))

	for i := range f.Activities {
		a[i] = &f.Activities[i]
	}

	return a
}

func (f *Feed) GetHub() string {
	if l := f.GetLink("hub"); l != nil {
		return l.Href
	}

	return ""
}

func (f *Feed) GetSalmon() string {
	if l := f.GetLink("salmon"); l != nil {
		return l.Href
	}

	return ""
}
