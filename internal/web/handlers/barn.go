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

type BarnManager struct {
	BarnRepo data.BarnRepo
}

// RegisterBarnRoutes wires barn management endpoints under /app.
func RegisterBarnRoutes(group *ghttp.RouterGroup, barnRepo data.BarnRepo) {
	bm := &BarnManager{
		BarnRepo: barnRepo,
	}

	// Barn management
	group.GET("/management/barns", bm.BarnsGet)
	group.POST("/management/barns", bm.BarnPost)
	group.GET("/management/barns/new", bm.BarnGet)
	group.GET("/management/barns/:id", bm.BarnGet)
	group.PUT("/management/barns/:id", bm.BarnPut)
	group.DELETE("/management/barns/:id", bm.BarnDelete)
}

// BarnsGet renders the barns management page.
func (bm *BarnManager) BarnsGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	barns, err := bm.BarnRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list barns: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		// For DataStar requests, return only the content fragment
		_ = middleware.TemplRender(
			r,
			pages.BarnsContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				barns,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.BarnsPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			barns,
		),
	)
}

// BarnPost creates a new barn.
func (bm *BarnManager) BarnPost(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	name := strings.TrimSpace(r.Get("name").String())
	capacityStr := strings.TrimSpace(r.Get("capacity").String())
	environmentControl := strings.TrimSpace(r.Get("environment_control").String())
	maintenanceSchedule := strings.TrimSpace(r.Get("maintenance_schedule").String())
	location := strings.TrimSpace(r.Get("location").String())

	errs := map[string]string{}
	if name == "" {
		errs["name"] = "Name is required"
	}

	var capacity *int
	if capacityStr != "" {
		if capVal, err := strconv.Atoi(capacityStr); err == nil {
			capacity = new(int)
			*capacity = capVal
		} else {
			errs["capacity"] = "Capacity must be a valid number"
		}
	}

	var envControl *string
	if environmentControl != "" {
		envControl = new(string)
		*envControl = environmentControl
	}

	var maintSchedule *string
	if maintenanceSchedule != "" {
		maintSchedule = new(string)
		*maintSchedule = maintenanceSchedule
	}

	var loc *string
	if location != "" {
		loc = new(string)
		*loc = location
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		createdBy := new(string)
		*createdBy = userIDStr
		updatedBy := new(string)
		*updatedBy = userIDStr
		barn := &domain.Barn{
			Name:                name,
			Capacity:            capacity,
			EnvironmentControl:  envControl,
			MaintenanceSchedule: maintSchedule,
			Location:            loc,
			Audit: domain.AuditFields{
				CreatedBy: createdBy,
				UpdatedBy: updatedBy,
			},
		}

		_, err := bm.BarnRepo.Create(r.GetCtx(), barn)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "create barn: %v", err)
			errs["form"] = "Failed to create barn"
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
		r.Response.RedirectTo(middleware.BasePath() + "/management/barns")
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/barns")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/barns")
}

// BarnGet renders a specific barn for editing or a new barn form.
func (bm *BarnManager) BarnGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	var barn *domain.Barn

	// Check if this is a request for a new barn (no ID provided)
	if idStr == "" || idStr == "new" {
		// This is a request for creating a new barn
		barn = nil
	} else {
		// This is a request for editing an existing barn
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			r.Response.WriteStatusExit(400, "Invalid barn ID")
			return
		}

		barn, err = bm.BarnRepo.FindByID(r.GetCtx(), id)
		if err != nil {
			if err == data.ErrNotFound {
				r.Response.WriteStatusExit(404, "Barn not found")
				return
			}
			g.Log().Errorf(r.GetCtx(), "find barn: %v", err)
			r.Response.WriteStatusExit(500, "Internal server error")
			return
		}
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		// For DataStar requests, return only the content fragment
		_ = middleware.TemplRender(
			r,
			pages.BarnContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				barn,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.BarnPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			barn,
		),
	)
}

// BarnPut updates an existing barn.
func (bm *BarnManager) BarnPut(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid barn ID")
		return
	}

	name := strings.TrimSpace(r.Get("name").String())
	capacityStr := strings.TrimSpace(r.Get("capacity").String())
	environmentControl := strings.TrimSpace(r.Get("environment_control").String())
	maintenanceSchedule := strings.TrimSpace(r.Get("maintenance_schedule").String())
	location := strings.TrimSpace(r.Get("location").String())

	errs := map[string]string{}
	if name == "" {
		errs["name"] = "Name is required"
	}

	var capacity *int
	if capacityStr != "" {
		if capVal, err := strconv.Atoi(capacityStr); err == nil {
			capacity = new(int)
			*capacity = capVal
		} else {
			errs["capacity"] = "Capacity must be a valid number"
		}
	}

	var envControl *string
	if environmentControl != "" {
		envControl = new(string)
		*envControl = environmentControl
	}

	var maintSchedule *string
	if maintenanceSchedule != "" {
		maintSchedule = new(string)
		*maintSchedule = maintenanceSchedule
	}

	var loc *string
	if location != "" {
		loc = new(string)
		*loc = location
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		updatedBy := new(string)
		*updatedBy = userIDStr
		barn := &domain.Barn{
			BarnID:              id,
			Name:                name,
			Capacity:            capacity,
			EnvironmentControl:  envControl,
			MaintenanceSchedule: maintSchedule,
			Location:            loc,
			Audit: domain.AuditFields{
				UpdatedBy: updatedBy,
			},
		}

		err := bm.BarnRepo.Update(r.GetCtx(), barn)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "update barn: %v", err)
			errs["form"] = "Failed to update barn"
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
		r.Response.RedirectTo(fmt.Sprintf("%s/management/barns/%d", middleware.BasePath(), id))
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", fmt.Sprintf("%s/management/barns/%d", middleware.BasePath(), id))
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the updated barn
	r.Response.RedirectTo(fmt.Sprintf("%s/management/barns/%d", middleware.BasePath(), id))
}

// BarnDelete soft deletes a barn.
func (bm *BarnManager) BarnDelete(r *ghttp.Request) {
	_, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid barn ID")
		return
	}

	err = bm.BarnRepo.SoftDelete(r.GetCtx(), id, time.Now())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "delete barn: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/barns")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/barns")
}
