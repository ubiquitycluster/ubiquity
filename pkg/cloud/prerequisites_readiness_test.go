package cloud

import "testing"

func TestRequiredCloudCRDsIncludeManagedServiceReadinessResources(t *testing.T) {
	required := map[string]struct{}{}
	for _, crd := range RequiredCloudCRDs() {
		required[crd] = struct{}{}
	}
	for _, resource := range AllManagedServiceReadinessResources() {
		if _, ok := required[resource]; !ok {
			t.Fatalf("RequiredCloudCRDs missing managed service resource %s", resource)
		}
	}
}

func TestRequiredCloudCRDsIncludeRestoreDrillResource(t *testing.T) {
	required := map[string]struct{}{}
	for _, crd := range RequiredCloudCRDs() {
		required[crd] = struct{}{}
	}
	if _, ok := required["restores.k8up.io"]; !ok {
		t.Fatalf("RequiredCloudCRDs missing restores.k8up.io: %v", RequiredCloudCRDs())
	}
}
