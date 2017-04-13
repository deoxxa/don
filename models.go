package main

import (
	"time"
)

type UIStatus struct {
	ID          int       `json:"id"`
	AuthorAcct  string    `json:"authorAcct"`
	AuthorName  string    `json:"authorName"`
	Time        time.Time `json:"time"`
	ContentText string    `json:"contentText"`
	ContentHTML string    `json:"contentHTML"`
}

type User struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	DisplayName *string   `json:"displayName"`
	Avatar      *string   `json:"avatar"`
}
