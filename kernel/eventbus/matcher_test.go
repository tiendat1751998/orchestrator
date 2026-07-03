package eventbus

import "testing"

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		pattern   string
		eventType string
		expected  bool
	}{
		// Global wildcard
		{"*", "anything.here", true},
		{"*", "task", true},
		{"*", "", true},

		// Exact match
		{"task.started", "task.started", true},
		{"task.started", "task.completed", false},
		{"task", "task", true},
		{"", "", true},

		// Wildcard segments
		{"task.*", "task.started", true},
		{"task.*", "task.completed", true},
		{"task.*", "task", false},
		{"task.*", "task.sub.deep", false},
		{"task.*", "mission.started", false},

		// Prefix wildcard
		{"*.started", "task.started", true},
		{"*.started", "mission.started", true},
		{"*.started", "task.completed", false},
		{"*.started", "started", false},

		// Multiple wildcards
		{"*.*", "task.started", true},
		{"*.*", "task", false},
		{"*.*.*", "task.sub.deep", true},
		{"*.*.*", "task.sub", false},
		{"task.*.started", "task.sub.started", true},
		{"task.*.started", "task.sub.completed", false},
	}

	for _, tc := range tests {
		t.Run(tc.pattern+"_vs_"+tc.eventType, func(t *testing.T) {
			result := matchPattern(tc.pattern, tc.eventType)
			if result != tc.expected {
				t.Errorf("matchPattern(%q, %q) = %v; expected %v", tc.pattern, tc.eventType, result, tc.expected)
			}
		})
	}
}
