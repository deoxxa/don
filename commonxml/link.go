package commonxml

type Link struct {
	Rel      string `xml:"rel,attr,omitempty" json:"rel,omitempty"`
	Type     string `xml:"type,attr,omitempty" json:"type,omitempty"`
	Href     string `xml:"href,attr" json:"href,omitempty"`
	HrefLang string `xml:"hreflang,attr,omitempty" json:"hrefLang,omitempty"`
	Template string `xml:"template,attr,omitempty" json:"template,omitempty"`
	Title    string `xml:"title,attr,omitempty" json:"title,omitempty"`
	Length   uint   `xml:"length,attr,omitempty" json:"length,omitempty"`
}

type HasLinks struct {
	Link []Link `xml:"link" json:"link,omitempty"`
}

func (v *HasLinks) GetLink(rel string) *Link {
	for _, l := range v.Link {
		if l.Rel == rel {
			return &l
		}
	}

	return nil
}

func (v *HasLinks) GetLinks(rel string) []Link {
	var a []Link

	for _, l := range v.Link {
		if l.Rel == rel {
			a = append(a, l)
		}
	}

	return a
}
