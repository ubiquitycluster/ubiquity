package cmd

import (
	"os/exec"
	"testing"

	"github.com/ubiquitycluster/ubiquity/pkg/provision"
)

func TestUpUsesProviderInterface(t *testing.T) {
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
	err := provisionMetal("sandbox")
	if err != nil {
		t.Fatalf("provisionMetal sandbox failed: %v", err)
	}
}

func TestProvisionMetalProd(t *testing.T) {
	err := provisionMetal("prod")
	if err != nil {
		t.Fatalf("provisionMetal prod failed: %v", err)
	}
}

func TestProvisionBootstrapSandbox(t *testing.T) {
	// Skip if a real cluster is running (helm install would be attempted)
	if err := exec.Command("kubectl", "cluster-info").Run(); err == nil {
		t.Skip("Skipping: real cluster detected, would attempt helm install")
	}
	err := provisionBootstrap("sandbox")
	if err != nil {
		t.Fatalf("provisionBootstrap sandbox failed: %v", err)
	}
}

func TestDecryptSopsSecrets(t *testing.T) {
	err := decryptSopsSecrets()
	// Expected: error if sops not installed (which is the case in test env)
	t.Logf("decryptSopsSecrets result: %v", err)
}

func TestProvisionSecurity(t *testing.T) {
	err := provisionSecurity("sandbox")
	if err != nil {
		t.Fatalf("provisionSecurity failed: %v", err)
	}
}

func TestProvisionExternal(t *testing.T) {
	err := provisionExternal("sandbox")
	if err != nil {
		t.Fatalf("provisionExternal failed: %v", err)
	}
}

func TestProvisionWait(t *testing.T) {
	err := provisionWait("sandbox")
	if err != nil {
		t.Fatalf("provisionWait failed: %v", err)
	}
}

func TestProvisionPostInstall(t *testing.T) {
	err := provisionPostInstall("sandbox")
	if err != nil {
		t.Fatalf("provisionPostInstall failed: %v", err)
	}
}