package provision

import (
	"os"
	"testing"
)

func cleanupState() {
	os.Remove(StatePath())
}

func TestNewState(t *testing.T) {
	defer cleanupState()
	s := NewState("sandbox")
	if s == nil {
		t.Fatal("NewState returned nil")
	}
	if s.Environment != "sandbox" {
		t.Errorf("expected environment 'sandbox', got '%s'", s.Environment)
	}
	if len(s.Phases) != len(PipelineOrder) {
		t.Errorf("expected %d phases, got %d", len(PipelineOrder), len(s.Phases))
	}
	for i, p := range s.Phases {
		if p.Name != PipelineOrder[i] {
			t.Errorf("phase %d: expected name '%s', got '%s'", i, PipelineOrder[i], p.Name)
		}
		if p.Status != PhasePending {
			t.Errorf("phase %s: expected status pending, got %s", p.Name, p.Status)
		}
	}
}

func TestPhaseLifecycle(t *testing.T) {
	defer cleanupState()
	s := NewState("test")

	if err := s.StartPhase("metal"); err != nil {
		t.Fatalf("StartPhase failed: %v", err)
	}
	if s.Phases[0].Status != PhaseRunning {
		t.Errorf("expected metal to be running, got %s", s.Phases[0].Status)
	}
	if s.CurrentPhase != "metal" {
		t.Errorf("expected CurrentPhase 'metal', got '%s'", s.CurrentPhase)
	}

	if err := s.CompletePhase("metal"); err != nil {
		t.Fatalf("CompletePhase failed: %v", err)
	}
	if s.Phases[0].Status != PhaseSuccess {
		t.Errorf("expected metal to be success, got %s", s.Phases[0].Status)
	}
	if s.CurrentPhase != "" {
		t.Errorf("expected CurrentPhase to be empty after completion, got '%s'", s.CurrentPhase)
	}
	if s.Phases[0].Duration == "" {
		t.Error("expected duration to be set after completion")
	}
}

func TestFailPhase(t *testing.T) {
	defer cleanupState()
	s := NewState("test")
	s.StartPhase("metal")
	err := s.FailPhase("metal", os.ErrNotExist)
	if err != nil {
		t.Fatalf("FailPhase failed: %v", err)
	}
	if s.Phases[0].Status != PhaseFailed {
		t.Errorf("expected failed, got %s", s.Phases[0].Status)
	}
	if s.Phases[0].Error == "" {
		t.Error("expected error message to be set")
	}
}

func TestSkipPhase(t *testing.T) {
	defer cleanupState()
	s := NewState("test")
	s.SkipPhase("bootstrap")
	if s.Phases[1].Status != PhaseSkipped {
		t.Errorf("expected skipped, got %s", s.Phases[1].Status)
	}
}

func TestUnknownPhase(t *testing.T) {
	defer cleanupState()
	s := NewState("test")
	if err := s.StartPhase("nonexistent"); err == nil {
		t.Error("expected error for unknown phase")
	}
}

func TestLoadStateNonExistent(t *testing.T) {
	cleanupState()
	s, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState should not error for missing file: %v", err)
	}
	if s != nil {
		t.Error("expected nil state for missing file")
	}
}

func TestSaveAndLoadState(t *testing.T) {
	defer cleanupState()
	s := NewState("prod")
	s.StartPhase("bootstrap")
	s.CompletePhase("bootstrap")

	if err := s.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}
	if loaded == nil {
		t.Fatal("LoadState returned nil after save")
	}
	if loaded.Environment != "prod" {
		t.Errorf("expected environment 'prod', got '%s'", loaded.Environment)
	}
	if len(loaded.Phases) != len(PipelineOrder) {
		t.Errorf("expected %d phases, got %d", len(PipelineOrder), len(loaded.Phases))
	}
}

func TestSummary(t *testing.T) {
	defer cleanupState()
	s := NewState("sandbox")
	summary := s.Summary()
	if summary == "" {
		t.Error("expected non-empty summary")
	}
	if !contains(summary, "sandbox") {
		t.Error("summary should contain environment name")
	}
	if !contains(summary, "metal") {
		t.Error("summary should contain phase names")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}