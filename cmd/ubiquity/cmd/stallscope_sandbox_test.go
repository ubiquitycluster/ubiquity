package cmd

import "testing"

func TestSandboxDeployTargetsIncludeStallscope(t *testing.T) {
	targets, err := collectSandboxDeployTargets()
	if err != nil {
		t.Fatalf("collectSandboxDeployTargets returned error: %v", err)
	}
	var found bool
	for _, target := range filterNvidiaAISandboxDeployTargets(targets) {
		if target.ChartDir == "platform/stallscope" {
			found = true
			if target.Namespace != "gpu-telemetry" {
				t.Fatalf("expected gpu-telemetry namespace, got %q", target.Namespace)
			}
			if target.Kind != "helm" {
				t.Fatalf("expected Stallscope to render as Helm target, got %q", target.Kind)
			}
		}
	}
	if !found {
		t.Fatal("NVIDIA AI sandbox targets should include platform/stallscope")
	}
}
