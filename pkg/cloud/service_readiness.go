package cloud

// ManagedServiceReadinessContract names the controller resource and live smoke marker required
// before a managed cloud primitive may be reported ready.
type ServiceReadinessContract struct {
	Service   ServiceType `json:"service" yaml:"service"`
	Resources []string    `json:"resources" yaml:"resources"`
	SmokeTest string      `json:"smokeTest" yaml:"smokeTest"`
}

var managedServiceReadinessContracts = map[ServiceType]ServiceReadinessContract{
	ServiceBucket:      {Service: ServiceBucket, Resources: []string{"objectbucketclaims.objectbucket.io"}, SmokeTest: "cloud-bucket-smoke-passed"},
	ServicePostgres:    {Service: ServicePostgres, Resources: []string{"clusters.postgresql.cnpg.io"}, SmokeTest: "cnpg-postgres-smoke-passed"},
	ServiceRedis:       {Service: ServiceRedis, Resources: []string{"redisfailovers.databases.spotahome.com"}, SmokeTest: "redis-smoke-passed"},
	ServiceKafka:       {Service: ServiceKafka, Resources: []string{"kafkas.kafka.strimzi.io"}, SmokeTest: "kafka-smoke-passed"},
	ServiceRegistry:    {Service: ServiceRegistry, Resources: []string{"projects.goharbor.io"}, SmokeTest: "harbor-registry-smoke-passed"},
	ServiceMariaDB:     {Service: ServiceMariaDB, Resources: []string{"mariadbs.k8s.mariadb.com"}, SmokeTest: "mariadb-smoke-passed"},
	ServiceMongoDB:     {Service: ServiceMongoDB, Resources: []string{"perconaservermongodbs.psmdb.percona.com"}, SmokeTest: "mongodb-smoke-passed"},
	ServiceNATS:        {Service: ServiceNATS, Resources: []string{"natsclusters.nats.io"}, SmokeTest: "nats-smoke-passed"},
	ServiceRabbitMQ:    {Service: ServiceRabbitMQ, Resources: []string{"rabbitmqclusters.rabbitmq.com"}, SmokeTest: "rabbitmq-smoke-passed"},
	ServiceClickHouse:  {Service: ServiceClickHouse, Resources: []string{"clickhouseinstallations.clickhouse.altinity.com"}, SmokeTest: "clickhouse-smoke-passed"},
	ServiceOpenSearch:  {Service: ServiceOpenSearch, Resources: []string{"opensearchclusters.opensearch.opster.io"}, SmokeTest: "opensearch-smoke-passed"},
	ServiceQdrant:      {Service: ServiceQdrant, Resources: []string{"qdrantclusters.qdrant.io"}, SmokeTest: "qdrant-smoke-passed"},
	ServiceOpenBao:     {Service: ServiceOpenBao, Resources: []string{"vaultconnections.secrets.hashicorp.com"}, SmokeTest: "openbao-vault-smoke-passed"},
	ServiceHTTPCache:   {Service: ServiceHTTPCache, Resources: []string{"httpproxies.projectcontour.io"}, SmokeTest: "http-cache-smoke-passed"},
	ServiceTCPBalancer: {Service: ServiceTCPBalancer, Resources: []string{"tcproutes.gateway.networking.k8s.io"}, SmokeTest: "tcp-balancer-smoke-passed"},
}

// ManagedServiceReadinessContract returns the service-specific readiness contract.
func ManagedServiceReadinessContract(serviceType ServiceType) (ServiceReadinessContract, bool) {
	contract, ok := managedServiceReadinessContracts[serviceType]
	return contract, ok
}

// ManagedServiceReadinessResources returns resource APIs whose status conditions prove the requested service type.
func ManagedServiceReadinessResources(serviceType ServiceType) []string {
	contract, ok := ManagedServiceReadinessContract(serviceType)
	if !ok {
		return nil
	}
	return append([]string{}, contract.Resources...)
}

// AllManagedServiceReadinessResources returns a deduplicated stable list for collector defaults.
func AllManagedServiceReadinessResources() []string {
	seen := map[string]struct{}{}
	var resources []string
	for _, service := range allManagedServiceTypes() {
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

// AllManagedServiceSmokeTests returns the stable list of smoke markers required by managed services.
func AllManagedServiceSmokeTests() []string {
	var smokeTests []string
	for _, service := range allManagedServiceTypes() {
		contract, ok := ManagedServiceReadinessContract(service)
		if ok && contract.SmokeTest != "" {
			smokeTests = append(smokeTests, contract.SmokeTest)
		}
	}
	return smokeTests
}

func allManagedServiceTypes() []ServiceType {
	return []ServiceType{
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
}
