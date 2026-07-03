# Micro-Task 1.38: Create contracts/context/metadata.go (Context Propagation)

## Info
- **File**: `contracts/context/metadata.go`
- **Package**: `agentcontext`
- **Depends on**: 1.28 (contracts/context/context.go)
- **Time**: 15 min
- **Verify**: `go build ./contracts/context/...`

## Purpose
Defines context propagation helpers to store and retrieve request-scoped telemetry metadata (Trace ID, Task ID, Mission ID, Agent Name) safely through nested Go routine execution context structures.

## EXACT code to create

```go
package agentcontext

import (
	"context"
)

// contextKey is a private type for context keys to prevent collisions.
type contextKey string

const (
	keyTraceID   contextKey = "trace_id"
	keyTaskID    contextKey = "task_id"
	keyMissionID contextKey = "mission_id"
	keyAgentName contextKey = "agent_name"
)

// WithTraceID returns a new context with the given Trace ID.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		return ctx
	}
	return context.WithValue(ctx, keyTraceID, traceID)
}

// TraceIDFrom extracts the Trace ID from context.
func TraceIDFrom(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(keyTraceID).(string); ok {
		return v
	}
	return ""
}

// WithTaskID returns a new context with the given Task ID.
func WithTaskID(ctx context.Context, taskID string) context.Context {
	if taskID == "" {
		return ctx
	}
	return context.WithValue(ctx, keyTaskID, taskID)
}

// TaskIDFrom extracts the Task ID from context.
func TaskIDFrom(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(keyTaskID).(string); ok {
		return v
	}
	return ""
}

// WithMissionID returns a new context with the given Mission ID.
func WithMissionID(ctx context.Context, missionID string) context.Context {
	if missionID == "" {
		return ctx
	}
	return context.WithValue(ctx, keyMissionID, missionID)
}

// MissionIDFrom extracts the Mission ID from context.
func MissionIDFrom(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(keyMissionID).(string); ok {
		return v
	}
	return ""
}

// WithAgentName returns a new context with the given Agent Name.
func WithAgentName(ctx context.Context, name string) context.Context {
	if name == "" {
		return ctx
	}
	return context.WithValue(ctx, keyAgentName, name)
}

// AgentNameFrom extracts the Agent Name from context.
func AgentNameFrom(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(keyAgentName).(string); ok {
		return v
	}
	return ""
}

// GetMetadata returns a map of all metadata present in the context.
// Useful for populating logger structured fields automatically.
func GetMetadata(ctx context.Context) map[string]string {
	meta := make(map[string]string)
	if ctx == nil {
		return meta
	}

	if traceID := TraceIDFrom(ctx); traceID != "" {
		meta["trace_id"] = traceID
	}
	if taskID := TaskIDFrom(ctx); taskID != "" {
		meta["task_id"] = taskID
	}
	if missionID := MissionIDFrom(ctx); missionID != "" {
		meta["mission_id"] = missionID
	}
	if agentName := AgentNameFrom(ctx); agentName != "" {
		meta["agent"] = agentName
	}

	return meta
}
```

## Rules
1. **Unexported Type Keys**: Keys must use unexported named types (e.g. `contextKey`) rather than base string literals to prevent naming collisions with external modules.
2. **Nil Safe Getters**: Getter methods must explicitly handle `nil` context inputs to prevent nil pointer panic crashes.
3. **Safe Assertions**: Always execute interface checks using the two-value type assertion syntax: `val, ok := ctx.Value(key).(string)`.

## ⚠️ Pitfalls

### Pitfall 1: Using plain string literals as context keys
```go
type contextKey string
const keyTraceID contextKey = "trace_id"
ctx = context.WithValue(ctx, keyTraceID, id) // Guaranteed unique by Go type system.
```
Use dedicated unexported types to block key namespaces collisions.

### Pitfall 2: Panicking on type assertion failures
Using single value type assertion (`ctx.Value(key).(string)`) directly will trigger a panic if the context contains a value of a different type or no value at all. Always assert with `v, ok := ...`.

## Verify
```bash
go build ./contracts/context/...
```

## Checklist
- [ ] File `contracts/context/metadata.go` exists
- [ ] Package: `agentcontext`
- [ ] `contextKey` type is declared as private unexported string type
- [ ] Defined trace_id, task_id, mission_id, and agent_name keys
- [ ] Setter helper methods WithXxx return modified context structures
- [ ] Getter helper methods XxxFrom handle nil context parameter checks
- [ ] Getter helper methods use safe type assertion `ok` checks
- [ ] `GetMetadata` aggregates all non-empty properties into a map
- [ ] `go build ./contracts/context/...` passes
