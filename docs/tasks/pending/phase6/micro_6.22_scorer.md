# Micro-Task 6.22: Create kernel/feedback/scorer.go

## Info
- **File**: `kernel/feedback/scorer.go`
- **Package**: `feedback`
- **Depends on**: 6.21
- **Time**: 15 min
- **Verify**: `go build ./kernel/feedback/...`

## Purpose
Tracks agent performance metrics (success rate, avg duration, token efficiency) with time-decayed weights for agent selection optimization.

## EXACT code to create

```go
package feedback

import (
	"math"
	"sync"
	"time"
)

// AgentScore holds accumulated performance metrics for an agent.
type AgentScore struct {
	AgentName    string        `json:"agent_name"`
	TotalTasks   int           `json:"total_tasks"`
	Successes    int           `json:"successes"`
	Failures     int           `json:"failures"`
	AvgDuration  time.Duration `json:"avg_duration"`
	AvgScore     float64       `json:"avg_score"`
	TokensUsed   int64         `json:"tokens_used"`
	LastUpdated  time.Time     `json:"last_updated"`
}

// SuccessRate returns the ratio of successful tasks.
func (s *AgentScore) SuccessRate() float64 {
	if s.TotalTasks == 0 {
		return 0
	}
	return float64(s.Successes) / float64(s.TotalTasks)
}

// CompositeScore computes a weighted score factoring success rate, speed, and quality.
func (s *AgentScore) CompositeScore() float64 {
	successWeight := 0.5
	qualityWeight := 0.3
	speedWeight := 0.2

	// Normalize duration to [0,1] where shorter = better (max 10 min baseline)
	speedScore := 1.0 - math.Min(s.AvgDuration.Seconds()/600.0, 1.0)

	// Apply time decay: recent performance matters more
	decayFactor := 1.0
	daysSinceUpdate := time.Since(s.LastUpdated).Hours() / 24
	if daysSinceUpdate > 7 {
		decayFactor = math.Exp(-0.1 * (daysSinceUpdate - 7))
	}

	raw := s.SuccessRate()*successWeight + s.AvgScore*qualityWeight + speedScore*speedWeight
	return raw * decayFactor
}

// Scorer aggregates agent performance statistics. Thread-safe.
type Scorer struct {
	mu     sync.RWMutex
	scores map[string]*AgentScore
}

// NewScorer constructs a new Scorer.
func NewScorer() *Scorer {
	return &Scorer{
		scores: make(map[string]*AgentScore),
	}
}

// Record updates an agent's score with a new task result.
func (s *Scorer) Record(agentName string, eval *Evaluation, duration time.Duration, tokens int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	score, exists := s.scores[agentName]
	if !exists {
		score = &AgentScore{AgentName: agentName}
		s.scores[agentName] = score
	}

	score.TotalTasks++
	if eval.Score >= 0.5 {
		score.Successes++
	} else {
		score.Failures++
	}

	// Running average for duration and quality score
	n := float64(score.TotalTasks)
	score.AvgDuration = time.Duration(
		(float64(score.AvgDuration)*(n-1) + float64(duration)) / n,
	)
	score.AvgScore = (score.AvgScore*(n-1) + eval.Score) / n
	score.TokensUsed += tokens
	score.LastUpdated = time.Now()
}

// Get returns the score for a specific agent. Returns nil if not found.
func (s *Scorer) Get(agentName string) *AgentScore {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.scores[agentName]
}

// All returns all agent scores.
func (s *Scorer) All() []*AgentScore {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*AgentScore, 0, len(s.scores))
	for _, v := range s.scores {
		result = append(result, v)
	}
	return result
}
```

## Rules
1. **Running Average**: Use online mean formula `(avg*(n-1) + new) / n` — no need to store all historical values.
2. **Time Decay**: Scores older than 7 days decay exponentially. Prevents stale agents from dominating rankings.
3. **Composite Score**: Weighted combination of success rate (50%), quality (30%), speed (20%).

## Verify
```bash
go build ./kernel/feedback/...
```

## Checklist
- [ ] AgentScore with composite scoring
- [ ] Time-decayed weights
- [ ] Running average calculations
- [ ] Thread-safe via RWMutex
- [ ] `go build ./kernel/feedback/...` passes
