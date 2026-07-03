# Micro-Task 1.27: Create contracts/workflow/workflow.go

## Info
- **File**: `contracts/workflow/workflow.go`
- **Package**: `workflow`
- **Depends on**: 1.07 (contracts/status.go), 1.18 (agent/task.go)
- **Time**: 10 min
- **Verify**: `go build ./contracts/workflow/...`

## Purpose
Declares structural representations (`Workflow` interface, `Step`, `Result`, `StepResult`) for executing human-defined, reusable sequences of agent tasks.

## EXACT code to create

```go
// Package workflow defines the contract for predefined execution flows.
// Workflows are reusable sequences of steps (e.g., "build and deploy", "review and merge").
package workflow

import (
	"context"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
)

// Workflow defines a sequence of steps to execute.
//
// Unlike missions (which are AI-planned), workflows are human-defined.
// They are predefined sequences stored in YAML files.
//
// Example workflow.yaml:
//
//	name: build-and-deploy
//	steps:
//	  - name: generate_code
//	    agent: backend
//	    task: code_generation
//	  - name: review_code
//	    agent: reviewer
//	    task: code_review
//	    depends_on: [generate_code]
//	  - name: deploy
//	    agent: devops
//	    task: deployment
//	    depends_on: [review_code]
//	    condition: "previous.status == 'success'"
type Workflow interface {
	// Name returns the workflow identifier.
	Name() string

	// Steps returns the list of steps in execution order.
	Steps() []Step

	// Execute runs all steps in the workflow.
	// The workflow engine manages dependencies and conditions.
	Execute(ctx context.Context, input map[string]any) (*Result, error)
}

// Step defines a single step in a workflow.
type Step struct {
	// Name identifies this step (unique within the workflow).
	Name string `yaml:"name" json:"name"`

	// Agent is the name of the agent to execute this step.
	Agent string `yaml:"agent" json:"agent"`

	// Task is the type of task to create (matches a Capability).
	Task string `yaml:"task" json:"task"`

	// DependsOn lists step names that must complete before this step.
	// The workflow engine builds a DAG from these dependencies.
	DependsOn []string `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`

	// Condition is an expression that must be true for this step to execute.
	// Example: "previous.status == 'success'"
	Condition string `yaml:"condition,omitempty" json:"condition,omitempty"`

	// OnFailure defines behavior when this step fails ("retry", "skip", "abort").
	// Default: "abort"
	OnFailure string `yaml:"on_failure,omitempty" json:"on_failure,omitempty"`

	// MaxRetries is the maximum retry attempts (used when OnFailure="retry").
	// Default: 0
	MaxRetries int `yaml:"max_retries,omitempty" json:"max_retries,omitempty"`
}

// Result represents the outcome of a workflow execution.
type Result struct {
	// Status is the overall workflow status.
	Status contracts.Status `json:"status"`

	// Steps maps step name → step result.
	Steps map[string]*StepResult `json:"steps"`

	// Duration is the total execution time.
	Duration time.Duration `json:"duration"`
}

// StepResult is the outcome of a single workflow step.
type StepResult struct {
	// Status of this step.
	Status contracts.Status `json:"status"`

	// Output is the step's text output.
	Output any `json:"output"`

	// Error is a human-readable error message (if failed).
	Error string `json:"error,omitempty"`

	// Duration is how long this step took.
	Duration time.Duration `json:"duration"`
}
```

## Rules
1. **Human vs Machine Directed**: Workflows represent static human plans loaded via configs, whereas Missions are dynamic plans generated on-the-fly by Planner LLMs.
2. **Failure Defaults**: If a step fails and `OnFailure` is blank, it must default to `"abort"`, stopping the entire workflow run immediately.
3. **Execution Dependency Graph**: Steps in a workflow build a topological dependency layout using the `DependsOn` step name strings.

## ⚠️ Pitfalls

### Pitfall 1: Confusing Workflow Steps with Planner Tasks
Workflows have static step declarations that specify the exact agent name (`backend`, `devops`). Planner tasks do not bind to specific agent instances, but declare capability dependencies (`Type: "code_generation"`), matched by the orchestrator registry at runtime.

### Pitfall 2: Circular dependencies in workflow configuration steps
If step A depends on step B, and step B depends on step A, executing the workflow will lock or fail validation. Implementations must run a validation cycle to detect dependency cycles before running the steps.

## Verify
```bash
go build ./contracts/workflow/...
```

## Checklist
- [ ] File `contracts/workflow/workflow.go` exists
- [ ] Package: `workflow`
- [ ] `Workflow` interface contains Name, Steps, and Execute methods
- [ ] `Step` contains Name, Agent, Task, DependsOn, Condition, OnFailure, and MaxRetries fields
- [ ] `Result` and `StepResult` structures are declared
- [ ] All structures define both YAML and JSON tags
- [ ] `go build ./contracts/workflow/...` passes
