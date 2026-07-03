# Micro-Task 2.09: Create kernel/logger/formatter.go

## Info
- **File to create**: `kernel/logger/formatter.go`
- **File to update**: `kernel/logger/logger.go` (Change default handler to use PrettyHandler)
- **Package**: `logger`
- **Depends on**: 2.07 (logger.go)
- **Time**: 15 min
- **Verify**: `go build ./kernel/logger/...`

## Purpose
Implements a custom `slog.Handler` (`PrettyHandler`) that formats log records into colorized, human-readable terminal lines (with emojis and cyan-colored attributes). This handler is designed specifically for local development outputs when `log_format` is set to `"text"`.

## EXACT code to create

### Part 1: Create `kernel/logger/formatter.go`

```go
package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"
)

// =============================================================================
// ANSI color codes for terminal output
// =============================================================================

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

// levelColors maps log levels to terminal colors.
var levelColors = map[slog.Level]string{
	slog.LevelDebug: colorGray,
	slog.LevelInfo:  colorGreen,
	slog.LevelWarn:  colorYellow,
	slog.LevelError: colorRed,
}

// levelIcons maps log levels to emoji indicators.
var levelIcons = map[slog.Level]string{
	slog.LevelDebug: "🔍",
	slog.LevelInfo:  "✅",
	slog.LevelWarn:  "⚠️",
	slog.LevelError: "❌",
}

// PrettyHandler is a custom slog.Handler that produces colorized terminal output.
//
// Output format:
//   14:30:05 ✅ INFO  task started          task_id=tsk-123 agent=backend
//   14:30:06 ❌ ERROR task failed           error="timeout" task_id=tsk-123
//
// This handler is NOT suitable for production (parsing colored output is hard).
// Use slog.JSONHandler for production.
type PrettyHandler struct {
	opts  slog.HandlerOptions
	out   io.Writer
	mu    sync.Mutex // Protects out from concurrent writes
	attrs []slog.Attr
	group string
}

// NewPrettyHandler creates a new PrettyHandler.
func NewPrettyHandler(out io.Writer, opts *slog.HandlerOptions) *PrettyHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &PrettyHandler{
		opts: *opts,
		out:  out,
	}
}

// Enabled reports whether the handler handles records at the given level.
func (h *PrettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

// Handle writes a log record to the output.
//
// Thread-safety: mu.Lock protects the io.Writer from interleaved output
// when multiple goroutines log concurrently.
func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
	// Format timestamp (HH:MM:SS — date is noise in dev)
	timeStr := r.Time.Format("15:04:05")

	// Get color and icon for level
	color := levelColors[r.Level]
	icon := levelIcons[r.Level]
	levelStr := r.Level.String()

	// Format attributes (key=value pairs)
	var attrsStr string
	// Include pre-set attrs from With()
	for _, a := range h.attrs {
		// Run through ReplaceAttr if configured
		if h.opts.ReplaceAttr != nil {
			a = h.opts.ReplaceAttr(nil, a)
		}
		if a.Key != "" {
			attrsStr += fmt.Sprintf(" %s%s%s=%v", colorCyan, a.Key, colorReset, a.Value.Any())
		}
	}
	// Include record-specific attrs
	r.Attrs(func(a slog.Attr) bool {
		if h.opts.ReplaceAttr != nil {
			a = h.opts.ReplaceAttr(nil, a)
		}
		if a.Key != "" {
			attrsStr += fmt.Sprintf(" %s%s%s=%v", colorCyan, a.Key, colorReset, a.Value.Any())
		}
		return true
	})

	// Build the line
	line := fmt.Sprintf("%s%s%s %s %s%-5s%s %s%s%s%s\n",
		colorGray, timeStr, colorReset,
		icon,
		color, levelStr, colorReset,
		colorBold, r.Message, colorReset,
		attrsStr,
	)

	// Thread-safe write
	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.out.Write([]byte(line))
	return err
}

// WithAttrs returns a new handler with the given attributes.
func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &PrettyHandler{
		opts:  h.opts,
		out:   h.out,
		attrs: newAttrs,
		group: h.group,
	}
}

// WithGroup returns a new handler with the given group name.
func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	return &PrettyHandler{
		opts:  h.opts,
		out:   h.out,
		attrs: h.attrs,
		group: name,
	}
}

// FormatDuration formats a duration in a human-readable way.
// Used as a helper when logging durations.
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	return d.Round(time.Second).String()
}
```

---

### Part 2: Update `kernel/logger/logger.go`

Modify the initialization block in [kernel/logger/logger.go](file:///d:/project/orchestrator/kernel/logger/logger.go) to bind `NewPrettyHandler` as the default output formatter when `Format` is not `"json"`:

```go
	switch strings.ToLower(opts.Format) {
	case "json":
		handler = slog.NewJSONHandler(opts.Output, handlerOpts)
	default:
		handler = NewPrettyHandler(opts.Output, handlerOpts)
	}
```

## Rules
1. **Thread-Safe Writing**: Wrap all `io.Writer` write calls inside `PrettyHandler.Handle` with a mutex lock to prevent concurrent workers from interleaving log line outputs.
2. **Immutability Principle**: Calls to `WithAttrs` and `WithGroup` must allocate and return a *new* instance of `PrettyHandler` rather than mutating existing field slices.

## ⚠️ Pitfalls

### Pitfall 1: Mutating handler properties directly inside `WithAttrs` calls
```go
newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
copy(newAttrs, h.attrs)
copy(newAttrs[len(h.attrs):], attrs)
return &PrettyHandler{attrs: newAttrs, ...} // Returns fresh isolated child handler copy.
```
Always treat handlers as immutable structures.

### Pitfall 2: Bypassing the ReplaceAttr hook inside custom handlers
If you skip applying `h.opts.ReplaceAttr` to attributes logged inside `PrettyHandler.Handle`, the logger will print sensitive keys (like api keys or secret passwords) in plain text, bypassing redaction policies.

## Verify
```bash
go build ./kernel/logger/...
```

## Checklist
- [ ] File `kernel/logger/formatter.go` exists
- [ ] Package: `logger`
- [ ] `PrettyHandler` implements `slog.Handler` (Enabled, Handle, WithAttrs, WithGroup)
- [ ] `Handle` wraps writes using a sync Mutex
- [ ] Emojis, colors, and duration converters are defined
- [ ] `logger.go` default initialized to `NewPrettyHandler`
- [ ] Custom handler formats respect `ReplaceAttr` hooks
- [ ] `go build ./kernel/logger/...` passes
