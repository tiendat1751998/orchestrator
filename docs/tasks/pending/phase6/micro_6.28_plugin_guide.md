# Micro-Task 6.28: Create docs/plugin-development.md

## Info
- **File**: `docs/plugin-development.md`
- **Depends on**: Phase 1 (contracts), Phase 3 (SDK)
- **Time**: 20 min
- **Verify**: Visual review

## Purpose
Step-by-step guide for creating custom agents, providers, and tools. Includes full working examples with code.

## Key sections to include

### 1. Creating a Custom Agent
```go
package myagent

import (
    "context"
    "log/slog"

    contractsagent "github.com/tiendat1751998/orchestrator/contracts/agent"
    "github.com/tiendat1751998/orchestrator/contracts/provider"
    sdkagent "github.com/tiendat1751998/orchestrator/sdk/agent"
)

type MyAgent struct {
    *sdkagent.BaseAgent
}

func NewMyAgent(manifestPath string, p provider.Provider, logger *slog.Logger) (*MyAgent, error) {
    manifest, err := sdkagent.LoadManifest(manifestPath)
    if err != nil { return nil, err }

    base, err := sdkagent.NewBaseAgent(manifest, p, logger)
    if err != nil { return nil, err }

    return &MyAgent{BaseAgent: base}, nil
}

func (a *MyAgent) Execute(ctx context.Context, task *contractsagent.Task) (*contractsagent.Result, error) {
    // Custom pre-processing
    result, err := a.BaseAgent.Execute(ctx, task)
    // Custom post-processing
    return result, err
}
```

### 2. Creating a Custom Provider
- Implement `provider.Provider` interface
- Extend `sdkprovider.BaseProvider` for common functionality
- Register via plugin manifest

### 3. Creating a Custom Tool
- Implement `tool.Tool` interface
- Define JSON Schema for parameters
- Register via `RegisterTool`

### 4. Manifest YAML Format
```yaml
name: my-agent
version: 1.0.0
role: specialist
capabilities:
  - code_generation
  - testing
prompts:
  system: prompts/system.md
```

### 5. Plugin Registration
```go
func init() {
    plugin.Register("my-agent", func(cfg provider.Config, logger *slog.Logger) (plugin.Plugin, error) {
        return NewMyAgent("manifests/my-agent.yaml", cfg, logger)
    })
}
```

## Checklist
- [ ] Custom agent example with BaseAgent embedding
- [ ] Custom provider example
- [ ] Custom tool example
- [ ] Manifest YAML format documentation
- [ ] Plugin registration instructions
