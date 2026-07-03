# Micro-Task 6.09: Create kernel/gateway/gateway.go

## Info
- **File**: `kernel/gateway/gateway.go`
- **Package**: `gateway`
- **Depends on**: Phase 5 (orchestrator), 6.10-6.13
- **Time**: 25 min
- **Verify**: `go build ./kernel/gateway/...`

## External dependencies to add
```bash
go get github.com/go-chi/chi/v5@latest
go get github.com/go-chi/cors@latest
```

## Purpose
Implements the HTTP gateway server (`Gateway` struct with `Start`/`Stop`) that wires all REST, SSE, and WebSocket route groups. Uses `chi` router for lightweight, stdlib-compatible routing.

## EXACT code to create

```go
// Package gateway implements the HTTP REST/WebSocket API server.
package gateway

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/tiendat1751998/orchestrator/kernel/orchestrator"
)

// Gateway serves the REST API and WebSocket connections.
// Thread-safe.
type Gateway struct {
	mu     sync.Mutex
	server *http.Server
	orch   *orchestrator.Orchestrator
	logger *slog.Logger
}

// NewGateway constructs a new Gateway.
func NewGateway(orch *orchestrator.Orchestrator, logger *slog.Logger) *Gateway {
	return &Gateway{
		orch:   orch,
		logger: logger,
	}
}

// Start binds to the given address and begins serving HTTP requests.
// Blocks until the server is shut down.
func (g *Gateway) Start(addr string) error {
	g.mu.Lock()
	if g.server != nil {
		g.mu.Unlock()
		return errors.New("gateway: server is already running")
	}

	router := g.buildRouter()

	g.server = &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // Disabled for SSE streaming
		IdleTimeout:  60 * time.Second,
		BaseContext: func(l net.Listener) context.Context {
			return context.Background()
		},
	}
	g.mu.Unlock()

	g.logger.Info("gateway starting", "addr", addr)

	if err := g.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("gateway: listen failed: %w", err)
	}
	return nil
}

// Stop performs a graceful shutdown with a 10-second deadline.
func (g *Gateway) Stop(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.server == nil {
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	g.logger.Info("gateway shutting down")

	err := g.server.Shutdown(shutdownCtx)
	g.server = nil
	return err
}

func (g *Gateway) buildRouter() chi.Router {
	r := chi.NewRouter()

	// Global middleware stack
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(requestLogger(g.logger))

	// API v1 route group
	r.Route("/api/v1", func(r chi.Router) {
		// Missions
		r.Post("/missions", g.createMission)
		r.Get("/missions", g.listMissions)
		r.Get("/missions/{id}", g.getMission)
		r.Delete("/missions/{id}", g.cancelMission)
		r.Get("/missions/{id}/stream", g.streamMission)

		// Registry queries
		r.Get("/agents", g.listAgents)
		r.Get("/providers", g.listProviders)

		// System
		r.Get("/health", g.healthCheck)
	})

	return r
}

// requestLogger is a chi middleware that logs request duration and status.
func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			logger.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration", time.Since(start).String(),
				"bytes", ww.BytesWritten(),
			)
		})
	}
}

// healthCheck returns system health status.
func (g *Gateway) healthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
```

## Rules
1. **WriteTimeout = 0**: SSE streaming requires disabling write timeout. A non-zero write timeout will kill SSE connections mid-stream.
2. **Recoverer Middleware**: Always include `chi/middleware.Recoverer` to convert panics in handlers to 500 responses instead of crashing the server.
3. **Graceful Shutdown**: Use `server.Shutdown()` with a bounded context (10s). Never `server.Close()` which drops active connections immediately.

## Pitfalls

### Pitfall 1: Using `http.ListenAndServe` without timeout configuration
```go
// WRONG:
http.ListenAndServe(addr, router) // No read/idle timeouts → slowloris DoS vulnerability

// CORRECT:
server := &http.Server{
    ReadTimeout: 15 * time.Second,
    IdleTimeout: 60 * time.Second,
}
```

### Pitfall 2: Blocking Start() preventing shutdown
`ListenAndServe` blocks. Run it in a goroutine in production, or use `Start` as a blocking call and `Stop` from another goroutine/signal handler.

## Verify
```bash
go build ./kernel/gateway/...
```

## Checklist
- [ ] File `kernel/gateway/gateway.go` exists
- [ ] Package: `gateway`
- [ ] Chi router with versioned API routes
- [ ] Request logging middleware
- [ ] Panic recovery middleware
- [ ] Graceful shutdown with timeout
- [ ] Health check endpoint
- [ ] `go build ./kernel/gateway/...` passes
