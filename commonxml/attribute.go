package commonxml

import (
	"encoding/xml"
)

type HasAttributes struct {
	Attributes []xml.Attr `xml:",any,attr" json:"attributes,omitempty"`
}

func (a *HasAttributes) GetAttribute(name xml.Name) *xml.Attr {
	for _, e := range a.Attributes {
		if e.Name == name {
			return &e
		}
	}

	return nil
}

func (a *HasAttributes) GetAttributes(name xml.Name) []xml.Attr {
	var r []xml.Attr

	for _, e := range a.Attributes {
		if name.Local == e.Name.Local && (name.Space == e.Name.Space || name.Space == "") {
			r = append(r, e)
		}
	}

	return r
}
