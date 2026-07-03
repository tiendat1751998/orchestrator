package scheduler

import (
	"fmt"
	"sync"

	"github.com/tiendat1751998/orchestrator/contracts"
)

// DependencyTracker tracks task dependencies and completion status.
//
// A task is "ready" when ALL its dependencies have been marked completed.
//
// Thread-safety: all methods are safe for concurrent use.
type DependencyTracker struct {
	mu sync.RWMutex

	// dependencies maps task ID → set of dependency task IDs.
	// A task with no entry or empty set has no dependencies (immediately ready).
	dependencies map[contracts.TaskID]map[contracts.TaskID]bool

	// completed tracks which tasks have been completed.
	completed map[contracts.TaskID]bool
}

// NewDependencyTracker creates a new tracker.
func NewDependencyTracker() *DependencyTracker {
	return &DependencyTracker{
		dependencies: make(map[contracts.TaskID]map[contracts.TaskID]bool),
		completed:    make(map[contracts.TaskID]bool),
	}
}

// AddDependency records that taskID depends on depID.
//
// Meaning: taskID CANNOT start until depID is completed.
//
// Returns error if:
//   - Self-dependency: taskID == depID
//   - Circular dependency: A→B, B→A
//
// Thread-safety: acquires write lock.
func (dt *DependencyTracker) AddDependency(taskID, depID contracts.TaskID) error {
	if taskID == depID {
		return fmt.Errorf("scheduler: self-dependency: task %q depends on itself", taskID)
	}

	dt.mu.Lock()
	defer dt.mu.Unlock()

	// Add the dependency
	if dt.dependencies[taskID] == nil {
		dt.dependencies[taskID] = make(map[contracts.TaskID]bool)
	}
	dt.dependencies[taskID][depID] = true

	// Check for circular dependency
	if dt.hasCircularDep(taskID) {
		// Rollback: remove the dependency we just added
		delete(dt.dependencies[taskID], depID)
		if len(dt.dependencies[taskID]) == 0 {
			delete(dt.dependencies, taskID)
		}
		return fmt.Errorf("scheduler: circular dependency detected involving task %q and %q", taskID, depID)
	}

	return nil
}

// AddDependencies records multiple dependencies for a task at once.
//
// This is a convenience wrapper around AddDependency.
// If any dependency causes an error, ALL dependencies for this call are rolled back.
func (dt *DependencyTracker) AddDependencies(taskID contracts.TaskID, depIDs []contracts.TaskID) error {
	for _, depID := range depIDs {
		if err := dt.AddDependency(taskID, depID); err != nil {
			// Rollback all dependencies added in this call
			dt.mu.Lock()
			for _, added := range depIDs {
				if dt.dependencies[taskID] != nil {
					delete(dt.dependencies[taskID], added)
				}
			}
			if len(dt.dependencies[taskID]) == 0 {
				delete(dt.dependencies, taskID)
			}
			dt.mu.Unlock()
			return err
		}
	}
	return nil
}

// MarkCompleted marks a task as completed.
// This may make dependent tasks ready.
func (dt *DependencyTracker) MarkCompleted(taskID contracts.TaskID) {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	dt.completed[taskID] = true
}

// IsReady checks if a task has all dependencies completed.
//
// Returns true if:
//   - Task has no dependencies, OR
//   - ALL dependency tasks are marked completed
func (dt *DependencyTracker) IsReady(taskID contracts.TaskID) bool {
	dt.mu.RLock()
	defer dt.mu.RUnlock()

	deps, hasDeps := dt.dependencies[taskID]
	if !hasDeps || len(deps) == 0 {
		return true // No dependencies
	}

	for depID := range deps {
		if !dt.completed[depID] {
			return false // At least one dependency not completed
		}
	}
	return true
}

// PendingDependencies returns the list of uncompleted dependencies for a task.
//
// Returns empty slice if the task is ready.
// Useful for debugging "why isn't my task running?".
func (dt *DependencyTracker) PendingDependencies(taskID contracts.TaskID) []contracts.TaskID {
	dt.mu.RLock()
	defer dt.mu.RUnlock()

	deps, hasDeps := dt.dependencies[taskID]
	if !hasDeps {
		return nil
	}

	var pending []contracts.TaskID
	for depID := range deps {
		if !dt.completed[depID] {
			pending = append(pending, depID)
		}
	}
	return pending
}

// =============================================================================
// Circular dependency detection (internal)
// =============================================================================

// hasCircularDep detects if adding current dependencies creates a cycle.
//
// Algorithm: DFS from taskID following dependency edges.
// If we reach taskID again → cycle exists.
//
// MUST be called while holding dt.mu.Lock().
func (dt *DependencyTracker) hasCircularDep(startID contracts.TaskID) bool {
	visited := make(map[contracts.TaskID]bool)
	return dt.dfs(startID, startID, visited)
}

// dfs performs depth-first search looking for cycles.
func (dt *DependencyTracker) dfs(currentID, startID contracts.TaskID, visited map[contracts.TaskID]bool) bool {
	if visited[currentID] {
		return false // Already visited this node, no cycle through here
	}
	visited[currentID] = true

	deps := dt.dependencies[currentID]
	for depID := range deps {
		if depID == startID {
			return true // Found a cycle back to the start
		}
		if dt.dfs(depID, startID, visited) {
			return true // Cycle found deeper in the graph
		}
	}

	return false
}

// Reset clears all tracked dependencies and completions.
// Used between missions.
func (dt *DependencyTracker) Reset() {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	dt.dependencies = make(map[contracts.TaskID]map[contracts.TaskID]bool)
	dt.completed = make(map[contracts.TaskID]bool)
}
