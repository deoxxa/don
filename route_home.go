package main

import (
	"net/http"
)

func (a *App) handleHomeGet(r *http.Request, ar *AppResponse) *AppResponse {
	var args getPublicTimelineArgs
	if err := decoder.Decode(&args, r.URL.Query()); err != nil {
		return ar.WithError(err)
	}

	activities, err := a.getPublicTimeline(args)
	if err != nil {
		return ar.WithError(err)
	}

	if activities == nil {
		activities = []Activity{}
	}

	return ar.ShallowMergeState(map[string]interface{}{
		"publicTimeline": map[string]interface{}{
			"loading":    false,
			"activities": activities,
			"error":      nil,
		},
	})
}
