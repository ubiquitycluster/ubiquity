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

func TestScheduledDependencyFreshnessReportWorkflow(t *testing.T) {
	workflow := mustRead(t, "../../.github/workflows/dependency-freshness.yml")
	script := mustRead(t, "../../test/e2e/dependency-freshness-report.sh")
	combined := workflow + "\n" + script
	for _, required := range []string{"schedule:", "workflow_dispatch:", "dependency-freshness-report.sh --dry-run", "go list -m -u all", "helm dependency list", "renovate", "dependabot", "actions", "container images", "dependency-freshness-report.md"} {
		if !strings.Contains(combined, required) {
			t.Fatalf("dependency freshness report missing %q", required)
		}
	}
}

func TestHelmHardeningChecksCoverUnitTestsAndDependencyFreshness(t *testing.T) {
	workflow := mustRead(t, "../../.github/workflows/ci.yaml")
	script := mustRead(t, "../../test/e2e/helm-hardening-checks.sh")
	combined := workflow + "\n" + script
	for _, required := range []string{"helm unittest", "helm dependency list", "helm dependency build", "helm lint", "helm template", "missing helm unittest coverage", "helm-hardening-checks.sh --dry-run"} {
		if !strings.Contains(combined, required) {
			t.Fatalf("helm hardening checks missing %q", required)
		}
	}
}

func TestNetworkPolicyBehaviorScriptProvesDenyAndAllowTraffic(t *testing.T) {
	script := mustRead(t, "../../test/e2e/network-policy-behavior.sh")
	for _, required := range []string{"deny-client", "allow-client", "expected denied traffic to fail", "expected allowed traffic to succeed", "allow-netpol-client-to-echo", "network-policy-deny-allow-proof-passed", "--dry-run"} {
		if !strings.Contains(script, required) {
			t.Fatalf("network policy behavior script missing %q", required)
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
