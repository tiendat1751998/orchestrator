// Package brain defines contracts for the deterministic decision and cognitive engines.
package brain

import (
	"context"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/fsm"
	// ponytail: goal import omitted as it is unused in the contracts defined here.
)

// ActionType represents the type of action the brain can decide.
type ActionType string

const (
	ActionDispatch ActionType = "dispatch"
	ActionRetry    ActionType = "retry"
	ActionFallback ActionType = "fallback"
	ActionSkip     ActionType = "skip"
	ActionAbort    ActionType = "abort"
	ActionWait     ActionType = "wait"
	ActionEscalate ActionType = "escalate"
)

// PastDecision records a previous decision for learning.
type PastDecision struct {
	Action    ActionType    `json:"action"`
	Target    string        `json:"target"`
	TaskType  string        `json:"task_type"`
	Success   bool          `json:"success"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
}

// Situation describes the current state for the brain to evaluate.
type Situation struct {
	TaskType  string         `json:"task_type"`
	TaskName  string         `json:"task_name,omitempty"`
	FailCount int            `json:"fail_count"`
	LastError string         `json:"last_error,omitempty"`
	Available []string       `json:"available"`
	History   []PastDecision `json:"history,omitempty"`
	Context   map[string]any `json:"context,omitempty"`
}

// Decision is the output of the brain's evaluation.
type Decision struct {
	Action     ActionType     `json:"action"`
	Target     string         `json:"target,omitempty"`
	Params     map[string]any `json:"params,omitempty"`
	Reasoning  string         `json:"reasoning"`
	Confidence float64        `json:"confidence"`
}

// DecisionEngine evaluates situations and produces decisions.
type DecisionEngine interface {
	Evaluate(ctx context.Context, situation Situation) (*Decision, error)
}

// AgentConfidence represents the capability rating of an agent.
type AgentConfidence struct {
	AgentID    string  `json:"agent_id"`
	Capability string  `json:"capability"`
	Score      float64 `json:"score"`
}

// ConfidenceEngine evaluates plan safety thresholds (RFC-0033).
type ConfidenceEngine interface {
	Evaluate(ctx context.Context, dag fsm.DAG, threshold float64) (fsm.DAG, error)
	GetConfidence(ctx context.Context, agentID, capability string) (*AgentConfidence, error)
}

// CompetencyRating represents the capability weight.
type CompetencyRating struct {
	AgentID     string  `json:"agent_id"`
	Capability  string  `json:"capability"`
	DomainSkill string  `json:"domain_skill"`
	Rating      float64 `json:"rating"`
}

// CapabilityGraph maps agent capabilities to skill nodes (RFC-0035).
type CapabilityGraph interface {
	AddCompetency(ctx context.Context, rating CompetencyRating) error
	RouteTask(ctx context.Context, capability string, skill string) (string, float64, error)
}

// SimulationResult represents the pre-flight audit report.
type SimulationResult struct {
	Valid         bool     `json:"valid"`
	ExpectedCost  float64  `json:"expected_cost_usd"`
	ExpectedTimeS int      `json:"expected_time_seconds"`
	Deadlocks     []string `json:"deadlocks,omitempty"`
	CircularDeps  []string `json:"circular_dependencies,omitempty"`
}

// MissionSimulator simulates FSM transitions on a plan DAG (RFC-0036).
type MissionSimulator interface {
	Simulate(ctx context.Context, dag fsm.DAG) (*SimulationResult, error)
}

// FailureContext contains metadata about the task failure.
type FailureContext struct {
	TaskID    string `json:"task_id"`
	Category  string `json:"category"`
	ErrorLog  string `json:"error_log"`
	Workspace string `json:"workspace_hash"`
}

// Replanner mutates plan DAGs on failure (RFC-0037).
type Replanner interface {
	Replan(ctx context.Context, fCtx FailureContext, activeDAG fsm.DAG) (fsm.DAG, error)
}

// ProjectROI represents the estimated value return.
type ProjectROI struct {
	DeveloperHoursSaved float64 `json:"developer_hours_saved"`
	LatencyImprovement  float64 `json:"latency_improvement_pct"`
	BugRiskReduction    float64 `json:"bug_risk_reduction"`
}

// EconomicEngine calculates plan business utility (RFC-0044).
type EconomicEngine interface {
	CalculateROI(ctx context.Context, steps []string, techStack []string) (*ProjectROI, error)
	ExpectedUtility(ctx context.Context, roi ProjectROI, projectedCost float64) float64
}

// VirtualEmployee represents a configured agent in the system.
type VirtualEmployee struct {
	ID                   string   `json:"id"`
	Role                 string   `json:"role"`
	ModelName            string   `json:"model_name"`
	VirtualSalaryPerHour float64  `json:"virtual_salary_per_hour"`
	ReliabilityScore     float64  `json:"reliability_score"`
	Capabilities         []string `json:"capabilities"`
}

// WorkforceManager coordinates virtual employees (RFC-0045).
type WorkforceManager interface {
	RegisterEmployee(ctx context.Context, emp VirtualEmployee) error
	GetBestFitAgent(ctx context.Context, roleRequired string, budgetMax float64) (*VirtualEmployee, error)
	UpdateReliability(ctx context.Context, employeeID string, success bool) error
}

// ProviderTrust represents the audited reliability of a model.
type ProviderTrust struct {
	ModelName   string  `json:"model_name"`
	TaskType    string  `json:"task_type"`
	SuccessRuns int     `json:"success_runs"`
	TotalRuns   int     `json:"total_runs"`
	TrustRating float64 `json:"trust_rating"`
}

// TrustEngine updates and queries provider trust ratings (RFC-0056).
type TrustEngine interface {
	AuditRecord(ctx context.Context, modelName string, taskType string, success bool) error
	GetTrustRating(ctx context.Context, modelName string, taskType string) (*ProviderTrust, error)
}
