# Micro-Task 2.20: Create kernel/registry/lifecycle.go

## Info
- **File**: `kernel/registry/lifecycle.go`
- **Package**: `registry`
- **Depends on**: 2.18 (registry.go)
- **Time**: 20 min
- **Verify**: `go build ./kernel/registry/...`

## Purpose
Manage plugin lifecycle: Init all → Start all → Stop all (reverse order).
The kernel calls these methods during startup and shutdown.

## EXACT code to create

```go
package registry

import (
	"context"
	"fmt"
	"log/slog"
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
// This ensures no half-started system (all-or-nothing startup).
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
			r.logger.Error("plugin start failed, rolling back",
				"failed_plugin", p.Name(),
				"error", err,
				"started_count", len(started),
			)
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
// Why? Agent calls Provider during execution. If Provider stops first,
// Agent tries to call a dead Provider → nil pointer → panic.
// By stopping Agent first, we ensure it's done using Provider before Provider stops.
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
	// Stop in reverse order
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
			// Continue stopping remaining plugins — don't abort
		}
	}

	if r.logger != nil {
		r.logger.Info("all plugins stopped")
	}

	return firstErr
}

// HealthCheckAll runs Health() on all plugins and returns results.
//
// Returns a map of plugin name → error (nil = healthy).
// Does NOT stop on first unhealthy plugin — checks all.
func (r *Registry) HealthCheckAll(ctx context.Context) map[string]error {
	plugins := r.AllPlugins()
	results := make(map[string]error, len(plugins))

	for _, p := range plugins {
		err := p.Health(ctx)
		results[p.Name()] = err

		if err != nil && r.logger != nil {
			r.logger.Warn("plugin unhealthy",
				"name", p.Name(),
				"error", err,
			)
		}
	}

	return results
}
```

## Pitfalls

### Pitfall 1: StartAll rollback on failure
```
Start: A ✓ → B ✓ → C ✗
Rollback: Stop B → Stop A (reverse of started, NOT reverse of all)
```
Only stop plugins that were ACTUALLY started. Don't try to stop C (it never started).

### Pitfall 2: StopAll continues on error
```go
if err := p.Stop(ctx); err != nil {
    // Log error but CONTINUE stopping remaining plugins
}
```
If Plugin A fails to stop → we STILL need to stop Plugin B, C, D.
Best-effort cleanup. Return FIRST error only.

### Pitfall 3: nil logger check
```go
if r.logger != nil {
    r.logger.Info(...)
}
```
Logger may be nil in tests. Every log call must check nil first.
Alternative: create a no-op logger. But nil check is simpler and explicit.

### Pitfall 4: stopPlugins uses context.Background()
```go
r.stopPlugins(context.Background(), started)
```
During rollback, the original context may already be cancelled.
Use a fresh `context.Background()` for rollback so cleanup actually happens.

### Pitfall 5: HealthCheckAll checks ALL plugins
```go
// NOT: return on first error
// YES: check all, return map of results
```
Returning on first error hides problems in other plugins.
Operator needs to see ALL unhealthy plugins at once.

## Checklist
- [ ] File `kernel/registry/lifecycle.go` exists
- [ ] `InitAll(ctx, configs)` — registration order, fail-fast
- [ ] `StartAll(ctx)` — registration order, rollback on failure
- [ ] `StopAll(ctx)` — REVERSE order, continue on error
- [ ] `stopPlugins(ctx, names)` — rollback helper, reverse order
- [ ] `HealthCheckAll(ctx)` — all plugins, returns map
- [ ] StartAll rollback: only stops ACTUALLY-STARTED plugins
- [ ] StopAll: returns FIRST error but stops ALL plugins
- [ ] Context cancellation check before each plugin
- [ ] Rollback uses context.Background() (not cancelled context)
- [ ] Nil logger checks on every log call
- [ ] Godoc with lifecycle order explanation
- [ ] `go build ./kernel/registry/...` no errors
