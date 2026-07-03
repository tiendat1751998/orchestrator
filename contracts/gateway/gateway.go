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
