# Micro-Task 1.21: Create contracts/agent/agent.go

## Info
- **File**: `contracts/agent/agent.go`
- **Package**: `agent`
- **Depends on**: 1.17, 1.18, 1.19
- **Time**: 10 min
- **Verify**: `go build ./contracts/agent/...`

## Purpose
Declares the core `Agent` interface that all agent persona plugins must implement.

## EXACT code to create

```go
package agent

import "context"

// Agent is the core interface that all AI agents must implement.
//
// Lifecycle:
//   Agent lifecycle (Init, Start, Stop) is managed by the Plugin interface
//   in contracts/plugin. This interface only defines runtime execution behavior.
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
```

## ⚠️ Pitfalls

### Pitfall 1: Returning both Result and error simultaneously
```go
return nil, contracts.ErrTaskTimeout // System timeout: return nil result, and the Go error.
```
Always follow the Go error convention: return either `(result, nil)` or `(nil, error)`. Never return both non-nil.

### Pitfall 2: Performing slow I/O inside CanHandle check
```go
func (a *MyAgent) CanHandle(t *Task) bool {
    return t.Type == "test" // Quick memory check
}
```
`CanHandle` is invoked frequently by the orchestrator. Keep it clean and avoid slow I/O operations inside it.

## Verify
```bash
go build ./contracts/agent/...
```

## Checklist
- [ ] File `contracts/agent/agent.go` exists
- [ ] Package: `agent`
- [ ] `Agent` interface is declared with Name, Role, Capabilities, Execute, and CanHandle methods
- [ ] `Execute` receives pointer to `Task` and returns `(*Result, error)`
- [ ] `CanHandle` receives pointer to `Task` and returns `bool`
- [ ] Interface does not contain lifecycle methods (Init/Start/Stop)
- [ ] `go build ./contracts/agent/...` passes
