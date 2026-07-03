// Package planner defines the locked contract for plan generation, scoring, and learning.
package planner

import (
	"context"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/fsm"
	"github.com/tiendat1751998/orchestrator/contracts/goal"
)

// Planner defines the locked contract for plan generation and optimization.
type Planner interface {
	// Plan generates candidate plan DAGs satisfying the goals.
	// It applies CSP constraint propagation to prune the search space.
	Plan(ctx context.Context, g goal.Goal) ([]fsm.DAG, error)

	// Score evaluates candidate plans mathematically using multi-objective Pareto Frontier calculations.
	Score(ctx context.Context, candidates []fsm.DAG) (fsm.DAG, error)

	// Explain generates the contrastive mathematical reasoning report detailing why the plan was chosen.
	Explain(ctx context.Context, chosen fsm.DAG, candidates []fsm.DAG) (string, error)

	// Learn updates template weights and failure association edges based on transition results.
	Learn(ctx context.Context, history fsm.TransitionRecord) error
}

// Mission represents a user's high-level goal.
// It is the top-level unit of work in the system.
type Mission struct {
	// ID uniquely identifies this mission.
	ID string `json:"id"`
	// Title is a short human-readable summary.
	Title string `json:"title"`
	// Description is the full mission specification.
	Description string `json:"description"`
	// Constraints are limitations (e.g., "no external APIs", "Go only").
	Constraints []string `json:"constraints,omitempty"`
	// Metadata carries additional key-value data.
	Metadata map[string]string `json:"metadata,omitempty"`
	// CreatedAt is when the mission was submitted.
	CreatedAt time.Time `json:"created_at"`
}
