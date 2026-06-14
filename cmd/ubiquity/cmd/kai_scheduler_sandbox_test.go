package cmd

import "testing"

func TestSandboxDeployTargetsIncludeKAIScheduler(t *testing.T) {
	targets, err := collectSandboxDeployTargets()
	if err != nil {
		t.Fatalf("collectSandboxDeployTargets returned error: %v", err)
	}
	var found bool
	for _, target := range filterNvidiaAISandboxDeployTargets(targets) {
		if target.ChartDir == "platform/kai-scheduler" {
			found = true
			if target.Namespace != "kai-scheduler" {
				t.Fatalf("expected kai-scheduler namespace, got %q", target.Namespace)
			}
		}
	}
	if !found {
		t.Fatal("NVIDIA AI sandbox targets should include platform/kai-scheduler")
	}
}
