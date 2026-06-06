package cloud

import (
	"fmt"
	"strings"
)

// ServiceType identifies an operator-backed service catalog item.
type ServiceType string

const (
	ServiceBucket   ServiceType = "bucket"
	ServicePostgres ServiceType = "postgres"
	ServiceRedis    ServiceType = "redis"
	ServiceKafka    ServiceType = "kafka"
	ServiceRegistry ServiceType = "registry"
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
