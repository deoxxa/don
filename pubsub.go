package main

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/tomnomnom/linkheader"
)

func PubSubMakeCallbackURL(baseURL, id string) string {
	return baseURL + "/" + id
}

type PubSubSubscription struct {
	ID          string
	Hub         string
	Topic       string
	CallbackURL string
	ExpiresAt   *time.Time
}

func (s *PubSubSubscription) Remaining(t time.Time) time.Duration {
	if s.ExpiresAt == nil {
		return 0
	}

	if s.ExpiresAt.Before(t) {
		return 0
	}

	return s.ExpiresAt.Sub(t)
}

type PubSubState interface {
	All() (subscriptions []PubSubSubscription, err error)
	Add(hub, topic, baseURL string) (id, callbackURL string, existed bool, err error)
	Get(hub, topic string) (subscription *PubSubSubscription, err error)
	GetByID(id string) (subscription *PubSubSubscription, err error)
	Set(hub, topic string, expiresAt time.Time) (err error)
	Del(hub, topic string) (err error)
}

type PubSubMessageHandler func(id string, s *PubSubSubscription, rd io.ReadCloser)

type PubSubClient struct {
	CallbackURL string
	State       PubSubState
	OnMessage   PubSubMessageHandler
}

func NewPubSubClient(callbackURL string, state PubSubState, onMessage PubSubMessageHandler) *PubSubClient {
	return &PubSubClient{
		CallbackURL: callbackURL,
		State:       state,
		OnMessage:   onMessage,
	}
}

func (c *PubSubClient) Refresh(forceUpdate bool, interval time.Duration) error {
	a, err := c.State.All()
	if err != nil {
		return errors.Wrap(err, "PubSubClient.Refresh")
	}

	var g WorkerGroup

	for _, e := range a {
		e := e

		callbackURL := c.CallbackURL + "/" + e.ID

		if forceUpdate || e.Remaining(time.Now()) < interval || e.CallbackURL != callbackURL {
			l := logrus.WithFields(logrus.Fields{
				"id":               e.ID,
				"hub":              e.Hub,
				"topic":            e.Topic,
				"callback_url":     e.CallbackURL,
				"new_callback_url": callbackURL,
				"expires_at":       e.ExpiresAt,
				"force_update":     forceUpdate,
			})

			l.Debug("pubsub: refreshing subscription")

			g.Add(func() error {
				if err := PubSubSubscribe(e.Hub, e.Topic, callbackURL); err != nil {
					l.WithError(err).Warn("pubsub: couldn't subscribe to topic")
					return errors.Wrap(err, "PubSubClient.RefreshWorker")
				}

				l.Debug("pubsub: subscribed successfully")
				return nil
			})
		}
	}

	return errors.Wrap(g.Run(4), "PubSubClient.Refresh")
}

func (c *PubSubClient) Subscribe(hub, topic string) error {
	_, callbackURL, existed, err := c.State.Add(hub, topic, c.CallbackURL)
	if err != nil {
		return errors.Wrap(err, "PubSubClient.Subscribe")
	}

	if !existed {
		if err := PubSubSubscribe(hub, topic, callbackURL); err != nil {
			return errors.Wrap(err, "PubSubClient.Subscribe")
		}
	}

	return nil
}

func (c *PubSubClient) Unsubscribe(hub, topic string) error {
	s, err := c.State.Get(hub, topic)
	if err != nil {
		return errors.Wrap(err, "PubSubClient.Unsubscribe")
	}

	if s != nil {
		if err := PubSubUnsubscribe(hub, topic, s.CallbackURL); err != nil {
			return errors.Wrap(err, "PubSubClient.Unsubscribe")
		}

		if err := c.State.Del(s.Hub, s.Topic); err != nil {
			return errors.Wrap(err, "PubSubClient.Unsubscribe")
		}
	}

	return nil
}

func (c *PubSubClient) Handler() *PubSubHandler {
	return &PubSubHandler{
		OnChallenge: func(id, topic, mode string, leaseTime time.Duration) error {
			logrus.WithFields(logrus.Fields{
				"id":         id,
				"topic":      topic,
				"mode":       mode,
				"lease_time": leaseTime,
			}).Debug("pubsub: received challenge")

			s, err := c.State.GetByID(id)
			if err != nil {
				return err
			}

			if s == nil {
				return errors.Errorf("PubSubClient.Handler: subscription not found")
			}

			return errors.Wrap(c.State.Set(s.Hub, s.Topic, time.Now().Add(leaseTime)), "PubSubClient.Handler")
		},
		OnMessage: func(id, topic string, rd io.ReadCloser) {
			logrus.WithFields(logrus.Fields{
				"id":    id,
				"topic": topic,
			}).Debug("pubsub: received message")

			defer rd.Close()

			if c.OnMessage != nil {
				s, err := c.State.GetByID(id)
				if err != nil {
					return
				}

				c.OnMessage(id, s, rd)
			}
		},
	}
}

func PubSubAlter(hub, topic, callbackURL, mode string) error {
	res, err := http.PostForm(hub, url.Values{
		"hub.callback":      []string{callbackURL},
		"hub.mode":          []string{mode},
		"hub.topic":         []string{topic},
		"hub.verify":        []string{"async"},
		"hub.lease_seconds": []string{"604800"},
	})

	if err != nil {
		return errors.Wrap(err, "PubSubAlter")
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return errors.Errorf("PubSubAlter: invalid status code; expected 2xx but got %d", res.StatusCode)
	}

	return nil
}

func PubSubSubscribe(hub, topic, callbackURL string) error {
	return errors.Wrap(PubSubAlter(hub, topic, callbackURL, "subscribe"), "PubSubSubscribe")
}

func PubSubUnsubscribe(hub, topic, callbackURL string) error {
	return errors.Wrap(PubSubAlter(hub, topic, callbackURL, "unsubscribe"), "PubSubSubscribe")
}

type PubSubHandler struct {
	OnChallenge func(id, topic, mode string, leaseTime time.Duration) error
	OnMessage   func(id, topic string, rd io.ReadCloser)
}

func (h *PubSubHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	bits := strings.Split(r.URL.Path, "/")
	id := bits[len(bits)-1]

	if r.Method == http.MethodGet && q.Get("hub.challenge") != "" {
		if h.OnChallenge != nil {
			leaseTime := time.Hour * 12
			if n, err := strconv.ParseInt(q.Get("hub.lease_seconds"), 10, 64); err == nil {
				leaseTime = time.Second * time.Duration(n)
			}

			if err := h.OnChallenge(id, q.Get("hub.topic"), q.Get("hub.mode"), leaseTime); err != nil {
				rw.WriteHeader(http.StatusNotFound)
				return
			}
		}

		rw.Write([]byte(r.URL.Query().Get("hub.challenge")))
		return
	}

	if r.Method != http.MethodPost {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	links := linkheader.ParseMultiple(r.Header["Link"])

	var topic string
	if a := links.FilterByRel("self"); len(a) > 0 {
		topic = a[0].URL
	}

	rw.WriteHeader(http.StatusAccepted)

	if h.OnMessage != nil {
		h.OnMessage(id, topic, r.Body)
	}
}
