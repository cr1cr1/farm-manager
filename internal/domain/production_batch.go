package domain

import "time"

// ProductionBatch represents batches of birds ready for processing
type ProductionBatch struct {
	BatchID        int64
	FlockID        int64
	DateReady      *time.Time
	NumberInBatch  *int
	WeightEstimate *float64
	Notes          *string
	Audit          AuditFields

	// Relations
	Flock *Flock
}
