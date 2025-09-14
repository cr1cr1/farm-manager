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

type MortalityRecordManager struct {
	MortalityRecordRepo data.MortalityRecordRepo
	FlockRepo           data.FlockRepo
}

// RegisterMortalityRecordRoutes wires mortality record management endpoints under /app.
func RegisterMortalityRecordRoutes(group *ghttp.RouterGroup, mortalityRecordRepo data.MortalityRecordRepo, flockRepo data.FlockRepo) {
	mrm := &MortalityRecordManager{
		MortalityRecordRepo: mortalityRecordRepo,
		FlockRepo:           flockRepo,
	}

	// Mortality record management
	group.GET("/management/mortality-records", mrm.MortalityRecordsGet)
	group.POST("/management/mortality-records", mrm.MortalityRecordPost)
	group.GET("/management/mortality-records/new", mrm.MortalityRecordGet)
	group.GET("/management/mortality-records/:id", mrm.MortalityRecordGet)
	group.PUT("/management/mortality-records/:id", mrm.MortalityRecordPut)
	group.DELETE("/management/mortality-records/:id", mrm.MortalityRecordDelete)
}

// MortalityRecordsGet renders the mortality records management page.
func (mrm *MortalityRecordManager) MortalityRecordsGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	mortalityRecords, err := mrm.MortalityRecordRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list mortality records: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.MortalityRecordsContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				mortalityRecords,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.MortalityRecordsPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			mortalityRecords,
		),
	)
}

// MortalityRecordPost creates a new mortality record.
func (mrm *MortalityRecordManager) MortalityRecordPost(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	flockIDStr := strings.TrimSpace(r.Get("flock_id").String())
	dateStr := strings.TrimSpace(r.Get("date").String())
	numberDeadStr := strings.TrimSpace(r.Get("number_dead").String())
	causeOfDeath := strings.TrimSpace(r.Get("cause_of_death").String())
	notes := strings.TrimSpace(r.Get("notes").String())

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

	var date *time.Time
	if dateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateStr); err == nil {
			date = &parsedDate
		} else {
			errs["date"] = "Date must be a valid date (YYYY-MM-DD)"
		}
	}

	var numberDead *int
	if numberDeadStr != "" {
		if numVal, err := strconv.Atoi(numberDeadStr); err == nil {
			numberDead = &numVal
		} else {
			errs["number_dead"] = "Number dead must be a valid number"
		}
	}

	var cause *string
	if causeOfDeath != "" {
		cause = &causeOfDeath
	}

	var notesPtr *string
	if notes != "" {
		notesPtr = &notes
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		mortalityRecord := &domain.MortalityRecord{
			FlockID:      flockID,
			Date:         date,
			NumberDead:   numberDead,
			CauseOfDeath: cause,
			Notes:        notesPtr,
			Audit: domain.AuditFields{
				CreatedBy: &userIDStr,
				UpdatedBy: &userIDStr,
			},
		}

		_, err := mrm.MortalityRecordRepo.Create(r.GetCtx(), mortalityRecord)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "create mortality record: %v", err)
			errs["form"] = "Failed to create mortality record"
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
		r.Response.RedirectTo(middleware.BasePath() + "/management/mortality-records")
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/mortality-records")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/mortality-records")
}

// MortalityRecordGet renders a specific mortality record for editing or a new mortality record form.
func (mrm *MortalityRecordManager) MortalityRecordGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	var mortalityRecord *domain.MortalityRecord

	// Check if this is a request for a new mortality record (no ID provided)
	if idStr == "" || idStr == "new" {
		// This is a request for creating a new mortality record
		mortalityRecord = nil
	} else {
		// This is a request for editing an existing mortality record
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			r.Response.WriteStatusExit(400, "Invalid mortality record ID")
			return
		}

		mortalityRecord, err = mrm.MortalityRecordRepo.FindByID(r.GetCtx(), id)
		if err != nil {
			if err == data.ErrNotFound {
				r.Response.WriteStatusExit(404, "Mortality record not found")
				return
			}
			g.Log().Errorf(r.GetCtx(), "find mortality record: %v", err)
			r.Response.WriteStatusExit(500, "Internal server error")
			return
		}
	}

	flocks, err := mrm.FlockRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list flocks: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.MortalityRecordContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				mortalityRecord,
				flocks,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.MortalityRecordPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			ThemeToString(user.Theme),
			mortalityRecord,
			flocks,
		),
	)
}

// MortalityRecordPut updates an existing mortality record.
func (mrm *MortalityRecordManager) MortalityRecordPut(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid mortality record ID")
		return
	}

	flockIDStr := strings.TrimSpace(r.Get("flock_id").String())
	dateStr := strings.TrimSpace(r.Get("date").String())
	numberDeadStr := strings.TrimSpace(r.Get("number_dead").String())
	causeOfDeath := strings.TrimSpace(r.Get("cause_of_death").String())
	notes := strings.TrimSpace(r.Get("notes").String())

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

	var date *time.Time
	if dateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateStr); err == nil {
			date = &parsedDate
		} else {
			errs["date"] = "Date must be a valid date (YYYY-MM-DD)"
		}
	}

	var numberDead *int
	if numberDeadStr != "" {
		if numVal, err := strconv.Atoi(numberDeadStr); err == nil {
			numberDead = &numVal
		} else {
			errs["number_dead"] = "Number dead must be a valid number"
		}
	}

	var cause *string
	if causeOfDeath != "" {
		cause = &causeOfDeath
	}

	var notesPtr *string
	if notes != "" {
		notesPtr = &notes
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		mortalityRecord := &domain.MortalityRecord{
			MortalityRecordID: id,
			FlockID:           flockID,
			Date:              date,
			NumberDead:        numberDead,
			CauseOfDeath:      cause,
			Notes:             notesPtr,
			Audit: domain.AuditFields{
				UpdatedBy: &userIDStr,
			},
		}

		err := mrm.MortalityRecordRepo.Update(r.GetCtx(), mortalityRecord)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "update mortality record: %v", err)
			errs["form"] = "Failed to update mortality record"
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
		r.Response.RedirectTo(fmt.Sprintf("%s/management/mortality-records/%d", middleware.BasePath(), id))
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", fmt.Sprintf("%s/management/mortality-records/%d", middleware.BasePath(), id))
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the updated mortality record
	r.Response.RedirectTo(fmt.Sprintf("%s/management/mortality-records/%d", middleware.BasePath(), id))
}

// MortalityRecordDelete soft deletes a mortality record.
func (mrm *MortalityRecordManager) MortalityRecordDelete(r *ghttp.Request) {
	_, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid mortality record ID")
		return
	}

	err = mrm.MortalityRecordRepo.SoftDelete(r.GetCtx(), id, time.Now())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "delete mortality record: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/mortality-records")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/mortality-records")
}
