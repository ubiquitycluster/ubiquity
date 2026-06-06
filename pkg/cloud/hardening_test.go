package cloud

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCloudChartsHaveValuesSchemas(t *testing.T) {
	charts := []string{
		"kubevirt-vm-disks",
		"tenant-vpc",
		"tenant-kubernetes-cluster",
		"managed-service",
		"platform-ops-policy",
	}
	for _, chart := range charts {
		t.Run(chart, func(t *testing.T) {
			path := "../../platform/" + chart + "/values.schema.json"
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("missing schema %s: %v", path, err)
			}
			var schema map[string]any
			if err := json.Unmarshal(content, &schema); err != nil {
				t.Fatalf("schema %s is not valid JSON: %v", path, err)
			}
			if schema["$schema"] == "" || schema["type"] != "object" {
				t.Fatalf("schema %s missing JSON-schema header/type: %#v", path, schema)
			}
		})
	}
}

func TestCloudServicesDocumentationHasCRDProvenanceAndReadinessBoundaries(t *testing.T) {
	content, err := os.ReadFile("../../docs/admin-guide/cloud-services.md")
	if err != nil {
		t.Fatal(err)
	}
	doc := string(content)
	for _, required := range []string{
		"## CRD and operator provenance",
		"ObjectBucketClaim", "objectbucket.io/v1alpha1",
		"CloudNativePG", "postgresql.cnpg.io/v1",
		"Strimzi", "kafka.strimzi.io/v1beta2",
		"K8up", "k8up.io/v1",
		"## Production readiness gates",
		"server-side dry-run", "controller reconciliation", "restore drill",
	} {
		if !strings.Contains(doc, required) {
			t.Fatalf("cloud services doc missing %q", required)
		}
	}
}
