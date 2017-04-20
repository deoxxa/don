package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"gopkg.in/hlandau/passlib.v1"
)

func (a *App) handleRegisterGet(r *http.Request, ar *AppResponse) *AppResponse {
	return ar.MergeMeta(map[string]string{
		"Title":       "Register",
		"Description": "Register a new account.",
	})
}

func (a *App) handleRegisterPost(r *http.Request, ar *AppResponse) *AppResponse {
	ar = ar.MergeMeta(map[string]string{
		"Title":       "Register",
		"Description": "Register a new account.",
	})

	if err := r.ParseForm(); err != nil {
		return ar.WithStatus(http.StatusBadRequest).WithError(errors.Wrap(err, "App.handleRegisterPost: couldn't parse form"))
	}

	var v struct {
		Email    string `schema:"email"`
		Username string `schema:"username"`
		Password string `schema:"password"`
	}

	if err := decoder.Decode(&v, r.PostForm); err != nil {
		return ar.WithError(errors.Wrap(err, "App.handleRegisterPost: couldn't decode form fields"))
	}

	u, err := a.userRegister(v.Email, v.Username, v.Password)

	if err != nil {
		ar = ar.WithUser(nil).WithError(errors.Wrap(err, "App.handleRegisterPost: couldn't register user")).ShallowMergeState(map[string]interface{}{
			"authentication": map[string]interface{}{
				"loading": false,
				"error":   err.Error(),
				"user":    nil,
			},
		})

		switch errors.Cause(err) {
		case errUsernameDisallowed, errUsernameInvalid:
			return ar.WithStatus(http.StatusForbidden)
		case errRegisterUsernameAlreadyExists:
			return ar.WithStatus(http.StatusConflict)
		default:
			return ar
		}
	}

	return ar.MergeUserContext(u).WithRedirect(r.URL.Query().Get("return_to"))
}

var (
	errRegisterUsernameAlreadyExists = errors.New("a user with that username already exists")
)

func (a *App) userRegister(email, username, password string) (*User, error) {
	if err := VerifyUsername(username); err != nil {
		return nil, errors.Wrap(err, "App.userRegister")
	}

	var id string
	if err := a.SQLDB.QueryRow("select id from users where username = $1 or email = $2", username, email).Scan(&id); err == nil {
		return nil, errors.Wrap(errRegisterUsernameAlreadyExists, "App.userRegister")
	} else if err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "App.userRegister")
	}

	hash, err := passlib.Hash(password)
	if err != nil {
		return nil, errors.Wrap(err, "App.userRegister")
	}

	u := User{
		ID:        uuid.NewV4().String(),
		CreatedAt: time.Now(),
		Username:  username,
		Email:     email,
	}

	if _, err := a.SQLDB.Exec("insert into users (id, created_at, username, email, hash) values ($1, $2, $3, $4, $5)", u.ID, u.CreatedAt, u.Username, u.Email, hash); err != nil {
		return nil, errors.Wrap(err, "App.userRegister")
	}

	return &u, nil
}
