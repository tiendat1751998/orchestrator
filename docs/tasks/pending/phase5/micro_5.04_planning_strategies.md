# Micro-Task 5.04: Create kernel/planner/pareto.go

- **File**: `kernel/planner/pareto.go`
- **Package**: `planner`
- **Depends on**: 5.01, 1.41 (brain/brain.go)
- **Time**: 25 min
- **Verify**: `go build ./kernel/planner/...`

## Purpose
Implements the multi-objective **Pareto Frontier Scoring** and **UCB-1 Exploration** functions. It grades candidate plan DAGs mathematically to select the option that maximizes quality, confidence, and exploration, while minimizing cost, duration, and risk.

## EXACT code to create

```go
package planner

import (
	"context"
	"math"
	
	"github.com/tiendat1751998/orchestrator/contracts/fsm"
)

// Weights configure the scoring preferences.
type Weights struct {
	Quality    float64
	Cost       float64
	Time       float64
	Confidence float64
	Risk       float64
}

// Scorer calculates the composite score.
type Scorer struct {
	weights Weights
	cFactor float64 // UCB exploration constant
}

// NewScorer constructs a new Pareto scorer.
func NewScorer(w Weights, c float64) *Scorer {
	return &Scorer{weights: w, cFactor: c}
}

// ScoreCandidate calculates a candidate plan's composite score.
func (s *Scorer) ScoreCandidate(
	ctx context.Context,
	dag fsm.DAG,
	q, c, t, conf, r float64,
	totalRuns int,
	usageCount int,
) float64 {
	// 1. Pareto score base calculation
	pareto := s.weights.Quality*q +
		s.weights.Confidence*conf +
		s.weights.Cost*(1.0-c) + // lower cost is better
		s.weights.Time*(1.0-t) - // lower duration is better
		s.weights.Risk*r
		
	// 2. UCB-1 Exploration Bonus
	exploration := 0.0
	if usageCount > 0 && totalRuns > 0 {
		exploration = s.cFactor * math.Sqrt(math.Log(float64(totalRuns))/float64(usageCount))
	} else if usageCount == 0 {
		exploration = 1.0 // Maximum bonus for new templates
	}
	
	return pareto + exploration
}
```

## Verify
```bash
go build ./kernel/planner/...
```

## Checklist
- [ ] File `kernel/planner/pareto.go` exists
- [ ] Package: `planner`
- [ ] Weights struct defines quality, cost, time, confidence, risk fields
- [ ] `ScoreCandidate` applies UCB-1 bonus on top of Pareto values
- [ ] `go build ./kernel/planner/...` passes
