# Micro-Task 5.19: Create kernel/resilience/health.go

## Info
- **File**: `kernel/resilience/health.go`
- **Package**: `resilience`
- **Depends on**: 5.18
- **Time**: 15 min
- **Verify**: `go build ./kernel/resilience/...`

## Purpose
Implements the registry health check aggregator (`HealthAggregator`) to query and summarize status reports from all registered plugin components.

## EXACT code to create

```go
package resilience

import (
	"context"
	"sync"

	cplugin "github.com/tiendat1751998/orchestrator/contracts/plugin"
)

// HealthCheckable defines components that expose health query methods.
type HealthCheckable interface {
	Health(ctx context.Context) (cplugin.HealthReport, error)
}

// HealthAggregator coordinates queries across registered components.
// Thread-safe.
type HealthAggregator struct {
	mu         sync.RWMutex
	components map[string]HealthCheckable
}

// NewHealthAggregator constructs a NewHealthAggregator.
func NewHealthAggregator() *HealthAggregator {
	return &HealthAggregator{
		components: make(map[string]HealthCheckable),
	}
}

// Register registers a component for health checks.
func (ha *HealthAggregator) Register(name string, component HealthCheckable) {
	ha.mu.Lock()
	defer ha.mu.Unlock()
	ha.components[name] = component
}

// CheckAll queries health status across components in parallel.
func (ha *HealthAggregator) CheckAll(ctx context.Context) map[string]cplugin.HealthReport {
	ha.mu.RLock()
	defer ha.mu.RUnlock()

	results := make(map[string]cplugin.HealthReport)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for name, comp := range ha.components {
		wg.Add(1)
		go func(n string, c HealthCheckable) {
			defer wg.Done()

			report, err := c.Health(ctx)
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				results[n] = cplugin.HealthReport{
					Status:  cplugin.HealthDown,
					Message: err.Error(),
				}
				return
			}
			results[n] = report
		}(name, comp)
	}

	wg.Wait()
	return results
}
```

## Pitfalls

### Pitfall 1: Blocking aggregators on hanging health queries
```go
// WRONG:
// Running health checks sequentially:
for name, comp := range components {
    comp.Health(ctx) // If one component hangs, the entire health endpoint blocks!
}

// CORRECT:
// Query components in parallel using goroutines.
```
Sequential health queries block the aggregator if a single plugin hangs. Run queries in parallel.

### Pitfall 2: Concurrent writes in results maps
When querying health checks in parallel goroutines, writing results directly to a shared map without synchronization will cause data races. Protect writes under locks.

## Verify
```bash
go build ./kernel/resilience/...
# Expected: clean compilation without errors
```

## Checklist
- [ ] File exists at `kernel/resilience/health.go`
- [ ] Package name is `resilience`
- [ ] All exported types have Godoc
- [ ] Health queries run in parallel goroutines
- [ ] Results map updates are guarded under locks
- [ ] Build command passes
