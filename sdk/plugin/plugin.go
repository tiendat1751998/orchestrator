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
		name:        name,
		pluginType:  pluginType,
		version:     version,
		initialized: false,
		started:     false,
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
