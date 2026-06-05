package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/ubiquitycluster/ubiquity/pkg/provision"
)

func isolateHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}

func requireLiveIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("UBIQUITY_LIVE_INTEGRATION") != "true" {
		t.Skip("skipping live cluster/provisioning integration test; set UBIQUITY_LIVE_INTEGRATION=true to run")
	}
}

func TestUpUsesProviderInterface(t *testing.T) {
	isolateHome(t)
	// Save original provider and restore after test
	origProvider := provider
	defer func() { provider = origProvider }()

	mock := &provision.MockProvider{}
	provider = mock

	// Set env flag on root (persistent flags) before subcommand runs
	rootCmd.PersistentFlags().Set("env", "sandbox")
	defer rootCmd.PersistentFlags().Set("env", "sandbox")

	// Execute the up command's RunE directly
	err := upCmd.RunE(upCmd, []string{})
	if err != nil {
		t.Fatalf("up command failed: %v", err)
	}

	// Verify phases were dispatched via the Provider
	expectedPhases := []string{"metal", "bootstrap", "security", "external", "wait", "post-install"}
	if len(mock.Calls) != len(expectedPhases) {
		t.Fatalf("expected %d calls, got %d: %v", len(expectedPhases), len(mock.Calls), mock.Calls)
	}
	for i, phase := range expectedPhases {
		expected := phase + ":sandbox"
		if mock.Calls[i] != expected {
			t.Errorf("call %d: expected %q, got %q", i, expected, mock.Calls[i])
		}
	}
}

func TestUpSandboxFlagSetsEnv(t *testing.T) {
	isolateHome(t)
	origProvider := provider
	defer func() { provider = origProvider }()

	mock := &provision.MockProvider{}
	provider = mock

	// Execute with sandbox flag set
	upCmd.Flags().Set("sandbox", "true")
	upCmd.Flags().Set("skip-security", "true")
	defer upCmd.Flags().Set("sandbox", "false")
	defer upCmd.Flags().Set("skip-security", "false")

	err := upCmd.RunE(upCmd, []string{})
	if err != nil {
		t.Fatalf("up --sandbox failed: %v", err)
	}

	if len(mock.Calls) == 0 {
		t.Error("expected at least one provider call for sandbox mode")
	}
}

func TestProvisionMetalSandbox(t *testing.T) {
	isolateHome(t)
	requireLiveIntegration(t)
	// Avoid creating or mutating a real k3d/docker cluster from unit tests.
	if err := exec.Command("kubectl", "cluster-info").Run(); err != nil {
		t.Skip("no live Kubernetes cluster; skipping sandbox provisioning side effects")
	}
	err := provisionMetal("sandbox")
	if err != nil {
		t.Fatalf("provisionMetal sandbox failed: %v", err)
	}
}

func TestReadK3sImage(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "commented file",
			input:    []byte("# Default k3s image\n# Override with K3S_IMAGE\nrancher/k3s:v1.32.13-k3s1\n"),
			expected: "rancher/k3s:v1.32.13-k3s1",
		},
		{
			name:     "plain image",
			input:    []byte("rancher/k3s:v1.30.0-k3s1"),
			expected: "rancher/k3s:v1.30.0-k3s1",
		},
		{
			name:     "trailing whitespace",
			input:    []byte("  rancher/k3s:v1.29.0-k3s1  \n"),
			expected: "rancher/k3s:v1.29.0-k3s1",
		},
		{
			name:     "all comments",
			input:    []byte("# only comments\n# nothing else\n"),
			expected: "",
		},
		{
			name:     "empty",
			input:    []byte(""),
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := readK3sImage(tt.input)
			if got != tt.expected {
				t.Errorf("readK3sImage() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestProvisionMetalProd(t *testing.T) {
	old := upLifecycle
	defer func() { upLifecycle = old }()
	upLifecycle = upLifecycleOptions{}
	err := provisionMetal("prod")
	if err != nil {
		t.Fatalf("provisionMetal prod failed: %v", err)
	}
	if upLifecycle.MetalBootstrapBackend != "ansible" || upLifecycle.NodeLifecycleBackend != "nico" {
		t.Fatalf("prod defaults = %s/%s, want ansible/nico", upLifecycle.MetalBootstrapBackend, upLifecycle.NodeLifecycleBackend)
	}
}

func TestLifecycleDefaultsSandboxNoneProdNICO(t *testing.T) {
	old := upLifecycle
	defer func() { upLifecycle = old }()
	upLifecycle = upLifecycleOptions{}
	applyLifecycleDefaults("sandbox")
	if upLifecycle.MetalBootstrapBackend != "none" || upLifecycle.NodeLifecycleBackend != "none" {
		t.Fatalf("sandbox defaults = %s/%s, want none/none", upLifecycle.MetalBootstrapBackend, upLifecycle.NodeLifecycleBackend)
	}
	upLifecycle = upLifecycleOptions{}
	applyLifecycleDefaults("prod")
	if upLifecycle.MetalBootstrapBackend != "ansible" || upLifecycle.NodeLifecycleBackend != "nico" {
		t.Fatalf("prod defaults = %s/%s, want ansible/nico", upLifecycle.MetalBootstrapBackend, upLifecycle.NodeLifecycleBackend)
	}
}

func TestDescribeNICOInstallPathIsGitOpsOnly(t *testing.T) {
	old := upLifecycle
	defer func() { upLifecycle = old }()
	upLifecycle = upLifecycleOptions{NodeLifecycleBackend: "nico", NICOValues: "values-nico.yaml", NICORESTValues: "values-rest.yaml", NICOSite: "lab-a"}
	out := describeNICOInstallPath("prod")
	for _, want := range []string{"GitOps wrappers", "values-nico.yaml", "values-rest.yaml", "lab-a"} {
		if !strings.Contains(out, want) {
			t.Fatalf("describeNICOInstallPath missing %q in %q", want, out)
		}
	}
}

func TestValidateLifecycleBackendsRejectsBMOWithoutMigrationAcknowledgement(t *testing.T) {
	old := upLifecycle
	defer func() { upLifecycle = old }()
	t.Setenv("UBIQUITY_ALLOW_BMO_MIGRATION", "")
	upLifecycle = upLifecycleOptions{MetalBootstrapBackend: "none", NodeLifecycleBackend: "bmo"}
	if err := validateLifecycleBackends(); err == nil || !strings.Contains(err.Error(), "fallback/migration-only") {
		t.Fatalf("expected explicit BMO migration error, got %v", err)
	}
}

func TestValidateLifecycleBackendsAllowsBMOWithMigrationAcknowledgement(t *testing.T) {
	old := upLifecycle
	defer func() { upLifecycle = old }()
	t.Setenv("UBIQUITY_ALLOW_BMO_MIGRATION", "true")
	upLifecycle = upLifecycleOptions{MetalBootstrapBackend: "none", NodeLifecycleBackend: "bmo"}
	if err := validateLifecycleBackends(); err != nil {
		t.Fatalf("expected acknowledged BMO migration fallback to validate, got %v", err)
	}
}

func TestProvisionBootstrapSandbox(t *testing.T) {
	isolateHome(t)
	requireLiveIntegration(t)
	// Only run when explicitly enabled; helm install/apply may be attempted.
	if err := exec.Command("kubectl", "cluster-info").Run(); err != nil {
		t.Skip("no live Kubernetes cluster; skipping bootstrap provisioning side effects")
	}
	err := provisionBootstrap("sandbox")
	if err != nil {
		t.Fatalf("provisionBootstrap sandbox failed: %v", err)
	}
}

func TestApplySandboxCharts(t *testing.T) {
	// applySandboxCharts has real cluster side effects when kubectl can reach a
	// live context, so unit tests should verify the deploy plan instead. Live
	// application is covered by the k3d sandbox proof path.
	targets, err := collectSandboxDeployTargets()
	if err != nil {
		t.Fatalf("collectSandboxDeployTargets failed: %v", err)
	}
	if len(targets) == 0 {
		t.Fatal("expected sandbox deploy targets")
	}
}

func TestDecryptSopsSecrets(t *testing.T) {
	isolateHome(t)
	err := decryptSopsSecrets()
	// Expected: error if sops not installed (which is the case in test env)
	t.Logf("decryptSopsSecrets result: %v", err)
}

func TestProvisionSecurity(t *testing.T) {
	isolateHome(t)
	requireLiveIntegration(t)
	err := provisionSecurity("sandbox")
	if err != nil {
		t.Fatalf("provisionSecurity failed: %v", err)
	}
}

func TestProvisionExternal(t *testing.T) {
	isolateHome(t)
	err := provisionExternal("sandbox")
	if err != nil {
		t.Fatalf("provisionExternal failed: %v", err)
	}
}

func TestProvisionWait(t *testing.T) {
	isolateHome(t)
	requireLiveIntegration(t)
	err := provisionWait("sandbox")
	if err != nil {
		t.Fatalf("provisionWait failed: %v", err)
	}
}

func TestProvisionPostInstallSandboxDoesNotInstallNICO(t *testing.T) {
	oldLifecycle := upLifecycle
	oldRunner := nicoGitOpsInstall
	defer func() {
		upLifecycle = oldLifecycle
		nicoGitOpsInstall = oldRunner
	}()
	upLifecycle = upLifecycleOptions{}
	var calls []nicoGitOpsTarget
	nicoGitOpsInstall = func(target nicoGitOpsTarget) error {
		calls = append(calls, target)
		return nil
	}

	err := provisionPostInstall("sandbox")
	if err != nil {
		t.Fatalf("provisionPostInstall sandbox failed: %v", err)
	}
	if len(calls) != 0 {
		t.Fatalf("sandbox/dev default node lifecycle none should not install NICo, got calls: %#v", calls)
	}
}

func TestProvisionPostInstallProdInstallsNICOGitOpsWrappers(t *testing.T) {
	oldLifecycle := upLifecycle
	oldRunner := nicoGitOpsInstall
	defer func() {
		upLifecycle = oldLifecycle
		nicoGitOpsInstall = oldRunner
	}()
	upLifecycle = upLifecycleOptions{NICOValues: "values-nico.yaml", NICORESTValues: "values-rest.yaml", NICOSite: "lab-a"}
	var calls []nicoGitOpsTarget
	nicoGitOpsInstall = func(target nicoGitOpsTarget) error {
		calls = append(calls, target)
		return nil
	}

	if err := provisionPostInstall("prod"); err != nil {
		t.Fatalf("provisionPostInstall prod failed: %v", err)
	}

	gotCharts := make([]string, 0, len(calls))
	for _, call := range calls {
		gotCharts = append(gotCharts, call.ChartDir)
	}
	wantCharts := []string{
		"system/nvidia-infra-controller-prereqs",
		"system/nvidia-infra-controller-core",
		"platform/nvidia-infra-controller-rest",
		"platform/node-lifecycle-exporter",
	}
	if !reflect.DeepEqual(gotCharts, wantCharts) {
		t.Fatalf("NICo GitOps wrapper chart calls = %#v, want %#v", gotCharts, wantCharts)
	}
	if calls[1].ValuesFiles[0] != "values-nico.yaml" {
		t.Fatalf("core values files = %#v, want values-nico.yaml", calls[1].ValuesFiles)
	}
	if calls[2].ValuesFiles[0] != "values-rest.yaml" {
		t.Fatalf("rest values files = %#v, want values-rest.yaml", calls[2].ValuesFiles)
	}
	for _, call := range calls {
		if call.Site != "lab-a" {
			t.Fatalf("site not propagated to %#v", call)
		}
	}
}

func TestProvisionPostInstallBMORequiresAcknowledgementAndWarnsFallbackOnly(t *testing.T) {
	oldLifecycle := upLifecycle
	defer func() { upLifecycle = oldLifecycle }()
	t.Setenv("UBIQUITY_ALLOW_BMO_MIGRATION", "")
	upLifecycle = upLifecycleOptions{MetalBootstrapBackend: "none", NodeLifecycleBackend: "bmo"}
	if err := validateLifecycleBackends(); err == nil || !strings.Contains(err.Error(), "fallback/migration-only") {
		t.Fatalf("expected BMO fallback-only error without acknowledgement, got %v", err)
	}

	t.Setenv("UBIQUITY_ALLOW_BMO_MIGRATION", "true")
	out := captureStdout(t, func() {
		if err := provisionPostInstall("prod"); err != nil {
			t.Fatalf("provisionPostInstall BMO fallback failed: %v", err)
		}
	})
	for _, want := range []string{"BMO", "fallback/migration-only"} {
		if !strings.Contains(out, want) {
			t.Fatalf("BMO fallback warning missing %q in %q", want, out)
		}
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	return buf.String()
}
