package cloud

import (
	"fmt"
	"strings"
)

// NetBirdMultiClusterOverlayRequest describes a placeholder-safe NetBird fleet
// bundle for one management cluster and one regional Ubiquity workload cluster.
type NetBirdMultiClusterOverlayRequest struct {
	ManagementCluster           string
	RegionalCluster             string
	Namespace                   string
	Region                      string
	Site                        string
	TrustTier                   string
	StorageProvider             string
	GPUClass                    string
	NetBirdServer               string
	RepoURL                     string
	TargetRevision              string
	PublicInferenceThroughMesh  bool
	AllowStretchedKubernetesWAN bool
}

// RenderNetBirdMultiClusterOverlay renders the GitOps resources needed to add a
// regional Ubiquity cluster to a NetBird-backed multi-cluster fleet. The bundle
// intentionally contains placeholders for credentials; real NetBird credentials,
// kubeconfigs, bearer tokens, and CA data must come from an external secret flow.
func RenderNetBirdMultiClusterOverlay(req NetBirdMultiClusterOverlayRequest) (string, error) {
	req = defaultNetBirdMultiClusterOverlay(req)
	if err := validateNetBirdMultiClusterOverlay(req); err != nil {
		return "", err
	}

	var b strings.Builder
	fmt.Fprintf(&b, `apiVersion: v1
kind: ConfigMap
metadata:
  name: ubiquity-netbird-overlay-policy
  namespace: ubiquity-system
  labels:
    app.kubernetes.io/part-of: ubiquity-multicluster
    ubiquity.io/overlay: netbird
data:
  managementCluster: %s
  regionalCluster: %s
  architecture: |
    NetBird private control/data overlay connects independent Ubiquity clusters.
    Policy boundary: do not stretch one Kubernetes cluster across regions.
    public inference traffic must not hairpin through NetBird; use Geo DNS or a global load balancer.
  netbirdPolicies: |
    - sourceGroup: argocd-controller
      destinationGroup: regional-kube-apis
      ports: ["tcp/6443"]
      purpose: GitOps reconciliation over NetBird only
    - sourceGroup: platform-admins
      destinationGroup: argocd-ui
      ports: ["tcp/443"]
      purpose: Private ArgoCD UI/API access
    - sourceGroup: regional-clusters
      destinationGroup: regional-clusters
      ports: []
      purpose: No broad regional east-west access by default
  trafficPolicy: |
    publicInference: Geo DNS / global load balancer
    netbirdPublicHairpin: forbidden
    localIngressRequired: true
  readinessGates: |
    - ubiquity cloud collect-readiness
    - ubiquity cloud readiness --readiness-file /tmp/cloud-readiness-evidence.json
    - ubiquity health --ai
    - rdma-network-smoke-test-passed
    - nico-policy-reconciled
    - nvidia.com/rdma allocatable when ubiquity.io/rdma=true
`, req.ManagementCluster, req.RegionalCluster)

	fmt.Fprintf(&b, `---
apiVersion: v1
kind: Secret
metadata:
  name: %s
  namespace: %s
  labels:
    argocd.argoproj.io/secret-type: cluster
    ubiquity.io/region: %s
    ubiquity.io/site: %s
    ubiquity.io/trust-tier: %s
    ubiquity.io/gpu: "true"
    ubiquity.io/gpu-class: %s
    ubiquity.io/rdma: "true"
    ubiquity.io/inference: "true"
    ubiquity.io/storage: %s
  annotations:
    ubiquity.io/netbird-resource: regional-kube-api
    ubiquity.io/readiness-gate: collect-readiness-and-ai-health
    ubiquity.io/secret-template-only: "true"
    ubiquity.io/credential-source: external-secret-manager
type: Opaque
stringData:
  name: %s
  server: %s
  config: |
    {
      "bearerToken": "PLACEHOLDER_BEARER_TOKEN_FROM_REMOTE_CLUSTER_SERVICE_ACCOUNT",
      "tlsClientConfig": {
        "caData": "PLACEHOLDER_BASE64_CLUSTER_CA_FROM_REMOTE_CLUSTER",
        "insecure": false
      }
    }
`, req.RegionalCluster, req.Namespace, req.Region, req.Site, req.TrustTier, req.GPUClass, req.StorageProvider, req.RegionalCluster, req.NetBirdServer)

	fmt.Fprintf(&b, `---
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: ubiquity-regional-ai-platform
  namespace: %s
spec:
  generators:
    - matrix:
        generators:
          - clusters:
              selector:
                matchLabels:
                  ubiquity.io/inference: "true"
                  ubiquity.io/gpu: "true"
          - git:
              repoURL: %s
              revision: %s
              directories:
                - path: system/network-policies
                - path: system/nvidia-gpu-operator
                - path: system/nvidia-network-operator
                - path: system/nvidia-nic-configuration-operator
                - path: platform/nim-operator
                - path: platform/ai-workload-tenancy
  template:
    metadata:
      name: '{{name}}-{{path.basename}}'
      labels:
        ubiquity.io/region: '{{metadata.labels.ubiquity\.io/region}}'
        ubiquity.io/site: '{{metadata.labels.ubiquity\.io/site}}'
        ubiquity.io/rdma: '{{metadata.labels.ubiquity\.io/rdma}}'
        ubiquity.io/inference: '{{metadata.labels.ubiquity\.io/inference}}'
    spec:
      project: regional-ai-platform
      source:
        repoURL: %s
        targetRevision: %s
        path: '{{path}}'
      destination:
        server: '{{server}}'
        namespace: '{{path.basename}}'
      syncPolicy:
        automated:
          prune: true
          selfHeal: true
        syncOptions:
          - CreateNamespace=true
          - ApplyOutOfSyncOnly=true
`, req.Namespace, req.RepoURL, req.TargetRevision, req.RepoURL, req.TargetRevision)

	fmt.Fprintf(&b, `---
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: ubiquity-rdma-readiness-smoke
  namespace: %s
spec:
  generators:
    - clusters:
        selector:
          matchLabels:
            ubiquity.io/rdma: "true"
            ubiquity.io/gpu: "true"
  template:
    metadata:
      name: '{{name}}-rdma-smoke'
      labels:
        ubiquity.io/region: '{{metadata.labels.ubiquity\.io/region}}'
        ubiquity.io/rdma: '{{metadata.labels.ubiquity\.io/rdma}}'
    spec:
      project: regional-ai-platform
      source:
        repoURL: %s
        targetRevision: %s
        path: test/e2e/manifests/nvidia-rdma-smoke
      destination:
        server: '{{server}}'
        namespace: ubiquity-readiness
      syncPolicy:
        automated:
          prune: true
          selfHeal: true
        syncOptions:
          - CreateNamespace=true
`, req.Namespace, req.RepoURL, req.TargetRevision)
	return b.String(), nil
}

func defaultNetBirdMultiClusterOverlay(req NetBirdMultiClusterOverlayRequest) NetBirdMultiClusterOverlayRequest {
	if req.Namespace == "" {
		req.Namespace = "argocd"
	}
	if req.TrustTier == "" {
		req.TrustTier = "production"
	}
	if req.StorageProvider == "" {
		req.StorageProvider = "vast"
	}
	if req.GPUClass == "" {
		req.GPUClass = "h100"
	}
	if req.NetBirdServer == "" {
		req.NetBirdServer = "https://NETBIRD_OVERLAY_IP_OR_DNS:6443"
	}
	if req.RepoURL == "" {
		req.RepoURL = "https://github.com/ubiquitycluster/ubiquity.git"
	}
	if req.TargetRevision == "" {
		req.TargetRevision = "pinned-by-gitops"
	}
	return req
}

func validateNetBirdMultiClusterOverlay(req NetBirdMultiClusterOverlayRequest) error {
	if req.ManagementCluster == "" {
		return fmt.Errorf("management cluster is required")
	}
	if req.RegionalCluster == "" {
		return fmt.Errorf("regional cluster is required")
	}
	if req.Region == "" {
		return fmt.Errorf("region is required")
	}
	if req.Site == "" {
		return fmt.Errorf("site is required")
	}
	for field, value := range map[string]string{
		"management cluster": req.ManagementCluster,
		"regional cluster":   req.RegionalCluster,
		"namespace":          req.Namespace,
		"region":             req.Region,
		"site":               req.Site,
		"trust tier":         req.TrustTier,
		"gpu class":          req.GPUClass,
		"storage provider":   req.StorageProvider,
	} {
		if !kubeName.MatchString(value) {
			return fmt.Errorf("%s %q must be DNS-compatible", field, value)
		}
	}
	if req.AllowStretchedKubernetesWAN {
		return fmt.Errorf("NetBird overlay must not stretch one Kubernetes cluster across regions")
	}
	if req.PublicInferenceThroughMesh {
		return fmt.Errorf("public inference traffic must not hairpin through NetBird")
	}
	if strings.Contains(strings.ToLower(req.TargetRevision), "latest") {
		return fmt.Errorf("target revision must not use latest")
	}
	for field, value := range map[string]string{
		"NetBird server":  req.NetBirdServer,
		"GitOps repo":     req.RepoURL,
		"target revision": req.TargetRevision,
	} {
		if strings.ContainsAny(value, "\r\n") {
			return fmt.Errorf("%s must be single-line", field)
		}
	}
	for _, secretMarker := range []string{"nb_" + "pat_", "setup" + "key:", "BEGIN " + "PRIVATE KEY", "eyJ" + "hbGci"} {
		if strings.Contains(req.NetBirdServer+req.RepoURL+req.TargetRevision, secretMarker) {
			return fmt.Errorf("NetBird overlay request contains secret-like material")
		}
	}
	return nil
}
