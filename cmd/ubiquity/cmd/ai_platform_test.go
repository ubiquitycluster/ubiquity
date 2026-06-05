package cmd

import "testing"

func TestAIPlatformCmdRegistered(t *testing.T) {
	cmd := findCommand(rootCmd, "ai-platform")
	if cmd == nil {
		t.Fatal("expected ai-platform command to be registered")
	}
	if cmd.Use != "ai-platform" {
		t.Fatalf("expected Use ai-platform, got %q", cmd.Use)
	}
}

func TestAIPlatformOutputIncludesBareMetalOrchestrationAlternatives(t *testing.T) {
	cmd := findCommand(rootCmd, "ai-platform")
	if cmd == nil {
		t.Fatal("expected ai-platform command to be registered")
	}
	cmd.Flags().Set("profile", "ai-production")
	defer cmd.Flags().Set("profile", "gpu-basic")

	output := captureOutput(func() {
		if err := cmd.RunE(cmd, []string{}); err != nil {
			t.Fatalf("ai-platform command failed: %v", err)
		}
	})

	assertContains(t, output, "Bare-metal orchestration alternatives:")
	assertContains(t, output, "deepops")
	assertContains(t, output, "cloud-native-stack")
	assertContains(t, output, "kai-scheduler")
	assertContains(t, output, "decision: reference")
	assertContains(t, output, "decision: evaluate")
}

func TestAIPlatformOutputIncludesStorageAlternatives(t *testing.T) {
	cmd := findCommand(rootCmd, "ai-platform")
	if cmd == nil {
		t.Fatal("expected ai-platform command to be registered")
	}
	cmd.Flags().Set("profile", "ai-production")
	defer cmd.Flags().Set("profile", "gpu-basic")

	output := captureOutput(func() {
		if err := cmd.RunE(cmd, []string{}); err != nil {
			t.Fatalf("ai-platform command failed: %v", err)
		}
	})

	assertContains(t, output, "Storage alternatives:")
	assertContains(t, output, "nvidia-aistore")
	assertContains(t, output, "adopt-for-ai-data-plane")
	assertContains(t, output, "replaces Longhorn for AI dataset/cache paths")
	assertContains(t, output, "not a generic PVC replacement")
}

func TestAIPlatformProfileOutputDemotesOllama(t *testing.T) {
	cmd := findCommand(rootCmd, "ai-platform")
	if cmd == nil {
		t.Fatal("expected ai-platform command to be registered")
	}
	cmd.Flags().Set("profile", "ai-production")
	defer cmd.Flags().Set("profile", "gpu-basic")

	output := captureOutput(func() {
		if err := cmd.RunE(cmd, []string{}); err != nil {
			t.Fatalf("ai-platform command failed: %v", err)
		}
	})

	assertContains(t, output, "Profile: ai-production")
	assertContains(t, output, "NVIDIA/gpu-operator")
	assertContains(t, output, "NVIDIA/k8s-nim-operator")
	assertContains(t, output, "Ollama: optional diagnostics only")
}

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !contains(haystack, needle) {
		t.Fatalf("expected output to contain %q, got:\n%s", needle, haystack)
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
