package cloud

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCoreServicesCapabilityContract(t *testing.T) {
	chart := mustRead(t, "../../system/core-services/Chart.yaml")
	values := mustRead(t, "../../system/core-services/values.yaml")
	templates := mustRead(t, "../../system/core-services/templates/applications.yaml")
	doc := mustRead(t, "../../docs/architecture/core-services.md")
	script := mustRead(t, "../../test/e2e/core-services-proof.sh")
	combined := chart + "\n" + values + "\n" + templates + "\n" + doc + "\n" + script

	requiredCapabilities := []string{
		"cert-manager",
		"cilium",
		"external-secrets",
		"longhorn",
		"network-policies",
		"kyverno",
		"kyverno-policies",
		"falco",
		"monitoring-system",
		"ingress-nginx",
		"metrics-server",
		"node-feature-discovery",
		"node-problem-detector",
		"snapshot-controller",
		"velero",
		"vertical-pod-autoscaler",
		"kubescape",
		"local-path-provisioner",
	}
	for _, capability := range requiredCapabilities {
		if !strings.Contains(combined, capability) {
			t.Fatalf("core services capability contract missing %q", capability)
		}
	}

	for _, forbidden := range []string{"flux", "Flux", "gitRepository", "HelmRelease"} {
		if strings.Contains(combined, forbidden) {
			t.Fatalf("core services must not include forbidden GitOps dependency %q", forbidden)
		}
	}

	for _, required := range []string{"repoURL", "targetRevision", "destination", "syncPolicy", "CreateNamespace=true", "velero.backupBucket is required", "core-services-proof-passed", "better than a monolithic chart"} {
		if !strings.Contains(combined, required) {
			t.Fatalf("core services contract missing %q", required)
		}
	}
}

func TestCoreServicesDoesNotMentionForbiddenVendorName(t *testing.T) {
	paths := []string{
		"../../docs/plans/2026-06-09-core-services-capability-parity.md",
		"../../docs/architecture/core-services.md",
		"../../system/core-services/Chart.yaml",
		"../../system/core-services/values.yaml",
		"../../system/core-services/templates/applications.yaml",
		"../../test/e2e/core-services-proof.sh",
	}
	for _, path := range paths {
		content := strings.ToLower(mustRead(t, filepath.Clean(path)))
		forbidden := "ns" + "cale"
		if strings.Contains(content, forbidden) {
			t.Fatalf("%s contains forbidden vendor name", path)
		}
	}
}
