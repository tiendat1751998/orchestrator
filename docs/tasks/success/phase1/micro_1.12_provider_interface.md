# Micro-Task 1.12: Create contracts/provider/provider.go

## Info
- **File**: `contracts/provider/provider.go`
- **Package**: `provider`
- **Depends on**: 1.09, 1.10, 1.11
- **Time**: 10 min
- **Verify**: `go build ./contracts/...`

## Purpose
Declares the core `Provider` contract interface that all backend AI adapters must satisfy.

## EXACT code to create

```go
package provider

import "context"

// Provider is the core interface that all AI providers must implement.
//
// Lifecycle:
//   Provider lifecycle (Init, Start, Stop) is managed by the Plugin interface
//   in contracts/plugin. This interface only defines runtime behavior.
type Provider interface {
	// Name returns the unique identifier for this provider.
	Name() string

	// Send sends a request and waits for the complete response.
	Send(ctx context.Context, req *Request) (*Response, error)

	// Stream sends a request and returns a channel for streaming the response.
	Stream(ctx context.Context, req *Request) (<-chan StreamChunk, error)

	// IsAvailable checks if the provider is ready to accept requests.
	IsAvailable(ctx context.Context) bool

	// Models returns the list of models supported by this provider.
	Models(ctx context.Context) ([]string, error)
}
```

## Pitfalls

### Pitfall 1: Mixing plugin lifecycle states with runtime interfaces
Adding `Init`, `Start`, or `Stop` hooks here violates separation of concerns. Keep initialization bounds restricted to the `Plugin` wrapper contracts.

### Pitfall 2: Permitting writers to access read-only stream channels
Always restrict stream channels returned by providers to read-only formats (`<-chan`) to prevent consumer goroutines from writing to channels and causing crashes.

## Verify
```bash
go build ./contracts/...
```

## Checklist
- [ ] File exists at `contracts/provider/provider.go`
- [ ] Package name is `provider`
- [ ] Interface defines Send, Stream, IsAvailable, and Models
- [ ] Stream returns a read-only channel format
- [ ] Build command passes
