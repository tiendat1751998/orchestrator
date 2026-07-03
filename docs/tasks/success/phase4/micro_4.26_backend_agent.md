# Micro-Task 4.26: Create plugins/agents/backend/agent.go

## Info
- **File**: `plugins/agents/backend/agent.go`
- **Package**: `backend`
- **Depends on**: 4.25
- **Time**: 15 min
- **Verify**: `go build ./plugins/agents/backend/...`

## Purpose
Implements the backend developer agent class wrapper, loading manifest configurations and embedding the SDK's `BaseAgent` structures.

## EXACT code to create

```go
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
```

## Pitfalls

### Pitfall 1: Bypassing SDK agent constructors
```go
// WRONG:
return &BackendAgent{
    BaseAgent: &sdkagent.BaseAgent{...}, // Fails validation checks of sdkagent package!
}

// CORRECT:
baseAgent, _ := sdkagent.NewBaseAgent(manifest, logger)
return &BackendAgent{BaseAgent: baseAgent}
```
Directly constructing base agent properties instead of calling `sdkagent.NewBaseAgent` skips validation routines, causing nil pointer panics.

### Pitfall 2: Bypassing lifecycle delegate updates
Ensure override hooks (`Init`, `Start`, `Stop`, `Health`) execute equivalent delegate methods on `BaseAgent` to keep registry statuses synchronized.

## Verify
```bash
go build ./plugins/agents/backend/...
```

## Checklist
- [ ] File exists at `plugins/agents/backend/agent.go`
- [ ] Package name is `backend`
- [ ] All exported types have Godoc
- [ ] `BackendAgent` embeds `sdkagent.BaseAgent`
- [ ] Lifecycle hooks call parent methods
- [ ] Build command passes
