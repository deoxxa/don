package main // import "fknsrs.biz/p/don"

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/Sirupsen/logrus"
	"gopkg.in/hlandau/passlib.v1"
)

func (a *App) handleLoginGet(rw http.ResponseWriter, r *http.Request) {
	s, u, err := a.getSessionAndUserFromRequest(r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.Save(r, rw); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := a.render(rw, r, "Login - DON", "Log into DON", map[string]interface{}{
		"authentication": map[string]interface{}{
			"loading": false,
			"error":   nil,
			"user":    u,
		},
	}); err != nil {
		logrus.WithError(err).Warn("handler: error sending response")
	}
}

func (a *App) handleLoginPost(rw http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	username := r.Form.Get("username")
	password := r.Form.Get("password")
	returnTo := r.Form.Get("return_to")

	var u User
	var hash string

	err := a.DB.QueryRow("select id, created_at, username, email, hash, display_name, avatar from users where username = $1 or email = $1", username).Scan(&u.ID, &u.CreatedAt, &u.Username, &u.Email, &hash, &u.DisplayName, &u.Avatar)

	switch err {
	case sql.ErrNoRows:
		if returnTo != "" {
			http.Redirect(rw, r, "/login?"+url.Values{"return_to": []string{returnTo}}.Encode(), http.StatusSeeOther)
			return
		}

		rw.WriteHeader(http.StatusUnauthorized)
		return
	case nil:
		newHash, err := passlib.Verify(password, hash)
		if err != nil {
			http.Redirect(rw, r, "/login?"+url.Values{"return_to": []string{returnTo}}.Encode(), http.StatusSeeOther)
			return
		}

		if newHash != "" {
			if _, err := a.DB.Exec("update users set hash = $1 where id = $2 and hash = $3", newHash, u.ID, hash); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		s, _, err := a.getSessionAndUserFromRequest(r)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		s.Values["user_id"] = u.ID

		if err := s.Save(r, rw); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if returnTo != "" {
			if strings.HasPrefix(returnTo, "/") {
				http.Redirect(rw, r, returnTo, http.StatusSeeOther)
				return
			}

			http.Redirect(rw, r, "/", http.StatusSeeOther)
			return
		}

		rw.Header().Set("content-type", "application/json; charset=utf8")
		rw.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(rw).Encode(map[string]interface{}{"user": u}); err != nil {
			logrus.WithError(err).Warn("handler: error sending response")
		}
	default:
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}
