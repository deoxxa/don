package main

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/umisama/go-sqlbuilder"
)

var (
	peopleTable = sqlbuilder.NewTable(
		"people",
		&sqlbuilder.TableOption{Unique: [][]string{{"hub", "topic"}}},
		sqlbuilder.IntColumn("ROWID", nil),
		sqlbuilder.StringColumn("feed_url", &sqlbuilder.ColumnOption{NotNull: true, PrimaryKey: true}),
		sqlbuilder.DateColumn("first_seen", &sqlbuilder.ColumnOption{NotNull: true}),
		sqlbuilder.StringColumn("name", nil),
		sqlbuilder.StringColumn("display_name", nil),
		sqlbuilder.StringColumn("email", nil),
		sqlbuilder.StringColumn("summary", nil),
		sqlbuilder.StringColumn("note", nil),
	)

	postsTable = sqlbuilder.NewTable(
		"posts",
		&sqlbuilder.TableOption{Unique: [][]string{{"feed_url", "id"}}},
		sqlbuilder.IntColumn("ROWID", nil),
		sqlbuilder.StringColumn("feed_url", &sqlbuilder.ColumnOption{NotNull: true}),
		sqlbuilder.StringColumn("id", &sqlbuilder.ColumnOption{NotNull: true}),
		sqlbuilder.StringColumn("created_at", &sqlbuilder.ColumnOption{NotNull: true}),
		sqlbuilder.StringColumn("raw_entry", &sqlbuilder.ColumnOption{NotNull: true}),
	)
)

func savePerson(db *sql.DB, feedURL string, author *AtomAuthor) error {
	tx, err := db.Begin()
	if err != nil {
		return errors.Wrap(err, "savePerson")
	}
	defer tx.Rollback()

	if err := savePersonTx(tx, feedURL, author); err != nil {
		return errors.Wrap(err, "savePerson")
	}

	return errors.Wrap(tx.Commit(), "savePerson")
}

func savePersonTx(tx *sql.Tx, feedURL string, author *AtomAuthor) error {
	var name, displayName, email, summary, note string
	if err := tx.QueryRow("select name, display_name, email, summary, note from people where feed_url = $1", feedURL).Scan(&name, &displayName, &email, &summary, &note); err != nil {
		if err != sql.ErrNoRows {
			return errors.Wrap(err, "savePersonTx")
		}

		if _, err := tx.Exec("insert into people (feed_url, first_seen, name, display_name, email, summary, note) values ($1, $2, $3, $4, $5, $6, $7)", feedURL, time.Now(), author.Name, author.DisplayName, author.Email, author.Summary, author.Note); err != nil {
			return errors.Wrap(err, "savePersonTx")
		}
	} else {
		if name == author.Name && displayName == author.DisplayName && email == author.Email && summary == author.Summary && note == author.Note {
			return nil
		}

		if _, err := tx.Exec("update people set name = $1, display_name = $2, email = $3, summary = $4, note = $5 where feed_url = $6", author.Name, author.DisplayName, author.Email, author.Summary, author.Note, feedURL); err != nil {
			return errors.Wrap(err, "savePersonTx")
		}
	}

	return nil
}

func saveEntry(db *sql.DB, feedURL string, entry *AtomEntry) error {
	tx, err := db.Begin()
	if err != nil {
		return errors.Wrap(err, "saveEntry")
	}
	defer tx.Rollback()

	if err := saveEntryTx(tx, feedURL, entry); err != nil {
		return errors.Wrap(err, "saveEntry")
	}

	return errors.Wrap(tx.Commit(), "saveEntry")
}

func saveEntryTx(tx *sql.Tx, feedURL string, entry *AtomEntry) error {
	var exists int
	if err := tx.QueryRow("select count(1) from posts where feed_url = $1 and id = $2", feedURL, entry.ID).Scan(&exists); err != nil {
		return errors.Wrap(err, "saveEntryTx")
	}

	if exists > 0 {
		return nil
	}

	d, err := json.Marshal(entry)
	if err != nil {
		return errors.Wrap(err, "saveEntryTx")
	}

	if _, err := tx.Exec("insert into posts (feed_url, id, created_at, raw_entry, xml_entry) values ($1, $2, $3, $4, $5)", feedURL, entry.ID, entry.Published, string(d), entry.InnerXML); err != nil {
		return errors.Wrap(err, "saveEntryTx")
	}

	return nil
}

type getPublicTimelineArgs struct {
	AfterID int `schema:"after_id"`
	Offset  int `schema:"offset,omitempty"`
	Limit   int `schema:"limit,omitempty"`
}

func getPublicTimeline(db *sql.DB, args getPublicTimelineArgs) ([]UIStatus, error) {
	qb := sqlbuilder.
		Select(postsTable.LeftOuterJoin(peopleTable, peopleTable.C("feed_url").Eq(postsTable.C("feed_url")))).
		Columns(
			postsTable.C("ROWID"),
			postsTable.C("feed_url"),
			postsTable.C("raw_entry"),
			peopleTable.C("name"),
			peopleTable.C("display_name"),
			peopleTable.C("email"),
		).
		OrderBy(true, postsTable.C("created_at"))

	if args.Offset > 0 && args.Offset < 225 {
		qb = qb.Offset(args.Offset)
	}

	if args.Limit > 0 && args.Limit <= 25 {
		qb = qb.Limit(args.Limit)
	} else {
		qb = qb.Limit(25)
	}

	if args.AfterID != 0 {
		qb = qb.Where(postsTable.C("ROWID").Gt(args.AfterID))
	}

	q, vars, err := qb.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "getPublicTimeline")
	}

	rows, err := db.Query(q, vars...)
	if err != nil {
		return nil, errors.Wrap(err, "getPublicTimeline")
	}
	defer rows.Close()

	var posts []UIStatus
	for rows.Next() {
		var id int
		var feedURL, rawEntry string
		var name, displayName, email sql.NullString
		if err := rows.Scan(&id, &feedURL, &rawEntry, &name, &displayName, &email); err != nil {
			return nil, errors.Wrap(err, "getPublicTimeline")
		}

		var entry AtomEntry
		if err := json.Unmarshal([]byte(rawEntry), &entry); err != nil {
			return nil, errors.Wrap(err, "getPublicTimeline")
		}

		post := UIStatus{ID: id}

		if name.Valid {
			post.AuthorAcct = email.String
			post.AuthorName = name.String
		}

		if t, err := time.Parse(time.RFC3339, entry.Published); err == nil {
			post.Time = t
		}

		if entry.Content != nil {
			post.ContentHTML = entry.Content.HTML()
			post.ContentText = entry.Content.Text()
		}

		posts = append(posts, post)
	}

	return posts, nil
}

type getPostsArgs struct {
	AfterID    int       `qstring:"after_id,omitempty"`
	BeforeID   int       `qstring:"before_id,omitempty"`
	AfterTime  time.Time `qstring:"after_time,omitempty"`
	BeforeTime time.Time `qstring:"before_time,omitempty"`
	People     []int     `qstring:"people,omitempty"`
	Sort       string    `qstring:"sort,omitempty"`
	Limit      int       `qstring:"limit,omitempty"`
}

func getPosts(db *sql.DB, args getPostsArgs) ([]UIStatus, error) {
	qb := sqlbuilder.
		Select(postsTable.LeftOuterJoin(peopleTable, peopleTable.C("feed_url").Eq(postsTable.C("feed_url")))).
		Columns(
			postsTable.C("ROWID"),
			postsTable.C("feed_url"),
			postsTable.C("raw_entry"),
			peopleTable.C("name"),
			peopleTable.C("display_name"),
			peopleTable.C("email"),
		)

	if args.Limit > 0 && args.Limit <= 75 {
		qb = qb.Limit(args.Limit)
	} else {
		qb = qb.Limit(75)
	}

	switch args.Sort {
	case "-created_at":
		qb = qb.OrderBy(false, postsTable.C("created_at"))
	case "created_at":
		qb = qb.OrderBy(true, postsTable.C("created_at"))
	case "-id":
		qb = qb.OrderBy(false, postsTable.C("ROWID"))
	default:
		qb = qb.OrderBy(true, postsTable.C("ROWID"))
	}

	var conditions []sqlbuilder.Condition

	if args.AfterID != 0 {
		conditions = append(conditions, postsTable.C("ROWID").Gt(args.AfterID))
	}
	if args.BeforeID != 0 {
		conditions = append(conditions, postsTable.C("ROWID").Lt(args.BeforeID))
	}
	if !args.AfterTime.IsZero() {
		conditions = append(conditions, postsTable.C("created_at").Gt(args.AfterTime))
	}
	if !args.BeforeTime.IsZero() {
		conditions = append(conditions, postsTable.C("created_at").Gt(args.BeforeTime))
	}

	if len(args.People) > 0 {
		var a []interface{}
		for _, id := range args.People {
			a = append(a, id)
		}

		conditions = append(conditions, peopleTable.C("ROWID").In(a...))
	}

	if len(conditions) > 0 {
		qb = qb.Where(sqlbuilder.And(conditions...))
	}

	q, vars, err := qb.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "getPosts")
	}

	rows, err := db.Query(q, vars...)
	if err != nil {
		return nil, errors.Wrap(err, "getPosts")
	}
	defer rows.Close()

	var posts []UIStatus
	for rows.Next() {
		var id int
		var feedURL, rawEntry string
		var name, displayName, email sql.NullString
		if err := rows.Scan(&id, &feedURL, &rawEntry, &name, &displayName, &email); err != nil {
			return nil, errors.Wrap(err, "getPosts")
		}

		var entry AtomEntry
		if err := json.Unmarshal([]byte(rawEntry), &entry); err != nil {
			return nil, errors.Wrap(err, "getPosts")
		}

		post := UIStatus{ID: id}

		if name.Valid {
			post.AuthorAcct = email.String
			post.AuthorName = name.String
		}

		if t, err := time.Parse(time.RFC3339, entry.Published); err == nil {
			post.Time = t
		}

		if entry.Content != nil {
			post.ContentHTML = entry.Content.HTML()
			post.ContentText = entry.Content.Text()
		}

		posts = append(posts, post)
	}

	return posts, nil
}
