package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// findMigrationsDir attempts to locate the db/migrations directory from various relative locations
// to support running from different working directories (e.g., package tests).
func findMigrationsDir() (string, error) {
	candidates := []string{
		"db/migrations",
		filepath.Join("..", "..", "db", "migrations"),
		filepath.Join("..", "db", "migrations"),
		filepath.Join("..", "..", "..", "db", "migrations"),
	}
	for _, d := range candidates {
		if info, err := os.Stat(d); err == nil && info.IsDir() {
			return d, nil
		}
	}
	return "", fmt.Errorf("db/migrations directory not found (looked in: %v)", candidates)
}

// Migrate executes all .sql files in db/migrations in lexicographic order.
// Each file may contain multiple SQL statements. Use IF NOT EXISTS to keep it idempotent.
func Migrate(ctx context.Context, db *sql.DB) error {
	dir, err := findMigrationsDir()
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %q: %w", dir, err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(strings.ToLower(name), ".sql") {
			files = append(files, filepath.Join(dir, name))
		}
	}
	sort.Strings(files)

	for _, p := range files {
		sqlBytes, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("read migration %q: %w", p, err)
		}
		sqlText := strings.TrimSpace(string(sqlBytes))
		if sqlText == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, sqlText); err != nil {
			return fmt.Errorf("exec migration %q: %w", p, err)
		}
	}

	return nil
}
