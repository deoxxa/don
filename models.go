package main

import (
	"encoding/json"
	"time"
)

type User struct {
	ID          string    `json:"id" sql:"id,text,primary_key,table=users"`
	CreatedAt   time.Time `json:"createdAt" sql:"created_at,datetime,not_null"`
	Username    string    `json:"username" sql:"username,text,not_null"`
	Email       string    `json:"email" sql:"email,text,not_null"`
	DisplayName *string   `json:"displayName" sql:"display_name,text"`
	Avatar      *string   `json:"avatar" sql:"avatar,text"`
}

type ActivityEvent struct {
	RowID    int64
	Activity *Activity
	JSON     json.RawMessage
}

type Activity struct {
	ID           string    `json:"id" sql:"id,text,primary_key,table=activities"`
	Permalink    string    `json:"permalink" sql:"permalink,text,not_null"`
	ActorID      *string   `json:"actorID" sql:"actor_id,text"`
	Actor        *Person   `json:"actor" sql:",reference"`
	ObjectID     string    `json:"objectID" sql:"object_id,text,not_null"`
	Object       Object    `json:"object" sql:",reference"`
	Verb         string    `json:"verb" sql:"verb,text,not_null"`
	Time         time.Time `json:"time" sql:"time,datetime,not_null"`
	Title        string    `json:"title" sql:"title,text"`
	InReplyToID  *string   `json:"inReplyToID" sql:"in_reply_to_id,text"`
	InReplyToURL *string   `json:"inReplyToURL" sql:"in_reply_to_url,text"`
}

type Person struct {
	ID          string    `json:"id" sql:"id,text,primary_key,table=people"`
	Host        string    `json:"host" sql:"host,text,not_null"`
	FirstSeen   time.Time `json:"firstSeen" sql:"first_seen,datetime,not_null"`
	Permalink   string    `json:"permalink" sql:"permalink,text,not_null"`
	DisplayName *string   `json:"displayName" sql:"display_name,text"`
	Avatar      *string   `json:"avatar" sql:"avatar,text"`
	Summary     *string   `json:"summary" sql:"summary,text"`
}

type Object struct {
	ID                  string  `json:"id" sql:"id,text,primary_key,table=objects"`
	Name                *string `json:"name" sql:"name,text"`
	Summary             *string `json:"summary" sql:"summary,text"`
	RepresentativeImage *string `json:"representativeImage" sql:"representative_image,text"`
	Permalink           *string `json:"permalink" sql:"permalink,text"`
	ObjectType          *string `json:"objectType" sql:"object_type,text"`
	Content             *string `json:"content" sql:"content,text"`
}
