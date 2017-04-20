package main

import (
	"time"
)

type User struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	DisplayName *string   `json:"displayName"`
	Avatar      *string   `json:"avatar"`
}

type Activity struct {
	ID           string    `json:"id"`
	Permalink    string    `json:"permalink"`
	ActorID      *string   `json:"actorID"`
	Actor        *Person   `json:"actor"`
	ObjectID     string    `json:"objectID"`
	Object       Object    `json:"object"`
	Verb         string    `json:"verb"`
	Time         time.Time `json:"time"`
	Title        *string   `json:"title"`
	InReplyToID  *string   `json:"inReplyToID"`
	InReplyToURL *string   `json:"inReplyToURL"`
}

type Person struct {
	ID          string    `json:"id"`
	Host        string    `json:"host"`
	FirstSeen   time.Time `json:"firstSeen"`
	Permalink   string    `json:"permalink"`
	DisplayName *string   `json:"displayName"`
	Avatar      *string   `json:"avatar"`
	Summary     *string   `json:"summary"`
}

type Object struct {
	ID                  string  `json:"id"`
	Name                *string `json:"name"`
	Summary             *string `json:"summary"`
	RepresentativeImage *string `json:"representativeImage"`
	Permalink           *string `json:"permalink"`
	ObjectType          *string `json:"objectType"`
	Content             *string `json:"content"`
}
