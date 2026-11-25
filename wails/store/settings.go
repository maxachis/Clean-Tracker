package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/maxachis/clean-tracker/wails/model"
)

// GetSettings returns the singleton settings row.
func (s *Store) GetSettings() (model.Settings, error) {
	var st model.Settings
	err := s.DB.QueryRow(`SELECT last_session_minutes, miss_sound FROM settings WHERE id = 1`).
		Scan(&st.LastSessionMinutes, &st.MissSound)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Settings{}, ErrNotFound
	}
	if err != nil {
		return model.Settings{}, fmt.Errorf("get settings: %w", err)
	}
	return st, nil
}

func (s *Store) UpdateSettings(st model.Settings) error {
	return s.execOne("update settings",
		`UPDATE settings SET last_session_minutes = ?, miss_sound = ? WHERE id = 1`,
		st.LastSessionMinutes, st.MissSound)
}
