package aiplatform

import (
	"strings"
	"testing"
)

func TestTelemetryProfilesIncludeStallscopeWorkloadProfiler(t *testing.T) {
	for _, profileName := range []string{"gpu-basic", "gpu-rdma", "gpu-mig", "ai-production"} {
		t.Run(profileName, func(t *testing.T) {
			profile, err := GetProfile(profileName)
			if err != nil {
				t.Fatalf("GetProfile(%q): %v", profileName, err)
			}
			if !profile.HasCapability(CapabilityTelemetry) {
				t.Fatalf("%s must retain telemetry capability", profileName)
			}
			component, ok := profile.ComponentsByName()["stallscope"]
			if !ok {
				t.Fatalf("%s must include stallscope for GPU workload stall telemetry", profileName)
			}
			if component.SourceRepo != "https://github.com/nshinde/stallscope" {
				t.Fatalf("stallscope source repo must be traceable to upstream, got %q", component.SourceRepo)
			}
			if component.ChartRepository != "file://platform/stallscope" || component.ChartName != "stallscope" {
				t.Fatalf("stallscope must be wired to the local wrapper chart, got chart=%q repo=%q", component.ChartName, component.ChartRepository)
			}
			if component.Namespace != "gpu-telemetry" {
				t.Fatalf("stallscope telemetry namespace must be gpu-telemetry, got %q", component.Namespace)
			}
			if !component.ProductionDefault {
				t.Fatalf("stallscope should be a production telemetry default for GPU workload diagnostics")
			}
			for _, required := range []string{"RDMA", "PFC", "performance"} {
				if !strings.Contains(component.Notes, required) {
					t.Fatalf("stallscope notes must mention %s diagnostics, got %q", required, component.Notes)
				}
			}
		})
	}
}
