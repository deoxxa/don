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

func (a *App) savePerson(p *activitystreams.Author) (*Person, error) {
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

	selectQuery := sqlbuilder.Select(peopleTable).Columns(
		peopleTable.C("first_seen"),
		peopleTable.C("display_name"),
		peopleTable.C("avatar"),
		peopleTable.C("summary"),
	).Where(peopleTable.C("id").Eq(accountURL.String()))

	selectQuerySQL, selectQueryVars, err := selectQuery.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "App.savePerson: couldn't make select query")
	}

	var firstSeen time.Time
	var displayName, avatar, summary string
	if err := tx.QueryRow(selectQuerySQL, selectQueryVars...).Scan(&firstSeen, &displayName, &avatar, &summary); err != nil && err != sql.ErrNoRows {
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

	if firstSeen.IsZero() {
		firstSeen = time.Now()

		insertQuery := sqlbuilder.Insert(peopleTable).Columns(
			peopleTable.C("id"),
			peopleTable.C("host"),
			peopleTable.C("first_seen"),
			peopleTable.C("permalink"),
			peopleTable.C("display_name"),
			peopleTable.C("avatar"),
			peopleTable.C("summary"),
		).Values(accountURL.String(), accountURL.Host, firstSeen, permalink, displayName, avatar, summary)

		insertQuerySQL, insertQueryVars, err := insertQuery.ToSql()
		if err != nil {
			return nil, errors.Wrap(err, "App.savePerson: couldn't make insert query")
		}

		if _, err := tx.Exec(insertQuerySQL, insertQueryVars...); err != nil {
			return nil, errors.Wrap(err, "App.savePerson: couldn't save person to db")
		}
	} else if changed {
		updateQuery := sqlbuilder.Update(peopleTable).
			Set(peopleTable.C("display_name"), displayName).
			Set(peopleTable.C("avatar"), avatar).
			Set(peopleTable.C("summary"), summary).
			Where(peopleTable.C("id").Eq(accountURL.String()))

		updateQuerySQL, updateQueryVars, err := updateQuery.ToSql()
		if err != nil {
			return nil, errors.Wrap(err, "App.savePerson: couldn't make update query")
		}

		if _, err := tx.Exec(updateQuerySQL, updateQueryVars...); err != nil {
			return nil, errors.Wrap(err, "App.savePerson: couldn't update person in db")
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "App.savePerson: couldn't commit transaction")
	}

	return &Person{
		ID:          accountURL.String(),
		Host:        accountURL.Host,
		FirstSeen:   firstSeen,
		Permalink:   permalink,
		DisplayName: &displayName,
		Avatar:      &avatar,
		Summary:     &summary,
	}, nil
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

	var person *Person

	var actorID sql.NullString
	if actor := e.GetActor(); actor != nil {
		if p, err := a.savePerson(actor); err != nil {
			return errors.Wrap(err, "saveActivity: couldn't save author")
		} else if p != nil {
			actorID.Valid = true
			actorID.String = p.ID

			person = p
		}
	}

	object, err := a.saveObject(o)
	if err != nil {
		return errors.Wrap(err, "saveActivity: couldn't save object")
	}

	if _, err := a.saveObject(e); err != nil {
		return errors.Wrap(err, "saveActivity: couldn't save activity as object")
	}

	tx, err := a.SQLDB.Begin()
	if err != nil {
		return errors.Wrap(err, "saveActivity: couldn't begin transaction")
	}
	defer tx.Rollback()

	var rowID sql.NullInt64
	if err := tx.QueryRow("select ROWID from activities where id = $1", e.GetID()).Scan(&rowID); err != nil && err != sql.ErrNoRows {
		return errors.Wrap(err, "saveActivity: couldn't query for existing activities")
	}

	var inReplyToID, inReplyToURL sql.NullString
	if e, ok := e.(activitystreams.HasInReplyTo); ok {
		if s := e.GetInReplyTo(); s != nil {
			inReplyToID.Valid = true
			inReplyToID.String = s.Ref

			inReplyToURL.Valid = true
			inReplyToURL.String = s.Href
		}
	}

	if rowID.Valid {
		return nil
	}

	res, err := tx.Exec("insert into activities (id, permalink, actor, object, verb, time, title, in_reply_to_id, in_reply_to_url) values ($1, $2, $3, $4, $5, $6, $7, $8, $9)", e.GetID(), e.GetPermalink(), actorID, e.GetObject().GetID(), e.GetVerb(), e.GetTime(), e.GetTitle(), inReplyToID, inReplyToURL)
	if err != nil {
		return errors.Wrap(err, "saveActivity: couldn't save activity to db")
	}

	insertID, err := res.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "saveActivity: couldn't get row id")
	}

	rowID.Valid = true
	rowID.Int64 = insertID

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "saveActivity: couldn't commit transaction")
	}

	activity := Activity{
		ID:        e.GetID(),
		Permalink: e.GetPermalink(),
		ObjectID:  object.ID,
		Object:    *object,
		Verb:      e.GetVerb(),
		Time:      e.GetTime(),
		Title:     e.GetTitle(),
	}

	if person != nil {
		activity.ActorID = &person.ID
		activity.Actor = person
	}

	if inReplyToID.Valid {
		activity.InReplyToID = &inReplyToID.String
	}
	if inReplyToURL.Valid {
		activity.InReplyToURL = &inReplyToURL.String
	}

	a.Emit(&ActivityEvent{RowID: rowID.Int64, Activity: &activity})

	return nil
}

func (a *App) saveObject(o activitystreams.ObjectLike) (*Object, error) {
	tx, err := a.SQLDB.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "saveObject: couldn't begin transaction")
	}
	defer tx.Rollback()

	var id, name, summary, representativeImage, permalink, objectType, content sql.NullString
	if err := tx.QueryRow("select id, name, summary, representative_image, permalink, object_type, content from objects where id = $1", o.GetID()).Scan(&id, &name, &summary, &representativeImage, &permalink, &objectType, &content); err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "saveObject: couldn't query for existing objects")
	}

	if !id.Valid {
		id.Valid = true
		id.String = o.GetID()

		if s := o.GetName(); s != "" {
			name.Valid = true
			name.String = s
		}

		if s := o.GetSummary(); s != "" {
			summary.Valid = true
			summary.String = s
		}

		if s := o.GetRepresentativeImage(); s != "" {
			representativeImage.Valid = true
			representativeImage.String = s
		}

		if s := o.GetPermalink(); s != "" {
			permalink.Valid = true
			permalink.String = s
		}

		if s := o.GetObjectType(); s != "" {
			objectType.Valid = true
			objectType.String = s
		}

		if hc, ok := o.(activitystreams.HasContent); ok {
			if c := hc.GetContent(); c != "" {
				content.Valid = true
				content.String = c
			}
		}

		if _, err := tx.Exec("insert into objects (id, name, summary, representative_image, permalink, object_type, content) values ($1, $2, $3, $4, $5, $6, $7)", o.GetID(), o.GetName(), o.GetSummary(), o.GetRepresentativeImage(), o.GetPermalink(), o.GetObjectType(), content); err != nil {
			return nil, errors.Wrap(err, "saveObject: couldn't save object to db")
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "saveObject: couldn't commit transaction")
	}

	object := &Object{ID: id.String}

	if name.Valid {
		object.Name = &name.String
	}
	if summary.Valid {
		object.Summary = &summary.String
	}
	if representativeImage.Valid {
		object.RepresentativeImage = &representativeImage.String
	}
	if permalink.Valid {
		object.Permalink = &permalink.String
	}
	if objectType.Valid {
		object.ObjectType = &objectType.String
	}
	if content.Valid {
		object.Content = &content.String
	}

	return object, nil
}

type getPublicTimelineArgs struct {
	Q       string    `schema:"q"`
	After   time.Time `schema:"after"`
	Before  time.Time `schema:"before"`
	Account string    `schema:"account"`
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

	var conditions []sqlbuilder.Condition

	if !args.After.IsZero() {
		conditions = append(conditions, activitiesTable.C("time").Gt(args.After))
	}
	if !args.Before.IsZero() {
		conditions = append(conditions, activitiesTable.C("time").Lt(args.Before))
	}
	if args.Account != "" {
		conditions = append(conditions, activitiesTable.C("actor").Eq("acct:"+strings.TrimPrefix(strings.TrimPrefix(args.Account, "@"), "acct:")))
	}

	if args.Q != "" {
		var l []sqlbuilder.Condition

		for _, s := range strings.Split(args.Q, " ") {
			s = strings.TrimSpace(s)

			if len(s) > 0 {
				l = append(l, objectsTable.C("content").Like("%"+s+"%"))
			}
		}

		if len(l) > 0 {
			conditions = append(conditions, sqlbuilder.And(l...))
		}
	}

	if len(conditions) > 0 {
		qb = qb.Where(sqlbuilder.And(conditions...))
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
