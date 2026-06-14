package aiplatform

import "testing"

func TestEvaluateReadinessFailsClosedWhenGPUOperatorMissing(t *testing.T) {
	status := EvaluateReadiness(ClusterSnapshot{})
	if status.Ready {
		t.Fatal("readiness must fail closed when NVIDIA platform evidence is missing")
	}
	if status.ChecksByName()["gpu-operator"].Ready {
		t.Fatal("gpu-operator check should not be ready without evidence")
	}
}

func TestEvaluateReadinessRequiresNIMForProduction(t *testing.T) {
	status := EvaluateReadiness(ClusterSnapshot{
		GPUOperatorReady:            true,
		GPUDriverReady:              true,
		GPUContainerToolkitReady:    true,
		GPUDevicePluginReady:        true,
		GPUFeatureDiscoveryReady:    true,
		GPUManagedDCGMExporterReady: true,
		GPUMIGManagerReady:          true,
		GPUOperatorValidatorReady:   true,
		DCGMMetricsScraped:          true,
		GPUAllocatableByNode:        map[string]int{"gpu-node-1": 8},
		NIMServicesReady:            0,
		LastNIMSmokeTestPassed:      false,
	})
	if status.Ready {
		t.Fatal("platform should not be ready without a ready NIM service and smoke test")
	}
	if status.ChecksByName()["nim-serving"].Ready {
		t.Fatal("nim-serving check should fail when no NIM service is ready")
	}
}

func TestEvaluateReadinessReportsEveryGPUOperatorOperand(t *testing.T) {
	status := EvaluateReadiness(ClusterSnapshot{GPUOperatorReady: true})
	checks := status.ChecksByName()
	for _, name := range []string{
		"gpu-driver",
		"gpu-runtime-toolkit",
		"device-plugin",
		"gpu-feature-discovery",
		"gpu-dcgm-exporter",
		"gpu-mig-manager",
		"gpu-validator",
	} {
		check, ok := checks[name]
		if !ok {
			t.Fatalf("missing readiness check %q", name)
		}
		if check.Ready {
			t.Fatalf("%s should fail closed without operand evidence", name)
		}
	}
}

func TestEvaluateReadinessPassesWithCompleteEvidence(t *testing.T) {
	status := EvaluateReadiness(ClusterSnapshot{
		GPUOperatorReady:                 true,
		GPUDriverReady:                   true,
		GPUContainerToolkitReady:         true,
		GPUDevicePluginReady:             true,
		GPUFeatureDiscoveryReady:         true,
		GPUManagedDCGMExporterReady:      true,
		GPUMIGManagerReady:               true,
		GPUOperatorValidatorReady:        true,
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
		t.Fatalf("expected complete evidence to be ready, got checks: %#v", status.Checks)
	}
}
