package activitystreams

import (
	"encoding/xml"
	"strings"

	"fknsrs.biz/p/don/commonxml"
)

type GenericObject struct {
	commonxml.HasLinks

	ID         string `xml:"http://www.w3.org/2005/Atom id,omitempty" json:"id,omitempty"`
	Title      string `xml:"http://www.w3.org/2005/Atom title,omitempty" json:"title,omitempty"`
	Summary    string `xml:"http://www.w3.org/2005/Atom summary,omitempty" json:"summary,omitempty"`
	ObjectType string `xml:"http://activitystrea.ms/spec/1.0/ object-type" json:"objectType,omitempty"`
}

func (o *GenericObject) GetID() string {
	return o.ID
}

func (o *GenericObject) GetName() string {
	return o.Title
}

func (o *GenericObject) GetSummary() string {
	return o.Summary
}

func (o *GenericObject) GetObjectType() string {
	return o.ObjectType
}

func (o *GenericObject) GetRepresentativeImage() string {
	for _, l := range o.GetLinks("preview") {
		if attr := l.GetAttribute(xml.Name{Local: "type"}); attr != nil && strings.HasPrefix(attr.Value, "image/") {
			return l.Href
		}
	}

	return ""
}

func (o *GenericObject) GetPermalink() string {
	for _, l := range o.GetLinks("alternate") {
		if l.Type == "text/html" {
			return l.Href
		}
	}

	return ""
}
