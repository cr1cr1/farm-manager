package domain

// Staff represents farm staff members
type Staff struct {
	StaffID     int64
	Name        string
	Role        *string
	Schedule    *string
	ContactInfo *string
	Audit       AuditFields
}
