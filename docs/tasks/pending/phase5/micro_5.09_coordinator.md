# Micro-Task 5.09: Create kernel/orchestrator/coordinator.go

## Info
- **File**: `kernel/orchestrator/coordinator.go`
- **Package**: `orchestrator`
- **Depends on**: 5.08
- **Time**: 15 min
- **Verify**: `go build ./kernel/orchestrator/...`

## Purpose
Implements the coordination context builder (`Coordinator` and helpers) to parse and inject dependencies outputs into target tasks parameters.

## EXACT code to create

```go
package orchestrator

import (
	"fmt"
	"log/slog"

	contractsagent "github.com/tiendat1751998/orchestrator/contracts/agent"
)

// Coordinator handles dependency injection between tasks.
type Coordinator struct {
	logger *slog.Logger
}

// NewCoordinator constructs a new Coordinator.
func NewCoordinator(logger *slog.Logger) *Coordinator {
	return &Coordinator{
		logger: logger,
	}
}

// InjectDependencyResults merges dependent task outputs into the target task's Input map.
func (c *Coordinator) InjectDependencyResults(task *contractsagent.Task, dependenciesResults map[string]*contractsagent.Result) error {
	if task == nil {
		return fmt.Errorf("coordinator: task cannot be nil")
	}

	if len(task.Dependencies) == 0 {
		return nil
	}

	depData := make(map[string]any)

	for _, depID := range task.Dependencies {
		res, ok := dependenciesResults[string(depID)]
		if !ok || res == nil {
			continue
		}

		depData[string(depID)] = map[string]any{
			"status": res.Status,
			"output": res.Output,
			"error":  res.Error,
		}
	}

	// Initialize Input map if nil
	if task.Input == nil {
		task.Input = make(map[string]any)
	}

	// Inject dependency results into the task's Input map
	task.Input["_dependency_results"] = depData
	return nil
}
```

## Pitfalls

### Pitfall 1: Overwriting existing parameters on injection
```go
// WRONG:
task.Parameters, _ = json.Marshal(map[string]any{
    "_dependency_results": depData, // Overwrites task description and instructions!
})
```
Replacing parameters map directly deletes task config keys. Always unmarshal parameters and merge results before saving updates.

### Pitfall 2: Silent failures when unmarshalling invalid JSON
If parameters contain malformed parameters, failing to report unmarshal errors will result in parameters being lost.

## Verify
```bash
go build ./kernel/orchestrator/...
# Expected: clean compilation without errors
```

## Checklist
- [ ] File exists at `kernel/orchestrator/coordinator.go`
- [ ] Package name is `orchestrator`
- [ ] All exported types have Godoc
- [ ] Injector preserves existing parameter keys
- [ ] Dependency results map output fields cleanly
- [ ] Build command passes
