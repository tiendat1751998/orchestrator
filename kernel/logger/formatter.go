package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
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
//
//	14:30:05 ✅ INFO  task started          task_id=tsk-123 agent=backend
//	14:30:06 ❌ ERROR task failed           error="timeout" task_id=tsk-123
//
// This handler is NOT suitable for production (parsing colored output is hard).
// Use slog.JSONHandler for production.
type PrettyHandler struct {
	opts   slog.HandlerOptions
	out    io.Writer
	mu     *sync.Mutex // Protects out from concurrent writes. Shared across WithAttrs/WithGroup copies.
	attrs  []slog.Attr
	groups []string
}

// NewPrettyHandler creates a new PrettyHandler.
func NewPrettyHandler(out io.Writer, opts *slog.HandlerOptions) *PrettyHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &PrettyHandler{
		opts: *opts,
		out:  out,
		mu:   &sync.Mutex{},
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

	// Get color and icon for level, fallback to standard ranges for custom levels
	color := levelColors[r.Level]
	if color == "" {
		if r.Level < slog.LevelInfo {
			color = colorGray
		} else if r.Level < slog.LevelWarn {
			color = colorGreen
		} else if r.Level < slog.LevelError {
			color = colorYellow
		} else {
			color = colorRed
		}
	}

	icon := levelIcons[r.Level]
	if icon == "" {
		if r.Level < slog.LevelInfo {
			icon = "🔍"
		} else if r.Level < slog.LevelWarn {
			icon = "✅"
		} else if r.Level < slog.LevelError {
			icon = "⚠️"
		} else {
			icon = "❌"
		}
	}

	levelStr := r.Level.String()

	// Format attributes (key=value pairs)
	var attrsStr string
	// Include pre-set attrs from With() (which are already resolved and replaced)
	for _, a := range h.attrs {
		var val any = a.Value.Any()
		if a.Value.Kind() == slog.KindDuration {
			val = FormatDuration(a.Value.Duration())
		}
		attrsStr += fmt.Sprintf(" %s%s%s=%v", colorCyan, a.Key, colorReset, val)
	}

	// Include record-specific attrs
	r.Attrs(func(a slog.Attr) bool {
		flattenAttr(h.groups, a, h.opts.ReplaceAttr, func(flat slog.Attr) {
			var val any = flat.Value.Any()
			if flat.Value.Kind() == slog.KindDuration {
				val = FormatDuration(flat.Value.Duration())
			}
			attrsStr += fmt.Sprintf(" %s%s%s=%v", colorCyan, flat.Key, colorReset, val)
		})
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
	if len(attrs) == 0 {
		return h
	}

	newAttrs := make([]slog.Attr, len(h.attrs))
	copy(newAttrs, h.attrs)

	for _, a := range attrs {
		flattenAttr(h.groups, a, h.opts.ReplaceAttr, func(flat slog.Attr) {
			newAttrs = append(newAttrs, flat)
		})
	}

	return &PrettyHandler{
		opts:   h.opts,
		out:    h.out,
		mu:     h.mu,
		attrs:  newAttrs,
		groups: h.groups,
	}
}

// WithGroup returns a new handler with the given group name.
func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name

	return &PrettyHandler{
		opts:   h.opts,
		out:    h.out,
		mu:     h.mu,
		attrs:  h.attrs,
		groups: newGroups,
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

// flattenAttr recursively resolves and flattens grouped attributes, applying ReplaceAttr if configured.
func flattenAttr(groups []string, a slog.Attr, replace func([]string, slog.Attr) slog.Attr, fn func(slog.Attr)) {
	a.Value = a.Value.Resolve()
	if replace != nil {
		a = replace(groups, a)
	}
	if a.Key == "" {
		return
	}

	if a.Value.Kind() == slog.KindGroup {
		attrs := a.Value.Group()
		if len(attrs) == 0 {
			return
		}
		newGroups := make([]string, len(groups)+1)
		copy(newGroups, groups)
		newGroups[len(groups)] = a.Key
		for _, child := range attrs {
			flattenAttr(newGroups, child, replace, fn)
		}
	} else {
		// Prefix the key with current groups
		if len(groups) > 0 {
			a.Key = strings.Join(groups, ".") + "." + a.Key
		}
		fn(a)
	}
}
