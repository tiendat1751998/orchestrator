package config

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp YAML: %v", err)
	}
	return path
}

func updateTempYAML(t *testing.T, path string, content string, offset time.Duration) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp YAML: %v", err)
	}
	newMod := time.Now().Add(offset)
	if err := os.Chtimes(path, newMod, newMod); err != nil {
		t.Fatalf("chtimes: %v", err)
	}
}

func TestWatcher_DetectionAndCallback(t *testing.T) {
	initialYAML := `
orchestrator:
  name: "initial-name"
  log_level: "info"
  log_format: "text"
  data_dir: "./data"
  max_concurrent_tasks: 5
  shutdown_timeout: "30s"
providers:
  default: "prov"
  configs:
    prov:
      type: "local"
      model: "gpt-4"
      timeout: "30s"
      max_retry: 3
`
	path := writeTempYAML(t, initialYAML)

	var mu sync.Mutex
	var lastConfig *Config
	callbackCount := 0

	onChange := func(cfg *Config) {
		mu.Lock()
		defer mu.Unlock()
		lastConfig = cfg
		callbackCount++
	}

	watcher := NewWatcher(path, 10*time.Millisecond, onChange)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := watcher.Start(ctx); err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}
	defer watcher.Stop()

	watcher.mu.Lock()
	initialMod := watcher.lastMod
	watcher.mu.Unlock()
	if initialMod.IsZero() {
		t.Errorf("expected lastMod to be set on start")
	}

	validYAML := `
orchestrator:
  name: "updated-name"
  log_level: "debug"
  log_format: "text"
  data_dir: "./data"
  max_concurrent_tasks: 5
  shutdown_timeout: "30s"
providers:
  default: "prov"
  configs:
    prov:
      type: "local"
      model: "gpt-4"
      timeout: "30s"
      max_retry: 3
`
	updateTempYAML(t, path, validYAML, 5*time.Second)

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if callbackCount != 1 {
		t.Errorf("expected callback to be called once, got %d", callbackCount)
	}
	if lastConfig == nil || lastConfig.Orchestrator.Name != "updated-name" {
		t.Errorf("expected config name to be 'updated-name', got: %+v", lastConfig)
	}
	mu.Unlock()
}

func TestWatcher_InvalidUpdates(t *testing.T) {
	initialYAML := `
orchestrator:
  name: "initial-name"
  log_level: "info"
  log_format: "text"
  data_dir: "./data"
  max_concurrent_tasks: 5
  shutdown_timeout: "30s"
providers:
  default: "prov"
  configs:
    prov:
      type: "local"
      model: "gpt-4"
      timeout: "30s"
      max_retry: 3
`
	path := writeTempYAML(t, initialYAML)

	var mu sync.Mutex
	var lastConfig *Config
	callbackCount := 0

	onChange := func(cfg *Config) {
		mu.Lock()
		defer mu.Unlock()
		lastConfig = cfg
		callbackCount++
	}

	watcher := NewWatcher(path, 10*time.Millisecond, onChange)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := watcher.Start(ctx); err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}
	defer watcher.Stop()

	invalidYAML := `
orchestrator:
  name: "bad-yaml"
  log_level: [unclosed brackets
`
	updateTempYAML(t, path, invalidYAML, 5*time.Second)
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if callbackCount != 0 {
		t.Errorf("callback should not be called for invalid YAML, count: %d", callbackCount)
	}
	mu.Unlock()

	invalidConfigYAML := `
orchestrator:
  name: "invalid-config"
  log_level: "super-debug"
  log_format: "text"
  data_dir: "./data"
  max_concurrent_tasks: 5
  shutdown_timeout: "30s"
providers:
  default: "prov"
  configs:
    prov:
      type: "local"
      model: "gpt-4"
      timeout: "30s"
      max_retry: 3
`
	updateTempYAML(t, path, invalidConfigYAML, 10*time.Second)
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if callbackCount != 0 {
		t.Errorf("callback should not be called for validation failure, count: %d", callbackCount)
	}
	mu.Unlock()

	validYAML := `
orchestrator:
  name: "recovered-name"
  log_level: "debug"
  log_format: "text"
  data_dir: "./data"
  max_concurrent_tasks: 5
  shutdown_timeout: "30s"
providers:
  default: "prov"
  configs:
    prov:
      type: "local"
      model: "gpt-4"
      timeout: "30s"
      max_retry: 3
`
	updateTempYAML(t, path, validYAML, 15*time.Second)
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if callbackCount != 1 {
		t.Errorf("expected callback to be called once after recovery, got %d", callbackCount)
	}
	if lastConfig == nil || lastConfig.Orchestrator.Name != "recovered-name" {
		t.Errorf("expected config name to be 'recovered-name', got: %+v", lastConfig)
	}
	mu.Unlock()
}

func TestWatcher_IdempotenceAndRestart(t *testing.T) {
	initialYAML := `
orchestrator:
  name: "idempotency-test"
  log_level: "info"
  log_format: "text"
  data_dir: "./data"
  max_concurrent_tasks: 5
  shutdown_timeout: "30s"
providers:
  default: "prov"
  configs:
    prov:
      type: "local"
      model: "gpt-4"
      timeout: "30s"
      max_retry: 3
`
	path := writeTempYAML(t, initialYAML)

	watcher := NewWatcher(path, 10*time.Millisecond, nil)

	ctx := context.Background()

	watcher.Stop()
	watcher.Stop()

	if err := watcher.Start(ctx); err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}

	if err := watcher.Start(ctx); err == nil {
		t.Errorf("expected error starting already running watcher, got nil")
	}

	watcher.Stop()
	watcher.Stop()

	if err := watcher.Start(ctx); err != nil {
		t.Fatalf("failed to restart watcher: %v", err)
	}

	watcher.Stop()
}

func TestWatcher_ContextShutdown(t *testing.T) {
	initialYAML := `
orchestrator:
  name: "ctx-test"
  log_level: "info"
  log_format: "text"
  data_dir: "./data"
  max_concurrent_tasks: 5
  shutdown_timeout: "30s"
providers:
  default: "prov"
  configs:
    prov:
      type: "local"
      model: "gpt-4"
      timeout: "30s"
      max_retry: 3
`
	path := writeTempYAML(t, initialYAML)

	watcher := NewWatcher(path, 10*time.Millisecond, nil)

	ctx, cancel := context.WithCancel(context.Background())

	if err := watcher.Start(ctx); err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}

	watcher.mu.Lock()
	runningBefore := watcher.running
	watcher.mu.Unlock()

	if !runningBefore {
		t.Errorf("expected watcher to be running")
	}

	cancel()

	time.Sleep(50 * time.Millisecond)

	watcher.mu.Lock()
	runningAfter := watcher.running
	watcher.mu.Unlock()

	if runningAfter {
		t.Errorf("expected watcher to stop running after context cancellation")
	}
}
