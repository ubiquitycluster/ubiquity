package aiplatform

import (
	"strings"
	"testing"
)

func TestNetworkOperatorValuesAreProfileDriven(t *testing.T) {
	values := mustRead(t, "../../system/nvidia-network-operator/values.yaml")
	for _, required := range []string{
		"aiPlatformProfile:",
		"rdma:",
		"resourceName:",
		"networkAttachment:",
		"validation:",
		"nvidia.com/rdma",
	} {
		if !strings.Contains(values, required) {
			t.Fatalf("NVIDIA Network Operator values must include profile-driven RDMA key %q", required)
		}
	}
}

func TestProductionSchedulerTenancyManifestsExist(t *testing.T) {
	for _, path := range []string{
		"../../platform/ai-workload-tenancy/Chart.yaml",
		"../../platform/ai-workload-tenancy/values.yaml",
		"../../platform/ai-workload-tenancy/templates/gpu-quota.yaml",
		"../../platform/ai-workload-tenancy/templates/priorityclasses.yaml",
	} {
		content := mustRead(t, path)
		if strings.TrimSpace(content) == "" {
			t.Fatalf("%s must not be empty", path)
		}
	}
}

func TestAIStoreEvaluationDocumentExists(t *testing.T) {
	doc := mustRead(t, "../../docs/reference/nvidia-ai-platform/aistore-evaluation.md")
	for _, required := range []string{"NVIDIA/aistore", "NVIDIA/ais-k8s", "decision", "enable only"} {
		if !strings.Contains(doc, required) {
			t.Fatalf("AIStore evaluation must include %q", required)
		}
	}
}
