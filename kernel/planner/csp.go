package planner

import (
	"context"
	"strings"

	"github.com/tiendat1751998/orchestrator/contracts/goal"
	"github.com/tiendat1751998/orchestrator/contracts/knowledge"
)

// CSPSolver prunes technology search spaces based on hard constraints.
type CSPSolver struct {
	graph knowledge.SkillGraph
}

// NewCSPSolver constructs a new solver.
func NewCSPSolver(graph knowledge.SkillGraph) *CSPSolver {
	return &CSPSolver{graph: graph}
}

// Filter prunes the available tech nodes that violate goal constraints.
// It handles database, language, and resource (offline/cloud) limits.
func (c *CSPSolver) Filter(ctx context.Context, constraints []goal.Constraint, available []knowledge.SkillNode) ([]knowledge.SkillNode, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if len(constraints) == 0 {
		// ponytail: return input directly if no constraints
		return available, nil
	}

	filtered := make([]knowledge.SkillNode, 0, len(available))

	for _, node := range available {
		valid := true
		for _, cons := range constraints {
			// 1. Database constraint pruning
			if cons.Type == "database_only" {
				isDbNode := strings.EqualFold(node.Category, "database") ||
					strings.EqualFold(node.ID, "postgres") ||
					strings.EqualFold(node.ID, "sqlite") ||
					strings.EqualFold(node.ID, "mysql") ||
					strings.EqualFold(node.ID, "mongodb")
				if isDbNode && !strings.EqualFold(node.ID, cons.Value) {
					valid = false
					break
				}
			}

			// 2. Language constraint pruning
			if cons.Type == "language" {
				isLangNode := strings.EqualFold(node.Category, "language") ||
					strings.EqualFold(node.ID, "go") ||
					strings.EqualFold(node.ID, "python") ||
					strings.EqualFold(node.ID, "javascript") ||
					strings.EqualFold(node.ID, "typescript") ||
					strings.EqualFold(node.ID, "rust")
				if isLangNode && !strings.EqualFold(node.ID, cons.Value) {
					valid = false
					break
				}
			}

			// 3. Resource/environment limits (e.g. offline-only/no cloud)
			if cons.Type == "offline_only" && (strings.EqualFold(cons.Value, "true") || strings.EqualFold(cons.Value, "yes")) {
				isCloudNode := strings.EqualFold(node.Category, "cloud") ||
					strings.EqualFold(node.ID, "openai") ||
					strings.EqualFold(node.ID, "anthropic") ||
					strings.EqualFold(node.ID, "aws") ||
					strings.EqualFold(node.ID, "gcp") ||
					strings.EqualFold(node.ID, "azure")
				if isCloudNode {
					valid = false
					break
				}
			}
		}

		if valid {
			filtered = append(filtered, node)
		}
	}

	return filtered, nil
}
