# Micro-Task 3.24: Create sdk/workflow/state.go

## Info
- **File**: `sdk/workflow/state.go`
- **Package**: `workflow`
- **Depends on**: 1.27 (workflow contract), 3.13 (workflow helper)
- **Time**: 20 min
- **Verify**: `go build ./sdk/workflow/...`

## Purpose
Implements the workflow runtime state resolver (`State` and recursive resolution methods) that manages inputs, step outputs, and maps templated fields (e.g. `{{steps.StepName.output}}`) dynamically during workflow executions.

## EXACT code to create

```go
package workflow

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/tiendat1751998/orchestrator/contracts/workflow"
)

// State holds the runtime execution states and inputs/outputs of a workflow.
// Safe for concurrent reading and writing.
type State struct {
	mu          sync.RWMutex
	Inputs      map[string]any
	StepResults map[string]*workflow.StepResult
}

// NewState initializes a new State instance with starting inputs.
func NewState(inputs map[string]any) *State {
	if inputs == nil {
		inputs = make(map[string]any)
	}
	return &State{
		Inputs:      inputs,
		StepResults: make(map[string]*workflow.StepResult),
	}
}

// SetStepResult stores the result of a completed workflow step.
func (s *State) SetStepResult(stepName string, res *workflow.StepResult) {
	if stepName == "" || res == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.StepResults[stepName] = res
}

// GetStepResult retrieves the result of a completed workflow step.
func (s *State) GetStepResult(stepName string) (*workflow.StepResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res, ok := s.StepResults[stepName]
	return res, ok
}

// ResolveValue inspects a value and resolves any template expressions recursively.
// Supported formats:
//   - "{{ inputs.KeyName }}"
//   - "{{ steps.StepName.status }}"
//   - "{{ steps.StepName.output }}"
//   - "{{ steps.StepName.error }}"
func (s *State) ResolveValue(val any) (any, error) {
	if val == nil {
		return nil, nil
	}

	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.String:
		return s.resolveString(v.String())

	case reflect.Map:
		resolvedMap := make(map[string]any)
		for _, key := range v.MapKeys() {
			kStr := fmt.Sprintf("%v", key.Interface())
			resolvedVal, err := s.ResolveValue(v.MapIndex(key).Interface())
			if err != nil {
				return nil, err
			}
			resolvedMap[kStr] = resolvedVal
		}
		return resolvedMap, nil

	case reflect.Slice:
		resolvedSlice := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			resolvedVal, err := s.ResolveValue(v.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			resolvedSlice[i] = resolvedVal
		}
		return resolvedSlice, nil
	}

	return val, nil
}

// resolveString extracts and evaluates a template expression.
func (s *State) resolveString(str string) (any, error) {
	trimmed := strings.TrimSpace(str)
	if !strings.HasPrefix(trimmed, "{{") || !strings.HasSuffix(trimmed, "}}") {
		return str, nil
	}

	expr := strings.TrimSpace(trimmed[2 : len(trimmed)-2])
	parts := strings.Split(expr, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("sdk/workflow: invalid template expression %q", str)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	switch parts[0] {
	case "inputs":
		key := parts[1]
		val, ok := s.Inputs[key]
		if !ok {
			return nil, fmt.Errorf("sdk/workflow: input key %q not found in state", key)
		}
		return val, nil

	case "steps":
		if len(parts) < 3 {
			return nil, fmt.Errorf("sdk/workflow: invalid steps expression %q (expected steps.stepName.property)", str)
		}
		stepName := parts[1]
		prop := parts[2]

		res, ok := s.StepResults[stepName]
		if !ok {
			return nil, fmt.Errorf("sdk/workflow: results for step %q not found in state", stepName)
		}

		switch prop {
		case "status":
			return string(res.Status), nil
		case "error":
			return res.Error, nil
		case "output":
			if len(parts) > 3 {
				return lookupNestedProperty(res.Output, parts[3:])
			}
			return res.Output, nil
		default:
			return nil, fmt.Errorf("sdk/workflow: unsupported step property %q in %q", prop, str)
		}
	}

	return nil, fmt.Errorf("sdk/workflow: unsupported template source %q in %q", parts[0], str)
}

func lookupNestedProperty(obj any, path []string) (any, error) {
	curr := obj
	for _, segment := range path {
		v := reflect.ValueOf(curr)
		if v.Kind() == reflect.Interface {
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Map:
			var found bool
			for _, key := range v.MapKeys() {
				if fmt.Sprintf("%v", key.Interface()) == segment {
					curr = v.MapIndex(key).Interface()
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("key %q not found in map", segment)
			}
		default:
			return nil, fmt.Errorf("cannot lookup field %q on type %T", segment, curr)
		}
	}
	return curr, nil
}
```

## Rules
1. **Recursive Template Resolutions**: Evaluate and resolve template strings recursively within collections (like nested maps or slices).
2. **Nested Property Extraction**: Support nested path lookup inside outputs (e.g. `{{steps.StepName.output.NestedKey}}`) using reflection lookup methods.
3. **Mutex Synchronization locks**: Guard operations on inputs and step results using a read/write mutex (`sync.RWMutex`).

## ⚠️ Pitfalls

### Pitfall 1: Unsynchronized reading during state updates
If parallel workflows modify state variables while another thread reads from the state, it will trigger data races. Guard all reads and writes with mutexes.

### Pitfall 2: Silent failures when accessing missing keys
If a template requests a key that does not exist in the inputs or steps registry, returning a blank value can hide errors. Always fail fast and return descriptive errors for missing keys.

## Verify
```bash
go build ./sdk/workflow/...
```

## Checklist
- [ ] File `sdk/workflow/state.go` exists
- [ ] Package: `workflow`
- [ ] `State` properties are protected using `sync.RWMutex` locks
- [ ] `ResolveValue` recursively parses slice and map items
- [ ] `inputs` variable expressions are resolved correctly
- [ ] `steps` outputs and statuses are resolved correctly
- [ ] Deep nested properties are extracted via reflection lookup helpers
- [ ] Missing variables trigger clean resolution errors
- [ ] `go build ./sdk/workflow/...` passes
