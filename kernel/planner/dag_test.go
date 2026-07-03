package planner

import (
	"testing"
)

func TestDAG_ValidateCycles(t *testing.T) {
	tests := []struct {
		name    string
		nodes   map[string]*DAGNode
		wantErr bool
		errSub  string
	}{
		{
			name:    "nil nodes",
			nodes:   nil,
			wantErr: false,
		},
		{
			name:    "empty nodes",
			nodes:   map[string]*DAGNode{},
			wantErr: false,
		},
		{
			name: "valid linear chain",
			nodes: map[string]*DAGNode{
				"A": {ID: "A", Dependencies: []string{}},
				"B": {ID: "B", Dependencies: []string{"A"}},
				"C": {ID: "C", Dependencies: []string{"B"}},
			},
			wantErr: false,
		},
		{
			name: "valid diamond graph",
			nodes: map[string]*DAGNode{
				"A": {ID: "A", Dependencies: []string{}},
				"B": {ID: "B", Dependencies: []string{"A"}},
				"C": {ID: "C", Dependencies: []string{"A"}},
				"D": {ID: "D", Dependencies: []string{"B", "C"}},
			},
			wantErr: false,
		},
		{
			name: "circular dependency self loop",
			nodes: map[string]*DAGNode{
				"A": {ID: "A", Dependencies: []string{"A"}},
			},
			wantErr: true,
			errSub:  "circular dependency loop detected",
		},
		{
			name: "circular dependency cycle",
			nodes: map[string]*DAGNode{
				"A": {ID: "A", Dependencies: []string{"C"}},
				"B": {ID: "B", Dependencies: []string{"A"}},
				"C": {ID: "C", Dependencies: []string{"B"}},
			},
			wantErr: true,
			errSub:  "circular dependency loop detected",
		},
		{
			name: "unresolved dependency",
			nodes: map[string]*DAGNode{
				"A": {ID: "A", Dependencies: []string{"B"}},
			},
			wantErr: true,
			errSub:  "unresolved dependency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DAG{
				Nodes: tt.nodes,
			}
			err := d.ValidateCycles()
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateCycles() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errSub != "" {
				// check that the error message matches or contains the expected substring
				if !contains(err.Error(), tt.errSub) {
					t.Errorf("ValidateCycles() error = %q, want substring %q", err.Error(), tt.errSub)
				}
			}
		})
	}
}

func TestDAG_NilDAG(t *testing.T) {
	var d *DAG
	err := d.ValidateCycles()
	if err == nil {
		t.Fatal("expected error for nil DAG, got nil")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || find(s, substr) >= 0)
}

func find(s, substr string) int {
	n := len(substr)
	for i := 0; i <= len(s)-n; i++ {
		if s[i:i+n] == substr {
			return i
		}
	}
	return -1
}
