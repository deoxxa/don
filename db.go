package main

import (
	"database/sql"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/umisama/go-sqlbuilder"

	"fknsrs.biz/p/don/acct"
	"fknsrs.biz/p/don/activitystreams"
)

var (
	peopleTable = sqlbuilder.NewTable(
		"people",
		nil,
		sqlbuilder.IntColumn("ROWID", nil),
		sqlbuilder.StringColumn("id", &sqlbuilder.ColumnOption{NotNull: true, PrimaryKey: true}),
		sqlbuilder.StringColumn("host", &sqlbuilder.ColumnOption{NotNull: true}),
		sqlbuilder.DateColumn("first_seen", &sqlbuilder.ColumnOption{NotNull: true}),
		sqlbuilder.StringColumn("permalink", &sqlbuilder.ColumnOption{NotNull: true}),
		sqlbuilder.StringColumn("display_name", nil),
		sqlbuilder.StringColumn("avatar", nil),
		sqlbuilder.StringColumn("summary", nil),
	)

	objectsTable = sqlbuilder.NewTable(
		"objects",
		nil,
		sqlbuilder.IntColumn("ROWID", nil),
		sqlbuilder.StringColumn("id", &sqlbuilder.ColumnOption{NotNull: true, PrimaryKey: true}),
		sqlbuilder.StringColumn("name", nil),
		sqlbuilder.StringColumn("summary", nil),
		sqlbuilder.StringColumn("representative_image", nil),
		sqlbuilder.StringColumn("permalink", nil),
		sqlbuilder.StringColumn("object_type", nil),
		sqlbuilder.StringColumn("content", nil),
	)

	activitiesTable = sqlbuilder.NewTable(
		"activities",
		nil,
		sqlbuilder.IntColumn("ROWID", nil),
		sqlbuilder.StringColumn("id", &sqlbuilder.ColumnOption{NotNull: true, PrimaryKey: true}),
		sqlbuilder.StringColumn("permalink", &sqlbuilder.ColumnOption{NotNull: true}),
		sqlbuilder.StringColumn("actor", nil),
		sqlbuilder.StringColumn("object", &sqlbuilder.ColumnOption{NotNull: true}),
		sqlbuilder.StringColumn("verb", &sqlbuilder.ColumnOption{NotNull: true}),
		sqlbuilder.DateColumn("time", &sqlbuilder.ColumnOption{NotNull: true}),
		sqlbuilder.StringColumn("title", nil),
		sqlbuilder.StringColumn("in_reply_to_id", nil),
		sqlbuilder.StringColumn("in_reply_to_url", nil),
	)
)

func (a *App) savePerson(p *activitystreams.Author) (*acct.URL, error) {
	var permalink string
	for _, l := range p.GetLinks("alternate") {
		if l.Type == "text/html" && permalink == "" {
			permalink = l.Href
		}
	}
	if permalink == "" {
		permalink = p.URI
	}

	if permalink == "" {
		return nil, errors.Errorf("App.savePerson: couldn't find permalink for user")
	}

	accountURLString, _, err := a.AccountURLCache.Get(permalink, p)

	if err != nil || len(accountURLString) == 0 {
		return nil, nil
	}

	accountURL, err := acct.FromString(string(accountURLString))
	if err != nil {
		return nil, errors.Wrap(err, "App.savePerson: couldn't parse account url")
	}

	tx, err := a.SQLDB.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "App.savePerson: couldn't begin transaction")
	}
	defer tx.Rollback()

	var exists bool
	var displayName, avatar, summary string
	var firstSeen time.Time
	if err := tx.QueryRow("select 1, first_seen, display_name, avatar, summary from people where id = $1", accountURL.String()).Scan(&exists, &firstSeen, &displayName, &avatar, &summary); err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "App.savePerson: couldn't query for existing person")
	}

	var changed bool

	if newDisplayName := strings.TrimSpace(p.DisplayName); newDisplayName != "" && newDisplayName != displayName {
		displayName = newDisplayName
		changed = true
	}
	if newSummary := strings.TrimSpace(p.Summary); newSummary != "" && newSummary != summary {
		summary = newSummary
		changed = true
	}
	if newAvatar := strings.TrimSpace(p.GetBestAvatar()); newAvatar != "" && newAvatar != avatar {
		avatar = newAvatar
		changed = true
	}

	if !exists {
		firstSeen = time.Now()
		if _, err := tx.Exec("insert into people (id, host, first_seen, permalink, display_name, avatar, summary) values ($1, $2, $3, $4, $5, $6, $7)", accountURL.String(), accountURL.Host, firstSeen, permalink, displayName, avatar, summary); err != nil {
			return nil, errors.Wrap(err, "App.savePerson: couldn't save person to db")
		}
	} else if changed {
		if _, err := tx.Exec("update people set display_name = $1, avatar = $2, summary = $3 where id = $4", displayName, avatar, summary, accountURL.String()); err != nil {
			return nil, errors.Wrap(err, "App.savePerson: couldn't update person in db")
		}
	}

	return accountURL, errors.Wrap(tx.Commit(), "App.savePerson: couldn't commit transaction")
}

func (a *App) saveActivity(e activitystreams.ActivityLike) error {
	o := e.GetObject()

	if o != e {
		if e2, ok := o.(activitystreams.ActivityLike); ok {
			if err := a.saveActivity(e2); err != nil {
				return errors.Wrap(err, "saveActivity: couldn't save nested activity")
			}
		}

		if p, ok := o.(*activitystreams.Author); ok {
			if _, err := a.savePerson(p); err != nil {
				return errors.Wrap(err, "saveActivity: couldn't save nested person")
			}
		}
	}

	var actorID sql.NullString
	if actor := e.GetActor(); actor != nil {
		if u, err := a.savePerson(actor); err != nil {
			return errors.Wrap(err, "saveActivity: couldn't save author")
		} else if u != nil {
			actorID.Valid = true
			actorID.String = u.String()
		}
	}

	if err := a.saveObject(o); err != nil {
		return errors.Wrap(err, "saveActivity: couldn't save object")
	}

	if err := a.saveObject(e); err != nil {
		return errors.Wrap(err, "saveActivity: couldn't save activity as object")
	}

	tx, err := a.SQLDB.Begin()
	if err != nil {
		return errors.Wrap(err, "saveActivity: couldn't begin transaction")
	}
	defer tx.Rollback()

	var exists int
	if err := tx.QueryRow("select count(*) from activities where id = $1", e.GetID()).Scan(&exists); err != nil {
		return errors.Wrap(err, "saveActivity: couldn't query for existing activities")
	}

	if exists == 0 {
		var inReplyToID, inReplyToURL sql.NullString
		if e, ok := e.(activitystreams.HasInReplyTo); ok {
			if s := e.GetInReplyTo(); s != nil {
				inReplyToID.Valid = true
				inReplyToID.String = s.Ref

				inReplyToURL.Valid = true
				inReplyToURL.String = s.Href
			}
		}

		if _, err := tx.Exec("insert into activities (id, permalink, actor, object, verb, time, title, in_reply_to_id, in_reply_to_url) values ($1, $2, $3, $4, $5, $6, $7, $8, $9)", e.GetID(), e.GetPermalink(), actorID, e.GetObject().GetID(), e.GetVerb(), e.GetTime(), e.GetTitle(), inReplyToID, inReplyToURL); err != nil {
			return errors.Wrap(err, "saveActivity: couldn't save activity to db")
		}
	}

	return errors.Wrap(tx.Commit(), "saveActivity: couldn't commit transaction")
}

func (a *App) saveObject(o activitystreams.ObjectLike) error {
	tx, err := a.SQLDB.Begin()
	if err != nil {
		return errors.Wrap(err, "saveObject: couldn't begin transaction")
	}
	defer tx.Rollback()

	var exists int
	if err := tx.QueryRow("select count(*) from objects where id = $1", o.GetID()).Scan(&exists); err != nil {
		return errors.Wrap(err, "saveObject: couldn't query for existing objects")
	}

	if exists == 0 {
		var content sql.NullString
		if hc, ok := o.(activitystreams.HasContent); ok {
			if c := hc.GetContent(); c != "" {
				content.Valid = true
				content.String = c
			}
		}

		if _, err := tx.Exec("insert into objects (id, name, summary, representative_image, permalink, object_type, content) values ($1, $2, $3, $4, $5, $6, $7)", o.GetID(), o.GetName(), o.GetSummary(), o.GetRepresentativeImage(), o.GetPermalink(), o.GetObjectType(), content); err != nil {
			return errors.Wrap(err, "saveObject: couldn't save object to db")
		}
	}

	return errors.Wrap(tx.Commit(), "saveObject: couldn't commit transaction")
}

type getPublicTimelineArgs struct {
	After  time.Time `schema:"after"`
	Before time.Time `schema:"before"`
}

func (a *App) getPublicTimeline(args getPublicTimelineArgs) ([]Activity, error) {
	qb := sqlbuilder.
		Select(activitiesTable.
			LeftOuterJoin(peopleTable, peopleTable.C("id").Eq(activitiesTable.C("actor"))).
			LeftOuterJoin(objectsTable, objectsTable.C("id").Eq(activitiesTable.C("object"))),
		).
		Columns(
			activitiesTable.C("id"),
			activitiesTable.C("permalink"),
			activitiesTable.C("actor"),
			activitiesTable.C("object"),
			activitiesTable.C("verb"),
			activitiesTable.C("time"),
			activitiesTable.C("title"),
			activitiesTable.C("in_reply_to_id"),
			activitiesTable.C("in_reply_to_url"),
			objectsTable.C("id"),
			objectsTable.C("name"),
			objectsTable.C("summary"),
			objectsTable.C("representative_image"),
			objectsTable.C("permalink"),
			objectsTable.C("object_type"),
			objectsTable.C("content"),
			peopleTable.C("id"),
			peopleTable.C("host"),
			peopleTable.C("first_seen"),
			peopleTable.C("permalink"),
			peopleTable.C("display_name"),
			peopleTable.C("avatar"),
			peopleTable.C("summary"),
		).
		OrderBy(true, activitiesTable.C("time")).
		Limit(50)

	if !args.After.IsZero() {
		qb = qb.Where(activitiesTable.C("time").Gt(args.After))
	}
	if !args.Before.IsZero() {
		qb = qb.Where(activitiesTable.C("time").Lt(args.Before))
	}

	q, vars, err := qb.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "getPublicTimeline")
	}

	rows, err := a.SQLDB.Query(q, vars...)
	if err != nil {
		return nil, errors.Wrap(err, "getPublicTimeline")
	}
	defer rows.Close()

	var activities []Activity
	for rows.Next() {
		var activity Activity

		var (
			personID          *string
			personHost        *string
			personFirstSeen   *time.Time
			personPermalink   *string
			personDisplayName *string
			personAvatar      *string
			personSummary     *string
		)

		if err := rows.Scan(
			&activity.ID,
			&activity.Permalink,
			&activity.ActorID,
			&activity.ObjectID,
			&activity.Verb,
			&activity.Time,
			&activity.Title,
			&activity.InReplyToID,
			&activity.InReplyToURL,
			&activity.Object.ID,
			&activity.Object.Name,
			&activity.Object.Summary,
			&activity.Object.RepresentativeImage,
			&activity.Object.Permalink,
			&activity.Object.ObjectType,
			&activity.Object.Content,
			&personID,
			&personHost,
			&personFirstSeen,
			&personPermalink,
			&personDisplayName,
			&personAvatar,
			&personSummary,
		); err != nil {
			return nil, err
		}

		if personID != nil && personHost != nil && personFirstSeen != nil && personPermalink != nil {
			activity.Actor = &Person{
				ID:          *personID,
				Host:        *personHost,
				FirstSeen:   *personFirstSeen,
				Permalink:   *personPermalink,
				DisplayName: personDisplayName,
				Avatar:      personAvatar,
				Summary:     personSummary,
			}
		}

		activities = append(activities, activity)
	}

	return activities, nil
}
