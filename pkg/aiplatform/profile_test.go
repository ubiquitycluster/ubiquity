package aiplatform

import "testing"

func TestProfileIncludesNvidiaSourceBackedComponents(t *testing.T) {
	profile, err := GetProfile("gpu-basic")
	if err != nil {
		t.Fatalf("GetProfile(gpu-basic) returned error: %v", err)
	}

	components := profile.ComponentsByName()
	gpuOperator := components["gpu-operator"]
	if gpuOperator.SourceRepo != "https://github.com/NVIDIA/gpu-operator" {
		t.Fatalf("gpu-basic must use NVIDIA/gpu-operator as source of truth, got %q", gpuOperator.SourceRepo)
	}
	if gpuOperator.ChartRepository != "https://helm.ngc.nvidia.com/nvidia" {
		t.Fatalf("gpu-operator chart repository mismatch: %q", gpuOperator.ChartRepository)
	}
	if !gpuOperator.ReplacesLocal {
		t.Fatal("gpu-operator should replace local/bespoke GPU enablement")
	}

	dcgm := components["dcgm-exporter"]
	if dcgm.SourceRepo != "https://github.com/NVIDIA/dcgm-exporter" {
		t.Fatalf("dcgm-exporter must be source-backed by NVIDIA/dcgm-exporter, got %q", dcgm.SourceRepo)
	}
	if !dcgm.ManagedByGPUOperator {
		t.Fatal("dcgm-exporter should be managed through GPU Operator by default")
	}
}

func TestProductionProfileRequiresServingNetworkAndValidation(t *testing.T) {
	profile, err := GetProfile("ai-production")
	if err != nil {
		t.Fatalf("GetProfile(ai-production) returned error: %v", err)
	}

	for _, required := range []Capability{CapabilityGPU, CapabilityRDMA, CapabilityServing, CapabilityTelemetry, CapabilityValidation} {
		if !profile.HasCapability(required) {
			t.Fatalf("ai-production missing required capability %q", required)
		}
	}

	components := profile.ComponentsByName()
	if components["nim-operator"].SourceRepo != "https://github.com/NVIDIA/k8s-nim-operator" {
		t.Fatalf("ai-production must use NVIDIA/k8s-nim-operator for production serving")
	}
	if components["ollama"].ProductionDefault {
		t.Fatal("Ollama must not be the production AI serving default")
	}
}

func TestAIProductionProfileIncludesKubeVirtVirtualization(t *testing.T) {
	profile, err := GetProfile("ai-production")
	if err != nil {
		t.Fatalf("GetProfile(ai-production) returned error: %v", err)
	}
	if !profile.HasCapability(CapabilityVirtualization) {
		t.Fatal("ai-production profile must include virtualization capability for KubeVirt VMs")
	}
	components := profile.ComponentsByName()
	for _, name := range []string{"kubevirt", "containerized-data-importer", "multus-cni"} {
		component, ok := components[name]
		if !ok {
			t.Fatalf("ai-production profile missing %s component", name)
		}
		if component.SourceRepo == "" {
			t.Fatalf("%s component must include source repo", name)
		}
	}
}

func TestUnknownProfileFailsClosed(t *testing.T) {
	_, err := GetProfile("unknown")
	if err == nil {
		t.Fatal("unknown profile should return an error")
	}
}
