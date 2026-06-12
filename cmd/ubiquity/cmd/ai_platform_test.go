package cmd

import (
	"context"
	"fmt"
	"testing"
)

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
	cmd.PersistentFlags().Set("profile", "ai-production")
	defer cmd.PersistentFlags().Set("profile", "gpu-basic")

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
	cmd.PersistentFlags().Set("profile", "ai-production")
	defer cmd.PersistentFlags().Set("profile", "gpu-basic")

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
	cmd.PersistentFlags().Set("profile", "ai-production")
	defer cmd.PersistentFlags().Set("profile", "gpu-basic")

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

func TestAIPlatformRenderAndApplySubcommandsAreActionable(t *testing.T) {
	render := findCommand(aiPlatformCmd, "render")
	if render == nil {
		t.Fatal("expected ai-platform render subcommand")
	}
	aiPlatformProfile = "ai-production"
	defer func() { aiPlatformProfile = "gpu-basic" }()
	output := captureOutput(func() {
		if err := render.RunE(render, []string{}); err != nil {
			t.Fatalf("render failed: %v", err)
		}
	})
	assertContains(t, output, "kind: ConfigMap")
	assertContains(t, output, "kind: Application")
	assertContains(t, output, "name: ai-platform-nvidia-gpu-operator")
	assertContains(t, output, "path: system/nvidia-gpu-operator")
	assertContains(t, output, "name: ai-platform-nvidia-network-operator")
	assertContains(t, output, "path: system/nvidia-network-operator")
	assertContains(t, output, "name: ai-platform-nim-operator")
	assertContains(t, output, "path: platform/nim-operator")
	assertContains(t, output, "name: ai-platform-kai-scheduler")
	assertContains(t, output, "path: platform/kai-scheduler")
	assertContains(t, output, "name: ai-platform-stallscope")
	assertContains(t, output, "path: platform/stallscope")
	assertContains(t, output, "namespace: gpu-telemetry")
	assertContains(t, output, "name: ai-platform-ai-workload-tenancy")
	assertContains(t, output, "path: platform/ai-workload-tenancy")
	assertContains(t, output, "name: ubiquity-ai-platform-profile")
	assertContains(t, output, "profile: ai-production")
	assertContains(t, output, "source: https://github.com/NVIDIA/gpu-operator")
	assertContains(t, output, "not-nvidia-certified: \"true\"")

	apply := findCommand(aiPlatformCmd, "apply")
	if apply == nil {
		t.Fatal("expected ai-platform apply subcommand")
	}
	oldRunner := runAIPlatformKubectl
	defer func() { runAIPlatformKubectl = oldRunner }()
	var gotArgs []string
	var gotManifest string
	runAIPlatformKubectl = func(ctx context.Context, args []string, stdin []byte) ([]byte, error) {
		gotArgs = append([]string(nil), args...)
		gotManifest = string(stdin)
		return []byte("configmap/ubiquity-ai-platform-profile configured\n"), nil
	}
	output = captureOutput(func() {
		if err := apply.RunE(apply, []string{}); err != nil {
			t.Fatalf("apply failed: %v", err)
		}
	})
	if fmt.Sprint(gotArgs) != "[apply --dry-run=server -f -]" {
		t.Fatalf("kubectl args = %#v", gotArgs)
	}
	assertContains(t, gotManifest, "profile: ai-production")
	assertContains(t, output, "configmap/ubiquity-ai-platform-profile configured")
}

func TestAIPlatformApplyCanMutateWithServerSideApply(t *testing.T) {
	apply := findCommand(aiPlatformCmd, "apply")
	if apply == nil {
		t.Fatal("expected ai-platform apply subcommand")
	}
	aiPlatformProfile = "ai-production"
	oldDryRun := aiPlatformApplyDryRun
	oldServerSide := aiPlatformApplyServerSide
	defer func() {
		aiPlatformProfile = "gpu-basic"
		aiPlatformApplyDryRun = oldDryRun
		aiPlatformApplyServerSide = oldServerSide
		apply.Flags().Set("dry-run", "true")
		apply.Flags().Set("server-side", "true")
	}()
	apply.Flags().Set("dry-run", "false")
	apply.Flags().Set("server-side", "true")

	oldRunner := runAIPlatformKubectl
	defer func() { runAIPlatformKubectl = oldRunner }()
	var gotArgs []string
	runAIPlatformKubectl = func(ctx context.Context, args []string, stdin []byte) ([]byte, error) {
		gotArgs = append([]string(nil), args...)
		return []byte("application/ai-platform-nvidia-gpu-operator configured\n"), nil
	}
	output := captureOutput(func() {
		if err := apply.RunE(apply, []string{}); err != nil {
			t.Fatalf("apply failed: %v", err)
		}
	})
	if fmt.Sprint(gotArgs) != "[apply --server-side -f -]" {
		t.Fatalf("kubectl args = %#v", gotArgs)
	}
	assertContains(t, output, "application/ai-platform-nvidia-gpu-operator configured")
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
