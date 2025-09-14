package domain

import "time"

// AuditFields captures common auditing metadata for records.
type AuditFields struct {
	CreatedAt time.Time  // NOT NULL
	UpdatedAt time.Time  // NOT NULL
	DeletedAt *time.Time // NULL
	CreatedBy *string    // NULL
	UpdatedBy *string    // NULL
}

// TouchCreated sets created/updated timestamps to now.
func (a *AuditFields) TouchCreated(now time.Time) {
	a.CreatedAt = now
	a.UpdatedAt = now
}

// TouchUpdated updates the updated timestamp to now.
func (a *AuditFields) TouchUpdated(now time.Time) {
	a.UpdatedAt = now
}
