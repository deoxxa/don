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
	addr = flag.String("addr", ":3000", "Address to listen on.")
)

func main() {
	flag.Parse()

	// this has to be here for rice to work
	_, _ = rice.FindBox("public")
	_, _ = rice.FindBox("templates")

	cfg := rice.Config{LocateOrder: []rice.LocateMethod{
		rice.LocateWorkingDirectory,
		rice.LocateFS,
		rice.LocateEmbedded,
	}}

	publicBox := cfg.MustFindBox("public")
	templateBox := cfg.MustFindBox("templates")

	rootTemplate := template.New("root")

	if err := templateBox.Walk("/", func(name string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(name, ".html") || strings.HasPrefix(name, "page_") {
			return nil
		}

		if _, err := rootTemplate.Parse(templateBox.MustString(name)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		panic(err)
	}

	templateHome := template.Must(template.Must(rootTemplate.Clone()).Parse(templateBox.MustString("page_home.html")))
	templateFeed := template.Must(template.Must(rootTemplate.Clone()).Parse(templateBox.MustString("page_feed.html")))

	m := mux.NewRouter()

	m.Methods("GET").Path("/").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("content-type", "text/html")
		if err := templateHome.Execute(rw, nil); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	m.Methods("GET").Path("/show-feed").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		feed, err := AtomFetch(r.URL.Query().Get("url"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.Header().Set("content-type", "text/html")
		rw.WriteHeader(http.StatusOK)

		if err := templateFeed.Execute(rw, map[string]interface{}{"Feed": feed}); err != nil {
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

	m.NotFoundHandler = http.FileServer(publicBox.HTTPBox())

	n := negroni.New()

	n.Use(negronilogrus.NewMiddleware())
	n.Use(negroni.NewRecovery())
	n.Use(gzip.Gzip(gzip.BestCompression))
	n.UseHandler(m)

	if err := http.ListenAndServe(*addr, n); err != nil {
		panic(err)
	}
}
