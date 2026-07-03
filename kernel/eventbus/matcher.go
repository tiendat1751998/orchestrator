package eventbus

import "strings"

// matchPattern checks if an event type matches a subscription pattern.
//
// Matching rules:
//
//	"*"             → matches EVERYTHING
//	"task.started"  → matches ONLY "task.started" (exact match)
//	"task.*"        → matches "task.started", "task.completed", "task.failed"
//	                  does NOT match "task" (no sub-event)
//	                  does NOT match "task.sub.deep" (only 1 level deep)
//	"*.started"     → matches "task.started", "mission.started" (prefix wildcard)
//
// Implementation:
// Split both pattern and eventType by "." into segments.
// Compare segment by segment. "*" segment matches any single segment.
//
// Examples:
//
//	matchPattern("task.*", "task.started")       → true
//	matchPattern("task.*", "task.completed")      → true
//	matchPattern("task.*", "task")                → false (different segment count)
//	matchPattern("task.*", "mission.started")     → false (first segment mismatch)
//	matchPattern("task.started", "task.started")  → true (exact match)
//	matchPattern("*", "anything.here")            → true (global wildcard)
//	matchPattern("*.*", "task.started")           → true
//	matchPattern("*.*", "task")                   → false (different segment count)
//
// ponytail: Uses a simple split-and-compare strategy. Avoids regexp/filepath matching for simplicity and performance.
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
