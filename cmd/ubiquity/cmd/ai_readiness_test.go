package cmd

import (
	"testing"

	"github.com/ubiquitycluster/ubiquity/pkg/aiplatform"
)

func TestRenderAIReadinessStatusFailsClosed(t *testing.T) {
	output := renderAIReadinessStatus(aiplatform.EvaluateReadiness(aiplatform.ClusterSnapshot{}))
	assertContains(t, output, "NVIDIA AI platform readiness: NOT READY")
	assertContains(t, output, "gpu-operator: FAIL")
	assertContains(t, output, "no allocatable NVIDIA GPUs or MIG resources found")
}

func TestRenderAIStoreReadinessStatusReportsOptionalDataPlaneBoundary(t *testing.T) {
	output := renderAIStoreReadinessStatus(aiplatform.EvaluateAIStoreReadiness(aiplatform.AIStoreSnapshot{}))
	assertContains(t, output, "NVIDIA AIStore data-plane readiness: NOT READY")
	assertContains(t, output, "optional AI dataset/cache/object path")
	assertContains(t, output, "not a generic PVC replacement")
	assertContains(t, output, "aistore-operator: FAIL")
}

func TestRenderAIReadinessStatusReady(t *testing.T) {
	status := aiplatform.EvaluateReadiness(aiplatform.ClusterSnapshot{
		GPUOperatorReady:                 true,
		GPUDevicePluginReady:             true,
		DCGMMetricsScraped:               true,
		GPUAllocatableByNode:             map[string]int{"gpu-node-1": 8},
		RDMAResourcesByNode:              map[string]int{"gpu-node-1": 4},
		NetworkAttachments:               []string{"default/rdma-ipoib"},
		LastRDMASmokeTestPassed:          true,
		NIMServicesReady:                 1,
		LastNIMSmokeTestPassed:           true,
		KAISchedulerReady:                true,
		KAIQueueReady:                    true,
		LastKAISchedulingSmokeTestPassed: true,
	})
	output := renderAIReadinessStatus(status)
	assertContains(t, output, "NVIDIA AI platform readiness: READY")
	assertContains(t, output, "nim-serving: OK")
}
