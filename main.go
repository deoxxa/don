package main // import "fknsrs.biz/p/don"

import (
	"bytes"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/Sirupsen/logrus"
	"github.com/dyninc/qstring"
	"github.com/gorilla/mux"
	"github.com/jtacoma/uritemplates"
	_ "github.com/mattn/go-sqlite3"
	"github.com/meatballhat/negroni-logrus"
	"github.com/olebedev/go-duktape"
	"github.com/sebest/xff"
	"github.com/timewasted/go-accept-headers"
	"github.com/umisama/go-sqlbuilder"
	"github.com/umisama/go-sqlbuilder/dialects"
	"github.com/urfave/negroni"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app                   = kingpin.New("don", "Really really small OStatus node.")
	addr                  = app.Flag("addr", "Address to listen on.").Envar("ADDR").Default(":5000").String()
	database              = app.Flag("database", "Where to put the SQLite database.").Envar("DATABASE").Default("don.db").String()
	publicURL             = app.Flag("public_url", "URL to use for callbacks etc.").Envar("PUBLIC_URL").Required().String()
	logLevel              = app.Flag("log_level", "How much to log.").Default("INFO").Envar("LOG_LEVEL").Enum("DEBUG", "INFO", "WARN", "ERROR", "FATAL", "PANIC")
	pubsubRefreshInterval = app.Flag("pubsub_refresh_interval", "PubSub subscription refresh interval.").Default("15m").Envar("PUBSUB_REFRESH_INTERVAL").Duration()
	recordDocuments       = app.Flag("record_documents", "Record all XML documents for debugging.").Envar("RECORD_DOCUMENTS").Bool()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	sqlbuilder.SetDialect(dialects.Sqlite{})

	ll, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		panic(err)
	}
	logrus.SetLevel(ll)

	logrus.WithFields(logrus.Fields{
		"addr":       *addr,
		"database":   *database,
		"public_url": *publicURL,
		"log_level":  *logLevel,
	}).Info("starting up")

	if http.DefaultClient.Transport == nil {
		http.DefaultClient.Transport = http.DefaultTransport
	}

	if tr, ok := http.DefaultClient.Transport.(*http.Transport); ok {
		logrus.Debug("hack: disabling http2 - POST requests are currently broken")
		tr.TLSNextProto = map[string]func(authority string, c *tls.Conn) http.RoundTripper{}
	} else {
		logrus.Debug("hack: couldn't disable http2 - POST requests may not work")
	}

	// this has to be here for rice to work
	_, _ = rice.FindBox("public")
	_, _ = rice.FindBox("templates")
	_, _ = rice.FindBox("migrations")
	_, _ = rice.FindBox("build")

	cfg := rice.Config{LocateOrder: []rice.LocateMethod{
		rice.LocateWorkingDirectory,
		rice.LocateFS,
		rice.LocateAppended,
	}}

	publicBox := cfg.MustFindBox("public")
	templateBox := cfg.MustFindBox("templates")
	migrationBox := cfg.MustFindBox("migrations")
	buildBox := cfg.MustFindBox("build")

	db, err := sql.Open("sqlite3", *database)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := migrate(db, migrationBox); err != nil {
		panic(err)
	}

	a, err := NewApp(db)
	if err != nil {
		panic(err)
	}

	psc := NewPubSubClient(*publicURL+"/pubsub", NewPubSubSQLState(db), a.OnMessage)

	go func() {
		time.Sleep(time.Second * 2)

		for {
			logrus.Debug("refreshing pubsub subscriptions")
			if err := psc.Refresh(false, *pubsubRefreshInterval); err != nil {
				logrus.WithError(err).Error("couldn't refresh pubsub subscriptions")
			} else {
				logrus.Debug("refreshed pubsub subscriptions")
			}

			time.Sleep(*pubsubRefreshInterval)
		}
	}()

	vmCount := 1
	vms := make(chan *duktape.Context, vmCount)
	for i := 0; i < vmCount; i++ {
		vms <- nil
	}

	jsPrelude := `console = { log: function() {} }; module = { exports: null };`

	withVM := func(fn func(vm *duktape.Context) error) error {
		vm := <-vms
		defer func() {
			vms <- vm
		}()

		defer func() {
			if e := recover(); e != nil {
				vm = nil
			}
		}()

		if vm == nil {
			c := duktape.New()

			if err := c.PevalString(jsPrelude); err != nil {
				return err
			}

			if err := c.PevalString(buildBox.MustString("entry-server-bundle.js")); err != nil {
				return err
			}

			vm = c
		}

		if err := fn(vm); err != nil {
			vm.DestroyHeap()
			vm = nil
			return err
		}

		return nil
	}

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

	templateFeed := template.Must(template.Must(rootTemplate.Clone()).Parse(templateBox.MustString("page_feed.html")))
	templateReact := template.Must(template.Must(rootTemplate.Clone()).Parse(templateBox.MustString("page_react.html")))

	m := mux.NewRouter()

	m.PathPrefix("/pubsub").Handler(psc.Handler())

	m.Methods("GET").Path("/health").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	m.Methods("GET").Path("/").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		var args getPublicTimelineArgs
		if err := qstring.Unmarshal(r.URL.Query(), &args); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		posts, err := getPublicTimeline(db, args)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		initialState := map[string]interface{}{
			"publicTimeline": map[string]interface{}{
				"loading": false,
				"posts":   posts,
				"error":   nil,
			},
		}

		d, err := json.Marshal(initialState)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		acceptable := accept.Parse(r.Header.Get("accept"))

		ct, err := acceptable.Negotiate("text/html", "application/json")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		switch ct {
		case "application/json":
			rw.Header().Set("content-type", "application/json; charset=utf8")
			io.Copy(rw, bytes.NewReader(d))
			return
		case "text/html", "":
			var html string
			if err := withVM(func(vm *duktape.Context) error {
				if err := vm.PevalString("module.exports"); err != nil {
					return err
				}

				vm.PushString(r.URL.String())
				vm.PushString(string(d))
				if rc := vm.Pcall(2); rc != 0 {
					return fmt.Errorf("invalid rc: expected 0 but got %d; duktape says %q", rc, vm.SafeToString(-1))
				}
				html = vm.SafeToString(-1)
				vm.Pop()

				return nil
			}); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

			data := map[string]interface{}{
				"HTML":     template.HTML(html),
				"JSON":     template.JS(d),
				"CSSFiles": []string{"/build/vendor-styles.css", "/build/entry-client-styles.css"},
				"JSFiles":  []string{"/build/vendor-bundle.js", "/build/entry-client-bundle.js"},
				"Meta": map[string]interface{}{
					"Title":       "Home - DON",
					"Description": "A very basic StatusNet node. Kind of like Mastodon, but worse.",
				},
			}

			rw.Header().Set("content-type", "text/html; charset=utf-8")
			if err := templateReact.Execute(rw, data); err != nil {
				panic(err)
			}
		}
	})

	m.Methods("GET").Path("/show-feed").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		feed, err := AtomFetch(r.URL.Query().Get("url"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		hubLink := feed.GetLink("hub")

		if hubLink != nil && hubLink.Href != "" && feed.ID != "" {
			if err := psc.Subscribe(hubLink.Href, feed.ID); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		rw.Header().Set("content-type", "text/html; charset=utf8")
		rw.WriteHeader(http.StatusOK)

		if feed.Author != nil {
			if err := savePerson(db, r.URL.Query().Get("url"), feed.Author); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		for _, e := range feed.Entry {
			if err := saveEntry(db, r.URL.Query().Get("url"), &e); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
		}

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

		var feedURL string
		if err := db.QueryRow("select feed_url from people where email = $1", strings.TrimPrefix(user, "acct:")).Scan(&feedURL); err == nil {
			rw.Header().Set("location", "/show-feed?"+url.Values{"url": []string{feedURL}}.Encode())
			rw.WriteHeader(http.StatusSeeOther)
			return
		} else if err != sql.ErrNoRows {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
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
		if feedLink == nil || feedLink.Href == "" {
			http.Error(rw, "no feed link found in webfinger response", http.StatusInternalServerError)
			return
		}

		rw.Header().Set("location", "/show-feed?"+url.Values{"url": []string{feedLink.Href}}.Encode())
		rw.WriteHeader(http.StatusSeeOther)
	})

	m.PathPrefix("/build").Handler(http.StripPrefix("/build", http.FileServer(buildBox.HTTPBox())))

	m.NotFoundHandler = http.FileServer(publicBox.HTTPBox())

	xffh, err := xff.Default()
	if err != nil {
		panic(err)
	}

	n := negroni.New()

	n.Use(xffh)
	n.Use(negronilogrus.NewMiddleware())
	n.Use(negroni.NewRecovery())
	n.UseHandler(m)

	logrus.Info("starting server")

	if err := http.ListenAndServe(*addr, n); err != nil {
		panic(err)
	}
}
