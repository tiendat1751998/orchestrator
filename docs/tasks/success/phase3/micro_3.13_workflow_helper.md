# Micro-Task 3.13: Create sdk/workflow/workflow.go

## Info
- **File**: `sdk/workflow/workflow.go`
- **Package**: `workflow`
- **Depends on**: 3.01 (base_plugin.md), 1.27 (workflow.go contract)
- **Time**: 15 min
- **Verify**: `go build ./sdk/workflow/...`

## Purpose
Implements the base workflow structure (`BaseWorkflow` and constructors) and a topological sorting algorithm (`SortSteps`) to schedule step execution based on dependency constraints, resolving cycles before execution.

## EXACT code to create

```go
// Package workflow provides base helpers and dependency tracking for human-defined workflows.
package workflow

import (
	"context"
	"errors"
	"fmt"

	contractsplugin "github.com/tiendat1751998/orchestrator/contracts/plugin"
	contractsworkflow "github.com/tiendat1751998/orchestrator/contracts/workflow"
	sdkplugin "github.com/tiendat1751998/orchestrator/sdk/plugin"
)

// BaseWorkflow embeds BasePlugin and provides default implementations for Name and Steps.
type BaseWorkflow struct {
	*sdkplugin.BasePlugin
	steps []contractsworkflow.Step
}

// NewBaseWorkflow constructs a BaseWorkflow.
func NewBaseWorkflow(name string, steps []contractsworkflow.Step) (*BaseWorkflow, error) {
	if name == "" {
		return nil, errors.New("sdk/workflow: workflow name cannot be empty")
	}

	basePlugin, err := sdkplugin.NewBasePlugin(name, contractsplugin.TypeWorkflow, "1.0.0")
	if err != nil {
		return nil, err
	}

	// Validate steps do not contain duplicate names
	seen := make(map[string]bool)
	for _, step := range steps {
		if step.Name == "" {
			return nil, errors.New("sdk/workflow: workflow step name cannot be empty")
		}
		if seen[step.Name] {
			return nil, fmt.Errorf("sdk/workflow: duplicate step name %q in workflow", step.Name)
		}
		seen[step.Name] = true
	}

	return &BaseWorkflow{
		BasePlugin: basePlugin,
		steps:      steps,
	}, nil
}

// Steps returns the defined steps list.
func (bw *BaseWorkflow) Steps() []contractsworkflow.Step {
	copied := make([]contractsworkflow.Step, len(bw.steps))
	copy(copied, bw.steps)
	return copied
}

// SortSteps topologically sorts the workflow steps based on their DependsOn field.
func SortSteps(steps []contractsworkflow.Step) ([]contractsworkflow.Step, error) {
	adj := make(map[string][]string)
	inDegree := make(map[string]int)
	stepsMap := make(map[string]contractsworkflow.Step)

	for _, step := range steps {
		stepsMap[step.Name] = step
		if _, exists := inDegree[step.Name]; !exists {
			inDegree[step.Name] = 0
		}
		for _, dep := range step.DependsOn {
			adj[dep] = append(adj[dep], step.Name)
			inDegree[step.Name]++
		}
	}

	// Kahn's algorithm topological sort
	var queue []string
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	var sorted []contractsworkflow.Step
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if step, ok := stepsMap[curr]; ok {
			sorted = append(sorted, step)
		}

		for _, neighbor := range adj[curr] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(sorted) != len(steps) {
		return nil, errors.New("sdk/workflow: circular dependency detected in workflow steps")
	}

	return sorted, nil
}

// Execute is a placeholder implementation. Concrete engines override this.
func (bw *BaseWorkflow) Execute(ctx context.Context, input map[string]any) (*contractsworkflow.Result, error) {
	return nil, errors.New("sdk/workflow: Execute method must be overridden by concrete workflows")
}
```

## Rules
1. **Duplicate Name Guards**: Enforce checks against duplicate step names inside constructors to avoid nodes conflicts during dependency sorting.
2. **Kahn's Sort Validation**: Execute Kahn's algorithm to resolve dependencies, checking `len(sorted) == len(steps)` to verify no circular dependencies exist.
3. **Array Copy Protection**: Always return copies of the internal step list to prevent external modifications to the workflow configuration.

## ⚠️ Pitfalls

### Pitfall 1: Bypassing duplicate step name validation checks
If steps are registered with identical names, maps lookups collide during topological sorting. Always reject duplicate name configs during initialization.

### Pitfall 2: Permitting circular dependencies to go un-flagged
If cycles are present, the sorted list will omit nodes in the dependency loop. Failing to check list lengths causes the engine to skip executing steps silently.

## Verify
```bash
go build ./sdk/workflow/...
```

## Checklist
- [ ] File `sdk/workflow/workflow.go` exists
- [ ] Package: `workflow`
- [ ] `BaseWorkflow` integrates lifecycle properties
- [ ] Constructor validates duplicate step names
- [ ] `Steps` returns copied arrays
- [ ] `SortSteps` initializes adjacency maps and in-degree counts
- [ ] Kahn's sorting detects circular dependencies and returns error
- [ ] Placeholder `Execute` returns clean error overrides warnings
- [ ] `go build ./sdk/workflow/...` passes
