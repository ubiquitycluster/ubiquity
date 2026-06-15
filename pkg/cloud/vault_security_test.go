package cloud

import (
	"strings"
	"testing"
)

func TestVaultSecretGenerationDoesNotBuildSourceInClusterWithRootToken(t *testing.T) {
	job := mustRead(t, "../../system/vault/templates/generate-secrets-job.yaml")
	values := mustRead(t, "../../system/vault/values.yaml")
	for _, forbidden := range []string{"golang:1.17-alpine", "go get .", "go run .", "vault-root", "vault-unseal-keys"} {
		if strings.Contains(job, forbidden) {
			t.Fatalf("generate-secrets job retains unsafe runtime/root-token pattern %q", forbidden)
		}
	}
	for _, required := range []string{"vault-secret-generator", "vault-secret-generator-token", "readOnlyRootFilesystem: true", "allowPrivilegeEscalation: false"} {
		if !strings.Contains(job+values, required) {
			t.Fatalf("generate-secrets job missing hardened pattern %q", required)
		}
	}
}

func TestVaultAndExternalSecretsDoNotUseRootTokenAsSteadyStateCredential(t *testing.T) {
	vaultCR := mustRead(t, "../../system/vault/templates/cr.yaml")
	if strings.Contains(vaultCR, "storeRootToken: true") {
		t.Fatal("Vault CR stores the root token by default")
	}
	if !strings.Contains(vaultCR, "storeRootToken: false") {
		t.Fatal("Vault CR should explicitly disable root-token storage by default")
	}

	store := mustRead(t, "../../platform/external-secrets/templates/clustersecretstore.yaml")
	for _, forbidden := range []string{"vault-root", "vault-unseal-keys"} {
		if strings.Contains(store, forbidden) {
			t.Fatalf("External Secrets store retains root-token reference %q", forbidden)
		}
	}
	for _, required := range []string{"vault-external-secrets-token", "key: token"} {
		if !strings.Contains(store, required) {
			t.Fatalf("External Secrets store missing scoped-token reference %q", required)
		}
	}
}
