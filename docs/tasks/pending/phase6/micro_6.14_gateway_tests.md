# Micro-Task 6.14: Create kernel/gateway/gateway_test.go

## Info
- **File**: `kernel/gateway/gateway_test.go`
- **Package**: `gateway_test`
- **Depends on**: 6.09-6.13
- **Time**: 20 min
- **Verify**: `go test ./kernel/gateway/...`

## Purpose
Integration tests for REST endpoints, SSE streaming, and health check using `httptest.Server`.

## EXACT code to create

```go
package gateway_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tiendat1751998/orchestrator/kernel/gateway"
)

func TestHealthCheck(t *testing.T) {
	g := gateway.NewGateway(nil, slog.Default())
	router := g.TestRouter() // exposed for testing

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatal("expected data in response")
	}
	if data["status"] != "healthy" {
		t.Errorf("expected healthy status, got %v", data["status"])
	}
}

func TestCreateMissionValidation(t *testing.T) {
	g := gateway.NewGateway(nil, slog.Default())
	router := g.TestRouter()

	// Empty body
	req := httptest.NewRequest("POST", "/api/v1/missions", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty title, got %d", w.Code)
	}
}

func TestListMissionsEmpty(t *testing.T) {
	g := gateway.NewGateway(nil, slog.Default())
	router := g.TestRouter()

	req := httptest.NewRequest("GET", "/api/v1/missions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRateLimiter(t *testing.T) {
	rl := gateway.NewSimpleRateLimiter(2, time.Minute)
	handler := rl.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, w.Code)
		}
	}

	// Third request should be rate limited
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
}
```

## Verify
```bash
go test ./kernel/gateway/... -v
```

## Checklist
- [ ] Health check returns 200 + healthy status
- [ ] Mission creation validates required fields
- [ ] Rate limiter blocks excess requests
- [ ] All tests pass
