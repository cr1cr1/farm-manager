package domain

import "time"

// Flock represents a group of birds being raised together
type Flock struct {
	FlockID       int64
	Breed         string
	HatchDate     *time.Time
	NumberOfBirds *int
	CurrentAge    *int
	BarnID        *int64
	HealthStatus  *string
	FeedTypeID    *int64
	Notes         *string
	Audit         AuditFields

	// Relations
	Barn     *Barn
	FeedType *FeedType
}
