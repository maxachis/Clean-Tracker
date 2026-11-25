package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/maxachis/clean-tracker/wails/model"
)

const choreColumns = `id, session_id, name, target_count, position`

func (s *Store) InsertChore(c model.Chore) error {
	_, err := s.DB.Exec(
		`INSERT INTO chores (`+choreColumns+`) VALUES (?, ?, ?, ?, ?)`,
		c.ID, c.SessionID, c.Name, c.TargetCount, c.Position,
	)
	if err != nil {
		return fmt.Errorf("insert chore %s: %w", c.ID, err)
	}
	return nil
}

func (s *Store) GetChore(id string) (model.Chore, error) {
	var c model.Chore
	err := s.DB.QueryRow(`SELECT `+choreColumns+` FROM chores WHERE id = ?`, id).
		Scan(&c.ID, &c.SessionID, &c.Name, &c.TargetCount, &c.Position)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Chore{}, fmt.Errorf("get chore %s: %w", id, ErrNotFound)
	}
	if err != nil {
		return model.Chore{}, fmt.Errorf("get chore %s: %w", id, err)
	}
	return c, nil
}

// ListChores returns a session's chores in list order.
func (s *Store) ListChores(sessionID string) ([]model.Chore, error) {
	rows, err := s.DB.Query(
		`SELECT `+choreColumns+` FROM chores WHERE session_id = ? ORDER BY position, id`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list chores %s: %w", sessionID, err)
	}
	defer rows.Close()
	chores := []model.Chore{}
	for rows.Next() {
		var c model.Chore
		if err := rows.Scan(&c.ID, &c.SessionID, &c.Name, &c.TargetCount, &c.Position); err != nil {
			return nil, fmt.Errorf("scan chore: %w", err)
		}
		chores = append(chores, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list chores %s: %w", sessionID, err)
	}
	return chores, nil
}

func (s *Store) UpdateChore(id, name string, count int) error {
	return s.execOne("update chore "+id, `UPDATE chores SET name = ?, target_count = ? WHERE id = ?`, name, count, id)
}

func (s *Store) DeleteChore(id string) error {
	return s.execOne("delete chore "+id, `DELETE FROM chores WHERE id = ?`, id)
}

func (s *Store) DeleteChoresForSession(sessionID string) error {
	if _, err := s.DB.Exec(`DELETE FROM chores WHERE session_id = ?`, sessionID); err != nil {
		return fmt.Errorf("delete chores for %s: %w", sessionID, err)
	}
	return nil
}

// SetChorePositions assigns positions 0..n-1 in the order of ids.
func (s *Store) SetChorePositions(ids []string) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("begin reorder: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // no-op after a successful commit
	for position, id := range ids {
		if _, err := tx.Exec(`UPDATE chores SET position = ? WHERE id = ?`, position, id); err != nil {
			return fmt.Errorf("reorder chore %s: %w", id, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit reorder: %w", err)
	}
	return nil
}

// RecentChoreNames returns distinct chore names from started sessions, the
// most recently created session first, up to limit.
func (s *Store) RecentChoreNames(limit int) ([]string, error) {
	rows, err := s.DB.Query(
		`SELECT c.name FROM chores c JOIN sessions s ON s.id = c.session_id
		 WHERE s.started_at IS NOT NULL
		 GROUP BY c.name ORDER BY MAX(s.created_at) DESC, MIN(c.position) LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("recent chore names: %w", err)
	}
	defer rows.Close()
	names := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan chore name: %w", err)
		}
		names = append(names, name)
	}
	return names, rows.Err()
}
