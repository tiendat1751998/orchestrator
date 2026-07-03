# Micro-Task 6.23: Create kernel/feedback/learner.go

## Info
- **File**: `kernel/feedback/learner.go`
- **Package**: `feedback`
- **Depends on**: 6.22 (scorer)
- **Time**: 15 min
- **Verify**: `go build ./kernel/feedback/...`

## Purpose
Persists historical task→agent performance mapping to JSON. Provides data for optimizing future agent selection based on task type history.

## EXACT code to create

```go
package feedback

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// TaskTypeStats maps task types to their historical agent performance.
type TaskTypeStats struct {
	TaskType string            `json:"task_type"`
	Agents   map[string]float64 `json:"agents"` // agentName → composite score
}

// Learner persists agent selection history for future optimization. Thread-safe.
type Learner struct {
	mu       sync.Mutex
	dataPath string
	history  map[string]*TaskTypeStats // taskType → stats
}

// NewLearner constructs a learner persisting data to the given directory.
func NewLearner(dataDir string) (*Learner, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("learner: failed to create data dir: %w", err)
	}

	l := &Learner{
		dataPath: filepath.Join(dataDir, "learning_data.json"),
		history:  make(map[string]*TaskTypeStats),
	}

	// Load existing history if available
	l.load()

	return l, nil
}

// RecordOutcome stores the result of an agent performing a task type.
func (l *Learner) RecordOutcome(taskType, agentName string, score float64) {
	l.mu.Lock()
	defer l.mu.Unlock()

	stats, exists := l.history[taskType]
	if !exists {
		stats = &TaskTypeStats{
			TaskType: taskType,
			Agents:   make(map[string]float64),
		}
		l.history[taskType] = stats
	}

	// Exponential moving average (α=0.3 weights recent data)
	existing, ok := stats.Agents[agentName]
	if ok {
		stats.Agents[agentName] = 0.7*existing + 0.3*score
	} else {
		stats.Agents[agentName] = score
	}

	l.persist()
}

// BestAgentFor returns the highest-scoring agent for a task type, or empty string if no data.
func (l *Learner) BestAgentFor(taskType string) string {
	l.mu.Lock()
	defer l.mu.Unlock()

	stats, exists := l.history[taskType]
	if !exists {
		return ""
	}

	var bestAgent string
	var bestScore float64
	for agent, score := range stats.Agents {
		if score > bestScore {
			bestScore = score
			bestAgent = agent
		}
	}
	return bestAgent
}

func (l *Learner) load() {
	data, err := os.ReadFile(l.dataPath)
	if err != nil {
		return // Fresh start
	}
	json.Unmarshal(data, &l.history)
}

func (l *Learner) persist() {
	data, _ := json.MarshalIndent(l.history, "", "  ")
	os.WriteFile(l.dataPath, data, 0644)
}
```

## Rules
1. **Exponential Moving Average**: α=0.3 for recent weighting. Adapts to changing agent performance.
2. **Lazy Load**: Load history from disk on construction. Persist on every update.

## Verify
```bash
go build ./kernel/feedback/...
```

## Checklist
- [ ] EMA-based scoring
- [ ] BestAgentFor returns optimal agent
- [ ] Persistence to JSON file
- [ ] `go build ./kernel/feedback/...` passes
