package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/cr1cr1/farm-manager/internal/domain"
)

// FeedTypeRepo defines operations for feed type management.
type FeedTypeRepo interface {
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context) ([]*domain.FeedType, error)
	FindByID(ctx context.Context, id int64) (*domain.FeedType, error)
	Create(ctx context.Context, feedType *domain.FeedType) (int64, error)
	Update(ctx context.Context, feedType *domain.FeedType) error
	SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error
}

type SQLiteFeedTypeRepo struct {
	DB *sql.DB
}

func NewSQLiteFeedTypeRepo(db *sql.DB) *SQLiteFeedTypeRepo {
	return &SQLiteFeedTypeRepo{DB: db}
}

func (r *SQLiteFeedTypeRepo) Count(ctx context.Context) (int64, error) {
	const q = `SELECT COUNT(1) FROM feed_types WHERE deleted_at IS NULL`
	var n int64
	if err := r.DB.QueryRowContext(ctx, q).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *SQLiteFeedTypeRepo) List(ctx context.Context) ([]*domain.FeedType, error) {
	const q = `
		SELECT feed_type_id, name, description, nutritional_info,
			   created_at, updated_at, deleted_at, created_by, updated_by
		FROM feed_types
		WHERE deleted_at IS NULL
		ORDER BY name
	`
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedTypes []*domain.FeedType
	for rows.Next() {
		var feedType domain.FeedType
		err := rows.Scan(
			&feedType.FeedTypeID,
			&feedType.Name,
			&feedType.Description,
			&feedType.NutritionalInfo,
			&feedType.Audit.CreatedAt,
			&feedType.Audit.UpdatedAt,
			&feedType.Audit.DeletedAt,
			&feedType.Audit.CreatedBy,
			&feedType.Audit.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		feedTypes = append(feedTypes, &feedType)
	}
	return feedTypes, rows.Err()
}

func (r *SQLiteFeedTypeRepo) FindByID(ctx context.Context, id int64) (*domain.FeedType, error) {
	const q = `
		SELECT feed_type_id, name, description, nutritional_info,
			   created_at, updated_at, deleted_at, created_by, updated_by
		FROM feed_types
		WHERE feed_type_id = ? AND deleted_at IS NULL
	`
	var feedType domain.FeedType
	err := r.DB.QueryRowContext(ctx, q, id).Scan(
		&feedType.FeedTypeID,
		&feedType.Name,
		&feedType.Description,
		&feedType.NutritionalInfo,
		&feedType.Audit.CreatedAt,
		&feedType.Audit.UpdatedAt,
		&feedType.Audit.DeletedAt,
		&feedType.Audit.CreatedBy,
		&feedType.Audit.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	return &feedType, nil
}

func (r *SQLiteFeedTypeRepo) Create(ctx context.Context, feedType *domain.FeedType) (int64, error) {
	const q = `
		INSERT INTO feed_types (name, description, nutritional_info,
							   created_at, updated_at, created_by, updated_by)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	feedType.Audit.CreatedAt = now
	feedType.Audit.UpdatedAt = now

	result, err := r.DB.ExecContext(ctx, q,
		feedType.Name,
		feedType.Description,
		feedType.NutritionalInfo,
		feedType.Audit.CreatedAt,
		feedType.Audit.UpdatedAt,
		feedType.Audit.CreatedBy,
		feedType.Audit.UpdatedBy,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *SQLiteFeedTypeRepo) Update(ctx context.Context, feedType *domain.FeedType) error {
	const q = `
		UPDATE feed_types
		SET name = ?, description = ?, nutritional_info = ?,
			updated_at = ?, updated_by = ?
		WHERE feed_type_id = ? AND deleted_at IS NULL
	`
	now := time.Now()
	feedType.Audit.UpdatedAt = now

	_, err := r.DB.ExecContext(ctx, q,
		feedType.Name,
		feedType.Description,
		feedType.NutritionalInfo,
		feedType.Audit.UpdatedAt,
		feedType.Audit.UpdatedBy,
		feedType.FeedTypeID,
	)
	return err
}

func (r *SQLiteFeedTypeRepo) SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error {
	const q = `UPDATE feed_types SET deleted_at = ? WHERE feed_type_id = ?`
	_, err := r.DB.ExecContext(ctx, q, deletedAt, id)
	return err
}
