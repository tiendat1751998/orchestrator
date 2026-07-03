# RFC-0011: Kubernetes-Style Scheduler

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0001 (Kernel Architecture), RFC-0009 (Resource Manager)

## Summary

This RFC specifies the architecture of the **Scheduler** in the AI Engineering Operating System (AEOS). To support complex task dependency trees, varying resource weights, and provider rate-limiting, the Scheduler is structured as a Kubernetes-inspired pipeline: Queue → Prioritize → Filter & Affinity → Resource Match → Worker Bind → Execute.

## Motivation

Issue 6 from the architecture review highlighted the need for a more advanced scheduling model. Simple FIFO (First-In, First-Out) or basic parallel task queues fall short when:
- Tasks depend on other tasks (DAG constraints).
- Certain tasks require specialized agents/models (Affinity).
- Multiple missions run simultaneously and compete for the same hardware resources or API rate quotas.

Implementing a decoupled, stage-based scheduler ensures fair task allocation, handles backpressure, and guarantees optimal execution paths.

## Design

### 1. Scheduling Pipeline Stages

Every task scheduled for execution is processed through these distinct stages:

```
  Plan Scheduled (Tasks pushed to Queue)
                   │
                   ▼
  ┌────────────────────────────────────────────────────────┐
  │ 1. Queue       : Holds pending tasks grouped by Priority│
  └────────────────┬───────────────────────────────────────┘
                   │
                   ▼
  ┌────────────────────────────────────────────────────────┐
  │ 2. Prioritize  : Sort tasks based on DAG dependencies   │
  │                  and user urgency metadata             │
  └────────────────┬───────────────────────────────────────┘
                   │ Sorted Tasks
                   ▼
  ┌────────────────────────────────────────────────────────┐
  │ 3. Filter &    : Select matching agents that support    │
  │    Affinity      required capabilities.                │
  └────────────────┬───────────────────────────────────────┘
                   │ Candidate Workers
                   ▼
  ┌────────────────────────────────────────────────────────┐
  │ 4. Resource    : Query Resource Manager (RAM, CPU,     │
  │    Match         API Quotas, Cooldown limits)          │
  └────────────────┬───────────────────────────────────────┘
                   │ Allocated Worker & Slot
                   ▼
  ┌────────────────────────────────────────────────────────┐
  │ 5. Bind        : Assign task to specific executor      │
  └────────────────┬───────────────────────────────────────┘
                   │
                   ▼
             Execute Task
```

---

### 2. Contracts (`contracts/scheduler/`)

```go
// contracts/scheduler/scheduler.go
package scheduler

import (
	"context"
	
	"github.com/tiendat1751998/orchestrator/contracts/agent"
)

// TaskPriority represents scheduling urgency.
type TaskPriority int

const (
	PriorityLow    TaskPriority = 100
	PriorityMedium TaskPriority = 500
	PriorityHigh   TaskPriority = 900
)

// SchedulingConstraints defines worker binding limits.
type SchedulingConstraints struct {
	RequiredAgent      string            `json:"required_agent,omitempty"`      // Exact match
	RequiredProvider   string            `json:"required_provider,omitempty"`   // e.g. Gemini
	AffinityCapabilities []string         `json:"affinity_capabilities,omitempty"`
	MinMemoryBytes     uint64            `json:"min_memory_bytes,omitempty"`
}

// Queue manages the active pending tasks.
type Queue interface {
	Push(ctx context.Context, task *agent.Task, priority TaskPriority, constraints SchedulingConstraints) error
	Pop(ctx context.Context) (*agent.Task, error)
	Len() int
}

// Scheduler handles task cycle pipeline assignments.
type Scheduler interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	
	// Schedule registers tasks in the queue.
	Schedule(ctx context.Context, tasks []*agent.Task, priority TaskPriority) error
}
```

---

### 3. Execution Pipeline Implementation (`kernel/execution/scheduler/`)

The scheduling loop executes as a background controller process:

```go
// kernel/execution/scheduler/controller.go
package scheduler

import (
	"context"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/contracts/plugin"
	"github.com/tiendat1751998/orchestrator/contracts/resource"
	"github.com/tiendat1751998/orchestrator/contracts/scheduler"
)

type schedulerController struct {
	queue           scheduler.Queue
	resourceManager resource.ResourceManager
	agentRegistry   *plugin.Registry[agent.Agent]
	dispatchChan    chan<- *agent.Task
	stopChan        chan struct{}
}

func (s *schedulerController) Start(ctx context.Context) error {
	go s.runLoop(ctx)
	return nil
}

func (s *schedulerController) runLoop(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			if s.queue.Len() == 0 {
				continue
			}

			// 1. Pop next high-priority task candidates
			task, err := s.queue.Pop(ctx)
			if err != nil {
				continue
			}

			// 2. Filter available agents matching capabilities
			candidates := s.agentRegistry.FindByCapability(task.Type)
			if len(candidates) == 0 {
				// No matching agent, mark failed or reschedule
				s.reschedule(ctx, task)
				continue
			}

			// 3. Select best candidate using Affinity/Resource allocations
			bound := false
			for _, candidate := range candidates {
				meta := candidate.Metadata()
				// Query resource availability (Estimated usage checks)
				ok, _ := s.resourceManager.Allocate(ctx, resource.ResourceRequest{
					ProviderName:    task.RequiredProvider,
					ModelName:       task.RequiredModel,
					EstimatedTokens: task.EstimatedTokens,
					CPURequirement:  task.CPURequirement,
					RAMRequirement:  task.RAMRequirement,
				})
				if ok {
					// 4. Bind & dispatch
					task.AssignedAgent = meta.Name
					s.dispatchChan <- task
					bound = true
					break
				}
			}

			if !bound {
				// Re-queue task if blocked by resource limitations (backpressure)
				s.reschedule(ctx, task)
			}
		}
	}
}

func (s *schedulerController) reschedule(ctx context.Context, t *agent.Task) {
	s.queue.Push(ctx, t, scheduler.PriorityMedium, scheduler.SchedulingConstraints{})
}
```

## Impact

- **DAG Handling**: Task dependency checks are executed before pushing to the Queue. Only tasks with all dependency states = `Completed` are added to the ready queue.
- **Resource Backpressure**: Tasks are held in the priority queue if the host CPU is overloaded or if targeted model APIs are in a cooldown state.

## Open Questions

1. **How is affinity score evaluated?**
   - The scheduler matches the requested capability tag directly. In Phase 5, we can use history success rates (from Cognitive Layer node scores) to score candidate bindings dynamically.
