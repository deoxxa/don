package main

import (
	"database/sql"
	"net/http"

	"github.com/pkg/errors"
	"gopkg.in/hlandau/passlib.v1"
)

func (a *App) handleLoginGet(r *http.Request, ar *AppResponse) *AppResponse {
	return ar.MergeMeta(map[string]string{
		"Title":       "Log in",
		"Description": "Log in to continue.",
	})
}

func (a *App) handleLoginPost(r *http.Request, ar *AppResponse) *AppResponse {
	ar = ar.MergeMeta(map[string]string{
		"Title":       "Log in",
		"Description": "Log in to continue.",
	})

	if err := r.ParseForm(); err != nil {
		return ar.WithStatus(http.StatusBadRequest).WithError(errors.Wrap(err, "App.handleLoginPost"))
	}

	var v struct {
		Username string `schema:"username"`
		Password string `schema:"password"`
	}

	if err := decoder.Decode(&v, r.PostForm); err != nil {
		return ar.WithError(errors.Wrap(err, "App.handleLoginPost"))
	}

	u, err := a.userLogin(v.Username, v.Password)

	if err != nil {
		ar = ar.WithUser(nil).WithError(errors.Wrap(err, "App.handleRegisterPost")).ShallowMergeState(map[string]interface{}{
			"authentication": map[string]interface{}{
				"loading": false,
				"error":   err.Error(),
				"user":    nil,
			},
		}).WithRedirect(r.URL.String())

		switch errors.Cause(err) {
		case errLoginUserNotFound:
			return ar.WithStatus(http.StatusUnauthorized)
		default:
			return ar
		}
	}

	return ar.MergeUserContext(u).WithRedirect(r.URL.Query().Get("return_to"))
}

var (
	errLoginUserNotFound = errors.New("no user with that username/password could be found")
)

func (a *App) userLogin(username, password string) (*User, error) {
	var u User
	var hash string

	if err := a.SQLDB.QueryRow("select id, created_at, username, email, hash, display_name, avatar from users where username = $1", username).Scan(&u.ID, &u.CreatedAt, &u.Username, &u.Email, &hash, &u.DisplayName, &u.Avatar); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.Wrap(errLoginUserNotFound, "App.userLogin")
		}

		return nil, errors.Wrap(err, "App.userLogin")
	}

	newHash, err := passlib.Verify(password, hash)
	if err != nil {
		return nil, errors.Wrap(errLoginUserNotFound, "App.userLogin")
	}

	if newHash != "" {
		if _, err := a.SQLDB.Exec("update users set hash = $1 where id = $2 and hash = $3", newHash, u.ID, hash); err != nil {
			return nil, errors.Wrap(err, "App.userLogin")
		}
	}

	return &u, nil
}
