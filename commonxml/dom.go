package commonxml

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type DOMType int

const (
	DOMTypeNone DOMType = iota
	DOMTypeElement
	DOMTypeText
)

type DOMNode struct {
	Type       DOMType
	XMLName    xml.Name
	Attributes []xml.Attr   `xml:",any,attr"`
	Children   []*DOMNode   `xml:",any"`
	CharData   xml.CharData `xml:",chardata"`
}

func (n *DOMNode) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.Type = DOMTypeElement
	n.XMLName = start.Name
	n.Attributes = start.Attr

	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		switch tok := t.(type) {
		case xml.CharData:
			data := make(xml.CharData, len(tok))
			copy(data, tok)
			n.Children = append(n.Children, &DOMNode{Type: DOMTypeText, CharData: data})
		case xml.StartElement:
			var v DOMNode
			if err := v.UnmarshalXML(d, tok); err != nil {
				return err
			}

			n.Children = append(n.Children, &v)
		case xml.EndElement:
			if tok.Name.Space == start.Name.Space && tok.Name.Local == start.Name.Local {
				return nil
			}

			return fmt.Errorf("DOMNode.UnmarshalXML: invalid end element %q %q; expected %q %q", tok.Name.Space, tok.Name.Local, start.Name.Space, start.Name.Local)
		default:
			return fmt.Errorf("DOMNode.UnmarshalXML: unhandled token type %T", t)
		}
	}

	return nil
}

func (n *DOMNode) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	switch n.Type {
	case DOMTypeText:
		return e.EncodeToken(n.CharData)
	case DOMTypeElement:
		if err := e.EncodeToken(xml.StartElement{Name: n.XMLName, Attr: n.Attributes}); err != nil {
			return err
		}

		for _, c := range n.Children {
			if err := c.MarshalXML(e, xml.StartElement{}); err != nil {
				return err
			}
		}

		if err := e.EncodeToken(xml.EndElement{Name: n.XMLName}); err != nil {
			return err
		}

		return nil
	default:
		return fmt.Errorf("DOMNode.MarshalXML: can't encode node type %v", n.Type)
	}
}

func (d *DOMNode) UnmarshalInto(v interface{}) error {
	x, err := xml.Marshal(d)
	if err != nil {
		return err
	}

	return xml.Unmarshal(x, v)
}

func (d *DOMNode) Text() string {
	var a []string

	for _, c := range d.Children {
		switch c.Type {
		case DOMTypeText:
			a = append(a, string(c.CharData))
		case DOMTypeElement:
			a = append(a, c.Text())
		}
	}

	return strings.Join(a, "")
}

func (d *DOMNode) GetChildByTagName(name xml.Name) *DOMNode {
	for _, e := range d.Children {
		if e.XMLName.Local == name.Local && (name.Space == "" || e.XMLName.Space == name.Space) {
			return e
		}
	}

	return nil
}

func (d *DOMNode) GetChildrenByTagName(name xml.Name) []*DOMNode {
	var a []*DOMNode

	for _, e := range d.Children {
		if e.XMLName.Local == name.Local && (name.Space == "" || e.XMLName.Space == name.Space) {
			a = append(a, e)
		}
	}

	return a
}

func (d *DOMNode) GetElementByTagName(name xml.Name) *DOMNode {
	for _, e := range d.Children {
		if e.XMLName.Local == name.Local && (name.Space == "" || e.XMLName.Space == name.Space) {
			return e
		}

		if v := e.GetElementByTagName(name); v != nil {
			return v
		}
	}

	return nil
}

func (d *DOMNode) GetElementsByTagName(name xml.Name) []*DOMNode {
	var a []*DOMNode

	for _, e := range d.Children {
		if e.XMLName.Local == name.Local && (name.Space == "" || e.XMLName.Space == name.Space) {
			a = append(a, e)
		}

		a = append(a, e.GetElementsByTagName(name)...)
	}

	return a
}

func (d *DOMNode) GetAttribute(name xml.Name) *xml.Attr {
	for _, e := range d.Attributes {
		if e.Name.Local == name.Local && (name.Space == "" || e.Name.Space == name.Space) {
			return &e
		}
	}

	return nil
}

func (d *DOMNode) GetAttributes(name xml.Name) []xml.Attr {
	var a []xml.Attr

	for _, e := range d.Attributes {
		if e.Name.Local == name.Local && (name.Space == "" || e.Name.Space == name.Space) {
			a = append(a, e)
		}
	}

	return a
}
