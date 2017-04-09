package main

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/juju/ratelimit"
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
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ExpiresAt   *time.Time
}

type PubSubState interface {
	All() (subscriptions []PubSubSubscription, err error)
	Add(hub, topic, baseURL string) (subscription *PubSubSubscription, oldCallbackURL string, err error)
	Get(hub, topic string) (subscription *PubSubSubscription, err error)
	GetByID(id string) (subscription *PubSubSubscription, err error)
	Set(hub, topic string, updatedAt, expiresAt time.Time) (err error)
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

	logrus.WithField("count", len(a)).Debug("pubsub: got subscriptions to refresh")

	m := make(map[string]*ratelimit.Bucket)
	var l sync.RWMutex

	getBucket := func(host string) *ratelimit.Bucket {
		l.RLock()
		if b, ok := m[host]; ok {
			l.RUnlock()
			return b
		}
		l.RUnlock()

		l.Lock()
		if b, ok := m[host]; ok {
			l.Unlock()
			return b
		}
		defer l.Unlock()

		m[host] = ratelimit.NewBucket(time.Second*30, 4)

		return m[host]
	}

	var g WorkerGroup

	for _, e := range a {
		e := e

		callbackURL := c.CallbackURL + "/" + e.ID

		l := logrus.WithFields(logrus.Fields{
			"id":               e.ID,
			"hub":              e.Hub,
			"topic":            e.Topic,
			"callback_url":     e.CallbackURL,
			"new_callback_url": callbackURL,
			"created_at":       e.CreatedAt,
			"updated_at":       e.UpdatedAt,
			"expires_at":       e.ExpiresAt,
			"force_update":     forceUpdate,
		})

		if e.ExpiresAt == nil {
			l.Debug("pubsub: not refreshing subscription which has never been verified")
			continue
		}

		if forceUpdate || e.ExpiresAt.Sub(time.Now()) < interval || e.CallbackURL != callbackURL {
			u, err := url.Parse(e.Hub)
			if err != nil {
				l.WithError(err).Warn("pubsub: couldn't parse hub url")
				return errors.Wrap(err, "PubSubClient.RefreshWorker")
			}

			dur, ok := getBucket(u.Host).TakeMaxDuration(1, *pubsubRefreshInterval)
			if !ok {
				l.Debug("pubsub: skipping renewing for now as we'd have to wait too long")
				continue
			}

			g.Add(func() error {
				if dur > 0 {
					l.WithField("duration", dur).Debug("pubsub: waiting so as not to overwhelm the endpoint")
					time.Sleep(dur)
				}

				l.Debug("pubsub: refreshing subscription")

				// simulate subscription
				time.Sleep(time.Second * 5)

				// if err := c.Subscribe(e.Hub, e.Topic); err != nil {
				// 	l.WithError(err).Warn("pubsub: couldn't subscribe to topic")
				// 	return errors.Wrap(err, "PubSubClient.RefreshWorker")
				// }

				l.Debug("pubsub: subscribed successfully")
				return nil
			})
		}
	}

	return errors.Wrap(g.Run(20), "PubSubClient.Refresh")
}

func (c *PubSubClient) Subscribe(hub, topic string) error {
	logrus.WithFields(logrus.Fields{"hub": hub, "topic": topic}).Debug("pubsub: subscribing")

	s, oldCallbackURL, err := c.State.Add(hub, topic, c.CallbackURL)
	if err != nil {
		return errors.Wrap(err, "PubSubClient.Subscribe")
	}

	if oldCallbackURL != s.CallbackURL {
		if oldCallbackURL != "" {
			if err := PubSubUnsubscribe(hub, topic, oldCallbackURL); err != nil {
				return errors.Wrap(err, "PubSubClient.Subscribe")
			}
		}

		if err := PubSubSubscribe(hub, topic, s.CallbackURL); err != nil {
			return errors.Wrap(err, "PubSubClient.Subscribe")
		}
	}

	return nil
}

func (c *PubSubClient) Unsubscribe(hub, topic string) error {
	logrus.WithFields(logrus.Fields{"hub": hub, "topic": topic}).Debug("pubsub: unsubscribing")

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
			l := logrus.WithFields(logrus.Fields{
				"id":         id,
				"topic":      topic,
				"mode":       mode,
				"lease_time": leaseTime,
			})

			l.Debug("pubsub: received challenge")

			s, err := c.State.GetByID(id)
			if err != nil {
				l.WithError(err).Warn("pubsub: error fetching subscription during challenge")
				return err
			}

			if s == nil {
				l.WithError(err).Warn("pubsub: subscription not found during challenge")
				return errors.Errorf("PubSubClient.Handler: subscription not found")
			}

			return errors.Wrap(c.State.Set(s.Hub, s.Topic, time.Now(), time.Now().Add(leaseTime)), "PubSubClient.Handler")
		},
		OnMessage: func(id, topic string, rd io.ReadCloser) {
			l := logrus.WithFields(logrus.Fields{
				"id":    id,
				"topic": topic,
			})

			l.Debug("pubsub: received message")

			defer rd.Close()

			if c.OnMessage != nil {
				s, err := c.State.GetByID(id)
				if err != nil {
					l.WithError(err).Warn("pubsub: error fetching subscription during reception")
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
