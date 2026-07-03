# Micro-Task 2.13: Create kernel/eventbus/matcher.go

## Info
- **File**: `kernel/eventbus/matcher.go`
- **Package**: `eventbus`
- **Depends on**: 2.12
- **Time**: 15 min
- **Verify**: `go build ./kernel/eventbus/...`

## Purpose
Implements segment-based wildcard pattern matching (`matchPattern`) to verify if published event types match subscription patterns.

## EXACT code to create

```go
package eventbus

import "strings"

// matchPattern checks if an event type matches a subscription pattern.
//
// Matching rules:
//   "*"             → matches EVERYTHING
//   "task.started"  → matches ONLY "task.started" (exact match)
//   "task.*"        → matches "task.started", "task.completed", "task.failed"
//                     does NOT match "task" (no sub-event)
//                     does NOT match "task.sub.deep" (only 1 level deep)
//   "*.started"     → matches "task.started", "mission.started" (prefix wildcard)
//
// Implementation:
// Split both pattern and eventType by "." into segments.
// Compare segment by segment. "*" segment matches any single segment.
//
// Examples:
//   matchPattern("task.*", "task.started")       → true
//   matchPattern("task.*", "task.completed")      → true
//   matchPattern("task.*", "task")                → false (different segment count)
//   matchPattern("task.*", "mission.started")     → false (first segment mismatch)
//   matchPattern("task.started", "task.started")  → true (exact match)
//   matchPattern("*", "anything.here")            → true (global wildcard)
//   matchPattern("*.*", "task.started")           → true
//   matchPattern("*.*", "task")                   → false (different segment count)
func matchPattern(pattern, eventType string) bool {
	// Special case: global wildcard matches everything
	if pattern == "*" {
		return true
	}

	// Exact match (most common case — fast path)
	if pattern == eventType {
		return true
	}

	// Split into segments for wildcard matching
	patternParts := strings.Split(pattern, ".")
	eventParts := strings.Split(eventType, ".")

	// Must have the same number of segments
	// "task.*" has 2 parts, "task.started" has 2 parts → OK
	// "task.*" has 2 parts, "task" has 1 part → NO match
	if len(patternParts) != len(eventParts) {
		return false
	}

	// Compare segment by segment
	for i := range patternParts {
		if patternParts[i] == "*" {
			continue // Wildcard segment matches anything
		}
		if patternParts[i] != eventParts[i] {
			return false // Segment mismatch
		}
	}

	return true
}
```

## Rules
1. **Global Wildcard Bypass**: If the subscription pattern is strictly `*`, match all events immediately without splitting.
2. **Exact Match Fast Path**: Exact string equality checks (`pattern == eventType`) must run first to avoid memory allocations from string splitting.
3. **Segment Bounds Guard**: The segment count must match exactly. A pattern like `task.*` (2 segments) must never match a deep type like `task.sub.deep` (3 segments).

## ⚠️ Pitfalls

### Pitfall 1: Using general regex or filesystem libraries for simple matching
Avoid standard helpers like `path.Match` which parse file paths and character classes (`[a-z]`), as they add unnecessary complexity. A simple segment-by-segment splitter is faster and safer for event topics.

### Pitfall 2: Overlooking segment count differences during wildcard lookups
If segment counts are not validated, a pattern like `task.*` might match `task` by replacing the wildcard with empty strings. Enforce matching slice counts strictly.

## Verify
```bash
go build ./kernel/eventbus/...
```

## Checklist
- [ ] File `kernel/eventbus/matcher.go` exists
- [ ] Package: `eventbus`
- [ ] `matchPattern` is declared as unexported helper function
- [ ] Pattern `*` matches all inputs immediately
- [ ] Fast path for exact match checks runs before splitting
- [ ] String splits are performed on dot characters (`.`)
- [ ] Validation checks for matching segment counts are enforced
- [ ] `go build ./kernel/eventbus/...` passes
