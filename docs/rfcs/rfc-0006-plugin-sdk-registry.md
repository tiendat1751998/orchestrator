# RFC-0006: Plugin SDK & Registry

- **Status**: PROPOSED
- **Priority**: P1 — Core
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0000 (State Machine), RFC-0001 (Kernel Architecture)

## Summary

This RFC details the design of the Plugin SDK and Registry in AEOS. Instead of multiple separate registries, the system defines a unified type-safe generic `Registry[T]` to manage Agents, Tools, and Providers. Plugins implement explicit lifecycle methods and report their capabilities to enable dynamic discovery.

## Motivation

Issue 9 from the architecture review suggested having 6 different registries (Plugin Registry, Capability Registry, Tool Registry, etc.). In Go, having 6 separate struct types with nearly identical CRUD logic creates code duplication. By leveraging Go generics, we can build a single, robust `Registry[T]` and implement specialised indexing (like Capability matching) as decorators or methods on the generic registry. This enforces the DRY principle while keeping full compile-time type safety.

## Design

### 1. Unified Plugin Interface (`contracts/plugin/`)

Every plugin in the system (whether an Agent, Tool, or Provider) must implement the base `Plugin` interface:

```go
// contracts/plugin/plugin.go
package plugin

import "context"

type PluginType string

const (
	TypeAgent    PluginType = "agent"
	TypeTool     PluginType = "tool"
	TypeProvider PluginType = "provider"
)

// PluginMetadata holds identity and lifecycle configurations.
type PluginMetadata struct {
	Name        string     `json:"name"`        // e.g. "gemini-provider"
	Version     string     `json:"version"`     // e.g. "1.0.0"
	Type        PluginType `json:"type"`
	Description string     `json:"description"`
	Author      string     `json:"author,omitempty"`
}

// Plugin is the base contract for all components loaded into the OS.
type Plugin interface {
	// Metadata returns the plugin's identity.
	Metadata() PluginMetadata

	// Init is called once when the plugin is loaded (configuration parsing).
	Init(ctx context.Context, config map[string]any) error

	// Start is called during kernel boot or hot-loading.
	Start(ctx context.Context) error

	// Stop is called during graceful shutdown or hot-unloading.
	Stop(ctx context.Context) error
}
```

---

### 2. Specialized Contracts (`contracts/`)

#### A. Providers (`contracts/provider/`)

As defined in RFC-0001, providers translate requests and responses. They are split into `APIProvider` and `CLIProvider`.

```go
// contracts/provider/provider.go
package provider

import (
	"context"
	"github.com/tiendat1751998/orchestrator/contracts/plugin"
)

type ProviderInfo struct {
	ModelName   string `json:"model_name"`
	MaxTokens   int    `json:"max_tokens"`
	SupportsImg bool   `json:"supports_img"`
}

type Provider interface {
	plugin.Plugin
	Info() ProviderInfo
	Capabilities() []string
}
```

#### B. Tools (`contracts/tool/`)

Tools are execution blocks that agents can run.

```go
// contracts/tool/tool.go
package tool

import (
	"context"
	"github.com/tiendat1751998/orchestrator/contracts/plugin"
)

// ToolSchema defines input requirements in JSON Schema format.
type ToolSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"` // JSON Schema structure
}

type Tool interface {
	plugin.Plugin
	Schema() ToolSchema
	Execute(ctx context.Context, args map[string]any) (map[string]any, error)
}
```

#### C. Agents (`contracts/agent/`)

Agents perform business tasks using tools and providers.

```go
// contracts/agent/agent.go
package agent

import (
	"context"
	"github.com/tiendat1751998/orchestrator/contracts/plugin"
)

type Agent interface {
	plugin.Plugin
	// Capabilities returns the task types this agent handles (e.g., "code_generation").
	Capabilities() []string
	// Execute performs a specific task.
	Execute(ctx context.Context, task *Task) (*Result, error)
}
```

---

### 3. Generic Registry (`kernel/plugin/`)

The kernel implements a thread-safe registry utilizing Go generics:

```go
// kernel/plugin/registry.go
package plugin

import (
	"fmt"
	"sync"

	"github.com/tiendat1751998/orchestrator/contracts/plugin"
	"github.com/tiendat1751998/orchestrator/contracts/tool"
)

// Registry manages a collection of plugins of type T.
type Registry[T plugin.Plugin] struct {
	mu           sync.RWMutex
	plugins      map[string]T
	capabilities map[string][]string // capability/tag -> []pluginName
}

func NewRegistry[T plugin.Plugin]() *Registry[T] {
	return &Registry[T]{
		plugins:      make(map[string]T),
		capabilities: make(map[string][]string),
	}
}

// Register adds a plugin to the registry.
func (r *Registry[T]) Register(p T) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	meta := p.Metadata()
	if _, exists := r.plugins[meta.Name]; exists {
		return fmt.Errorf("plugin %s already registered", meta.Name)
	}

	r.plugins[meta.Name] = p
	r.indexPlugin(meta.Name, p)
	return nil
}

// Get retrieves a plugin by name.
func (r *Registry[T]) Get(name string) (T, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.plugins[name]
	return p, exists
}

// List returns all registered plugins.
func (r *Registry[T]) List() []T {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]T, 0, len(r.plugins))
	for _, p := range r.plugins {
		list = append(list, p)
	}
	return list
}

// FindByCapability returns plugins that match a capability.
func (r *Registry[T]) FindByCapability(cap string) []T {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := r.capabilities[cap]
	result := make([]T, 0, len(names))
	for _, name := range names {
		if p, ok := r.plugins[name]; ok {
			result = append(result, p)
		}
	}
	return result
}

// Unregister removes a plugin and triggers its Stop method.
func (r *Registry[T]) Unregister(ctx context.Context, name string) error {
	r.mu.Lock()
	p, exists := r.plugins[name]
	if !exists {
		r.mu.Unlock()
		return nil
	}
	delete(r.plugins, name)
	
	// Rebuild capability index
	r.rebuildIndex()
	r.mu.Unlock()

	return p.Stop(ctx)
}

func (r *Registry[T]) indexPlugin(name string, p T) {
	if agentPlugin, ok := any(p).(interface{ Capabilities() []string }); ok {
		for _, cap := range agentPlugin.Capabilities() {
			r.capabilities[cap] = append(r.capabilities[cap], name)
		}
	}
	if toolPlugin, ok := any(p).(interface{ Schema() tool.ToolSchema }); ok {
		toolName := toolPlugin.Schema().Name
		r.capabilities[toolName] = append(r.capabilities[toolName], name)
	}
}

func (r *Registry[T]) rebuildIndex() {
	r.capabilities = make(map[string][]string)
	for name, p := range r.plugins {
		r.indexPlugin(name, p)
	}
}
```

---

### 4. Plugin Manager (`kernel/plugin/lifecycle.go`)

The `PluginManager` handles loading plugins, parsing configuration, calling `Init`, and orchestrating hot-reloads.

```go
// kernel/plugin/manager.go
package plugin

import (
	"context"
	
	"github.com/tiendat1751998/orchestrator/contracts/plugin"
)

type PluginManager interface {
	// Load registers and starts a compiled Go plugin (.so) or configuration-defined script.
	Load(ctx context.Context, path string, config map[string]any) (plugin.Plugin, error)
	// Unload stops and removes a plugin.
	Unload(ctx context.Context, name string) error
	
	// Registries
	Agents() *Registry[agent.Agent]
	Tools() *Registry[tool.Tool]
	Providers() *Registry[provider.Provider]
}
```

## Impact

- **Dry Generic Registry**: Leverages Go generics `[T plugin.Plugin]` to index Agents, Tools, and Providers with 0 code duplication.
- **Dynamic Capability Matching**: `FindByCapability` inspects agent capabilities and tool schemas at registration time automatically.

## Open Questions

1. **Go Plugin Limitations**:
   - Go's standard `plugin` package (`.so` files) has strict runtime environment constraints (must match exact Go version, tags, and compiler flags).
   - *Recommendation*: For Phase 1, load plugins directly inside the main binary executable (static compiled plugins) using manual initialization code. Add dynamic gRPC plugin loading in Phase 4.
