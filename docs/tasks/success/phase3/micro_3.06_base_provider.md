# Micro-Task 3.06: Create sdk/provider/provider.go

## Info
- **File**: `sdk/provider/provider.go`
- **Package**: `provider`
- **Depends on**: 3.01 (base_plugin.md), 1.11 (provider config contract), 1.12 (provider interface contract)
- **Time**: 15 min
- **Verify**: `go build ./sdk/provider/...`

## Purpose
Implements the reusable, thread-safe base helper struct (`BaseProvider` and constructors) for AI model provider integrations. It reduces boilerplate for writing new provider plugins by providing default lifecycle setups, helper accessors, and models list management.

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
func (bp *BaseProvider) Models(ctx context.Context) ([]string, error) {
	if !bp.IsStarted() {
		return nil, fmt.Errorf("sdk/provider: provider %q is not running", bp.Name())
	}
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	copied := make([]string, len(bp.models))
	copy(copied, bp.models)
	return copied, nil
}

// UpdateModels replaces the active list of supported models.
func (bp *BaseProvider) UpdateModels(models []string) {
	bp.mu.Lock()
	bp.models = make([]string, 0, len(models))
	for _, m := range models {
		if m != "" {
			bp.models = append(bp.models, m)
		}
	}
	bp.mu.Unlock()
}

// IsAvailable performs a baseline check to see if the provider plugin is operational.
func (bp *BaseProvider) IsAvailable(ctx context.Context) bool {
	return bp.IsStarted()
}

// Timeout returns the duration wait configured for operations.
func (bp *BaseProvider) Timeout() time.Duration {
	return bp.cfg.TimeoutOrDefault()
}
```

## Rules
1. **Prevent Shared Memory Mutations**: Copy slice buffers before returning them inside `Models()` calls to avoid shared memory mutation issues outside lock boundaries.
2. **Access Locks**: Methods that read or update internal lists (such as `Models` and `UpdateModels`) must synchronize using `sync.RWMutex`.
3. **Guard Startup States**: Reject method invocations with errors if they are executed before the provider has started.

## ⚠️ Pitfalls

### Pitfall 1: Returning references to internal slices directly
Returning slices directly (e.g., `return bp.models, nil`) allows callers to modify the elements of the internal slice without locking, causing data races. Always return copies instead.

### Pitfall 2: Bypassing startup status checks
Querying provider data structures (like listing models) when the provider has stopped can return outdated or empty values. Always verify `IsStarted()` beforehand.

## Verify
```bash
go build ./sdk/provider/...
```

## Checklist
- [ ] File `sdk/provider/provider.go` exists
- [ ] Package: `provider`
- [ ] `BaseProvider` embeds `*sdkplugin.BasePlugin`
- [ ] `Models` returns copies of slice lists to prevent external edits
- [ ] `UpdateModels` and `Models` are protected by `sync.RWMutex` locks
- [ ] `IsAvailable` performs basic running state checks
- [ ] Provider configuration exposes `Timeout` defaults
- [ ] `go build ./sdk/provider/...` passes
