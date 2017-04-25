package activitystreams

import (
	"html"
	"strings"
)

type Comment struct {
	GenericObject

	Content []Content `xml:"http://www.w3.org/2005/Atom content,omitempty" json:"content,omitempty"`
}

func (c *Comment) GetContent() string {
	for _, e := range c.Content {
		if e.Type == "html" {
			return e.Body
		}
	}

	for _, e := range c.Content {
		return strings.Replace(html.EscapeString(e.Body), "\n", "<br>", -1)
	}

	return ""
}
