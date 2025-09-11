package handlers

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/a-h/templ"
	"github.com/cr1cr1/farm-manager/internal/data"
	"github.com/cr1cr1/farm-manager/internal/domain"
	"github.com/cr1cr1/farm-manager/internal/web/middleware"
	"github.com/cr1cr1/farm-manager/internal/web/templates/pages"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	Repo data.UserRepo
}

// RegisterAuthRoutes wires auth endpoints under /app.
func RegisterAuthRoutes(group *ghttp.RouterGroup, repo data.UserRepo) {
	h := &Auth{Repo: repo}
	group.GET("/login", h.LoginGet)
	group.POST("/login", h.LoginPost)
	group.POST("/logout", h.LogoutPost)
}

// LoginGet renders the login page. If already authenticated, redirects to /app.
func (h *Auth) LoginGet(r *ghttp.Request) {
	if u, ok := middleware.CurrentUsername(r); ok && u != "" {
		r.Response.RedirectTo(middleware.BasePath())
		return
	}

	// Seed default admin once if no users exist.
	if err := h.ensureSeedAdmin(r); err != nil {
		g.Log().Errorf(r.GetCtx(), "seed admin failed: %v", err)
	}

	csrf := middleware.HiddenCsrfFieldValue(r)
	title := "Login"
	username := ""
	errs := map[string]string{}

	component := pages.LoginPage(middleware.BasePath(), title, csrf, errs, username)
	r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = component.Render(r.GetCtx(), r.Response.Writer)
}

// LoginPost attempts to authenticate and set session. On failure, re-render with errors.
func (h *Auth) LoginPost(r *ghttp.Request) {
	// CSRF is validated by middleware.Csrf()
	username := strings.TrimSpace(r.Get("username").String())
	password := r.Get("password").String()

	errs := map[string]string{}
	if username == "" {
		errs["username"] = "Username is required"
	}
	if password == "" {
		errs["password"] = "Password is required"
	}

	if len(errs) == 0 {
		u, err := h.Repo.FindByUsername(r.GetCtx(), username)
		if err != nil {
			if err == data.ErrNotFound {
				errs["form"] = "Invalid username or password"
			} else {
				g.Log().Errorf(r.GetCtx(), "find user: %v", err)
				errs["form"] = "Authentication failed"
			}
		} else {
			if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil {
				// Success
				middleware.SetLoggedIn(r, u.Username)
				middleware.SetNoCache(r)

				// Check if this is a DataStar request
				isDataStarRequest := r.Header.Get("datastar-request") == "true"
				if isDataStarRequest {
					// For DataStar requests, return JavaScript to redirect
					basePath := middleware.BasePath()
					js := fmt.Sprintf("window.location.href = %q;", basePath)
					r.Response.Header().Set("Content-Type", "text/javascript")
					r.Response.WriteHeader(200)
					r.Response.Writer.Write([]byte(js))
					return
				} else {
					// For regular requests, redirect
					r.Response.RedirectTo(middleware.BasePath())
					return
				}
			}
			errs["form"] = "Invalid username or password"
		}
	}

	csrf := middleware.HiddenCsrfFieldValue(r)
	title := "Login"

	// Check if this is a DataStar request (has datastar header)
	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	var component templ.Component
	if isDataStarRequest && len(errs) > 0 {
		// Return fragment for DataStar requests with errors
		component = pages.LoginFragment(middleware.BasePath(), csrf, errs, username)
	} else {
		// Return full page for initial load or non-DataStar requests
		component = pages.LoginPage(middleware.BasePath(), title, csrf, errs, username)
	}

	r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = component.Render(r.GetCtx(), r.Response.Writer)
}

// LogoutPost logs the user out and redirects to /app/login.
func (h *Auth) LogoutPost(r *ghttp.Request) {
	middleware.ClearLogin(r)
	middleware.SetNoCache(r)
	r.Response.RedirectTo(middleware.BasePath() + "/login")
}

// ensureSeedAdmin creates a default admin user if users table is empty.
// Default username: admin; password must be provided via ADMIN_PASSWORD (no default).
func (h *Auth) ensureSeedAdmin(r *ghttp.Request) error {
	n, err := h.Repo.Count(r.GetCtx())
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	pass := strings.TrimSpace(os.Getenv("ADMIN_PASSWORD"))
	if pass == "" {
		return fmt.Errorf("ADMIN_PASSWORD is required for first-run admin seeding")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u := &domain.User{
		Username:            "admin",
		PasswordHash:        string(hash),
		ForcePasswordChange: false, // optional flow not implemented here
	}
	_, err = h.Repo.Create(r.GetCtx(), u)
	if err != nil {
		// Ignore unique errors if seeded concurrently.
		if err == sql.ErrNoRows {
			return nil
		}
		g.Log().Noticef(r.GetCtx(), "admin seed may already exist: %v", err)
	}
	return nil
}
