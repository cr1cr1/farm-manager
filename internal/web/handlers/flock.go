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

type FlockManager struct {
	FlockRepo    data.FlockRepo
	BarnRepo     data.BarnRepo
	FeedTypeRepo data.FeedTypeRepo
}

// RegisterFlockRoutes wires flock management endpoints under /app.
func RegisterFlockRoutes(group *ghttp.RouterGroup, flockRepo data.FlockRepo, barnRepo data.BarnRepo, feedTypeRepo data.FeedTypeRepo) {
	fm := &FlockManager{
		FlockRepo:    flockRepo,
		BarnRepo:     barnRepo,
		FeedTypeRepo: feedTypeRepo,
	}

	// Flock management
	group.GET("/management/flocks", fm.FlocksGet)
	group.POST("/management/flocks", fm.FlockPost)
	group.GET("/management/flocks/new", fm.FlockGet)
	group.GET("/management/flocks/:id", fm.FlockGet)
	group.PUT("/management/flocks/:id", fm.FlockPut)
	group.DELETE("/management/flocks/:id", fm.FlockDelete)
}

// FlocksGet renders the flocks management page.
func (fm *FlockManager) FlocksGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	flocks, err := fm.FlockRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list flocks: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.FlocksContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				flocks,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.FlocksPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			flocks,
		),
	)
}

// FlockPost creates a new flock.
func (fm *FlockManager) FlockPost(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	breed := strings.TrimSpace(r.Get("breed").String())
	hatchDateStr := strings.TrimSpace(r.Get("hatch_date").String())
	numberOfBirdsStr := strings.TrimSpace(r.Get("number_of_birds").String())
	currentAgeStr := strings.TrimSpace(r.Get("current_age").String())
	barnIDStr := strings.TrimSpace(r.Get("barn_id").String())
	healthStatus := strings.TrimSpace(r.Get("health_status").String())
	feedTypeIDStr := strings.TrimSpace(r.Get("feed_type_id").String())
	notes := strings.TrimSpace(r.Get("notes").String())

	errs := map[string]string{}
	if breed == "" {
		errs["breed"] = "Breed is required"
	}

	var hatchDate *time.Time
	if hatchDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", hatchDateStr); err == nil {
			hatchDate = &parsedDate
		} else {
			errs["hatch_date"] = "Hatch date must be a valid date (YYYY-MM-DD)"
		}
	}

	var numberOfBirds *int
	if numberOfBirdsStr != "" {
		if numVal, err := strconv.Atoi(numberOfBirdsStr); err == nil {
			numberOfBirds = &numVal
		} else {
			errs["number_of_birds"] = "Number of birds must be a valid number"
		}
	}

	var currentAge *int
	if currentAgeStr != "" {
		if ageVal, err := strconv.Atoi(currentAgeStr); err == nil {
			currentAge = &ageVal
		} else {
			errs["current_age"] = "Current age must be a valid number"
		}
	}

	var barnID *int64
	if barnIDStr != "" {
		if idVal, err := strconv.ParseInt(barnIDStr, 10, 64); err == nil {
			barnID = &idVal
		} else {
			errs["barn_id"] = "Barn ID must be a valid number"
		}
	}

	var healthStat *string
	if healthStatus != "" {
		healthStat = &healthStatus
	}

	var feedTypeID *int64
	if feedTypeIDStr != "" {
		if idVal, err := strconv.ParseInt(feedTypeIDStr, 10, 64); err == nil {
			feedTypeID = &idVal
		} else {
			errs["feed_type_id"] = "Feed type ID must be a valid number"
		}
	}

	var notesPtr *string
	if notes != "" {
		notesPtr = &notes
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		flock := &domain.Flock{
			Breed:         breed,
			HatchDate:     hatchDate,
			NumberOfBirds: numberOfBirds,
			CurrentAge:    currentAge,
			BarnID:        barnID,
			HealthStatus:  healthStat,
			FeedTypeID:    feedTypeID,
			Notes:         notesPtr,
			Audit: domain.AuditFields{
				CreatedBy: &userIDStr,
				UpdatedBy: &userIDStr,
			},
		}

		_, err := fm.FlockRepo.Create(r.GetCtx(), flock)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "create flock: %v", err)
			errs["form"] = "Failed to create flock"
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
		r.Response.RedirectTo(middleware.BasePath() + "/management/flocks")
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/flocks")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/flocks")
}

// FlockGet renders a specific flock for editing or a new flock form.
func (fm *FlockManager) FlockGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	var flock *domain.Flock

	// Check if this is a request for a new flock (no ID provided)
	if idStr == "" || idStr == "new" {
		// This is a request for creating a new flock
		flock = nil
	} else {
		// This is a request for editing an existing flock
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			r.Response.WriteStatusExit(400, "Invalid flock ID")
			return
		}

		flock, err = fm.FlockRepo.FindByID(r.GetCtx(), id)
		if err != nil {
			if err == data.ErrNotFound {
				r.Response.WriteStatusExit(404, "Flock not found")
				return
			}
			g.Log().Errorf(r.GetCtx(), "find flock: %v", err)
			r.Response.WriteStatusExit(500, "Internal server error")
			return
		}
	}

	barns, err := fm.BarnRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list barns: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	feedTypes, err := fm.FeedTypeRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list feed types: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.FlockContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				flock,
				barns,
				feedTypes,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.FlockPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			flock,
			barns,
			feedTypes,
		),
	)
}

// FlockPut updates an existing flock.
func (fm *FlockManager) FlockPut(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid flock ID")
		return
	}

	breed := strings.TrimSpace(r.Get("breed").String())
	hatchDateStr := strings.TrimSpace(r.Get("hatch_date").String())
	numberOfBirdsStr := strings.TrimSpace(r.Get("number_of_birds").String())
	currentAgeStr := strings.TrimSpace(r.Get("current_age").String())
	barnIDStr := strings.TrimSpace(r.Get("barn_id").String())
	healthStatus := strings.TrimSpace(r.Get("health_status").String())
	feedTypeIDStr := strings.TrimSpace(r.Get("feed_type_id").String())
	notes := strings.TrimSpace(r.Get("notes").String())

	errs := map[string]string{}
	if breed == "" {
		errs["breed"] = "Breed is required"
	}

	var hatchDate *time.Time
	if hatchDateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", hatchDateStr); err == nil {
			hatchDate = &parsedDate
		} else {
			errs["hatch_date"] = "Hatch date must be a valid date (YYYY-MM-DD)"
		}
	}

	var numberOfBirds *int
	if numberOfBirdsStr != "" {
		if numVal, err := strconv.Atoi(numberOfBirdsStr); err == nil {
			numberOfBirds = &numVal
		} else {
			errs["number_of_birds"] = "Number of birds must be a valid number"
		}
	}

	var currentAge *int
	if currentAgeStr != "" {
		if ageVal, err := strconv.Atoi(currentAgeStr); err == nil {
			currentAge = &ageVal
		} else {
			errs["current_age"] = "Current age must be a valid number"
		}
	}

	var barnID *int64
	if barnIDStr != "" {
		if idVal, err := strconv.ParseInt(barnIDStr, 10, 64); err == nil {
			barnID = &idVal
		} else {
			errs["barn_id"] = "Barn ID must be a valid number"
		}
	}

	var healthStat *string
	if healthStatus != "" {
		healthStat = &healthStatus
	}

	var feedTypeID *int64
	if feedTypeIDStr != "" {
		if idVal, err := strconv.ParseInt(feedTypeIDStr, 10, 64); err == nil {
			feedTypeID = &idVal
		} else {
			errs["feed_type_id"] = "Feed type ID must be a valid number"
		}
	}

	var notesPtr *string
	if notes != "" {
		notesPtr = &notes
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		flock := &domain.Flock{
			FlockID:       id,
			Breed:         breed,
			HatchDate:     hatchDate,
			NumberOfBirds: numberOfBirds,
			CurrentAge:    currentAge,
			BarnID:        barnID,
			HealthStatus:  healthStat,
			FeedTypeID:    feedTypeID,
			Notes:         notesPtr,
			Audit: domain.AuditFields{
				UpdatedBy: &userIDStr,
			},
		}

		err := fm.FlockRepo.Update(r.GetCtx(), flock)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "update flock: %v", err)
			errs["form"] = "Failed to update flock"
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
		r.Response.RedirectTo(fmt.Sprintf("%s/management/flocks/%d", middleware.BasePath(), id))
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", fmt.Sprintf("%s/management/flocks/%d", middleware.BasePath(), id))
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the updated flock
	r.Response.RedirectTo(fmt.Sprintf("%s/management/flocks/%d", middleware.BasePath(), id))
}

// FlockDelete soft deletes a flock.
func (fm *FlockManager) FlockDelete(r *ghttp.Request) {
	_, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid flock ID")
		return
	}

	err = fm.FlockRepo.SoftDelete(r.GetCtx(), id, time.Now())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "delete flock: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/flocks")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/flocks")
}
