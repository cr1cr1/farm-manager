package domain

import "time"

// SlaughterRecord tracks the slaughtering process for production batches
type SlaughterRecord struct {
	SlaughterID       int64
	BatchID           int64
	Date              *time.Time
	NumberSlaughtered *int
	MeatYield         *float64
	Waste             *float64
	StaffID           *int64
	Audit             AuditFields

	// Relations
	Batch *ProductionBatch
	Staff *Staff
}
