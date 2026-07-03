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
