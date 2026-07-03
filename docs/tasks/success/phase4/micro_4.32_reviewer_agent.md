# Micro-Task 4.32: Create plugins/agents/reviewer/agent.go

## Info
- **File**: `plugins/agents/reviewer/agent.go`
- **Package**: `reviewer`
- **Depends on**: 4.31
- **Time**: 15 min
- **Verify**: `go build ./plugins/agents/reviewer/...`

## Purpose
Implements the Code Reviewer agent package constructor, loading YAML configurations and wrapping the SDK BaseAgent.

## EXACT code to create

```go
// Package reviewer implements the Code reviewer agent plugin.
package reviewer

import (
	"context"
	"fmt"
	"log/slog"

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
	manifest, err := sdkagent.LoadManifest(manifestPath)
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
```

## Pitfalls

### Pitfall 1: Bypassing manifest path validation checks
Verify constructors parse paths using standard filepath resolution to avoid absolute mapping crashes.

### Pitfall 2: Silent failures when executing delegate calls
Lifecycle functions must wrap equivalent parent methods to avoid registration failures.

## Verify
```bash
go build ./plugins/agents/reviewer/...
```

## Checklist
- [ ] File exists at `plugins/agents/reviewer/agent.go`
- [ ] Package name is `reviewer`
- [ ] All exported types have Godoc
- [ ] `ReviewerAgent` embeds `sdkagent.BaseAgent`
- [ ] Build command passes
