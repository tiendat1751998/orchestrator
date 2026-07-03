package kernel

import (
	"context"
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/plugin"
	"github.com/tiendat1751998/orchestrator/kernel/config"
)

type mockPlugin struct {
	name        string
	pluginType  plugin.Type
	version     string
	initCalled  bool
	startCalled bool
	stopCalled  bool
	initErr     error
	startErr    error
	stopErr     error
}

func (m *mockPlugin) Name() string      { return m.name }
func (m *mockPlugin) Type() plugin.Type { return m.pluginType }
func (m *mockPlugin) Version() string   { return m.version }

func (m *mockPlugin) Init(_ context.Context, _ map[string]any) error {
	m.initCalled = true
	return m.initErr
}

func (m *mockPlugin) Start(_ context.Context) error {
	m.startCalled = true
	return m.startErr
}

func (m *mockPlugin) Stop(_ context.Context) error {
	m.stopCalled = true
	return m.stopErr
}

func (m *mockPlugin) Health(_ context.Context) (plugin.HealthReport, error) {
	return plugin.HealthReport{Status: plugin.HealthOK}, nil
}

func TestNewKernel(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		k, err := New(nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if k != nil {
			t.Errorf("expected kernel to be nil, got %v", k)
		}
	})

	t.Run("valid config", func(t *testing.T) {
		cfg := config.DefaultConfig()
		k, err := New(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if k == nil {
			t.Fatal("expected kernel, got nil")
		}
		if k.State() != StateCreated {
			t.Errorf("expected state to be StateCreated, got %v", k.State())
		}
		if k.Config() != cfg {
			t.Errorf("expected config to be %v, got %v", cfg, k.Config())
		}
		if k.Logger() == nil {
			t.Error("expected logger to be initialized")
		}
		if k.EventBus() == nil {
			t.Error("expected eventbus to be initialized")
		}
		if k.Registry() == nil {
			t.Error("expected registry to be initialized")
		}
	})
}

func TestKernelLifecycle(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Orchestrator.Name = "test-kernel"
	cfg.Orchestrator.DataDir = t.TempDir()

	k, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create kernel: %v", err)
	}

	p := &mockPlugin{
		name:       "test-plugin",
		pluginType: plugin.TypeSearch,
		version:    "1.0.0",
	}

	// Register plugin before Start should succeed
	if err := k.RegisterPlugin(p); err != nil {
		t.Fatalf("failed to register plugin: %v", err)
	}

	// Double start should fail transition
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := k.Start(ctx); err != nil {
		t.Fatalf("failed to start kernel: %v", err)
	}

	if k.State() != StateRunning {
		t.Errorf("expected state to be Running, got %v", k.State())
	}

	// Registering plugin after start should fail
	p2 := &mockPlugin{
		name:       "test-plugin-2",
		pluginType: plugin.TypeSearch,
		version:    "1.0.0",
	}
	if err := k.RegisterPlugin(p2); err == nil {
		t.Error("expected error registering plugin after start, got nil")
	}

	// Check if plugin methods were called
	if !p.initCalled {
		t.Error("expected plugin Init to be called")
	}
	if !p.startCalled {
		t.Error("expected plugin Start to be called")
	}

	// Stop the kernel
	if err := k.Stop(ctx); err != nil {
		t.Fatalf("failed to stop kernel: %v", err)
	}

	if k.State() != StateStopped {
		t.Errorf("expected state to be Stopped, got %v", k.State())
	}

	if !p.stopCalled {
		t.Error("expected plugin Stop to be called")
	}

	// Stop again should be a no-op (idempotent)
	if err := k.Stop(ctx); err != nil {
		t.Errorf("expected second stop to be no-op, got error: %v", err)
	}
}

func TestKernelStartFailure(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Orchestrator.Name = "test-kernel-fail"
	cfg.Orchestrator.DataDir = t.TempDir()

	k, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create kernel: %v", err)
	}

	// Force initialization error
	p := &mockPlugin{
		name:       "test-plugin",
		pluginType: plugin.TypeSearch,
		version:    "1.0.0",
		initErr:    context.DeadlineExceeded,
	}

	if err := k.RegisterPlugin(p); err != nil {
		t.Fatalf("failed to register plugin: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := k.Start(ctx); err == nil {
		t.Fatal("expected start error, got nil")
	}

	if k.State() != StateStopped {
		t.Errorf("expected state to be Stopped after failure, got %v", k.State())
	}
}
