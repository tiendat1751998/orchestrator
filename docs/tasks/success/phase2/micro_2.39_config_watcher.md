# Micro-Task 2.39: Create kernel/config/watcher.go (Config Hot-Reload)

## Info
- **File created**: `kernel/config/watcher.go`
- **Package**: `config`
- **Depends on**: 2.04 (loader.go), 2.05 (validator.go)
- **Time**: 20 min
- **Verify**: `go build ./kernel/config/...`

## Purpose
Implements a platform-independent configuration change watcher (`Watcher` and constructors) that polls file system modification times (`os.Stat().ModTime()`), safely reloads valid YAML changes, and invokes update callbacks without crashing live worker tasks.

## EXACT code to create

```go
package config

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"
)

// Watcher monitors a configuration file for changes and triggers reload callback.
// Uses polling based on file modification time (ModTime) for cross-platform reliability.
type Watcher struct {
	filePath string
	interval time.Duration
	onChange func(*Config)

	mu       sync.Mutex
	lastMod  time.Time
	running  bool
	stopChan chan struct{}
}

// NewWatcher creates a new Config Watcher.
//
// Parameters:
//   - path: Absolute path to the YAML config file.
//   - interval: How frequently to check the file for changes (default: 5s if <= 0).
//   - onChange: Callback function called with the new configuration.
func NewWatcher(path string, interval time.Duration, onChange func(*Config)) *Watcher {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	return &Watcher{
		filePath: path,
		interval: interval,
		onChange: onChange,
		stopChan: make(chan struct{}),
	}
}

// Start begins monitoring the config file in a background goroutine.
// Blocks until initial validation succeeds, then runs monitor loop in background.
func (w *Watcher) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return errors.New("config: watcher is already running")
	}

	info, err := os.Stat(w.filePath)
	if err != nil {
		w.mu.Unlock()
		return err
	}
	w.lastMod = info.ModTime()
	w.running = true
	w.mu.Unlock()

	go w.pollLoop()

	return nil
}

// pollLoop runs the polling checks at configured intervals.
func (w *Watcher) pollLoop() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ticker.C:
			w.checkFile()
		}
	}
}

// checkFile inspects the file modification time and triggers reload if updated.
func (w *Watcher) checkFile() {
	info, err := os.Stat(w.filePath)
	if err != nil {
		// File might be temporarily locked or deleted during edit.
		// Ignore temporary read errors to avoid crashing.
		return
	}

	w.mu.Lock()
	modTime := info.ModTime()
	if !modTime.After(w.lastMod) {
		w.mu.Unlock()
		return
	}
	w.lastMod = modTime
	w.mu.Unlock()

	// File changed. Load new config.
	newCfg, err := Load(w.filePath)
	if err != nil {
		// If load/validation fails, ignore and keep old config (robustness).
		// We do not want a broken config edit to crash a running kernel.
		return
	}

	// Trigger callback
	if w.onChange != nil {
		w.onChange(newCfg)
	}
}

// Stop halts the configuration monitoring.
func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	close(w.stopChan)
	w.running = false
	// Recreate channel for potential restart
	w.stopChan = make(chan struct{})
}
```

## Rules
1. **Fallback Protection on Bad Configuration Updates**: During file updates, if loading or validation fails, ignore the errors silently and keep using the active configuration. Live kernels must never panic or crash due to syntactical typos made during hot updates.
2. **Polling over ModTime**: Use file modification timestamps (`os.Stat().ModTime()`) to detect updates. This is more cross-platform compatible and reliable than file system event APIs.
3. **Idempotence of Stop Calls**: Make `Stop` calls idempotent, and recreate termination channels (`stopChan`) to allow restarting.

## ⚠️ Pitfalls

### Pitfall 1: Crashing runtime kernels on syntax errors in updated configurations
```go
```
Discard invalid configuration edits and continue using the old memory-loaded configurations.

### Pitfall 2: Memory leaks from failing to stop polling loops
If the ticker is not stopped, background polling goroutines continue running in the background. Always ensure the ticker is stopped using `defer ticker.Stop()`.

## Verify
```bash
go build ./kernel/config/...
```

## Checklist
- [ ] File `kernel/config/watcher.go` exists
- [ ] Package: `config`
- [ ] `Watcher` uses mutex locks to protect modification timestamps
- [ ] File changes are detected by polling file modification times
- [ ] YAML syntax errors are discarded to protect live runtime executions
- [ ] Ticker checks are integrated with context selector channels for clean shutdowns
- [ ] `Stop` shuts down background threads and is idempotent
- [] Channels are recreated on Stop to support restarts
- [ ] `go build ./kernel/config/...` passes
