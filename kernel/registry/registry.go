// Package registry provides a thread-safe plugin registry.
//
// The registry is the kernel's service locator. It stores references to all
// registered plugins (agents, providers, tools) and provides lookup methods.
//
// Thread-safety: all methods are safe for concurrent use (sync.RWMutex).
//
// Registration rules:
//   - Names must be unique within a plugin type (two agents named "backend" = error)
//   - Plugins must implement the plugin.Plugin interface for lifecycle management
//   - Agents must additionally implement agent.Agent
//   - Providers must additionally implement provider.Provider
//   - Tools must additionally implement tool.Tool
package registry

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/contracts/plugin"
	"github.com/tiendat1751998/orchestrator/contracts/provider"
	"github.com/tiendat1751998/orchestrator/contracts/tool"
)

// Registry stores and retrieves plugins by name and type.
//
// Internal storage uses separate maps per plugin type for O(1) lookup.
// A master map tracks all plugins regardless of type for lifecycle management.
type Registry struct {
	mu sync.RWMutex

	// all stores every registered plugin (for lifecycle: InitAll, StartAll, StopAll).
	// Key: plugin name. Value: plugin instance.
	all map[string]plugin.Plugin

	// Type-specific maps for fast lookup.
	// These hold the SAME instances as 'all', just type-asserted for convenience.
	agents    map[string]agent.Agent
	providers map[string]provider.Provider
	tools     map[string]tool.Tool

	// order tracks registration order for deterministic Init/Start.
	// StopAll reverses this order.
	order []string

	logger *slog.Logger
}

// New creates an empty Registry.
func New(logger *slog.Logger) *Registry {
	return &Registry{
		all:       make(map[string]plugin.Plugin),
		agents:    make(map[string]agent.Agent),
		providers: make(map[string]provider.Provider),
		tools:     make(map[string]tool.Tool),
		order:     make([]string, 0),
		logger:    logger,
	}
}

// Register adds a plugin to the registry.
//
// The plugin is stored in the master map AND in the type-specific map
// based on its Type() return value.
//
// Errors:
//   - Duplicate name: two plugins with the same name
//   - Type mismatch: plugin.Type() is "agent" but it doesn't implement agent.Agent
//
// Thread-safety: acquires write lock.
func (r *Registry) Register(p plugin.Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := p.Name()

	// Check duplicate
	if _, exists := r.all[name]; exists {
		return fmt.Errorf("registry: plugin %q already registered", name)
	}

	// Store in master map
	r.all[name] = p
	r.order = append(r.order, name)

	// Store in type-specific map
	switch p.Type() {
	case plugin.TypeAgent:
		a, ok := p.(agent.Agent)
		if !ok {
			// Rollback master map registration
			delete(r.all, name)
			r.order = r.order[:len(r.order)-1]
			return fmt.Errorf("registry: plugin %q has type %q but does not implement agent.Agent interface", name, p.Type())
		}
		r.agents[name] = a

	case plugin.TypeProvider:
		prov, ok := p.(provider.Provider)
		if !ok {
			delete(r.all, name)
			r.order = r.order[:len(r.order)-1]
			return fmt.Errorf("registry: plugin %q has type %q but does not implement provider.Provider interface", name, p.Type())
		}
		r.providers[name] = prov

	case plugin.TypeTool:
		t, ok := p.(tool.Tool)
		if !ok {
			delete(r.all, name)
			r.order = r.order[:len(r.order)-1]
			return fmt.Errorf("registry: plugin %q has type %q but does not implement tool.Tool interface", name, p.Type())
		}
		r.tools[name] = t

	default:
		// Other plugin types (search, memory, workflow, context) are stored
		// only in the master map. Type-specific maps can be added later.
	}

	if r.logger != nil {
		r.logger.Info("plugin registered",
			"name", name,
			"type", p.Type(),
			"version", p.Version(),
		)
	}

	return nil
}

// Unregister removes a plugin from the registry.
//
// Returns error if plugin not found.
// Does NOT call Stop() — caller must stop the plugin first.
//
// Thread-safety: acquires write lock.
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.all[name]
	if !exists {
		return fmt.Errorf("registry: plugin %q not found", name)
	}

	// Remove from type-specific map
	switch p.Type() {
	case plugin.TypeAgent:
		delete(r.agents, name)
	case plugin.TypeProvider:
		delete(r.providers, name)
	case plugin.TypeTool:
		delete(r.tools, name)
	}

	// Remove from master map
	delete(r.all, name)

	// Remove from order slice
	for i, n := range r.order {
		if n == name {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}

	if r.logger != nil {
		r.logger.Info("plugin unregistered", "name", name)
	}

	return nil
}

// =============================================================================
// Lookup Methods
// =============================================================================

// GetAgent returns a registered agent by name.
// Returns error if not found.
func (r *Registry) GetAgent(name string) (agent.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	a, exists := r.agents[name]
	if !exists {
		return nil, fmt.Errorf("registry: agent %q not found", name)
	}
	return a, nil
}

// GetProvider returns a registered provider by name.
// Returns error if not found.
func (r *Registry) GetProvider(name string) (provider.Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("registry: provider %q not found", name)
	}
	return p, nil
}

// GetTool returns a registered tool by name.
// Returns error if not found.
func (r *Registry) GetTool(name string) (tool.Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("registry: tool %q not found", name)
	}
	return t, nil
}

// GetPlugin returns any registered plugin by name.
func (r *Registry) GetPlugin(name string) (plugin.Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.all[name]
	if !exists {
		return nil, fmt.Errorf("registry: plugin %q not found", name)
	}
	return p, nil
}

// =============================================================================
// Listing Methods
// =============================================================================

// ListAgents returns all registered agent names.
func (r *Registry) ListAgents() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.agents))
	for name := range r.agents {
		names = append(names, name)
	}
	return names
}

// ListProviders returns all registered provider names.
func (r *Registry) ListProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// ListTools returns all registered tool names.
func (r *Registry) ListTools() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// AllPlugins returns all registered plugins in registration order.
// Used by lifecycle manager for Init/Start.
func (r *Registry) AllPlugins() []plugin.Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]plugin.Plugin, 0, len(r.order))
	for _, name := range r.order {
		if p, ok := r.all[name]; ok {
			result = append(result, p)
		}
	}
	return result
}

// AllPluginsReversed returns all plugins in reverse registration order.
// Used by lifecycle manager for Stop (LIFO).
func (r *Registry) AllPluginsReversed() []plugin.Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]plugin.Plugin, 0, len(r.order))
	for i := len(r.order) - 1; i >= 0; i-- {
		if p, ok := r.all[r.order[i]]; ok {
			result = append(result, p)
		}
	}
	return result
}

// Count returns the total number of registered plugins.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.all)
}
