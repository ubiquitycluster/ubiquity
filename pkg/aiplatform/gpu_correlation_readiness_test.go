package aiplatform

import "testing"

func TestEvaluateReadinessFailsClosedOnNICoKubernetesGPUCorrelationIssue(t *testing.T) {
	status := EvaluateReadiness(ClusterSnapshot{
		GPUOperatorReady:                 true,
		GPUDevicePluginReady:             true,
		GPUAllocatableByNode:             map[string]int{"cn01": 8},
		RDMAResourcesByNode:              map[string]int{"cn01": 1},
		NetworkAttachments:               []string{"default/rdma"},
		LastRDMASmokeTestPassed:          true,
		DCGMMetricsScraped:               true,
		NIMServicesReady:                 1,
		LastNIMSmokeTestPassed:           true,
		KAISchedulerReady:                true,
		KAIQueueReady:                    true,
		LastKAISchedulingSmokeTestPassed: true,
		GPUCorrelationIssues:             []string{"cn01: NICo GPU inventory missing from Kubernetes"},
	})
	check := status.ChecksByName()["gpu-nico-kubernetes-correlation"]
	if status.Ready || check.Ready {
		t.Fatalf("expected GPU correlation issue to fail readiness, status=%+v check=%+v", status, check)
	}
}
