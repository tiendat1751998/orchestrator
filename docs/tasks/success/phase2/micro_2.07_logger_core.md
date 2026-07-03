# Micro-Task 2.07: Create kernel/logger/logger.go

## Info
- **File**: `kernel/logger/logger.go`
- **Package**: `logger`
- **Depends on**: 2.01 (config struct)
- **Time**: 20 min
- **Verify**: `go build ./kernel/logger/...`

## Purpose
Establishes the structured logger (`Logger`, `Options`) using Go's standard library `log/slog` to support JSON (production) and Text (development) logging formats.

## EXACT code to create

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
	Level string

	// Format sets the output format.
	// Values: "json" (structured), "text" (human-readable)
	Format string

	// Output is the writer for log output.
	// Default: os.Stderr (diagnostic output should go to stderr)
	Output io.Writer
}

// Logger wraps slog.Logger with additional functionality.
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
		ReplaceAttr: replaceAttr,            // Hook for redacting sensitive fields
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
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		slog: l.slog.With(args...),
	}
}

// WithGroup returns a new Logger with the given group name.
// Attributes are nested under the group name in JSON output.
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{
		slog: l.slog.WithGroup(name),
	}
}

// Slog returns the underlying slog.Logger.
func (l *Logger) Slog() *slog.Logger {
	return l.slog
}
```

> [!NOTE]
> The `replaceAttr` reference will be implemented inside the redaction/formatting step in Task 2.10. For intermediate compilation, we will define a temporary placeholder inside `logger.go` if needed, or simply proceed. Let's make sure `replaceAttr` is resolved safely. If `replaceAttr` is not yet defined, we can define a placeholder:
> ```go
> func replaceAttr(groups []string, a slog.Attr) slog.Attr { return a }
> ```
> inside `logger.go` to guarantee compilation, which will be overwritten in Task 2.10.

```go
// Placeholder for replaceAttr to ensure compile checks pass before Task 2.10 is added.
func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	return a
}
```

## Rules
1. **Output target choice**: Diagnostic logs must write to `os.Stderr` (not `os.Stdout`). This ensures stdout remains clean for piping structured command outputs.
2. **Debug-Only caller sources**: Only capture source paths (`AddSource: true`) when the log level is strictly set to `slog.LevelDebug`. Capturing call stacks is expensive.
3. **Typo Resiliency**: If an invalid log level value (e.g. `verbose`) is configured, default to `slog.LevelInfo` rather than panicking or crashing.

## ⚠️ Pitfalls

### Pitfall 1: Mixing diagnostic logs with program outputs on stdout
```go
opts.Output = os.Stderr // Diagnostics are sent to stderr, keeping stdout clean for command results.
```
Send all runtime log statements to standard error to support clean pipeline flows.

### Pitfall 2: Performance losses from permanent source tracing
Using `AddSource: true` invokes `runtime.Caller` to look up stack frames for every log line, which slows down hot path execution. Enforce source line lookup only during debug mode runs.

## Verify
```bash
go build ./kernel/logger/...
```

## Checklist
- [ ] File `kernel/logger/logger.go` exists
- [ ] Package: `logger`
- [ ] `Options` configure Level, Format, and Output properties
- [ ] `Logger` struct wraps `*slog.Logger`
- [ ] `New` handles default parameters when zero values are passed
- [ ] Log levels parse with case-insensitive switches
- [ ] Sub-logger builders `With` and `WithGroup` exist
- [ ] Default output writer is set to `os.Stderr`
- [ ] `go build ./kernel/logger/...` passes
