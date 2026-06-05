package aiplatform

import "testing"

func TestEvaluateReadinessRequiresRDMAResourceEvidence(t *testing.T) {
	status := EvaluateReadiness(ClusterSnapshot{
		GPUOperatorReady:                 true,
		GPUDevicePluginReady:             true,
		DCGMMetricsScraped:               true,
		GPUAllocatableByNode:             map[string]int{"gpu-node-1": 8},
		NetworkAttachments:               []string{"default/rdma-ipoib"},
		NIMServicesReady:                 1,
		LastNIMSmokeTestPassed:           true,
		KAISchedulerReady:                true,
		KAIQueueReady:                    true,
		LastKAISchedulingSmokeTestPassed: true,
	})
	if status.Ready {
		t.Fatal("platform should not be ready without RDMA resource evidence")
	}
	check := status.ChecksByName()["rdma-network"]
	if check.Ready || check.Message == "" {
		t.Fatalf("expected rdma-network check to fail explicitly, got %#v", check)
	}
}

func TestEvaluateReadinessRequiresNetworkAttachmentEvidence(t *testing.T) {
	status := EvaluateReadiness(ClusterSnapshot{
		GPUOperatorReady:                 true,
		GPUDevicePluginReady:             true,
		DCGMMetricsScraped:               true,
		GPUAllocatableByNode:             map[string]int{"gpu-node-1": 8},
		RDMAResourcesByNode:              map[string]int{"gpu-node-1": 4},
		NIMServicesReady:                 1,
		LastNIMSmokeTestPassed:           true,
		KAISchedulerReady:                true,
		KAIQueueReady:                    true,
		LastKAISchedulingSmokeTestPassed: true,
	})
	if status.Ready {
		t.Fatal("platform should not be ready without RDMA NetworkAttachmentDefinition evidence")
	}
	check := status.ChecksByName()["rdma-network"]
	if check.Ready || check.Message == "" {
		t.Fatalf("expected rdma-network check to fail explicitly, got %#v", check)
	}
}

func TestEvaluateReadinessRequiresRDMASmokeTestEvidence(t *testing.T) {
	status := EvaluateReadiness(ClusterSnapshot{
		GPUOperatorReady:                 true,
		GPUDevicePluginReady:             true,
		DCGMMetricsScraped:               true,
		GPUAllocatableByNode:             map[string]int{"gpu-node-1": 8},
		RDMAResourcesByNode:              map[string]int{"gpu-node-1": 4},
		NetworkAttachments:               []string{"default/rdma-ipoib"},
		NIMServicesReady:                 1,
		LastNIMSmokeTestPassed:           true,
		KAISchedulerReady:                true,
		KAIQueueReady:                    true,
		LastKAISchedulingSmokeTestPassed: true,
	})
	if status.Ready {
		t.Fatal("platform should not be ready without RDMA smoke-test evidence")
	}
	check := status.ChecksByName()["rdma-network"]
	if check.Ready || check.Message == "" {
		t.Fatalf("expected rdma-network check to fail explicitly, got %#v", check)
	}
}

func TestEvaluateReadinessAcceptsCompleteRDMAEvidence(t *testing.T) {
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
		t.Fatalf("expected complete RDMA evidence to be ready, got %#v", status.Checks)
	}
}
