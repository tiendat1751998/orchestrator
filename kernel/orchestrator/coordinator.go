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
