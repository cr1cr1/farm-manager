package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/cr1cr1/farm-manager/internal/domain"
)

// SlaughterRecordRepo defines operations for slaughterrecord management.
type SlaughterRecordRepo interface {
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context) ([]*domain.SlaughterRecord, error)
	FindByID(ctx context.Context, id int64) (*domain.SlaughterRecord, error)
	Create(ctx context.Context, s *domain.SlaughterRecord) (int64, error)
	Update(ctx context.Context, s *domain.SlaughterRecord) error
	SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error
}

type SQLiteSlaughterRecordRepo struct {
	DB *sql.DB
}

func NewSQLiteSlaughterRecordRepo(db *sql.DB) *SQLiteSlaughterRecordRepo {
	return &SQLiteSlaughterRecordRepo{DB: db}
}

func (r *SQLiteSlaughterRecordRepo) Count(ctx context.Context) (int64, error) {
	const q = `SELECT COUNT(1) FROM slaughter_records WHERE deleted_at IS NULL`
	var n int64
	if err := r.DB.QueryRowContext(ctx, q).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *SQLiteSlaughterRecordRepo) List(ctx context.Context) ([]*domain.SlaughterRecord, error) {
	const q = `SELECT slaughter_id, batch_id, date, number_slaughtered, meat_yield, waste, staff_id, created_at, updated_at, deleted_at, created_by, updated_by FROM slaughter_records WHERE deleted_at IS NULL ORDER BY slaughter_id`
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*domain.SlaughterRecord
	for rows.Next() {
		var item domain.SlaughterRecord
		err := rows.Scan(
			&item.SlaughterID,
			&item.BatchID,
			&item.Date,
			&item.NumberSlaughtered,
			&item.MeatYield,
			&item.Waste,
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

func (r *SQLiteSlaughterRecordRepo) FindByID(ctx context.Context, id int64) (*domain.SlaughterRecord, error) {
	const q = `SELECT slaughter_id, batch_id, date, number_slaughtered, meat_yield, waste, staff_id, created_at, updated_at, deleted_at, created_by, updated_by FROM slaughter_records WHERE slaughter_id = ? AND deleted_at IS NULL`
	var item domain.SlaughterRecord
	err := r.DB.QueryRowContext(ctx, q, id).Scan(
		&item.SlaughterID,
		&item.BatchID,
		&item.Date,
		&item.NumberSlaughtered,
		&item.MeatYield,
		&item.Waste,
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

func (r *SQLiteSlaughterRecordRepo) Create(ctx context.Context, s *domain.SlaughterRecord) (int64, error) {
	const q = `INSERT INTO slaughter_records (batch_id, date, number_slaughtered, meat_yield, waste, staff_id, created_at, updated_at, created_by, updated_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	s.Audit.CreatedAt = now
	s.Audit.UpdatedAt = now

	result, err := r.DB.ExecContext(ctx, q,
		s.BatchID,
		s.Date,
		s.NumberSlaughtered,
		s.MeatYield,
		s.Waste,
		s.StaffID,
		s.Audit.CreatedAt,
		s.Audit.UpdatedAt,
		s.Audit.CreatedBy,
		s.Audit.UpdatedBy,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *SQLiteSlaughterRecordRepo) Update(ctx context.Context, s *domain.SlaughterRecord) error {
	const q = `UPDATE slaughter_records SET batch_id = ?, date = ?, number_slaughtered = ?, meat_yield = ?, waste = ?, staff_id = ?, updated_at = ?, updated_by = ? WHERE slaughter_id = ? AND deleted_at IS NULL`
	s.Audit.UpdatedAt = time.Now()

	_, err := r.DB.ExecContext(ctx, q,
		s.BatchID,
		s.Date,
		s.NumberSlaughtered,
		s.MeatYield,
		s.Waste,
		s.StaffID,
		s.Audit.UpdatedAt,
		s.Audit.UpdatedBy,
		s.SlaughterID,
	)
	return err
}

func (r *SQLiteSlaughterRecordRepo) SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error {
	const q = `UPDATE slaughter_records SET deleted_at = ? WHERE slaughter_id = ?`
	_, err := r.DB.ExecContext(ctx, q, deletedAt, id)
	return err
}
