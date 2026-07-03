# Micro-Task 6.24: Create kernel/feedback/ranking.go

## Info
- **File**: `kernel/feedback/ranking.go`
- **Package**: `feedback`
- **Depends on**: 6.22 (scorer), 6.23 (learner)
- **Time**: 15 min
- **Verify**: `go build ./kernel/feedback/...`

## Purpose
Agent ranking engine that combines scorer data and learner history to select the optimal agent for each task type. Falls back to round-robin if insufficient data.

## EXACT code to create

```go
package feedback

import (
	"sort"
	"sync"
	"sync/atomic"
)

// RankingEngine selects the best agent for a given task type. Thread-safe.
type RankingEngine struct {
	scorer       *Scorer
	learner      *Learner
	roundRobinIdx atomic.Int64
}

// NewRankingEngine constructs a ranking engine.
func NewRankingEngine(scorer *Scorer, learner *Learner) *RankingEngine {
	return &RankingEngine{
		scorer:  scorer,
		learner: learner,
	}
}

// RankedAgent holds an agent name and its computed rank score.
type RankedAgent struct {
	Name  string  `json:"name"`
	Score float64 `json:"score"`
}

// Rank returns agents sorted by suitability for the given task type (best first).
func (re *RankingEngine) Rank(taskType string, availableAgents []string) []RankedAgent {
	if len(availableAgents) == 0 {
		return nil
	}

	ranked := make([]RankedAgent, 0, len(availableAgents))

	for _, name := range availableAgents {
		score := 0.5 // Default neutral score

		// Factor 1: Historical scorer data (60% weight)
		if agentScore := re.scorer.Get(name); agentScore != nil {
			score = 0.4*score + 0.6*agentScore.CompositeScore()
		}

		// Factor 2: Learner task-type affinity (40% boost if best)
		bestAgent := re.learner.BestAgentFor(taskType)
		if bestAgent == name {
			score *= 1.2 // 20% boost for historically best agent
		}

		ranked = append(ranked, RankedAgent{Name: name, Score: score})
	}

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Score > ranked[j].Score
	})

	return ranked
}

// BestAgent returns the single best agent for a task type.
// Falls back to round-robin if all scores are equal.
func (re *RankingEngine) BestAgent(taskType string, availableAgents []string) string {
	if len(availableAgents) == 0 {
		return ""
	}

	ranked := re.Rank(taskType, availableAgents)
	if len(ranked) == 0 {
		return availableAgents[0]
	}

	// If top scores are tied (within 1%), use round-robin
	if len(ranked) > 1 && (ranked[0].Score-ranked[1].Score) < 0.01 {
		idx := re.roundRobinIdx.Add(1)
		return availableAgents[int(idx)%len(availableAgents)]
	}

	return ranked[0].Name
}
```

## Rules
1. **Multi-Factor Ranking**: Combine scorer (historical performance) and learner (task-type affinity).
2. **Round-Robin Fallback**: When agents are equally ranked, round-robin prevents starvation.
3. **Atomic Counter**: Use `atomic.Int64` for the round-robin index — no mutex needed for a simple counter.

## Verify
```bash
go build ./kernel/feedback/...
```

## Checklist
- [ ] Rank returns sorted agents
- [ ] BestAgent with round-robin fallback
- [ ] Multi-factor scoring
- [ ] `go build ./kernel/feedback/...` passes
