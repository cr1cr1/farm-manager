package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/cr1cr1/farm-manager/internal/domain"
)

// StaffRepo defines operations for staff management.
type StaffRepo interface {
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context) ([]*domain.Staff, error)
	FindByID(ctx context.Context, id int64) (*domain.Staff, error)
	Create(ctx context.Context, staff *domain.Staff) (int64, error)
	Update(ctx context.Context, staff *domain.Staff) error
	SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error
}

type SQLiteStaffRepo struct {
	DB *sql.DB
}

func NewSQLiteStaffRepo(db *sql.DB) *SQLiteStaffRepo {
	return &SQLiteStaffRepo{DB: db}
}

func (r *SQLiteStaffRepo) Count(ctx context.Context) (int64, error) {
	const q = `SELECT COUNT(1) FROM staff WHERE deleted_at IS NULL`
	var n int64
	if err := r.DB.QueryRowContext(ctx, q).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *SQLiteStaffRepo) List(ctx context.Context) ([]*domain.Staff, error) {
	const q = `
		SELECT staff_id, name, role, schedule, contact_info,
			   created_at, updated_at, deleted_at, created_by, updated_by
		FROM staff
		WHERE deleted_at IS NULL
		ORDER BY name
	`
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var staff []*domain.Staff
	for rows.Next() {
		var s domain.Staff
		err := rows.Scan(
			&s.StaffID,
			&s.Name,
			&s.Role,
			&s.Schedule,
			&s.ContactInfo,
			&s.Audit.CreatedAt,
			&s.Audit.UpdatedAt,
			&s.Audit.DeletedAt,
			&s.Audit.CreatedBy,
			&s.Audit.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		staff = append(staff, &s)
	}
	return staff, rows.Err()
}

func (r *SQLiteStaffRepo) FindByID(ctx context.Context, id int64) (*domain.Staff, error) {
	const q = `
		SELECT staff_id, name, role, schedule, contact_info,
			   created_at, updated_at, deleted_at, created_by, updated_by
		FROM staff
		WHERE staff_id = ? AND deleted_at IS NULL
	`
	var staff domain.Staff
	err := r.DB.QueryRowContext(ctx, q, id).Scan(
		&staff.StaffID,
		&staff.Name,
		&staff.Role,
		&staff.Schedule,
		&staff.ContactInfo,
		&staff.Audit.CreatedAt,
		&staff.Audit.UpdatedAt,
		&staff.Audit.DeletedAt,
		&staff.Audit.CreatedBy,
		&staff.Audit.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	return &staff, nil
}

func (r *SQLiteStaffRepo) Create(ctx context.Context, staff *domain.Staff) (int64, error) {
	const q = `
		INSERT INTO staff (name, role, schedule, contact_info,
						   created_at, updated_at, created_by, updated_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	staff.Audit.CreatedAt = now
	staff.Audit.UpdatedAt = now

	result, err := r.DB.ExecContext(ctx, q,
		staff.Name,
		staff.Role,
		staff.Schedule,
		staff.ContactInfo,
		staff.Audit.CreatedAt,
		staff.Audit.UpdatedAt,
		staff.Audit.CreatedBy,
		staff.Audit.UpdatedBy,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *SQLiteStaffRepo) Update(ctx context.Context, staff *domain.Staff) error {
	const q = `
		UPDATE staff
		SET name = ?, role = ?, schedule = ?, contact_info = ?,
			updated_at = ?, updated_by = ?
		WHERE staff_id = ? AND deleted_at IS NULL
	`
	now := time.Now()
	staff.Audit.UpdatedAt = now

	_, err := r.DB.ExecContext(ctx, q,
		staff.Name,
		staff.Role,
		staff.Schedule,
		staff.ContactInfo,
		staff.Audit.UpdatedAt,
		staff.Audit.UpdatedBy,
		staff.StaffID,
	)
	return err
}

func (r *SQLiteStaffRepo) SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error {
	const q = `UPDATE staff SET deleted_at = ? WHERE staff_id = ?`
	_, err := r.DB.ExecContext(ctx, q, deletedAt, id)
	return err
}
