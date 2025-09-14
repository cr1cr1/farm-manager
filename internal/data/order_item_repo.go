package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/cr1cr1/farm-manager/internal/domain"
)

// OrderItemRepo defines operations for orderitem management.
type OrderItemRepo interface {
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context) ([]*domain.OrderItem, error)
	FindByID(ctx context.Context, id int64) (*domain.OrderItem, error)
	Create(ctx context.Context, o *domain.OrderItem) (int64, error)
	Update(ctx context.Context, o *domain.OrderItem) error
	SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error
}

type SQLiteOrderItemRepo struct {
	DB *sql.DB
}

func NewSQLiteOrderItemRepo(db *sql.DB) *SQLiteOrderItemRepo {
	return &SQLiteOrderItemRepo{DB: db}
}

func (r *SQLiteOrderItemRepo) Count(ctx context.Context) (int64, error) {
	const q = `SELECT COUNT(1) FROM order_items WHERE deleted_at IS NULL`
	var n int64
	if err := r.DB.QueryRowContext(ctx, q).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *SQLiteOrderItemRepo) List(ctx context.Context) ([]*domain.OrderItem, error) {
	const q = `SELECT order_item_id, order_id, product_description, quantity, unit_price, total_price, created_at, updated_at, deleted_at, created_by, updated_by FROM order_items WHERE deleted_at IS NULL ORDER BY order_item_id`
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		err := rows.Scan(
			&item.OrderItemID,
			&item.OrderID,
			&item.ProductDescription,
			&item.Quantity,
			&item.UnitPrice,
			&item.TotalPrice,
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

func (r *SQLiteOrderItemRepo) FindByID(ctx context.Context, id int64) (*domain.OrderItem, error) {
	const q = `SELECT order_item_id, order_id, product_description, quantity, unit_price, total_price, created_at, updated_at, deleted_at, created_by, updated_by FROM order_items WHERE order_item_id = ? AND deleted_at IS NULL`
	var item domain.OrderItem
	err := r.DB.QueryRowContext(ctx, q, id).Scan(
		&item.OrderItemID,
		&item.OrderID,
		&item.ProductDescription,
		&item.Quantity,
		&item.UnitPrice,
		&item.TotalPrice,
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

func (r *SQLiteOrderItemRepo) Create(ctx context.Context, o *domain.OrderItem) (int64, error) {
	const q = `INSERT INTO order_items (order_id, product_description, quantity, unit_price, total_price, created_at, updated_at, created_by, updated_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	o.Audit.CreatedAt = now
	o.Audit.UpdatedAt = now

	result, err := r.DB.ExecContext(ctx, q,
		o.OrderID,
		o.ProductDescription,
		o.Quantity,
		o.UnitPrice,
		o.TotalPrice,
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

func (r *SQLiteOrderItemRepo) Update(ctx context.Context, o *domain.OrderItem) error {
	const q = `UPDATE order_items SET order_id = ?, product_description = ?, quantity = ?, unit_price = ?, total_price = ?, updated_at = ?, updated_by = ? WHERE order_item_id = ? AND deleted_at IS NULL`
	o.Audit.UpdatedAt = time.Now()

	_, err := r.DB.ExecContext(ctx, q,
		o.OrderID,
		o.ProductDescription,
		o.Quantity,
		o.UnitPrice,
		o.TotalPrice,
		o.Audit.UpdatedAt,
		o.Audit.UpdatedBy,
		o.OrderItemID,
	)
	return err
}

func (r *SQLiteOrderItemRepo) SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error {
	const q = `UPDATE order_items SET deleted_at = ? WHERE order_item_id = ?`
	_, err := r.DB.ExecContext(ctx, q, deletedAt, id)
	return err
}
