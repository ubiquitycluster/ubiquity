package cloud

import (
	"fmt"
	"strings"
)

// ServiceType identifies an operator-backed service catalog item.
type ServiceType string

const (
	ServiceBucket      ServiceType = "bucket"
	ServicePostgres    ServiceType = "postgres"
	ServiceRedis       ServiceType = "redis"
	ServiceKafka       ServiceType = "kafka"
	ServiceRegistry    ServiceType = "registry"
	ServiceMariaDB     ServiceType = "mariadb"
	ServiceMongoDB     ServiceType = "mongodb"
	ServiceNATS        ServiceType = "nats"
	ServiceRabbitMQ    ServiceType = "rabbitmq"
	ServiceClickHouse  ServiceType = "clickhouse"
	ServiceOpenSearch  ServiceType = "opensearch"
	ServiceQdrant      ServiceType = "qdrant"
	ServiceOpenBao     ServiceType = "openbao"
	ServiceHTTPCache   ServiceType = "http-cache"
	ServiceTCPBalancer ServiceType = "tcp-balancer"
)

// ManagedServiceRequest renders a service CR for an operator already installed by the platform.
type ManagedServiceRequest struct {
	Name         string
	Namespace    string
	Type         ServiceType
	StorageClass string
	Size         string
	Replicas     int
}

// RenderManagedService renders a supported service catalog primitive without installing/replacing its operator.
func RenderManagedService(req ManagedServiceRequest) (string, error) {
	req = defaultManagedService(req)
	if err := validateManagedService(req); err != nil {
		return "", err
	}
	switch req.Type {
	case ServiceBucket:
		return fmt.Sprintf(`apiVersion: objectbucket.io/v1alpha1
kind: ObjectBucketClaim
metadata:
  name: %s
  namespace: %s
  labels:
    ubiquity.ai/service-type: bucket
spec:
  bucketName: %s
  storageClassName: %s
`, req.Name, req.Namespace, req.Name, req.StorageClass), nil
	case ServicePostgres:
		return fmt.Sprintf(`apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: %s
  namespace: %s
  labels:
    ubiquity.ai/service-type: postgres
spec:
  instances: %d
  storage:
    size: %s
    storageClass: %s
`, req.Name, req.Namespace, req.Replicas, req.Size, req.StorageClass), nil
	case ServiceRedis:
		return fmt.Sprintf(`apiVersion: databases.spotahome.com/v1
kind: RedisFailover
metadata:
  name: %s
  namespace: %s
  labels:
    ubiquity.ai/service-type: redis
spec:
  sentinel:
    replicas: 3
  redis:
    replicas: %d
    storage:
      persistentVolumeClaim:
        metadata:
          name: %s-data
        spec:
          storageClassName: %s
          resources:
            requests:
              storage: %s
`, req.Name, req.Namespace, req.Replicas, req.Name, req.StorageClass, req.Size), nil
	case ServiceKafka:
		return fmt.Sprintf(`apiVersion: kafka.strimzi.io/v1beta2
kind: Kafka
metadata:
  name: %s
  namespace: %s
  labels:
    ubiquity.ai/service-type: kafka
spec:
  kafka:
    replicas: %d
    storage:
      type: persistent-claim
      class: %s
      size: %s
  zookeeper:
    replicas: 3
    storage:
      type: persistent-claim
      class: %s
      size: %s
`, req.Name, req.Namespace, req.Replicas, req.StorageClass, req.Size, req.StorageClass, req.Size), nil
	case ServiceRegistry:
		return fmt.Sprintf(`apiVersion: goharbor.io/v1alpha1
kind: Project
metadata:
  name: %s
  namespace: %s
  labels:
    ubiquity.ai/service-type: registry
spec:
  public: false
  storageLimit: %s
`, req.Name, req.Namespace, req.Size), nil
	case ServiceMariaDB:
		return fmt.Sprintf(`apiVersion: k8s.mariadb.com/v1alpha1
kind: MariaDB
metadata:
  name: %s
  namespace: %s
  labels:
    ubiquity.ai/service-type: mariadb
spec:
  replicas: %d
  storage:
    size: %s
    storageClassName: %s
  rootPasswordSecretKeyRef:
    name: %s-root
    key: password
`, req.Name, req.Namespace, req.Replicas, req.Size, req.StorageClass, req.Name), nil
	case ServiceMongoDB:
		return fmt.Sprintf(`apiVersion: psmdb.percona.com/v1
kind: PerconaServerMongoDB
metadata:
  name: %s
  namespace: %s
  labels:
    ubiquity.ai/service-type: mongodb
spec:
  replsets:
    - name: rs0
      size: %d
      volumeSpec:
        persistentVolumeClaim:
          storageClassName: %s
          resources:
            requests:
              storage: %s
`, req.Name, req.Namespace, req.Replicas, req.StorageClass, req.Size), nil
	case ServiceNATS:
		return fmt.Sprintf(`apiVersion: nats.io/v1alpha2
kind: NatsCluster
metadata:
  name: %s
  namespace: %s
  labels:
    ubiquity.ai/service-type: nats
spec:
  size: %d
`, req.Name, req.Namespace, req.Replicas), nil
	case ServiceRabbitMQ:
		return fmt.Sprintf(`apiVersion: rabbitmq.com/v1beta1
kind: RabbitmqCluster
metadata:
  name: %s
  namespace: %s
  labels:
    ubiquity.ai/service-type: rabbitmq
spec:
  replicas: %d
  persistence:
    storageClassName: %s
    storage: %s
`, req.Name, req.Namespace, req.Replicas, req.StorageClass, req.Size), nil
	case ServiceClickHouse:
		return fmt.Sprintf(`apiVersion: clickhouse.altinity.com/v1
kind: ClickHouseInstallation
metadata:
  name: %s
  namespace: %s
  labels:
    ubiquity.ai/service-type: clickhouse
spec:
  configuration:
    clusters:
      - name: default
        layout:
          shardsCount: 1
          replicasCount: %d
  templates:
    volumeClaimTemplates:
      - name: data
        spec:
          storageClassName: %s
          resources:
            requests:
              storage: %s
`, req.Name, req.Namespace, req.Replicas, req.StorageClass, req.Size), nil
	case ServiceOpenSearch:
		return fmt.Sprintf(`apiVersion: opensearch.opster.io/v1
kind: OpenSearchCluster
metadata:
  name: %s
  namespace: %s
  labels:
    ubiquity.ai/service-type: opensearch
spec:
  general:
    serviceName: %s
  dashboards:
    enable: true
  nodePools:
    - component: masters
      replicas: %d
      diskSize: %s
`, req.Name, req.Namespace, req.Name, req.Replicas, req.Size), nil
	case ServiceQdrant:
		return fmt.Sprintf(`apiVersion: qdrant.io/v1
kind: QdrantCluster
metadata:
  name: %s
  namespace: %s
  labels:
    ubiquity.ai/service-type: qdrant
spec:
  replicas: %d
  persistence:
    storageClassName: %s
    size: %s
`, req.Name, req.Namespace, req.Replicas, req.StorageClass, req.Size), nil
	case ServiceOpenBao:
		return fmt.Sprintf(`apiVersion: secrets.hashicorp.com/v1beta1
kind: VaultConnection
metadata:
  name: %s
  namespace: %s
  labels:
    ubiquity.ai/service-type: openbao
spec:
  address: https://%s.%s.svc:8200
`, req.Name, req.Namespace, req.Name, req.Namespace), nil
	case ServiceHTTPCache:
		return fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  namespace: %s
  labels:
    ubiquity.ai/service-type: http-cache
spec:
  replicas: %d
  selector:
    matchLabels:
      app.kubernetes.io/name: %s
  template:
    metadata:
      labels:
        app.kubernetes.io/name: %s
    spec:
      containers:
        - name: cache
          image: nginx:stable
`, req.Name, req.Namespace, req.Replicas, req.Name, req.Name), nil
	case ServiceTCPBalancer:
		return fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: %s
  namespace: %s
  labels:
    ubiquity.ai/service-type: tcp-balancer
spec:
  type: LoadBalancer
  ports:
    - name: tcp
      port: 443
      targetPort: 443
  selector:
    ubiquity.ai/tcp-backend: %s
`, req.Name, req.Namespace, req.Name), nil
	default:
		return "", fmt.Errorf("unsupported managed service %q", req.Type)
	}
}

func defaultManagedService(req ManagedServiceRequest) ManagedServiceRequest {
	if req.Namespace == "" {
		req.Namespace = "tenant-a"
	}
	if req.Type == "" {
		req.Type = ServiceBucket
	}
	if req.StorageClass == "" {
		req.StorageClass = "standard"
	}
	if req.Size == "" {
		req.Size = "100Gi"
	}
	if req.Replicas == 0 {
		req.Replicas = 3
	}
	return req
}

func validateManagedService(req ManagedServiceRequest) error {
	if !kubeName.MatchString(req.Name) {
		return fmt.Errorf("managed service name %q must be DNS-compatible", req.Name)
	}
	if !kubeName.MatchString(req.Namespace) {
		return fmt.Errorf("managed service namespace %q must be DNS-compatible", req.Namespace)
	}
	if strings.TrimSpace(req.StorageClass) == "" || strings.TrimSpace(req.Size) == "" {
		return fmt.Errorf("managed service storage class and size are required")
	}
	if req.Replicas < 1 {
		return fmt.Errorf("managed service replicas must be at least 1")
	}
	return nil
}
