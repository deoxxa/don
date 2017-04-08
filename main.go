package main // import "fknsrs.biz/p/don"

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/jtacoma/uritemplates"
	_ "github.com/mattn/go-sqlite3"
	"github.com/meatballhat/negroni-logrus"
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
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

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

	db, err := sql.Open("sqlite3", *database)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		create table if not exists pubsub_state (id text not null primary key, hub text not null, topic text not null, callback_url text not null, expires_at datetime, unique (hub, topic));
	  create table if not exists people (feed_url text not null primary key, first_seen datetime not null, name text, display_name text, email text, summary text, note text);
		create table if not exists posts (feed_url text not null, id text not null, created_at datetime not null, raw_entry text not null, primary key (feed_url, id));
	`); err != nil {
		panic(err)
	}

	savePerson := func(feedURL string, author *AtomAuthor) error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		var name, displayName, email, summary, note string
		if err := tx.QueryRow("select name, display_name, email, summary, note from people where feed_url = $1", feedURL).Scan(&name, &displayName, &email, &summary, &note); err != nil {
			if err != sql.ErrNoRows {
				return err
			}

			if _, err := tx.Exec("insert into people (feed_url, first_seen, name, display_name, email, summary, note) values ($1, $2, $3, $4, $5, $6, $7)", feedURL, time.Now(), author.Name, author.DisplayName, author.Email, author.Summary, author.Note); err != nil {
				return err
			}
		} else {
			if name == author.Name && displayName == author.DisplayName && email == author.Email && summary == author.Summary && note == author.Note {
				return nil
			}

			if _, err := tx.Exec("update people set name = $1, display_name = $2, email = $3, summary = $4, note = $5 where feed_url = $6", author.Name, author.DisplayName, author.Email, author.Summary, author.Note, feedURL); err != nil {
				return err
			}
		}

		return tx.Commit()
	}

	saveEntry := func(feedURL string, entry *AtomEntry) error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		var exists int
		if err := tx.QueryRow("select count(1) from posts where feed_url = $1 and id = $2", feedURL, entry.ID).Scan(&exists); err != nil {
			return err
		}

		if exists > 0 {
			return nil
		}

		d, err := json.Marshal(entry)
		if err != nil {
			return err
		}

		if _, err := tx.Exec("insert into posts (feed_url, id, created_at, raw_entry) values ($1, $2, $3, $4)", feedURL, entry.ID, entry.Published, string(d)); err != nil {
			return err
		}

		return tx.Commit()
	}

	psc := NewPubSubClient(*publicURL+"/pubsub", NewPubSubSQLState(db), func(id string, s *PubSubSubscription, rd io.ReadCloser) {
		if s == nil {
			logrus.WithField("id", id).Debug("pubsub: unsolicited message")
			return
		}

		l := logrus.WithFields(logrus.Fields{
			"id":    s.ID,
			"hub":   s.Hub,
			"topic": s.Topic,
		})

		l.Debug("pubsub: received message")

		var v AtomFeed
		if err := xml.NewDecoder(rd).Decode(&v); err != nil {
			l.WithError(err).Debug("pubsub: couldn't parse body")
			return
		}

		l.Debug("pubsub: parsed message")

		if v.Author != nil {
			if err := savePerson(s.Topic, v.Author); err != nil {
				l.WithError(err).Debug("pubsub: couldn't save author")
				return
			}
		}

		for _, e := range v.Entry {
			if err := saveEntry(s.Topic, &e); err != nil {
				l.WithError(err).Debug("pubsub: couldn't save entry")
			} else {
				l.Debug("pubsub: saved entry")
			}
		}
	})

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

	m.PathPrefix("/pubsub").Handler(psc.Handler())

	m.Methods("GET").Path("/").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("select posts.feed_url, posts.raw_entry, people.name, people.display_name, people.email from posts left join people on people.feed_url = posts.feed_url order by posts.created_at desc limit 25")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var posts []AtomFeed
		for rows.Next() {
			var feedURL, rawEntry string
			var name, displayName, email sql.NullString
			if err := rows.Scan(&feedURL, &rawEntry, &name, &displayName, &email); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

			var entry AtomEntry
			if err := json.Unmarshal([]byte(rawEntry), &entry); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

			post := AtomFeed{
				Entry: []AtomEntry{entry},
				Link:  []AtomLink{AtomLink{Rel: "self", Href: feedURL}},
			}

			if name.Valid {
				post.Author = &AtomAuthor{
					Name:        name.String,
					DisplayName: displayName.String,
					Email:       email.String,
				}
			}

			posts = append(posts, post)
		}

		rw.Header().Set("content-type", "text/html")
		if err := templateHome.Execute(rw, map[string]interface{}{"Posts": posts}); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	m.Methods("GET").Path("/health").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	m.Methods("GET").Path("/show-feed").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		feed, err := AtomFetch(r.URL.Query().Get("url"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		hubLink := feed.GetLink("hub")
		selfLink := feed.GetLink("self")

		if hubLink != nil && selfLink != nil && hubLink.Href != "" && selfLink.Href != "" {
			if err := psc.Subscribe(hubLink.Href, selfLink.Href); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		rw.Header().Set("content-type", "text/html")
		rw.WriteHeader(http.StatusOK)

		if feed.Author != nil {
			if err := savePerson(r.URL.Query().Get("url"), feed.Author); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		for _, e := range feed.Entry {
			if err := saveEntry(r.URL.Query().Get("url"), &e); err != nil {
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
	n.UseHandler(m)

	logrus.Info("starting server")

	if err := http.ListenAndServe(*addr, n); err != nil {
		panic(err)
	}
}
