package cloud

import (
	"strings"
	"testing"
)

func TestProductionSecurityCIIncludesSBOMAndSigning(t *testing.T) {
	workflow := mustRead(t, "../../.github/workflows/ci.yaml")
	script := mustRead(t, "../../test/e2e/sbom-and-signing-proof.sh")
	for path, content := range map[string]string{
		"ci":     workflow,
		"script": script,
	} {
		for _, required := range []string{"syft", "cyclonedx", "cosign sign", "cosign verify", "sbom.spdx.json", "sbom.cyclonedx.json", "--dry-run"} {
			if !strings.Contains(content, required) {
				t.Fatalf("%s missing %q", path, required)
			}
		}
	}
}

func TestKyvernoPolicyTestsCoverAllowAndDenyFixtures(t *testing.T) {
	workflow := mustRead(t, "../../.github/workflows/ci.yaml")
	testSpec := mustRead(t, "../../system/kyverno-policies/kyverno-test/kyverno-test.yaml")
	for _, required := range []string{"kyverno test system/kyverno-policies/kyverno-test", "valid-pod", "privileged-pod", "missing-label-pod", "default-network-policy-denies", "result: pass", "result: fail"} {
		if !strings.Contains(workflow+"\n"+testSpec, required) {
			t.Fatalf("kyverno CI/test coverage missing %q", required)
		}
	}
}

func TestRuntimeSecurityValidationIncludesFalcoAlertingAndDashboard(t *testing.T) {
	workflow := mustRead(t, "../../.github/workflows/ci.yaml")
	script := mustRead(t, "../../test/e2e/runtime-security-validation.sh")
	rules := mustRead(t, "../../system/falco-rules/templates/rules.yaml")
	dashboard := mustRead(t, "../../grafana/dashboards/cluster.json")
	combined := workflow + "\n" + script + "\n" + rules + "\n" + dashboard
	for _, required := range []string{"falco-rules", "Falco", "Falcosidekick", "Alertmanager", "Terminal shell in container", "runtime-security-validation.sh --dry-run", "falco_events_total", "--dry-run"} {
		if !strings.Contains(combined, required) {
			t.Fatalf("runtime security validation missing %q", required)
		}
	}
}
