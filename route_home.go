package main

import (
	"net/http"
)

func (a *App) handleHomeGet(r *http.Request, ar *AppResponse) *AppResponse {
	var args getPublicTimelineArgs
	if err := decoder.Decode(&args, r.URL.Query()); err != nil {
		return ar.WithError(err)
	}

	posts, err := getPublicTimeline(a.DB, args)
	if err != nil {
		return ar.WithError(err)
	}

	return ar.ShallowMergeState(map[string]interface{}{
		"publicTimeline": map[string]interface{}{
			"loading": false,
			"posts":   posts,
			"error":   nil,
		},
	})
}
