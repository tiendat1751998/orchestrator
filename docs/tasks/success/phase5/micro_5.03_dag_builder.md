# Micro-Task 5.03: Create kernel/planner/dag.go

- **File**: `kernel/planner/dag.go`
- **Package**: `planner`
- **Depends on**: 5.01, 1.23 (contracts/event/event.go)
- **Time**: 20 min
- **Verify**: `go build ./kernel/planner/...`

## Purpose
Implements Directed Acyclic Graph (DAG) management and Kahn's cycle validation for the FSM. It ensures plan paths are topologically sorted and clean of circular references.

## EXACT code to create

```go
package planner

import (
	"errors"
	"fmt"
	"sync"

	"github.com/tiendat1751998/orchestrator/contracts/fsm"
)

// DAGNode wraps task configs with statuses.
type DAGNode struct {
	ID           string           `json:"id"`
	Dependencies []string         `json:"dependencies"`
	Status       fsm.State        `json:"status"`
}

// DAG represents a Directed Acyclic Graph of execution nodes.
type DAG struct {
	mu    sync.RWMutex
	Nodes map[string]*DAGNode `json:"nodes"`
}

// ValidateCycles implements Kahn's algorithm to detect circular dependencies.
func (d *DAG) ValidateCycles() error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	inDegree := make(map[string]int)
	adj := make(map[string][]string)

	for id := range d.Nodes {
		inDegree[id] = 0
	}

	for id, node := range d.Nodes {
		for _, dep := range node.Dependencies {
			if _, exists := d.Nodes[dep]; !exists {
				return fmt.Errorf("dag: task %q has unresolved dependency %q", id, dep)
			}
			adj[dep] = append(adj[dep], id)
			inDegree[id]++
		}
	}

	var queue []string
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	visited := 0
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		visited++

		for _, neighbor := range adj[curr] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if visited != len(d.Nodes) {
		return errors.New("dag: circular dependency loop detected")
	}

	return nil
}
```

## Verify
```bash
go build ./kernel/planner/...
```

## Checklist
- [ ] File `kernel/planner/dag.go` exists
- [ ] Package: `planner`
- [ ] `DAG` struct holds nodes in thread-safe map
- [ ] `ValidateCycles` implements Kahn's algorithm
- [ ] `go build ./kernel/planner/...` passes
