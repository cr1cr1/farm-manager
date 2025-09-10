package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const (
	defaultDSN = "file:./data/app.db?cache=shared&mode=rwc"
)

// dsnFromEnv returns the SQLite DSN from env or a sane default.
func dsnFromEnv() string {
	if v := os.Getenv("SQLITE_DSN"); v != "" {
		return v
	}
	return defaultDSN
}

// ensureDataDir tries to create the directory that contains the SQLite file if the DSN points to a file: URL.
func ensureDataDir(dsn string) error {
	// Only handle file: DSNs. Others (e.g., :memory:) are ignored.
	raw := dsn
	if !strings.HasPrefix(raw, "file:") {
		return nil
	}
	path := strings.TrimPrefix(raw, "file:")
	// Strip query params
	if i := strings.IndexByte(path, '?'); i >= 0 {
		path = path[:i]
	}
	if path == "" || path == ":memory:" {
		return nil
	}
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

// Open opens (and creates if needed) the SQLite database with required PRAGMAs.
func Open(ctx context.Context) (*sql.DB, error) {
	dsn := dsnFromEnv()
	if err := ensureDataDir(dsn); err != nil {
		return nil, fmt.Errorf("ensure data dir: %w", err)
	}

	// modernc.org/sqlite registers as driver name "sqlite".
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Reasonable pool defaults for SQLite without WAL.
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)

	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctxPing); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	// Enforce pragma settings per connection.
	if err := applyPragmas(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func applyPragmas(ctx context.Context, db *sql.DB) error {
	// busy_timeout in milliseconds; keep conservative default.
	stmts := []string{
		"PRAGMA foreign_keys = ON;",
		"PRAGMA busy_timeout = 5000;",
	}
	for _, s := range stmts {
		if _, err := db.ExecContext(ctx, s); err != nil {
			return fmt.Errorf("apply pragma %q: %w", s, err)
		}
	}
	return nil
}

// Close closes the database, returning the first error encountered.
func Close(db *sql.DB) error {
	if db == nil {
		return nil
	}
	if err := db.Close(); err != nil {
		return err
	}
	return nil
}

// WithDB is a small helper to open a DB, run a function, and close it.
func WithDB(ctx context.Context, fn func(*sql.DB) error) error {
	db, err := Open(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = Close(db)
	}()
	if fn == nil {
		return errors.New("nil db function")
	}
	return fn(db)
}
