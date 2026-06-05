package aiplatform

import "testing"

func TestBareMetalOrchestrationAlternativesAreNvidiaSourceBacked(t *testing.T) {
	alternatives := BareMetalOrchestrationAlternatives()
	if len(alternatives) < 5 {
		t.Fatalf("expected at least five NVIDIA bare-metal orchestration alternatives, got %d", len(alternatives))
	}

	byName := orchestrationAlternativesByName(alternatives)
	for _, required := range []string{"deepops", "cloud-native-stack", "gpu-operator", "network-operator", "kai-scheduler"} {
		alternative, ok := byName[required]
		if !ok {
			t.Fatalf("missing NVIDIA bare-metal orchestration alternative %q", required)
		}
		if alternative.SourceRepo == "" {
			t.Fatalf("%s must cite a source repository", required)
		}
		if alternative.Evaluation == "" {
			t.Fatalf("%s must include reviewer-visible evaluation notes", required)
		}
	}
}

func TestBareMetalOrchestrationDecisionReplacesWeakLocalPaths(t *testing.T) {
	byName := orchestrationAlternativesByName(BareMetalOrchestrationAlternatives())

	gpuOperator := byName["gpu-operator"]
	if gpuOperator.Decision != OrchestrationDecisionAdopt {
		t.Fatalf("gpu-operator should be adopted, got %q", gpuOperator.Decision)
	}
	if !gpuOperator.ReplacesLocal {
		t.Fatal("gpu-operator should replace bespoke GPU node enablement")
	}

	networkOperator := byName["network-operator"]
	if networkOperator.Decision != OrchestrationDecisionAdopt {
		t.Fatalf("network-operator should be adopted for RDMA profiles, got %q", networkOperator.Decision)
	}
	if !networkOperator.ReplacesLocal {
		t.Fatal("network-operator should replace bespoke RDMA/network attachment glue when RDMA is enabled")
	}

	deepOps := byName["deepops"]
	if deepOps.Decision != OrchestrationDecisionReference {
		t.Fatalf("deepops should be a source-backed reference path, got %q", deepOps.Decision)
	}
	if deepOps.ReplacesLocal {
		t.Fatal("deepops should not be treated as a drop-in replacement for Ubiquity's k3d/Cluster API sandbox path")
	}
}

func TestProductionProfileCarriesBareMetalOrchestrationCapability(t *testing.T) {
	profile, err := GetProfile("ai-production")
	if err != nil {
		t.Fatalf("GetProfile(ai-production) returned error: %v", err)
	}
	if !profile.HasCapability(CapabilityBareMetalOrchestration) {
		t.Fatal("ai-production should track bare-metal orchestration capability and alternatives")
	}
}

func orchestrationAlternativesByName(alternatives []OrchestrationAlternative) map[string]OrchestrationAlternative {
	byName := make(map[string]OrchestrationAlternative, len(alternatives))
	for _, alternative := range alternatives {
		byName[alternative.Name] = alternative
	}
	return byName
}
