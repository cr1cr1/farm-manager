package domain

import "time"

// Order represents customer orders
type Order struct {
	OrderID      int64
	CustomerID   int64
	OrderDate    *time.Time
	DeliveryDate *time.Time
	TotalAmount  *float64
	Status       *string
	Audit        AuditFields

	// Relations
	Customer *Customer
}
