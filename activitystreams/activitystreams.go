package activitystreams

import (
	"time"

	"github.com/pkg/errors"

	"fknsrs.biz/p/don/commonxml"
)

func Fetch(u string) (*Feed, error) {
	var f Feed
	if err := commonxml.Fetch(u, &f); err != nil {
		return nil, errors.Wrap(err, "activitystreams.Fetch")
	}

	return &f, nil
}

func Parse(d []byte) (*Feed, error) {
	var f Feed
	if err := commonxml.Parse(d, &f); err != nil {
		return nil, errors.Wrap(err, "activitystreams.Parse")
	}

	return &f, nil
}

type ObjectLike interface {
	GetID() string
	GetName() string
	GetSummary() string
	GetRepresentativeImage() string
	GetPermalink() string
	GetObjectType() string
}

type ActivityLike interface {
	ObjectLike

	GetActor() *Author
	GetObject() ObjectLike
	GetVerb() string
	GetTime() time.Time
	GetTitle() string
}

type NoteLike interface {
	ObjectLike

	GetContent() string
}

type HasContent interface {
	GetContent() string
}

type HasInReplyTo interface {
	GetInReplyTo() *InReplyTo
}
