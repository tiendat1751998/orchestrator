# Micro-Task 1.33: Tạo contracts/gateway/gateway.go

## Thông tin
- **File tạo**: `contracts/gateway/gateway.go`
- **Package**: `gateway`
- **Dependencies trước**: 1.06
- **Thời gian**: 5 phút
- **Verify**: `go build ./contracts/gateway/...`

## Nội dung CHÍNH XÁC cần tạo

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

## ⚠️ Pitfalls cần tránh
1. **Start is blocking but uses goroutine**: Start() blocks waiting for ctx.Done(). The actual HTTP server runs in a goroutine inside Start().
2. **Graceful shutdown**: Stop() should wait for in-flight requests to complete (with a deadline).
3. **Address() before Start()**: Returns "" if not started yet. Callers must check.

## Checklist
- [ ] File `contracts/gateway/gateway.go` tồn tại
- [ ] Package: `package gateway`
- [ ] Gateway interface với 3 methods (Start, Stop, Address)
- [ ] Godoc comments
- [ ] `go build ./contracts/gateway/...` không lỗi
