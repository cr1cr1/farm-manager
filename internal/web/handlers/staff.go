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

type StaffManager struct {
	StaffRepo data.StaffRepo
}

// RegisterStaffRoutes wires staff management endpoints under /app.
func RegisterStaffRoutes(group *ghttp.RouterGroup, staffRepo data.StaffRepo) {
	sm := &StaffManager{
		StaffRepo: staffRepo,
	}

	// Staff management
	group.GET("/management/staff", sm.StaffGet)
	group.POST("/management/staff", sm.StaffPost)
	group.GET("/management/staff/:id", sm.StaffGetByID)
	group.GET("/management/staff/new", sm.StaffGetByID)
	group.PUT("/management/staff/:id", sm.StaffPut)
	group.DELETE("/management/staff/:id", sm.StaffDelete)
}

// StaffGet renders the staff management page.
func (sm *StaffManager) StaffGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	staff, err := sm.StaffRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list staff: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.StaffContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				staff,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.StaffPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			staff,
		),
	)
}

// StaffPost creates a new staff member.
func (sm *StaffManager) StaffPost(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	name := strings.TrimSpace(r.Get("name").String())
	role := strings.TrimSpace(r.Get("role").String())
	schedule := strings.TrimSpace(r.Get("schedule").String())
	contactInfo := strings.TrimSpace(r.Get("contact_info").String())

	errs := map[string]string{}
	if name == "" {
		errs["name"] = "Name is required"
	}

	var rol *string
	if role != "" {
		rol = new(string)
		*rol = role
	}

	var sched *string
	if schedule != "" {
		sched = new(string)
		*sched = schedule
	}

	var contact *string
	if contactInfo != "" {
		contact = new(string)
		*contact = contactInfo
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		createdBy := new(string)
		*createdBy = userIDStr
		updatedBy := new(string)
		*updatedBy = userIDStr
		staff := &domain.Staff{
			Name:        name,
			Role:        rol,
			Schedule:    sched,
			ContactInfo: contact,
			Audit: domain.AuditFields{
				CreatedBy: createdBy,
				UpdatedBy: updatedBy,
			},
		}

		_, err := sm.StaffRepo.Create(r.GetCtx(), staff)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "create staff: %v", err)
			errs["form"] = "Failed to create staff member"
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
		r.Response.RedirectTo(middleware.BasePath() + "/management/staff")
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/staff")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/staff")
}

// StaffGetByID renders a specific staff member for editing.
func (sm *StaffManager) StaffGetByID(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	var staff *domain.Staff
	if idStr == "new" {
		staff = nil
	} else {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			r.Response.WriteStatusExit(400, "Invalid staff ID")
			return
		}

		staff, err = sm.StaffRepo.FindByID(r.GetCtx(), id)
		if err != nil {
			if err == data.ErrNotFound {
				r.Response.WriteStatusExit(404, "Staff member not found")
				return
			}
			g.Log().Errorf(r.GetCtx(), "find staff: %v", err)
			r.Response.WriteStatusExit(500, "Internal server error")
			return
		}
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.StaffMemberContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				staff,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.StaffMemberPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			staff,
		),
	)
}

// StaffPut updates an existing staff member.
func (sm *StaffManager) StaffPut(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid staff ID")
		return
	}

	name := strings.TrimSpace(r.Get("name").String())
	role := strings.TrimSpace(r.Get("role").String())
	schedule := strings.TrimSpace(r.Get("schedule").String())
	contactInfo := strings.TrimSpace(r.Get("contact_info").String())

	errs := map[string]string{}
	if name == "" {
		errs["name"] = "Name is required"
	}

	var rol *string
	if role != "" {
		rol = new(string)
		*rol = role
	}

	var sched *string
	if schedule != "" {
		sched = new(string)
		*sched = schedule
	}

	var contact *string
	if contactInfo != "" {
		contact = new(string)
		*contact = contactInfo
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		updatedBy := new(string)
		*updatedBy = userIDStr
		staff := &domain.Staff{
			StaffID:     id,
			Name:        name,
			Role:        rol,
			Schedule:    sched,
			ContactInfo: contact,
			Audit: domain.AuditFields{
				UpdatedBy: updatedBy,
			},
		}

		err := sm.StaffRepo.Update(r.GetCtx(), staff)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "update staff: %v", err)
			errs["form"] = "Failed to update staff member"
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
		r.Response.RedirectTo(fmt.Sprintf("%s/management/staff/%d", middleware.BasePath(), id))
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", fmt.Sprintf("%s/management/staff/%d", middleware.BasePath(), id))
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the updated staff member
	r.Response.RedirectTo(fmt.Sprintf("%s/management/staff/%d", middleware.BasePath(), id))
}

// StaffDelete soft deletes a staff member.
func (sm *StaffManager) StaffDelete(r *ghttp.Request) {
	_, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid staff ID")
		return
	}

	err = sm.StaffRepo.SoftDelete(r.GetCtx(), id, time.Now())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "delete staff: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/staff")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/staff")
}
