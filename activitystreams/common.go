package activitystreams

import (
	"encoding/xml"
)

type Content struct {
	Type string `xml:"type,attr" json:"type,omitempty"`
	Body string `xml:",chardata" json:"body,omitempty"`
}

type InReplyTo struct {
	XMLName xml.Name `xml:"http://purl.org/syndication/thread/1.0 in-reply-to" json:"-"`
	Ref     string   `xml:"ref,attr" json:"ref"`
	Href    string   `xml:"href,attr" json:"href"`
}
