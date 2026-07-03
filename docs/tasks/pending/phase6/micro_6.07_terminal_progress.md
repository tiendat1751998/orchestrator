# Micro-Task 6.07: Create cmd/orchestrator/ui/progress.go

## Info
- **File**: `cmd/orchestrator/ui/progress.go`
- **Package**: `ui`
- **Depends on**: 6.01, 5.08 (orchestrator)
- **Time**: 30 min
- **Verify**: `go build ./cmd/orchestrator/...`

## Purpose
Implements the terminal progress renderer using ANSI escape codes for live mission tracking. Renders task DAG progress, token counter, elapsed timer, and status indicators. Pure ANSI — no heavy TUI framework dependencies.

## EXACT code to create

```go
// Package ui implements terminal rendering utilities for the orchestrator CLI.
package ui

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/tiendat1751998/orchestrator/kernel/orchestrator"
)

// ANSI escape sequences for terminal control.
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
	clearLine   = "\033[2K"
	moveUp      = "\033[%dA"
	hideCursor  = "\033[?25l"
	showCursor  = "\033[?25h"
)

// TaskStatus represents the visual state of a task.
type TaskStatus int

const (
	TaskPending TaskStatus = iota
	TaskRunning
	TaskCompleted
	TaskFailed
)

// TaskDisplay holds rendering state for a single task row.
type TaskDisplay struct {
	Name      string
	Agent     string
	Status    TaskStatus
	Duration  time.Duration
	SpinIndex int
}

// ProgressRenderer renders live mission progress to the terminal.
// Thread-safe.
type ProgressRenderer struct {
	mu           sync.Mutex
	writer       io.Writer
	missionTitle string
	tasks        []*TaskDisplay
	totalTokens  int
	startTime    time.Time
	running      bool
	stopCh       chan struct{}
	lastLines    int
}

// NewProgressRenderer constructs a new ProgressRenderer.
func NewProgressRenderer(w io.Writer) *ProgressRenderer {
	return &ProgressRenderer{
		writer: w,
		stopCh: make(chan struct{}),
	}
}

// SetMission configures the mission title for display.
func (p *ProgressRenderer) SetMission(title string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.missionTitle = title
}

// Start begins the background render loop (spinner animation).
func (p *ProgressRenderer) Start() {
	p.mu.Lock()
	p.running = true
	p.startTime = time.Now()
	p.mu.Unlock()

	fmt.Fprint(p.writer, hideCursor)

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-p.stopCh:
				return
			case <-ticker.C:
				p.render()
			}
		}
	}()
}

// Stop halts the render loop and restores cursor.
func (p *ProgressRenderer) Stop() {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return
	}
	p.running = false
	p.mu.Unlock()

	close(p.stopCh)
	fmt.Fprint(p.writer, showCursor)
}

// Update applies an orchestrator progress update to the display state.
func (p *ProgressRenderer) Update(update orchestrator.ProgressUpdate) {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch update.Type {
	case "task_started":
		p.tasks = append(p.tasks, &TaskDisplay{
			Name:   update.TaskName,
			Agent:  update.AgentName,
			Status: TaskRunning,
		})

	case "task_completed":
		for _, t := range p.tasks {
			if t.Name == update.TaskName {
				t.Status = TaskCompleted
				t.Duration = update.Duration
				break
			}
		}

	case "task_failed":
		for _, t := range p.tasks {
			if t.Name == update.TaskName {
				t.Status = TaskFailed
				t.Duration = update.Duration
				break
			}
		}

	case "tokens_update":
		p.totalTokens = update.TotalTokens
	}
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func (p *ProgressRenderer) render() {
	p.mu.Lock()
	defer p.mu.Unlock()

	var buf strings.Builder

	// Clear previous output
	if p.lastLines > 0 {
		buf.WriteString(fmt.Sprintf(moveUp, p.lastLines))
	}

	lines := 0

	// Header
	buf.WriteString(clearLine)
	buf.WriteString(fmt.Sprintf("%s🎯 Mission: %s%s\n", colorBold, p.missionTitle, colorReset))
	lines++

	buf.WriteString(clearLine)
	buf.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	lines++

	// Task rows
	for i, t := range p.tasks {
		buf.WriteString(clearLine)

		var icon, color string
		switch t.Status {
		case TaskPending:
			icon = "⏳"
			color = colorGray
		case TaskRunning:
			t.SpinIndex = (t.SpinIndex + 1) % len(spinnerFrames)
			icon = spinnerFrames[t.SpinIndex]
			color = colorYellow
		case TaskCompleted:
			icon = "✅"
			color = colorGreen
		case TaskFailed:
			icon = "❌"
			color = colorRed
		}

		durationStr := "..."
		if t.Duration > 0 {
			durationStr = fmt.Sprintf("%.1fs", t.Duration.Seconds())
		}

		buf.WriteString(fmt.Sprintf("  %s %s[%d/%d] %-30s (%s)%s %s\n",
			icon, color, i+1, len(p.tasks), t.Name, t.Agent, colorReset, durationStr))
		lines++
	}

	// Footer
	buf.WriteString(clearLine)
	buf.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	lines++

	elapsed := time.Since(p.startTime)
	buf.WriteString(clearLine)
	buf.WriteString(fmt.Sprintf("  %s⏱️  Elapsed: %s | 💰 Tokens: %s%s\n",
		colorCyan, formatDuration(elapsed), formatNumber(p.totalTokens), colorReset))
	lines++

	p.lastLines = lines
	fmt.Fprint(p.writer, buf.String())
}

// RenderFinalResult prints the completed mission summary.
func (p *ProgressRenderer) RenderFinalResult(result *orchestrator.MissionResult, elapsed time.Duration) {
	fmt.Fprintf(p.writer, "\n%s✅ Mission completed in %s%s\n", colorGreen, formatDuration(elapsed), colorReset)
	if result != nil {
		fmt.Fprintf(p.writer, "   Tasks: %d completed, %d failed\n", result.CompletedTasks, result.FailedTasks)
		fmt.Fprintf(p.writer, "   Tokens: %s total\n", formatNumber(result.TotalTokens))
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
}

func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%d,%03d", n/1000, n%1000)
}
```

## Rules
1. **Pure ANSI**: No external TUI libraries (bubbletea, termbox). ANSI escape codes work on all modern terminals including Windows Terminal.
2. **Thread-Safe Rendering**: All state mutations go through mutex. The orchestrator callback fires from a goroutine; the render loop fires from another.
3. **Cursor Management**: Hide cursor on start, restore on stop. Prevents visual flicker during redraws.

## Pitfalls

### Pitfall 1: Corrupting terminal state on crash
```go
// WRONG:
fmt.Print(hideCursor) // If program panics, cursor stays hidden forever

// CORRECT:
defer fmt.Print(showCursor) // Restore cursor even on panic
```
Always defer cursor restoration. A hidden cursor after a crash is extremely confusing for users.

### Pitfall 2: ANSI codes on legacy Windows cmd.exe
Windows cmd.exe (pre-Windows 10) does not support ANSI escape codes natively. On Windows, enable virtual terminal processing via `golang.org/x/sys/windows` kernel32 calls. Windows Terminal and PowerShell 7+ support ANSI natively.

## Verify
```bash
go build ./cmd/orchestrator/...
```

## Checklist
- [ ] File `cmd/orchestrator/ui/progress.go` exists
- [ ] Package: `ui`
- [ ] ANSI spinner animation for active tasks
- [ ] Color-coded status indicators (green/yellow/red/gray)
- [ ] Token counter and elapsed timer in footer
- [ ] Thread-safe state updates via mutex
- [ ] Cursor hidden during rendering, restored on stop
- [ ] `go build ./cmd/orchestrator/...` passes
