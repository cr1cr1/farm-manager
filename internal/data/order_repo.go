package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/cr1cr1/farm-manager/internal/domain"
)

// OrderRepo defines operations for order management.
type OrderRepo interface {
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context) ([]*domain.Order, error)
	FindByID(ctx context.Context, id int64) (*domain.Order, error)
	Create(ctx context.Context, o *domain.Order) (int64, error)
	Update(ctx context.Context, o *domain.Order) error
	SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error
}

type SQLiteOrderRepo struct {
	DB *sql.DB
}

func NewSQLiteOrderRepo(db *sql.DB) *SQLiteOrderRepo {
	return &SQLiteOrderRepo{DB: db}
}

func (r *SQLiteOrderRepo) Count(ctx context.Context) (int64, error) {
	const q = `SELECT COUNT(1) FROM orders WHERE deleted_at IS NULL`
	var n int64
	if err := r.DB.QueryRowContext(ctx, q).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *SQLiteOrderRepo) List(ctx context.Context) ([]*domain.Order, error) {
	const q = `SELECT order_id, customer_id, order_date, delivery_date, total_amount, status, created_at, updated_at, deleted_at, created_by, updated_by FROM orders WHERE deleted_at IS NULL ORDER BY order_id`
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*domain.Order
	for rows.Next() {
		var item domain.Order
		err := rows.Scan(
			&item.OrderID,
			&item.CustomerID,
			&item.OrderDate,
			&item.DeliveryDate,
			&item.TotalAmount,
			&item.Status,
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

func (r *SQLiteOrderRepo) FindByID(ctx context.Context, id int64) (*domain.Order, error) {
	const q = `SELECT order_id, customer_id, order_date, delivery_date, total_amount, status, created_at, updated_at, deleted_at, created_by, updated_by FROM orders WHERE order_id = ? AND deleted_at IS NULL`
	var item domain.Order
	err := r.DB.QueryRowContext(ctx, q, id).Scan(
		&item.OrderID,
		&item.CustomerID,
		&item.OrderDate,
		&item.DeliveryDate,
		&item.TotalAmount,
		&item.Status,
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

func (r *SQLiteOrderRepo) Create(ctx context.Context, o *domain.Order) (int64, error) {
	const q = `INSERT INTO orders (customer_id, order_date, delivery_date, total_amount, status, created_at, updated_at, created_by, updated_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	o.Audit.CreatedAt = now
	o.Audit.UpdatedAt = now

	result, err := r.DB.ExecContext(ctx, q,
		o.CustomerID,
		o.OrderDate,
		o.DeliveryDate,
		o.TotalAmount,
		o.Status,
		o.Audit.CreatedAt,
		o.Audit.UpdatedAt,
		o.Audit.CreatedBy,
		o.Audit.UpdatedBy,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *SQLiteOrderRepo) Update(ctx context.Context, o *domain.Order) error {
	const q = `UPDATE orders SET customer_id = ?, order_date = ?, delivery_date = ?, total_amount = ?, status = ?, updated_at = ?, updated_by = ? WHERE order_id = ? AND deleted_at IS NULL`
	now := time.Now()
	o.Audit.UpdatedAt = now

	_, err := r.DB.ExecContext(ctx, q,
		o.CustomerID,
		o.OrderDate,
		o.DeliveryDate,
		o.TotalAmount,
		o.Status,
		now,
		o.Audit.UpdatedBy,
		o.OrderID,
	)
	return err
}

func (r *SQLiteOrderRepo) SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error {
	const q = `UPDATE orders SET deleted_at = ? WHERE order_id = ?`
	_, err := r.DB.ExecContext(ctx, q, deletedAt, id)
	return err
}
