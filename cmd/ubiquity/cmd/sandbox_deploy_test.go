package cmd

import "testing"

func TestSandboxDeployTargetsIncludeNvidiaAIComponents(t *testing.T) {
	targets, err := collectSandboxDeployTargets()
	if err != nil {
		t.Fatalf("collect sandbox deploy targets: %v", err)
	}

	want := map[string]string{
		"system/nvidia-gpu-operator":     "gpu-operator",
		"platform/nim-operator":          "nim-operator",
		"platform/ai-workload-tenancy":   "ai-workload-tenancy",
		"platform/stallscope":            "gpu-telemetry",
		"system/nvidia-network-operator": "nvidia-network-operator",
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
		{Stack: "platform", Name: "nim-operator", ChartDir: "platform/nim-operator", Namespace: "nim-operator"},
		{Stack: "platform", Name: "ai-workload-tenancy", ChartDir: "platform/ai-workload-tenancy", Namespace: "ai-workload-tenancy"},
		{Stack: "platform", Name: "stallscope", ChartDir: "platform/stallscope", Namespace: "gpu-telemetry"},
	}); err != nil {
		t.Fatalf("NVIDIA AI sandbox charts should render without NVIDIA devices: %v", err)
	}
}
