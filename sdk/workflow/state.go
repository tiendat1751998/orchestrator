package workflow

import (
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

	// ponytail: lock is only held during specific map/state reads to prevent recursive RLock deadlocks in Go.
	switch parts[0] {
	case "inputs":
		key := parts[1]
		s.mu.RLock()
		val, ok := s.Inputs[key]
		s.mu.RUnlock()
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

		s.mu.RLock()
		res, ok := s.StepResults[stepName]
		s.mu.RUnlock()
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
		if curr == nil {
			return nil, fmt.Errorf("cannot lookup field %q on nil value", segment)
		}
		v := reflect.ValueOf(curr)
		for v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return nil, fmt.Errorf("cannot lookup field %q on nil pointer or interface", segment)
			}
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
