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

const (
	THEME_SYSTEM = iota
	THEME_DARK
	THEME_LIGHT
)

// ThemeToString converts theme int to string for HTML attribute use.
func ThemeToString(theme int) string {
	switch theme {
	case THEME_DARK:
		return "dark"
	case THEME_LIGHT:
		return "light"
	default:
		return "system"
	}
}

// ParseTheme converts string input to theme int, defaulting to system.
func ParseTheme(themeStr string) int {
	switch strings.ToLower(themeStr) {
	case "dark":
		return THEME_DARK
	case "light":
		return THEME_LIGHT
	default:
		return THEME_SYSTEM
	}
}

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
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	csrf := middleware.CsrfToken(r)
	errs := map[string]string{}
	success := ""

	_ = middleware.TemplRender(
		r,
		pages.ProfilePage(
			middleware.BasePath(),
			"Profile",
			csrf,
			errs,
			success,
			user.Username,
			ThemeToString(user.Theme),
		),
	)
}

// PasswordPost updates the current user's password after validation.
func (p *Profile) PasswordPost(r *ghttp.Request) {
	// CSRF is validated by middleware.Csrf()
	user, ok := middleware.CurrentUser(r)
	if !ok {
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
		if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(current)) != nil {
			errs["current_password"] = "Current password is incorrect"
		} else {
			hash, err := bcrypt.GenerateFromPassword([]byte(newpw), bcrypt.DefaultCost)
			if err != nil {
				g.Log().Errorf(r.GetCtx(), "hash password: %v", err)
				errs["form"] = "Unable to update password"
			} else {
				if err := p.Repo.UpdatePassword(r.GetCtx(), user.ID, string(hash), false); err != nil {
					g.Log().Errorf(r.GetCtx(), "update password: %v", err)
					errs["form"] = "Unable to update password"
				} else {
					middleware.SetNoCache(r)
					success = "Password updated"
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
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	// Update theme in database
	if err := p.Repo.UpdateTheme(r.GetCtx(), user.ID, ParseTheme(r.Get("theme").String())); err != nil {
		g.Log().Errorf(r.GetCtx(), "update theme: %v", err)
		r.Response.WriteHeader(500)
		return
	}

	// Return success response for DataStar - HTML fragment for status display
	r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.Response.WriteHeader(200)
	r.Response.Write([]byte(`<div id="theme-status" class="alert-success">Theme preference saved successfully.</div>`))
}
