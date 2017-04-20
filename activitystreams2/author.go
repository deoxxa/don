package activitystreams2

import (
	"encoding/xml"
	"strconv"
	"strings"

	"fknsrs.biz/p/don/commonxml"
)

type Author struct {
	commonxml.HasLinks

	XMLName           xml.Name       `json:"-"`
	ID                string         `xml:"http://www.w3.org/2005/Atom id,omitempty" json:"id,omitempty"`
	Name              string         `xml:"http://www.w3.org/2005/Atom name" json:"name,omitempty"`
	URI               string         `xml:"http://www.w3.org/2005/Atom uri,omitempty" json:"uri,omitempty"`
	Email             string         `xml:"http://www.w3.org/2005/Atom email,omitempty" json:"email,omitempty"`
	Summary           string         `xml:"http://www.w3.org/2005/Atom summary,omitempty" json:"summary,omitempty"`
	ObjectType        string         `xml:"http://activitystrea.ms/spec/1.0/ object-type" json:"objectType,omitempty"`
	PreferredUsername string         `xml:"http://portablecontacts.net/spec/1.0 preferredUsername" json:"preferredUsername,omitempty"`
	DisplayName       string         `xml:"http://portablecontacts.net/spec/1.0 displayName" json:"displayName,omitempty"`
	Note              string         `xml:"http://portablecontacts.net/spec/1.0 note" json:"note,omitempty"`
	URLs              []AuthorURL    `xml:"http://portablecontacts.net/spec/1.0 urls" json:"urls,omitempty"`
	Address           *AuthorAddress `xml:"http://portablecontacts.net/spec/1.0 address" json:"address,omitempty"`
	Scope             string         `xml:"http://mastodon.social/schema/1.0 scope" json:"scope,omitempty"`
}

type AuthorURL struct {
	Type    string `xml:"http://portablecontacts.net/spec/1.0 type" json:"type"`
	Value   string `xml:"http://portablecontacts.net/spec/1.0 value" json:"value"`
	Primary bool   `xml:"http://portablecontacts.net/spec/1.0 primary" json:"primary"`
}

type AuthorAddress struct {
	Formatted string `xml:"http://portablecontacts.net/spec/1.0 formatted" json:"type"`
}

func (a *Author) GetID() string {
	return a.ID
}

func (a *Author) GetName() string {
	return a.Name
}

func (a *Author) GetSummary() string {
	return a.Summary
}

func (a *Author) GetRepresentativeImage() string {
	for _, l := range a.GetLinks("preview") {
		if strings.HasPrefix(l.Type, "image/") {
			return l.Href
		}
	}

	return ""
}

func (a *Author) GetPermalink() string {
	for _, l := range a.GetLinks("alternate") {
		if l.Type == "text/html" {
			return l.Href
		}
	}

	return a.URI
}

func (a *Author) GetObjectType() string {
	return "http://activitystrea.ms/schema/1.0/person"
}

func (a *Author) GetBestAvatar() string {
	var best string

	var bestSize int64
	for _, l := range a.GetLinks("avatar") {
		if best == "" {
			best = l.Href
		}

		attr := l.GetAttribute(xml.Name{Space: "http://purl.org/syndication/atommedia", Local: "height"})
		if attr == nil {
			continue
		}

		if n, err := strconv.ParseInt(attr.Value, 10, 64); err == nil && n > bestSize {
			bestSize = n
			best = l.Href
		}
	}

	return best
}
