package planner_test

import (
	"context"
	"math"
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts/fsm"
	"github.com/tiendat1751998/orchestrator/contracts/goal"
	"github.com/tiendat1751998/orchestrator/contracts/knowledge"
	"github.com/tiendat1751998/orchestrator/kernel/planner"
)

// mockSkillGraph implements knowledge.SkillGraph for testing purposes.
type mockSkillGraph struct {
	skills []knowledge.SkillNode
	edges  []knowledge.SkillEdge
}

func (m *mockSkillGraph) AddSkill(ctx context.Context, skill knowledge.SkillNode) error {
	m.skills = append(m.skills, skill)
	return nil
}

func (m *mockSkillGraph) LinkSkills(ctx context.Context, edge knowledge.SkillEdge) error {
	m.edges = append(m.edges, edge)
	return nil
}

func (m *mockSkillGraph) ResolveDependencies(ctx context.Context, skillID string) ([]knowledge.SkillNode, error) {
	// Simple lookup helper for testing dependencies if needed
	var result []knowledge.SkillNode
	for _, edge := range m.edges {
		if edge.FromID == skillID {
			for _, skill := range m.skills {
				if skill.ID == edge.ToID {
					result = append(result, skill)
				}
			}
		}
	}
	return result, nil
}

func TestCSPSolver_Filter(t *testing.T) {
	// 1. Mock SkillGraph and register tech nodes (Postgres, SQLite, Go, OpenAI).
	g := &mockSkillGraph{}
	ctx := context.Background()

	postgresNode := knowledge.SkillNode{
		ID:          "postgres",
		Name:        "PostgreSQL",
		Category:    "database",
		Description: "Relational database",
	}
	sqliteNode := knowledge.SkillNode{
		ID:          "sqlite",
		Name:        "SQLite",
		Category:    "database",
		Description: "Embedded database",
	}
	goNode := knowledge.SkillNode{
		ID:          "go",
		Name:        "Go",
		Category:    "language",
		Description: "System programming",
	}
	pythonNode := knowledge.SkillNode{
		ID:          "python",
		Name:        "Python",
		Category:    "language",
		Description: "Scripting language",
	}
	openAINode := knowledge.SkillNode{
		ID:          "openai",
		Name:        "OpenAI",
		Category:    "cloud",
		Description: "Cloud LLM",
	}

	_ = g.AddSkill(ctx, postgresNode)
	_ = g.AddSkill(ctx, sqliteNode)
	_ = g.AddSkill(ctx, goNode)
	_ = g.AddSkill(ctx, pythonNode)
	_ = g.AddSkill(ctx, openAINode)

	solver := planner.NewCSPSolver(g)
	available := []knowledge.SkillNode{postgresNode, sqliteNode, goNode, pythonNode, openAINode}

	t.Run("DatabaseOnlyConstraint", func(t *testing.T) {
		// 2. Configure a "database_only: sqlite" constraint.
		constraints := []goal.Constraint{
			{
				Type:  "database_only",
				Value: "sqlite",
			},
		}

		// 3. Verify that the CSP filter successfully prunes the Postgres node.
		filtered, err := solver.Filter(ctx, constraints, available)
		if err != nil {
			t.Fatalf("Filter returned unexpected error: %v", err)
		}

		var hasSQLite, hasPostgres, hasGo bool
		for _, node := range filtered {
			switch node.ID {
			case "sqlite":
				hasSQLite = true
			case "postgres":
				hasPostgres = true
			case "go":
				hasGo = true
			}
		}

		if !hasSQLite {
			t.Errorf("expected sqlite node to be kept")
		}
		if hasPostgres {
			t.Errorf("expected postgres node to be pruned")
		}
		if !hasGo {
			t.Errorf("expected go node to be kept (non-database node should not be affected by database_only)")
		}
	})

	t.Run("LanguageConstraint", func(t *testing.T) {
		constraints := []goal.Constraint{
			{
				Type:  "language",
				Value: "go",
			},
		}

		filtered, err := solver.Filter(ctx, constraints, available)
		if err != nil {
			t.Fatalf("Filter returned unexpected error: %v", err)
		}

		var hasGo, hasPython, hasSQLite bool
		for _, node := range filtered {
			switch node.ID {
			case "go":
				hasGo = true
			case "python":
				hasPython = true
			case "sqlite":
				hasSQLite = true
			}
		}

		if !hasGo {
			t.Errorf("expected go node to be kept")
		}
		if hasPython {
			t.Errorf("expected python node to be pruned")
		}
		if !hasSQLite {
			t.Errorf("expected sqlite node to be kept (non-language node should not be affected by language)")
		}
	})

	t.Run("OfflineOnlyConstraint", func(t *testing.T) {
		constraints := []goal.Constraint{
			{
				Type:  "offline_only",
				Value: "true",
			},
		}

		filtered, err := solver.Filter(ctx, constraints, available)
		if err != nil {
			t.Fatalf("Filter returned unexpected error: %v", err)
		}

		var hasOpenAI, hasSQLite bool
		for _, node := range filtered {
			switch node.ID {
			case "openai":
				hasOpenAI = true
			case "sqlite":
				hasSQLite = true
			}
		}

		if hasOpenAI {
			t.Errorf("expected openai node to be pruned due to offline_only constraint")
		}
		if !hasSQLite {
			t.Errorf("expected sqlite node to be kept")
		}
	})

	t.Run("NoConstraints", func(t *testing.T) {
		filtered, err := solver.Filter(ctx, nil, available)
		if err != nil {
			t.Fatalf("Filter returned unexpected error: %v", err)
		}
		if len(filtered) != len(available) {
			t.Errorf("expected %d nodes, got %d", len(available), len(filtered))
		}
	})

	t.Run("ContextCancelled", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := solver.Filter(cancelCtx, []goal.Constraint{{Type: "database_only", Value: "sqlite"}}, available)
		if err == nil {
			t.Fatal("expected error due to cancelled context, got nil")
		}
	})

	t.Run("NilContext", func(t *testing.T) {
		filtered, err := solver.Filter(nil, nil, available) //nolint:staticcheck
		if err != nil {
			t.Fatalf("Filter returned unexpected error on nil context: %v", err)
		}
		if len(filtered) != len(available) {
			t.Errorf("expected %d nodes, got %d", len(available), len(filtered))
		}
	})
}

func TestScorer_ParetoAndUCB(t *testing.T) {
	// 1. Configure weights for Quality (+), Cost (-), and Risk (-).
	// Quality weight is positive, Cost weight is positive but (1.0-c) is used, Risk weight is positive but subtracted.
	weights := planner.Weights{
		Quality:    0.5,
		Cost:       0.3,
		Time:       0.0,
		Confidence: 0.0,
		Risk:       0.2,
	}
	cFactor := 1.5
	scorer := planner.NewScorer(weights, cFactor)
	ctx := context.Background()

	// 2. Score two candidate DAGs (one cheap/low-success, one premium/high-success).
	dagCheap := fsm.DAG{
		Nodes: map[string]*fsm.DAGNode{
			"task1": {
				ID:           "task1",
				Dependencies: nil,
				Status:       "pending",
			},
		},
	}
	dagPremium := fsm.DAG{
		Nodes: map[string]*fsm.DAGNode{
			"task1": {
				ID:           "task1",
				Dependencies: nil,
				Status:       "pending",
			},
			"task2": {
				ID:           "task2",
				Dependencies: []string{"task1"},
				Status:       "pending",
			},
		},
	}

	// Cheap / Low-success: Quality: 0.3, Cost: 0.1, Risk: 0.7
	// Pareto: 0.5*0.3 + 0.3*(1-0.1) - 0.2*0.7 = 0.15 + 0.27 - 0.14 = 0.28
	// UCB Bonus (runs=100, usage=10): 1.5 * sqrt(ln(100)/10) = 1.5 * sqrt(4.60517/10) = 1.5 * 0.678614 = 1.017921
	// Total Cheap Score: 0.28 + 1.017921 = 1.297921
	scoreCheap := scorer.ScoreCandidate(ctx, dagCheap, 0.3, 0.1, 0.0, 0.0, 0.7, 100, 10)
	expectedCheap := 0.28 + 1.0179210637267784

	if math.Abs(scoreCheap-expectedCheap) > 1e-9 {
		t.Errorf("Expected cheap score %f, got %f", expectedCheap, scoreCheap)
	}

	// Premium / High-success: Quality: 0.9, Cost: 0.8, Risk: 0.1
	// Pareto: 0.5*0.9 + 0.3*(1-0.8) - 0.2*0.1 = 0.45 + 0.06 - 0.02 = 0.49
	// UCB Bonus (runs=100, usage=10): 1.0179210637267784
	// Total Premium Score: 0.49 + 1.017921 = 1.507921
	scorePremium := scorer.ScoreCandidate(ctx, dagPremium, 0.9, 0.8, 0.0, 0.0, 0.1, 100, 10)
	expectedPremium := 0.49 + 1.0179210637267784

	if math.Abs(scorePremium-expectedPremium) > 1e-9 {
		t.Errorf("Expected premium score %f, got %f", expectedPremium, scorePremium)
	}

	if scorePremium <= scoreCheap {
		t.Errorf("Expected premium plan score (%f) to be higher than cheap plan score (%f)", scorePremium, scoreCheap)
	}

	// 3. Verify that the UCB-1 bonus is applied correctly when usage counts are low.
	t.Run("UCBBonusZeroUsage", func(t *testing.T) {
		// If usageCount is 0, bonus is 1.0.
		// Pareto cheap base: 0.28. Expected: 1.28.
		scoreZeroUsage := scorer.ScoreCandidate(ctx, dagCheap, 0.3, 0.1, 0.0, 0.0, 0.7, 100, 0)
		expectedZeroUsage := 0.28 + 1.0
		if math.Abs(scoreZeroUsage-expectedZeroUsage) > 1e-9 {
			t.Errorf("Expected score with 0 usage to be %f, got %f", expectedZeroUsage, scoreZeroUsage)
		}
	})

	t.Run("UCBBonusLowVsHighUsage", func(t *testing.T) {
		// Low usage (1 run out of 100) vs High usage (100 runs out of 100)
		// Low usage: UCB Bonus = 1.5 * sqrt(ln(100)/1) = 1.5 * sqrt(4.60517018) = 1.5 * 2.145966 = 3.218949
		// High usage: UCB Bonus = 1.5 * sqrt(ln(100)/100) = 1.5 * sqrt(0.0460517) = 1.5 * 0.214596 = 0.321894
		scoreLowUsage := scorer.ScoreCandidate(ctx, dagCheap, 0.3, 0.1, 0.0, 0.0, 0.7, 100, 1)
		scoreHighUsage := scorer.ScoreCandidate(ctx, dagCheap, 0.3, 0.1, 0.0, 0.0, 0.7, 100, 100)

		bonusLow := scoreLowUsage - 0.28
		bonusHigh := scoreHighUsage - 0.28

		if bonusLow <= bonusHigh {
			t.Errorf("Expected low usage bonus (%f) to be higher than high usage bonus (%f)", bonusLow, bonusHigh)
		}

		expectedBonusLow := 1.5 * math.Sqrt(math.Log(100.0)/1.0)
		expectedBonusHigh := 1.5 * math.Sqrt(math.Log(100.0)/100.0)

		if math.Abs(bonusLow-expectedBonusLow) > 1e-9 {
			t.Errorf("Expected low usage bonus to be %f, got %f", expectedBonusLow, bonusLow)
		}
		if math.Abs(bonusHigh-expectedBonusHigh) > 1e-9 {
			t.Errorf("Expected high usage bonus to be %f, got %f", expectedBonusHigh, bonusHigh)
		}
	})
}
