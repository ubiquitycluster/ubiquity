package cloud

import (
	"strings"
	"testing"
)

func TestRenderNetBirdMultiClusterOverlayProducesPlaceholderSafeBundle(t *testing.T) {
	manifest, err := RenderNetBirdMultiClusterOverlay(NetBirdMultiClusterOverlayRequest{
		ManagementCluster: "ubiquity-management",
		RegionalCluster:   "spanish-fork-gpu-01",
		Region:            "us-west",
		Site:              "spanish-fork",
		StorageProvider:   "vast",
		GPUClass:          "h100",
		NetBirdServer:     "https://NETBIRD_OVERLAY_IP_OR_DNS:6443",
	})
	if err != nil {
		t.Fatalf("RenderNetBirdMultiClusterOverlay returned error: %v", err)
	}
	for _, required := range []string{
		"kind: ConfigMap",
		"name: ubiquity-netbird-overlay-policy",
		"NetBird private control/data overlay",
		"do not stretch one Kubernetes cluster across regions",
		"public inference traffic must not hairpin through NetBird",
		"kind: Secret",
		"argocd.argoproj.io/secret-type: cluster",
		"spanish-fork-gpu-01",
		"ubiquity.io/region: us-west",
		"ubiquity.io/site: spanish-fork",
		"ubiquity.io/gpu: \"true\"",
		"ubiquity.io/rdma: \"true\"",
		"ubiquity.io/inference: \"true\"",
		"ubiquity.io/storage: vast",
		"PLACEHOLDER_BEARER_TOKEN_FROM_REMOTE_CLUSTER_SERVICE_ACCOUNT",
		"PLACEHOLDER_BASE64_CLUSTER_CA_FROM_REMOTE_CLUSTER",
		"kind: ApplicationSet",
		"ubiquity-regional-ai-platform",
		"ubiquity-rdma-readiness-smoke",
		"nvidia-nic-configuration-operator",
		"rdma-network-smoke-test-passed",
		"nico-policy-reconciled",
	} {
		if !strings.Contains(manifest, required) {
			t.Fatalf("manifest missing %q:\n%s", required, manifest)
		}
	}
	for _, forbidden := range []string{"nb_" + "pat_", "setup" + "key:", "BEGIN " + "PRIVATE KEY", "eyJ" + "hbGci", "latest"} {
		if strings.Contains(manifest, forbidden) {
			t.Fatalf("manifest contains forbidden secret or unpinned marker %q:\n%s", forbidden, manifest)
		}
	}
}

func TestRenderNetBirdMultiClusterOverlayFailsClosedForMissingRequiredFields(t *testing.T) {
	for _, req := range []NetBirdMultiClusterOverlayRequest{
		{RegionalCluster: "spanish-fork-gpu-01", Region: "us-west", Site: "spanish-fork"},
		{ManagementCluster: "ubiquity-management", Region: "us-west", Site: "spanish-fork"},
		{ManagementCluster: "ubiquity-management", RegionalCluster: "spanish-fork-gpu-01", Site: "spanish-fork"},
		{ManagementCluster: "ubiquity-management", RegionalCluster: "spanish-fork-gpu-01", Region: "us-west"},
	} {
		if _, err := RenderNetBirdMultiClusterOverlay(req); err == nil {
			t.Fatalf("expected validation error for request %#v", req)
		}
	}
}

func TestRenderNetBirdMultiClusterOverlayRejectsPublicHairpinRouting(t *testing.T) {
	_, err := RenderNetBirdMultiClusterOverlay(NetBirdMultiClusterOverlayRequest{
		ManagementCluster:           "ubiquity-management",
		RegionalCluster:             "spanish-fork-gpu-01",
		Region:                      "us-west",
		Site:                        "spanish-fork",
		PublicInferenceThroughMesh:  true,
		AllowStretchedKubernetesWAN: false,
	})
	if err == nil || !strings.Contains(err.Error(), "public inference traffic") {
		t.Fatalf("expected public inference hairpin validation error, got %v", err)
	}

	_, err = RenderNetBirdMultiClusterOverlay(NetBirdMultiClusterOverlayRequest{
		ManagementCluster:           "ubiquity-management",
		RegionalCluster:             "spanish-fork-gpu-01",
		Region:                      "us-west",
		Site:                        "spanish-fork",
		AllowStretchedKubernetesWAN: true,
	})
	if err == nil || !strings.Contains(err.Error(), "stretch") {
		t.Fatalf("expected stretched Kubernetes validation error, got %v", err)
	}
}
