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

type SlaughterRecordManager struct {
	SlaughterRecordRepo data.SlaughterRecordRepo
	ProductionBatchRepo data.ProductionBatchRepo
	StaffRepo           data.StaffRepo
}

// RegisterSlaughterRecordRoutes wires slaughter record management endpoints under /app.
func RegisterSlaughterRecordRoutes(group *ghttp.RouterGroup, slaughterRecordRepo data.SlaughterRecordRepo, productionBatchRepo data.ProductionBatchRepo, staffRepo data.StaffRepo) {
	srm := &SlaughterRecordManager{
		SlaughterRecordRepo: slaughterRecordRepo,
		ProductionBatchRepo: productionBatchRepo,
		StaffRepo:           staffRepo,
	}

	// Slaughter record management
	group.GET("/management/slaughter-records", srm.SlaughterRecordsGet)
	group.POST("/management/slaughter-records", srm.SlaughterRecordPost)
	group.GET("/management/slaughter-records/new", srm.SlaughterRecordGet)
	group.GET("/management/slaughter-records/:id", srm.SlaughterRecordGet)
	group.PUT("/management/slaughter-records/:id", srm.SlaughterRecordPut)
	group.DELETE("/management/slaughter-records/:id", srm.SlaughterRecordDelete)
}

// SlaughterRecordsGet renders the slaughter records management page.
func (srm *SlaughterRecordManager) SlaughterRecordsGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	slaughterRecords, err := srm.SlaughterRecordRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list slaughter records: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.SlaughterRecordsContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				slaughterRecords,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.SlaughterRecordsPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			slaughterRecords,
		),
	)
}

// SlaughterRecordPost creates a new slaughter record.
func (srm *SlaughterRecordManager) SlaughterRecordPost(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	batchIDStr := strings.TrimSpace(r.Get("batch_id").String())
	dateStr := strings.TrimSpace(r.Get("date").String())
	numberSlaughteredStr := strings.TrimSpace(r.Get("number_slaughtered").String())
	meatYieldStr := strings.TrimSpace(r.Get("meat_yield").String())
	wasteStr := strings.TrimSpace(r.Get("waste").String())
	staffIDStr := strings.TrimSpace(r.Get("staff_id").String())

	errs := map[string]string{}
	if batchIDStr == "" {
		errs["batch_id"] = "Batch ID is required"
	}

	var batchID int64
	if batchIDStr != "" {
		if idVal, err := strconv.ParseInt(batchIDStr, 10, 64); err == nil {
			batchID = idVal
		} else {
			errs["batch_id"] = "Batch ID must be a valid number"
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

	var numberSlaughtered *int
	if numberSlaughteredStr != "" {
		if numVal, err := strconv.Atoi(numberSlaughteredStr); err == nil {
			numberSlaughtered = &numVal
		} else {
			errs["number_slaughtered"] = "Number slaughtered must be a valid number"
		}
	}

	var meatYield *float64
	if meatYieldStr != "" {
		if yieldVal, err := strconv.ParseFloat(meatYieldStr, 64); err == nil {
			meatYield = &yieldVal
		} else {
			errs["meat_yield"] = "Meat yield must be a valid number"
		}
	}

	var waste *float64
	if wasteStr != "" {
		if wasteVal, err := strconv.ParseFloat(wasteStr, 64); err == nil {
			waste = &wasteVal
		} else {
			errs["waste"] = "Waste must be a valid number"
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
		slaughterRecord := &domain.SlaughterRecord{
			BatchID:           batchID,
			Date:              date,
			NumberSlaughtered: numberSlaughtered,
			MeatYield:         meatYield,
			Waste:             waste,
			StaffID:           staffID,
			Audit: domain.AuditFields{
				CreatedBy: &userIDStr,
				UpdatedBy: &userIDStr,
			},
		}

		_, err := srm.SlaughterRecordRepo.Create(r.GetCtx(), slaughterRecord)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "create slaughter record: %v", err)
			errs["form"] = "Failed to create slaughter record"
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
		r.Response.RedirectTo(middleware.BasePath() + "/management/slaughter-records")
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/slaughter-records")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/slaughter-records")
}

// SlaughterRecordGet renders a specific slaughter record for editing or a new slaughter record form.
func (srm *SlaughterRecordManager) SlaughterRecordGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	var slaughterRecord *domain.SlaughterRecord

	// Check if this is a request for a new slaughter record (no ID provided)
	if idStr == "" || idStr == "new" {
		// This is a request for creating a new slaughter record
		slaughterRecord = nil
	} else {
		// This is a request for editing an existing slaughter record
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			r.Response.WriteStatusExit(400, "Invalid slaughter record ID")
			return
		}

		slaughterRecord, err = srm.SlaughterRecordRepo.FindByID(r.GetCtx(), id)
		if err != nil {
			if err == data.ErrNotFound {
				r.Response.WriteStatusExit(404, "Slaughter record not found")
				return
			}
			g.Log().Errorf(r.GetCtx(), "find slaughter record: %v", err)
			r.Response.WriteStatusExit(500, "Internal server error")
			return
		}
	}

	// Fetch production batches and staff for dropdowns
	productionBatches, err := srm.ProductionBatchRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list production batches: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	staff, err := srm.StaffRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list staff: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.SlaughterRecordContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				slaughterRecord,
				productionBatches,
				staff,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.SlaughterRecordPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			slaughterRecord,
			productionBatches,
			staff,
		),
	)
}

// SlaughterRecordPut updates an existing slaughter record.
func (srm *SlaughterRecordManager) SlaughterRecordPut(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid slaughter record ID")
		return
	}

	batchIDStr := strings.TrimSpace(r.Get("batch_id").String())
	dateStr := strings.TrimSpace(r.Get("date").String())
	numberSlaughteredStr := strings.TrimSpace(r.Get("number_slaughtered").String())
	meatYieldStr := strings.TrimSpace(r.Get("meat_yield").String())
	wasteStr := strings.TrimSpace(r.Get("waste").String())
	staffIDStr := strings.TrimSpace(r.Get("staff_id").String())

	errs := map[string]string{}
	if batchIDStr == "" {
		errs["batch_id"] = "Batch ID is required"
	}

	var batchID int64
	if batchIDStr != "" {
		if idVal, err := strconv.ParseInt(batchIDStr, 10, 64); err == nil {
			batchID = idVal
		} else {
			errs["batch_id"] = "Batch ID must be a valid number"
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

	var numberSlaughtered *int
	if numberSlaughteredStr != "" {
		if numVal, err := strconv.Atoi(numberSlaughteredStr); err == nil {
			numberSlaughtered = &numVal
		} else {
			errs["number_slaughtered"] = "Number slaughtered must be a valid number"
		}
	}

	var meatYield *float64
	if meatYieldStr != "" {
		if yieldVal, err := strconv.ParseFloat(meatYieldStr, 64); err == nil {
			meatYield = &yieldVal
		} else {
			errs["meat_yield"] = "Meat yield must be a valid number"
		}
	}

	var waste *float64
	if wasteStr != "" {
		if wasteVal, err := strconv.ParseFloat(wasteStr, 64); err == nil {
			waste = &wasteVal
		} else {
			errs["waste"] = "Waste must be a valid number"
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
		slaughterRecord := &domain.SlaughterRecord{
			SlaughterID:       id,
			BatchID:           batchID,
			Date:              date,
			NumberSlaughtered: numberSlaughtered,
			MeatYield:         meatYield,
			Waste:             waste,
			StaffID:           staffID,
			Audit: domain.AuditFields{
				UpdatedBy: &userIDStr,
			},
		}

		err := srm.SlaughterRecordRepo.Update(r.GetCtx(), slaughterRecord)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "update slaughter record: %v", err)
			errs["form"] = "Failed to update slaughter record"
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
		r.Response.RedirectTo(fmt.Sprintf("%s/management/slaughter-records/%d", middleware.BasePath(), id))
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", fmt.Sprintf("%s/management/slaughter-records/%d", middleware.BasePath(), id))
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the updated slaughter record
	r.Response.RedirectTo(fmt.Sprintf("%s/management/slaughter-records/%d", middleware.BasePath(), id))
}

// SlaughterRecordDelete soft deletes a slaughter record.
func (srm *SlaughterRecordManager) SlaughterRecordDelete(r *ghttp.Request) {
	_, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid slaughter record ID")
		return
	}

	err = srm.SlaughterRecordRepo.SoftDelete(r.GetCtx(), id, time.Now())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "delete slaughter record: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/slaughter-records")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/slaughter-records")
}
