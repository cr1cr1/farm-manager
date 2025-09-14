package domain

import "time"

// InventoryItem represents items stored in inventory
type InventoryItem struct {
	InventoryItemID int64
	Name            string
	Type            *string
	Quantity        *float64
	Unit            *string
	ExpirationDate  *time.Time
	SupplierInfo    *string
	Notes           *string
	Audit           AuditFields
}
