# Micro-Task 2.08: Tạo kernel/logger/fields.go

## Thông tin
- **File tạo**: `kernel/logger/fields.go`
- **Package**: `logger`
- **Dependencies trước**: 2.07 (logger.go)
- **Thời gian**: 10 phút
- **Verify**: `go build ./kernel/logger/...`

## Mục đích
Định nghĩa standard field names và convenience methods cho common patterns.
Đảm bảo TẤT CẢ components dùng CÙNG field names.

## Nội dung CHÍNH XÁC cần tạo

```go
package logger

// =============================================================================
// Standard field name constants
// =============================================================================
//
// WHY constants?
// → Typos in string literals are invisible bugs:
//     log.Info("done", "taks_id", id)  // typo: "taks_id" instead of "task_id"
//     → No error, but structured query task_id=X won't find this log entry.
// → Constants: compiler catches typos at build time.

const (
	// FieldComponent identifies the system component.
	// Example: "kernel", "scheduler", "registry", "runtime"
	FieldComponent = "component"

	// FieldTaskID identifies a specific task.
	// Format: "tsk-" + random hex
	FieldTaskID = "task_id"

	// FieldMissionID identifies a specific mission.
	// Format: "msn-" + random hex
	FieldMissionID = "mission_id"

	// FieldAgentName identifies the agent executing a task.
	// Example: "backend", "reviewer", "devops"
	FieldAgentName = "agent"

	// FieldProviderName identifies the AI provider.
	// Example: "antigravity", "gemini-api"
	FieldProviderName = "provider"

	// FieldToolName identifies a tool being called.
	// Example: "read_file", "run_command", "git_commit"
	FieldToolName = "tool"

	// FieldDuration records execution time.
	// Value type: time.Duration (displayed as "1.234s")
	FieldDuration = "duration"

	// FieldError records an error message.
	// Value type: string or error
	FieldError = "error"

	// FieldStatus records a status value.
	// Example: "success", "failed", "timeout"
	FieldStatus = "status"

	// FieldTokens records token usage.
	// Value type: int
	FieldTokens = "tokens"

	// FieldEventType records an event type.
	// Example: "task.started", "task.completed"
	FieldEventType = "event_type"
)

// =============================================================================
// Convenience methods for creating component loggers
// =============================================================================

// WithTask creates a sub-logger with task_id pre-set.
// All subsequent log calls include the task_id automatically.
//
// Usage:
//
//	taskLog := log.WithTask("tsk-a1b2c3d4")
//	taskLog.Info("executing")       // includes task_id=tsk-a1b2c3d4
//	taskLog.Info("completed")       // includes task_id=tsk-a1b2c3d4
func (l *Logger) WithTask(taskID string) *Logger {
	return l.With(FieldTaskID, taskID)
}

// WithMission creates a sub-logger with mission_id pre-set.
func (l *Logger) WithMission(missionID string) *Logger {
	return l.With(FieldMissionID, missionID)
}

// WithAgent creates a sub-logger with agent name pre-set.
func (l *Logger) WithAgent(agentName string) *Logger {
	return l.With(FieldAgentName, agentName)
}

// WithComponent creates a sub-logger with component name pre-set.
//
// Usage:
//
//	kernelLog := log.WithComponent("kernel")
//	registryLog := log.WithComponent("registry")
func (l *Logger) WithComponent(name string) *Logger {
	return l.With(FieldComponent, name)
}

// WithProvider creates a sub-logger with provider name pre-set.
func (l *Logger) WithProvider(name string) *Logger {
	return l.With(FieldProviderName, name)
}

// WithTool creates a sub-logger with tool name pre-set.
func (l *Logger) WithTool(name string) *Logger {
	return l.With(FieldToolName, name)
}
```

## ⚠️ Pitfalls cần tránh

### Pitfall 1: Field name convention = snake_case
```
✅ ĐÚNG: "task_id", "agent", "mission_id"
❌ SAI:   "taskId", "AgentName", "MISSION_ID"
```
snake_case = JSON convention, compatible với log aggregators (Elastic, Loki).

### Pitfall 2: WithTask returns NEW logger
```go
log.WithTask("tsk-123")  // ❌ SAI — discards the result
taskLog := log.WithTask("tsk-123")  // ✅ ĐÚNG — use returned logger
```
`With*()` methods return a NEW Logger instance. Original logger is unchanged (immutable pattern).

## Checklist
- [ ] File `kernel/logger/fields.go` tồn tại
- [ ] 11 field name constants
- [ ] 6 convenience methods: WithTask, WithMission, WithAgent, WithComponent, WithProvider, WithTool
- [ ] All convenience methods return `*Logger`
- [ ] snake_case naming consistently
- [ ] Godoc examples
- [ ] `go build ./kernel/logger/...` không lỗi
