package main

import (
	"bytes"
	"database/sql"
	"encoding/xml"
	"io"
	"io/ioutil"
	"time"

	"github.com/Sirupsen/logrus"
)

type App struct {
	DB *sql.DB
}

func NewApp(db *sql.DB) (*App, error) {
	return &App{DB: db}, nil
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
