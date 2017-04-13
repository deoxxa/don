package main

import (
	"net/http"

	"github.com/Sirupsen/logrus"
)

func (a *App) handleLogoutGet(rw http.ResponseWriter, r *http.Request) {
	s, u, err := a.getSessionAndUserFromRequest(r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.Save(r, rw); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := a.render(rw, r, "Logout - DON", "Log out of DON", map[string]interface{}{
		"authentication": map[string]interface{}{
			"loading": false,
			"error":   nil,
			"user":    u,
		},
	}); err != nil {
		logrus.WithError(err).Warn("handler: error sending response")
	}
}

func (a *App) handleLogoutPost(rw http.ResponseWriter, r *http.Request) {
	s, _, err := a.getSessionAndUserFromRequest(r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	s.Options.MaxAge = -1

	if err := s.Save(r, rw); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(rw, r, "/", http.StatusSeeOther)
}
