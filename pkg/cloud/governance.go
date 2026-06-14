package cloud

import "fmt"

// CloudGovernanceRequest renders cross-cutting governance for tenant cloud primitives.
type CloudGovernanceRequest struct {
	Name      string
	Namespace string
}

// RenderCloudGovernance renders security, GitOps, observability, cost, networking, storage, and upgrade policies.
func RenderCloudGovernance(req CloudGovernanceRequest) (string, error) {
	if req.Name == "" {
		req.Name = "tenant-a-governance"
	}
	if req.Namespace == "" {
		req.Namespace = "tenant-a"
	}
	if !kubeName.MatchString(req.Name) || !kubeName.MatchString(req.Namespace) {
		return "", fmt.Errorf("cloud governance name and namespace must be DNS-compatible")
	}
	return fmt.Sprintf(`apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: %s-cloud-admin
  namespace: %s
rules:
  - apiGroups: ["", "networking.k8s.io", "kubevirt.io", "cdi.kubevirt.io"]
    resources: ["services", "configmaps", "networkpolicies", "virtualmachines", "datavolumes"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: %s-cloud-admin
  namespace: %s
subjects:
  - kind: Group
    name: ubiquity:%s:cloud-admins
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: %s-cloud-admin
---
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: require-cloud-provenance
spec:
  validationFailureAction: Enforce
  rules:
    - name: require-cloud-provenance
      match:
        any:
          - resources:
              kinds: ["ConfigMap", "Service", "Deployment"]
      validate:
        message: "cloud primitives require ubiquity.ai/provenance annotations"
        pattern:
          metadata:
            annotations:
              ubiquity.ai/provenance: "?*"
---
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: %s-gitops-lifecycle
  namespace: argocd
  labels:
    ubiquity.ai/gitops-lifecycle: tenant-cloud
spec:
  project: default
  source:
    repoURL: https://example.invalid/ubiquity-gitops.git
    targetRevision: HEAD
    path: tenants/%s/cloud
  destination:
    server: https://kubernetes.default.svc
    namespace: %s
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: %s-cloud
  namespace: %s
spec:
  selector:
    matchLabels:
      ubiquity.ai/cloud-observed: "true"
  endpoints:
    - port: metrics
---
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: %s-cloud
  namespace: %s
spec:
  groups:
    - name: cloud-primitive-degraded
      rules:
        - alert: CloudPrimitiveDegraded
          expr: ubiquity_cloud_primitive_ready == 0
          for: 10m
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: %s-cost-allocation
  namespace: %s
  labels:
    opencost.io/allocation: tenant
data:
  tenant: %s
  gpuCostAllocation: "required"
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: %s-gateway
  namespace: %s
spec:
  gatewayClassName: ubiquity-tenant
  listeners:
    - name: https
      protocol: HTTPS
      port: 443
---
apiVersion: externaldns.k8s.io/v1alpha1
kind: DNSEndpoint
metadata:
  name: %s-dns
  namespace: %s
spec:
  endpoints: []
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: %s-vpn-egress
  namespace: %s
  labels:
    ubiquity.ai/network-boundary: vpn
spec:
  podSelector: {}
  policyTypes: ["Egress"]
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: %s-expandable
  annotations:
    ubiquity.ai/storage-profile: tenant-cloud
provisioner: driver.longhorn.io
allowVolumeExpansion: true
reclaimPolicy: Retain
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: %s-upgrade-policy
  namespace: %s
data:
  upgradePolicy: "staged-server-dry-run-then-reconcile"
  rollbackPolicy: "retain-previous-rendered-manifest-and-controller-version"
`, req.Name, req.Namespace, req.Name, req.Namespace, req.Namespace, req.Name, req.Name, req.Namespace, req.Namespace, req.Name, req.Namespace, req.Name, req.Namespace, req.Name, req.Namespace, req.Namespace, req.Name, req.Namespace, req.Name, req.Namespace, req.Name, req.Namespace, req.Name, req.Name, req.Namespace), nil
}
