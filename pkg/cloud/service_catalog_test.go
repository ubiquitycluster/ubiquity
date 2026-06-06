package cloud

import (
	"strings"
	"testing"
)

func TestRenderManagedServiceCatalogSupportsOperatorBackedServices(t *testing.T) {
	cases := []struct {
		service  ServiceType
		required []string
	}{
		{ServiceBucket, []string{"kind: ObjectBucketClaim", "storageClassName: object-store", "bucketName: datasets"}},
		{ServicePostgres, []string{"apiVersion: postgresql.cnpg.io/v1", "kind: Cluster", "instances: 3", "storage:", "size: 200Gi"}},
		{ServiceRedis, []string{"kind: RedisFailover", "sentinel:", "replicas: 3"}},
		{ServiceKafka, []string{"apiVersion: kafka.strimzi.io/v1beta2", "kind: Kafka", "replicas: 3"}},
		{ServiceRegistry, []string{"kind: Project", "ubiquity.ai/service-type: registry"}},
	}
	for _, tc := range cases {
		t.Run(string(tc.service), func(t *testing.T) {
			manifest, err := RenderManagedService(ManagedServiceRequest{Name: "datasets", Namespace: "tenant-a", Type: tc.service, StorageClass: "object-store", Size: "200Gi", Replicas: 3})
			if err != nil {
				t.Fatalf("RenderManagedService returned error: %v", err)
			}
			for _, required := range tc.required {
				if !strings.Contains(manifest, required) {
					t.Fatalf("manifest missing %q:\n%s", required, manifest)
				}
			}
		})
	}
}

func TestRenderManagedServiceFailsClosedForUnsupportedService(t *testing.T) {
	_, err := RenderManagedService(ManagedServiceRequest{Name: "bad", Type: ServiceType("oracle")})
	if err == nil || !strings.Contains(err.Error(), "unsupported managed service") {
		t.Fatalf("expected unsupported service error, got %v", err)
	}
}
