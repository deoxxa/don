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
