package aiplatform

import (
	"os"
	"strings"
	"testing"
)

func TestNIMOperatorChartDefinesNGCSecretAndSampleServiceFlow(t *testing.T) {
	for _, path := range []string{
		"../../platform/nim-operator/values.yaml",
		"../../platform/nim-operator/templates/ngc-secrets.yaml",
		"../../platform/nim-operator/templates/sample-nimservice.yaml",
		"../../platform/nim-operator/templates/smoke-test.yaml",
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected NIM operator integration file %s: %v", path, err)
		}
	}

	values := mustRead(t, "../../platform/nim-operator/values.yaml")
	for _, required := range []string{"ngc:", "apiKeySecretName:", "pullSecretName:", "commitCredentials: false"} {
		if !strings.Contains(values, required) {
			t.Fatalf("NIM values must include safe NGC secret setting %q", required)
		}
	}

	sample := mustRead(t, "../../platform/nim-operator/templates/sample-nimservice.yaml")
	for _, required := range []string{"kind: NIMCache", "kind: NIMService", "apps.nvidia.com/v1alpha1", "nvidia.com/gpu"} {
		if !strings.Contains(sample, required) {
			t.Fatalf("sample NIM service template must include %q", required)
		}
	}
}

func TestNIMSmokeTestCreatesReadinessEvidenceConfigMap(t *testing.T) {
	smoke := mustRead(t, "../../platform/nim-operator/templates/smoke-test.yaml")
	for _, required := range []string{
		"kind: ServiceAccount",
		"kind: Role",
		"kind: RoleBinding",
		"kind: Job",
		"serviceAccountName:",
		"curl --fail",
		"kubectl create configmap",
		"successConfigMapName",
		"nim-smoke-test-passed",
	} {
		if !strings.Contains(smoke, required) {
			t.Fatalf("NIM smoke test template must include %q", required)
		}
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(content)
}
