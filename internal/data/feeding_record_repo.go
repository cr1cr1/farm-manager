package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/cr1cr1/farm-manager/internal/domain"
)

// FeedingRecordRepo defines operations for feedingrecord management.
type FeedingRecordRepo interface {
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context) ([]*domain.FeedingRecord, error)
	FindByID(ctx context.Context, id int64) (*domain.FeedingRecord, error)
	Create(ctx context.Context, f *domain.FeedingRecord) (int64, error)
	Update(ctx context.Context, f *domain.FeedingRecord) error
	SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error
}

type SQLiteFeedingRecordRepo struct {
	DB *sql.DB
}

func NewSQLiteFeedingRecordRepo(db *sql.DB) *SQLiteFeedingRecordRepo {
	return &SQLiteFeedingRecordRepo{DB: db}
}

func (r *SQLiteFeedingRecordRepo) Count(ctx context.Context) (int64, error) {
	const q = `SELECT COUNT(1) FROM feeding_records WHERE deleted_at IS NULL`
	var n int64
	if err := r.DB.QueryRowContext(ctx, q).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *SQLiteFeedingRecordRepo) List(ctx context.Context) ([]*domain.FeedingRecord, error) {
	const q = `SELECT feeding_record_id, flock_id, feed_type_id, amount_given, date_time, staff_id, created_at, updated_at, deleted_at, created_by, updated_by FROM feeding_records WHERE deleted_at IS NULL ORDER BY feeding_record_id`
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*domain.FeedingRecord
	for rows.Next() {
		var item domain.FeedingRecord
		err := rows.Scan(
			&item.FeedingRecordID,
			&item.FlockID,
			&item.FeedTypeID,
			&item.AmountGiven,
			&item.DateTime,
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

func (r *SQLiteFeedingRecordRepo) FindByID(ctx context.Context, id int64) (*domain.FeedingRecord, error) {
	const q = `SELECT feeding_record_id, flock_id, feed_type_id, amount_given, date_time, staff_id, created_at, updated_at, deleted_at, created_by, updated_by FROM feeding_records WHERE feeding_record_id = ? AND deleted_at IS NULL`
	var item domain.FeedingRecord
	err := r.DB.QueryRowContext(ctx, q, id).Scan(
		&item.FeedingRecordID,
		&item.FlockID,
		&item.FeedTypeID,
		&item.AmountGiven,
		&item.DateTime,
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

func (r *SQLiteFeedingRecordRepo) Create(ctx context.Context, f *domain.FeedingRecord) (int64, error) {
	const q = `INSERT INTO feeding_records (flock_id, feed_type_id, amount_given, date_time, staff_id, created_at, updated_at, created_by, updated_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	f.Audit.CreatedAt = now
	f.Audit.UpdatedAt = now

	result, err := r.DB.ExecContext(ctx, q,
		f.FlockID,
		f.FeedTypeID,
		f.AmountGiven,
		f.DateTime,
		f.StaffID,
		f.Audit.CreatedAt,
		f.Audit.UpdatedAt,
		f.Audit.CreatedBy,
		f.Audit.UpdatedBy,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *SQLiteFeedingRecordRepo) Update(ctx context.Context, f *domain.FeedingRecord) error {
	const q = `UPDATE feeding_records SET flock_id = ?, feed_type_id = ?, amount_given = ?, date_time = ?, staff_id = ?, updated_at = ?, updated_by = ? WHERE feeding_record_id = ? AND deleted_at IS NULL`
	f.Audit.UpdatedAt = time.Now()

	_, err := r.DB.ExecContext(ctx, q,
		f.FlockID,
		f.FeedTypeID,
		f.AmountGiven,
		f.DateTime,
		f.StaffID,
		f.Audit.UpdatedAt,
		f.Audit.UpdatedBy,
		f.FeedingRecordID,
	)
	return err
}

func (r *SQLiteFeedingRecordRepo) SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error {
	const q = `UPDATE feeding_records SET deleted_at = ? WHERE feeding_record_id = ?`
	_, err := r.DB.ExecContext(ctx, q, deletedAt, id)
	return err
}
