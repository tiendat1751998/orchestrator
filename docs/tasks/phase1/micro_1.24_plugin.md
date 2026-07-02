# Micro-Task 1.24: Tạo contracts/plugin/plugin.go

## Thông tin
- **File tạo**: `contracts/plugin/plugin.go`
- **Package**: `plugin`
- **Dependencies trước**: 1.06 (contracts/types.go)
- **Thời gian**: 10 phút
- **Verify**: `go build ./contracts/plugin/...`

## Nội dung CHÍNH XÁC cần tạo

```go
// Package plugin defines the lifecycle interface for all pluggable components.
// Every agent, provider, tool, etc. must implement Plugin for lifecycle management.
package plugin

import "context"

// Type identifies the category of a plugin.
// Used by the registry to organize plugins by category.
type Type string

const (
	TypeAgent    Type = "agent"
	TypeProvider Type = "provider"
	TypeTool     Type = "tool"
	TypeSearch   Type = "search"
	TypeMemory   Type = "memory"
	TypeWorkflow Type = "workflow"
	TypeContext  Type = "context"
)

// String returns the string representation.
func (t Type) String() string { return string(t) }

// Plugin is the lifecycle interface for all pluggable components.
//
// Every component that registers with the kernel (agents, providers, tools)
// must implement this interface. The kernel manages the lifecycle:
//
//	Init → Start → [Health checks] → Stop
//
// Lifecycle order rules:
//   - Init is called for ALL plugins before any Start.
//   - Start is called in dependency order (EventBus → Registry → Provider → Agent).
//   - Stop is called in REVERSE dependency order (Agent → Provider → Registry → EventBus).
//
// WHY Init and Start are separate?
//   - Init: Load config, validate settings, allocate memory (NO network, NO goroutines)
//   - Start: Open connections, start background goroutines, become operational
//   - Separation allows: init all → validate all → start all in dependency order
//   - If init fails → don't start anything → clean error message
//   - If start fails → stop already-started plugins in reverse order
type Plugin interface {
	// Name returns the unique identifier for this plugin.
	// Must be unique within its Type category.
	// Convention: lowercase, alphanumeric + hyphens
	Name() string

	// Type returns the plugin category.
	Type() Type

	// Version returns the plugin version (semver format).
	Version() string

	// Init loads configuration and validates settings.
	//
	// Called once before Start. Must NOT:
	//   - Open network connections
	//   - Start goroutines
	//   - Perform I/O operations
	//
	// Parameters:
	//   - ctx: for cancellation (e.g., if init takes too long)
	//   - config: plugin-specific configuration from the main config file
	//
	// WHY config is map[string]any?
	// → Each plugin has different config fields.
	// → Plugin unmarshals into its own typed config struct.
	// → Example:
	//     type MyConfig struct {
	//         APIKey string `mapstructure:"api_key"`
	//     }
	//     mapstructure.Decode(config, &myConfig)
	Init(ctx context.Context, config map[string]any) error

	// Start makes the plugin operational.
	//
	// Called after Init. This is where you:
	//   - Open database connections
	//   - Start background goroutines
	//   - Connect to external services
	//
	// If Start fails, the kernel will call Stop for cleanup.
	Start(ctx context.Context) error

	// Stop gracefully shuts down the plugin.
	//
	// Called when the kernel is shutting down.
	// Must:
	//   - Close all connections
	//   - Stop all goroutines (use context cancellation or done channels)
	//   - Release all resources
	//
	// Must complete within the context deadline.
	// If ctx is cancelled, perform best-effort cleanup and return.
	//
	// Stop is called in REVERSE order of Start.
	// This ensures: Agent.Stop() before Provider.Stop()
	// (so agents don't try to use a stopped provider).
	Stop(ctx context.Context) error

	// Health checks if the plugin is functioning correctly.
	//
	// Called periodically by the kernel supervisor.
	// Return nil if healthy, or an error describing the problem.
	//
	// The kernel uses Health results to:
	//   - Update circuit breaker state
	//   - Log warnings
	//   - Trigger alerts
	//
	// Must complete quickly (< 5 seconds).
	// Must NOT perform heavy operations.
	Health(ctx context.Context) error
}
```

## ⚠️ Pitfalls cần tránh
1. **Init vs Start TÁCH RIÊNG**: Init = pure config (no I/O). Start = operational (I/O). Tách ra để kernel có thể init tất cả trước, validate, rồi start theo dependency order.
2. **Stop REVERSE order**: Agent stop TRƯỚC Provider. Nếu Provider stop trước → Agent gọi Provider đã chết → nil pointer / panic.
3. **Health returns error**: nil = healthy. non-nil = unhealthy + error message. Circuit breaker dùng kết quả này.
4. **Config là `map[string]any`**: Plugin tự unmarshal vào typed struct. KHÔNG dùng global config struct vì mỗi plugin khác nhau.
5. **Stop must be idempotent**: Gọi Stop() 2 lần KHÔNG nên panic. Check if already stopped.

## Checklist
- [ ] File `contracts/plugin/plugin.go` tồn tại
- [ ] Package: `package plugin`
- [ ] Type type với 7 constants
- [ ] Plugin interface với 7 methods (Name, Type, Version, Init, Start, Stop, Health)
- [ ] Init nhận `map[string]any` config
- [ ] Tất cả lifecycle methods nhận `context.Context`
- [ ] Godoc comments chi tiết
- [ ] Lifecycle order documented trong comments
- [ ] `go build ./contracts/plugin/...` không lỗi
