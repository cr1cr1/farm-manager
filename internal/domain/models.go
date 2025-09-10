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

// User represents an application user used for session authentication only.
type User struct {
	ID                  int64
	Username            string
	PasswordHash        string
	ForcePasswordChange bool
	Audit               AuditFields
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

// MarkDeleted sets the soft-delete timestamp.
func (u *User) MarkDeleted(now time.Time) {
	u.Audit.DeletedAt = &now
}
