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

type FeedTypeManager struct {
	FeedTypeRepo data.FeedTypeRepo
}

// RegisterFeedTypeRoutes wires feed type management endpoints under /app.
func RegisterFeedTypeRoutes(group *ghttp.RouterGroup, feedTypeRepo data.FeedTypeRepo) {
	ftm := &FeedTypeManager{
		FeedTypeRepo: feedTypeRepo,
	}

	// FeedType management
	group.GET("/management/feed-types", ftm.FeedTypesGet)
	group.POST("/management/feed-types", ftm.FeedTypePost)
	group.GET("/management/feed-types/new", ftm.FeedTypeGet)
	group.GET("/management/feed-types/:id", ftm.FeedTypeGet)
	group.PUT("/management/feed-types/:id", ftm.FeedTypePut)
	group.DELETE("/management/feed-types/:id", ftm.FeedTypeDelete)
}

// FeedTypesGet renders the feed types management page.
func (ftm *FeedTypeManager) FeedTypesGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	feedTypes, err := ftm.FeedTypeRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list feed types: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		// For DataStar requests, return only the content fragment
		_ = middleware.TemplRender(
			r,
			pages.FeedTypesContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				feedTypes,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.FeedTypesPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			feedTypes,
		),
	)
}

// FeedTypePost creates a new feed type.
func (ftm *FeedTypeManager) FeedTypePost(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	name := strings.TrimSpace(r.Get("name").String())
	description := strings.TrimSpace(r.Get("description").String())
	nutritionalInfo := strings.TrimSpace(r.Get("nutritional_info").String())

	errs := map[string]string{}
	if name == "" {
		errs["name"] = "Name is required"
	}

	var desc *string
	if description != "" {
		desc = new(string)
		*desc = description
	}

	var nutritional *string
	if nutritionalInfo != "" {
		nutritional = new(string)
		*nutritional = nutritionalInfo
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		createdBy := new(string)
		*createdBy = userIDStr
		updatedBy := new(string)
		*updatedBy = userIDStr
		feedType := &domain.FeedType{
			Name:            name,
			Description:     desc,
			NutritionalInfo: nutritional,
			Audit: domain.AuditFields{
				CreatedBy: createdBy,
				UpdatedBy: updatedBy,
			},
		}

		_, err := ftm.FeedTypeRepo.Create(r.GetCtx(), feedType)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "create feed type: %v", err)
			errs["form"] = "Failed to create feed type"
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
		r.Response.RedirectTo(middleware.BasePath() + "/management/feed-types")
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/feed-types")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/feed-types")
}

// FeedTypeGet renders a specific feed type for editing or a new feed type form.
func (ftm *FeedTypeManager) FeedTypeGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	var feedType *domain.FeedType

	// Check if this is a request for a new feed type (no ID provided)
	if idStr == "" || idStr == "new" {
		// This is a request for creating a new feed type
		feedType = nil
	} else {
		// This is a request for editing an existing feed type
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			r.Response.WriteStatusExit(400, "Invalid feed type ID")
			return
		}

		feedType, err = ftm.FeedTypeRepo.FindByID(r.GetCtx(), id)
		if err != nil {
			if err == data.ErrNotFound {
				r.Response.WriteStatusExit(404, "Feed type not found")
				return
			}
			g.Log().Errorf(r.GetCtx(), "find feed type: %v", err)
			r.Response.WriteStatusExit(500, "Internal server error")
			return
		}
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		// For DataStar requests, return only the content fragment
		_ = middleware.TemplRender(
			r,
			pages.FeedTypeContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				feedType,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.FeedTypePage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			feedType,
		),
	)
}

// FeedTypePut updates an existing feed type.
func (ftm *FeedTypeManager) FeedTypePut(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid feed type ID")
		return
	}

	name := strings.TrimSpace(r.Get("name").String())
	description := strings.TrimSpace(r.Get("description").String())
	nutritionalInfo := strings.TrimSpace(r.Get("nutritional_info").String())

	errs := map[string]string{}
	if name == "" {
		errs["name"] = "Name is required"
	}

	var desc *string
	if description != "" {
		desc = new(string)
		*desc = description
	}

	var nutritional *string
	if nutritionalInfo != "" {
		nutritional = new(string)
		*nutritional = nutritionalInfo
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		updatedBy := new(string)
		*updatedBy = userIDStr
		feedType := &domain.FeedType{
			FeedTypeID:      id,
			Name:            name,
			Description:     desc,
			NutritionalInfo: nutritional,
			Audit: domain.AuditFields{
				UpdatedBy: updatedBy,
			},
		}

		err := ftm.FeedTypeRepo.Update(r.GetCtx(), feedType)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "update feed type: %v", err)
			errs["form"] = "Failed to update feed type"
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
		r.Response.RedirectTo(fmt.Sprintf("%s/management/feed-types/%d", middleware.BasePath(), id))
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", fmt.Sprintf("%s/management/feed-types/%d", middleware.BasePath(), id))
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the updated feed type
	r.Response.RedirectTo(fmt.Sprintf("%s/management/feed-types/%d", middleware.BasePath(), id))
}

// FeedTypeDelete soft deletes a feed type.
func (ftm *FeedTypeManager) FeedTypeDelete(r *ghttp.Request) {
	_, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid feed type ID")
		return
	}

	err = ftm.FeedTypeRepo.SoftDelete(r.GetCtx(), id, time.Now())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "delete feed type: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/feed-types")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/feed-types")
}
