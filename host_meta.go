package main

import (
	"encoding/xml"
	"net/http"

	"github.com/pkg/errors"
)

func HostMetaFetch(domain string) (*HostMetaResponse, error) {
	req, err := http.NewRequest("GET", "https://"+domain+"/.well-known/host-meta", nil)
	if err != nil {
		return nil, errors.Wrap(err, "HostMetaFetch")
	}
	req.Header.Set("accept", "application/xrd+xml")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "HostMetaFetch")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.Errorf("HostMetaFetch: invalid status code; expected 200 but got %d", res.StatusCode)
	}

	var v HostMetaResponse
	if err := xml.NewDecoder(res.Body).Decode(&v); err != nil {
		return nil, errors.Wrap(err, "HostMetaFetch")
	}

	return &v, nil
}

type HostMetaResponse struct {
	Properties map[string]*string `json:"properties,omitempty"`
	Links      []HostMetaLink     `json:"links,omitempty"`
}

func (r *HostMetaResponse) GetLink(rel string) *HostMetaLink {
	for _, l := range r.Links {
		if l.Rel == rel {
			return &l
		}
	}

	return nil
}

type hostMetaResponseXML struct {
	XMLName    xml.Name              `xml:"http://docs.oasis-open.org/ns/xri/xrd-1.0 XRD"`
	Properties []hostMetaXMLProperty `xml:"Property"`
	Links      []HostMetaLink        `xml:"Link"`
}

func (r *HostMetaResponse) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	v := hostMetaResponseXML{Links: r.Links}

	for key, value := range r.Properties {
		if value == nil {
			v.Properties = append(v.Properties, hostMetaXMLProperty{Type: key, Value: ""})
		} else {
			v.Properties = append(v.Properties, hostMetaXMLProperty{Type: key, Value: *value})
		}
	}

	return e.Encode(v)
}

func (r *HostMetaResponse) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v hostMetaResponseXML
	if err := d.DecodeElement(&v, &start); err != nil {
		return errors.Wrap(err, "HostMetaResponse.UnmarshalXML")
	}

	r.Links = v.Links

	for _, p := range v.Properties {
		if p.Value == "" {
			r.Properties[p.Type] = nil
		} else {
			r.Properties[p.Type] = &p.Value
		}
	}

	return nil
}

type HostMetaLink struct {
	Rel        string             `json:"rel,omitempty"`
	Type       string             `json:"type,omitempty"`
	Href       string             `json:"href,omitempty"`
	Template   string             `json:"template,omitempty"`
	Titles     map[string]string  `json:"titles,omitempty"`
	Properties map[string]*string `json:"properties,omitempty"`
}

type hostMetaLinkXML struct {
	XMLName    xml.Name              `xml:"Link" json:"-"`
	Rel        string                `xml:"rel,attr,omitempty"`
	Type       string                `xml:"type,attr,omitempty"`
	Href       string                `xml:"href,attr,omitempty"`
	Template   string                `xml:"template,attr,omitempty"`
	Titles     []hostMetaXMLTitle    `xml:"title"`
	Properties []hostMetaXMLProperty `xml:"property"`
}

type hostMetaXMLTitle struct {
	Lang  string `xml:"lang,attr,omitempty"`
	Value string `xml:",chardata"`
}

type hostMetaXMLProperty struct {
	Type  string `xml:"type,attr,omitempty"`
	Value string `xml:",chardata"`
}

func (m *HostMetaLink) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v hostMetaLinkXML
	if err := d.DecodeElement(&v, &start); err != nil {
		return errors.Wrap(err, "HostMetaLink.UnmarshalXML")
	}

	m.Rel = v.Rel
	m.Type = v.Type
	m.Href = v.Href
	m.Template = v.Template
	m.Titles = make(map[string]string)
	m.Properties = make(map[string]*string)

	for _, t := range v.Titles {
		m.Titles[t.Lang] = t.Value
	}

	for _, p := range v.Properties {
		if p.Value == "" {
			m.Properties[p.Type] = nil
		} else {
			m.Properties[p.Type] = &p.Value
		}
	}

	return nil
}

func (m *HostMetaLink) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	v := hostMetaLinkXML{
		Rel:      m.Rel,
		Type:     m.Type,
		Href:     m.Href,
		Template: m.Template,
	}

	for key, value := range m.Titles {
		v.Titles = append(v.Titles, hostMetaXMLTitle{Lang: key, Value: value})
	}
	for key, value := range m.Properties {
		if value == nil {
			v.Properties = append(v.Properties, hostMetaXMLProperty{Type: key, Value: ""})
		} else {
			v.Properties = append(v.Properties, hostMetaXMLProperty{Type: key, Value: *value})
		}
	}

	return e.Encode(v)
}
