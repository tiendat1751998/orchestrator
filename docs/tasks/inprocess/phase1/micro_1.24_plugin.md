# Micro-Task 1.24: Create contracts/plugin/plugin.go

## Info
- **File**: `contracts/plugin/plugin.go`
- **Package**: `plugin`
- **Depends on**: 1.06 (contracts/types.go)
- **Time**: 10 min
- **Verify**: `go build ./contracts/plugin/...`

## Purpose
Defines the `Plugin` and `Type` declarations that govern the lifecycle stages of all dynamic modules (agents, providers, tools, and storage engines) registered within the orchestrator kernel.

## EXACT code to create

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

	// Health checks if the plugin is functioning correctly and returns a detailed report.
	//
	// Parameters:
	//   - ctx: for timeout enforcement. Checkers should abort and return error on cancellation.
	//
	// Returns:
	//   - HealthReport: structured health status.
	//   - error: system level failure during the health check itself (e.g. context timeout).
	//            If the plugin is simply unhealthy (down), return (HealthReport{Status: HealthDown}, nil)
	//            rather than a non-nil error.
	Health(ctx context.Context) (HealthReport, error)
}
```

## Rules
1. **Name Format**: Plugin names must use alphanumeric characters and hyphens, all lowercase.
2. **Type Segregation**: Plugins must return their exact type constant.
3. **Lifecycle Phase Isolation**:
   - `Init`: Configuration parsing, dependency checks, structure allocations. No network requests, file writes, or background goroutines.
   - `Start`: Operational setup, launching workers, checking network resources.
   - `Stop`: Teardown of resources.
   - `Health`: Live status checking. Must return structured `HealthReport` and a Go `error` only for check execution failures.

## ⚠️ Pitfalls

### Pitfall 1: Launching goroutines or opening network sockets in `Init`
If a plugin starts background threads or dials connections in `Init`, it violates lifecycle isolation. If later plugins fail `Init`, the kernel shuts down without calling `Stop` on unstarted plugins, resulting in leaked goroutines or dangling connections. Keep `Init` purely synchronous and memory-only.

### Pitfall 2: Non-idempotent `Stop` implementations
If `Stop()` panics when called multiple times, the kernel's recovery sequences will crash during complex error cleanups. Make sure `Stop()` checks if the plugin has already shutdown before releasing channels or resources.

### Pitfall 3: Blocking indefinitely in `Stop`
If a database connection or a background worker fails to shut down, `Stop()` might hang forever, preventing the orchestrator from exiting. Always respect the passed `context.Context` deadline and force shutdown if the context expires.

## Verify
```bash
go build ./contracts/plugin/...
```

## Checklist
- [ ] File `contracts/plugin/plugin.go` exists
- [ ] Package: `plugin`
- [ ] `Type` constant enums are defined
- [ ] `Plugin` interface declared with Name, Type, Version, Init, Start, Stop, and Health methods
- [ ] `Health` returns `(HealthReport, error)`
- [ ] `Init` config parameter is a `map[string]any`
- [ ] `go build ./contracts/plugin/...` passes
