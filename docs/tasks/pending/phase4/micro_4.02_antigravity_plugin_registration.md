# Micro-Task 4.02: Create plugins/providers/antigravity/plugin.go

## Info
- **File**: `plugins/providers/antigravity/plugin.go`
- **Package**: `antigravity`
- **Depends on**: 4.01
- **Time**: 15 min
- **Verify**: `go build ./plugins/providers/antigravity/...`

## Purpose
Implements the Plugin loader interfaces for the Antigravity provider package, wrapping the runtime initialization and delegating state methods to the base providers.

## EXACT code to create

```go
// Package antigravity implements the Antigravity CLI adapter provider plugin.
package antigravity

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	cplugin "github.com/tiendat1751998/orchestrator/contracts/plugin"
	cprovider "github.com/tiendat1751998/orchestrator/contracts/provider"
	sdkprovider "github.com/tiendat1751998/orchestrator/sdk/provider"
)

// AntigravityProvider adapts the Antigravity CLI model runner into a contracts-compliant provider.
type AntigravityProvider struct {
	*sdkprovider.BaseProvider

	logger *slog.Logger
	binary string
}

// NewAntigravityProvider constructs a new AntigravityProvider instance.
func NewAntigravityProvider(cfg *cprovider.Config, logger *slog.Logger) (*AntigravityProvider, error) {
	if cfg == nil {
		return nil, errors.New("antigravity: configuration cannot be nil")
	}

	// Supported default models for Antigravity CLI
	defaultModels := []string{"gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.0-flash"}

	baseProvider, err := sdkprovider.NewBaseProvider(cfg, defaultModels, logger)
	if err != nil {
		return nil, fmt.Errorf("antigravity: failed to initialize base provider: %w", err)
	}

	// Extract custom configuration parameters
	binaryPath := cfg.Binary
	if binaryPath == "" {
		binaryPath = "antigravity" // Fallback default path search
	}

	return &AntigravityProvider{
		BaseProvider: baseProvider,
		logger:       logger,
		binary:       binaryPath,
	}, nil
}

// Init implements contracts/plugin.Plugin interface.
func (p *AntigravityProvider) Init(ctx context.Context, config map[string]any) error {
	// Call base provider to perform default state validations
	if err := p.BaseProvider.Init(ctx, config); err != nil {
		return err
	}

	if p.logger != nil {
		p.logger.Info("antigravity provider initialized", "binary", p.binary)
	}
	return nil
}

// Start implements contracts/plugin.Plugin interface.
func (p *AntigravityProvider) Start(ctx context.Context) error {
	if err := p.BaseProvider.Start(ctx); err != nil {
		return err
	}

	if p.logger != nil {
		p.logger.Info("antigravity provider started")
	}
	return nil
}

// Stop implements contracts/plugin.Plugin interface.
func (p *AntigravityProvider) Stop(ctx context.Context) error {
	if err := p.BaseProvider.Stop(ctx); err != nil {
		return err
	}

	if p.logger != nil {
		p.logger.Info("antigravity provider stopped")
	}
	return nil
}

// Health implements contracts/plugin.Plugin interface.
func (p *AntigravityProvider) Health(ctx context.Context) (cplugin.HealthReport, error) {
	report, err := p.BaseProvider.Health(ctx)
	if err != nil {
		return report, err
	}

	// Dynamic check to verify binary availability on health query
	if !p.IsAvailable(ctx) {
		report.Status = cplugin.HealthDown
		report.Message = "antigravity CLI binary not found or not executable"
	}

	return report, nil
}
```

## Pitfalls

### Pitfall 1: Bypassing BaseProvider lifecycle calls
```go
// WRONG:
func (p *AntigravityProvider) Init(ctx context.Context, config map[string]any) error {
    // Custom logic only
    return nil // Skips initialization locks of BaseProvider, causing races!
}

// CORRECT:
func (p *AntigravityProvider) Init(ctx context.Context, config map[string]any) error {
    if err := p.BaseProvider.Init(ctx, config); err != nil {
        return err
    }
    return nil
}
```
If you override lifecycle methods (`Init`, `Start`, `Stop`) without calling the embedded `BaseProvider` counterpart methods, you break status flags (`IsStarted()`, `IsInitialized()`), causing runtime failures.

### Pitfall 2: Silent failures from missing binary configs
If `binary` is missing and no fallback is set, executing CLI processes will fail later during task runs. Provide default executable searches.

## Verify
```bash
go build ./plugins/providers/antigravity/...
```

## Checklist
- [ ] File exists at `plugins/providers/antigravity/plugin.go`
- [ ] Package name is `antigravity`
- [ ] All exported types have Godoc
- [ ] `AntigravityProvider` embeds `sdkprovider.BaseProvider`
- [ ] Override methods call parent `BaseProvider` lifecycles
- [ ] `Health()` checks binary execution availability
- [ ] Build command passes
