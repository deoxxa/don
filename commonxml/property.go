package commonxml

type Property struct {
	Type  string `xml:"type,attr,omitempty" json:"type,omitempty"`
	Value string `xml:",chardata" json:"value,omitempty"`
}

type HasProperties struct {
	Property []Property `xml:"link" json:"link,omitempty"`
}

func (v *HasProperties) GetProperty(typ string) *Property {
	for _, l := range v.Property {
		if l.Type == typ {
			return &l
		}
	}

	return nil
}

func (v *HasProperties) GetProperties(typ string) []Property {
	var a []Property

	for _, l := range v.Property {
		if l.Type == typ {
			a = append(a, l)
		}
	}

	return a
}
