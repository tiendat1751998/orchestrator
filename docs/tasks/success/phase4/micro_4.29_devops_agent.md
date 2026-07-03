# Micro-Task 4.29: Create plugins/agents/devops/agent.go

## Info
- **File**: `plugins/agents/devops/agent.go`
- **Package**: `devops`
- **Depends on**: 4.28
- **Time**: 15 min
- **Verify**: `go build ./plugins/agents/devops/...`

## Purpose
Implements the DevOps Developer agent package constructor, loading YAML configurations and wrapping the SDK BaseAgent.

## EXACT code to create

```go
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
```

## Pitfalls

### Pitfall 1: Bypassing registry registrations
Ensure constructors load the manifest from correct paths to avoid startup crashes in registries.

### Pitfall 2: Bypassing delegate lifecycle chains
Override methods must invoke equivalent base hooks to preserve state trackers.

## Verify
```bash
go build ./plugins/agents/devops/...
```

## Checklist
- [ ] File exists at `plugins/agents/devops/agent.go`
- [ ] Package name is `devops`
- [ ] All exported types have Godoc
- [ ] `DevOpsAgent` embeds `sdkagent.BaseAgent`
- [ ] Build command passes
