# Micro-Task 2.20: Create kernel/registry/lifecycle.go

## Info
- **File**: `kernel/registry/lifecycle.go`
- **Package**: `registry`
- **Depends on**: 2.18 (registry.go), 1.40 (health.go/plugin.go signatures)
- **Time**: 20 min
- **Verify**: `go build ./kernel/registry/...`

## Purpose
Implements the registry plugin lifecycle orchestrator (`InitAll`, `StartAll`, `StopAll`, `stopPlugins`, and `HealthCheckAll`) that coordinates configuration, start, teardown rollbacks, and health assessments.

## EXACT code to create

```go
package registry

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tiendat1751998/orchestrator/contracts/plugin"
)

// InitAll initializes all registered plugins in registration order.
//
// Init is the first lifecycle phase:
//   - Load configuration
//   - Validate settings
//   - Allocate memory
//   - NO network connections, NO goroutines
//
// If ANY plugin fails to init, the entire operation stops and returns the error.
// Already-initialized plugins are NOT rolled back (Init has no side effects to clean up).
//
// Parameters:
//   - ctx: for cancellation (e.g., init timeout)
//   - configs: maps plugin name → plugin-specific config
//              Each plugin receives only ITS config, not the entire config tree.
func (r *Registry) InitAll(ctx context.Context, configs map[string]map[string]any) error {
	plugins := r.AllPlugins() // Registration order

	for _, p := range plugins {
		// Check context cancellation before each plugin init
		select {
		case <-ctx.Done():
			return fmt.Errorf("registry: init cancelled: %w", ctx.Err())
		default:
		}

		// Get plugin-specific config (may be nil if not configured)
		pluginConfig := configs[p.Name()]

		if r.logger != nil {
			r.logger.Info("initializing plugin",
				"name", p.Name(),
				"type", p.Type(),
			)
		}

		if err := p.Init(ctx, pluginConfig); err != nil {
			return fmt.Errorf("registry: init plugin %q: %w", p.Name(), err)
		}
	}

	if r.logger != nil {
		r.logger.Info("all plugins initialized", "count", len(plugins))
	}

	return nil
}

// StartAll starts all registered plugins in registration order.
//
// Start is the second lifecycle phase:
//   - Open network connections
//   - Start background goroutines
//   - Become operational
//
// If a plugin fails to start:
//   1. Log the error
//   2. Stop all ALREADY-STARTED plugins in reverse order
//   3. Return the error
//
// Parameters:
//   - ctx: for cancellation (e.g., startup timeout)
func (r *Registry) StartAll(ctx context.Context) error {
	plugins := r.AllPlugins()
	started := make([]string, 0, len(plugins)) // Track started plugins for rollback

	for _, p := range plugins {
		select {
		case <-ctx.Done():
			// Cancel during startup → stop already-started plugins
			r.stopPlugins(context.Background(), started)
			return fmt.Errorf("registry: start cancelled: %w", ctx.Err())
		default:
		}

		if r.logger != nil {
			r.logger.Info("starting plugin",
				"name", p.Name(),
				"type", p.Type(),
			)
		}

		if err := p.Start(ctx); err != nil {
			// Rollback: stop already-started plugins
			if r.logger != nil {
				r.logger.Error("plugin start failed, rolling back",
					"failed_plugin", p.Name(),
					"error", err,
					"started_count", len(started),
				)
			}
			r.stopPlugins(context.Background(), started)
			return fmt.Errorf("registry: start plugin %q: %w", p.Name(), err)
		}

		started = append(started, p.Name())
	}

	if r.logger != nil {
		r.logger.Info("all plugins started", "count", len(plugins))
	}

	return nil
}

// StopAll stops all registered plugins in REVERSE registration order.
//
// Stop is the final lifecycle phase:
//   - Close network connections
//   - Stop background goroutines
//   - Release resources
//
// REVERSE order is CRITICAL:
//   Start order: EventBus → Registry → Provider → Agent
//   Stop order:  Agent → Provider → Registry → EventBus
//
// If a plugin fails to stop:
//   - Log the error
//   - Continue stopping remaining plugins (best-effort cleanup)
//   - Return the FIRST error encountered
//
// Parameters:
//   - ctx: for deadline (e.g., 30-second shutdown timeout)
func (r *Registry) StopAll(ctx context.Context) error {
	plugins := r.AllPluginsReversed() // REVERSE order
	return r.stopPluginInstances(ctx, plugins)
}

// stopPlugins stops specific plugins by name in REVERSE order.
// Used for rollback during failed StartAll.
func (r *Registry) stopPlugins(ctx context.Context, names []string) {
	for i := len(names) - 1; i >= 0; i-- {
		name := names[i]
		p, err := r.GetPlugin(name)
		if err != nil {
			continue // Plugin already unregistered
		}

		if r.logger != nil {
			r.logger.Info("stopping plugin (rollback)", "name", name)
		}

		if err := p.Stop(ctx); err != nil {
			if r.logger != nil {
				r.logger.Error("plugin stop failed during rollback",
					"name", name,
					"error", err,
				)
			}
		}
	}
}

// stopPluginInstances stops a list of plugin instances.
func (r *Registry) stopPluginInstances(ctx context.Context, plugins []plugin.Plugin) error {
	var firstErr error

	for _, p := range plugins {
		select {
		case <-ctx.Done():
			if r.logger != nil {
				r.logger.Warn("stop timeout reached, forcing shutdown",
					"remaining", len(plugins),
				)
			}
			if firstErr == nil {
				firstErr = fmt.Errorf("registry: stop timeout: %w", ctx.Err())
			}
			return firstErr
		default:
		}

		if r.logger != nil {
			r.logger.Info("stopping plugin",
				"name", p.Name(),
				"type", p.Type(),
			)
		}

		if err := p.Stop(ctx); err != nil {
			if r.logger != nil {
				r.logger.Error("plugin stop failed",
					"name", p.Name(),
					"error", err,
				)
			}
			if firstErr == nil {
				firstErr = fmt.Errorf("registry: stop plugin %q: %w", p.Name(), err)
			}
		}
	}

	if r.logger != nil {
		r.logger.Info("all plugins stopped")
	}

	return firstErr
}

// HealthCheckAll runs Health() on all plugins and returns their reports.
//
// Does NOT stop on first unhealthy plugin — checks all.
func (r *Registry) HealthCheckAll(ctx context.Context) map[string]plugin.HealthReport {
	plugins := r.AllPlugins()
	results := make(map[string]plugin.HealthReport, len(plugins))

	for _, p := range plugins {
		report, err := p.Health(ctx)
		if err != nil {
			// If health check itself failed (e.g. system error), enrich report status
			report.Status = plugin.HealthDown
			report.Message = fmt.Sprintf("health check failed: %v", err)
		}
		results[p.Name()] = report

		if !report.IsHealthy() && r.logger != nil {
			r.logger.Warn("plugin unhealthy",
				"name", p.Name(),
				"status", report.Status,
				"message", report.Message,
			)
		}
	}

	return results
}
```

## Rules
1. **Rollback Start Boundaries**: If `Start` fails, roll back execution by calling `Stop()` on all *already-started* components in reverse order. Do not try to clean up the component that failed to launch.
2. **Reverse Cleanup order (LIFO)**: Shutdown routines (`StopAll`) must stop components in reverse registration order (e.g. Agents → Providers → Registry → EventBus) to ensure callers are deactivated before their dependency layers stop.
3. **Structured Health Reporting**: Align `HealthCheckAll` outputs with the updated `Plugin.Health` signature (`(HealthReport, error)`), returning a map keyed by name containing `plugin.HealthReport` records.
4. **Isolated Rollback Contexts**: Cleanup rollbacks triggered inside `StartAll` failures must run under a fresh `context.Background()` context to ensure execution even if the startup context was canceled.

## ⚠️ Pitfalls

### Pitfall 1: Bypassing rollbacks for partially started clusters
If component startup crashes halfway through (e.g. A starts, B starts, C fails), failing to stop A and B leaves open sockets and leaked goroutines. Perform cleanups inside rollbacks.

### Pitfall 2: Crashing compiler checks from outdated `HealthCheckAll` signatures
```go
```
Always use `plugin.HealthReport` maps inside `HealthCheckAll` calls to align package layers.

## Verify
```bash
go build ./kernel/registry/...
```

## Checklist
- [ ] File `kernel/registry/lifecycle.go` exists
- [ ] Package: `registry`
- [ ] `InitAll` initializes plugins in registration order
- [ ] `StartAll` rolls back partially initialized components in reverse order
- [ ] `StopAll` stops components in reverse order (LIFO)
- [ ] `stopPlugins` uses fresh `context.Background()` contexts for rollback runs
- [ ] `HealthCheckAll` returns `map[string]plugin.HealthReport` records
- [ ] Log calls include safety guards checking for nil loggers
- [ ] `go build ./kernel/registry/...` passes
