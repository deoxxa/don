package main

import (
	"github.com/gorilla/sessions"
)

type AppResponse struct {
	Session  *sessions.Session
	User     *User
	Status   int
	Error    error
	Redirect string
	State    map[string]interface{}
	Meta     map[string]string
}

func NewAppResponse() *AppResponse {
	return &AppResponse{
		State: make(map[string]interface{}),
		Meta:  make(map[string]string),
	}
}

func (a *AppResponse) WithSession(session *sessions.Session) *AppResponse {
	c := *a
	c.Session = session
	return &c
}

func (a *AppResponse) WithUser(user *User) *AppResponse {
	c := *a
	c.User = user
	return &c
}

func (a *AppResponse) WithStatus(status int) *AppResponse {
	c := *a
	c.Status = status
	return &c
}

func (a *AppResponse) WithError(err error) *AppResponse {
	c := *a
	c.Error = err
	return &c
}

func (a *AppResponse) WithRedirect(redirect string) *AppResponse {
	c := *a
	c.Redirect = redirect
	return &c
}

func (a *AppResponse) WithState(state map[string]interface{}) *AppResponse {
	c := *a
	c.State = state
	return &c
}

func (a *AppResponse) WithMeta(meta map[string]string) *AppResponse {
	c := *a
	c.Meta = meta
	return &c
}

func (a *AppResponse) GetMeta(name, defaultValue string) string {
	if v, ok := a.Meta[name]; ok {
		return v
	}

	return defaultValue
}

func (a *AppResponse) ShallowMergeState(state map[string]interface{}) *AppResponse {
	s := make(map[string]interface{})

	for k, v := range a.State {
		s[k] = v
	}

	for k, v := range state {
		s[k] = v
	}

	return a.WithState(s)
}

func (a *AppResponse) MergeMeta(meta map[string]string) *AppResponse {
	m := make(map[string]string)

	for k, v := range a.Meta {
		m[k] = v
	}

	for k, v := range meta {
		m[k] = v
	}

	return a.WithMeta(m)
}

func (a *AppResponse) MergeUserContext(user *User) *AppResponse {
	if user == nil {
		delete(a.Session.Values, "user_id")
	} else {
		a.Session.Values["user_id"] = user.ID
	}

	return a.WithUser(user).ShallowMergeState(map[string]interface{}{
		"authentication": map[string]interface{}{
			"loading": false,
			"error":   nil,
			"user":    user,
		},
	})
}
