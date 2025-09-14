package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/cr1cr1/farm-manager/internal/domain"
)

// HealthCheckRepo defines operations for healthcheck management.
type HealthCheckRepo interface {
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context) ([]*domain.HealthCheck, error)
	FindByID(ctx context.Context, id int64) (*domain.HealthCheck, error)
	Create(ctx context.Context, h *domain.HealthCheck) (int64, error)
	Update(ctx context.Context, h *domain.HealthCheck) error
	SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error
}

type SQLiteHealthCheckRepo struct {
	DB *sql.DB
}

func NewSQLiteHealthCheckRepo(db *sql.DB) *SQLiteHealthCheckRepo {
	return &SQLiteHealthCheckRepo{DB: db}
}

func (r *SQLiteHealthCheckRepo) Count(ctx context.Context) (int64, error) {
	const q = `SELECT COUNT(1) FROM health_checks WHERE deleted_at IS NULL`
	var n int64
	if err := r.DB.QueryRowContext(ctx, q).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *SQLiteHealthCheckRepo) List(ctx context.Context) ([]*domain.HealthCheck, error) {
	const q = `SELECT health_check_id, flock_id, check_date, health_status, vaccinations_given, treatments_administered, notes, staff_id, created_at, updated_at, deleted_at, created_by, updated_by FROM health_checks WHERE deleted_at IS NULL ORDER BY health_check_id`
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*domain.HealthCheck
	for rows.Next() {
		var item domain.HealthCheck
		err := rows.Scan(
			&item.HealthCheckID,
			&item.FlockID,
			&item.CheckDate,
			&item.HealthStatus,
			&item.VaccinationsGiven,
			&item.TreatmentsAdministered,
			&item.Notes,
			&item.StaffID,
			&item.Audit.CreatedAt,
			&item.Audit.UpdatedAt,
			&item.Audit.DeletedAt,
			&item.Audit.CreatedBy,
			&item.Audit.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, rows.Err()
}

func (r *SQLiteHealthCheckRepo) FindByID(ctx context.Context, id int64) (*domain.HealthCheck, error) {
	const q = `SELECT health_check_id, flock_id, check_date, health_status, vaccinations_given, treatments_administered, notes, staff_id, created_at, updated_at, deleted_at, created_by, updated_by FROM health_checks WHERE health_check_id = ? AND deleted_at IS NULL`
	var item domain.HealthCheck
	err := r.DB.QueryRowContext(ctx, q, id).Scan(
		&item.HealthCheckID,
		&item.FlockID,
		&item.CheckDate,
		&item.HealthStatus,
		&item.VaccinationsGiven,
		&item.TreatmentsAdministered,
		&item.Notes,
		&item.StaffID,
		&item.Audit.CreatedAt,
		&item.Audit.UpdatedAt,
		&item.Audit.DeletedAt,
		&item.Audit.CreatedBy,
		&item.Audit.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *SQLiteHealthCheckRepo) Create(ctx context.Context, h *domain.HealthCheck) (int64, error) {
	const q = `INSERT INTO health_checks (flock_id, check_date, health_status, vaccinations_given, treatments_administered, notes, staff_id, created_at, updated_at, created_by, updated_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	h.Audit.CreatedAt = now
	h.Audit.UpdatedAt = now

	result, err := r.DB.ExecContext(ctx, q,
		h.FlockID,
		h.CheckDate,
		h.HealthStatus,
		h.VaccinationsGiven,
		h.TreatmentsAdministered,
		h.Notes,
		h.StaffID,
		h.Audit.CreatedAt,
		h.Audit.UpdatedAt,
		h.Audit.CreatedBy,
		h.Audit.UpdatedBy,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *SQLiteHealthCheckRepo) Update(ctx context.Context, h *domain.HealthCheck) error {
	const q = `UPDATE health_checks SET flock_id = ?, check_date = ?, health_status = ?, vaccinations_given = ?, treatments_administered = ?, notes = ?, staff_id = ?, updated_at = ?, updated_by = ? WHERE health_check_id = ? AND deleted_at IS NULL`
	h.Audit.UpdatedAt = time.Now()

	_, err := r.DB.ExecContext(ctx, q,
		h.FlockID,
		h.CheckDate,
		h.HealthStatus,
		h.VaccinationsGiven,
		h.TreatmentsAdministered,
		h.Notes,
		h.StaffID,
		h.Audit.UpdatedAt,
		h.Audit.UpdatedBy,
		h.HealthCheckID,
	)
	return err
}

func (r *SQLiteHealthCheckRepo) SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error {
	const q = `UPDATE health_checks SET deleted_at = ? WHERE health_check_id = ?`
	_, err := r.DB.ExecContext(ctx, q, deletedAt, id)
	return err
}
