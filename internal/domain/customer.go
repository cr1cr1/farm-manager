package domain

// Customer represents customers who purchase products
type Customer struct {
	CustomerID      int64
	Name            string
	ContactInfo     *string
	DeliveryAddress *string
	CustomerType    *string
	Audit           AuditFields
}
