// Package orchestrator defines the contract for the main orchestration engine.
// The orchestrator coordinates the entire mission execution flow:
// Mission → Plan → Schedule → Execute → Aggregate → Result
package orchestrator

import (
	"context"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/contracts/planner"
)

// Orchestrator coordinates the entire mission execution flow.
//
// It is the top-level component that:
//  1. Receives a Mission from the user
//  2. Uses the Planner to decompose it into tasks
//  3. Uses the Scheduler to order and dispatch tasks
//  4. Uses the Runtime to execute tasks via agents
//  5. Aggregates results into a MissionResult
//
// Thread-safety: ExecuteMission must be safe for concurrent use.
// Multiple missions can execute simultaneously.
type Orchestrator interface {
	// ExecuteMission runs a mission from start to finish.
	//
	// This is the main entry point. It blocks until the mission completes
	// or the context is cancelled.
	//
	// Returns:
	//   - *MissionResult: the aggregated result of all tasks
	//   - error: system-level errors (kernel not started, etc.)
	ExecuteMission(ctx context.Context, mission *planner.Mission) (*MissionResult, error)

	// Status returns the current status of a running or completed mission.
	//
	// Can be called while ExecuteMission is running (from another goroutine)
	// to get real-time progress updates.
	Status(missionID string) (*MissionStatus, error)

	// Cancel cancels a running mission.
	// All in-flight tasks are cancelled via context cancellation.
	// Returns nil if the mission was successfully cancelled.
	Cancel(missionID string) error
}

// MissionResult is the final output of a completed mission.
type MissionResult struct {
	// MissionID links to the original mission.
	MissionID string `json:"mission_id"`

	// Status is the overall mission outcome.
	Status contracts.Status `json:"status"`

	// Tasks maps task ID → task result.
	Tasks map[string]*agent.Result `json:"tasks"`

	// Summary is a human-readable summary of what was accomplished.
	Summary string `json:"summary"`

	// Artifacts are all files/outputs produced across all tasks.
	Artifacts []agent.Artifact `json:"artifacts"`

	// Duration is the total mission execution time.
	Duration time.Duration `json:"duration"`
}

// MissionStatus provides real-time progress of a running mission.
// Used by CLI progress display and API status endpoint.
type MissionStatus struct {
	// MissionID identifies the mission.
	MissionID string `json:"mission_id"`

	// Status is the current overall status.
	Status contracts.Status `json:"status"`

	// CurrentTask is the name of the currently executing task.
	CurrentTask string `json:"current_task,omitempty"`

	// TotalTasks is the total number of tasks in the plan.
	TotalTasks int `json:"total_tasks"`

	// DoneTasks is the number of completed tasks (success + failed + skipped).
	DoneTasks int `json:"done_tasks"`

	// FailedTasks is the number of failed tasks.
	FailedTasks int `json:"failed_tasks"`

	// Elapsed is time since mission started.
	Elapsed time.Duration `json:"elapsed"`
}

// Progress returns the completion percentage (0-100).
func (s *MissionStatus) Progress() int {
	if s.TotalTasks == 0 {
		return 0
	}
	return (s.DoneTasks * 100) / s.TotalTasks
}
