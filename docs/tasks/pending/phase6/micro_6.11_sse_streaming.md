# Micro-Task 6.11: Create kernel/gateway/sse.go

## Info
- **File**: `kernel/gateway/sse.go`
- **Package**: `gateway`
- **Depends on**: 6.09, 5.08 (orchestrator)
- **Time**: 20 min
- **Verify**: `go build ./kernel/gateway/...`

## Purpose
Implements Server-Sent Events (SSE) endpoint for real-time mission progress streaming via `GET /api/v1/missions/{id}/stream`.

## EXACT code to create

```go
package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// streamMission handles GET /api/v1/missions/{id}/stream (SSE).
func (g *Gateway) streamMission(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "mission ID is required")
		return
	}

	// Verify response writer supports flushing (required for SSE)
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	ctx := r.Context()

	// Heartbeat keep-alive ticker to detect dead connections
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	// Progress channel from orchestrator (placeholder until wired to real mission)
	updates := make(chan sseEvent, 10)

	// Simulate sending initial connection event
	sendSSE(w, flusher, sseEvent{
		Event: "connected",
		Data:  fmt.Sprintf(`{"mission_id":"%s","timestamp":"%s"}`, id, time.Now().UTC().Format(time.RFC3339)),
	})

	for {
		select {
		case <-ctx.Done():
			sendSSE(w, flusher, sseEvent{Event: "disconnected", Data: `{"reason":"client_closed"}`})
			return

		case update, ok := <-updates:
			if !ok {
				sendSSE(w, flusher, sseEvent{Event: "complete", Data: `{"status":"done"}`})
				return
			}
			sendSSE(w, flusher, update)

		case <-heartbeat.C:
			sendSSE(w, flusher, sseEvent{Event: "heartbeat", Data: fmt.Sprintf(`{"t":"%s"}`, time.Now().UTC().Format(time.RFC3339))})
		}
	}
}

type sseEvent struct {
	Event string
	Data  string
}

func sendSSE(w http.ResponseWriter, flusher http.Flusher, evt sseEvent) {
	if evt.Event != "" {
		fmt.Fprintf(w, "event: %s\n", evt.Event)
	}
	fmt.Fprintf(w, "data: %s\n\n", evt.Data)
	flusher.Flush()
}

// marshalSSEData serializes an object to JSON for SSE data field.
func marshalSSEData(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return `{"error":"marshal_failed"}`
	}
	return string(data)
}
```

## Rules
1. **Flusher Check**: Always verify `http.Flusher` interface before streaming. Not all `ResponseWriter` implementations support it.
2. **Heartbeat**: Send periodic heartbeats (15s) to detect dead TCP connections. Without heartbeats, half-open connections consume resources indefinitely.
3. **X-Accel-Buffering**: Set to `no` to disable nginx proxy buffering which defeats SSE streaming.

## Pitfalls

### Pitfall 1: SSE connections timing out
```go
// WRONG:
server := &http.Server{WriteTimeout: 30 * time.Second} // Kills SSE streams after 30s!

// CORRECT:
server := &http.Server{WriteTimeout: 0} // Disable write timeout for SSE
```

## Verify
```bash
go build ./kernel/gateway/...
```

## Checklist
- [ ] File `kernel/gateway/sse.go` exists
- [ ] SSE headers: `text/event-stream`, `no-cache`, `keep-alive`
- [ ] Flusher interface check before streaming
- [ ] Heartbeat keep-alive every 15 seconds
- [ ] Clean disconnect on context cancellation
- [ ] `go build ./kernel/gateway/...` passes
