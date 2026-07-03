# Micro-Task 1.33: Create contracts/gateway/gateway.go

## Info
- **File**: `contracts/gateway/gateway.go`
- **Package**: `gateway`
- **Depends on**: 1.06
- **Time**: 5 min
- **Verify**: `go build ./contracts/gateway/...`

## Purpose
Declares the unified `Gateway` interface representing external-facing network endpoints (such as REST API, gRPC, or WebSocket servers) that expose orchestrator capabilities.

## EXACT code to create

```go
// Package gateway defines the contract for external-facing servers.
// The gateway exposes the orchestrator's functionality via REST API, gRPC, or WebSocket.
package gateway

import "context"

// Gateway is the unified entry point for external requests.
//
// Implementations:
//   - REST API server (Phase 6)
//   - gRPC server (future)
//   - WebSocket server (future)
//   - Message queue consumer (future)
type Gateway interface {
	// Start begins listening for incoming requests.
	// This is a blocking call — it runs until ctx is cancelled or Stop is called.
	//
	// Typical implementation:
	//   server := &http.Server{Addr: addr, Handler: router}
	//   go server.ListenAndServe()
	//   <-ctx.Done()
	//   server.Shutdown(shutdownCtx)
	Start(ctx context.Context) error

	// Stop gracefully shuts down the gateway.
	// Must complete within the context deadline.
	// Ongoing requests should be allowed to finish.
	Stop(ctx context.Context) error

	// Address returns the listen address (e.g., ":8080", "0.0.0.0:9090").
	// Returns empty string if the gateway hasn't started yet.
	Address() string
}
```

## Rules
1. **Blocking Start**: The `Start` execution blocks until context is cancelled or `Stop` is called. The underlying web listener runs in a separate goroutine spawned within `Start`.
2. **Graceful Shutdown**: `Stop` must use graceful shutdown methods to allow ongoing HTTP requests to complete before closing connections.
3. **Empty Address State**: `Address` must return `""` if the listener is not yet bound or started.

## ⚠️ Pitfalls

### Pitfall 1: Non-blocking implementation of `Start()`
If `Start()` spawns a listener and returns immediately without blocking on `ctx.Done()`, the calling bootstrap kernel will assume startup is complete and continue, which can cause race conditions or premature process exit.

### Pitfall 2: Not setting a timeout during `Stop` execution
If a client request hangs or remains open indefinitely during gateway shutdown, and `Stop` does not enforce a timeout on the graceful shutdown context, the shutdown process will block indefinitely. Always wrap the shutdown context with a deadline.

## Verify
```bash
go build ./contracts/gateway/...
```

## Checklist
- [ ] File `contracts/gateway/gateway.go` exists
- [ ] Package: `gateway`
- [ ] `Gateway` interface declares `Start`, `Stop`, and `Address` methods
- [ ] `Start` and `Stop` methods receive `context.Context` parameters
- [ ] `go build ./contracts/gateway/...` passes
