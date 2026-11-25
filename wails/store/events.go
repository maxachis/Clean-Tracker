package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/maxachis/clean-tracker/wails/model"
)

// InsertEvent appends a fact to a session's log. choreID is empty except for done events.
func (s *Store) InsertEvent(sessionID string, kind model.EventKind, choreID string, at time.Time) error {
	var chore sql.NullString
	if choreID != "" {
		chore = sql.NullString{String: choreID, Valid: true}
	}
	_, err := s.DB.Exec(
		`INSERT INTO events (session_id, kind, chore_id, at) VALUES (?, ?, ?, ?)`,
		sessionID, string(kind), chore, formatTime(at),
	)
	if err != nil {
		return fmt.Errorf("insert %s event for %s: %w", kind, sessionID, err)
	}
	return nil
}

// ListEvents returns a session's events in insertion order.
func (s *Store) ListEvents(sessionID string) ([]model.Event, error) {
	rows, err := s.DB.Query(
		`SELECT id, kind, chore_id, at FROM events WHERE session_id = ? ORDER BY id`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list events %s: %w", sessionID, err)
	}
	defer rows.Close()
	events := []model.Event{}
	for rows.Next() {
		var (
			e     model.Event
			kind  string
			chore sql.NullString
			at    string
		)
		if err := rows.Scan(&e.ID, &kind, &chore, &at); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		e.Kind = model.EventKind(kind)
		e.ChoreID = chore.String
		if e.At, err = parseTime(at); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
