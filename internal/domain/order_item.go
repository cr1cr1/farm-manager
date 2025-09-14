package domain

// OrderItem represents individual items within an order
type OrderItem struct {
	OrderItemID        int64
	OrderID            int64
	ProductDescription *string
	Quantity           *float64
	UnitPrice          *float64
	TotalPrice         *float64
	Audit              AuditFields

	// Relations
	Order *Order
}
