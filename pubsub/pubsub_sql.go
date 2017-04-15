package pubsub

import (
	"database/sql"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
)

type SQLiteState struct {
	m  sync.Mutex
	DB *sql.DB
}

func NewSQLiteState(db *sql.DB) *SQLiteState {
	return &SQLiteState{DB: db}
}

func (s *SQLiteState) All() ([]Subscription, error) {
	s.m.Lock()
	defer s.m.Unlock()

	var a []Subscription

	rows, err := s.DB.Query("select id, hub, topic, callback_url, created_at, updated_at, expires_at from pubsub_state")
	if err != nil {
		return nil, errors.Wrap(err, "SQLiteState.All")
	}
	defer rows.Close()

	for rows.Next() {
		var id, hub, topic, callbackURL string
		var createdAt, updatedAt time.Time
		var expiresAt *time.Time
		if err := rows.Scan(&id, &hub, &topic, &callbackURL, &createdAt, &updatedAt, &expiresAt); err != nil {
			return nil, errors.Wrap(err, "SQLiteState.All")
		}

		a = append(a, Subscription{
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

func (s *SQLiteState) Add(hub, topic, baseURL string) (*Subscription, string, error) {
	s.m.Lock()
	defer s.m.Unlock()

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, "", errors.Wrap(err, "SQLiteState.Add: couldn't open transaction")
	}
	defer tx.Rollback()

	var createdAt, updatedAt time.Time
	var expiresAt *time.Time
	var id, callbackURL, oldCallbackURL string
	if err := tx.QueryRow("select id, callback_url, created_at, updated_at, expires_at from pubsub_state where hub = $1 and topic = $2", hub, topic).Scan(&id, &oldCallbackURL, &createdAt, &updatedAt, &expiresAt); err != nil {
		if err != sql.ErrNoRows {
			return nil, "", errors.Wrap(err, "SQLiteState.Add: couldn't select subscription record")
		}

		id = uuid.NewV4().String()
		createdAt = time.Now()
		updatedAt = createdAt
		callbackURL = baseURL + "/" + id

		if _, err := tx.Exec("insert into pubsub_state (id, hub, topic, callback_url, created_at, updated_at) values ($1, $2, $3, $4, $5, $6)", id, hub, topic, callbackURL, createdAt, updatedAt); err != nil {
			return nil, "", errors.Wrap(err, "SQLiteState.Add: couldn't insert subscription record")
		}
	} else {
		callbackURL = oldCallbackURL

		if newCallbackURL := baseURL + "/" + id; newCallbackURL != oldCallbackURL {
			callbackURL = newCallbackURL
			updatedAt = time.Now()

			if _, err := tx.Exec("update pubsub_state set callback_url = $1, updated_at = $2, expires_at = NULL where id = $3", callbackURL, updatedAt, id); err != nil {
				return nil, "", errors.Wrap(err, "SQLiteState.Add: couldn't update subscription record")
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, "", errors.Wrap(err, "SQLiteState.Add: couldn't close transaction")
	}

	return &Subscription{
		ID:          id,
		Hub:         hub,
		Topic:       topic,
		CallbackURL: callbackURL,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		ExpiresAt:   expiresAt,
	}, oldCallbackURL, nil
}

func (s *SQLiteState) Get(hub, topic string) (*Subscription, error) {
	s.m.Lock()
	defer s.m.Unlock()

	var v Subscription

	if err := s.DB.QueryRow("select id, hub, topic, callback_url, created_at, updated_at, expires_at from pubsub_state where hub = $1 and topic = $2", hub, topic).Scan(&v.ID, &v.Hub, &v.Topic, &v.CallbackURL, &v.CreatedAt, &v.ExpiresAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, errors.Wrap(err, "SQLiteState.Get: couldn't select subscription record")
	}

	return &v, nil
}

func (s *SQLiteState) GetByID(id string) (*Subscription, error) {
	s.m.Lock()
	defer s.m.Unlock()

	var v Subscription

	if err := s.DB.QueryRow("select id, hub, topic, callback_url, created_at, updated_at, expires_at from pubsub_state where id = $1", id).Scan(&v.ID, &v.Hub, &v.Topic, &v.CallbackURL, &v.CreatedAt, &v.UpdatedAt, &v.ExpiresAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, errors.Wrap(err, "SQLiteState.GetByID: couldn't select subscription record")
	}

	return &v, nil
}

func (s *SQLiteState) Set(hub, topic string, updatedAt, expiresAt time.Time) error {
	s.m.Lock()
	defer s.m.Unlock()

	if _, err := s.DB.Exec("update pubsub_state set updated_at = $1, expires_at = $2 where hub = $3 and topic = $4", updatedAt, expiresAt, hub, topic); err != nil {
		return errors.Wrap(err, "SQLiteState.Set: couldn't update subscription record")
	}

	return nil
}

func (s *SQLiteState) Del(hub, topic string) error {
	s.m.Lock()
	defer s.m.Unlock()

	if _, err := s.DB.Exec("delete from pubsub_state where hub = $1 and topic = $2", hub, topic); err != nil {
		return errors.Wrap(err, "SQLiteState.Del: couldn't delete subscription record")
	}

	return nil
}
