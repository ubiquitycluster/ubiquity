package cloud

import (
	"strings"
	"testing"
)

func TestRequiredCloudOperatorBundlesIncludeOwnershipAndAirgapMetadata(t *testing.T) {
	bundles := RequiredCloudOperatorBundles()
	if len(bundles) < 10 {
		t.Fatalf("expected broad operator bundle coverage, got %d", len(bundles))
	}
	seen := map[string]CloudOperatorBundle{}
	for _, bundle := range bundles {
		seen[bundle.Name] = bundle
		if bundle.OwnedCRD == "" || bundle.Controller == "" || bundle.InstallNamespace == "" || bundle.Source == "" || bundle.AirgapArtifact == "" {
			t.Fatalf("bundle missing provenance/ownership fields: %#v", bundle)
		}
	}
	for _, required := range []string{"kubevirt-cdi", "cloudnative-pg", "strimzi", "cluster-api", "k8up", "longhorn", "gateway-api", "opencost", "kyverno", "argocd"} {
		if _, ok := seen[required]; !ok {
			t.Fatalf("missing operator bundle %q in %#v", required, seen)
		}
	}
}

func TestRenderCloudOperatorBundlesCreatesInstallPlanConfigMap(t *testing.T) {
	manifest, err := RenderCloudOperatorBundles(CloudOperatorBundlesRequest{Name: "cloud-operators", Namespace: "ubiquity-system"})
	if err != nil {
		t.Fatalf("RenderCloudOperatorBundles returned error: %v", err)
	}
	for _, required := range []string{"kind: ConfigMap", "ubiquity.ai/operator-install-plan: cloud", "kubevirt-cdi", "datavolumes.cdi.kubevirt.io", "airgapArtifact", "controller", "installNamespace"} {
		if !strings.Contains(manifest, required) {
			t.Fatalf("manifest missing %q:\n%s", required, manifest)
		}
	}
}

func TestRenderCloudOperatorBundlesRejectsBadNames(t *testing.T) {
	_, err := RenderCloudOperatorBundles(CloudOperatorBundlesRequest{Name: "Bad_Name", Namespace: "ubiquity-system"})
	if err == nil || !strings.Contains(err.Error(), "DNS-compatible") {
		t.Fatalf("expected DNS validation error, got %v", err)
	}
}

func TestRequiredCloudOperatorBundlesIncludeRestoreOwnership(t *testing.T) {
	manifest, err := RenderCloudOperatorBundles(CloudOperatorBundlesRequest{Name: "cloud-operators", Namespace: "ubiquity-system"})
	if err != nil {
		t.Fatalf("RenderCloudOperatorBundles returned error: %v", err)
	}
	for _, required := range []string{"restores.k8up.io", "K8up Operator"} {
		if !strings.Contains(manifest, required) {
			t.Fatalf("operator bundle manifest missing %q:\n%s", required, manifest)
		}
	}
}
