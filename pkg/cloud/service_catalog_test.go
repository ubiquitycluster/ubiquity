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
		{ServiceMariaDB, []string{"apiVersion: k8s.mariadb.com/v1alpha1", "kind: MariaDB", "rootPasswordSecretKeyRef"}},
		{ServiceMongoDB, []string{"apiVersion: psmdb.percona.com/v1", "kind: PerconaServerMongoDB", "replsets:"}},
		{ServiceNATS, []string{"apiVersion: nats.io/v1alpha2", "kind: NatsCluster", "size: 3"}},
		{ServiceRabbitMQ, []string{"apiVersion: rabbitmq.com/v1beta1", "kind: RabbitmqCluster", "replicas: 3"}},
		{ServiceClickHouse, []string{"apiVersion: clickhouse.altinity.com/v1", "kind: ClickHouseInstallation", "templates:"}},
		{ServiceOpenSearch, []string{"apiVersion: opensearch.opster.io/v1", "kind: OpenSearchCluster", "dashboards:"}},
		{ServiceQdrant, []string{"apiVersion: qdrant.io/v1", "kind: QdrantCluster", "persistence:"}},
		{ServiceOpenBao, []string{"apiVersion: secrets.hashicorp.com/v1beta1", "kind: VaultConnection", "ubiquity.ai/service-type: openbao"}},
		{ServiceHTTPCache, []string{"kind: Deployment", "ubiquity.ai/service-type: http-cache", "nginx:stable"}},
		{ServiceTCPBalancer, []string{"kind: Service", "ubiquity.ai/service-type: tcp-balancer", "type: LoadBalancer"}},
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
