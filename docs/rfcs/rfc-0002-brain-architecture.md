# RFC-0002: Brain Architecture — Cognitive Engines

- **Status**: PROPOSED → **REVISED**
- **Priority**: P0 — Foundation
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Revised**: 2026-07-03 (Review fixes: A1, A2, B1, B2, B3, B4)
- **Depends on**: RFC-0000 (State Machine), RFC-0001 (Kernel Architecture)

## Summary

The Brain is the deterministic decision-making center of the system. It contains multiple specialized engines, each responsible for a different cognitive function. AI is NEVER called for decisions — only for content generation via Execution Runtime.

## Motivation

Issue 1 from architecture review: Brain was too small (just Rule → Decision). A real "second brain" needs multiple cognitive engines.

## Design

### Package Structure Decision (Fix for Issue A1)

> [!IMPORTANT]
> **Decision: Flat `package brain` with multiple files.**
>
> Rationale: Go convention is one package per directory. Sub-packages (`brain/decision/`, `brain/planning/`) create import complexity — they can't import each other without circular deps. A single `package brain` in contracts, split across files, is cleaner. The kernel implementation uses sub-packages internally.

```
contracts/brain/           # FLAT package: package brain
├── brain.go               # Brain facade interface
├── decision.go            # DecisionEngine, Situation, Decision, Rule
├── planning.go            # PlanningEngine, Plan, PlanTemplate, TaskSpec
├── policy.go              # PolicyEngine, PolicyAction, PolicyResult, Policy
├── context.go             # ContextEngine, ContextRequest, AssembledContext
├── cognitive.go           # Reflection, Learning, Prediction, Optimization
└── types.go               # Shared types: ActionType, ExecutionStrategy, PlanSource

kernel/brain/              # SUB-PACKAGES for implementation
├── runtime.go
├── decision/              # implements brain.DecisionEngine
├── planning/              # implements brain.PlanningEngine
├── policy/                # implements brain.PolicyEngine
├── context/               # implements brain.ContextEngine
└── cognitive/             # implements brain.ReflectionEngine, etc. (Phase 5+)
```

### Mission Type (Fix for Issue B2)

> [!IMPORTANT]
> `Mission` is a top-level domain concept. It gets its own contract package.

```go
// contracts/mission/mission.go
package mission

import "time"

// Mission represents a user's high-level goal.
// It is the top-level unit of work in the system.
type Mission struct {
    // ID uniquely identifies this mission.
    ID          string            `json:"id"`
    // Title is a short human-readable summary.
    Title       string            `json:"title"`
    // Description is the full mission specification.
    Description string            `json:"description"`
    // Constraints are limitations (e.g., "no external APIs", "Go only").
    Constraints []string          `json:"constraints,omitempty"`
    // Metadata carries additional key-value data.
    Metadata    map[string]string `json:"metadata,omitempty"`
    // CreatedAt is when the mission was submitted.
    CreatedAt   time.Time         `json:"created_at"`
}
```

### Brain Facade

```go
// contracts/brain/brain.go
package brain

import (
    "context"
    
    "github.com/tiendat1751998/orchestrator/contracts/mission"
)

// Brain is the top-level cognitive interface.
// It coordinates all sub-engines to make decisions.
//
// External callers (cmd, modules) use Brain.
// Kernel-internal code may access individual engines directly.
type Brain interface {
    // Decide evaluates a situation and returns an action decision.
    Decide(ctx context.Context, situation Situation) (*Decision, error)
    
    // Plan creates an execution plan for a mission.
    Plan(ctx context.Context, m *mission.Mission) (*Plan, error)
    
    // Enforce checks if an action is allowed by policies.
    Enforce(ctx context.Context, action PolicyAction) (*PolicyResult, error)
    
    // AssembleContext builds the optimal context for an AI call.
    AssembleContext(ctx context.Context, req ContextRequest) (*AssembledContext, error)
    
    // Sub-engine accessors (for kernel-internal use)
    DecisionEngine() DecisionEngine
    PlanningEngine() PlanningEngine
    PolicyEngine() PolicyEngine
    ContextEngine() ContextEngine
}
```

### Decision Engine (Fix for Issue B1 — full definitions)

```go
// contracts/brain/types.go — Shared types
package brain

// ActionType represents the type of action the brain can decide.
type ActionType string

const (
    ActionDispatch  ActionType = "dispatch"   // Dispatch task to executor
    ActionRetry     ActionType = "retry"      // Retry failed task
    ActionFallback  ActionType = "fallback"   // Switch to alternative executor
    ActionSkip      ActionType = "skip"       // Skip non-critical task
    ActionAbort     ActionType = "abort"      // Abort entire mission
    ActionWait      ActionType = "wait"       // Wait for dependencies/cooldown
    ActionEscalate  ActionType = "escalate"   // Escalate to human operator
)

// ExecutionStrategy determines how tasks in a plan are executed.
type ExecutionStrategy string

const (
    StrategySequential ExecutionStrategy = "sequential" // Tasks run one at a time
    StrategyParallel   ExecutionStrategy = "parallel"   // Independent tasks run concurrently
    StrategyHybrid     ExecutionStrategy = "hybrid"     // Mix of sequential and parallel
)

// PlanSource indicates how a plan was created.
type PlanSource string

const (
    PlanSourceTemplate PlanSource = "template" // From knowledge store template (0ms)
    PlanSourceRules    PlanSource = "rules"    // From decision engine rules (fast)
    PlanSourceAI       PlanSource = "ai"       // From AI provider (fallback, slow)
)
```

```go
// contracts/brain/decision.go — Decision Engine
package brain

import (
    "context"
    "time"
)

// Situation describes the current state for the brain to evaluate.
type Situation struct {
    // TaskType is the category (e.g., "code_generation", "code_review").
    TaskType string `json:"task_type"`
    // TaskName is the specific task name for granular matching.
    TaskName string `json:"task_name,omitempty"`
    // FailCount is how many times this task has failed.
    FailCount int `json:"fail_count"`
    // LastError is the error from the most recent failure.
    LastError string `json:"last_error,omitempty"`
    // Available lists executor names currently available.
    Available []string `json:"available"`
    // History contains past decisions for similar situations.
    History []PastDecision `json:"history,omitempty"`
    // Context carries additional key-value metadata.
    Context map[string]any `json:"context,omitempty"`
}

// Decision is the brain's output — what action to take.
type Decision struct {
    // Action is the type of action.
    Action ActionType `json:"action"`
    // Target is the executor name (empty for Abort/Skip).
    Target string `json:"target,omitempty"`
    // Params carries action-specific parameters.
    Params map[string]any `json:"params,omitempty"`
    // Reasoning explains the decision (for audit logs).
    Reasoning string `json:"reasoning"`
    // Confidence is the brain's confidence (0.0-1.0).
    Confidence float64 `json:"confidence"`
}

// PastDecision records a previous decision for learning.
type PastDecision struct {
    Action    ActionType    `json:"action"`
    Target    string        `json:"target"`
    TaskType  string        `json:"task_type"`
    Success   bool          `json:"success"`
    Duration  time.Duration `json:"duration"`
    Timestamp time.Time     `json:"timestamp"`
}

// Rule is a deterministic decision rule.
// Rules are evaluated in priority order (lower = higher priority).
type Rule struct {
    // Name identifies this rule (for logging).
    Name string `json:"name"`
    // Priority determines evaluation order (lower = first). Range: 0-1000.
    Priority int `json:"priority"`
    // Description explains the rule (for docs).
    Description string `json:"description,omitempty"`
    // Condition evaluates whether this rule matches.
    // MUST be a pure function (no I/O, no side effects).
    Condition func(Situation) bool `json:"-"`
    // Action to take when matched.
    Action ActionType `json:"action"`
    // TargetSelector determines executor. Nil = first available.
    TargetSelector func(Situation) string `json:"-"`
    // Params are static parameters for the decision.
    Params map[string]any `json:"params,omitempty"`
}

// Strategy selects the best executor from available options.
type Strategy interface {
    Select(taskType string, available []string, history []PastDecision) string
    Name() string
}

// DecisionEngine evaluates situations via rules and strategies.
// Thread-safe: Evaluate may be called concurrently.
type DecisionEngine interface {
    // Evaluate processes a situation through rules (priority order).
    // First matching rule wins. If none match, default decision returned.
    // Never returns nil Decision on success.
    Evaluate(ctx context.Context, situation Situation) (*Decision, error)
    // RegisterRule adds a rule. Duplicate names rejected.
    RegisterRule(rule Rule) error
    // Rules returns all rules sorted by priority.
    Rules() []Rule
}
```

### Planning Engine (Fix for Issues A2, B4)

```go
// contracts/brain/planning.go — Planning Engine
package brain

import (
    "context"
    
    "github.com/tiendat1751998/orchestrator/contracts/mission"
)

// TaskSpec is a lightweight task description created by the planner.
// The kernel maps TaskSpec → agent.Task when dispatching for execution.
//
// Why not agent.Task directly?
// Brain defines WHAT to do (task spec). Agent defines HOW to do it.
// Brain should not know about agent-specific fields (result, retry config, etc.).
type TaskSpec struct {
    // ID uniquely identifies this task within the plan.
    ID string `json:"id"`
    // Name is a short description.
    Name string `json:"name"`
    // Type categorizes the task (e.g., "code_generation", "testing").
    Type string `json:"type"`
    // Description is the full task specification.
    Description string `json:"description"`
    // DependsOn lists task IDs that must complete before this task starts.
    DependsOn []string `json:"depends_on,omitempty"`
    // Params carries task-specific parameters.
    Params map[string]any `json:"params,omitempty"`
    // EstimatedDuration is the planner's time estimate.
    EstimatedDuration string `json:"estimated_duration,omitempty"`
    // Critical indicates if failure should abort the mission.
    Critical bool `json:"critical"`
}

// Plan is the output of planning — a DAG of task specs.
type Plan struct {
    // ID uniquely identifies this plan.
    ID string `json:"id"`
    // MissionID links to the mission this plan serves.
    MissionID string `json:"mission_id"`
    // Tasks are all tasks in the plan.
    Tasks []TaskSpec `json:"tasks"`
    // Strategy determines execution order (sequential, parallel, hybrid).
    Strategy ExecutionStrategy `json:"strategy"`
    // Source indicates how the plan was created (template, rules, AI).
    Source PlanSource `json:"source"`
    // TemplateID is set when plan was created from a template.
    TemplateID string `json:"template_id,omitempty"`
}

// PlanTemplate is a reusable plan blueprint stored in Knowledge.
type PlanTemplate struct {
    // ID uniquely identifies this template.
    ID string `json:"id"`
    // Name is a short description (e.g., "REST API Project").
    Name string `json:"name"`
    // Description explains when to use this template.
    Description string `json:"description"`
    // Tags for matching (e.g., ["rest", "api", "go"]).
    Tags []string `json:"tags"`
    // TaskSpecs are the template's tasks (with placeholder params).
    TaskSpecs []TaskSpec `json:"task_specs"`
    // Strategy is the default execution strategy.
    Strategy ExecutionStrategy `json:"strategy"`
    // Params defines required template parameters and their defaults.
    // Example: {"project_name": "", "language": "go"}
    Params map[string]any `json:"params"`
    // Score is the template's confidence (0-1), updated by learning.
    Score float64 `json:"score"`
}

// TaskFailure describes why a task failed (for replanning).
type TaskFailure struct {
    // TaskID is the failed task's ID.
    TaskID string `json:"task_id"`
    // Error is the failure message.
    Error string `json:"error"`
    // Attempt is which attempt failed (1-based).
    Attempt int `json:"attempt"`
}

// PlanningEngine decomposes missions into task DAGs.
// Strategy: template-first → rule-based → AI fallback.
type PlanningEngine interface {
    // Plan creates a plan for a mission.
    // Tries templates first, then rules, then AI (last resort).
    Plan(ctx context.Context, m *mission.Mission) (*Plan, error)
    // PlanFromTemplate creates a plan from a specific template.
    // Fully deterministic — no AI calls, executes in <1ms.
    PlanFromTemplate(ctx context.Context, templateID string, params map[string]any) (*Plan, error)
    // Replan creates a revised plan after task failure.
    // Uses deterministic rules (not AI) for recovery decisions.
    Replan(ctx context.Context, plan *Plan, failure TaskFailure) (*Plan, error)
    // RegisterTemplate adds a plan template.
    RegisterTemplate(template PlanTemplate) error
    // Templates returns all registered templates.
    Templates() []PlanTemplate
}
```

### Policy Engine

```go
// contracts/brain/policy.go — Policy Engine
package brain

import "context"

// PolicyAction describes an action being attempted.
type PolicyAction struct {
    // Actor is who is attempting (agent name, user).
    Actor string `json:"actor"`
    // Action is what they're doing (execute_shell, write_file, etc.).
    Action string `json:"action"`
    // Resource is what they're acting on (file path, URL, etc.).
    Resource string `json:"resource"`
    // Context carries additional metadata.
    Context map[string]any `json:"context,omitempty"`
}

// PolicyResult is the outcome of policy evaluation.
type PolicyResult struct {
    // Allowed indicates if the action is permitted.
    Allowed bool `json:"allowed"`
    // Reason explains the decision.
    Reason string `json:"reason"`
    // EvaluatedPolicies lists which policies were checked.
    EvaluatedPolicies []string `json:"evaluated_policies"`
}

// Policy is a named security/resource policy.
type Policy struct {
    // Name identifies this policy.
    Name string `json:"name"`
    // Priority determines evaluation order (lower = first).
    Priority int `json:"priority"`
    // Description explains the policy.
    Description string `json:"description,omitempty"`
    // Evaluate checks if the action is allowed.
    // MUST be deterministic (no I/O).
    Evaluate func(PolicyAction) PolicyResult `json:"-"`
}

// PolicyEngine enforces security, resource, and operational policies.
type PolicyEngine interface {
    Enforce(ctx context.Context, action PolicyAction) (*PolicyResult, error)
    RegisterPolicy(policy Policy) error
    Policies() []Policy
}
```

### Context Engine (Fix for Issue B3)

```go
// contracts/brain/context.go — Context Engine
package brain

import "context"

// ContextSourceType identifies where context items come from.
type ContextSourceType string

const (
    SourceCode      ContextSourceType = "code"      // Source code files
    SourceDoc       ContextSourceType = "doc"        // Documentation
    SourceKnowledge ContextSourceType = "knowledge"  // Knowledge graph nodes
    SourceHistory   ContextSourceType = "history"    // Past decisions/outcomes
    SourceArtifact  ContextSourceType = "artifact"   // Previous mission artifacts
    SourceMemory    ContextSourceType = "memory"     // Working memory items
)

// ContextSource specifies where to pull context from.
type ContextSource struct {
    // Type is the source category.
    Type ContextSourceType `json:"type"`
    // Filter narrows the source (e.g., file glob, tag list, query).
    Filter string `json:"filter,omitempty"`
    // MaxItems limits items from this source.
    MaxItems int `json:"max_items,omitempty"`
}

// ContextRequest describes what context to assemble.
type ContextRequest struct {
    // TaskType is the kind of task needing context.
    TaskType string `json:"task_type"`
    // TaskDescription helps the ranker determine relevance.
    TaskDescription string `json:"task_description,omitempty"`
    // MaxTokens is the token budget for the assembled context.
    MaxTokens int `json:"max_tokens"`
    // Sources specifies where to pull context from (ordered by priority).
    Sources []ContextSource `json:"sources"`
}

// AssembledContext is the output — ranked, compressed context.
type AssembledContext struct {
    // Items are the context items, ranked by relevance.
    Items []ContextItem `json:"items"`
    // TotalTokens is the actual token count used.
    TotalTokens int `json:"total_tokens"`
    // Truncated indicates whether any items were truncated.
    Truncated bool `json:"truncated"`
    // SourcesSummary lists which sources contributed.
    SourcesSummary map[ContextSourceType]int `json:"sources_summary"`
}

// ContextItem is a single piece of context.
type ContextItem struct {
    // Source identifies where this came from.
    Source ContextSourceType `json:"source"`
    // Reference is a specific identifier (file path, node ID, etc.).
    Reference string `json:"reference,omitempty"`
    // Content is the actual text.
    Content string `json:"content"`
    // Tokens is the token count.
    Tokens int `json:"tokens"`
    // Score is the relevance score (0-1).
    Score float64 `json:"score"`
}

// ContextEngine assembles optimal context for AI provider calls.
type ContextEngine interface {
    // Assemble builds context from multiple sources, ranked and compressed.
    Assemble(ctx context.Context, req ContextRequest) (*AssembledContext, error)
}
```

### Cognitive Layer (Phase 5+)

```go
// contracts/brain/cognitive.go — Cognitive Layer contracts
package brain

import "context"

// MissionOutcome is the input for reflection — what happened during a mission.
type MissionOutcome struct {
    MissionID    string     `json:"mission_id"`
    Success      bool       `json:"success"`
    TaskOutcomes []TaskOutcome `json:"task_outcomes"`
    TotalDuration string    `json:"total_duration"`
}

type TaskOutcome struct {
    TaskID   string  `json:"task_id"`
    TaskType string  `json:"task_type"`
    Success  bool    `json:"success"`
    Agent    string  `json:"agent"`
    Quality  float64 `json:"quality"`
    Error    string  `json:"error,omitempty"`
}

// Reflection is the output of self-review.
type Reflection struct {
    Strengths   []string `json:"strengths"`
    Weaknesses  []string `json:"weaknesses"`
    Improvements []string `json:"improvements"`
    NewPatterns []string `json:"new_patterns"`
}

// Prediction estimates likely outcomes.
type Prediction struct {
    SuccessProbability float64            `json:"success_probability"`
    EstimatedDuration  string             `json:"estimated_duration"`
    RiskFactors        []string           `json:"risk_factors"`
    Recommendations    []string           `json:"recommendations"`
}

// OptimizationResult describes what was optimized.
type OptimizationResult struct {
    RulesUpdated     int `json:"rules_updated"`
    StrategiesAdjusted int `json:"strategies_adjusted"`
    TemplatesCreated int `json:"templates_created"`
    Notes            []string `json:"notes"`
}

// ReflectionEngine performs post-mission self-review.
type ReflectionEngine interface {
    Reflect(ctx context.Context, outcome MissionOutcome) (*Reflection, error)
}

// LearningEngine learns from outcomes to improve future decisions.
type LearningEngine interface {
    Learn(ctx context.Context, outcome MissionOutcome) error
    UpdateScores(ctx context.Context) error
}

// PredictionEngine estimates likely outcomes before execution.
type PredictionEngine interface {
    Predict(ctx context.Context, plan *Plan) (*Prediction, error)
}

// OptimizationEngine tunes strategies based on accumulated data.
type OptimizationEngine interface {
    Optimize(ctx context.Context) (*OptimizationResult, error)
}
```

### Implementation Phases

| Engine | Contracts | Kernel Implementation | Phase |
|---|---|---|---|
| Decision Engine | `contracts/brain/decision.go` | `kernel/brain/decision/` | Phase 1-2 |
| Planning Engine | `contracts/brain/planning.go` | `kernel/brain/planning/` | Phase 1-2 |
| Policy Engine | `contracts/brain/policy.go` | `kernel/brain/policy/` | Phase 1-2 |
| Context Engine | `contracts/brain/context.go` | `kernel/brain/context/` | Phase 2-3 |
| Learning Engine | `contracts/brain/cognitive.go` | `kernel/brain/cognitive/` | Phase 3+ |
| Reflection Engine | `contracts/brain/cognitive.go` | `kernel/brain/cognitive/` | Phase 5+ |
| Prediction Engine | `contracts/brain/cognitive.go` | `kernel/brain/cognitive/` | Phase 5+ |
| Optimization Engine | `contracts/brain/cognitive.go` | `kernel/brain/cognitive/` | Phase 5+ |

**Rule**: All interfaces defined in Phase 1 (contracts). Implementations phased by need.

### Post-Mission Cognitive Loop (Phase 5+)

```
Mission Completed
       ↓
  Reflection      ← "What went well? What failed? Why?"
       ↓
  Learning        ← "Update agent scores. Record patterns."
       ↓
  Knowledge Update ← "Store new templates. Update facts."
       ↓
  Optimization    ← "Tune rule weights. Adjust strategies."
       ↓
  Policy Update   ← "Tighten/loosen policies based on data."
       ↓
     Done         ← System is now smarter for next mission
```

## Impact

### Replaces
- `contracts/planner/` → absorbed into `contracts/brain/planning.go`
- `contracts/orchestrator/` → orchestration distributed across Brain + Execution

### New Packages
- `contracts/brain/` — All brain interfaces (flat package, multiple files)
- `contracts/mission/` — Mission type definition

### Layer Compliance
- `contracts/brain/` imports only stdlib + `contracts/mission/` ✅
- `contracts/mission/` imports only stdlib ✅
- `kernel/brain/` imports only `contracts/` ✅

## Resolved Questions

1. ~~**Brain as facade vs direct engine access**~~ **RESOLVED**: Brain facade for external callers. Sub-engine accessors for kernel-internal use.

2. ~~**Context Engine scope**~~ **RESOLVED**: Single Context Engine in Brain Runtime. Execution Runtime accesses via interface.

3. ~~**Flat vs sub-packages**~~ **RESOLVED**: Flat `package brain` in contracts (multiple files). Sub-packages in kernel implementation.

4. ~~**Plan.Tasks type**~~ **RESOLVED**: `brain.TaskSpec` (lightweight). Kernel maps TaskSpec → agent.Task.
