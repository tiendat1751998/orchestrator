# Micro-Task 6.10: Create kernel/gateway/rest.go

## Info
- **File**: `kernel/gateway/rest.go`
- **Package**: `gateway`
- **Depends on**: 6.09, 6.15 (mission manager)
- **Time**: 25 min
- **Verify**: `go build ./kernel/gateway/...`

## Purpose
Implements RESTful CRUD handlers for missions, agents listing, and providers listing. All responses use a consistent JSON envelope `{data, error, meta}`.

## EXACT code to create

```go
package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tiendat1751998/orchestrator/kernel/planner"
)

// apiResponse is the standard JSON response envelope.
type apiResponse struct {
	Data  any            `json:"data,omitempty"`
	Error *apiError      `json:"error,omitempty"`
	Meta  map[string]any `json:"meta,omitempty"`
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(apiResponse{
		Data: data,
		Meta: map[string]any{"timestamp": time.Now().UTC().Format(time.RFC3339)},
	})
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(apiResponse{
		Error: &apiError{Code: status, Message: msg},
	})
}

type createMissionRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Constraints []string `json:"constraints,omitempty"`
}

// createMission handles POST /api/v1/missions.
func (g *Gateway) createMission(w http.ResponseWriter, r *http.Request) {
	var req createMissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	mission := &planner.Mission{
		Title:       req.Title,
		Description: req.Description,
		Constraints: req.Constraints,
	}

	// Execute mission asynchronously
	go func() {
		ctx := r.Context()
		_, err := g.orch.ExecuteMission(ctx, mission, nil)
		if err != nil {
			g.logger.Error("mission execution failed", "title", mission.Title, "error", err)
		}
	}()

	writeJSON(w, http.StatusAccepted, map[string]any{
		"id":     mission.ID,
		"title":  mission.Title,
		"status": "accepted",
	})
}

// listMissions handles GET /api/v1/missions.
func (g *Gateway) listMissions(w http.ResponseWriter, r *http.Request) {
	// Placeholder: return empty list until mission store is wired
	writeJSON(w, http.StatusOK, []any{})
}

// getMission handles GET /api/v1/missions/{id}.
func (g *Gateway) getMission(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "mission ID is required")
		return
	}

	// Placeholder: return mock until mission store is wired
	writeJSON(w, http.StatusOK, map[string]any{
		"id":     id,
		"status": "unknown",
	})
}

// cancelMission handles DELETE /api/v1/missions/{id}.
func (g *Gateway) cancelMission(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "mission ID is required")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":     id,
		"status": "cancelled",
	})
}

// listAgents handles GET /api/v1/agents.
func (g *Gateway) listAgents(w http.ResponseWriter, r *http.Request) {
	// Placeholder: will be wired to registry
	writeJSON(w, http.StatusOK, []any{})
}

// listProviders handles GET /api/v1/providers.
func (g *Gateway) listProviders(w http.ResponseWriter, r *http.Request) {
	// Placeholder: will be wired to registry
	writeJSON(w, http.StatusOK, []any{})
}
```

## Rules
1. **Consistent Envelope**: ALL responses use `{data, error, meta}`. Never return raw objects.
2. **Async Mission Execution**: `POST /missions` returns `202 Accepted` immediately. Mission runs in a goroutine. Progress tracked via SSE stream.
3. **URL Parameters**: Use `chi.URLParam(r, "id")` — never `r.URL.Query().Get("id")` for path parameters.

## Pitfalls

### Pitfall 1: Blocking request handlers during long missions
```go
// WRONG:
func (g *Gateway) createMission(w http.ResponseWriter, r *http.Request) {
    result, _ := g.orch.ExecuteMission(ctx, mission) // Blocks for 30 min!
    writeJSON(w, 200, result)
}

// CORRECT:
go func() { g.orch.ExecuteMission(ctx, mission, nil) }()
writeJSON(w, 202, map[string]any{"status": "accepted"})
```

## Verify
```bash
go build ./kernel/gateway/...
```

## Checklist
- [ ] File `kernel/gateway/rest.go` exists
- [ ] Consistent JSON envelope on all responses
- [ ] POST /missions returns 202 Accepted
- [ ] Mission execution runs asynchronously
- [ ] `go build ./kernel/gateway/...` passes
