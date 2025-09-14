package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cr1cr1/farm-manager/internal/data"
	"github.com/cr1cr1/farm-manager/internal/domain"
	"github.com/cr1cr1/farm-manager/internal/web/middleware"
	"github.com/cr1cr1/farm-manager/internal/web/templates/pages"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type HealthCheckManager struct {
	HealthCheckRepo data.HealthCheckRepo
	FlockRepo       data.FlockRepo
	StaffRepo       data.StaffRepo
}

// RegisterHealthCheckRoutes wires health check management endpoints under /app.
func RegisterHealthCheckRoutes(group *ghttp.RouterGroup, healthCheckRepo data.HealthCheckRepo, flockRepo data.FlockRepo, staffRepo data.StaffRepo) {
	hcm := &HealthCheckManager{
		HealthCheckRepo: healthCheckRepo,
		FlockRepo:       flockRepo,
		StaffRepo:       staffRepo,
	}

	// Health check management
	group.GET("/management/health-checks", hcm.HealthChecksGet)
	group.POST("/management/health-checks", hcm.HealthCheckPost)
	group.GET("/management/health-checks/new", hcm.HealthCheckGet)
	group.GET("/management/health-checks/:id", hcm.HealthCheckGet)
	group.PUT("/management/health-checks/:id", hcm.HealthCheckPut)
	group.DELETE("/management/health-checks/:id", hcm.HealthCheckDelete)
}

// HealthChecksGet renders the health checks management page.
func (hcm *HealthCheckManager) HealthChecksGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	healthChecks, err := hcm.HealthCheckRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list health checks: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.HealthChecksContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				healthChecks,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.HealthChecksPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			healthChecks,
		),
	)
}

// HealthCheckPost creates a new health check.
func (hcm *HealthCheckManager) HealthCheckPost(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	flockIDStr := strings.TrimSpace(r.Get("flock_id").String())
	checkDateStr := strings.TrimSpace(r.Get("check_date").String())
	healthStatus := strings.TrimSpace(r.Get("health_status").String())
	vaccinationsGiven := strings.TrimSpace(r.Get("vaccinations_given").String())
	treatmentsAdministered := strings.TrimSpace(r.Get("treatments_administered").String())
	notes := strings.TrimSpace(r.Get("notes").String())
	staffIDStr := strings.TrimSpace(r.Get("staff_id").String())

	errs := map[string]string{}
	if flockIDStr == "" {
		errs["flock_id"] = "Flock ID is required"
	}

	var flockID int64
	if flockIDStr != "" {
		if idVal, err := strconv.ParseInt(flockIDStr, 10, 64); err == nil {
			flockID = idVal
		} else {
			errs["flock_id"] = "Flock ID must be a valid number"
		}
	}

	var checkDate *time.Time
	if checkDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", checkDateStr); err == nil {
			checkDate = &parsedDate
		} else {
			errs["check_date"] = "Check date must be a valid date (YYYY-MM-DD)"
		}
	}

	var healthStat *string
	if healthStatus != "" {
		healthStat = &healthStatus
	}

	var vaccinations *string
	if vaccinationsGiven != "" {
		vaccinations = &vaccinationsGiven
	}

	var treatments *string
	if treatmentsAdministered != "" {
		treatments = &treatmentsAdministered
	}

	var notesPtr *string
	if notes != "" {
		notesPtr = &notes
	}

	var staffID *int64
	if staffIDStr != "" {
		if idVal, err := strconv.ParseInt(staffIDStr, 10, 64); err == nil {
			staffID = &idVal
		} else {
			errs["staff_id"] = "Staff ID must be a valid number"
		}
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		healthCheck := &domain.HealthCheck{
			FlockID:                flockID,
			CheckDate:              checkDate,
			HealthStatus:           healthStat,
			VaccinationsGiven:      vaccinations,
			TreatmentsAdministered: treatments,
			Notes:                  notesPtr,
			StaffID:                staffID,
			Audit: domain.AuditFields{
				CreatedBy: &userIDStr,
				UpdatedBy: &userIDStr,
			},
		}

		_, err := hcm.HealthCheckRepo.Create(r.GetCtx(), healthCheck)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "create health check: %v", err)
			errs["form"] = "Failed to create health check"
		}
	}

	if len(errs) > 0 {
		if isDataStarRequest {
			// For DataStar requests, return validation errors
			r.Response.Header().Set("Content-Type", "application/json")
			r.Response.WriteJson(map[string]interface{}{
				"errors": errs,
			})
			return
		}
		// For regular requests, redirect back with errors
		r.Response.RedirectTo(middleware.BasePath() + "/management/health-checks")
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/health-checks")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/health-checks")
}

// HealthCheckGet renders a specific health check for editing or a new health check form.
func (hcm *HealthCheckManager) HealthCheckGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	var healthCheck *domain.HealthCheck

	// Check if this is a request for a new health check (no ID provided)
	if idStr == "" || idStr == "new" {
		// This is a request for creating a new health check
		healthCheck = nil
	} else {
		// This is a request for editing an existing health check
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			r.Response.WriteStatusExit(400, "Invalid health check ID")
			return
		}

		healthCheck, err = hcm.HealthCheckRepo.FindByID(r.GetCtx(), id)
		if err != nil {
			if err == data.ErrNotFound {
				r.Response.WriteStatusExit(404, "Health check not found")
				return
			}
			g.Log().Errorf(r.GetCtx(), "find health check: %v", err)
			r.Response.WriteStatusExit(500, "Internal server error")
			return
		}
	}

	flocks, err := hcm.FlockRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list flocks: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	staff, err := hcm.StaffRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list staff: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.HealthCheckContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				healthCheck,
				flocks,
				staff,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.HealthCheckPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			healthCheck,
			flocks,
			staff,
		),
	)
}

// HealthCheckPut updates an existing health check.
func (hcm *HealthCheckManager) HealthCheckPut(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid health check ID")
		return
	}

	flockIDStr := strings.TrimSpace(r.Get("flock_id").String())
	checkDateStr := strings.TrimSpace(r.Get("check_date").String())
	healthStatus := strings.TrimSpace(r.Get("health_status").String())
	vaccinationsGiven := strings.TrimSpace(r.Get("vaccinations_given").String())
	treatmentsAdministered := strings.TrimSpace(r.Get("treatments_administered").String())
	notes := strings.TrimSpace(r.Get("notes").String())
	staffIDStr := strings.TrimSpace(r.Get("staff_id").String())

	errs := map[string]string{}
	if flockIDStr == "" {
		errs["flock_id"] = "Flock ID is required"
	}

	var flockID int64
	if flockIDStr != "" {
		if idVal, err := strconv.ParseInt(flockIDStr, 10, 64); err == nil {
			flockID = idVal
		} else {
			errs["flock_id"] = "Flock ID must be a valid number"
		}
	}

	var checkDate *time.Time
	if checkDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", checkDateStr); err == nil {
			checkDate = &parsedDate
		} else {
			errs["check_date"] = "Check date must be a valid date (YYYY-MM-DD)"
		}
	}

	var healthStat *string
	if healthStatus != "" {
		healthStat = &healthStatus
	}

	var vaccinations *string
	if vaccinationsGiven != "" {
		vaccinations = &vaccinationsGiven
	}

	var treatments *string
	if treatmentsAdministered != "" {
		treatments = &treatmentsAdministered
	}

	var notesPtr *string
	if notes != "" {
		notesPtr = &notes
	}

	var staffID *int64
	if staffIDStr != "" {
		if idVal, err := strconv.ParseInt(staffIDStr, 10, 64); err == nil {
			staffID = &idVal
		} else {
			errs["staff_id"] = "Staff ID must be a valid number"
		}
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		healthCheck := &domain.HealthCheck{
			HealthCheckID:          id,
			FlockID:                flockID,
			CheckDate:              checkDate,
			HealthStatus:           healthStat,
			VaccinationsGiven:      vaccinations,
			TreatmentsAdministered: treatments,
			Notes:                  notesPtr,
			StaffID:                staffID,
			Audit: domain.AuditFields{
				UpdatedBy: &userIDStr,
			},
		}

		err := hcm.HealthCheckRepo.Update(r.GetCtx(), healthCheck)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "update health check: %v", err)
			errs["form"] = "Failed to update health check"
		}
	}

	if len(errs) > 0 {
		if isDataStarRequest {
			// For DataStar requests, return validation errors
			r.Response.Header().Set("Content-Type", "application/json")
			r.Response.WriteJson(map[string]interface{}{
				"errors": errs,
			})
			return
		}
		// For regular requests, redirect back with errors
		r.Response.RedirectTo(fmt.Sprintf("%s/management/health-checks/%d", middleware.BasePath(), id))
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", fmt.Sprintf("%s/management/health-checks/%d", middleware.BasePath(), id))
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the updated health check
	r.Response.RedirectTo(fmt.Sprintf("%s/management/health-checks/%d", middleware.BasePath(), id))
}

// HealthCheckDelete soft deletes a health check.
func (hcm *HealthCheckManager) HealthCheckDelete(r *ghttp.Request) {
	_, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid health check ID")
		return
	}

	err = hcm.HealthCheckRepo.SoftDelete(r.GetCtx(), id, time.Now())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "delete health check: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/health-checks")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/health-checks")
}
