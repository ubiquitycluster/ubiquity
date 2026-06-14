/*
Copyright © 2026 Ubiquity Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/ubiquitycluster/ubiquity/blob/main/LICENSE

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package provision

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PhaseState represents the status of a single provisioning phase.
type PhaseState string

const (
	PhasePending PhaseState = "pending"
	PhaseRunning PhaseState = "running"
	PhaseSuccess PhaseState = "success"
	PhaseFailed  PhaseState = "failed"
	PhaseSkipped PhaseState = "skipped"
)

// Phase tracks a single phase in the provisioning pipeline.
type Phase struct {
	Name      string     `json:"name"`
	Status    PhaseState `json:"status"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	Duration  string     `json:"duration,omitempty"`
	Error     string     `json:"error,omitempty"`
	LogURL    string     `json:"log_url,omitempty"`
}

// State represents the full provisioning state.
type State struct {
	Phases       []Phase `json:"phases"`
	Environment  string  `json:"environment"`
	UpdatedAt    string  `json:"updated_at"`
	CurrentPhase string  `json:"current_phase,omitempty"`
}

// PipelineOrder defines the order of provisioning phases.
var PipelineOrder = []string{
	"metal",
	"bootstrap",
	"security",
	"external",
	"wait",
	"post-install",
}

// NewState creates a new provisioning state for the given environment.
func NewState(env string) *State {
	phases := make([]Phase, len(PipelineOrder))
	for i, name := range PipelineOrder {
		phases[i] = Phase{
			Name:   name,
			Status: PhasePending,
		}
	}
	return &State{
		Phases:      phases,
		Environment: env,
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
}

// StatePath returns the path to the state file.
func StatePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ubiquity", "state.json")
}

// LoadState reads the provisioning state from disk.
func LoadState() (*State, error) {
	path := StatePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	state := &State{}
	if err := json.Unmarshal(data, state); err != nil {
		return nil, fmt.Errorf("parsing state file: %w", err)
	}
	return state, nil
}

// Save persists the provisioning state to disk.
func (s *State) Save() error {
	s.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	path := StatePath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}
	return nil
}

// StartPhase marks a phase as running and records the start time.
func (s *State) StartPhase(name string) error {
	for i := range s.Phases {
		if s.Phases[i].Name == name {
			now := time.Now()
			s.Phases[i].Status = PhaseRunning
			s.Phases[i].StartedAt = &now
			s.Phases[i].Error = ""
			s.CurrentPhase = name
			return s.Save()
		}
	}
	return fmt.Errorf("unknown phase: %s", name)
}

// CompletePhase marks a phase as completed successfully.
func (s *State) CompletePhase(name string) error {
	for i := range s.Phases {
		if s.Phases[i].Name == name {
			now := time.Now()
			s.Phases[i].Status = PhaseSuccess
			s.Phases[i].EndedAt = &now
			if s.Phases[i].StartedAt != nil {
				s.Phases[i].Duration = now.Sub(*s.Phases[i].StartedAt).Round(time.Second).String()
			}
			// Clear current phase if it was this one
			if s.CurrentPhase == name {
				s.CurrentPhase = ""
			}
			return s.Save()
		}
	}
	return fmt.Errorf("unknown phase: %s", name)
}

// FailPhase marks a phase as failed with an error message.
func (s *State) FailPhase(name string, err error) error {
	for i := range s.Phases {
		if s.Phases[i].Name == name {
			now := time.Now()
			s.Phases[i].Status = PhaseFailed
			s.Phases[i].EndedAt = &now
			s.Phases[i].Error = err.Error()
			if s.Phases[i].StartedAt != nil {
				s.Phases[i].Duration = now.Sub(*s.Phases[i].StartedAt).Round(time.Second).String()
			}
			return s.Save()
		}
	}
	return fmt.Errorf("unknown phase: %s", name)
}

// SkipPhase marks a phase as skipped.
func (s *State) SkipPhase(name string) error {
	for i := range s.Phases {
		if s.Phases[i].Name == name {
			s.Phases[i].Status = PhaseSkipped
			return s.Save()
		}
	}
	return fmt.Errorf("unknown phase: %s", name)
}

// Summary renders a human-readable summary of the provisioning state.
func (s *State) Summary() string {
	out := fmt.Sprintf("Ubiquity Cluster Status (%s)\n", s.Environment)
	out += "=====================\n\n"
	out += fmt.Sprintf("%-15s %-10s %s\n", "Phase", "Status", "Duration")
	out += fmt.Sprintf("%-15s %-10s %s\n", "─────", "──────", "────────")

	for _, p := range s.Phases {
		duration := p.Duration
		if duration == "" && p.Status == PhasePending {
			duration = "—"
		}
		if duration == "" {
			duration = "running…"
		}
		status := string(p.Status)
		if p.Status == PhaseRunning {
			status = "running…"
		}
		out += fmt.Sprintf("%-15s %-10s %s\n", p.Name, status, duration)

		if p.Error != "" {
			out += fmt.Sprintf("  └─ Error: %s\n", p.Error)
		}
	}

	if s.CurrentPhase != "" {
		out += fmt.Sprintf("\nCurrently provisioning: %s\n", s.CurrentPhase)
	}

	out += fmt.Sprintf("\nLast updated: %s\n", s.UpdatedAt)
	return out
}