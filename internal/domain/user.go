package domain

import "time"

// User represents an application user used for session authentication only.
type User struct {
	ID                  int64
	Username            string
	PasswordHash        string
	ForcePasswordChange bool
	Theme               int // 0 = system, 1 = dark, 2 = light
	Audit               AuditFields
}

// MarkDeleted sets the soft-delete timestamp.
func (u *User) MarkDeleted(now time.Time) {
	u.Audit.DeletedAt = &now
}
