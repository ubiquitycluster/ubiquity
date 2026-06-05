package aiplatform

import "testing"

func TestParseMIGAllocatableByNodeReadsOnlyPositiveNvidiaMIGResources(t *testing.T) {
	input := []byte(`{
		"items": [
			{"metadata":{"name":"cpu-1"},"status":{"allocatable":{"cpu":"64"}}},
			{"metadata":{"name":"mig-1"},"status":{"allocatable":{"nvidia.com/mig-1g.10gb":"7","nvidia.com/mig-2g.20gb":"0"}}},
			{"metadata":{"name":"mig-2"},"status":{"allocatable":{"nvidia.com/mig-3g.40gb":"2","nvidia.com/gpu":"0"}}}
		]
	}`)

	parsed, err := ParseMIGAllocatableByNode(input)
	if err != nil {
		t.Fatalf("ParseMIGAllocatableByNode returned unexpected error: %v", err)
	}
	if len(parsed) != 2 {
		t.Fatalf("expected MIG evidence for two nodes, got %#v", parsed)
	}
	if parsed["mig-1"]["nvidia.com/mig-1g.10gb"] != 7 {
		t.Fatalf("expected mig-1 1g resources, got %#v", parsed["mig-1"])
	}
	if _, ok := parsed["mig-1"]["nvidia.com/mig-2g.20gb"]; ok {
		t.Fatalf("zero-count MIG resources must not be counted: %#v", parsed["mig-1"])
	}
	if parsed["mig-2"]["nvidia.com/mig-3g.40gb"] != 2 {
		t.Fatalf("expected mig-2 3g resources, got %#v", parsed["mig-2"])
	}
}

func TestEvaluateReadinessAcceptsMIGAllocatableWhenFullGPUResourceIsAbsent(t *testing.T) {
	status := EvaluateReadiness(ClusterSnapshot{
		GPUOperatorReady:                 true,
		GPUDevicePluginReady:             true,
		DCGMMetricsScraped:               true,
		GPUAllocatableByNode:             map[string]int{},
		MIGAllocatableByNode:             map[string]map[string]int{"mig-node-1": {"nvidia.com/mig-1g.10gb": 7}},
		RDMAResourcesByNode:              map[string]int{"mig-node-1": 4},
		NetworkAttachments:               []string{"default/rdma-ipoib"},
		LastRDMASmokeTestPassed:          true,
		NIMServicesReady:                 1,
		LastNIMSmokeTestPassed:           true,
		KAISchedulerReady:                true,
		KAIQueueReady:                    true,
		LastKAISchedulingSmokeTestPassed: true,
	})
	if !status.Ready {
		t.Fatalf("expected MIG allocatable resources to satisfy GPU capacity readiness, got checks: %#v", status.Checks)
	}
	message := status.ChecksByName()["gpu-allocatable"].Message
	if message == "" || !containsAll(message, []string{"MIG", "7"}) {
		t.Fatalf("expected gpu-allocatable message to report MIG evidence, got %q", message)
	}
}

func containsAll(value string, needles []string) bool {
	for _, needle := range needles {
		if !contains(value, needle) {
			return false
		}
	}
	return true
}

func contains(value, needle string) bool {
	for i := 0; i+len(needle) <= len(value); i++ {
		if value[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
