# Micro-Task 2.19: Create kernel/registry/finder.go

## Info
- **File**: `kernel/registry/finder.go`
- **Package**: `registry`
- **Depends on**: 2.18 (registry.go)
- **Time**: 15 min
- **Verify**: `go build ./kernel/registry/...`

## Purpose
Implements agent discovery utilities (`FindAgentForTask`, `FindAllAgentsForTask`, `FindAgentByCapability`) that scan registered agents to match tasks based on capabilities.

## EXACT code to create

```go
package registry

import (
	"fmt"

	"github.com/tiendat1751998/orchestrator/contracts/agent"
)

// FindAgentForTask finds the most suitable agent for a task.
//
// Selection algorithm:
//   1. Iterate all registered agents
//   2. Call agent.CanHandle(task) on each
//   3. Among agents that CAN handle:
//      a. Prefer agents where task.Type exactly matches a Capability
//      b. If multiple match, prefer the one registered first (stable selection)
//   4. Return the first match
//
// Returns error if no agent can handle the task.
// The error message lists all available agents and their capabilities
// so the user can diagnose what's missing.
func (r *Registry) FindAgentForTask(task *agent.Task) (agent.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Iterate in registration order (deterministic)
	for _, name := range r.order {
		a, isAgent := r.agents[name]
		if !isAgent {
			continue // Skip non-agent plugins
		}

		if a.CanHandle(task) {
			if r.logger != nil {
				r.logger.Debug("agent matched for task",
					"agent", name,
					"task_name", task.Name,
					"task_type", task.Type,
				)
			}
			return a, nil
		}
	}

	return nil, r.noAgentError(task)
}

// FindAllAgentsForTask returns ALL agents that can handle a task.
func (r *Registry) FindAllAgentsForTask(task *agent.Task) []agent.Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matches []agent.Agent
	for _, name := range r.order {
		a, isAgent := r.agents[name]
		if !isAgent {
			continue
		}
		if a.CanHandle(task) {
			matches = append(matches, a)
		}
	}
	return matches
}

// FindAgentByCapability finds the first agent with a specific capability.
func (r *Registry) FindAgentByCapability(cap agent.Capability) (agent.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, name := range r.order {
		a, isAgent := r.agents[name]
		if !isAgent {
			continue
		}
		// Scan capabilities list manually to verify match
		for _, c := range a.Capabilities() {
			if c == cap {
				return a, nil
			}
		}
	}

	return nil, fmt.Errorf("registry: no agent found with capability %q", cap)
}

// noAgentError builds a detailed error message listing all available agents.
//
// Example output:
//   registry: no agent can handle task "implement_handler" (type: "code_generation")
//   Available agents:
//     - "reviewer" [code_review, testing]
//     - "devops" [deployment]
func (r *Registry) noAgentError(task *agent.Task) error {
	if len(r.agents) == 0 {
		return fmt.Errorf(
			"registry: no agent can handle task %q (type: %q) — no agents are registered",
			task.Name, task.Type,
		)
	}

	msg := fmt.Sprintf(
		"registry: no agent can handle task %q (type: %q)\nAvailable agents:\n",
		task.Name, task.Type,
	)

	for _, name := range r.order {
		a, isAgent := r.agents[name]
		if !isAgent {
			continue
		}

		caps := make([]string, len(a.Capabilities()))
		for i, c := range a.Capabilities() {
			caps[i] = string(c)
		}
		msg += fmt.Sprintf("  - %q %v\n", name, caps)
	}

	return fmt.Errorf("%s", msg)
}
```

## Rules
1. **Deterministic Selection order**: Loop search targets using `order` slices to guarantee selection outcomes. Do not iterate map keys directly, which introduces random select behavior.
2. **I/O Free Match checks**: The checking functions (e.g. `CanHandle`) must run instantly (under 1ms) without executing file reads or network requests.
3. **Trace diagnostics**: When agent discovery fails, the returned error must list all available agents and their registered capability arrays to simplify debugging.

## ⚠️ Pitfalls

### Pitfall 1: Iterating maps directly when matching agents
```go
for _, name := range r.order { // Iterating the registered order slice ensures stable, reproducible agent matching.
    if a, ok := r.agents[name]; ok && a.CanHandle(task) { return a, nil }
}
```
Always use indexing slices to make loop selections deterministic.

### Pitfall 2: Omiting agent details during routing failures
Returning a simple `"no agent matched"` error when routing tasks makes debugging difficult. Always build detailed trace errors listing all registered agents and their capability configurations.

## Verify
```bash
go build ./kernel/registry/...
```

## Checklist
- [ ] File `kernel/registry/finder.go` exists
- [ ] Package: `registry`
- [ ] `FindAgentForTask` checks agents in registration order
- [ ] `FindAllAgentsForTask` aggregates all matches in an array
- [ ] `FindAgentByCapability` searches for capability matches
- [ ] `noAgentError` formats detailed listings of registered agents
- [ ] Lookups use `RLock` read locks
- [ ] `go build ./kernel/registry/...` passes
