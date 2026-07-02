# Micro-Task 2.07: Tạo kernel/logger/logger.go

## Thông tin
- **File tạo**: `kernel/logger/logger.go`
- **Package**: `logger`
- **Dependencies trước**: 2.01 (config struct)
- **Thời gian**: 20 phút
- **Verify**: `go build ./kernel/logger/...`

## Mục đích
Structured logging dựa trên `log/slog` standard library.
Hỗ trợ JSON (production) và Text (development) output.

## Tại sao dùng slog?
1. Standard library (Go 1.21+) → không dependency ngoài
2. Structured logging (key=value) → machine-parseable
3. Performance tốt (zero-allocation hot path)
4. Go team maintain → stable API

## Nội dung CHÍNH XÁC cần tạo

```go
// Package logger provides structured logging built on Go's log/slog.
//
// Usage:
//
//	log := logger.New(logger.Options{Level: "info", Format: "text"})
//	log.Info("task started", "task_id", "tsk-123", "agent", "backend")
//	log.Error("task failed", "error", err, "task_id", "tsk-123")
//
// Log levels: Debug < Info < Warn < Error
// Log formats: "json" (structured, for production), "text" (colored, for dev)
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Options configures the logger.
type Options struct {
	// Level sets the minimum log level.
	// Values: "debug", "info", "warn", "error"
	// Default: "info"
	Level string

	// Format sets the output format.
	// Values: "json" (structured), "text" (human-readable)
	// Default: "text"
	Format string

	// Output is the writer for log output.
	// Default: os.Stderr (NOT os.Stdout — stdout is for application output)
	//
	// WHY Stderr?
	// → Stdout is for program output (results, data).
	// → Stderr is for diagnostics (logs, errors).
	// → This allows: orchestrator 2>error.log | process_output
	Output io.Writer
}

// Logger wraps slog.Logger with additional functionality.
//
// WHY wrap instead of using slog.Logger directly?
// → Allows adding methods (WithTask, WithAgent) for common patterns.
// → Allows swapping implementation in tests (mock logger).
// → Allows adding redaction logic (see redact.go).
type Logger struct {
	slog *slog.Logger
}

// New creates a new Logger with the given options.
//
// If options have zero values, defaults are applied:
//   Level: "info", Format: "text", Output: os.Stderr
func New(opts Options) *Logger {
	// Apply defaults
	if opts.Level == "" {
		opts.Level = "info"
	}
	if opts.Format == "" {
		opts.Format = "text"
	}
	if opts.Output == nil {
		opts.Output = os.Stderr
	}

	level := parseLevel(opts.Level)

	var handler slog.Handler
	handlerOpts := &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug, // Add source file/line in debug mode
	}

	switch strings.ToLower(opts.Format) {
	case "json":
		handler = slog.NewJSONHandler(opts.Output, handlerOpts)
	default:
		handler = slog.NewTextHandler(opts.Output, handlerOpts)
	}

	return &Logger{
		slog: slog.New(handler),
	}
}

// parseLevel converts string log level to slog.Level.
//
// WHY function instead of map lookup?
// → Switch is exhaustive — compiler warns on missing cases (with linter).
// → Default to Info for unknown values (fail-safe, not fail-hard).
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// =============================================================================
// Standard log methods
// =============================================================================

// Debug logs at Debug level (detailed diagnostic info, not shown in production).
func (l *Logger) Debug(msg string, args ...any) {
	l.slog.Debug(msg, args...)
}

// Info logs at Info level (general operational events).
func (l *Logger) Info(msg string, args ...any) {
	l.slog.Info(msg, args...)
}

// Warn logs at Warn level (potential issues that don't prevent operation).
func (l *Logger) Warn(msg string, args ...any) {
	l.slog.Warn(msg, args...)
}

// Error logs at Error level (failures that need attention).
func (l *Logger) Error(msg string, args ...any) {
	l.slog.Error(msg, args...)
}

// =============================================================================
// Context-aware log methods
// =============================================================================

// DebugContext logs at Debug level with context (for request tracing).
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.slog.DebugContext(ctx, msg, args...)
}

// InfoContext logs at Info level with context.
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.slog.InfoContext(ctx, msg, args...)
}

// WarnContext logs at Warn level with context.
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.slog.WarnContext(ctx, msg, args...)
}

// ErrorContext logs at Error level with context.
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.slog.ErrorContext(ctx, msg, args...)
}

// =============================================================================
// Sub-logger creation
// =============================================================================

// With returns a new Logger with the given attributes always included.
//
// Use this to create component-specific loggers:
//
//	kernelLog := log.With("component", "kernel")
//	kernelLog.Info("started")
//	// Output: time=... level=INFO msg=started component=kernel
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		slog: l.slog.With(args...),
	}
}

// WithGroup returns a new Logger with the given group name.
// Attributes are nested under the group name in JSON output.
//
//	log.WithGroup("provider").Info("connected", "name", "antigravity")
//	// JSON output: {"provider": {"name": "antigravity"}}
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{
		slog: l.slog.WithGroup(name),
	}
}

// Slog returns the underlying slog.Logger.
// Use when you need to pass logger to libraries that accept *slog.Logger.
func (l *Logger) Slog() *slog.Logger {
	return l.slog
}
```

## ⚠️ Pitfalls cần tránh

### Pitfall 1: Output = Stderr, NOT Stdout
```go
// ❌ SAI:
opts.Output = os.Stdout
// Mixes logs with application output → piping broken

// ✅ ĐÚNG:
opts.Output = os.Stderr
// orchestrator 2>error.log | process_output → works correctly
```

### Pitfall 2: AddSource chỉ ở Debug
`AddSource: true` thêm file:line vào mỗi log entry. Hữu ích cho debugging nhưng CHẬM (runtime.Caller()). Chỉ bật ở Debug level.

### Pitfall 3: slog args format
```go
// ❌ SAI — positional format:
log.Info(fmt.Sprintf("task %s started by %s", taskID, agentName))

// ✅ ĐÚNG — structured key-value:
log.Info("task started", "task_id", taskID, "agent", agentName)
// Output: time=... level=INFO msg="task started" task_id=tsk-123 agent=backend
```
Structured logs cho phép query: `task_id=tsk-123 AND level=ERROR`.

### Pitfall 4: Log field naming consistency
```go
// ❌ SAI — inconsistent naming:
log.Info("started", "taskId", id)    // camelCase
log.Info("done", "task_id", id)      // snake_case
log.Info("failed", "TaskID", id)     // PascalCase

// ✅ ĐÚNG — ALWAYS snake_case:
log.Info("started", "task_id", id)
log.Info("done", "task_id", id)
log.Info("failed", "task_id", id)
```

### Pitfall 5: Unknown log level = Info (fail-safe)
```go
// User typo in config: log_level: "verbose"
// Result: defaults to Info (NOT error, NOT panic)
// Reason: logging system crashing = worse than wrong log level
```

## Checklist
- [ ] File `kernel/logger/logger.go` tồn tại
- [ ] Package: `package logger`
- [ ] Options struct (Level, Format, Output)
- [ ] Logger struct wrapping `*slog.Logger`
- [ ] `New(opts)` constructor with defaults
- [ ] `parseLevel()` — string → slog.Level
- [ ] 4 standard methods: Debug, Info, Warn, Error
- [ ] 4 context methods: DebugContext, InfoContext, WarnContext, ErrorContext
- [ ] `With()` — create sub-logger with persistent fields
- [ ] `WithGroup()` — create sub-logger with group
- [ ] `Slog()` — expose underlying slog.Logger
- [ ] JSON và Text handlers
- [ ] AddSource chỉ ở Debug level
- [ ] Default output: os.Stderr
- [ ] Dùng `log/slog` standard library (KHÔNG zap, logrus)
- [ ] Godoc comments với usage examples
- [ ] `go build ./kernel/logger/...` không lỗi
