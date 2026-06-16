package cloud

import (
	"encoding/json"
	"os"
	"path/filepath"
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

func TestDependencyAutomationConfigsCoverSupportedEcosystems(t *testing.T) {
	dependabot := mustRead(t, "../../.github/dependabot.yml")
	renovate := mustRead(t, "../../renovate.json")
	if _, err := os.Stat("../../renovate.json5"); !os.IsNotExist(err) {
		t.Fatalf("renovate.json5 should be removed after JSON migration, stat err=%v", err)
	}
	var renovateJSON map[string]any
	if err := json.Unmarshal([]byte(renovate), &renovateJSON); err != nil {
		t.Fatalf("renovate.json is not valid JSON: %v", err)
	}

	for path, content := range map[string]string{
		"dependabot": dependabot,
		"renovate":   renovate,
	} {
		for _, required := range []string{"gomod", "github-actions", "docker", "dependencies"} {
			if !strings.Contains(content, required) {
				t.Fatalf("%s config missing %q", path, required)
			}
		}
	}
	for _, required := range []string{"directory: \"/\"", "directory: \"/tools\"", "directory: \"/test\"", "directory: \"/terratest\"", "open-pull-requests-limit"} {
		if !strings.Contains(dependabot, required) {
			t.Fatalf("dependabot config missing %q", required)
		}
	}
	for _, required := range []string{"regexManagers", "Chart\\\\.ya?ml", "datasourceTemplate", "helm", "registryUrl", "Go modules", "GitHub Actions"} {
		if !strings.Contains(renovate, required) {
			t.Fatalf("renovate config missing %q", required)
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

func TestCISupplyChainAndLintGatesArePinnedAndFailClosed(t *testing.T) {
	ci := mustRead(t, "../../.github/workflows/ci.yaml")
	release := mustRead(t, "../../.github/workflows/release.yaml")
	opus := mustRead(t, "../../.github/workflows/opus-build.yml")
	vulncheck := mustRead(t, "../../.github/workflows/vulncheck.yaml")
	combined := ci + "\n" + release + "\n" + opus + "\n" + vulncheck

	for _, forbidden := range []string{"ansible-lint metal/ --exclude metal/roles/automatic_upgrade/ || true", "terraform fmt -check -recursive) || true", "@master", "@latest", "releases/latest", "version: latest", "git diff-tree --no-commit-id"} {
		if strings.Contains(combined, forbidden) {
			t.Fatalf("CI/release workflow retains mutable or fail-open pattern %q", forbidden)
		}
	}
	for _, required := range []string{"github.event.pull_request.base.sha", "git diff --name-only", "GITLEAKS_VERSION", "GOVULNCHECK_VERSION", "HELM_UNITTEST_VERSION", "KUBECONFORM_VERSION", "aquasecurity/trivy-action@v0.36.0", "anchore/sbom-action@v0.24.0", "GORELEASER_VERSION", "scripts/check-graphify-freshness.sh --strict"} {
		if !strings.Contains(combined, required) {
			t.Fatalf("CI/release workflow missing pinned/fail-closed pattern %q", required)
		}
	}
}

func TestGraphifyFreshnessCheckIsWiredIntoGithubCI(t *testing.T) {
	workflow := mustRead(t, filepath.Clean("../../.github/workflows/ci.yaml"))
	if !strings.Contains(workflow, "scripts/check-graphify-freshness.sh --strict") {
		t.Fatal("CI must run Graphify strict freshness checks")
	}
}
