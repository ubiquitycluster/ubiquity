package cloud

import (
	"strings"
	"testing"
)

func TestRenderCloudGovernanceCoversSecurityGitOpsObservabilityNetworkingStorageAndUpgrades(t *testing.T) {
	manifest, err := RenderCloudGovernance(CloudGovernanceRequest{Name: "tenant-a-governance", Namespace: "tenant-a"})
	if err != nil {
		t.Fatalf("RenderCloudGovernance returned error: %v", err)
	}
	for _, required := range []string{
		"kind: Role", "kind: RoleBinding", "kind: ClusterPolicy", "require-cloud-provenance",
		"kind: Application", "argoproj.io/v1alpha1", "ubiquity.ai/gitops-lifecycle",
		"kind: ServiceMonitor", "kind: PrometheusRule", "cloud-primitive-degraded",
		"kind: ConfigMap", "name: tenant-a-governance-cost-allocation", "opencost.io/allocation",
		"kind: Gateway", "gateway.networking.k8s.io/v1", "kind: DNSEndpoint",
		"kind: NetworkPolicy", "vpn", "kind: StorageClass", "allowVolumeExpansion: true",
		"kind: ConfigMap", "upgradePolicy", "rollbackPolicy",
	} {
		if !strings.Contains(manifest, required) {
			t.Fatalf("manifest missing %q:\n%s", required, manifest)
		}
	}
}

func TestRenderCloudGovernanceFailsClosedForBadNames(t *testing.T) {
	_, err := RenderCloudGovernance(CloudGovernanceRequest{Name: "Bad_Name", Namespace: "tenant-a"})
	if err == nil || !strings.Contains(err.Error(), "DNS-compatible") {
		t.Fatalf("expected DNS validation error, got %v", err)
	}
}
