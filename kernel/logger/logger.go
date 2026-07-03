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
//
//	Level: "info", Format: "text", Output: os.Stderr
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
		Level:       level,
		AddSource:   level == slog.LevelDebug, // Add source file/line in debug mode
		ReplaceAttr: replaceAttr,              // Hook for redacting sensitive fields
	}

	switch strings.ToLower(opts.Format) {
	case "json":
		handler = slog.NewJSONHandler(opts.Output, handlerOpts)
	default:
		handler = NewPrettyHandler(opts.Output, handlerOpts)
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

// replaceAttr is the slog attributes filter callback used to redact secrets.
func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if IsSensitiveField(a.Key) {
		return slog.String(a.Key, RedactedValue)
	}
	return a
}
