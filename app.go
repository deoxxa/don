package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"github.com/timewasted/go-accept-headers"
)

type App struct {
	DB       *sql.DB
	Store    sessions.Store
	Renderer ReactRenderer
	Template *template.Template
	BuildBox *rice.Box
}

func NewApp(db *sql.DB, store sessions.Store, renderer ReactRenderer, template *template.Template, buildBox *rice.Box) (*App, error) {
	return &App{DB: db, Store: store, Renderer: renderer, Template: template, BuildBox: buildBox}, nil
}

func (a *App) OnMessage(id string, s *PubSubSubscription, rd io.ReadCloser) {
	var v AtomFeed

	if *recordDocuments {
		d, err := ioutil.ReadAll(rd)
		if err != nil {
			logrus.WithField("id", id).WithError(err).Debug("pubsub: couldn't read message")
			return
		}

		if err := xml.NewDecoder(bytes.NewReader(d)).Decode(&v); err != nil {
			logrus.WithField("id", id).WithError(err).Debug("pubsub: couldn't parse body")
			return
		}

		if _, err := a.DB.Exec("insert into documents (created_at, xml) values ($1, $2)", time.Now(), string(d)); err != nil {
			logrus.WithField("id", id).WithError(err).Debug("pubsub: couldn't save document")
			return
		}
	} else {
		if err := xml.NewDecoder(rd).Decode(&v); err != nil {
			logrus.WithField("id", id).WithError(err).Debug("pubsub: couldn't parse body")
			return
		}
	}

	if s == nil {
		logrus.WithField("id", id).Debug("pubsub: unsolicited message")
		return
	}

	l := logrus.WithFields(logrus.Fields{
		"id":    s.ID,
		"hub":   s.Hub,
		"topic": s.Topic,
	})

	if v.Author != nil {
		if err := savePerson(a.DB, s.Topic, v.Author); err != nil {
			l.WithError(err).Debug("pubsub: couldn't save author")
			return
		}
	}

	for _, e := range v.Entry {
		if err := saveEntry(a.DB, s.Topic, &e); err != nil {
			l.WithError(err).Debug("pubsub: couldn't save entry")
		} else {
			l.Debug("pubsub: saved entry")
		}
	}
}

func (a *App) getSessionAndUserFromRequest(r *http.Request) (*sessions.Session, *User, error) {
	s, err := a.Store.Get(r, "login")
	if err != nil {
		return nil, nil, errors.Wrap(err, "App.getSessionAndUserFromRequest")
	}

	userIDValue, ok := s.Values["user_id"]
	if !ok {
		return s, nil, nil
	}

	userID, ok := userIDValue.(string)
	if !ok {
		return s, nil, errors.Errorf("App.getSessionAndUserFromRequest: invalid type %T for user_id", userIDValue)
	}

	var u User
	if err := a.DB.QueryRow("select id, created_at, username, email, display_name, avatar from users where id = $1", userID).Scan(&u.ID, &u.CreatedAt, &u.Username, &u.Email, &u.DisplayName, &u.Avatar); err != nil {
		if err == sql.ErrNoRows {
			return s, nil, nil
		}

		return s, nil, errors.Wrap(err, "App.getSessionAndUserFromRequest")
	}

	return s, &u, nil
}

type AppHandlerFunc func(r *http.Request, ar *AppResponse) *AppResponse

func (a *App) HandlerFor(fn AppHandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		ar, err := a.StandardContext(rw, r)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := a.SendResponse(rw, r, fn(r, ar)); err != nil {
			logrus.WithError(err).Warn("error sending response")
		}
	}
}

func (a *App) StandardContext(rw http.ResponseWriter, r *http.Request) (*AppResponse, error) {
	s, u, err := a.getSessionAndUserFromRequest(r)
	if err != nil {
		return nil, errors.Wrap(err, "App.StandardContext")
	}

	return NewAppResponse().WithSession(s).WithUser(u).ShallowMergeState(map[string]interface{}{
		"authentication": map[string]interface{}{
			"loading": false,
			"error":   nil,
			"user":    u,
		},
	}), nil
}

func (a *App) SendResponse(rw http.ResponseWriter, r *http.Request, ar *AppResponse) error {
	acceptable := accept.Parse(r.Header.Get("accept"))

	ct, err := acceptable.Negotiate("text/html", "application/json")
	if err != nil {
		ar = ar.WithError(err)
	}

	if ar.Session != nil {
		if ar.User == nil {
			delete(ar.Session.Values, "user_id")
		} else {
			ar.Session.Values["user_id"] = ar.User.ID
		}

		if err := ar.Session.Save(r, rw); err != nil {
			ar = ar.WithError(err)
		}
	}

	if ar.Error != nil {
		ar = ar.ShallowMergeState(map[string]interface{}{
			"server": map[string]interface{}{
				"error": ar.Error.Error(),
			},
		})
	} else {
		ar = ar.ShallowMergeState(map[string]interface{}{
			"server": map[string]interface{}{
				"error": nil,
			},
		})
	}

	if ar.Error != nil && ct == "application/json" {
		status := http.StatusSeeOther
		if ar.Status != 0 {
			status = ar.Status
		}

		http.Error(rw, ar.Error.Error(), status)
		return nil
	}

	if ar.Redirect != "" && (ct == "text/html" || ct == "") {
		status := http.StatusSeeOther
		if ar.Status != 0 {
			status = ar.Status
		}

		http.Redirect(rw, r, ar.Redirect, status)
		return nil
	}

	d, err := json.Marshal(ar.State)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return errors.Wrap(err, "App.render")
	}

	switch ct {
	case "application/json":
		rw.Header().Set("content-type", "application/json; charset=utf8")
		if ar.Status != 0 {
			rw.WriteHeader(ar.Status)
		}
		if _, err := io.Copy(rw, bytes.NewReader(d)); err != nil {
			return errors.Wrap(err, "App.render")
		}
	case "text/html", "":
		html, err := a.Renderer.Render(a.BuildBox.MustString("entry-server-bundle.js"), r.URL.String(), string(d))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return errors.Wrap(err, "App.render")
		}

		data := map[string]interface{}{
			"HTML":     template.HTML(html),
			"JSON":     template.JS(d),
			"CSSFiles": []string{"/build/vendor-styles.css", "/build/entry-client-styles.css"},
			"JSFiles":  []string{"/build/vendor-bundle.js", "/build/entry-client-bundle.js"},
			"Meta": map[string]interface{}{
				"Title":       ar.GetMeta("title", "DON"),
				"Description": ar.GetMeta("description", "A very basic StatusNet node. Kind of like Mastodon, but worse."),
			},
		}

		if *externalJS != "" {
			data["CSSFiles"] = []string{}
			data["JSFiles"] = []string{*externalJS}
		}

		rw.Header().Set("content-type", "text/html; charset=utf-8")
		if ar.Status != 0 {
			rw.WriteHeader(ar.Status)
		}
		if a.Template.Execute(rw, data); err != nil {
			return errors.Wrap(err, "App.render")
		}
	}

	return nil
}
