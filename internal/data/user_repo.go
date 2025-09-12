package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/cr1cr1/farm-manager/internal/domain"
)

var (
	// ErrNotFound is returned when a user cannot be found.
	ErrNotFound = sql.ErrNoRows
)

// UserRepo defines minimal operations for authentication.
type UserRepo interface {
	// Count returns the number of non-deleted users.
	Count(ctx context.Context) (int64, error)
	// FindByUsername returns a user by username (excluding soft-deleted).
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	// Create inserts a new user with a pre-hashed password.
	Create(ctx context.Context, u *domain.User) (int64, error)
	// UpdatePassword updates the password hash and optionally clears force flag.
	UpdatePassword(ctx context.Context, userID int64, newHash string, forceChange bool) error
	// UpdateTheme updates the user's theme preference.
	UpdateTheme(ctx context.Context, userID int64, theme int) error
	// SoftDelete marks the user as deleted.
	SoftDelete(ctx context.Context, userID int64, deletedAt time.Time) error
}

type SQLiteUserRepo struct {
	DB *sql.DB
}

func NewSQLiteUserRepo(db *sql.DB) *SQLiteUserRepo {
	return &SQLiteUserRepo{DB: db}
}

func (r *SQLiteUserRepo) Count(ctx context.Context) (int64, error) {
	const q = `SELECT COUNT(1) FROM users WHERE deleted_at IS NULL`
	var n int64
	if err := r.DB.QueryRowContext(ctx, q).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *SQLiteUserRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	const q = `
SELECT id, username, password_hash, force_password_change, theme, created_at, updated_at, deleted_at, created_by, updated_by
FROM users
WHERE username = ? AND (deleted_at IS NULL)
LIMIT 1`
	row := r.DB.QueryRowContext(ctx, q, username)
	u, err := scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

func (r *SQLiteUserRepo) Create(ctx context.Context, u *domain.User) (int64, error) {
	// created_at/updated_at defaulted by schema; we supply fields explicitly for clarity.
	const q = `
INSERT INTO users (username, password_hash, force_password_change, created_at, updated_at, created_by, updated_by)
VALUES (?, ?, ?, strftime('%Y-%m-%dT%H:%M:%fZ','now'), strftime('%Y-%m-%dT%H:%M:%fZ','now'), ?, ?)`
	var createdBy, updatedBy interface{}
	if u.Audit.CreatedBy != nil {
		createdBy = *u.Audit.CreatedBy
	}
	if u.Audit.UpdatedBy != nil {
		updatedBy = *u.Audit.UpdatedBy
	}
	res, err := r.DB.ExecContext(ctx, q, u.Username, u.PasswordHash, boolToInt(u.ForcePasswordChange), createdBy, updatedBy)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *SQLiteUserRepo) UpdatePassword(ctx context.Context, userID int64, newHash string, forceChange bool) error {
	const q = `
UPDATE users
SET password_hash = ?, force_password_change = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
WHERE id = ? AND (deleted_at IS NULL)`
	_, err := r.DB.ExecContext(ctx, q, newHash, boolToInt(forceChange), userID)
	return err
}

func (r *SQLiteUserRepo) UpdateTheme(ctx context.Context, userID int64, theme int) error {
	const q = `
UPDATE users
SET theme = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
WHERE id = ? AND (deleted_at IS NULL)`
	_, err := r.DB.ExecContext(ctx, q, theme, userID)
	return err
}

func (r *SQLiteUserRepo) SoftDelete(ctx context.Context, userID int64, deletedAt time.Time) error {
	const q = `
UPDATE users
SET deleted_at = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
WHERE id = ? AND (deleted_at IS NULL)`
	_, err := r.DB.ExecContext(ctx, q, deletedAt.UTC().Format(time.RFC3339Nano), userID)
	return err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(rs rowScanner) (*domain.User, error) {
	var (
		id           int64
		username     string
		passwordHash string
		force        int
		theme        int
		createdAtStr string
		updatedAtStr string
		deletedAtStr sql.NullString
		createdByStr sql.NullString
		updatedByStr sql.NullString
	)
	if err := rs.Scan(&id, &username, &passwordHash, &force, &theme, &createdAtStr, &updatedAtStr, &deletedAtStr, &createdByStr, &updatedByStr); err != nil {
		return nil, err
	}

	createdAt, err := time.Parse(time.RFC3339Nano, createdAtStr)
	if err != nil {
		return nil, err
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, updatedAtStr)
	if err != nil {
		return nil, err
	}
	var deletedAt *time.Time
	if deletedAtStr.Valid && deletedAtStr.String != "" {
		t, err := time.Parse(time.RFC3339Nano, deletedAtStr.String)
		if err != nil {
			return nil, err
		}
		deletedAt = &t
	}
	var createdBy *string
	var updatedBy *string
	if createdByStr.Valid && createdByStr.String != "" {
		s := createdByStr.String
		createdBy = &s
	}
	if updatedByStr.Valid && updatedByStr.String != "" {
		s := updatedByStr.String
		updatedBy = &s
	}

	return &domain.User{
		ID:                  id,
		Username:            username,
		PasswordHash:        passwordHash,
		ForcePasswordChange: intToBool(force),
		Theme:               theme,
		Audit: domain.AuditFields{
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			DeletedAt: deletedAt,
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		},
	}, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intToBool(i int) bool {
	return i != 0
}
