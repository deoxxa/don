package activitystreams2

import (
	"html"
	"strings"
)

type Note struct {
	GenericObject

	Content []Content `xml:"http://www.w3.org/2005/Atom content,omitempty" json:"content,omitempty"`
}

func (n *Note) GetContent() string {
	for _, c := range n.Content {
		if c.Type == "html" {
			return c.Body
		}
	}

	for _, c := range n.Content {
		return strings.Replace(html.EscapeString(c.Body), "\n", "<br>", -1)
	}

	return ""
}
