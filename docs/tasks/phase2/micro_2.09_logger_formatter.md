# Micro-Task 2.09: Tạo kernel/logger/formatter.go

## Thông tin
- **File tạo**: `kernel/logger/formatter.go`
- **Package**: `logger`
- **Dependencies trước**: 2.07
- **Thời gian**: 15 phút
- **Verify**: `go build ./kernel/logger/...`

## Mục đích
Custom slog.Handler cho terminal output đẹp (colors, icons).
Dùng trong development mode (log_format: "text").

## Nội dung CHÍNH XÁC cần tạo

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
// ANSI color codes cho terminal output
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
// Output format:
//   HH:MM:SS ICON LEVEL  message  key=value key=value
//
// Thread-safety: mu.Lock protects the io.Writer from interleaved output
// when multiple goroutines log concurrently.
func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
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
		attrsStr += fmt.Sprintf(" %s%s%s=%v", colorCyan, a.Key, colorReset, a.Value.Any())
	}
	// Include record-specific attrs
	r.Attrs(func(a slog.Attr) bool {
		attrsStr += fmt.Sprintf(" %s%s%s=%v", colorCyan, a.Key, colorReset, a.Value.Any())
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
//
// Examples:
//
//	FormatDuration(150 * time.Millisecond) → "150ms"
//	FormatDuration(2 * time.Second)        → "2.00s"
//	FormatDuration(90 * time.Second)       → "1m30s"
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

## ⚠️ Pitfalls cần tránh

### Pitfall 1: Mutex cho io.Writer
```go
h.mu.Lock()
defer h.mu.Unlock()
_, err := h.out.Write([]byte(line))
```
Nhiều goroutines gọi Handle() cùng lúc → output bị interleaved nếu không lock.

### Pitfall 2: WithAttrs creates NEW handler (immutable)
```go
// ❌ SAI — modify existing:
h.attrs = append(h.attrs, attrs...)
return h

// ✅ ĐÚNG — create new:
newAttrs := make([]slog.Attr, ...)
return &PrettyHandler{..., attrs: newAttrs}
```
slog contract yêu cầu WithAttrs trả về handler MỚI, KHÔNG modify handler cũ.

### Pitfall 3: ANSI colors trên Windows
Windows Command Prompt cũ KHÔNG hỗ trợ ANSI codes.
Windows Terminal (mới) hỗ trợ. PowerShell 7+ hỗ trợ.
Nếu cần hỗ trợ CMD cũ → dùng slog.TextHandler thay thế PrettyHandler.
Phase này: chỉ dùng PrettyHandler cho development, TextHandler/JSONHandler cho production.

## Checklist
- [ ] File `kernel/logger/formatter.go` tồn tại
- [ ] PrettyHandler implements `slog.Handler` interface (4 methods)
- [ ] `Enabled()` checks minimum level
- [ ] `Handle()` formats: time + icon + level + message + attrs
- [ ] `WithAttrs()` returns NEW handler (immutable)
- [ ] `WithGroup()` returns NEW handler
- [ ] Mutex protects io.Writer
- [ ] ANSI color constants
- [ ] `FormatDuration()` helper
- [ ] `go build ./kernel/logger/...` không lỗi
