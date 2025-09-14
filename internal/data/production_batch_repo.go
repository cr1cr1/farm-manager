package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/cr1cr1/farm-manager/internal/domain"
)

// ProductionBatchRepo defines operations for productionbatch management.
type ProductionBatchRepo interface {
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context) ([]*domain.ProductionBatch, error)
	FindByID(ctx context.Context, id int64) (*domain.ProductionBatch, error)
	Create(ctx context.Context, p *domain.ProductionBatch) (int64, error)
	Update(ctx context.Context, p *domain.ProductionBatch) error
	SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error
}

type SQLiteProductionBatchRepo struct {
	DB *sql.DB
}

func NewSQLiteProductionBatchRepo(db *sql.DB) *SQLiteProductionBatchRepo {
	return &SQLiteProductionBatchRepo{DB: db}
}

func (r *SQLiteProductionBatchRepo) Count(ctx context.Context) (int64, error) {
	const q = `SELECT COUNT(1) FROM production_batches WHERE deleted_at IS NULL`
	var n int64
	if err := r.DB.QueryRowContext(ctx, q).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *SQLiteProductionBatchRepo) List(ctx context.Context) ([]*domain.ProductionBatch, error) {
	const q = `SELECT batch_id, flock_id, date_ready, number_in_batch, weight_estimate, notes, created_at, updated_at, deleted_at, created_by, updated_by FROM production_batches WHERE deleted_at IS NULL ORDER BY batch_id`
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*domain.ProductionBatch
	for rows.Next() {
		var item domain.ProductionBatch
		err := rows.Scan(
			&item.BatchID,
			&item.FlockID,
			&item.DateReady,
			&item.NumberInBatch,
			&item.WeightEstimate,
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

func (r *SQLiteProductionBatchRepo) FindByID(ctx context.Context, id int64) (*domain.ProductionBatch, error) {
	const q = `SELECT batch_id, flock_id, date_ready, number_in_batch, weight_estimate, notes, created_at, updated_at, deleted_at, created_by, updated_by FROM production_batches WHERE batch_id = ? AND deleted_at IS NULL`
	var item domain.ProductionBatch
	err := r.DB.QueryRowContext(ctx, q, id).Scan(
		&item.BatchID,
		&item.FlockID,
		&item.DateReady,
		&item.NumberInBatch,
		&item.WeightEstimate,
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

func (r *SQLiteProductionBatchRepo) Create(ctx context.Context, p *domain.ProductionBatch) (int64, error) {
	const q = `INSERT INTO production_batches (flock_id, date_ready, number_in_batch, weight_estimate, notes, created_at, updated_at, created_by, updated_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	p.Audit.CreatedAt = now
	p.Audit.UpdatedAt = now

	result, err := r.DB.ExecContext(ctx, q,
		p.FlockID,
		p.DateReady,
		p.NumberInBatch,
		p.WeightEstimate,
		p.Notes,
		p.Audit.CreatedAt,
		p.Audit.UpdatedAt,
		p.Audit.CreatedBy,
		p.Audit.UpdatedBy,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *SQLiteProductionBatchRepo) Update(ctx context.Context, p *domain.ProductionBatch) error {
	const q = `UPDATE production_batches SET flock_id = ?, date_ready = ?, number_in_batch = ?, weight_estimate = ?, notes = ?, updated_at = ?, updated_by = ? WHERE batch_id = ? AND deleted_at IS NULL`
	p.Audit.UpdatedAt = time.Now()

	_, err := r.DB.ExecContext(ctx, q,
		p.FlockID,
		p.DateReady,
		p.NumberInBatch,
		p.WeightEstimate,
		p.Notes,
		p.Audit.UpdatedAt,
		p.Audit.UpdatedBy,
		p.BatchID,
	)
	return err
}

func (r *SQLiteProductionBatchRepo) SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error {
	const q = `UPDATE production_batches SET deleted_at = ? WHERE batch_id = ?`
	_, err := r.DB.ExecContext(ctx, q, deletedAt, id)
	return err
}
