package cloud

// ManagedServiceReadinessResources returns resource APIs whose status conditions prove the requested service type.
func ManagedServiceReadinessResources(serviceType ServiceType) []string {
	resource, ok := map[ServiceType]string{
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
	}[serviceType]
	if !ok {
		return nil
	}
	return []string{resource}
}

// AllManagedServiceReadinessResources returns a deduplicated stable list for collector defaults.
func AllManagedServiceReadinessResources() []string {
	services := []ServiceType{
		ServiceBucket,
		ServicePostgres,
		ServiceRedis,
		ServiceKafka,
		ServiceRegistry,
		ServiceMariaDB,
		ServiceMongoDB,
		ServiceNATS,
		ServiceRabbitMQ,
		ServiceClickHouse,
		ServiceOpenSearch,
		ServiceQdrant,
		ServiceOpenBao,
		ServiceHTTPCache,
		ServiceTCPBalancer,
	}
	seen := map[string]struct{}{}
	var resources []string
	for _, service := range services {
		for _, resource := range ManagedServiceReadinessResources(service) {
			if _, ok := seen[resource]; ok {
				continue
			}
			seen[resource] = struct{}{}
			resources = append(resources, resource)
		}
	}
	return resources
}
