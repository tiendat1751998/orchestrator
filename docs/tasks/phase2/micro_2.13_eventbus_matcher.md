# Micro-Task 2.13: Tạo kernel/eventbus/matcher.go

## Thông tin
- **File tạo**: `kernel/eventbus/matcher.go`
- **Package**: `eventbus`
- **Dependencies trước**: 2.12
- **Thời gian**: 15 phút
- **Verify**: `go build ./kernel/eventbus/...`

## Mục đích
Wildcard pattern matching cho event type subscription.
`"task.*"` matches `"task.started"`, `"task.completed"`, etc.

## Nội dung CHÍNH XÁC cần tạo

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

## ⚠️ Pitfalls cần tránh

### Pitfall 1: "*" matches ALL events
Global wildcard `"*"` is special — matches any event type regardless of segments.
DO NOT split it and try to match segments.

### Pitfall 2: Segment count MUST match
```
"task.*"  (2 segments) vs "task" (1 segment) → NO match
"task.*"  (2 segments) vs "task.started" (2 segments) → match
"task.*"  (2 segments) vs "task.sub.deep" (3 segments) → NO match
```
This prevents `"task.*"` from matching deeply nested events like `"task.sub.deep"`.

### Pitfall 3: Fast path for exact match
Most subscriptions use exact event types (`"task.started"`).
Check exact match BEFORE splitting into segments (avoid allocation).

### Pitfall 4: KHÔNG dùng path.Match()
`path.Match()` uses filesystem matching rules (`?` matches single char, `[...]` matches char class).
We want simpler rules: `*` matches single segment. That's it.

## Test cases cho matcher (implement trong 2.17)
```go
// MUST pass:
matchPattern("*", "task.started")        // true
matchPattern("*", "anything")            // true
matchPattern("task.started", "task.started") // true
matchPattern("task.*", "task.started")   // true
matchPattern("task.*", "task.completed") // true
matchPattern("*.started", "task.started") // true
matchPattern("*.*", "task.started")      // true

// MUST fail:
matchPattern("task.*", "task")           // false (1 vs 2 segments)
matchPattern("task.*", "mission.started") // false (first segment mismatch)
matchPattern("task.started", "task.completed") // false (exact mismatch)
matchPattern("*.*", "task")              // false (1 vs 2 segments)
matchPattern("task.*", "task.sub.deep")  // false (2 vs 3 segments)
```

## Checklist
- [ ] File `kernel/eventbus/matcher.go` tồn tại
- [ ] `matchPattern()` function — unexported (internal)
- [ ] Global wildcard `"*"` matches everything
- [ ] Exact match fast path
- [ ] Segment-based comparison
- [ ] `"*"` segment matches any single segment
- [ ] Different segment counts → no match
- [ ] KHÔNG dùng `path.Match()`
- [ ] Godoc với matching rules và examples
- [ ] `go build ./kernel/eventbus/...` không lỗi
