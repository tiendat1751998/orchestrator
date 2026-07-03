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
