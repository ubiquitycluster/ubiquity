package cmd

import "testing"

func TestSandboxDeployTargetsIncludeNvidiaAIComponents(t *testing.T) {
	targets, err := collectSandboxDeployTargets()
	if err != nil {
		t.Fatalf("collect sandbox deploy targets: %v", err)
	}

	want := map[string]string{
		"platform/ai-platform-console":             "ai-platform",
		"platform/ai-workload-tenancy":             "ai-workloads",
		"platform/stallscope":                      "gpu-telemetry",
		"platform/nim-operator":                    "nim-operator",
		"system/nvidia-gpu-operator":               "gpu-operator",
		"system/nvidia-network-operator":           "nvidia-network-operator",
		"system/nvidia-nic-configuration-operator": "network-operator",
	}
	got := map[string]sandboxDeployTarget{}
	for _, target := range targets {
		got[target.ChartDir] = target
	}
	for chartDir, namespace := range want {
		target, ok := got[chartDir]
		if !ok {
			t.Fatalf("sandbox deploy target %s missing from plan", chartDir)
		}
		if target.Namespace != namespace {
			t.Fatalf("sandbox deploy target %s namespace = %q, want %q", chartDir, target.Namespace, namespace)
		}
	}
}

func TestValidateNvidiaAISandboxChartsRenderWithoutDevices(t *testing.T) {
	if err := validateSandboxDeployTargets([]sandboxDeployTarget{
		{Stack: "system", Name: "nvidia-gpu-operator", ChartDir: "system/nvidia-gpu-operator", Namespace: "gpu-operator"},
		{Stack: "system", Name: "nvidia-nic-configuration-operator", ChartDir: "system/nvidia-nic-configuration-operator", Namespace: "network-operator"},
		{Stack: "platform", Name: "nim-operator", ChartDir: "platform/nim-operator", Namespace: "nim-operator"},
		{Stack: "platform", Name: "stallscope", ChartDir: "platform/stallscope", Namespace: "gpu-telemetry"},
		{Stack: "platform", Name: "ai-platform-console", ChartDir: "platform/ai-platform-console", Namespace: "ai-platform"},
		{Stack: "platform", Name: "ai-workload-tenancy", ChartDir: "platform/ai-workload-tenancy", Namespace: "ai-workloads"},
	}); err != nil {
		t.Fatalf("NVIDIA AI sandbox charts should render without NVIDIA devices: %v", err)
	}
}
