package main

import (
	"database/sql"
	"encoding/json"
	"time"
)

func savePerson(db *sql.DB, feedURL string, author *AtomAuthor) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var name, displayName, email, summary, note string
	if err := tx.QueryRow("select name, display_name, email, summary, note from people where feed_url = $1", feedURL).Scan(&name, &displayName, &email, &summary, &note); err != nil {
		if err != sql.ErrNoRows {
			return err
		}

		if _, err := tx.Exec("insert into people (feed_url, first_seen, name, display_name, email, summary, note) values ($1, $2, $3, $4, $5, $6, $7)", feedURL, time.Now(), author.Name, author.DisplayName, author.Email, author.Summary, author.Note); err != nil {
			return err
		}
	} else {
		if name == author.Name && displayName == author.DisplayName && email == author.Email && summary == author.Summary && note == author.Note {
			return nil
		}

		if _, err := tx.Exec("update people set name = $1, display_name = $2, email = $3, summary = $4, note = $5 where feed_url = $6", author.Name, author.DisplayName, author.Email, author.Summary, author.Note, feedURL); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func saveEntry(db *sql.DB, feedURL string, entry *AtomEntry) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists int
	if err := tx.QueryRow("select count(1) from posts where feed_url = $1 and id = $2", feedURL, entry.ID).Scan(&exists); err != nil {
		return err
	}

	if exists > 0 {
		return nil
	}

	d, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	if _, err := tx.Exec("insert into posts (feed_url, id, created_at, raw_entry) values ($1, $2, $3, $4)", feedURL, entry.ID, entry.Published, string(d)); err != nil {
		return err
	}

	return tx.Commit()
}

func getPublicTimeline(db *sql.DB, offset, limit int) ([]UIStatus, error) {
	rows, err := db.Query("select posts.ROWID, posts.feed_url, posts.raw_entry, people.name, people.display_name, people.email from posts left join people on people.feed_url = posts.feed_url order by posts.created_at desc limit $1 offset $2", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []UIStatus
	for rows.Next() {
		var id, feedURL, rawEntry string
		var name, displayName, email sql.NullString
		if err := rows.Scan(&id, &feedURL, &rawEntry, &name, &displayName, &email); err != nil {
			return nil, err
		}

		var entry AtomEntry
		if err := json.Unmarshal([]byte(rawEntry), &entry); err != nil {
			return nil, err
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
