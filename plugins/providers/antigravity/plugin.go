// Package antigravity implements the Antigravity CLI adapter provider plugin.
package antigravity

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	cplugin "github.com/tiendat1751998/orchestrator/contracts/plugin"
	cprovider "github.com/tiendat1751998/orchestrator/contracts/provider"
	"github.com/tiendat1751998/orchestrator/plugins/providers/antigravity/session"
	sdkprovider "github.com/tiendat1751998/orchestrator/sdk/provider"
)

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

	// 5 maximum concurrent CLI processes, 5 minutes idle timeout
	sm := session.NewSessionManager(binaryPath, 5, 5*time.Minute)

	return &AntigravityProvider{
		BaseProvider: baseProvider,
		logger:       logger,
		binary:       binaryPath,
		sm:           sm,
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
	var stopErr error
	if p.sm != nil {
		stopErr = p.sm.Stop()
	}

	if err := p.BaseProvider.Stop(ctx); err != nil {
		return err
	}

	if p.logger != nil {
		p.logger.Info("antigravity provider stopped")
	}
	return stopErr
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
