package aiplatform

import "testing"

func TestProductionProfileUsesKAISchedulerForAIWorkloadScheduling(t *testing.T) {
	profile, err := GetProfile("ai-production")
	if err != nil {
		t.Fatalf("GetProfile(ai-production) returned error: %v", err)
	}

	component := profile.ComponentsByName()["kai-scheduler"]
	if component.SourceRepo != "https://github.com/NVIDIA/KAI-Scheduler" {
		t.Fatalf("ai-production should track NVIDIA/KAI-Scheduler, got %q", component.SourceRepo)
	}
	if component.ChartRepository != "oci://ghcr.io/kai-scheduler/kai-scheduler" {
		t.Fatalf("unexpected KAI Scheduler chart repository %q", component.ChartRepository)
	}
	if !component.ReplacesLocal {
		t.Fatal("KAI Scheduler should replace local priority/quota-only scheduling for production AI workloads")
	}
	if !component.ProductionDefault {
		t.Fatal("KAI Scheduler should be the production default scheduler component for ai-production")
	}
}
