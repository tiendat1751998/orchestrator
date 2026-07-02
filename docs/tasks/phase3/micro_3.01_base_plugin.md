# Micro-Task 3.01: Create sdk/plugin/plugin.go

## Info
- **File**: `sdk/plugin/plugin.go`
- **Package**: `plugin`
- **Depends on**: 1.24 (plugin.go contract), 1.40 (health.go report)
- **Time**: 15 min
- **Verify**: `go build ./sdk/plugin/...`

## Purpose
Implements a reusable, thread-safe base helper struct (`BasePlugin`) that implements the `contracts/plugin.Plugin` interface. It eliminates boilerplate code for custom plugins by providing default lifecycle handlers and structured health reporting.

## EXACT code to create

```go
// Package plugin provides base helpers and default implementations for the Plugin interface.
package plugin

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/plugin"
)

// BasePlugin provides a thread-safe default implementation of the plugin.Plugin interface.
// Other plugins (agents, providers, tools) can embed this struct to inherit lifecycle accessors.
type BasePlugin struct {
	mu          sync.RWMutex
	name        string
	pluginType  plugin.Type
	version     string
	initialized bool
	started     bool
}

// NewBasePlugin constructs a new BasePlugin.
//
// Parameters:
//   - name: unique plugin identifier (lowercase alphanumeric and hyphens).
//   - pluginType: category of the plugin (agent, provider, tool, etc.).
//   - version: semver-compliant version string.
func NewBasePlugin(name string, pluginType plugin.Type, version string) (*BasePlugin, error) {
	if name == "" {
		return nil, errors.New("sdk/plugin: plugin name cannot be empty")
	}
	if version == "" {
		version = "1.0.0"
	}
	return &BasePlugin{
		name:       name,
		pluginType: pluginType,
		version:    version,
	}, nil
}

// Name returns the unique identifier for this plugin.
// Safe for concurrent use.
func (p *BasePlugin) Name() string {
	return p.name
}

// Type returns the plugin category.
// Safe for concurrent use.
func (p *BasePlugin) Type() plugin.Type {
	return p.pluginType
}

// Version returns the plugin version.
// Safe for concurrent use.
func (p *BasePlugin) Version() string {
	return p.version
}

// Init loads configuration and validates settings.
// Returns an error if called when already initialized.
func (p *BasePlugin) Init(ctx context.Context, config map[string]any) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return errors.New("sdk/plugin: plugin " + p.name + " already initialized")
	}
	p.initialized = true
	return nil
}

// Start transitions the plugin to running state.
func (p *BasePlugin) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.initialized {
		return errors.New("sdk/plugin: plugin " + p.name + " must be initialized before start")
	}
	if p.started {
		return errors.New("sdk/plugin: plugin " + p.name + " already started")
	}
	p.started = true
	return nil
}

// Stop transitions the plugin to stopped state.
func (p *BasePlugin) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.started {
		return nil // Already stopped or never started, idempotent no-op
	}
	p.started = false
	return nil
}

// Health performs a default health check.
// Returns an OK HealthReport with timestamp.
func (p *BasePlugin) Health(ctx context.Context) (plugin.HealthReport, error) {
	p.mu.RLock()
	initialized := p.initialized
	started := p.started
	p.mu.RUnlock()

	if !initialized {
		return plugin.HealthReport{
			Status:    plugin.HealthDown,
			Message:   "plugin not initialized",
			Timestamp: time.Now(),
		}, nil
	}

	if !started {
		return plugin.HealthReport{
			Status:    plugin.HealthDown,
			Message:   "plugin initialized but not started",
			Timestamp: time.Now(),
		}, nil
	}

	return plugin.HealthReport{
		Status:    plugin.HealthOK,
		Timestamp: time.Now(),
	}, nil
}

// IsInitialized returns true if the plugin is initialized.
func (p *BasePlugin) IsInitialized() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.initialized
}

// IsStarted returns true if the plugin is started.
func (p *BasePlugin) IsStarted() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.started
}
```

## ⚠️ Pitfalls

### Pitfall 1: Missing State Locks on Accessors
```go
// ❌ WRONG:
func (p *BasePlugin) IsStarted() bool {
    return p.started // Race condition if another thread is modifying in Start() or Stop()
}

// ✅ CORRECT:
func (p *BasePlugin) IsStarted() bool {
    p.mu.RLock()
    defer p.mu.RUnlock()
    return p.started
}
```
State flags accessed across threads must be guarded by locks or atomic types.

### Pitfall 2: Stop Method is not Idempotent
```go
// ❌ WRONG:
func (p *BasePlugin) Stop(ctx context.Context) error {
    if !p.started {
        return errors.New("not started") // Panic or error on second stop -> breaks shutdown recovery
    }
    p.started = false
    return nil
}

// ✅ CORRECT:
func (p *BasePlugin) Stop(ctx context.Context) error {
    p.mu.Lock()
    defer p.mu.Unlock()
    if !p.started {
        return nil // Safe no-op
    }
    p.started = false
    return nil
}
```
Shutdown loops frequently call `Stop` multiple times or on failed components. Always ensure `Stop` behaves as an idempotent no-op when already stopped.

## Verify
```bash
go build ./sdk/plugin/...
```

## Checklist
- [ ] File `sdk/plugin/plugin.go` exists
- [ ] Package: `plugin`
- [ ] Constructor `NewBasePlugin()` enforces name validations
- [ ] `BasePlugin` implements `Name()`, `Type()`, and `Version()`
- [ ] State transition checks on `Init()`, `Start()`, `Stop()` using mutex protection
- [ ] `Health()` returns structured `plugin.HealthReport` matching contract 1.40
- [ ] `IsInitialized()` and `IsStarted()` accessors are protected by RWMutex
- [ ] `go build ./sdk/plugin/...` passes
