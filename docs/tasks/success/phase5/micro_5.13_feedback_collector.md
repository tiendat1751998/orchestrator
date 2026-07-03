# Micro-Task 5.13: Create kernel/orchestrator/feedback.go

## Info
- **File**: `kernel/orchestrator/feedback.go`
- **Package**: `orchestrator`
- **Depends on**: 5.12
- **Time**: 15 min
- **Verify**: `go build ./kernel/orchestrator/...`

## Purpose
Implements the orchestration feedback loops (`FeedbackCollector` and types) that track task speeds and costs to help optimize future task planning.

## EXACT code to create

```go
package orchestrator

import (
	"sync"
	"time"
)

// AgentMetrics tracks performance statistics for a specific agent type.
type AgentMetrics struct {
	AgentName   string
	SuccessCount int
	FailureCount int
	TotalTokens  int64
	TotalDuration time.Duration
}

// FeedbackCollector collects running metrics for feedback loop planning.
// Thread-safe.
type FeedbackCollector struct {
	mu      sync.RWMutex
	metrics map[string]*AgentMetrics
}

// NewFeedbackCollector constructs a new FeedbackCollector.
func NewFeedbackCollector() *FeedbackCollector {
	return &FeedbackCollector{
		metrics: make(map[string]*AgentMetrics),
	}
}

// RecordSuccess registers a successful task execution.
func (fc *FeedbackCollector) RecordSuccess(agentName string, tokens int64, duration time.Duration) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	m, ok := fc.metrics[agentName]
	if !ok {
		m = &AgentMetrics{AgentName: agentName}
		fc.metrics[agentName] = m
	}

	m.SuccessCount++
	m.TotalTokens += tokens
	m.TotalDuration += duration
}

// RecordFailure registers a failed task execution.
func (fc *FeedbackCollector) RecordFailure(agentName string) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	m, ok := fc.metrics[agentName]
	if !ok {
		m = &AgentMetrics{AgentName: agentName}
		fc.metrics[agentName] = m
	}

	m.FailureCount++
}

// GetMetrics returns a copy of the collected metrics.
func (fc *FeedbackCollector) GetMetrics() map[string]AgentMetrics {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	copied := make(map[string]AgentMetrics)
	for k, v := range fc.metrics {
		if v != nil {
			copied[k] = *v
		}
	}
	return copied
}
```

## Pitfalls

### Pitfall 1: Data races during concurrent updates
```go
// WRONG:
func (fc *FeedbackCollector) RecordSuccess(agentName string) {
    fc.metrics[agentName].SuccessCount++ // Data race! Multiple task goroutines modify map concurrently.
}

// CORRECT:
fc.mu.Lock()
defer fc.mu.Unlock()
```
Modifying maps from concurrent threads without using mutex locks triggers race conditions. Guard map updates with locks.

### Pitfall 2: Returning direct pointers to internal maps
Returning raw references to internal maps allows external code to mutate states without acquiring locks. Return deep copies instead.

## Verify
```bash
go build ./kernel/orchestrator/...
# Expected: clean compilation without errors
```

## Checklist
- [ ] File exists at `kernel/orchestrator/feedback.go`
- [ ] Package name is `orchestrator`
- [ ] All exported types have Godoc
- [ ] Metric writes are guarded under mutex locks
- [ ] GetMetrics returns a copied map rather than internal pointers
- [ ] Build command passes
