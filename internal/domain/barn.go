package domain

// Barn represents a physical barn structure for housing flocks
type Barn struct {
	BarnID              int64
	Name                string
	Capacity            *int
	EnvironmentControl  *string
	MaintenanceSchedule *string
	Location            *string
	Audit               AuditFields
}
