package aiplatform

import (
	"strings"
	"testing"
)

func TestGatedGPUE2EScriptExistsAndRequiresExplicitFlag(t *testing.T) {
	script := mustRead(t, "../../test/e2e/nvidia-ai-platform.sh")
	for _, required := range []string{
		"UBIQUITY_RUN_GPU_E2E=true",
		"nvidia-smi",
		"dcgm",
		"nim-smoke-test",
		"nvidia.com/rdma",
		"nvidia.com/mig-",
		"network-attachment-definitions.k8s.cni.cncf.io",
		"readyReplicas",
		"availableReplicas",
		"numberAvailable",
		"rdma-network-smoke-test-passed",
		"kai-scheduler-default",
		"default-queue",
		"kai-scheduling-smoke-test-passed",
		"exit 0",
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("GPU E2E script must include %q", required)
		}
	}
}

func TestNvidiaAIPlatformDocsIncludeProvenanceAndOperations(t *testing.T) {
	doc := mustRead(t, "../../docs/admin-guide/nvidia-ai-platform.md")
	for _, required := range []string{
		"Supported NVIDIA components",
		"NVIDIA/gpu-operator",
		"NVIDIA/k8s-nim-operator",
		"NVIDIA/network-operator",
		"nvidia.com/rdma",
		"nvidia.com/mig-",
		"NetworkAttachmentDefinition",
		"rdma-network-smoke-test-passed",
		"NGC credentials",
		"No NVIDIA approval or certification",
		"ubiquity ai-platform",
		"ubiquity test --sandbox-deploy",
		"Quick start",
		"What sandbox deploy proves",
		"How sandbox deploy works",
		"Production deployment flow",
		"NVIDIA/KAI-Scheduler",
		"NVIDIA/deepops",
		"NVIDIA/cloud-native-stack",
		"Longhorn",
		"not a generic PVC replacement",
		"kubectl apply --server-side --force-conflicts",
		"Troubleshooting",
		"ubiquity health",
	} {
		if !strings.Contains(doc, required) {
			t.Fatalf("NVIDIA AI platform docs must include %q", required)
		}
	}
}

func TestCIContainsGatedGPUE2EJob(t *testing.T) {
	workflow := mustRead(t, "../../.github/workflows/ci.yaml")
	for _, required := range []string{"gpu-e2e", "UBIQUITY_RUN_GPU_E2E", "test/e2e/nvidia-ai-platform.sh"} {
		if !strings.Contains(workflow, required) {
			t.Fatalf("CI workflow must include gated GPU E2E marker %q", required)
		}
	}
}
