package agent

import "context"

// Agent is the core interface that all AI agents must implement.
//
// Lifecycle:
//
//	Agent lifecycle (Init, Start, Stop) is managed by the Plugin interface
//	in contracts/plugin. This interface only defines runtime execution behavior.
type Agent interface {
	// Name returns the unique identifier for this agent (e.g., "backend").
	Name() string

	// Role returns the human-readable role description (e.g., "Backend Developer").
	Role() string

	// Capabilities returns the list of capabilities this agent has.
	Capabilities() []Capability

	// Execute performs a task and returns the execution result.
	//
	// Error handling convention:
	//   - System errors (e.g. panic, network down): return (nil, error)
	//   - Task failures (e.g. AI logic failure): return (Result{Status: StatusFailed}, nil)
	//   - Never return both non-nil Result and non-nil error simultaneously.
	Execute(ctx context.Context, task *Task) (*Result, error)

	// CanHandle checks if this agent is capable of executing the given task.
	// Must execute quickly without performing I/O (ideally < 1ms).
	CanHandle(task *Task) bool
}
