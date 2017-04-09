package main

import (
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
)

type PubSubSQLState struct{ DB *sql.DB }

func NewPubSubSQLState(db *sql.DB) *PubSubSQLState {
	return &PubSubSQLState{DB: db}
}

func (s *PubSubSQLState) All() ([]PubSubSubscription, error) {
	var a []PubSubSubscription

	rows, err := s.DB.Query("select id, hub, topic, callback_url, created_at, updated_at, expires_at from pubsub_state")
	if err != nil {
		return nil, errors.Wrap(err, "PubSubSQLState.All")
	}
	defer rows.Close()

	for rows.Next() {
		var id, hub, topic, callbackURL string
		var createdAt, updatedAt time.Time
		var expiresAt *time.Time
		if err := rows.Scan(&id, &hub, &topic, &callbackURL, &createdAt, &updatedAt, &expiresAt); err != nil {
			return nil, errors.Wrap(err, "PubSubSQLState.All")
		}

		a = append(a, PubSubSubscription{
			ID:          id,
			Hub:         hub,
			Topic:       topic,
			CallbackURL: callbackURL,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
			ExpiresAt:   expiresAt,
		})
	}

	return a, nil
}

func (s *PubSubSQLState) Add(hub, topic, baseURL string) (*PubSubSubscription, string, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, "", errors.Wrap(err, "PubSubSQLState.Add")
	}
	defer tx.Rollback()

	var createdAt, updatedAt time.Time
	var expiresAt *time.Time
	var id, callbackURL, oldCallbackURL string
	if err := tx.QueryRow("select id, callback_url, created_at, updated_at, expires_at from pubsub_state where hub = $1 and topic = $2", hub, topic).Scan(&id, &oldCallbackURL, &createdAt, &updatedAt, &expiresAt); err != nil {
		if err != sql.ErrNoRows {
			return nil, "", errors.Wrap(err, "PubSubSQLState.Add")
		}

		id = uuid.NewV4().String()
		createdAt = time.Now()
		updatedAt = createdAt
		callbackURL = baseURL + "/" + id

		if _, err := tx.Exec("insert into pubsub_state (id, hub, topic, callback_url, created_at, updated_at) values ($1, $2, $3, $4, $5, $6)", id, hub, topic, callbackURL, createdAt, updatedAt); err != nil {
			return nil, "", errors.Wrap(err, "PubSubSQLState.Add")
		}
	} else {
		callbackURL = oldCallbackURL

		if newCallbackURL := baseURL + "/" + id; newCallbackURL != oldCallbackURL {
			callbackURL = newCallbackURL
			updatedAt = time.Now()

			if _, err := tx.Exec("update pubsub_state set callback_url = $1, updated_at = $2, expires_at = NULL where id = $3", callbackURL, updatedAt, id); err != nil {
				return nil, "", errors.Wrap(err, "PubSubSQLState.Add")
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, "", errors.Wrap(err, "PubSubSQLState.Add")
	}

	return &PubSubSubscription{
		ID:          id,
		Hub:         hub,
		Topic:       topic,
		CallbackURL: callbackURL,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		ExpiresAt:   expiresAt,
	}, oldCallbackURL, nil
}

func (s *PubSubSQLState) Get(hub, topic string) (*PubSubSubscription, error) {
	var v PubSubSubscription

	if err := s.DB.QueryRow("select id, hub, topic, callback_url, created_at, updated_at, expires_at from pubsub_state where hub = $1 and topic = $2", hub, topic).Scan(&v.ID, &v.Hub, &v.Topic, &v.CallbackURL, &v.CreatedAt, &v.ExpiresAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, errors.Wrap(err, "PubSubSQLState.Get")
	}

	return &v, nil
}

func (s *PubSubSQLState) GetByID(id string) (*PubSubSubscription, error) {
	var v PubSubSubscription

	if err := s.DB.QueryRow("select id, hub, topic, callback_url, created_at, updated_at, expires_at from pubsub_state where id = $1", id).Scan(&v.ID, &v.Hub, &v.Topic, &v.CallbackURL, &v.CreatedAt, &v.UpdatedAt, &v.ExpiresAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, errors.Wrap(err, "PubSubSQLState.GetByID")
	}

	return &v, nil
}

func (s *PubSubSQLState) Set(hub, topic string, updatedAt, expiresAt time.Time) error {
	if _, err := s.DB.Exec("update pubsub_state set updated_at = $1, expires_at = $2 where hub = $3 and topic = $4", updatedAt, expiresAt, hub, topic); err != nil {
		return errors.Wrap(err, "PubSubSQLState.Set")
	}

	return nil
}

func (s *PubSubSQLState) Del(hub, topic string) error {
	if _, err := s.DB.Exec("delete from pubsub_state where hub = $1 and topic = $2", hub, topic); err != nil {
		return errors.Wrap(err, "PubSubSQLState.Del")
	}

	return nil
}
