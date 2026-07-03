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
