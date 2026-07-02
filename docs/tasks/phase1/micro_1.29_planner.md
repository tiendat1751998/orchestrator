# Micro-Task 1.29: Tạo contracts/planner/planner.go

## Thông tin
- **File tạo**: `contracts/planner/planner.go`
- **Package**: `planner`
- **Dependencies trước**: 1.18 (agent/task.go)
- **Thời gian**: 10 phút
- **Verify**: `go build ./contracts/planner/...`

## Nội dung CHÍNH XÁC cần tạo

```go
// Package planner defines the contract for mission planning.
// The planner decomposes high-level missions into executable task graphs.
package planner

import (
	"context"

	"github.com/tiendat1751998/orchestrator/contracts/agent"
)

// Planner decomposes missions into executable task graphs (DAGs).
//
// The planner uses an AI provider to analyze the mission and break it
// into specific, actionable tasks with dependencies.
//
// Example flow:
//   Mission: "Build a REST API for user management"
//   ↓ Plan()
//   Tasks: [
//     {name: "design_api", type: "architecture"},
//     {name: "implement_handlers", type: "code_generation", depends: ["design_api"]},
//     {name: "write_tests", type: "testing", depends: ["implement_handlers"]},
//     {name: "review_code", type: "code_review", depends: ["implement_handlers"]},
//   ]
type Planner interface {
	// Plan decomposes a mission into a list of tasks with dependencies.
	//
	// The returned tasks form a DAG (Directed Acyclic Graph).
	// The orchestrator uses the DAG to determine execution order.
	//
	// Returns:
	//   - []*agent.Task: list of tasks (may be empty for trivial missions)
	//   - error: if planning fails (provider error, invalid mission)
	Plan(ctx context.Context, mission *Mission) ([]*agent.Task, error)

	// Replan creates a new plan when a task fails.
	//
	// The planner analyzes the failure and decides how to recover:
	//   - Retry the same task with different approach
	//   - Replace with alternative tasks
	//   - Skip if not critical
	//   - Abort the mission
	//
	// Parameters:
	//   - mission: the original mission
	//   - failedTask: the task that failed
	//   - err: the error that caused the failure
	//
	// Returns a new list of tasks (replacing the failed task and its dependents).
	Replan(ctx context.Context, mission *Mission, failedTask *agent.Task, err error) ([]*agent.Task, error)
}

// Mission is a high-level goal from the user.
// This is the input to the entire orchestration system.
//
// Example:
//
//	Mission{
//	    ID:          "msn-a1b2c3d4",
//	    Title:       "Build REST API",
//	    Description: "Build a REST API for user management with Go and Gin framework",
//	    Constraints: []string{"use Go", "use Gin framework", "no external ORM"},
//	}
type Mission struct {
	// ID uniquely identifies this mission.
	ID string `json:"id"`

	// Title is a short summary of the mission.
	Title string `json:"title"`

	// Description is a detailed explanation of what the user wants.
	// This is the main input for the AI planner.
	Description string `json:"description"`

	// Constraints are rules the planner and agents must follow.
	// Example: ["use Go", "no external dependencies", "follow hexagonal architecture"]
	Constraints []string `json:"constraints,omitempty"`

	// Metadata for extensibility.
	Metadata map[string]string `json:"metadata,omitempty"`
}
```

## ⚠️ Pitfalls cần tránh
1. **Import agent package**: Planner returns `[]*agent.Task`. OK — planner knows about tasks but tasks don't know about planner.
2. **Mission vs Task**: Mission = what the user wants (high-level). Task = what an agent does (specific). Planner converts mission → tasks.
3. **Replan max attempts**: Replan itself doesn't enforce max attempts. The orchestrator enforces (typically max 3 replans).

## Checklist
- [ ] File `contracts/planner/planner.go` tồn tại
- [ ] Package: `package planner`
- [ ] Import `contracts/agent` cho Task type
- [ ] Planner interface với 2 methods (Plan, Replan)
- [ ] Mission struct với 5 fields
- [ ] Plan returns `[]*agent.Task`
- [ ] Replan nhận failedTask và error
- [ ] Godoc comments với examples
- [ ] `go build ./contracts/planner/...` không lỗi
- [ ] Không có import cycle
