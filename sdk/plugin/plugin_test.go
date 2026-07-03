package plugin

import (
	"context"
	"sync"
	"testing"

	cplugin "github.com/tiendat1751998/orchestrator/contracts/plugin"
)

func TestNewBasePlugin(t *testing.T) {
	t.Run("empty name returns error", func(t *testing.T) {
		p, err := NewBasePlugin("", cplugin.TypeAgent, "1.2.3")
		if err == nil {
			t.Error("expected error for empty plugin name, got nil")
		}
		if p != nil {
			t.Errorf("expected nil plugin, got %v", p)
		}
	})

	t.Run("empty version defaults to 1.0.0", func(t *testing.T) {
		p, err := NewBasePlugin("test-plugin", cplugin.TypeAgent, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Version() != "1.0.0" {
			t.Errorf("expected version to default to '1.0.0', got '%s'", p.Version())
		}
	})

	t.Run("valid parameters initialized correctly", func(t *testing.T) {
		name := "test-plugin"
		pType := cplugin.TypeAgent
		ver := "1.2.3"
		p, err := NewBasePlugin(name, pType, ver)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name() != name {
			t.Errorf("expected Name() = %s, got %s", name, p.Name())
		}
		if p.Type() != pType {
			t.Errorf("expected Type() = %v, got %v", pType, p.Type())
		}
		if p.Version() != ver {
			t.Errorf("expected Version() = %s, got %s", ver, p.Version())
		}
		if p.IsInitialized() {
			t.Error("expected plugin not to be initialized initially")
		}
		if p.IsStarted() {
			t.Error("expected plugin not to be started initially")
		}
	})
}

func TestBasePluginLifecycle(t *testing.T) {
	ctx := context.Background()
	p, err := NewBasePlugin("test-plugin", cplugin.TypeAgent, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify initial health report is DOWN (not initialized)
	report, err := p.Health(ctx)
	if err != nil {
		t.Fatalf("unexpected error from Health: %v", err)
	}
	if report.Status != cplugin.HealthDown {
		t.Errorf("expected Health status to be Down, got %s", report.Status)
	}
	if report.Message != "plugin not initialized" {
		t.Errorf("expected Health message to be 'plugin not initialized', got '%s'", report.Message)
	}

	// Trying to start before init should fail
	if err := p.Start(ctx); err == nil {
		t.Error("expected error starting uninitialized plugin, got nil")
	}

	// Initialize plugin
	if err := p.Init(ctx, nil); err != nil {
		t.Fatalf("unexpected error initializing: %v", err)
	}
	if !p.IsInitialized() {
		t.Error("expected IsInitialized() to be true")
	}

	// Re-initializing should fail
	if err := p.Init(ctx, nil); err == nil {
		t.Error("expected error on re-initialization, got nil")
	}

	// Verify health report is DOWN (initialized but not started)
	report, err = p.Health(ctx)
	if err != nil {
		t.Fatalf("unexpected error from Health: %v", err)
	}
	if report.Status != cplugin.HealthDown {
		t.Errorf("expected Health status to be Down, got %s", report.Status)
	}
	if report.Message != "plugin initialized but not started" {
		t.Errorf("expected Health message to be 'plugin initialized but not started', got '%s'", report.Message)
	}

	// Start plugin
	if err := p.Start(ctx); err != nil {
		t.Fatalf("unexpected error starting: %v", err)
	}
	if !p.IsStarted() {
		t.Error("expected IsStarted() to be true")
	}

	// Re-starting should fail
	if err := p.Start(ctx); err == nil {
		t.Error("expected error on re-starting, got nil")
	}

	// Verify health report is OK
	report, err = p.Health(ctx)
	if err != nil {
		t.Fatalf("unexpected error from Health: %v", err)
	}
	if report.Status != cplugin.HealthOK {
		t.Errorf("expected Health status to be OK, got %s", report.Status)
	}

	// Stop plugin
	if err := p.Stop(ctx); err != nil {
		t.Fatalf("unexpected error stopping: %v", err)
	}
	if p.IsStarted() {
		t.Error("expected IsStarted() to be false after Stop()")
	}

	// Idempotency: Stop again should be a no-op and not fail
	if err := p.Stop(ctx); err != nil {
		t.Fatalf("unexpected error on second Stop: %v", err)
	}
}

func TestBasePluginThreadSafety(t *testing.T) {
	ctx := context.Background()
	p, err := NewBasePlugin("test-plugin", cplugin.TypeAgent, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var wg sync.WaitGroup
	workers := 20
	iterations := 100

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// Concurrent reads
				_ = p.Name()
				_ = p.Type()
				_ = p.Version()
				_ = p.IsInitialized()
				_ = p.IsStarted()
				_, _ = p.Health(ctx)

				// Interleaved lifecycle mutations
				if j%10 == 0 {
					_ = p.Init(ctx, nil)
				}
				if j%10 == 3 {
					_ = p.Start(ctx)
				}
				if j%10 == 7 {
					_ = p.Stop(ctx)
				}
			}
		}(i)
	}

	wg.Wait()
}
