package main

import (
	"flag"
	"html/template"
	"net/http"
	"net/url"

	"github.com/jtacoma/uritemplates"
)

var (
	homeTemplate = template.Must(template.New("home").Parse(`
		<html>
			<head>
				<title>don</title>
			</head>
			<body>
				<header>
					<a href="/">home</a>
				</header>

				<form id="find-feed" method="get" action="/find-feed">
					<label for="user">user</label>
					<input id="user" name="user" placeholder="e.g. acct:your-username@mastodon.social" />
					<input type="submit" />
				</form>
			</body>
		</html>
	`))

	feedTemplate = template.Must(template.New("feed").Parse(`
		{{$feed := .Feed}}

		<html>
			<head>
				<title>{{$feed.Author}} @ don</title>
			</head>
			<body>
				<header>
					<a href="/">home</a>
				</header>

				<h1>{{$feed.Author}}</h1>

				{{range $entry := $feed.Entry}}
					{{if $object := $entry.Object}}
						<h4>{{$object.Published}} (shared from {{$object.Author}} @ {{$entry.Published}})</h4>

						{{if $content := $object.Content}}
							<p>{{$content.HTML}}</p>
						{{end}}
					{{else if $content := $entry.Content}}
						<h4>{{$entry.Published}}</h4>

						<p>{{$content.HTML}}</p>
					{{else if eq $entry.Verb "http://activitystrea.ms/schema/1.0/delete"}}
						<h4>{{$entry.Published}}</h4>

						<s>(deleted @ {{$entry.Updated}})</s>
					{{end}}

					<hr>
				{{end}}

				{{if $link := $feed.GetLink "next"}}
					<a href="/show-feed?url={{$link.Href}}" rel="next">earlier</a>
				{{end}}
			</body>
		</html>
	`))
)

var (
	addr = flag.String("addr", ":3000", "Address to listen on.")
)

func main() {
	flag.Parse()

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("content-type", "text/html")
		if err := homeTemplate.Execute(rw, nil); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/find-feed", func(rw http.ResponseWriter, r *http.Request) {
		user := r.URL.Query().Get("user")

		acct, err := AcctFromString(user)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		hm, err := HostMetaFetch(acct.Host)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		lrdd := hm.GetLink("lrdd")
		if lrdd == nil {
			http.Error(rw, "no lrdd link found in host metadata", http.StatusInternalServerError)
			return
		}

		var lrddHref string
		switch {
		case lrdd.Href != "":
			lrddHref = lrdd.Href
		case lrdd.Template != "":
			lrddHrefTemplate, err := uritemplates.Parse(lrdd.Template)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

			s, err := lrddHrefTemplate.Expand(map[string]interface{}{"uri": acct.String()})
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

			lrddHref = s
		}

		wf, err := WebfingerFetch(lrddHref)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		feedLink := wf.GetLink("http://schemas.google.com/g/2010#updates-from")
		if feedLink == nil {
			http.Error(rw, "no feed link found in webfinger response", http.StatusInternalServerError)
			return
		}

		rw.Header().Set("location", "/show-feed?"+url.Values{"url": []string{feedLink.Href}}.Encode())
		rw.WriteHeader(http.StatusSeeOther)
	})

	http.HandleFunc("/show-feed", func(rw http.ResponseWriter, r *http.Request) {
		feed, err := AtomFetch(r.URL.Query().Get("url"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.Header().Set("content-type", "text/html")
		rw.WriteHeader(http.StatusOK)

		if err := feedTemplate.Execute(rw, map[string]interface{}{"Feed": feed}); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	if err := http.ListenAndServe(*addr, nil); err != nil {
		panic(err)
	}
}
