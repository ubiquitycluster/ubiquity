package cloud

import "testing"

func TestManagedServiceReadinessResourcesCoverEveryServiceType(t *testing.T) {
	cases := map[ServiceType]string{
		ServiceBucket:      "objectbucketclaims.objectbucket.io",
		ServicePostgres:    "clusters.postgresql.cnpg.io",
		ServiceRedis:       "redisfailovers.databases.spotahome.com",
		ServiceKafka:       "kafkas.kafka.strimzi.io",
		ServiceRegistry:    "projects.goharbor.io",
		ServiceMariaDB:     "mariadbs.k8s.mariadb.com",
		ServiceMongoDB:     "perconaservermongodbs.psmdb.percona.com",
		ServiceNATS:        "natsclusters.nats.io",
		ServiceRabbitMQ:    "rabbitmqclusters.rabbitmq.com",
		ServiceClickHouse:  "clickhouseinstallations.clickhouse.altinity.com",
		ServiceOpenSearch:  "opensearchclusters.opensearch.opster.io",
		ServiceQdrant:      "qdrantclusters.qdrant.io",
		ServiceOpenBao:     "vaultconnections.secrets.hashicorp.com",
		ServiceHTTPCache:   "httpproxies.projectcontour.io",
		ServiceTCPBalancer: "tcproutes.gateway.networking.k8s.io",
	}
	for serviceType, want := range cases {
		got := ManagedServiceReadinessResources(serviceType)
		if len(got) != 1 || got[0] != want {
			t.Fatalf("service %s readiness resources = %v, want %s", serviceType, got, want)
		}
	}
}

func TestManagedServiceReadinessContractsRequireSpecificSmokeMarkers(t *testing.T) {
	cases := map[ServiceType]string{
		ServiceBucket:      "cloud-bucket-smoke-passed",
		ServicePostgres:    "cnpg-postgres-smoke-passed",
		ServiceRedis:       "redis-smoke-passed",
		ServiceKafka:       "kafka-smoke-passed",
		ServiceRegistry:    "harbor-registry-smoke-passed",
		ServiceMariaDB:     "mariadb-smoke-passed",
		ServiceMongoDB:     "mongodb-smoke-passed",
		ServiceNATS:        "nats-smoke-passed",
		ServiceRabbitMQ:    "rabbitmq-smoke-passed",
		ServiceClickHouse:  "clickhouse-smoke-passed",
		ServiceOpenSearch:  "opensearch-smoke-passed",
		ServiceQdrant:      "qdrant-smoke-passed",
		ServiceOpenBao:     "openbao-vault-smoke-passed",
		ServiceHTTPCache:   "http-cache-smoke-passed",
		ServiceTCPBalancer: "tcp-balancer-smoke-passed",
	}
	for serviceType, wantSmoke := range cases {
		contract, ok := ManagedServiceReadinessContract(serviceType)
		if !ok {
			t.Fatalf("missing readiness contract for %s", serviceType)
		}
		if contract.SmokeTest != wantSmoke {
			t.Fatalf("service %s smoke marker = %q, want %q", serviceType, contract.SmokeTest, wantSmoke)
		}
		if len(contract.Resources) == 0 {
			t.Fatalf("service %s has no readiness resources", serviceType)
		}
	}
	all := AllManagedServiceSmokeTests()
	if len(all) != len(cases) {
		t.Fatalf("all smoke tests = %v, want %d entries", all, len(cases))
	}
}

func TestAllManagedServiceReadinessResourcesAreDeduplicated(t *testing.T) {
	resources := AllManagedServiceReadinessResources()
	seen := map[string]struct{}{}
	for _, resource := range resources {
		if _, ok := seen[resource]; ok {
			t.Fatalf("duplicate readiness resource %s in %v", resource, resources)
		}
		seen[resource] = struct{}{}
	}
	if len(resources) < 15 {
		t.Fatalf("expected expanded managed-service readiness resources, got %v", resources)
	}
}
