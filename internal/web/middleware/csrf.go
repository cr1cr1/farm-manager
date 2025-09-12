package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
)

const (
	defaultCSRFCookie = "csrf_token"
	defaultCSRFHeader = "X-CSRF-Token"
)

// CsrfCookieName returns the cookie name for the CSRF token.
func CsrfCookieName() string {
	if v := os.Getenv("CSRF_COOKIE_NAME"); v != "" {
		return v
	}
	return defaultCSRFCookie
}

// CsrfHeaderName returns the header name used to validate the CSRF token.
func CsrfHeaderName() string {
	if v := os.Getenv("CSRF_HEADER_NAME"); v != "" {
		return v
	}
	return defaultCSRFHeader
}

// CsrfToken returns (and ensures issuing) the CSRF token for the current session/request.
func CsrfToken(r *ghttp.Request) string {
	name := CsrfCookieName()
	v := r.Cookie.Get(name).String()
	if v != "" {
		return v
	}
	token := newToken()
	// Set cookie with conservative defaults.
	// GoFrame sets Path=/ by default; enforce no-cache for safety via headers.
	r.Cookie.Set(name, token)
	// Best-effort cookie flags via raw header (for stricter policies).
	c := &http.Cookie{
		Name:     name,
		Value:    token,
		Path:     "/",
		HttpOnly: true, // token not needed by JS; we render it server-side into forms
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil, // secure when TLS present
		MaxAge:   0,            // session cookie
	}
	http.SetCookie(r.Response.Writer, c)
	return token
}

// Csrf is middleware that issues CSRF token cookies on safe methods and validates tokens on mutating methods.
// Validation passes when either:
// - Header CsrfHeaderName equals the CSRF cookie value; OR
// - Form field named CsrfCookieName equals the CSRF cookie value.
func Csrf() ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		method := strings.ToUpper(r.Method)
		switch method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
			// Ensure token cookie exists for subsequent form rendering.
			_ = CsrfToken(r)
			r.Middleware.Next()
			return
		default:
			cookie := r.Cookie.Get(CsrfCookieName()).String()
			if cookie == "" {
				forbid(r)
				return
			}
			header := r.Header.Get(CsrfHeaderName())
			form := r.Get(CsrfCookieName()).String() // posted form value if present
			if header == cookie || form == cookie {
				r.Middleware.Next()
				return
			}
			forbid(r)
			return
		}
	}
}

func forbid(r *ghttp.Request) {
	r.Response.WriteStatus(http.StatusForbidden)
	r.Response.Header().Set("Cache-Control", "no-store")
	r.Response.Write("Forbidden: invalid CSRF token")
}

// newToken generates a URL-safe random token.
func newToken() string {
	var b [32]byte
	_, _ = rand.Read(b[:])
	return base64.RawURLEncoding.EncodeToString(b[:])
}

// SetNoCache adds no-store cache headers; useful for auth flows.
func SetNoCache(r *ghttp.Request) {
	r.Response.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	r.Response.Header().Set("Pragma", "no-cache")
	r.Response.Header().Set("Expires", time.Unix(0, 0).UTC().Format(http.TimeFormat))
}
