package cloud

import (
	"fmt"
	"strings"
)

// CloudPrerequisitesRequest renders a contract documenting required CRDs/controllers for cloud primitives.
type CloudPrerequisitesRequest struct {
	Name      string
	Namespace string
}

// RequiredCloudCRDs returns the CRDs required by the currently supported cloud primitive renderers.
func RequiredCloudCRDs() []string {
	return []string{
		"datavolumes.cdi.kubevirt.io",
		"virtualmachines.kubevirt.io",
		"network-attachment-definitions.k8s.cni.cncf.io",
		"objectbucketclaims.objectbucket.io",
		"clusters.postgresql.cnpg.io",
		"redisfailovers.databases.spotahome.com",
		"kafkas.kafka.strimzi.io",
		"projects.goharbor.io",
		"clusters.cluster.x-k8s.io",
		"schedules.k8up.io",
		"volumesnapshotclasses.snapshot.storage.k8s.io",
	}
}

// RequiredCloudOperators returns expected operator/controller families for the cloud primitives.
func RequiredCloudOperators() []string {
	return []string{
		"KubeVirt CDI",
		"KubeVirt",
		"Multus",
		"ObjectBucket controller",
		"CloudNativePG",
		"Redis Operator",
		"Strimzi",
		"Harbor operator or compatible registry controller",
		"Cluster API",
		"Kamaji-compatible control-plane provider",
		"K8up",
		"snapshot-controller",
		"Longhorn CSI",
	}
}

// RenderCloudPrerequisites renders a ConfigMap contract for GitOps and CI checks.
func RenderCloudPrerequisites(req CloudPrerequisitesRequest) (string, error) {
	if req.Name == "" {
		req.Name = "cloud-prereqs"
	}
	if req.Namespace == "" {
		req.Namespace = "ubiquity-system"
	}
	if !kubeName.MatchString(req.Name) || !kubeName.MatchString(req.Namespace) {
		return "", fmt.Errorf("cloud prerequisite name and namespace must be DNS-compatible")
	}
	var b strings.Builder
	fmt.Fprintf(&b, `apiVersion: v1
kind: ConfigMap
metadata:
  name: %s
  namespace: %s
  labels:
    ubiquity.ai/prerequisite-contract: cloud
data:
  serverSideDryRunRequired: "true"
  reconciliationRequired: "true"
  restoreDrillRequired: "true"
  crds: |
`, req.Name, req.Namespace)
	for _, crd := range RequiredCloudCRDs() {
		fmt.Fprintf(&b, "    - %s\n", crd)
	}
	b.WriteString("  operators: |\n")
	for _, operator := range RequiredCloudOperators() {
		fmt.Fprintf(&b, "    - %s\n", operator)
	}
	return b.String(), nil
}
