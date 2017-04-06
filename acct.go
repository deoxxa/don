package main

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

type Acct struct {
	User string `json:"user"`
	Host string `json:"host"`
}

func (a *Acct) URL() *url.URL {
	return &url.URL{
		Scheme: "acct",
		Opaque: a.User + "@" + a.Host,
	}
}

func (a *Acct) String() string {
	return a.URL().String()
}

func AcctFromString(s string) (*Acct, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, errors.Wrap(err, "AcctFromString")
	}

	if u.Scheme != "acct" {
		return nil, errors.Errorf("AcctFromString: invalid scheme; expected acct but got %q", u.Scheme)
	}

	bits := strings.Split(u.Opaque, "@")
	if len(bits) != 2 {
		return nil, errors.Errorf("AcctFromString: expected two strings separated by @ but got %d", len(bits))
	}

	return &Acct{User: bits[0], Host: bits[1]}, nil
}
