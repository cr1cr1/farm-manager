package domain

import "database/sql"

// FeedingRecord tracks feed given to flocks
type FeedingRecord struct {
	FeedingRecordID int64
	FlockID         int64
	FeedTypeID      int64
	AmountGiven     *float64
	DateTime        sql.NullTime
	StaffID         *int64
	Audit           AuditFields

	// Relations
	Flock    *Flock
	FeedType *FeedType
	Staff    *Staff
}
