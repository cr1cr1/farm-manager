package handlers

import (
	"database/sql"
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

type FeedingRecordManager struct {
	FeedingRecordRepo data.FeedingRecordRepo
	FlockRepo         data.FlockRepo
	FeedTypeRepo      data.FeedTypeRepo
	StaffRepo         data.StaffRepo
}

// RegisterFeedingRecordRoutes wires feeding record management endpoints under /app.
func RegisterFeedingRecordRoutes(group *ghttp.RouterGroup, feedingRecordRepo data.FeedingRecordRepo, flockRepo data.FlockRepo, feedTypeRepo data.FeedTypeRepo, staffRepo data.StaffRepo) {
	frm := &FeedingRecordManager{
		FeedingRecordRepo: feedingRecordRepo,
		FlockRepo:         flockRepo,
		FeedTypeRepo:      feedTypeRepo,
		StaffRepo:         staffRepo,
	}

	// Feeding record management
	group.GET("/management/feeding-records", frm.FeedingRecordsGet)
	group.POST("/management/feeding-records", frm.FeedingRecordPost)
	group.GET("/management/feeding-records/new", frm.FeedingRecordGet)
	group.GET("/management/feeding-records/:id", frm.FeedingRecordGet)
	group.PUT("/management/feeding-records/:id", frm.FeedingRecordPut)
	group.DELETE("/management/feeding-records/:id", frm.FeedingRecordDelete)
}

// FeedingRecordsGet renders the feeding records management page.
func (frm *FeedingRecordManager) FeedingRecordsGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	feedingRecords, err := frm.FeedingRecordRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list feeding records: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.FeedingRecordsContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				feedingRecords,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.FeedingRecordsPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			feedingRecords,
		),
	)
}

// FeedingRecordPost creates a new feeding record.
func (frm *FeedingRecordManager) FeedingRecordPost(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	flockIDStr := strings.TrimSpace(r.Get("flock_id").String())
	feedTypeIDStr := strings.TrimSpace(r.Get("feed_type_id").String())
	amountGivenStr := strings.TrimSpace(r.Get("amount_given").String())
	dateTimeStr := strings.TrimSpace(r.Get("date_time").String())
	staffIDStr := strings.TrimSpace(r.Get("staff_id").String())

	errs := map[string]string{}
	if flockIDStr == "" {
		errs["flock_id"] = "Flock ID is required"
	}
	if feedTypeIDStr == "" {
		errs["feed_type_id"] = "Feed type ID is required"
	}

	var flockID int64
	if flockIDStr != "" {
		if idVal, err := strconv.ParseInt(flockIDStr, 10, 64); err == nil {
			flockID = idVal
		} else {
			errs["flock_id"] = "Flock ID must be a valid number"
		}
	}

	var feedTypeID int64
	if feedTypeIDStr != "" {
		if idVal, err := strconv.ParseInt(feedTypeIDStr, 10, 64); err == nil {
			feedTypeID = idVal
		} else {
			errs["feed_type_id"] = "Feed type ID must be a valid number"
		}
	}

	var amountGiven *float64
	if amountGivenStr != "" {
		if amtVal, err := strconv.ParseFloat(amountGivenStr, 64); err == nil {
			amountGiven = &amtVal
		} else {
			errs["amount_given"] = "Amount given must be a valid number"
		}
	}

	var dateTime sql.NullTime
	if dateTimeStr != "" {
		if parsedDateTime, err := time.Parse("2006-01-02T15:04", dateTimeStr); err == nil {
			dateTime = sql.NullTime{Time: parsedDateTime, Valid: true}
		} else {
			errs["date_time"] = "Date time must be a valid datetime (YYYY-MM-DDTHH:MM)"
		}
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
		feedingRecord := &domain.FeedingRecord{
			FlockID:     flockID,
			FeedTypeID:  feedTypeID,
			AmountGiven: amountGiven,
			DateTime:    dateTime,
			StaffID:     staffID,
			Audit: domain.AuditFields{
				CreatedBy: &userIDStr,
				UpdatedBy: &userIDStr,
			},
		}

		_, err := frm.FeedingRecordRepo.Create(r.GetCtx(), feedingRecord)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "create feeding record: %v", err)
			errs["form"] = "Failed to create feeding record"
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
		r.Response.RedirectTo(middleware.BasePath() + "/management/feeding-records")
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/feeding-records")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/feeding-records")
}

// FeedingRecordGet renders a specific feeding record for editing or a new feeding record form.
func (frm *FeedingRecordManager) FeedingRecordGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	var feedingRecord *domain.FeedingRecord

	// Check if this is a request for a new feeding record (no ID provided)
	if idStr == "" || idStr == "new" {
		// This is a request for creating a new feeding record
		feedingRecord = nil
	} else {
		// This is a request for editing an existing feeding record
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			r.Response.WriteStatusExit(400, "Invalid feeding record ID")
			return
		}

		feedingRecord, err = frm.FeedingRecordRepo.FindByID(r.GetCtx(), id)
		if err != nil {
			if err == data.ErrNotFound {
				r.Response.WriteStatusExit(404, "Feeding record not found")
				return
			}
			g.Log().Errorf(r.GetCtx(), "find feeding record: %v", err)
			r.Response.WriteStatusExit(500, "Internal server error")
			return
		}
	}

	flocks, err := frm.FlockRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list flocks: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	feedTypes, err := frm.FeedTypeRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list feed types: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	staff, err := frm.StaffRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list staff: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.FeedingRecordContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				feedingRecord,
				flocks,
				feedTypes,
				staff,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.FeedingRecordPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			feedingRecord,
			flocks,
			feedTypes,
			staff,
		),
	)
}

// FeedingRecordPut updates an existing feeding record.
func (frm *FeedingRecordManager) FeedingRecordPut(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid feeding record ID")
		return
	}

	flockIDStr := strings.TrimSpace(r.Get("flock_id").String())
	feedTypeIDStr := strings.TrimSpace(r.Get("feed_type_id").String())
	amountGivenStr := strings.TrimSpace(r.Get("amount_given").String())
	dateTimeStr := strings.TrimSpace(r.Get("date_time").String())
	staffIDStr := strings.TrimSpace(r.Get("staff_id").String())

	errs := map[string]string{}
	if flockIDStr == "" {
		errs["flock_id"] = "Flock ID is required"
	}
	if feedTypeIDStr == "" {
		errs["feed_type_id"] = "Feed type ID is required"
	}

	var flockID int64
	if flockIDStr != "" {
		if idVal, err := strconv.ParseInt(flockIDStr, 10, 64); err == nil {
			flockID = idVal
		} else {
			errs["flock_id"] = "Flock ID must be a valid number"
		}
	}

	var feedTypeID int64
	if feedTypeIDStr != "" {
		if idVal, err := strconv.ParseInt(feedTypeIDStr, 10, 64); err == nil {
			feedTypeID = idVal
		} else {
			errs["feed_type_id"] = "Feed type ID must be a valid number"
		}
	}

	var amountGiven *float64
	if amountGivenStr != "" {
		if amtVal, err := strconv.ParseFloat(amountGivenStr, 64); err == nil {
			amountGiven = &amtVal
		} else {
			errs["amount_given"] = "Amount given must be a valid number"
		}
	}

	var dateTime sql.NullTime
	if dateTimeStr != "" {
		if parsedDateTime, err := time.Parse("2006-01-02T15:04", dateTimeStr); err == nil {
			dateTime = sql.NullTime{Time: parsedDateTime, Valid: true}
		} else {
			errs["date_time"] = "Date time must be a valid datetime (YYYY-MM-DDTHH:MM)"
		}
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
		feedingRecord := &domain.FeedingRecord{
			FeedingRecordID: id,
			FlockID:         flockID,
			FeedTypeID:      feedTypeID,
			AmountGiven:     amountGiven,
			DateTime:        dateTime,
			StaffID:         staffID,
			Audit: domain.AuditFields{
				UpdatedBy: &userIDStr,
			},
		}

		err := frm.FeedingRecordRepo.Update(r.GetCtx(), feedingRecord)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "update feeding record: %v", err)
			errs["form"] = "Failed to update feeding record"
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
		r.Response.RedirectTo(fmt.Sprintf("%s/management/feeding-records/%d", middleware.BasePath(), id))
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", fmt.Sprintf("%s/management/feeding-records/%d", middleware.BasePath(), id))
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the updated feeding record
	r.Response.RedirectTo(fmt.Sprintf("%s/management/feeding-records/%d", middleware.BasePath(), id))
}

// FeedingRecordDelete soft deletes a feeding record.
func (frm *FeedingRecordManager) FeedingRecordDelete(r *ghttp.Request) {
	_, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid feeding record ID")
		return
	}

	err = frm.FeedingRecordRepo.SoftDelete(r.GetCtx(), id, time.Now())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "delete feeding record: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/feeding-records")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/feeding-records")
}
