package main // import "fknsrs.biz/p/don"

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/dyninc/qstring"
)

func (a *App) handleHomeGet(rw http.ResponseWriter, r *http.Request) {
	var args getPublicTimelineArgs
	if err := qstring.Unmarshal(r.URL.Query(), &args); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	s, u, err := a.getSessionAndUserFromRequest(r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	posts, err := getPublicTimeline(a.DB, args)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.Save(r, rw); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := a.render(rw, r, "Home - DON", "", map[string]interface{}{
		"authentication": map[string]interface{}{
			"loading": false,
			"error":   nil,
			"user":    u,
		},
		"publicTimeline": map[string]interface{}{
			"loading": false,
			"posts":   posts,
			"error":   nil,
		},
	}); err != nil {
		logrus.WithError(err).Warn("handler: error sending response")
	}
}
