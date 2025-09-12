package middleware

import (
	"os"

	"github.com/cr1cr1/farm-manager/internal/domain"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

const SessionUserKey = "auth:user"

// BasePath returns the application base path, defaulting to /app.
func BasePath() string {
	if v := os.Getenv("APP_BASE_PATH"); v != "" {
		return v
	}
	return "/app"
}

// CurrentUser returns the logged-in user from session if present.
func CurrentUser(r *ghttp.Request) (*domain.User, bool) {
	v := r.Session.MustGet(SessionUserKey)
	if v == nil || v.IsNil() {
		return nil, false
	}

	var user domain.User
	if err := v.Scan(&user); err != nil {
		return nil, false
	}

	return &user, true
}

// SetLoggedIn marks the request session as logged-in for the given user.
func SetLoggedIn(r *ghttp.Request, user *domain.User) {
	// TODO encrypt session data when storing on disk
	err := r.Session.Set(SessionUserKey, user)
	if err != nil {
		// log error but do not block login
		g.Log().Errorf(r.GetCtx(), "set session user: %v", err)
	}
}

// ClearLogin removes login information from the session.
func ClearLogin(r *ghttp.Request) {
	r.Session.Remove(SessionUserKey)
}

// RequireAuth redirects to /app/login (or APP_BASE_PATH/login) when not authenticated.
func RequireAuth() ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		if _, ok := CurrentUser(r); !ok {
			// Ensure no caching on redirect responses.
			r.Response.Header().Set("Cache-Control", "no-store")
			r.Response.RedirectTo(BasePath() + "/login")
			return
		}
		r.Middleware.Next()
	}
}
