# Micro-Task 4.16: Create plugins/providers/antigravity/provider_test.go

## Info
- **File**: `plugins/providers/antigravity/provider_test.go`
- **Package**: `antigravity_test`
- **Depends on**: 4.14, 4.15
- **Time**: 20 min
- **Verify**: `go test -v -race -count=1 ./plugins/providers/antigravity/...`

## Purpose
Implements integration unit tests for the Antigravity and Gemini provider drivers, verifying process calls, parsers, and connection sessions.

## EXACT code to create

```go
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
		Name: "antigravity-test",
		Settings: map[string]any{
			"binary": "go", // Use a guaranteed standard CLI binary for lifecycle checks
		},
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
		Settings: map[string]any{
			// Empty key settings
		},
	}

	_, err := antigravity.NewGeminiProvider(cfg)
	if err == nil {
		t.Error("expected error due to missing api_key configuration setting, got nil")
	}
}

func TestGeminiProvider_SuccessConstruction(t *testing.T) {
	cfg := &provider.Config{
		Name: "gemini-test",
		Settings: map[string]any{
			"api_key": "dummy-key-value",
		},
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
```

## Pitfalls

### Pitfall 1: Requiring real API keys for unit tests
Running tests that attempt to connect to live Gemini server endpoints without active API keys causes them to fail. Isolate tests from live APIs by using dummy configurations to test initialization, and wrap network operations.

### Pitfall 2: Reusing processes without resetting manager loops
If test tasks leak CLI sessions, they can trigger resource limits in subsequent tests. Always execute `p.Stop()` in defer blocks.

## Verify
```bash
go test -v -race -count=1 ./plugins/providers/antigravity/...
```

## Checklist
- [ ] File exists at `plugins/providers/antigravity/provider_test.go`
- [ ] Package name is `antigravity_test`
- [ ] All exported types have Godoc
- [ ] Lifecycle transitions execute successfully
- [ ] Constructor rejects configurations missing API keys
- [ ] Build command passes
