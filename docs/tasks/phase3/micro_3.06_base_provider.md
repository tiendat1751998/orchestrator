# Micro-Task 3.06: Create sdk/provider/provider.go

## Info
- **File**: `sdk/provider/provider.go`
- **Package**: `provider`
- **Depends on**: 3.01 (base_plugin.md), 1.11 (provider config contract), 1.12 (provider interface contract)
- **Time**: 15 min
- **Verify**: `go build ./sdk/provider/...`

## Purpose
Implements the reusable, thread-safe base helper struct (`BaseProvider`) for AI model provider integrations. It reduces boilerplate for writing new provider plugins by providing default lifecycle setups, helper accessors, and models list management.

## EXACT code to create

```go
// Package provider provides base helpers and request builders for AI provider integrations.
package provider

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	contractsplugin "github.com/tiendat1751998/orchestrator/contracts/plugin"
	contractsprovider "github.com/tiendat1751998/orchestrator/contracts/provider"
	sdkplugin "github.com/tiendat1751998/orchestrator/sdk/plugin"
)

// BaseProvider implements core shared methods of contractsprovider.Provider interface.
// Custom providers embed this struct to inherit default implementations of Name, IsAvailable, and Models.
type BaseProvider struct {
	*sdkplugin.BasePlugin

	logger *slog.Logger
	cfg    *contractsprovider.Config
	
	mu     sync.RWMutex
	models []string
}

// NewBaseProvider constructs a BaseProvider.
//
// Parameters:
//   - cfg: provider configuration struct from YAML.
//   - models: list of model strings supported by the provider (e.g. ["gemini-2.5-pro"]).
//   - logger: logger instance.
func NewBaseProvider(cfg *contractsprovider.Config, models []string, logger *slog.Logger) (*BaseProvider, error) {
	if cfg == nil {
		return nil, errors.New("sdk/provider: configuration cannot be nil")
	}
	if cfg.Name == "" {
		return nil, errors.New("sdk/provider: provider name cannot be empty")
	}

	basePlugin, err := sdkplugin.NewBasePlugin(cfg.Name, contractsplugin.TypeProvider, "1.0.0")
	if err != nil {
		return nil, err
	}

	// Clean models list
	cleanModels := make([]string, 0, len(models))
	for _, m := range models {
		if m != "" {
			cleanModels = append(cleanModels, m)
		}
	}

	return &BaseProvider{
		BasePlugin: basePlugin,
		cfg:        cfg,
		models:     cleanModels,
		logger:     logger,
	}, nil
}

// Config returns the read-only copy of the provider's configuration.
func (bp *BaseProvider) Config() contractsprovider.Config {
	return *bp.cfg
}

// Name returns the unique provider identifier.
func (bp *BaseProvider) Name() string {
	return bp.cfg.Name
}

// Models returns the list of models supported by this provider.
// Safe for concurrent use.
func (bp *BaseProvider) Models(ctx context.Context) ([]string, error) {
	if !bp.IsStarted() {
		return nil, fmt.Errorf("sdk/provider: provider %q is not running", bp.Name())
	}
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	// Return a copy to prevent external mutation
	copied := make([]string, len(bp.models))
	copy(copied, bp.models)
	return copied, nil
}

// UpdateModels replaces the active list of supported models.
// Thread-safe. Called when provider discovers models dynamically from API.
func (bp *BaseProvider) UpdateModels(models []string) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	cleanModels := make([]string, 0, len(models))
	for _, m := range models {
		if m != "" {
			cleanModels = append(cleanModels, m)
		}
	}
	bp.models = cleanModels
}

// IsAvailable performs a baseline check to see if the provider plugin is operational.
// Concrete implementations (like CLI providers) should override this method to test files or ping endpoints.
func (bp *BaseProvider) IsAvailable(ctx context.Context) bool {
	// Base check: is the plugin initialized and started?
	return bp.IsStarted()
}

// Timeout returns the duration wait configured for operations.
func (bp *BaseProvider) Timeout() time.Duration {
	return bp.cfg.TimeoutOrDefault()
}
```

## ⚠️ Pitfalls

### Pitfall 1: Returning internal slice reference directly
```go
// ❌ WRONG:
func (bp *BaseProvider) Models(ctx context.Context) ([]string, error) {
    return bp.models, nil // Return directly -> caller can mutate the slice elements
}

// ✅ CORRECT:
func (bp *BaseProvider) Models(ctx context.Context) ([]string, error) {
    bp.mu.RLock()
    defer bp.mu.RUnlock()
    copied := make([]string, len(bp.models))
    copy(copied, bp.models)
    return copied, nil
}
```
Slice references share memory under the hood. Always copy the slice before returning it from thread-safe boundaries.

### Pitfall 2: No check for running state in method calls
Calling `Models(ctx)` on a stopped or uninitialized provider must fail fast instead of returning outdated data. Verify `bp.IsStarted()` at method entry.

## Verify
```bash
go build ./sdk/provider/...
```

## Checklist
- [ ] File `sdk/provider/provider.go` exists
- [ ] Package: `provider`
- [ ] `BaseProvider` embeds `*sdkplugin.BasePlugin`
- [ ] `Models` returns a copied slice to prevent external writes
- [ ] Thread safety (sync.RWMutex) on `UpdateModels` and `Models` functions
- [ ] Availability baseline checks `IsStarted()` state
- [ ] Cấu hình `TimeoutOrDefault` được bộc lộ thông qua method `Timeout`
- [ ] `go build ./sdk/provider/...` thành công
