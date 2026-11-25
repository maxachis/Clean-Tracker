package migrations

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/maxachis/clean-tracker/wails/model"

	_ "modernc.org/sqlite"
)

func TestRun_IsIdempotentAndSeedsSettings(t *testing.T) {
	// Run executes on every app start, so a second pass must neither fail
	// nor overwrite settings the user has changed.
	dsns := map[string]string{
		"memory":      ":memory:",
		"file-backed": filepath.Join(t.TempDir(), "fresh.db"),
	}
	for name, dsn := range dsns {
		t.Run(name, func(t *testing.T) {
			db, err := sql.Open("sqlite", dsn)
			if err != nil {
				t.Fatalf("open: %v", err)
			}
			db.SetMaxOpenConns(1)
			t.Cleanup(func() { _ = db.Close() })

			if err := Run(db); err != nil {
				t.Fatalf("first run: %v", err)
			}
			if _, err := db.Exec(`UPDATE settings SET last_session_minutes = 45`); err != nil {
				t.Fatalf("change setting: %v", err)
			}
			if err := Run(db); err != nil {
				t.Fatalf("second run: %v", err)
			}

			var minutes int
			var sound bool
			if err := db.QueryRow(`SELECT last_session_minutes, miss_sound FROM settings WHERE id = 1`).Scan(&minutes, &sound); err != nil {
				t.Fatalf("query: %v", err)
			}
			if minutes != 45 || sound != model.DefaultMissSound {
				t.Fatalf("settings = %d/%v, want 45/%v", minutes, sound, model.DefaultMissSound)
			}
		})
	}
}

func TestRun_RejectsChoreCountBelowOne(t *testing.T) {
	// The schema itself enforces target_count >= 1 as a backstop to service validation.
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()
	if err := Run(db); err != nil {
		t.Fatalf("run: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO sessions (id, length_seconds, created_at) VALUES ('s', 60, 'x')`); err != nil {
		t.Fatalf("insert session: %v", err)
	}
	_, err = db.Exec(`INSERT INTO chores (id, session_id, name, target_count, position) VALUES ('c', 's', 'n', 0, 0)`)
	if err == nil {
		t.Fatal("expected CHECK constraint failure for target_count = 0")
	}
}
