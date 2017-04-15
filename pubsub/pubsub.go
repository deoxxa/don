package pubsub

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

	"fknsrs.biz/p/don/workergroup"
)

func MakeCallbackURL(baseURL, id string) string {
	return baseURL + "/" + id
}

type Subscription struct {
	ID          string
	Hub         string
	Topic       string
	CallbackURL string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ExpiresAt   *time.Time
}

type State interface {
	All() (subscriptions []Subscription, err error)
	Add(hub, topic, baseURL string) (subscription *Subscription, oldCallbackURL string, err error)
	Get(hub, topic string) (subscription *Subscription, err error)
	GetByID(id string) (subscription *Subscription, err error)
	Set(hub, topic string, updatedAt, expiresAt time.Time) (err error)
	Del(hub, topic string) (err error)
}

type MessageHandler func(id string, s *Subscription, rd io.ReadCloser)

type Client struct {
	CallbackURL string
	State       State
	OnMessage   MessageHandler
}

func NewClient(callbackURL string, state State, onMessage MessageHandler) *Client {
	return &Client{
		CallbackURL: callbackURL,
		State:       state,
		OnMessage:   onMessage,
	}
}

func (c *Client) Refresh(forceUpdate bool, interval time.Duration) error {
	a, err := c.State.All()
	if err != nil {
		return errors.Wrap(err, "Client.Refresh")
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

	var g workergroup.Group

	n := 0

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
				return errors.Wrap(err, "Client.RefreshWorker")
			}

			dur, ok := getBucket(u.Host).TakeMaxDuration(1, interval)
			if !ok {
				l.Debug("pubsub: skipping renewing for now as we'd have to wait too long")
				continue
			}

			n++

			g.Add(func() error {
				if dur > 0 {
					l.WithField("duration", dur).Debug("pubsub: waiting so as not to overwhelm the endpoint")
					time.Sleep(dur)
				}

				l.Debug("pubsub: refreshing subscription")

				if err := c.Subscribe(e.Hub, e.Topic); err != nil {
					l.WithError(err).Warn("pubsub: couldn't subscribe to topic")
					return errors.Wrap(err, "Client.RefreshWorker")
				}

				l.Debug("pubsub: subscribed successfully")
				return nil
			})
		}
	}

	return errors.Wrap(g.Run(n), "Client.Refresh")
}

func (c *Client) Subscribe(hub, topic string) error {
	logrus.WithFields(logrus.Fields{"hub": hub, "topic": topic}).Debug("pubsub: subscribing")

	s, oldCallbackURL, err := c.State.Add(hub, topic, c.CallbackURL)
	if err != nil {
		return errors.Wrap(err, "Client.Subscribe")
	}

	if oldCallbackURL != s.CallbackURL {
		if oldCallbackURL != "" {
			if err := Unsubscribe(hub, topic, oldCallbackURL); err != nil {
				return errors.Wrap(err, "Client.Subscribe")
			}
		}

		if err := Subscribe(hub, topic, s.CallbackURL); err != nil {
			return errors.Wrap(err, "Client.Subscribe")
		}
	}

	return nil
}

func (c *Client) Unsubscribe(hub, topic string) error {
	logrus.WithFields(logrus.Fields{"hub": hub, "topic": topic}).Debug("pubsub: unsubscribing")

	s, err := c.State.Get(hub, topic)
	if err != nil {
		return errors.Wrap(err, "Client.Unsubscribe")
	}

	if s != nil {
		if err := Unsubscribe(hub, topic, s.CallbackURL); err != nil {
			return errors.Wrap(err, "Client.Unsubscribe")
		}

		if err := c.State.Del(s.Hub, s.Topic); err != nil {
			return errors.Wrap(err, "Client.Unsubscribe")
		}
	}

	return nil
}

func (c *Client) Handler() *Handler {
	return &Handler{
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
				return errors.Wrap(err, "Client.Handler")
			}

			if s == nil {
				l.WithError(err).Warn("pubsub: subscription not found during challenge")
				return errors.Errorf("Client.Handler: subscription not found")
			}

			return errors.Wrap(c.State.Set(s.Hub, s.Topic, time.Now(), time.Now().Add(leaseTime)), "Client.Handler")
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

func Alter(hub, topic, callbackURL, mode string) error {
	f := url.Values{
		"hub.callback":      []string{callbackURL},
		"hub.mode":          []string{mode},
		"hub.topic":         []string{topic},
		"hub.verify":        []string{"async"},
		"hub.lease_seconds": []string{"604800"},
	}

	res, err := http.PostForm(hub, f)
	if err != nil {
		return errors.Wrap(err, "Alter")
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return errors.Errorf("Alter: invalid status code; expected 2xx but got %d", res.StatusCode)
	}

	return nil
}

func Subscribe(hub, topic, callbackURL string) error {
	return errors.Wrap(Alter(hub, topic, callbackURL, "subscribe"), "Subscribe")
}

func Unsubscribe(hub, topic, callbackURL string) error {
	return errors.Wrap(Alter(hub, topic, callbackURL, "unsubscribe"), "Subscribe")
}

type Handler struct {
	OnChallenge func(id, topic, mode string, leaseTime time.Duration) error
	OnMessage   func(id, topic string, rd io.ReadCloser)
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
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
