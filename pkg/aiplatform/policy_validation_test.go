package aiplatform

import (
	"strings"
	"testing"
)

func TestCIExercisesKyvernoAndNetworkPolicyBehavioralTests(t *testing.T) {
	workflow := mustRead(t, "../../.github/workflows/ci.yaml")
	for _, required := range []string{
		"kyverno/action-install-cli",
		"kyverno test system/kyverno-policies/kyverno-test",
		"test/e2e/network-policy-behavior.sh --dry-run",
	} {
		if !strings.Contains(workflow, required) {
			t.Fatalf("CI workflow must include policy validation marker %q", required)
		}
	}
}

func TestNetworkPolicyBehaviorScriptDocumentsDenyAndDNSProof(t *testing.T) {
	script := mustRead(t, "../../test/e2e/network-policy-behavior.sh")
	for _, required := range []string{
		"default-deny-ingress",
		"default-deny-egress",
		"allow-dns",
		"kubectl exec",
		"nslookup kubernetes.default.svc.cluster.local",
		"wget --timeout=3",
		"UBIQUITY_RUN_NETWORK_POLICY_E2E=true",
		"--dry-run",
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("network policy behavior script must include %q", required)
		}
	}
}
