package middleware

import (
	"os"

	"github.com/gogf/gf/v2/net/ghttp"
)

const SessionUserKey = "auth:username"

// BasePath returns the application base path, defaulting to /app.
func BasePath() string {
	if v := os.Getenv("APP_BASE_PATH"); v != "" {
		return v
	}
	return "/app"
}

// CurrentUsername returns the logged-in username from session if present.
func CurrentUsername(r *ghttp.Request) (string, bool) {
	v := r.Session.MustGet(SessionUserKey)
	if v == nil || v.IsNil() {
		return "", false
	}
	s := v.String()
	if s == "" {
		return "", false
	}
	return s, true
}

// SetLoggedIn marks the request session as logged-in for the given username.
func SetLoggedIn(r *ghttp.Request, username string) {
	r.Session.Set(SessionUserKey, username)
}

// ClearLogin removes login information from the session.
func ClearLogin(r *ghttp.Request) {
	r.Session.Remove(SessionUserKey)
}

// RequireAuth redirects to /app/login (or APP_BASE_PATH/login) when not authenticated.
func RequireAuth() ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		if _, ok := CurrentUsername(r); !ok {
			// Ensure no caching on redirect responses.
			r.Response.Header().Set("Cache-Control", "no-store")
			r.Response.RedirectTo(BasePath() + "/login")
			return
		}
		r.Middleware.Next()
	}
}
