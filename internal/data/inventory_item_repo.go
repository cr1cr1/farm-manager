package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/cr1cr1/farm-manager/internal/domain"
)

// InventoryItemRepo defines operations for inventoryitem management.
type InventoryItemRepo interface {
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context) ([]*domain.InventoryItem, error)
	FindByID(ctx context.Context, id int64) (*domain.InventoryItem, error)
	Create(ctx context.Context, i *domain.InventoryItem) (int64, error)
	Update(ctx context.Context, i *domain.InventoryItem) error
	SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error
}

type SQLiteInventoryItemRepo struct {
	DB *sql.DB
}

func NewSQLiteInventoryItemRepo(db *sql.DB) *SQLiteInventoryItemRepo {
	return &SQLiteInventoryItemRepo{DB: db}
}

func (r *SQLiteInventoryItemRepo) Count(ctx context.Context) (int64, error) {
	const q = `SELECT COUNT(1) FROM inventory_items WHERE deleted_at IS NULL`
	var n int64
	if err := r.DB.QueryRowContext(ctx, q).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *SQLiteInventoryItemRepo) List(ctx context.Context) ([]*domain.InventoryItem, error) {
	const q = `SELECT inventory_item_id, name, type, quantity, unit, expiration_date, supplier_info, notes, created_at, updated_at, deleted_at, created_by, updated_by FROM inventory_items WHERE deleted_at IS NULL ORDER BY inventory_item_id`
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*domain.InventoryItem
	for rows.Next() {
		var item domain.InventoryItem
		err := rows.Scan(
			&item.InventoryItemID,
			&item.Name,
			&item.Type,
			&item.Quantity,
			&item.Unit,
			&item.ExpirationDate,
			&item.SupplierInfo,
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

func (r *SQLiteInventoryItemRepo) FindByID(ctx context.Context, id int64) (*domain.InventoryItem, error) {
	const q = `SELECT inventory_item_id, name, type, quantity, unit, expiration_date, supplier_info, notes, created_at, updated_at, deleted_at, created_by, updated_by FROM inventory_items WHERE inventory_item_id = ? AND deleted_at IS NULL`
	var item domain.InventoryItem
	err := r.DB.QueryRowContext(ctx, q, id).Scan(
		&item.InventoryItemID,
		&item.Name,
		&item.Type,
		&item.Quantity,
		&item.Unit,
		&item.ExpirationDate,
		&item.SupplierInfo,
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

func (r *SQLiteInventoryItemRepo) Create(ctx context.Context, i *domain.InventoryItem) (int64, error) {
	const q = `INSERT INTO inventory_items (name, type, quantity, unit, expiration_date, supplier_info, notes, created_at, updated_at, created_by, updated_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	i.Audit.CreatedAt = now
	i.Audit.UpdatedAt = now

	result, err := r.DB.ExecContext(ctx, q,
		i.Name,
		i.Type,
		i.Quantity,
		i.Unit,
		i.ExpirationDate,
		i.SupplierInfo,
		i.Notes,
		i.Audit.CreatedAt,
		i.Audit.UpdatedAt,
		i.Audit.CreatedBy,
		i.Audit.UpdatedBy,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *SQLiteInventoryItemRepo) Update(ctx context.Context, i *domain.InventoryItem) error {
	const q = `UPDATE inventory_items SET name = ?, type = ?, quantity = ?, unit = ?, expiration_date = ?, supplier_info = ?, notes = ?, updated_at = ?, updated_by = ? WHERE inventory_item_id = ? AND deleted_at IS NULL`
	i.Audit.UpdatedAt = time.Now()

	_, err := r.DB.ExecContext(ctx, q,
		i.Name,
		i.Type,
		i.Quantity,
		i.Unit,
		i.ExpirationDate,
		i.SupplierInfo,
		i.Notes,
		i.Audit.UpdatedAt,
		i.Audit.UpdatedBy,
		i.InventoryItemID,
	)
	return err
}

func (r *SQLiteInventoryItemRepo) SoftDelete(ctx context.Context, id int64, deletedAt time.Time) error {
	const q = `UPDATE inventory_items SET deleted_at = ? WHERE inventory_item_id = ?`
	_, err := r.DB.ExecContext(ctx, q, deletedAt, id)
	return err
}
