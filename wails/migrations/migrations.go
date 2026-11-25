package migrations

import (
	"database/sql"
	_ "embed"
	"fmt"

	"github.com/maxachis/clean-tracker/wails/model"
)

//go:embed schema.sql
var schema string

// Run applies the embedded schema and seeds default settings. Every step is
// idempotent, so it's safe to invoke on every startup.
func Run(db *sql.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	_, err := db.Exec(
		`INSERT OR IGNORE INTO settings (id, last_session_minutes, miss_sound) VALUES (1, ?, ?)`,
		model.DefaultLastSessionMinutes, model.DefaultMissSound,
	)
	if err != nil {
		return fmt.Errorf("seed settings: %w", err)
	}
	return nil
}
