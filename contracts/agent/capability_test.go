package agent

import "testing"

func TestCapability_IsValid(t *testing.T) {
	tests := []struct {
		cap   Capability
		valid bool
	}{
		{CapabilityCodeGeneration, true},
		{CapabilityCodeReview, true},
		{CapabilityArchitecture, true},
		{CapabilityTesting, true},
		{CapabilityDocumentation, true},
		{CapabilityDeployment, true},
		{CapabilityDebugging, true},
		{CapabilityRefactoring, true},
		{CapabilityDataAnalysis, true},
		{CapabilityResearch, true},
		{Capability("invalid_capability"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.cap), func(t *testing.T) {
			if got := tt.cap.IsValid(); got != tt.valid {
				t.Errorf("Capability.IsValid() = %v, want %v for %v", got, tt.valid, tt.cap)
			}
		})
	}
}

func TestHasCapability(t *testing.T) {
	caps := []Capability{CapabilityCodeGeneration, CapabilityTesting}
	if !HasCapability(caps, CapabilityCodeGeneration) {
		t.Errorf("HasCapability() = false, want true")
	}
	if HasCapability(caps, CapabilityCodeReview) {
		t.Errorf("HasCapability() = true, want false")
	}
}
