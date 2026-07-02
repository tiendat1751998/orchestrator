# Micro-Task 2.19: Create kernel/registry/finder.go

## Info
- **File**: `kernel/registry/finder.go`
- **Package**: `registry`
- **Depends on**: 2.18 (registry.go)
- **Time**: 15 min
- **Verify**: `go build ./kernel/registry/...`

## Purpose
Find the best agent for a given task based on capabilities and scoring.
The orchestrator calls this when it needs to dispatch a task.

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
// Why not scoring/ranking?
//   For Phase 2, simple first-match is sufficient.
//   Scoring (based on past performance) will be added in Phase 5 via the
//   feedback.Scorer interface. The algorithm here can be swapped out without
//   changing the registry interface.
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
//
// Useful for:
//   - Parallel execution (send task to multiple agents, pick best result)
//   - Fallback chains (if agent A fails, try agent B)
//   - Debugging (see which agents could handle a task)
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
//
// Useful when you know the capability you need but don't have a full Task yet.
func (r *Registry) FindAgentByCapability(cap agent.Capability) (agent.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, name := range r.order {
		a, isAgent := r.agents[name]
		if !isAgent {
			continue
		}
		if agent.HasCapability(a.Capabilities(), cap) {
			return a, nil
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
//
// This helps the user understand WHY no agent matched and what they need to add.
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

## Pitfalls

### Pitfall 1: CanHandle must be fast
`FindAgentForTask` iterates all agents and calls `CanHandle()` on each.
If CanHandle does I/O (network, file read) → O(N × I/O time) → slow.
Contract requires: CanHandle < 1ms, no I/O.

### Pitfall 2: Deterministic ordering
```go
for _, name := range r.order {  // registration order
```
NOT `for name := range r.agents` — map iteration is random.
Random ordering → different agent selected on different runs → non-reproducible behavior.

### Pitfall 3: Detailed error message
When no agent matches, the error message MUST list:
- Task name and type
- All available agents with their capabilities
Without this, debugging "why didn't my task get assigned?" is impossible.

### Pitfall 4: RLock (not Lock) for reads
FindAgentForTask is called frequently (every task dispatch).
Using Lock instead of RLock = sequential dispatching = poor performance.

## Checklist
- [ ] File `kernel/registry/finder.go` exists
- [ ] `FindAgentForTask(task)` — first match, registration order
- [ ] `FindAllAgentsForTask(task)` — all matches
- [ ] `FindAgentByCapability(cap)` — find by single capability
- [ ] `noAgentError(task)` — detailed error with agent listing
- [ ] Uses registration order (not map iteration)
- [ ] RLock for all read operations
- [ ] Error messages include task name, type, and available agents
- [ ] `go build ./kernel/registry/...` no errors
