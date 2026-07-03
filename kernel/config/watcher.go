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
	stopChan := w.stopChan
	w.mu.Unlock()

	go w.pollLoop(ctx, stopChan)

	return nil
}

// pollLoop runs the polling checks at configured intervals.
// ponytail: polling ModTime is cross-platform reliable but has polling latency (interval). Upgrade path: fsnotify/inotify if immediate hot-reload is required for large scale deployment.
func (w *Watcher) pollLoop(ctx context.Context, stopChan chan struct{}) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ctx.Done():
			w.Stop()
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

	// Validate config before triggering callback
	if err := Validate(newCfg); err != nil {
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
