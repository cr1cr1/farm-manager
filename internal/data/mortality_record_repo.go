package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/cr1cr1/farm-manager/internal/domain"
)

// MortalityRecordRepo defines operations for mortalityrecord management.
type MortalityRecordRepo interface {
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context) ([]*domain.MortalityRecord, error)
	FindByID(ctx context.Context, id int64) (*domain.MortalityRecord, error)
	Create(ctx context.Context, m *domain.MortalityRecord) (int64, error)
	Update(ctx context.Context, m *domain.MortalityRecord) error
	SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error
}

type SQLiteMortalityRecordRepo struct {
	DB *sql.DB
}

func NewSQLiteMortalityRecordRepo(db *sql.DB) *SQLiteMortalityRecordRepo {
	return &SQLiteMortalityRecordRepo{DB: db}
}

func (r *SQLiteMortalityRecordRepo) Count(ctx context.Context) (int64, error) {
	const q = `SELECT COUNT(1) FROM mortality_records WHERE deleted_at IS NULL`
	var n int64
	if err := r.DB.QueryRowContext(ctx, q).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *SQLiteMortalityRecordRepo) List(ctx context.Context) ([]*domain.MortalityRecord, error) {
	const q = `SELECT mortality_record_id, flock_id, date, number_dead, cause_of_death, notes, created_at, updated_at, deleted_at, created_by, updated_by FROM mortality_records WHERE deleted_at IS NULL ORDER BY mortality_record_id`
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*domain.MortalityRecord
	for rows.Next() {
		var item domain.MortalityRecord
		err := rows.Scan(
			&item.MortalityRecordID,
			&item.FlockID,
			&item.Date,
			&item.NumberDead,
			&item.CauseOfDeath,
			&item.Notes,
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

func (r *SQLiteMortalityRecordRepo) FindByID(ctx context.Context, id int64) (*domain.MortalityRecord, error) {
	const q = `SELECT mortality_record_id, flock_id, date, number_dead, cause_of_death, notes, created_at, updated_at, deleted_at, created_by, updated_by FROM mortality_records WHERE mortality_record_id = ? AND deleted_at IS NULL`
	var item domain.MortalityRecord
	err := r.DB.QueryRowContext(ctx, q, id).Scan(
		&item.MortalityRecordID,
		&item.FlockID,
		&item.Date,
		&item.NumberDead,
		&item.CauseOfDeath,
		&item.Notes,
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

func (r *SQLiteMortalityRecordRepo) Create(ctx context.Context, m *domain.MortalityRecord) (int64, error) {
	const q = `INSERT INTO mortality_records (flock_id, date, number_dead, cause_of_death, notes, created_at, updated_at, created_by, updated_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	m.Audit.CreatedAt = now
	m.Audit.UpdatedAt = now

	result, err := r.DB.ExecContext(ctx, q,
		m.FlockID,
		m.Date,
		m.NumberDead,
		m.CauseOfDeath,
		m.Notes,
		m.Audit.CreatedAt,
		m.Audit.UpdatedAt,
		m.Audit.CreatedBy,
		m.Audit.UpdatedBy,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *SQLiteMortalityRecordRepo) Update(ctx context.Context, m *domain.MortalityRecord) error {
	const q = `UPDATE mortality_records SET flock_id = ?, date = ?, number_dead = ?, cause_of_death = ?, notes = ?, updated_at = ?, updated_by = ? WHERE mortality_record_id = ? AND deleted_at IS NULL`
	m.Audit.UpdatedAt = time.Now()

	_, err := r.DB.ExecContext(ctx, q,
		m.FlockID,
		m.Date,
		m.NumberDead,
		m.CauseOfDeath,
		m.Notes,
		m.Audit.UpdatedAt,
		m.Audit.UpdatedBy,
		m.MortalityRecordID,
	)
	return err
}

func (r *SQLiteMortalityRecordRepo) SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error {
	const q = `UPDATE mortality_records SET deleted_at = ? WHERE mortality_record_id = ?`
	_, err := r.DB.ExecContext(ctx, q, deletedAt, id)
	return err
}
