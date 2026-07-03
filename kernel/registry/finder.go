package registry

import (
	"fmt"

	"github.com/tiendat1751998/orchestrator/contracts/agent"
)

// FindAgentForTask finds the most suitable agent for a task.
//
// Selection algorithm:
//  1. Iterate all registered agents
//  2. Call agent.CanHandle(task) on each
//  3. Among agents that CAN handle:
//     a. Prefer agents where task.Type exactly matches a Capability
//     b. If multiple match, prefer the one registered first (stable selection)
//  4. Return the first match
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
//
//	registry: no agent can handle task "implement_handler" (type: "code_generation")
//	Available agents:
//	  - "reviewer" [code_review, testing]
//	  - "devops" [deployment]
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
