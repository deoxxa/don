package main // import "fknsrs.biz/p/don"

import (
	"flag"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	"github.com/jtacoma/uritemplates"
	"github.com/meatballhat/negroni-logrus"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/urfave/negroni"
)

var (
	homeTemplate = template.Must(template.New("home").Parse(`
		<html>
			<head>
				<title>DON</title>
				<link rel="stylesheet" href="/style.css">
			</head>
			<body>
				<header>
					<h1><a href="/">DON</a></h1>
				</header>

				<div class="wrapper">
					<form id="find-feed" method="get" action="/find-feed" autocomplete="off">
						<input id="user" name="user" type="text" placeholder="e.g. your-username@your-provider.com" value="TheAdmin@mastodon.cloud" />
						<br>
						<input type="submit" value="Go!" />
					</form>

					<p class="blurb">
						This is a <em>ridiculously</em> simple, read-only OStatus client. Mostly an experiment. <a href="https://www.fknsrs.biz/p/don">Source code is available</a>.
					</p>
				</div>
			</body>
		</html>
	`))

	feedTemplate = template.Must(template.New("feed").Parse(`
		{{$feed := .Feed}}

		<html>
			<head>
				<title>{{$feed.Author}} @ DON</title>
				<link rel="stylesheet" href="/style.css">
			</head>
			<body>
				<header>
					<h1><a href="/">DON</a></h1>
				</header>

				<div class="wrapper">
					<h1>{{$feed.Author}}</h1>

					{{range $entry := $feed.Entry}}
						{{if $object := $entry.Object}}
							<h4>{{$object.Published}} (shared from {{$object.Author}} @ {{$entry.Published}})</h4>

							{{if $content := $object.Content}}
								<p>{{$content.HTML}}</p>
							{{end}}
						{{else if $content := $entry.Content}}
							{{if $inReplyTo := $entry.InReplyTo}}
								<h4>{{$entry.Published}} (<a href="{{$inReplyTo.Href}}">in reply to</a>)</h4>
							{{else}}
								<h4>{{$entry.Published}}</h4>
							{{end}}

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
				</div>
			</body>
		</html>
	`))
)

var (
	addr = flag.String("addr", ":3000", "Address to listen on.")
)

func main() {
	flag.Parse()

	// this has to be here for rice to work
	_, _ = rice.FindBox("public")

	cfg := rice.Config{LocateOrder: []rice.LocateMethod{
		rice.LocateWorkingDirectory,
		rice.LocateFS,
		rice.LocateEmbedded,
	}}

	box := cfg.MustFindBox("public")

	m := mux.NewRouter()

	m.Methods("GET").Path("/").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("content-type", "text/html")
		if err := homeTemplate.Execute(rw, nil); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	m.Methods("GET").Path("/find-feed").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		user := r.URL.Query().Get("user")
		if !strings.HasPrefix(user, "acct:") {
			user = "acct:" + user
		}

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

	m.Methods("GET").Path("/show-feed").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
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

	m.NotFoundHandler = http.FileServer(box.HTTPBox())

	n := negroni.New()

	n.Use(negronilogrus.NewMiddleware())
	n.Use(negroni.NewRecovery())
	n.Use(gzip.Gzip(gzip.BestCompression))
	n.UseHandler(m)

	if err := http.ListenAndServe(*addr, n); err != nil {
		panic(err)
	}
}
