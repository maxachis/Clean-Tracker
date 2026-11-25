package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/maxachis/clean-tracker/wails/migrations"

	_ "modernc.org/sqlite"
)

const (
	// busyTimeoutMS is how long a write waits for another connection's write
	// to finish before giving up with "database is locked".
	busyTimeoutMS = 5000

	memoryDSN = ":memory:"
	dbSubdir  = "clean-tracker"
	dbFile    = "clean-tracker.db"
	dirMode   = 0o755
)

// ErrNotFound is returned when a requested row does not exist.
var ErrNotFound = errors.New("not found")

type Store struct {
	DB *sql.DB
}

// Open returns a Store backed by the given DSN. Pass ":memory:" for tests.
func Open(dsn string) (*Store, error) {
	db, err := sql.Open("sqlite", withPragmas(dsn))
	if err != nil {
		return nil, fmt.Errorf("open sqlite %s: %w", dsn, err)
	}
	if dsn == memoryDSN {
		// Every pooled connection would otherwise get its own empty database.
		db.SetMaxOpenConns(1)
	}
	if err := migrations.Run(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{DB: db}, nil
}

// withPragmas adds the busy_timeout and foreign_keys pragmas to the DSN so
// the driver sets them on every pooled connection, not just the first.
func withPragmas(dsn string) string {
	separator := "?"
	if strings.Contains(dsn, "?") {
		separator = "&"
	}
	return fmt.Sprintf("%s%s_pragma=busy_timeout(%d)&_pragma=foreign_keys(1)", dsn, separator, busyTimeoutMS)
}

// DefaultDBPath returns the per-user location for the clean-tracker DB file,
// creating the parent directory if it does not exist.
func DefaultDBPath() (string, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config dir: %w", err)
	}
	dir := filepath.Join(cfg, dbSubdir)
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", dir, err)
	}
	return filepath.Join(dir, dbFile), nil
}

func (s *Store) Close() error {
	return s.DB.Close()
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTime(raw string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse time %q: %w", raw, err)
	}
	return t.UTC(), nil
}

func parseNullTime(raw sql.NullString) (*time.Time, error) {
	if !raw.Valid {
		return nil, nil
	}
	t, err := parseTime(raw.String)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func nullableTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return formatTime(*t)
}
