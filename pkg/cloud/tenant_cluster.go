package cloud

import (
	"fmt"
	"regexp"
	"strings"
)

var kubeVersion = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)

// TenantClusterRequest describes a tenant Kubernetes workload cluster without replacing NICo host lifecycle.
type TenantClusterRequest struct {
	Name              string
	Namespace         string
	KubernetesVersion string
	ControlPlaneClass string
	NodePoolClass     string
	WorkerReplicas    int
}

// RenderTenantKubernetesCluster renders CAPI/Kamaji-style workload cluster intent tied to NICo-managed workers.
func RenderTenantKubernetesCluster(req TenantClusterRequest) (string, error) {
	req = defaultTenantCluster(req)
	if err := validateTenantCluster(req); err != nil {
		return "", err
	}
	return fmt.Sprintf(`apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: %s
  namespace: %s
  labels:
    app.kubernetes.io/part-of: ubiquity-tenant-kubernetes
    ubiquity.ai/node-lifecycle: nico-primary
  annotations:
    ubiquity.ai/design-boundary: "Tenant clusters consume NICo-managed/bootstrap nodes; they do not replace physical node lifecycle management."
spec:
  clusterNetwork:
    pods:
      cidrBlocks: ["10.244.0.0/16"]
    services:
      cidrBlocks: ["10.96.0.0/12"]
  controlPlaneRef:
    apiVersion: controlplane.cluster.x-k8s.io/v1alpha1
    kind: TenantControlPlane
    name: %s-control-plane
---
apiVersion: controlplane.cluster.x-k8s.io/v1alpha1
kind: TenantControlPlane
metadata:
  name: %s-control-plane
  namespace: %s
spec:
  controlPlaneClass: %s
  version: %s
---
apiVersion: cluster.x-k8s.io/v1beta1
kind: MachineDeployment
metadata:
  name: %s-workers
  namespace: %s
  labels:
    cluster.x-k8s.io/cluster-name: %s
    ubiquity.ai/node-pool-class: %s
spec:
  clusterName: %s
  replicas: %d
  selector:
    matchLabels:
      cluster.x-k8s.io/cluster-name: %s
  template:
    metadata:
      labels:
        cluster.x-k8s.io/cluster-name: %s
        ubiquity.ai/node-pool-class: %s
    spec:
      clusterName: %s
      version: %s
`, req.Name, req.Namespace, req.Name, req.Name, req.Namespace, req.ControlPlaneClass, req.KubernetesVersion, req.Name, req.Namespace, req.Name, req.NodePoolClass, req.Name, req.WorkerReplicas, req.Name, req.Name, req.NodePoolClass, req.Name, req.KubernetesVersion), nil
}

func defaultTenantCluster(req TenantClusterRequest) TenantClusterRequest {
	if req.Namespace == "" {
		req.Namespace = "tenant-clusters"
	}
	if req.KubernetesVersion == "" {
		req.KubernetesVersion = "v1.31.4"
	}
	if req.ControlPlaneClass == "" {
		req.ControlPlaneClass = "kamaji"
	}
	if req.NodePoolClass == "" {
		req.NodePoolClass = "nico-managed-workers"
	}
	if req.WorkerReplicas == 0 {
		req.WorkerReplicas = 3
	}
	return req
}

func validateTenantCluster(req TenantClusterRequest) error {
	if !kubeName.MatchString(req.Name) {
		return fmt.Errorf("tenant cluster name %q must be DNS-compatible", req.Name)
	}
	if !kubeName.MatchString(req.Namespace) {
		return fmt.Errorf("tenant cluster namespace %q must be DNS-compatible", req.Namespace)
	}
	if !kubeVersion.MatchString(req.KubernetesVersion) {
		return fmt.Errorf("tenant cluster version %q must look like v1.31.4", req.KubernetesVersion)
	}
	if strings.TrimSpace(req.ControlPlaneClass) == "" || strings.TrimSpace(req.NodePoolClass) == "" {
		return fmt.Errorf("tenant cluster control-plane and node-pool classes are required")
	}
	if req.WorkerReplicas < 0 {
		return fmt.Errorf("tenant cluster worker replicas must be non-negative")
	}
	return nil
}
