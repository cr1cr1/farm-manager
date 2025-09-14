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

type ProductionBatchManager struct {
	ProductionBatchRepo data.ProductionBatchRepo
	FlockRepo           data.FlockRepo
	StaffRepo           data.StaffRepo
}

// RegisterProductionBatchRoutes wires production batch management endpoints under /app.
func RegisterProductionBatchRoutes(group *ghttp.RouterGroup, productionBatchRepo data.ProductionBatchRepo, flockRepo data.FlockRepo, staffRepo data.StaffRepo) {
	pbm := &ProductionBatchManager{
		ProductionBatchRepo: productionBatchRepo,
		FlockRepo:           flockRepo,
		StaffRepo:           staffRepo,
	}

	// Production batch management
	group.GET("/management/production-batches", pbm.ProductionBatchesGet)
	group.POST("/management/production-batches", pbm.ProductionBatchPost)
	group.GET("/management/production-batches/new", pbm.ProductionBatchGet)
	group.GET("/management/production-batches/:id", pbm.ProductionBatchGet)
	group.PUT("/management/production-batches/:id", pbm.ProductionBatchPut)
	group.DELETE("/management/production-batches/:id", pbm.ProductionBatchDelete)
}

// ProductionBatchesGet renders the production batches management page.
func (pbm *ProductionBatchManager) ProductionBatchesGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	productionBatches, err := pbm.ProductionBatchRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list production batches: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.ProductionBatchesContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				productionBatches,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.ProductionBatchesPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			productionBatches,
		),
	)
}

// ProductionBatchPost creates a new production batch.
func (pbm *ProductionBatchManager) ProductionBatchPost(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	flockIDStr := strings.TrimSpace(r.Get("flock_id").String())
	dateReadyStr := strings.TrimSpace(r.Get("date_ready").String())
	numberInBatchStr := strings.TrimSpace(r.Get("number_in_batch").String())
	weightEstimateStr := strings.TrimSpace(r.Get("weight_estimate").String())
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

	var dateReady *time.Time
	if dateReadyStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateReadyStr); err == nil {
			dateReady = &parsedDate
		} else {
			errs["date_ready"] = "Date ready must be a valid date (YYYY-MM-DD)"
		}
	}

	var numberInBatch *int
	if numberInBatchStr != "" {
		if numVal, err := strconv.Atoi(numberInBatchStr); err == nil {
			numberInBatch = &numVal
		} else {
			errs["number_in_batch"] = "Number in batch must be a valid number"
		}
	}

	var weightEstimate *float64
	if weightEstimateStr != "" {
		if weightVal, err := strconv.ParseFloat(weightEstimateStr, 64); err == nil {
			weightEstimate = &weightVal
		} else {
			errs["weight_estimate"] = "Weight estimate must be a valid number"
		}
	}

	var notesPtr *string
	if notes != "" {
		notesPtr = &notes
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		productionBatch := &domain.ProductionBatch{
			FlockID:        flockID,
			DateReady:      dateReady,
			NumberInBatch:  numberInBatch,
			WeightEstimate: weightEstimate,
			Notes:          notesPtr,
			Audit: domain.AuditFields{
				CreatedBy: &userIDStr,
				UpdatedBy: &userIDStr,
			},
		}

		_, err := pbm.ProductionBatchRepo.Create(r.GetCtx(), productionBatch)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "create production batch: %v", err)
			errs["form"] = "Failed to create production batch"
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
		r.Response.RedirectTo(middleware.BasePath() + "/management/production-batches")
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/production-batches")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/production-batches")
}

// ProductionBatchGet renders a specific production batch for editing or a new production batch form.
func (pbm *ProductionBatchManager) ProductionBatchGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	var productionBatch *domain.ProductionBatch

	// Check if this is a request for a new production batch (no ID provided)
	if idStr == "" || idStr == "new" {
		// This is a request for creating a new production batch
		productionBatch = nil
	} else {
		// This is a request for editing an existing production batch
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			r.Response.WriteStatusExit(400, "Invalid production batch ID")
			return
		}

		productionBatch, err = pbm.ProductionBatchRepo.FindByID(r.GetCtx(), id)
		if err != nil {
			if err == data.ErrNotFound {
				r.Response.WriteStatusExit(404, "Production batch not found")
				return
			}
			g.Log().Errorf(r.GetCtx(), "find production batch: %v", err)
			r.Response.WriteStatusExit(500, "Internal server error")
			return
		}
	}

	flocks, err := pbm.FlockRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list flocks: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	staff, err := pbm.StaffRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list staff: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.ProductionBatchContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				productionBatch,
				flocks,
				staff,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.ProductionBatchPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			ThemeToString(user.Theme),
			productionBatch,
			flocks,
			staff,
		),
	)
}

// ProductionBatchPut updates an existing production batch.
func (pbm *ProductionBatchManager) ProductionBatchPut(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid production batch ID")
		return
	}

	flockIDStr := strings.TrimSpace(r.Get("flock_id").String())
	dateReadyStr := strings.TrimSpace(r.Get("date_ready").String())
	numberInBatchStr := strings.TrimSpace(r.Get("number_in_batch").String())
	weightEstimateStr := strings.TrimSpace(r.Get("weight_estimate").String())
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

	var dateReady *time.Time
	if dateReadyStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateReadyStr); err == nil {
			dateReady = &parsedDate
		} else {
			errs["date_ready"] = "Date ready must be a valid date (YYYY-MM-DD)"
		}
	}

	var numberInBatch *int
	if numberInBatchStr != "" {
		if numVal, err := strconv.Atoi(numberInBatchStr); err == nil {
			numberInBatch = &numVal
		} else {
			errs["number_in_batch"] = "Number in batch must be a valid number"
		}
	}

	var weightEstimate *float64
	if weightEstimateStr != "" {
		if weightVal, err := strconv.ParseFloat(weightEstimateStr, 64); err == nil {
			weightEstimate = &weightVal
		} else {
			errs["weight_estimate"] = "Weight estimate must be a valid number"
		}
	}

	var notesPtr *string
	if notes != "" {
		notesPtr = &notes
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		productionBatch := &domain.ProductionBatch{
			BatchID:        id,
			FlockID:        flockID,
			DateReady:      dateReady,
			NumberInBatch:  numberInBatch,
			WeightEstimate: weightEstimate,
			Notes:          notesPtr,
			Audit: domain.AuditFields{
				UpdatedBy: &userIDStr,
			},
		}

		err := pbm.ProductionBatchRepo.Update(r.GetCtx(), productionBatch)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "update production batch: %v", err)
			errs["form"] = "Failed to update production batch"
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
		r.Response.RedirectTo(fmt.Sprintf("%s/management/production-batches/%d", middleware.BasePath(), id))
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", fmt.Sprintf("%s/management/production-batches/%d", middleware.BasePath(), id))
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the updated production batch
	r.Response.RedirectTo(fmt.Sprintf("%s/management/production-batches/%d", middleware.BasePath(), id))
}

// ProductionBatchDelete soft deletes a production batch.
func (pbm *ProductionBatchManager) ProductionBatchDelete(r *ghttp.Request) {
	_, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid production batch ID")
		return
	}

	err = pbm.ProductionBatchRepo.SoftDelete(r.GetCtx(), id, time.Now())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "delete production batch: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/production-batches")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/production-batches")
}
