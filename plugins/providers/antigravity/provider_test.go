package antigravity_test

import (
	"context"
	"os"
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts/provider"
	"github.com/tiendat1751998/orchestrator/plugins/providers/antigravity"
)

func TestProvider_InitAndLifecycle(t *testing.T) {
	cfg := &provider.Config{
		Name:   "antigravity-test",
		Binary: "go", // Use a guaranteed standard CLI binary for lifecycle checks
	}

	p, err := antigravity.NewProvider(cfg)
	if err != nil {
		t.Fatalf("failed to construct provider: %v", err)
	}

	ctx := context.Background()

	// 1. Verify availability of dummy standard CLI path
	if !p.IsAvailable(ctx) {
		t.Log("Warning: binary search LookPath did not find binary")
	}

	// 2. Test lifecycle state transitions
	err = p.Init(ctx, nil)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	err = p.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	err = p.Stop(ctx)
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestGeminiProvider_FailureOnMissingAPIKey(t *testing.T) {
	cfg := &provider.Config{
		Name: "gemini-test",
	}

	_, err := antigravity.NewGeminiProvider(cfg)
	if err == nil {
		t.Error("expected error due to missing api_key configuration setting, got nil")
	}
}

func TestGeminiProvider_SuccessConstruction(t *testing.T) {
	cfg := &provider.Config{
		Name:   "gemini-test",
		APIKey: "dummy-key-value",
	}

	gp, err := antigravity.NewGeminiProvider(cfg)
	if err != nil {
		t.Fatalf("failed to construct gemini provider: %v", err)
	}

	gp.Init(context.Background(), nil)
	gp.Start(context.Background())
	defer gp.Stop(context.Background())

	if !gp.IsAvailable(context.Background()) {
		t.Error("expected gemini provider to report available when API key is set")
	}
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
