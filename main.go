package main // import "fknsrs.biz/p/don"

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/Sirupsen/logrus"
	"github.com/bernerdschaefer/eventsource"
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"
	"github.com/jtacoma/uritemplates"
	_ "github.com/mattn/go-sqlite3"
	"github.com/meatballhat/negroni-logrus"
	"github.com/sebest/xff"
	"github.com/umisama/go-sqlbuilder"
	"github.com/umisama/go-sqlbuilder/dialects"
	"github.com/urfave/negroni"
	"gopkg.in/alecthomas/kingpin.v2"

	"fknsrs.biz/p/don/acct"
	"fknsrs.biz/p/don/activitystreams"
	"fknsrs.biz/p/don/hostmeta"
	"fknsrs.biz/p/don/pubsub"
	"fknsrs.biz/p/don/react"
	"fknsrs.biz/p/don/webfinger"
)

var (
	app                   = kingpin.New("don", "Really really small OStatus node.")
	addr                  = app.Flag("addr", "Address to listen on.").Envar("ADDR").Default(":5000").String()
	database              = app.Flag("database", "Where to put the SQLite database.").Envar("DATABASE").Default("don.db").String()
	cache                 = app.Flag("cache", "Where to put the cache file.").Envar("CACHE").Default("don.cache").String()
	publicURL             = app.Flag("public_url", "URL to use for callbacks etc.").Envar("PUBLIC_URL").Required().String()
	logLevel              = app.Flag("log_level", "How much to log.").Default("INFO").Envar("LOG_LEVEL").Enum("DEBUG", "INFO", "WARN", "ERROR", "FATAL", "PANIC")
	pubsubRefreshInterval = app.Flag("pubsub_refresh_interval", "PubSub subscription refresh interval.").Default("15m").Envar("PUBSUB_REFRESH_INTERVAL").Duration()
	recordDocuments       = app.Flag("record_documents", "Record all XML documents for debugging.").Envar("RECORD_DOCUMENTS").Bool()
	reactRenderer         = app.Flag("react_renderer", "React server rendering strategy.").Envar("REACT_RENDERER").Default("duktape").Enum("duktape", "node")
	externalJS            = app.Flag("external_js", "Load client JS from an external location.").Envar("EXTERNAL_JS").String()
	cookieSigningKey      = app.Flag("cookie_signing_key", "Key for signing cookies.").Envar("COOKIE_SIGNING_KEY").Required().HexBytes()
	cookieEncryptionKey   = app.Flag("cookie_encryption_key", "Key for encrypting cookies.").Envar("COOKIE_ENCRYPTION_KEY").Required().HexBytes()
	sqlQueryLog           = app.Flag("sql_query_log", "Enable SQL query logging.").Envar("SQL_QUERY_LOG").Bool()
)

var decoder *schema.Decoder

func init() {
	decoder = schema.NewDecoder()
}

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	sqlbuilder.SetDialect(dialects.Postgresql{})

	ll, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		panic(err)
	}
	logrus.SetLevel(ll)

	logrus.WithFields(logrus.Fields{
		"addr":                    *addr,
		"database":                *database,
		"public_url":              *publicURL,
		"log_level":               *logLevel,
		"pubsub_refresh_interval": *pubsubRefreshInterval,
		"record_documents":        *recordDocuments,
		"react_renderer":          *reactRenderer,
		"external_js":             *externalJS,
		"cookie_signing_key":      strings.Repeat("*", len(*cookieSigningKey)),
		"cookie_encryption_key":   strings.Repeat("*", len(*cookieEncryptionKey)),
	}).Info("starting up")

	if http.DefaultClient.Transport == nil {
		http.DefaultClient.Transport = http.DefaultTransport
	}

	if tr, ok := http.DefaultClient.Transport.(*http.Transport); ok {
		logrus.Debug("hack: disabling http2 - client POST requests are currently broken")
		tr.TLSNextProto = map[string]func(authority string, c *tls.Conn) http.RoundTripper{}
	} else {
		logrus.Debug("hack: couldn't disable http2 - client POST requests may not work")
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

	sqlDB, err := sql.Open("sqlite3", *database)
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()

	boltDB, err := bolt.Open(*cache, 0644, nil)
	if err != nil {
		panic(err)
	}
	defer boltDB.Close()

	if err := migrate(sqlDB, migrationBox); err != nil {
		panic(err)
	}

	ss := sessions.NewCookieStore(*cookieSigningKey, *cookieEncryptionKey)
	ss.Options = &sessions.Options{HttpOnly: true, Secure: strings.HasPrefix(*publicURL, "https:")}

	var renderer react.Renderer
	switch *reactRenderer {
	case "duktape":
		renderer = react.NewDuktapeRenderer(1)
	case "node":
		renderer = react.NewNodeJSRenderer(1)
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

	templateReact := template.Must(template.Must(rootTemplate.Clone()).Parse(templateBox.MustString("page_react.html")))

	a, err := NewApp(sqlDB, boltDB, ss, renderer, templateReact, buildBox)
	if err != nil {
		panic(err)
	}

	psc := pubsub.NewClient(*publicURL+"/pubsub", pubsub.NewSQLiteState(sqlDB), a.OnMessage)

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

	m := mux.NewRouter()

	m.Methods("GET").Path("/health").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	m.PathPrefix("/pubsub").Handler(psc.Handler())

	m.Methods("GET").Path("/").HandlerFunc(a.HandlerFor(a.handleHomeGet))
	m.Methods("GET").Path("/login").HandlerFunc(a.HandlerFor(a.handleLoginGet))
	m.Methods("POST").Path("/login").HandlerFunc(a.HandlerFor(a.handleLoginPost))
	m.Methods("GET").Path("/register").HandlerFunc(a.HandlerFor(a.handleRegisterGet))
	m.Methods("POST").Path("/register").HandlerFunc(a.HandlerFor(a.handleRegisterPost))
	m.Methods("GET").Path("/logout").HandlerFunc(a.HandlerFor(a.handleLogoutGet))
	m.Methods("POST").Path("/logout").HandlerFunc(a.HandlerFor(a.handleLogoutPost))

	m.Methods("GET").Path("/api/feed").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		var args getPublicTimelineArgs
		if err := decoder.Decode(&args, r.URL.Query()); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := a.StandardContext(rw, r); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		cn, ok := rw.(http.CloseNotifier)
		if !ok {
			http.Error(rw, "couldn't make CloseNotifier out of request", http.StatusInternalServerError)
			return
		}

		stop := cn.CloseNotify()
		ch := make(chan *ActivityEvent, 25)

		a.AddListener(ch)
		defer func() { a.RemoveListener(ch) }()

		rw.Header().Set("content-type", "text/event-stream")
		rw.WriteHeader(http.StatusOK)

		enc := eventsource.NewEncoder(rw)

	loop:
		for {
			select {
			case <-stop:
				break loop
			case ev := <-ch:
				if args.Q != "" {
					if ev.Activity.Object.Content == nil {
						continue
					}

					for _, s := range strings.Split(args.Q, " ") {
						s = strings.TrimSpace(s)

						if len(s) > 0 && !strings.Contains(*ev.Activity.Object.Content, s) {
							continue loop
						}
					}
				}

				if err := enc.Encode(eventsource.Event{
					Type: "activity",
					ID:   fmt.Sprintf("%d", ev.RowID),
					Data: ev.JSON,
				}); err != nil {
					panic(err)
				}

				if err := enc.Flush(); err != nil {
					panic(err)
				}
			case <-time.After(time.Second * 30):
				if err := enc.WriteField("", []byte("heartbeat")); err != nil {
					panic(err)
				}

				if err := enc.Flush(); err != nil {
					panic(err)
				}
			}
		}
	})

	m.Methods("POST").Path("/ingest-xml").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		var f activitystreams.Feed

		if err := xml.NewDecoder(r.Body).Decode(&f); err != nil {
			logrus.WithError(err).Debug("ingest-xml: couldn't parse body")
			return
		}

		for _, e := range f.Activities {
			if err := a.saveActivity(&e); err != nil {
				panic(err)
			}
		}
	})

	m.Methods("GET").Path("/show-feed").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		feedData, _, err := a.FeedCache.Get(r.URL.Query().Get("url"), nil)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		feed, err := activitystreams.Parse(feedData)
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

		for _, e := range feed.Activities {
			if err := a.saveActivity(&e); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		rw.Header().Set("content-type", "application/json")
		if err := json.NewEncoder(rw).Encode(feed); err != nil {
			panic(err)
		}
	})

	m.Methods("GET").Path("/find-feed").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		user := r.URL.Query().Get("user")
		if !strings.HasPrefix(user, "acct:") {
			user = "acct:" + user
		}

		acct, err := acct.FromString(user)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		wf, err := webfinger.Fetch(webfinger.MakeURL(acct.Host, acct.String(), nil))
		if err != nil {
			hm, err := hostmeta.Fetch(acct.Host)
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

			wf, err = webfinger.Fetch(lrddHref)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
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
