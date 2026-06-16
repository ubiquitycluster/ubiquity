package tui

import (
	"testing"

	"github.com/ubiquitycluster/ubiquity/pkg/provision"
)

func TestRenderStatusWithState(t *testing.T) {
	state := provision.NewState("sandbox")
	result := RenderStatus(state)
	if result == "" {
		t.Fatal("RenderStatus returned empty string")
	}
}

func TestRenderStatusContainsPhases(t *testing.T) {
	state := provision.NewState("test")
	result := RenderStatus(state)
	phases := []string{"metal", "bootstrap", "security", "external", "wait", "post-install"}
	for _, phase := range phases {
		if !stringContains(result, phase) {
			t.Errorf("RenderStatus missing phase %q", phase)
		}
	}
}

func TestRenderStatusNil(t *testing.T) {
	result := RenderStatus(nil)
	if result == "" {
		t.Error("RenderStatus returned empty for nil")
	}
}

func TestPrintStatusNoPanic(t *testing.T) {
	PrintStatus(nil)                        // should not panic
	PrintStatus(provision.NewState("test")) // should not panic
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestStatusColorHandlesAllStates(t *testing.T) {
	states := []provision.PhaseState{
		provision.PhaseSuccess,
		provision.PhaseRunning,
		provision.PhaseFailed,
		provision.PhaseSkipped,
		provision.PhasePending,
	}
	for _, s := range states {
		result := StatusColor(s)
		if result == "" {
			t.Errorf("StatusColor returned empty for state %q", s)
		}
	}
}
