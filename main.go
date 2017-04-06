package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/jtacoma/uritemplates"
	"github.com/kr/text"
)

func formatText(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")

	for i := range lines {
		lines[i] = text.Indent(text.Wrap(lines[i], 78), "  ")
	}

	return strings.Join(lines, "\n")
}

var (
	user = flag.String("user", "acct:gargron@mastodon.social", "User to look up.")
)

func main() {
	flag.Parse()

	fmt.Printf("parsing acct url: %q\n", *user)
	acct, err := AcctFromString(*user)
	if err != nil {
		panic(err)
	}
	fmt.Printf("done\n")

	fmt.Printf("fetching host metadata: %s\n", acct.Host)
	hm, err := HostMetaFetch(acct.Host)
	if err != nil {
		panic(err)
	}
	fmt.Printf("done\n")

	fmt.Printf("getting lrdd link...\n")
	lrdd := hm.GetLink("lrdd")
	if lrdd == nil {
		panic(fmt.Errorf("no lrdd link found in host metadata"))
	}

	var lrddHref string
	switch {
	case lrdd.Href != "":
		lrddHref = lrdd.Href
	case lrdd.Template != "":
		lrddHrefTemplate, err := uritemplates.Parse(lrdd.Template)
		if err != nil {
			panic(err)
		}

		s, err := lrddHrefTemplate.Expand(map[string]interface{}{"uri": acct.String()})
		if err != nil {
			panic(err)
		}

		lrddHref = s
	}
	fmt.Printf("done: %s\n", lrddHref)

	fmt.Printf("fetching lrdd link as webfinger...\n")
	wf, err := WebfingerFetch(lrddHref)
	if err != nil {
		panic(err)
	}
	fmt.Printf("done\n")

	if feedLink := wf.GetLink("http://schemas.google.com/g/2010#updates-from"); feedLink != nil {
		href := feedLink.Href

	begin:
		feed, err := AtomFetch(href)
		if err != nil {
			panic(err)
		}

		for _, e := range feed.Entry {
			if e.Verb == "http://activitystrea.ms/schema/1.0/share" && e.ObjectType == "http://activitystrea.ms/schema/1.0/activity" && e.Object != nil {
				fmt.Printf("[%s] (retweet from %s; %s)\n", e.Published, e.Object.Author, e.Object.Published)
				fmt.Printf("%s\n", formatText(e.Object.Content.String()))
			} else {
				fmt.Printf("[%s]\n%s\n", e.Published, formatText(e.Content.String()))
			}
		}

		if l := feed.GetLink("next"); l != nil {
			href = l.Href
			goto begin
		}
	}
}
