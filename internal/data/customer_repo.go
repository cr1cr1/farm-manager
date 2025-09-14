package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/cr1cr1/farm-manager/internal/domain"
)

// CustomerRepo defines operations for customer management.
type CustomerRepo interface {
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context) ([]*domain.Customer, error)
	FindByID(ctx context.Context, id int64) (*domain.Customer, error)
	Create(ctx context.Context, c *domain.Customer) (int64, error)
	Update(ctx context.Context, c *domain.Customer) error
	SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error
}

type SQLiteCustomerRepo struct {
	DB *sql.DB
}

func NewSQLiteCustomerRepo(db *sql.DB) *SQLiteCustomerRepo {
	return &SQLiteCustomerRepo{DB: db}
}

func (r *SQLiteCustomerRepo) Count(ctx context.Context) (int64, error) {
	const q = `SELECT COUNT(1) FROM customers WHERE deleted_at IS NULL`
	var n int64
	if err := r.DB.QueryRowContext(ctx, q).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *SQLiteCustomerRepo) List(ctx context.Context) ([]*domain.Customer, error) {
	const q = `SELECT customer_id, name, contact_info, delivery_address, customer_type, created_at, updated_at, deleted_at, created_by, updated_by FROM customers WHERE deleted_at IS NULL ORDER BY customer_id`
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*domain.Customer
	for rows.Next() {
		var item domain.Customer
		err := rows.Scan(
			&item.CustomerID,
			&item.Name,
			&item.ContactInfo,
			&item.DeliveryAddress,
			&item.CustomerType,
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

func (r *SQLiteCustomerRepo) FindByID(ctx context.Context, id int64) (*domain.Customer, error) {
	const q = `SELECT customer_id, name, contact_info, delivery_address, customer_type, created_at, updated_at, deleted_at, created_by, updated_by FROM customers WHERE customer_id = ? AND deleted_at IS NULL`
	var item domain.Customer
	err := r.DB.QueryRowContext(ctx, q, id).Scan(
		&item.CustomerID,
		&item.Name,
		&item.ContactInfo,
		&item.DeliveryAddress,
		&item.CustomerType,
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

func (r *SQLiteCustomerRepo) Create(ctx context.Context, c *domain.Customer) (int64, error) {
	const q = `INSERT INTO customers (name, contact_info, delivery_address, customer_type, created_at, updated_at, created_by, updated_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	c.Audit.CreatedAt = now
	c.Audit.UpdatedAt = now

	result, err := r.DB.ExecContext(ctx, q,
		c.Name,
		c.ContactInfo,
		c.DeliveryAddress,
		c.CustomerType,
		c.Audit.CreatedAt,
		c.Audit.UpdatedAt,
		c.Audit.CreatedBy,
		c.Audit.UpdatedBy,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *SQLiteCustomerRepo) Update(ctx context.Context, c *domain.Customer) error {
	const q = `UPDATE customers SET name = ?, contact_info = ?, delivery_address = ?, customer_type = ?, updated_at = ?, updated_by = ? WHERE customer_id = ? AND deleted_at IS NULL`
	now := time.Now()
	c.Audit.UpdatedAt = now

	_, err := r.DB.ExecContext(ctx, q,
		c.Name,
		c.ContactInfo,
		c.DeliveryAddress,
		c.CustomerType,
		now,
		c.Audit.UpdatedBy,
		c.CustomerID,
	)
	return err
}

func (r *SQLiteCustomerRepo) SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error {
	const q = `UPDATE customers SET deleted_at = ? WHERE customer_id = ?`
	_, err := r.DB.ExecContext(ctx, q, deletedAt, id)
	return err
}
