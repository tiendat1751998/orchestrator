# Micro-Task 3.01: Create sdk/plugin/plugin.go

## Info
- **File**: `sdk/plugin/plugin.go`
- **Package**: `plugin`
- **Depends on**: 1.24 (plugin.go contract), 1.40 (health.go report)
- **Time**: 15 min
- **Verify**: `go build ./sdk/plugin/...`

## Purpose
Implements a reusable, thread-safe base helper struct (`BasePlugin` and constructors) that implements the `contracts/plugin.Plugin` interface. It eliminates boilerplate code for custom plugins by providing default lifecycle handlers and structured health reporting.

## EXACT code to create

```go
// Package plugin provides base helpers and default implementations for the Plugin interface.
package plugin

import (
	"context"
	"errors"
	"sync"
	"time"

	cplugin "github.com/tiendat1751998/orchestrator/contracts/plugin"
)

// BasePlugin provides a thread-safe default implementation of the plugin.Plugin interface.
// Other plugins (agents, providers, tools) can embed this struct to inherit lifecycle accessors.
type BasePlugin struct {
	mu          sync.RWMutex
	name        string
	pluginType  cplugin.Type
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
func NewBasePlugin(name string, pluginType cplugin.Type, version string) (*BasePlugin, error) {
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
func (p *BasePlugin) Name() string {
	return p.name
}

// Type returns the plugin category.
func (p *BasePlugin) Type() cplugin.Type {
	return p.pluginType
}

// Version returns the plugin version.
func (p *BasePlugin) Version() string {
	return p.version
}

// Init loads configuration and validates settings.
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
func (p *BasePlugin) Health(ctx context.Context) (cplugin.HealthReport, error) {
	p.mu.RLock()
	initialized := p.initialized
	started := p.started
	p.mu.RUnlock()

	if !initialized {
		return cplugin.HealthReport{
			Status:    cplugin.HealthDown,
			Message:   "plugin not initialized",
			Timestamp: time.Now(),
		}, nil
	}

	if !started {
		return cplugin.HealthReport{
			Status:    cplugin.HealthDown,
			Message:   "plugin initialized but not started",
			Timestamp: time.Now(),
		}, nil
	}

	return cplugin.HealthReport{
		Status:    cplugin.HealthOK,
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

## Rules
1. **Package Import Aliasing**: Alias the imported `contracts/plugin` package to `cplugin` to avoid namespace collisions with the SDK package `plugin`.
2. **Access Locks**: All accessor functions (such as `IsStarted` and `IsInitialized`) must acquire read locks (`RLock()`) to prevent race conditions during updates.
3. **Idempotence of Stop Calls**: Make `Stop` operations idempotent. Repeated calls to `Stop` on already terminated plugins must execute as no-ops instead of raising errors.

## ⚠️ Pitfalls

### Pitfall 1: Package shadowing compile errors
Declaring the SDK package name as `plugin` while importing `contracts/plugin` without an alias results in namespace collisions. Use `cplugin` for the contract package.

### Pitfall 2: Race conditions on status accessors
Reading lifecycle status fields (`started`, `initialized`) directly without acquiring locking guards allows other threads to modify them in the background. Wrap access behind locks.

## Verify
```bash
go build ./sdk/plugin/...
```

## Checklist
- [ ] File `sdk/plugin/plugin.go` exists
- [ ] Package: `plugin`
- [ ] Import `contracts/plugin` aliased to `cplugin`
- [ ] Base plugin struct tracks initialized and started flags
- [ ] Constructor performs name validations
- [ ] State transitions check flags under mutex locks
- [ ] `Stop` executes cleanly on already stopped instances
- [ ] `Health` returns structured reports matching standard contracts
- [ ] Read accessors use `RLock` protections
- [ ] `go build ./sdk/plugin/...` passes
