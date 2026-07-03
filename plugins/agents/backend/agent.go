// Package backend implements the Backend developer agent plugin.
package backend

import (
	"context"
	"fmt"
	"log/slog"

	cplugin "github.com/tiendat1751998/orchestrator/contracts/plugin"
	"github.com/tiendat1751998/orchestrator/contracts/provider"
	sdkagent "github.com/tiendat1751998/orchestrator/sdk/agent"
)

// BackendAgent handles software development tasks.
// Extends BaseAgent. Thread-safe.
type BackendAgent struct {
	*sdkagent.BaseAgent
}

// NewBackendAgent constructs a BackendAgent.
func NewBackendAgent(manifestPath string, p provider.Provider, logger *slog.Logger) (*BackendAgent, error) {
	manifest, err := sdkagent.LoadManifest(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("backend: failed to load manifest: %w", err)
	}

	baseAgent, err := sdkagent.NewBaseAgent(manifest, p, logger)
	if err != nil {
		return nil, fmt.Errorf("backend: failed to construct base agent: %w", err)
	}

	return &BackendAgent{
		BaseAgent: baseAgent,
	}, nil
}

// Init implements contracts/plugin.Plugin interface.
func (a *BackendAgent) Init(ctx context.Context, config map[string]any) error {
	return a.BaseAgent.Init(ctx, config)
}

// Start implements contracts/plugin.Plugin interface.
func (a *BackendAgent) Start(ctx context.Context) error {
	return a.BaseAgent.Start(ctx)
}

// Stop implements contracts/plugin.Plugin interface.
func (a *BackendAgent) Stop(ctx context.Context) error {
	return a.BaseAgent.Stop(ctx)
}

// Health implements contracts/plugin.Plugin interface.
func (a *BackendAgent) Health(ctx context.Context) (cplugin.HealthReport, error) {
	return a.BaseAgent.Health(ctx)
}
