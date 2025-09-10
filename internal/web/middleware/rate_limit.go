package middleware

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
)

const (
	defaultRPS   = 10
	defaultBurst = 20
)

type tokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	lastRefill time.Time
	rps        float64
	burst      float64
}

func (b *tokenBucket) allow(now time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill
	elapsed := now.Sub(b.lastRefill).Seconds()
	if elapsed > 0 {
		b.tokens += elapsed * b.rps
		if b.tokens > b.burst {
			b.tokens = b.burst
		}
		b.lastRefill = now
	}

	// Consume
	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true
	}
	return false
}

type limiter struct {
	rps   float64
	burst float64

	mu      sync.Mutex
	buckets map[string]*tokenBucket
}

func newLimiter(rps, burst int) *limiter {
	return &limiter{
		rps:     float64(rps),
		burst:   float64(burst),
		buckets: make(map[string]*tokenBucket),
	}
}

func (l *limiter) get(ip string) *tokenBucket {
	l.mu.Lock()
	defer l.mu.Unlock()
	if b, ok := l.buckets[ip]; ok {
		return b
	}
	b := &tokenBucket{
		tokens:     float64(l.burst),
		lastRefill: time.Now(),
		rps:        l.rps,
		burst:      l.burst,
	}
	l.buckets[ip] = b
	return b
}

func envInt(name string, def int) int {
	if v := os.Getenv(name); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

// RateLimit is a token-bucket limiter per client IP.
// Defaults can be configured via env:
// - RATE_LIMIT_RPS (default 10)
// - RATE_LIMIT_BURST (default 20)
func RateLimit() ghttp.HandlerFunc {
	rps := envInt("RATE_LIMIT_RPS", defaultRPS)
	burst := envInt("RATE_LIMIT_BURST", defaultBurst)
	l := newLimiter(rps, burst)

	return func(r *ghttp.Request) {
		ip := r.GetClientIp()
		if ip == "" {
			ip = "unknown"
		}
		if !l.get(ip).allow(time.Now()) {
			r.Response.WriteStatus(http.StatusTooManyRequests)
			r.Response.Header().Set("Retry-After", "1")
			r.Response.Header().Set("Cache-Control", "no-store")
			r.Response.Write("Too Many Requests")
			return
		}
		r.Middleware.Next()
	}
}
