package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/cr1cr1/farm-manager/internal/domain"
)

// BarnRepo defines operations for barn management.
type BarnRepo interface {
	// Count returns the number of non-deleted barns.
	Count(ctx context.Context) (int64, error)
	// List returns all non-deleted barns.
	List(ctx context.Context) ([]*domain.Barn, error)
	// FindByID returns a barn by ID (excluding soft-deleted).
	FindByID(ctx context.Context, id int64) (*domain.Barn, error)
	// Create inserts a new barn.
	Create(ctx context.Context, barn *domain.Barn) (int64, error)
	// Update modifies an existing barn.
	Update(ctx context.Context, barn *domain.Barn) error
	// SoftDelete marks the barn as deleted.
	SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error
}

type SQLiteBarnRepo struct {
	DB *sql.DB
}

func NewSQLiteBarnRepo(db *sql.DB) *SQLiteBarnRepo {
	return &SQLiteBarnRepo{DB: db}
}

func (r *SQLiteBarnRepo) Count(ctx context.Context) (int64, error) {
	const q = `SELECT COUNT(1) FROM barns WHERE deleted_at IS NULL`
	var n int64
	if err := r.DB.QueryRowContext(ctx, q).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *SQLiteBarnRepo) List(ctx context.Context) ([]*domain.Barn, error) {
	const q = `
		SELECT barn_id, name, capacity, environment_control, maintenance_schedule, location,
			   created_at, updated_at, deleted_at, created_by, updated_by
		FROM barns
		WHERE deleted_at IS NULL
		ORDER BY name
	`
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var barns []*domain.Barn
	for rows.Next() {
		var barn domain.Barn
		err := rows.Scan(
			&barn.BarnID,
			&barn.Name,
			&barn.Capacity,
			&barn.EnvironmentControl,
			&barn.MaintenanceSchedule,
			&barn.Location,
			&barn.Audit.CreatedAt,
			&barn.Audit.UpdatedAt,
			&barn.Audit.DeletedAt,
			&barn.Audit.CreatedBy,
			&barn.Audit.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		barns = append(barns, &barn)
	}
	return barns, rows.Err()
}

func (r *SQLiteBarnRepo) FindByID(ctx context.Context, id int64) (*domain.Barn, error) {
	const q = `
		SELECT barn_id, name, capacity, environment_control, maintenance_schedule, location,
			   created_at, updated_at, deleted_at, created_by, updated_by
		FROM barns
		WHERE barn_id = ? AND deleted_at IS NULL
	`
	var barn domain.Barn
	err := r.DB.QueryRowContext(ctx, q, id).Scan(
		&barn.BarnID,
		&barn.Name,
		&barn.Capacity,
		&barn.EnvironmentControl,
		&barn.MaintenanceSchedule,
		&barn.Location,
		&barn.Audit.CreatedAt,
		&barn.Audit.UpdatedAt,
		&barn.Audit.DeletedAt,
		&barn.Audit.CreatedBy,
		&barn.Audit.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	return &barn, nil
}

func (r *SQLiteBarnRepo) Create(ctx context.Context, barn *domain.Barn) (int64, error) {
	const q = `
		INSERT INTO barns (name, capacity, environment_control, maintenance_schedule, location,
						   created_at, updated_at, created_by, updated_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	barn.Audit.CreatedAt = now
	barn.Audit.UpdatedAt = now

	result, err := r.DB.ExecContext(ctx, q,
		barn.Name,
		barn.Capacity,
		barn.EnvironmentControl,
		barn.MaintenanceSchedule,
		barn.Location,
		barn.Audit.CreatedAt,
		barn.Audit.UpdatedAt,
		barn.Audit.CreatedBy,
		barn.Audit.UpdatedBy,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *SQLiteBarnRepo) Update(ctx context.Context, barn *domain.Barn) error {
	const q = `
		UPDATE barns
		SET name = ?, capacity = ?, environment_control = ?, maintenance_schedule = ?, location = ?,
			updated_at = ?, updated_by = ?
		WHERE barn_id = ? AND deleted_at IS NULL
	`
	now := time.Now()
	barn.Audit.UpdatedAt = now

	_, err := r.DB.ExecContext(ctx, q,
		barn.Name,
		barn.Capacity,
		barn.EnvironmentControl,
		barn.MaintenanceSchedule,
		barn.Location,
		barn.Audit.UpdatedAt,
		barn.Audit.UpdatedBy,
		barn.BarnID,
	)
	return err
}

func (r *SQLiteBarnRepo) SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error {
	const q = `UPDATE barns SET deleted_at = ? WHERE barn_id = ?`
	_, err := r.DB.ExecContext(ctx, q, deletedAt, id)
	return err
}
