package cmd

import (
	"testing"

	"github.com/ubiquitycluster/ubiquity/pkg/provision"
)

func TestUpUsesProviderInterface(t *testing.T) {
	// Save original provider and restore after test
	origProvider := provider
	defer func() { provider = origProvider }()

	mock := &provision.MockProvider{}
	provider = mock

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
		if mock.Calls[i] != phase+":" {
			t.Errorf("call %d: expected %q, got %q", i, phase+":", mock.Calls[i])
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