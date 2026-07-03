// Package reviewer implements the Code reviewer agent plugin.
package reviewer

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	cplugin "github.com/tiendat1751998/orchestrator/contracts/plugin"
	"github.com/tiendat1751998/orchestrator/contracts/provider"
	sdkagent "github.com/tiendat1751998/orchestrator/sdk/agent"
)

// ReviewerAgent handles audits and code reviews.
// Extends BaseAgent. Thread-safe.
type ReviewerAgent struct {
	*sdkagent.BaseAgent
}

// NewReviewerAgent constructs a ReviewerAgent.
func NewReviewerAgent(manifestPath string, p provider.Provider, logger *slog.Logger) (*ReviewerAgent, error) {
	if manifestPath == "" {
		return nil, fmt.Errorf("reviewer: manifest path cannot be empty")
	}

	// Verify constructors parse paths using standard filepath resolution to avoid absolute mapping crashes.
	resolvedPath := filepath.Clean(manifestPath)

	manifest, err := sdkagent.LoadManifest(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("reviewer: failed to load manifest: %w", err)
	}

	baseAgent, err := sdkagent.NewBaseAgent(manifest, p, logger)
	if err != nil {
		return nil, fmt.Errorf("reviewer: failed to construct base agent: %w", err)
	}

	return &ReviewerAgent{
		BaseAgent: baseAgent,
	}, nil
}

// Init implements contracts/plugin.Plugin interface.
func (a *ReviewerAgent) Init(ctx context.Context, config map[string]any) error {
	return a.BaseAgent.Init(ctx, config)
}

// Start implements contracts/plugin.Plugin interface.
func (a *ReviewerAgent) Start(ctx context.Context) error {
	return a.BaseAgent.Start(ctx)
}

// Stop implements contracts/plugin.Plugin interface.
func (a *ReviewerAgent) Stop(ctx context.Context) error {
	return a.BaseAgent.Stop(ctx)
}

// Health implements contracts/plugin.Plugin interface.
func (a *ReviewerAgent) Health(ctx context.Context) (cplugin.HealthReport, error) {
	return a.BaseAgent.Health(ctx)
}
