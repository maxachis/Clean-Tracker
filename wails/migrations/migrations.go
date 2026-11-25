package migrations

import (
	"database/sql"
	_ "embed"
	"fmt"

	"github.com/maxachis/clean-tracker/wails/model"
)

//go:embed schema.sql
var schema string

// Run applies the embedded schema, adds columns introduced after the first
// release, and seeds default settings. Every step is
// idempotent, so it's safe to invoke on every startup.
func Run(db *sql.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	if err := addMinActionColumns(db); err != nil {
		return err
	}
	_, err := db.Exec(
		`INSERT OR IGNORE INTO settings (id, last_session_minutes, miss_sound, min_action_seconds) VALUES (1, ?, ?, ?)`,
		model.DefaultLastSessionMinutes, model.DefaultMissSound, model.DefaultMinActionSeconds,
	)
	if err != nil {
		return fmt.Errorf("seed settings: %w", err)
	}
	return nil
}

// addMinActionColumns upgrades databases created before the per-action minimum
// existed; the CREATE TABLEs in the schema only cover fresh databases.
func addMinActionColumns(db *sql.DB) error {
	definition := fmt.Sprintf("INTEGER NOT NULL DEFAULT %d", model.DefaultMinActionSeconds)
	for _, table := range []string{"sessions", "settings"} {
		if err := addColumnIfMissing(db, table, "min_action_seconds", definition); err != nil {
			return err
		}
	}
	return nil
}

// addColumnIfMissing adds a column to an existing table unless it is already
// there. SQLite has no ADD COLUMN IF NOT EXISTS.
func addColumnIfMissing(db *sql.DB, table, column, definition string) error {
	exists, err := columnExists(db, table, column)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if _, err := db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition)); err != nil {
		return fmt.Errorf("add column %s.%s: %w", table, column, err)
	}
	return nil
}

func columnExists(db *sql.DB, table, column string) (bool, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, fmt.Errorf("inspect table %s: %w", table, err)
	}
	defer func() { _ = rows.Close() }() // read-only cursor; Err() below reports failures
	for rows.Next() {
		var (
			cid        int
			name, typ  string
			notNull    int
			defaultVal sql.NullString
			pk         int
		)
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultVal, &pk); err != nil {
			return false, fmt.Errorf("scan table_info %s: %w", table, err)
		}
		if name == column {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("inspect table %s: %w", table, err)
	}
	return false, nil
}
