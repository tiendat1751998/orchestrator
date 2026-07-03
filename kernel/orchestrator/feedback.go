package orchestrator

import (
	"sync"
	"time"
)

// AgentMetrics tracks performance statistics for a specific agent type.
type AgentMetrics struct {
	AgentName     string
	SuccessCount  int
	FailureCount  int
	TotalTokens   int64
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
