package main // import "fknsrs.biz/p/don"

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"
	"gopkg.in/hlandau/passlib.v1"
)

func (a *App) handleRegisterGet(rw http.ResponseWriter, r *http.Request) {
	s, u, err := a.getSessionAndUserFromRequest(r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.Save(r, rw); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := a.render(rw, r, "Register - DON", "Register a new account.", map[string]interface{}{
		"authentication": map[string]interface{}{
			"loading": false,
			"error":   nil,
			"user":    u,
		},
	}); err != nil {
		logrus.WithError(err).Warn("handler: error sending response")
	}
}

func (a *App) handleRegisterPost(rw http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	username := r.Form.Get("username")
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	returnTo := r.Form.Get("return_to")

	var id string
	if err := a.DB.QueryRow("select id from users where username = $1 or email = $2", username, email).Scan(&id); err == nil {
		http.Error(rw, "A user with that username or email already exists.", http.StatusConflict)
		return
	} else if err != sql.ErrNoRows {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	hash, err := passlib.Hash(password)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	u := User{
		ID:        uuid.NewV4().String(),
		CreatedAt: time.Now(),
		Username:  username,
		Email:     email,
	}

	if _, err := a.DB.Exec("insert into users (id, created_at, username, email, hash) values ($1, $2, $3, $4, $5)", u.ID, u.CreatedAt, u.Username, u.Email, hash); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
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

	if strings.HasPrefix(returnTo, "/") {
		http.Redirect(rw, r, returnTo, http.StatusSeeOther)
		return
	}

	rw.Header().Set("content-type", "application/json; charset=utf8")
	rw.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(rw).Encode(map[string]interface{}{
		"authentication": map[string]interface{}{
			"loading": false,
			"error":   nil,
			"user":    u,
		},
	}); err != nil {
		logrus.WithError(err).Warn("handler: error sending response")
	}
}
