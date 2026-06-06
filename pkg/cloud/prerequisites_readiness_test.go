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
