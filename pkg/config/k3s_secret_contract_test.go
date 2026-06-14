package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestK3SEncryptionSecretIsNotCommittedAsLiteral(t *testing.T) {
	root := configRepoRoot(t)

	for _, rel := range []string{
		"metal/group_vars/metal.yml",
		"metal/roles/k3s/defaults/main.yml",
	} {
		content := readConfigTestFile(t, root, rel)
		if strings.Contains(content, "TODO") && strings.Contains(content, "k3s_encryption_secret") {
			t.Fatalf("%s keeps k3s_encryption_secret as TODO debt", rel)
		}
		for _, forbidden := range []string{
			"dHBhczJxN2FvOW52a2pzcWc3cWd1YXoyamttNGNrc2w=",
			"qh5+jYTGNBfcimR1C09yqnE6H6218M48WBjnCGiDCn0=",
		} {
			if strings.Contains(content, forbidden) {
				t.Fatalf("%s contains a committed K3s encryption secret fixture", rel)
			}
		}
	}
}

func TestK3SEncryptionSecretFailsClosedBeforeTemplateDeployment(t *testing.T) {
	root := configRepoRoot(t)
	content := readConfigTestFile(t, root, "metal/roles/k3s/tasks/cluster_bootstrap.yml")

	secretAssert := strings.Index(content, "Validate K3s encryption secret is provided")
	deployTemplate := strings.Index(content, "Deploy encryption provider config")
	if secretAssert == -1 {
		t.Fatal("cluster bootstrap must validate k3s_encryption_secret before rendering encryption config")
	}
	if deployTemplate == -1 {
		t.Fatal("cluster bootstrap must still deploy the encryption provider config")
	}
	if secretAssert > deployTemplate {
		t.Fatal("k3s_encryption_secret validation must run before encryption provider config deployment")
	}

	for _, required := range []string{
		"k3s_encryption_secret is defined",
		"k3s_encryption_secret != ''",
		"not k3s_encryption_secret.startswith('CHANGEME')",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("cluster bootstrap validation missing assertion %q", required)
		}
	}
}

func configRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func readConfigTestFile(t *testing.T, root, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(b)
}
