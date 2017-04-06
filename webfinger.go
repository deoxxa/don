package main

import (
	"encoding/json"
	"encoding/xml"
	"mime"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

var ErrWebfingerNotFound = errors.New("Webfinger: resource not found")

func WebfingerMakeURL(domain, resource string, rel []string) string {
	return "https://" + domain + "/.well-known/webfinger?" + url.Values{
		"resource": []string{resource},
		"rel":      rel,
	}.Encode()
}

func WebfingerFetch(u string) (*WebfingerResponse, error) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "WebfingerFetch")
	}
	req.Header.Set("accept", "application/jrd+json, application/xrd+xml")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "WebfingerFetch")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		if res.StatusCode == 404 {
			return nil, ErrWebfingerNotFound
		}

		return nil, errors.Errorf("WebfingerFetch: invalid status code; expected 200 but got %d", res.StatusCode)
	}

	media := "application/jrd+json"
	if mt, _, err := mime.ParseMediaType(res.Header.Get("content-type")); err == nil {
		media = mt
	}

	var v WebfingerResponse

	switch media {
	case "application/xrd+xml":
		if err := xml.NewDecoder(res.Body).Decode(&v); err != nil {
			return nil, errors.Wrap(err, "WebfingerFetch")
		}
	default:
		if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
			return nil, errors.Wrap(err, "WebfingerFetch")
		}
	}

	return &v, nil
}

type WebfingerResponse struct {
	Subject    string             `json:"subject"`
	Aliases    []string           `json:"aliases,omitempty"`
	Properties map[string]*string `json:"properties,omitempty"`
	Links      []WebfingerLink    `json:"links,omitempty"`
}

func (r *WebfingerResponse) GetLink(rel string) *WebfingerLink {
	for _, l := range r.Links {
		if l.Rel == rel {
			return &l
		}
	}

	return nil
}

type webfingerResponseXML struct {
	XMLName    xml.Name               `xml:"http://docs.oasis-open.org/ns/xri/xrd-1.0 XRD"`
	Subject    string                 `xml:"Subject"`
	Aliases    []string               `xml:"Alias"`
	Properties []webfingerXMLProperty `xml:"Property"`
	Links      []WebfingerLink        `xml:"Link"`
}

func (r *WebfingerResponse) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	v := webfingerResponseXML{
		Subject: r.Subject,
		Aliases: r.Aliases,
		Links:   r.Links,
	}

	for key, value := range r.Properties {
		if value == nil {
			v.Properties = append(v.Properties, webfingerXMLProperty{Type: key, Value: ""})
		} else {
			v.Properties = append(v.Properties, webfingerXMLProperty{Type: key, Value: *value})
		}
	}

	return e.Encode(v)
}

type WebfingerLink struct {
	Rel        string             `json:"rel,omitempty"`
	Type       string             `json:"type,omitempty"`
	Href       string             `json:"href,omitempty"`
	Template   string             `json:"template,omitempty"`
	Titles     map[string]string  `json:"titles,omitempty"`
	Properties map[string]*string `json:"properties,omitempty"`
}

type webfingerLinkXML struct {
	XMLName    xml.Name               `xml:"Link" json:"-"`
	Rel        string                 `xml:"rel,attr,omitempty"`
	Type       string                 `xml:"type,attr,omitempty"`
	Href       string                 `xml:"href,attr,omitempty"`
	Template   string                 `xml:"template,attr,omitempty"`
	Titles     []webfingerXMLTitle    `xml:"title"`
	Properties []webfingerXMLProperty `xml:"property"`
}

type webfingerXMLTitle struct {
	Lang  string `xml:"lang,attr,omitempty"`
	Value string `xml:",chardata"`
}

type webfingerXMLProperty struct {
	Type  string `xml:"type,attr,omitempty"`
	Value string `xml:",chardata"`
}

func (m *WebfingerLink) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v webfingerLinkXML
	if err := d.DecodeElement(&v, &start); err != nil {
		return errors.Wrap(err, "WebfingerLink.UnmarshalXML")
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

func (m *WebfingerLink) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	v := webfingerLinkXML{
		Rel:      m.Rel,
		Type:     m.Type,
		Href:     m.Href,
		Template: m.Template,
	}

	for key, value := range m.Titles {
		v.Titles = append(v.Titles, webfingerXMLTitle{Lang: key, Value: value})
	}
	for key, value := range m.Properties {
		if value == nil {
			v.Properties = append(v.Properties, webfingerXMLProperty{Type: key, Value: ""})
		} else {
			v.Properties = append(v.Properties, webfingerXMLProperty{Type: key, Value: *value})
		}
	}

	return e.Encode(v)
}

type WebfingerSource interface {
	Fetch(subject string, rel []string) (*WebfingerResponse, error)
}

type WebfingerHandler struct{ WebfingerSource }

func (s *WebfingerHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	subject := q.Get("resource")
	if subject == "" {
		http.Error(rw, "missing resource query parameter", http.StatusBadRequest)
		return
	}

	res, err := s.Fetch(subject, q["rel"])
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if rel := q["rel"]; len(rel) != 0 {
		m := make(map[string]bool)
		for _, e := range rel {
			m[e] = true
		}

		var a []WebfingerLink

		for _, l := range res.Links {
			if m[l.Rel] {
				a = append(a, l)
			}
		}

		res.Links = a
	}

	rw.Header().Set("content-type", "application/jrd+json")
	rw.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(rw).Encode(res); err != nil {
		panic(err)
	}
}
