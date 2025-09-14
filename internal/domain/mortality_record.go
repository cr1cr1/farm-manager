package domain

import "time"

// MortalityRecord tracks bird deaths in flocks
type MortalityRecord struct {
	MortalityRecordID int64
	FlockID           int64
	Date              *time.Time
	NumberDead        *int
	CauseOfDeath      *string
	Notes             *string
	Audit             AuditFields

	// Relations
	Flock *Flock
}
