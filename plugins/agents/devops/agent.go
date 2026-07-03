// Package devops implements the DevOps engineer agent plugin.
package devops

import (
	"context"
	"fmt"
	"log/slog"

	cplugin "github.com/tiendat1751998/orchestrator/contracts/plugin"
	"github.com/tiendat1751998/orchestrator/contracts/provider"
	sdkagent "github.com/tiendat1751998/orchestrator/sdk/agent"
)

// DevOpsAgent handles environments and deployments.
// Extends BaseAgent. Thread-safe.
type DevOpsAgent struct {
	*sdkagent.BaseAgent
}

// NewDevOpsAgent constructs a DevOpsAgent.
func NewDevOpsAgent(manifestPath string, p provider.Provider, logger *slog.Logger) (*DevOpsAgent, error) {
	manifest, err := sdkagent.LoadManifest(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("devops: failed to load manifest: %w", err)
	}

	baseAgent, err := sdkagent.NewBaseAgent(manifest, p, logger)
	if err != nil {
		return nil, fmt.Errorf("devops: failed to construct base agent: %w", err)
	}

	return &DevOpsAgent{
		BaseAgent: baseAgent,
	}, nil
}

// Init implements contracts/plugin.Plugin interface.
func (a *DevOpsAgent) Init(ctx context.Context, config map[string]any) error {
	return a.BaseAgent.Init(ctx, config)
}

// Start implements contracts/plugin.Plugin interface.
func (a *DevOpsAgent) Start(ctx context.Context) error {
	return a.BaseAgent.Start(ctx)
}

// Stop implements contracts/plugin.Plugin interface.
func (a *DevOpsAgent) Stop(ctx context.Context) error {
	return a.BaseAgent.Stop(ctx)
}

// Health implements contracts/plugin.Plugin interface.
func (a *DevOpsAgent) Health(ctx context.Context) (cplugin.HealthReport, error) {
	return a.BaseAgent.Health(ctx)
}
