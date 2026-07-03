# RFC-0010: Cognitive Architecture (AEOS Brain Core)

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0002 (Brain Architecture), RFC-0003 (Knowledge Engine), RFC-0008 (Event Model & Event Sourcing)

## Summary

This RFC specifies the **Cognitive Architecture** of the AI Engineering Operating System (AEOS). Rather than relying on static prompt templates or raw LLM adjustments, AEOS implements a 5-tier cognitive model (Perception, Memory, Reasoning, Learning, Action) that incorporates 4 distinct dimensions of memory (Working, Episodic, Semantic, Procedural) and uses specialized Experience and Pattern Engines. This allows the system's reasoning capability ("IQ") to increase autonomously as it runs more projects.

## Motivation

AI Orchestrators typically suffer from "cold start" and repetitive failure patterns. Without a structured cognitive core:
- The system cannot remember specific historical failures (e.g. Mission #25 failed because of a Database Deadlock) and will repeat the same plan.
- The system cannot infer implicit rules of thumb for a codebase (e.g. "this project always uses Gin for REST, and doesn't use GORM").
- Planning must be generated from scratch on every run, leading to high cost, high latency, and unpredictable outputs.

Implementing a cognitive architecture allows the system to observe its own execution, accumulate episodes, mine patterns, and optimize its reasoning logic.

## Design

### 1. The 5-Tier Cognitive Core

The AEOS Brain Core is structured into five distinct operational tiers:

```
 Perception Tier   ◄── Reads AST, Git Diffs, Execution Logs, System Metrics
       │
       ▼
  Memory Tier      ◄── Working (RAM) | Episodic (Events) | Semantic (Graph) | Procedural (Skills)
       │
       ▼
 Reasoning Tier    ◄── Decomposes, Plans, and runs Meta-Thinking
       │
       ▼
  Learning Tier    ◄── Reflection Engine, Experience Engine, Pattern Engine
       │
       ▼
  Action Tier      ◄── Dispatches tasks to Sandboxed Agents & Tools
```

| Tier | Responsibility | Core Components |
|---|---|---|
| **Perception** | Sensing the workspace & system state | AST Parser, Git Diff Reader, Log Stream Analyzer, Metric Monitors |
| **Memory** | Retaining short-term and long-term data | Working Memory, Episodic Log, Semantic Graph, Procedural Skills |
| **Reasoning** | Thinking, planning, and self-review | Planner, Meta-Thinking Controller, Rule Evaluator |
| **Learning** | Self-improving feedback loops | Reflection Engine, Experience Engine, Pattern Engine |
| **Action** | Executing decisions via sandboxes | Scheduler, Process Manager, Capability Guards |

---

### 2. The 4 Memory Dimensions (RFC-0005 Expanded)

To make AEOS a true "Second Brain", the memory tier manages four isolated dimensions:

```
  ┌────────────────────────────────────────────────────────┐
  │                      Memory Tier                       │
  ├────────────────────────────────────────────────────────┤
  │ 1. Working Memory    : Mission-scoped active variables │
  │                        (RAM map).                      │
  ├────────────────────────────────────────────────────────┤
  │ 2. Episodic Memory   : Chronological database of past  │
  │                        missions, outcomes, and failure │
  │                        episodes (e.g., deadlock logs). │
  ├────────────────────────────────────────────────────────┤
  │ 3. Semantic Memory   : Graph database storing ontology │
  │                        relations (e.g. "Gin" is-a      │
  │                        "HTTP Router").                 │
  ├────────────────────────────────────────────────────────┤
  │ 4. Procedural Memory : Collection of codebase habits   │
  │                        and skills (e.g., Gin coding    │
  │                        styles, GORM exclusions).       │
  └────────────────────────────────────────────────────────┘
```

---

### 3. Contracts (`contracts/brain/cognitive.go`)

We define the contracts for the new Cognitive layers, replacing the old basic learning engine definitions:

```go
package brain

import (
	"context"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/event"
	"github.com/tiendat1751998/orchestrator/contracts/mission"
)

type ErrorCategory string

const (
	ErrCategoryExecution      ErrorCategory = "execution"      // Code/Logic compiler errors (penalize scorecard)
	ErrCategoryInfrastructure ErrorCategory = "infrastructure" // Network timeouts, API 429/500 limits (NO penalty)
	ErrCategorySystem         ErrorCategory = "system"         // Host out-of-memory, file permission errors (NO penalty)
)

// TaskPerformance tracks metrics for a single task execution block.
type TaskPerformance struct {
	TaskID        string        `json:"task_id"`
	TaskType      string        `json:"task_type"`
	ExecutorName  string        `json:"executor_name"`
	Duration      time.Duration `json:"duration"`
	Success       bool          `json:"success"`
	RetriesCount  int           `json:"retries_count"`
	QualityScore  float64       `json:"quality_score"` // Heuristic rating (0.0 to 1.0)
	ErrorCategory ErrorCategory `json:"error_category,omitempty"`
	ErrorMessage  string        `json:"error_message,omitempty"`
}

// ReflectionReport is the output of the Reflection Engine.
type ReflectionReport struct {
	MissionID        string            `json:"mission_id"`
	Success          bool              `json:"success"`
	TaskPerformances []TaskPerformance `json:"task_performances"`
	RootCauseError   string            `json:"root_cause_error,omitempty"`
	LessonLearned    string            `json:"lesson_learned,omitempty"`
	ExtractTemplate  bool              `json:"extract_template"`
}

// DecisionRecord represents a historical decision and its technical rationale.
type DecisionRecord struct {
	ID        string    `json:"id"`
	MissionID string    `json:"mission_id"`
	Action    string    `json:"action"`    // e.g. "split_service"
	Target    string    `json:"target"`    // e.g. "UserService"
	Rationale string    `json:"rationale"` // Technical reasoning, e.g. "Circular Dependency risk"
	Timestamp time.Time `json:"timestamp"`
}

// SkillNode represents an agent's capability level in a Skill Tree.
type SkillNode struct {
	AgentID   string    `json:"agent_id"`
	SkillName string    `json:"skill_name"` // e.g. "CRUD", "CQRS", "DDD"
	Level     int       `json:"level"`      // e.g. 1, 2, 3
	ExpPoints int       `json:"exp_points"`
	Unlocked  bool      `json:"unlocked"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TruthReport represents detailed reports for each validation stage.
type TruthReport struct {
	Passed          bool     `json:"passed"`
	CompilerOutput  string   `json:"compiler_output,omitempty"`
	LinterWarnings  []string `json:"linter_warnings,omitempty"`
	TestsPassed     int      `json:"tests_passed"`
	TestsFailed     int      `json:"tests_failed"`
	SecurityIssues  []string `json:"security_issues,omitempty"`
	RejectedReasons []string `json:"rejected_reasons,omitempty"`
}

// TruthPipeline executes multi-stage validation checks on generated assets before learning.
type TruthPipeline interface {
	Verify(ctx context.Context, codePath string) (*TruthReport, error)
}

// ReflectionEngine inspects episodic timeline events.
type ReflectionEngine interface {
	// Reflect evaluates timeline events to construct an execution performance report.
	Reflect(ctx context.Context, m *mission.Mission, events []event.Event) (*ReflectionReport, error)
}

// Experience represents a technology recipe or configuration pair (e.g., Gin + sqlc + Redis).
type Experience struct {
	ID        string    `json:"id"`
	Category  string    `json:"category"` // e.g. "http_db_cache"
	Stack     []string  `json:"stack"`    // e.g. ["gin", "sqlc", "redis"]
	UseCount  int       `json:"use_count"`
	Successes int       `json:"successes"`
	Score     float64   `json:"score"` // Confidence score (0.0 to 1.0)
	UpdatedAt time.Time `json:"updated_at"`
}

// ExperienceEngine accumulates and scores technology stack recipes.
type ExperienceEngine interface {
	// QueryExperience returns the best technology stack experience matching constraints.
	QueryExperience(ctx context.Context, category string, tags []string) (*Experience, error)
	// RecordUsage updates successes and usage statistics for an experience stack.
	RecordUsage(ctx context.Context, experienceID string, success bool) error
}

// Pattern represents a structural code pattern (e.g. Saga, Outbox, DDD Repository).
type Pattern struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ContextTags []string `json:"context_tags"`
	CodeSample  string   `json:"code_sample,omitempty"`
	Score       float64  `json:"score"`
}

// PatternEngine mines design patterns from successful project directories.
type PatternEngine interface {
	// Mine extracts potential pattern files and registers them in the Knowledge Graph.
	Mine(ctx context.Context, workspacePath string) ([]Pattern, error)
	// Match returns relevant patterns for the planner.
	Match(ctx context.Context, tags []string) ([]Pattern, error)
}

// MetaThinkingController runs pre-plan evaluations to ensure quality.
type MetaThinkingController interface {
	// EvaluatePlan checks the generated plan for issues (redundancies, loops, locks)
	// before scheduling it.
	EvaluatePlan(ctx context.Context, plan *Plan) (*PlanValidationResult, error)
}

type PlanValidationResult struct {
	IsValid     bool     `json:"is_valid"`
	Warnings    []string `json:"warnings,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
}
```

---

### 4. Implementations & Learning Engines (`kernel/brain/cognitive/`)

#### A. Experience Scorecard Update (EMA Model)
The `ExperienceEngine` recalculates scorecard confidence for stack configurations. If a stack of `sqlc + Gin + Redis` fails because of a coding logic issue (classified as `ErrCategoryExecution`), we penalize the score. If it fails due to network outage, we preserve the score:

```go
// kernel/brain/cognitive/experience.go
package cognitive

import (
	"context"
	
	"github.com/tiendat1751998/orchestrator/contracts/brain"
	"github.com/tiendat1751998/orchestrator/contracts/knowledge"
)

type expEngine struct {
	graph knowledge.KnowledgeGraph
	alpha float64 // Learning rate (e.g., 0.1)
}

func NewExperienceEngine(g knowledge.KnowledgeGraph) brain.ExperienceEngine {
	return &expEngine{graph: g, alpha: 0.10}
}

func (ee *expEngine) RecordUsage(ctx context.Context, expID string, success bool) error {
	node, err := ee.graph.GetNode(ctx, "experience:"+expID)
	if err != nil {
		return err
	}

	outcome := 0.0
	if success {
		outcome = 1.0
	}

	// Update score using EMA formula
	node.Score = (node.Score * (1 - ee.alpha)) + (outcome * ee.alpha)
	node.UsedCount++
	if success {
		node.SuccessCount++
	}

	return ee.graph.UpdateNode(ctx, *node)
}
```

#### B. Episodic Fault Checking (Reflection Engine)
During reflection, if a mission fails, the `ReflectionEngine` writes an **Episodic Failure Entry** to the Knowledge Graph.
- For example, if Mission #25 failed because of a `Deadlock` in the Postgres SQL transaction:
- The engine creates a Node with tag `["postgres", "deadlock", "failure-episode"]`.
- When the **Planner** starts a new mission with similar Postgres constraints, the Semantic Search queries related failure episodes and injects a warning:
  > *"Warning: Past Episode (Mission #25) failed due to a Deadlock. Avoid nested transactions when planning database handlers."*
This effectively gives the system an "IQ" that rises as it gains experience.

## Impact

- **Decoupled Cognition**: Reasoning, reflection, and experience updates run in background worker threads without adding latency to active execution runtimes.
- **Permanent Skills**: Codebase constraints (Gin over GORM) are stored inside the Graph's Procedural Memory node and automatically appended to the Context Engine prompt builder.

## Open Questions

1. **How do we evaluate the QualityScore heuristically?**
   - We execute automated testing suites and code linters. If the linter reports 0 warnings, the quality score defaults to $1.0$. If compilation fails, the quality is $0.0$. If tests pass but have high latency, the quality is rated proportionally.
