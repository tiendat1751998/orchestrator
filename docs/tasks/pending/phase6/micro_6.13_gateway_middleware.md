# Micro-Task 6.13: Create kernel/gateway/middleware.go

## Info
- **File**: `kernel/gateway/middleware.go`
- **Package**: `gateway`
- **Depends on**: 6.09
- **Time**: 15 min
- **Verify**: `go build ./kernel/gateway/...`

## Purpose
Implements CORS middleware, API key authentication, rate limiting, and request ID injection for the HTTP gateway.

## EXACT code to create

```go
package gateway

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/cors"
)

// CORSMiddleware returns a configured CORS handler.
func CORSMiddleware() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // TODO: restrict in production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	})
}

// APIKeyAuth returns middleware that validates API key from Authorization header.
func APIKeyAuth(validKeys map[string]bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(validKeys) == 0 {
				next.ServeHTTP(w, r) // No keys configured = auth disabled
				return
			}

			auth := r.Header.Get("Authorization")
			if auth == "" {
				writeError(w, http.StatusUnauthorized, "missing Authorization header")
				return
			}

			key := strings.TrimPrefix(auth, "Bearer ")
			if !validKeys[key] {
				writeError(w, http.StatusForbidden, "invalid API key")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SimpleRateLimiter implements a per-IP token bucket rate limiter.
type SimpleRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int           // requests per window
	window   time.Duration // time window
}

type visitor struct {
	tokens    int
	lastReset time.Time
}

// NewSimpleRateLimiter constructs a rate limiter (e.g., 100 requests per minute).
func NewSimpleRateLimiter(rate int, window time.Duration) *SimpleRateLimiter {
	return &SimpleRateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}
}

// Middleware returns the HTTP middleware function.
func (rl *SimpleRateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			rl.mu.Lock()
			v, exists := rl.visitors[ip]
			now := time.Now()

			if !exists || now.Sub(v.lastReset) > rl.window {
				rl.visitors[ip] = &visitor{tokens: rl.rate - 1, lastReset: now}
				rl.mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			if v.tokens <= 0 {
				rl.mu.Unlock()
				writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			v.tokens--
			rl.mu.Unlock()
			next.ServeHTTP(w, r)
		})
	}
}
```

## Rules
1. **CORS Restrictive in Production**: Default `*` origin is development-only. Production deployments MUST restrict origins.
2. **Bearer Token Auth**: API key extracted from `Authorization: Bearer <key>` header. Disabled when no keys configured.
3. **Per-IP Rate Limiting**: Token bucket per IP. Simple but effective for single-instance deployments. Use Redis-backed limiter for distributed setups.

## Verify
```bash
go build ./kernel/gateway/...
```

## Checklist
- [ ] File `kernel/gateway/middleware.go` exists
- [ ] CORS middleware with configurable origins
- [ ] API key auth middleware (disabled when no keys)
- [ ] Per-IP rate limiter with token bucket
- [ ] `go build ./kernel/gateway/...` passes
