package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/cr1cr1/farm-manager/internal/domain"
)

// FlockRepo defines operations for flock management.
type FlockRepo interface {
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context) ([]*domain.Flock, error)
	FindByID(ctx context.Context, id int64) (*domain.Flock, error)
	Create(ctx context.Context, flock *domain.Flock) (int64, error)
	Update(ctx context.Context, flock *domain.Flock) error
	SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error
}

type SQLiteFlockRepo struct {
	DB *sql.DB
}

func NewSQLiteFlockRepo(db *sql.DB) *SQLiteFlockRepo {
	return &SQLiteFlockRepo{DB: db}
}

func (r *SQLiteFlockRepo) Count(ctx context.Context) (int64, error) {
	const q = `SELECT COUNT(1) FROM flocks WHERE deleted_at IS NULL`
	var n int64
	if err := r.DB.QueryRowContext(ctx, q).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *SQLiteFlockRepo) List(ctx context.Context) ([]*domain.Flock, error) {
	const q = `
		SELECT f.flock_id, f.breed, f.hatch_date, f.number_of_birds, f.current_age,
			   f.barn_id, f.health_status, f.feed_type_id, f.notes,
			   f.created_at, f.updated_at, f.deleted_at, f.created_by, f.updated_by,
			   b.name as barn_name, ft.name as feed_type_name
		FROM flocks f
		LEFT JOIN barns b ON f.barn_id = b.barn_id AND b.deleted_at IS NULL
		LEFT JOIN feed_types ft ON f.feed_type_id = ft.feed_type_id AND ft.deleted_at IS NULL
		WHERE f.deleted_at IS NULL
		ORDER BY f.breed, f.flock_id
	`
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flocks []*domain.Flock
	for rows.Next() {
		var flock domain.Flock
		var barnName, feedTypeName sql.NullString

		err := rows.Scan(
			&flock.FlockID,
			&flock.Breed,
			&flock.HatchDate,
			&flock.NumberOfBirds,
			&flock.CurrentAge,
			&flock.BarnID,
			&flock.HealthStatus,
			&flock.FeedTypeID,
			&flock.Notes,
			&flock.Audit.CreatedAt,
			&flock.Audit.UpdatedAt,
			&flock.Audit.DeletedAt,
			&flock.Audit.CreatedBy,
			&flock.Audit.UpdatedBy,
			&barnName,
			&feedTypeName,
		)
		if err != nil {
			return nil, err
		}

		// Populate relations if they exist
		if flock.BarnID != nil {
			flock.Barn = &domain.Barn{Name: barnName.String}
		}
		if flock.FeedTypeID != nil {
			flock.FeedType = &domain.FeedType{Name: feedTypeName.String}
		}

		flocks = append(flocks, &flock)
	}
	return flocks, rows.Err()
}

func (r *SQLiteFlockRepo) FindByID(ctx context.Context, id int64) (*domain.Flock, error) {
	const q = `
		SELECT f.flock_id, f.breed, f.hatch_date, f.number_of_birds, f.current_age,
			   f.barn_id, f.health_status, f.feed_type_id, f.notes,
			   f.created_at, f.updated_at, f.deleted_at, f.created_by, f.updated_by,
			   b.name as barn_name, ft.name as feed_type_name
		FROM flocks f
		LEFT JOIN barns b ON f.barn_id = b.barn_id AND b.deleted_at IS NULL
		LEFT JOIN feed_types ft ON f.feed_type_id = ft.feed_type_id AND ft.deleted_at IS NULL
		WHERE f.flock_id = ? AND f.deleted_at IS NULL
	`
	var flock domain.Flock
	var barnName, feedTypeName sql.NullString

	err := r.DB.QueryRowContext(ctx, q, id).Scan(
		&flock.FlockID,
		&flock.Breed,
		&flock.HatchDate,
		&flock.NumberOfBirds,
		&flock.CurrentAge,
		&flock.BarnID,
		&flock.HealthStatus,
		&flock.FeedTypeID,
		&flock.Notes,
		&flock.Audit.CreatedAt,
		&flock.Audit.UpdatedAt,
		&flock.Audit.DeletedAt,
		&flock.Audit.CreatedBy,
		&flock.Audit.UpdatedBy,
		&barnName,
		&feedTypeName,
	)
	if err != nil {
		return nil, err
	}

	// Populate relations if they exist
	if flock.BarnID != nil {
		flock.Barn = &domain.Barn{Name: barnName.String}
	}
	if flock.FeedTypeID != nil {
		flock.FeedType = &domain.FeedType{Name: feedTypeName.String}
	}

	return &flock, nil
}

func (r *SQLiteFlockRepo) Create(ctx context.Context, flock *domain.Flock) (int64, error) {
	const q = `
		INSERT INTO flocks (breed, hatch_date, number_of_birds, current_age,
						   barn_id, health_status, feed_type_id, notes,
						   created_at, updated_at, created_by, updated_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	flock.Audit.CreatedAt = now
	flock.Audit.UpdatedAt = now

	result, err := r.DB.ExecContext(ctx, q,
		flock.Breed,
		flock.HatchDate,
		flock.NumberOfBirds,
		flock.CurrentAge,
		flock.BarnID,
		flock.HealthStatus,
		flock.FeedTypeID,
		flock.Notes,
		flock.Audit.CreatedAt,
		flock.Audit.UpdatedAt,
		flock.Audit.CreatedBy,
		flock.Audit.UpdatedBy,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *SQLiteFlockRepo) Update(ctx context.Context, flock *domain.Flock) error {
	const q = `
		UPDATE flocks
		SET breed = ?, hatch_date = ?, number_of_birds = ?, current_age = ?,
			barn_id = ?, health_status = ?, feed_type_id = ?, notes = ?,
			updated_at = ?, updated_by = ?
		WHERE flock_id = ? AND deleted_at IS NULL
	`
	now := time.Now()
	flock.Audit.UpdatedAt = now

	_, err := r.DB.ExecContext(ctx, q,
		flock.Breed,
		flock.HatchDate,
		flock.NumberOfBirds,
		flock.CurrentAge,
		flock.BarnID,
		flock.HealthStatus,
		flock.FeedTypeID,
		flock.Notes,
		flock.Audit.UpdatedAt,
		flock.Audit.UpdatedBy,
		flock.FlockID,
	)
	return err
}

func (r *SQLiteFlockRepo) SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error {
	const q = `UPDATE flocks SET deleted_at = ? WHERE flock_id = ?`
	_, err := r.DB.ExecContext(ctx, q, deletedAt, id)
	return err
}
