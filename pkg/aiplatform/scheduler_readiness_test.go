package aiplatform

import "testing"

func TestEvaluateReadinessRequiresKAISchedulerEvidenceForProductionScheduling(t *testing.T) {
	status := EvaluateReadiness(ClusterSnapshot{
		GPUOperatorReady:       true,
		GPUDevicePluginReady:   true,
		DCGMMetricsScraped:     true,
		GPUAllocatableByNode:   map[string]int{"gpu-node-1": 8},
		NIMServicesReady:       1,
		LastNIMSmokeTestPassed: true,
	})

	if status.Ready {
		t.Fatal("platform should not be ready without KAI Scheduler readiness and scheduling smoke-test evidence")
	}
	check := status.ChecksByName()["kai-scheduler"]
	if check.Ready {
		t.Fatal("kai-scheduler check should fail when scheduler evidence is missing")
	}
}

func TestEvaluateReadinessPassesWithKAISchedulerEvidence(t *testing.T) {
	status := EvaluateReadiness(ClusterSnapshot{
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

	if !status.Ready {
		t.Fatalf("expected complete KAI Scheduler evidence to be ready, got checks: %#v", status.Checks)
	}
	if !status.ChecksByName()["kai-scheduler"].Ready {
		t.Fatalf("kai-scheduler check should pass with scheduler, queue, and smoke-test evidence")
	}
}
