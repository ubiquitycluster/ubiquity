package cloud

import (
	"strings"
	"testing"
)

func TestRenderCloudPrerequisitesListsCRDsOperatorsAndFailClosedGates(t *testing.T) {
	manifest, err := RenderCloudPrerequisites(CloudPrerequisitesRequest{Name: "cloud-prereqs", Namespace: "ubiquity-system"})
	if err != nil {
		t.Fatalf("RenderCloudPrerequisites returned error: %v", err)
	}
	for _, required := range []string{
		"kind: ConfigMap", "name: cloud-prereqs", "ubiquity.ai/prerequisite-contract: cloud",
		"datavolumes.cdi.kubevirt.io", "virtualmachines.kubevirt.io", "objectbucketclaims.objectbucket.io",
		"clusters.postgresql.cnpg.io", "kafkas.kafka.strimzi.io", "schedules.k8up.io",
		"serverSideDryRunRequired: \"true\"", "reconciliationRequired: \"true\"", "restoreDrillRequired: \"true\"",
	} {
		if !strings.Contains(manifest, required) {
			t.Fatalf("manifest missing %q:\n%s", required, manifest)
		}
	}
}

func TestRequiredCloudCRDsIncludesEveryCatalogService(t *testing.T) {
	crds := RequiredCloudCRDs()
	for _, required := range []string{
		"redisfailovers.databases.spotahome.com",
		"projects.goharbor.io",
		"clusters.cluster.x-k8s.io",
		"volumesnapshotclasses.snapshot.storage.k8s.io",
	} {
		if !containsString(crds, required) {
			t.Fatalf("RequiredCloudCRDs missing %s: %#v", required, crds)
		}
	}
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
