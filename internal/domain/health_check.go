package domain

import "time"

// HealthCheck represents health inspections and treatments for flocks
type HealthCheck struct {
	HealthCheckID          int64
	FlockID                int64
	CheckDate              *time.Time
	HealthStatus           *string
	VaccinationsGiven      *string
	TreatmentsAdministered *string
	Notes                  *string
	StaffID                *int64
	Audit                  AuditFields

	// Relations
	Flock *Flock
	Staff *Staff
}
