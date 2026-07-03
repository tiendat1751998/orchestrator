package agentcontext

import (
	"context"
	"reflect"
	"testing"
)

func TestMetadataContext(t *testing.T) {
	// Test nil contexts don't panic and return empty string/map
	t.Run("NilContext", func(t *testing.T) {
		var nilCtx context.Context

		if got := TraceIDFrom(nilCtx); got != "" {
			t.Errorf("TraceIDFrom(nil) = %q; want empty string", got)
		}
		if got := TaskIDFrom(nilCtx); got != "" {
			t.Errorf("TaskIDFrom(nil) = %q; want empty string", got)
		}
		if got := MissionIDFrom(nilCtx); got != "" {
			t.Errorf("MissionIDFrom(nil) = %q; want empty string", got)
		}
		if got := AgentNameFrom(nilCtx); got != "" {
			t.Errorf("AgentNameFrom(nil) = %q; want empty string", got)
		}

		gotMeta := GetMetadata(nilCtx)
		if len(gotMeta) != 0 {
			t.Errorf("GetMetadata(nil) = %v; want empty map", gotMeta)
		}
	})

	// Test basic context propagation and safe retrieval
	t.Run("BasicPropagation", func(t *testing.T) {
		ctx := context.Background()

		ctx = WithTraceID(ctx, "t-123")
		ctx = WithTaskID(ctx, "task-456")
		ctx = WithMissionID(ctx, "mission-789")
		ctx = WithAgentName(ctx, "test-agent")

		if got := TraceIDFrom(ctx); got != "t-123" {
			t.Errorf("TraceIDFrom = %q; want %q", got, "t-123")
		}
		if got := TaskIDFrom(ctx); got != "task-456" {
			t.Errorf("TaskIDFrom = %q; want %q", got, "task-456")
		}
		if got := MissionIDFrom(ctx); got != "mission-789" {
			t.Errorf("MissionIDFrom = %q; want %q", got, "mission-789")
		}
		if got := AgentNameFrom(ctx); got != "test-agent" {
			t.Errorf("AgentNameFrom = %q; want %q", got, "test-agent")
		}

		expectedMeta := map[string]string{
			"trace_id":   "t-123",
			"task_id":    "task-456",
			"mission_id": "mission-789",
			"agent":      "test-agent",
		}
		gotMeta := GetMetadata(ctx)
		if !reflect.DeepEqual(gotMeta, expectedMeta) {
			t.Errorf("GetMetadata = %v; want %v", gotMeta, expectedMeta)
		}
	})

	// Test that empty strings in setters return the original context without modification
	t.Run("EmptyStrings", func(t *testing.T) {
		ctx := context.Background()

		ctxWithEmptyTrace := WithTraceID(ctx, "")
		if ctxWithEmptyTrace != ctx {
			t.Error("WithTraceID with empty traceID should return original context")
		}

		ctxWithEmptyTask := WithTaskID(ctx, "")
		if ctxWithEmptyTask != ctx {
			t.Error("WithTaskID with empty taskID should return original context")
		}

		ctxWithEmptyMission := WithMissionID(ctx, "")
		if ctxWithEmptyMission != ctx {
			t.Error("WithMissionID with empty missionID should return original context")
		}

		ctxWithEmptyAgent := WithAgentName(ctx, "")
		if ctxWithEmptyAgent != ctx {
			t.Error("WithAgentName with empty name should return original context")
		}
	})

	// Test wrong type inside context doesn't panic and returns empty string
	t.Run("WrongTypeAssert", func(t *testing.T) {
		// Create a context where keys are the correct private type but store wrong types (integers instead of strings)
		ctx := context.Background()
		ctx = context.WithValue(ctx, keyTraceID, 123)
		ctx = context.WithValue(ctx, keyTaskID, 456)
		ctx = context.WithValue(ctx, keyMissionID, 789)
		ctx = context.WithValue(ctx, keyAgentName, 999)

		if got := TraceIDFrom(ctx); got != "" {
			t.Errorf("TraceIDFrom wrong type = %q; want empty string", got)
		}
		if got := TaskIDFrom(ctx); got != "" {
			t.Errorf("TaskIDFrom wrong type = %q; want empty string", got)
		}
		if got := MissionIDFrom(ctx); got != "" {
			t.Errorf("MissionIDFrom wrong type = %q; want empty string", got)
		}
		if got := AgentNameFrom(ctx); got != "" {
			t.Errorf("AgentNameFrom wrong type = %q; want empty string", got)
		}

		gotMeta := GetMetadata(ctx)
		if len(gotMeta) != 0 {
			t.Errorf("GetMetadata wrong type = %v; want empty map", gotMeta)
		}
	})
}
