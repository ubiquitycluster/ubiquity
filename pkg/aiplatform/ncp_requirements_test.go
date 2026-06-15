package aiplatform

import "testing"

func TestNCPRequirementsCoverLayeredReferencePlatform(t *testing.T) {
	requirements := NCPRequirements()
	if len(requirements) < 6 {
		t.Fatalf("expected layered NCP requirement map, got %d entries", len(requirements))
	}
	seen := map[string]NCPRequirement{}
	for _, requirement := range requirements {
		if requirement.ID == "" || requirement.Layer == "" || requirement.Capability == "" || requirement.ReadinessSignal == "" {
			t.Fatalf("requirement must be self describing: %#v", requirement)
		}
		if len(requirement.UbiquityEvidence) == 0 {
			t.Fatalf("requirement %s must cite Ubiquity-owned evidence", requirement.ID)
		}
		seen[requirement.ID] = requirement
	}
	for _, id := range []string{
		"iaas-bare-metal-vm-lifecycle",
		"caas-gpu-kubernetes-substrate",
		"caas-rdma-networking",
		"paas-serving-scheduling",
		"tenant-workload-isolation",
		"observability-validation",
	} {
		if _, ok := seen[id]; !ok {
			t.Fatalf("missing NCP requirement %s", id)
		}
	}
}

func TestNCPRequirementsMapToImplementedUbiquityArtifacts(t *testing.T) {
	requirements := NCPRequirements()
	wantEvidence := []string{
		"system/nvidia-gpu-operator",
		"system/nvidia-network-operator",
		"system/nvidia-nic-configuration-operator",
		"platform/nim-operator",
		"platform/kai-scheduler",
		"platform/ai-workload-tenancy",
		"pkg/aiplatform/readiness.go",
		"test/e2e/nvidia-ai-platform-final-demo.sh",
	}
	for _, want := range wantEvidence {
		found := false
		for _, requirement := range requirements {
			for _, evidence := range requirement.UbiquityEvidence {
				if evidence == want {
					found = true
				}
			}
		}
		if !found {
			t.Fatalf("NCP requirement map must include evidence artifact %q", want)
		}
	}
}
