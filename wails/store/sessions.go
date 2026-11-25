package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/maxachis/clean-tracker/wails/model"
)

const sessionColumns = `id, length_seconds, min_action_seconds, created_at, started_at, ended_at, end_status`

func scanSession(row interface{ Scan(...any) error }) (model.Session, error) {
	var (
		s                      model.Session
		created                string
		started, ended, status sql.NullString
	)
	if err := row.Scan(&s.ID, &s.LengthSeconds, &s.MinActionSeconds, &created, &started, &ended, &status); err != nil {
		return model.Session{}, err
	}
	var err error
	if s.CreatedAt, err = parseTime(created); err != nil {
		return model.Session{}, err
	}
	if s.StartedAt, err = parseNullTime(started); err != nil {
		return model.Session{}, err
	}
	if s.EndedAt, err = parseNullTime(ended); err != nil {
		return model.Session{}, err
	}
	if status.Valid {
		end := model.SessionStatus(status.String)
		s.EndStatus = &end
	}
	return s, nil
}

// querySession runs a single-row session query, mapping a missing row to ErrNotFound.
func (s *Store) querySession(what, where string, args ...any) (model.Session, error) {
	sess, err := scanSession(s.DB.QueryRow(`SELECT `+sessionColumns+` FROM sessions WHERE `+where, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return model.Session{}, fmt.Errorf("%s: %w", what, ErrNotFound)
	}
	if err != nil {
		return model.Session{}, fmt.Errorf("%s: %w", what, err)
	}
	return sess, nil
}

func (s *Store) InsertSession(sess model.Session) error {
	_, err := s.DB.Exec(
		`INSERT INTO sessions (`+sessionColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		sess.ID, sess.LengthSeconds, sess.MinActionSeconds, formatTime(sess.CreatedAt),
		nullableTime(sess.StartedAt), nullableTime(sess.EndedAt), nullableStatus(sess.EndStatus),
	)
	if err != nil {
		return fmt.Errorf("insert session %s: %w", sess.ID, err)
	}
	return nil
}

func nullableStatus(st *model.SessionStatus) any {
	if st == nil {
		return nil
	}
	return string(*st)
}

func (s *Store) GetSession(id string) (model.Session, error) {
	return s.querySession("get session "+id, `id = ?`, id)
}

// GetDraft returns the session that has not been started, or ErrNotFound.
func (s *Store) GetDraft() (model.Session, error) {
	return s.querySession("get draft", `started_at IS NULL ORDER BY created_at DESC LIMIT 1`)
}

// GetRunning returns the started session that has not ended, or ErrNotFound.
func (s *Store) GetRunning() (model.Session, error) {
	return s.querySession("get running session", `started_at IS NOT NULL AND end_status IS NULL ORDER BY started_at DESC LIMIT 1`)
}

func (s *Store) UpdateSessionLength(id string, lengthSeconds int) error {
	return s.execOne("update session length "+id, `UPDATE sessions SET length_seconds = ? WHERE id = ?`, lengthSeconds, id)
}

// UpdateSessionMinAction sets the session's own copy of the per-action minimum.
func (s *Store) UpdateSessionMinAction(id string, seconds int) error {
	return s.execOne("update session min action "+id, `UPDATE sessions SET min_action_seconds = ? WHERE id = ?`, seconds, id)
}

func (s *Store) MarkSessionStarted(id string, at time.Time) error {
	return s.execOne("start session "+id, `UPDATE sessions SET started_at = ? WHERE id = ?`, formatTime(at), id)
}

func (s *Store) MarkSessionEnded(id string, at time.Time, status model.SessionStatus) error {
	return s.execOne("end session "+id,
		`UPDATE sessions SET ended_at = ?, end_status = ? WHERE id = ?`, formatTime(at), string(status), id)
}

// execOne runs a statement that must touch exactly one row.
func (s *Store) execOne(what, query string, args ...any) error {
	res, err := s.DB.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", what, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", what, err)
	}
	if n == 0 {
		return fmt.Errorf("%s: %w", what, ErrNotFound)
	}
	return nil
}
