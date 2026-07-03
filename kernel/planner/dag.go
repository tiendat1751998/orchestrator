package planner

import (
	"errors"
	"fmt"
	"sync"

	"github.com/tiendat1751998/orchestrator/contracts/fsm"
)

// DAGNode wraps task configs with statuses.
type DAGNode struct {
	ID           string    `json:"id"`
	Dependencies []string  `json:"dependencies"`
	Status       fsm.State `json:"status"`
}

// DAG represents a Directed Acyclic Graph of execution nodes.
type DAG struct {
	mu    sync.RWMutex
	Nodes map[string]*DAGNode `json:"nodes"`
}

// ValidateCycles implements Kahn's algorithm to detect circular dependencies.
func (d *DAG) ValidateCycles() error {
	if d == nil {
		return errors.New("dag: nil dag")
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.Nodes == nil {
		return nil
	}

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
