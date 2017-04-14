package main

import (
	"net/http"
)

func (a *App) handleLogoutGet(r *http.Request, ar *AppResponse) *AppResponse {
	return ar.MergeMeta(map[string]string{
		"Title":       "Log out",
		"Description": "Log out of don.",
	})
}

func (a *App) handleLogoutPost(r *http.Request, ar *AppResponse) *AppResponse {
	ar.Session.Options.MaxAge = -1
	return ar.WithRedirect("/")
}
