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
