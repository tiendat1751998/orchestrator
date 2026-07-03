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
