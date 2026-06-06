package cloud

import (
	"fmt"
	"strings"
)

// CloudOperatorBundle records operator/CRD ownership and air-gap artifact expectations.
type CloudOperatorBundle struct {
	Name             string
	Controller       string
	OwnedCRD         string
	InstallNamespace string
	Source           string
	Version          string
	AirgapArtifact   string
}

// CloudOperatorBundlesRequest renders the operator install-plan contract.
type CloudOperatorBundlesRequest struct {
	Name      string
	Namespace string
}

// RequiredCloudOperatorBundles returns Ubiquity's supported cloud-controller install plan.
func RequiredCloudOperatorBundles() []CloudOperatorBundle {
	return []CloudOperatorBundle{
		{Name: "kubevirt-cdi", Controller: "CDI Operator", OwnedCRD: "datavolumes.cdi.kubevirt.io", InstallNamespace: "cdi", Source: "https://github.com/kubevirt/containerized-data-importer", Version: "pinned-by-platform", AirgapArtifact: "platform/vendor/cdi"},
		{Name: "kubevirt", Controller: "KubeVirt Operator", OwnedCRD: "virtualmachines.kubevirt.io", InstallNamespace: "kubevirt", Source: "https://github.com/kubevirt/kubevirt", Version: "pinned-by-platform", AirgapArtifact: "platform/vendor/kubevirt"},
		{Name: "multus", Controller: "Multus CNI", OwnedCRD: "network-attachment-definitions.k8s.cni.cncf.io", InstallNamespace: "kube-system", Source: "https://github.com/k8snetworkplumbingwg/multus-cni", Version: "pinned-by-platform", AirgapArtifact: "platform/vendor/multus"},
		{Name: "cloudnative-pg", Controller: "CloudNativePG Operator", OwnedCRD: "clusters.postgresql.cnpg.io", InstallNamespace: "cnpg-system", Source: "https://github.com/cloudnative-pg/cloudnative-pg", Version: "pinned-by-platform", AirgapArtifact: "platform/vendor/cloudnative-pg"},
		{Name: "strimzi", Controller: "Strimzi Operator", OwnedCRD: "kafkas.kafka.strimzi.io", InstallNamespace: "strimzi-system", Source: "https://github.com/strimzi/strimzi-kafka-operator", Version: "pinned-by-platform", AirgapArtifact: "platform/vendor/strimzi"},
		{Name: "cluster-api", Controller: "Cluster API", OwnedCRD: "clusters.cluster.x-k8s.io", InstallNamespace: "capi-system", Source: "https://github.com/kubernetes-sigs/cluster-api", Version: "pinned-by-platform", AirgapArtifact: "platform/vendor/cluster-api"},
		{Name: "k8up", Controller: "K8up Operator", OwnedCRD: "schedules.k8up.io", InstallNamespace: "k8up-system", Source: "https://github.com/k8up-io/k8up", Version: "pinned-by-platform", AirgapArtifact: "platform/vendor/k8up"},
		{Name: "longhorn", Controller: "Longhorn CSI", OwnedCRD: "volumesnapshotclasses.snapshot.storage.k8s.io", InstallNamespace: "longhorn-system", Source: "https://github.com/longhorn/longhorn", Version: "pinned-by-platform", AirgapArtifact: "platform/vendor/longhorn"},
		{Name: "gateway-api", Controller: "Gateway API Controller", OwnedCRD: "gateways.gateway.networking.k8s.io", InstallNamespace: "gateway-system", Source: "https://gateway-api.sigs.k8s.io", Version: "pinned-by-platform", AirgapArtifact: "platform/vendor/gateway-api"},
		{Name: "external-dns", Controller: "external-dns", OwnedCRD: "dnsendpoints.externaldns.k8s.io", InstallNamespace: "external-dns", Source: "https://github.com/kubernetes-sigs/external-dns", Version: "pinned-by-platform", AirgapArtifact: "platform/vendor/external-dns"},
		{Name: "opencost", Controller: "OpenCost", OwnedCRD: "opencost.io/allocation", InstallNamespace: "opencost", Source: "https://github.com/opencost/opencost", Version: "pinned-by-platform", AirgapArtifact: "platform/vendor/opencost"},
		{Name: "kyverno", Controller: "Kyverno", OwnedCRD: "clusterpolicies.kyverno.io", InstallNamespace: "kyverno", Source: "https://github.com/kyverno/kyverno", Version: "pinned-by-platform", AirgapArtifact: "platform/vendor/kyverno"},
		{Name: "argocd", Controller: "Argo CD", OwnedCRD: "applications.argoproj.io", InstallNamespace: "argocd", Source: "https://github.com/argoproj/argo-cd", Version: "pinned-by-platform", AirgapArtifact: "platform/vendor/argocd"},
	}
}

// RenderCloudOperatorBundles renders a ConfigMap install-plan contract for GitOps and air-gap review.
func RenderCloudOperatorBundles(req CloudOperatorBundlesRequest) (string, error) {
	if req.Name == "" {
		req.Name = "cloud-operators"
	}
	if req.Namespace == "" {
		req.Namespace = "ubiquity-system"
	}
	if !kubeName.MatchString(req.Name) || !kubeName.MatchString(req.Namespace) {
		return "", fmt.Errorf("cloud operator bundle name and namespace must be DNS-compatible")
	}
	var b strings.Builder
	fmt.Fprintf(&b, `apiVersion: v1
kind: ConfigMap
metadata:
  name: %s
  namespace: %s
  labels:
    ubiquity.ai/operator-install-plan: cloud
data:
  bundles: |
`, req.Name, req.Namespace)
	for _, bundle := range RequiredCloudOperatorBundles() {
		fmt.Fprintf(&b, "    - name: %s\n", bundle.Name)
		fmt.Fprintf(&b, "      controller: %s\n", bundle.Controller)
		fmt.Fprintf(&b, "      ownedCRD: %s\n", bundle.OwnedCRD)
		fmt.Fprintf(&b, "      installNamespace: %s\n", bundle.InstallNamespace)
		fmt.Fprintf(&b, "      source: %s\n", bundle.Source)
		fmt.Fprintf(&b, "      version: %s\n", bundle.Version)
		fmt.Fprintf(&b, "      airgapArtifact: %s\n", bundle.AirgapArtifact)
	}
	return b.String(), nil
}
