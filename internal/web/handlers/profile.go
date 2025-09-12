package handlers

import (
	"strings"

	"github.com/cr1cr1/farm-manager/internal/data"
	"github.com/cr1cr1/farm-manager/internal/web/middleware"
	"github.com/cr1cr1/farm-manager/internal/web/templates/pages"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"golang.org/x/crypto/bcrypt"
)

type Profile struct {
	Repo data.UserRepo
}

// RegisterProfileRoutes wires profile endpoints under /app.
func RegisterProfileRoutes(group *ghttp.RouterGroup, repo data.UserRepo) {
	p := &Profile{Repo: repo}
	group.GET("/profile", p.ProfileGet)
	group.POST("/profile/password", p.PasswordPost)
	group.POST("/profile/theme", p.ThemePost)
}

// ProfileGet renders the profile page.
func (p *Profile) ProfileGet(r *ghttp.Request) {
	username, _ := middleware.CurrentUsername(r)
	csrf := middleware.CsrfToken(r)
	errs := map[string]string{}
	success := ""

	// Load user to get theme preference
	var userTheme int
	if username != "" {
		if u, err := p.Repo.FindByUsername(r.GetCtx(), username); err == nil {
			userTheme = u.Theme
		}
	}

	_ = middleware.TemplRender(
		r,
		pages.ProfilePage(
			middleware.BasePath(),
			"Profile",
			csrf,
			errs,
			username,
			success,
			userTheme,
		),
	)
}

// PasswordPost updates the current user's password after validation.
func (p *Profile) PasswordPost(r *ghttp.Request) {
	// CSRF is validated by middleware.Csrf()
	username, ok := middleware.CurrentUsername(r)
	if !ok || username == "" {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	current := r.Get("current_password").String()
	newpw := r.Get("new_password").String()
	confirm := r.Get("confirm_password").String()

	errs := map[string]string{}
	if strings.TrimSpace(current) == "" {
		errs["current_password"] = "Current password is required"
	}
	if len(newpw) < 8 {
		errs["new_password"] = "New password must be at least 8 characters"
	}
	if newpw != confirm {
		errs["confirm_password"] = "Passwords do not match"
	}

	success := ""
	if len(errs) == 0 {
		u, err := p.Repo.FindByUsername(r.GetCtx(), username)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "find user: %v", err)
			errs["form"] = "Unable to update password"
		} else {
			if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(current)) != nil {
				errs["current_password"] = "Current password is incorrect"
			} else {
				hash, err := bcrypt.GenerateFromPassword([]byte(newpw), bcrypt.DefaultCost)
				if err != nil {
					g.Log().Errorf(r.GetCtx(), "hash password: %v", err)
					errs["form"] = "Unable to update password"
				} else {
					if err := p.Repo.UpdatePassword(r.GetCtx(), u.ID, string(hash), false); err != nil {
						g.Log().Errorf(r.GetCtx(), "update password: %v", err)
						errs["form"] = "Unable to update password"
					} else {
						middleware.SetNoCache(r)
						success = "Password updated"
					}
				}
			}
		}
	}

	_ = middleware.TemplRender(
		r,
		pages.ProfilePasswordFragment(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			errs,
			success,
		),
	)
}

// ThemePost updates the current user's theme preference.
func (p *Profile) ThemePost(r *ghttp.Request) {
	// CSRF is validated by middleware.Csrf()
	username, ok := middleware.CurrentUsername(r)
	if !ok || username == "" {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	themeStr := r.Get("theme").String()
	var theme int
	switch themeStr {
	case "dark":
		theme = 1
	case "light":
		theme = 2
	default:
		theme = 0 // default to system
	}

	// Get user to update theme
	u, err := p.Repo.FindByUsername(r.GetCtx(), username)
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "find user: %v", err)
		r.Response.WriteHeader(500)
		return
	}

	// Update theme in database
	if err := p.Repo.UpdateTheme(r.GetCtx(), u.ID, theme); err != nil {
		g.Log().Errorf(r.GetCtx(), "update theme: %v", err)
		r.Response.WriteHeader(500)
		return
	}

	// Return success response for DataStar - HTML fragment for status display
	r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.Response.WriteHeader(200)
	r.Response.Write([]byte(`<div id="theme-status" class="alert-success">Theme preference saved successfully.</div>`))
}
