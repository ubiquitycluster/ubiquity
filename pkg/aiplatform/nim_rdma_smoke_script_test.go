package aiplatform

import (
	"strings"
	"testing"
)

func TestNIMGPUServingSmokeScriptIsGatedAndRecordsEvidence(t *testing.T) {
	script := mustRead(t, "../../test/e2e/nim-gpu-serving-smoke.sh")
	for _, required := range []string{
		"UBIQUITY_RUN_NIM_GPU_SMOKE=true",
		"nim-smoke-test-passed",
		"kubectl wait",
		"curl --fail",
		"NIMService",
		"--dry-run",
		"fail closed",
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("NIM GPU serving smoke script missing %q", required)
		}
	}
}

func TestNVIDIARDMASmokeScriptIsGatedAndRecordsEvidence(t *testing.T) {
	script := mustRead(t, "../../test/e2e/nvidia-rdma-smoke.sh")
	for _, required := range []string{
		"UBIQUITY_RUN_NVIDIA_RDMA_SMOKE=true",
		"rdma-network-smoke-test-passed",
		"nvidia.com/rdma",
		"network-attachment-definitions.k8s.cni.cncf.io",
		"kubectl apply",
		"--dry-run",
		"fail closed",
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("NVIDIA RDMA smoke script missing %q", required)
		}
	}
}
