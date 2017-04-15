package acct

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

type URL struct {
	User string `json:"user"`
	Host string `json:"host"`
}

func (a *URL) URL() *url.URL {
	return &url.URL{
		Scheme: "acct",
		Opaque: a.User + "@" + a.Host,
	}
}

func (a *URL) String() string {
	return a.URL().String()
}

func FromString(s string) (*URL, error) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "@")

	if !strings.HasPrefix(s, "acct:") {
		s = "acct:" + s
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, errors.Wrap(err, "FromString")
	}

	if u.Scheme != "acct" {
		return nil, errors.Errorf("FromString: invalid scheme; expected acct but got %q", u.Scheme)
	}

	bits := strings.Split(u.Opaque, "@")
	if len(bits) != 2 {
		return nil, errors.Errorf("FromString: expected two strings separated by @ but got %d", len(bits))
	}

	return &URL{User: bits[0], Host: bits[1]}, nil
}
