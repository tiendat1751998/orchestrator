// Package goal defines the contracts for goal specification and target parsing.
package goal

import (
	"context"
)

// Objective represents a sub-goal in the mission.
type Objective struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	DependsOn   []string `json:"depends_on,omitempty"`
}

// Constraint represents a budget, tech-stack, or security limit.
type Constraint struct {
	Type  string `json:"type"`  // e.g. "budget_usd", "offline_only", "language", "framework"
	Value string `json:"value"` // e.g. "30.0", "true", "go", "gin"
}

// Goal represents the formalized target contract.
type Goal struct {
	RawInput           string       `json:"raw_input"`
	Objectives         []Objective  `json:"objectives"`
	Constraints        []Constraint `json:"constraints"`
	AcceptanceCriteria []string     `json:"acceptance_criteria"`
	Milestones         []string     `json:"milestones"`
}

// GoalEngine translates raw user strings to structured Goals.
type GoalEngine interface {
	// Parse translates raw user input into the structured Goal model.
	Parse(ctx context.Context, input string) (*Goal, error)
}
