package cloud

import (
	"strings"
	"testing"
)

func TestMultiClusterNetBirdArchitectureContract(t *testing.T) {
	doc := mustRead(t, "../../docs/architecture/multi-cluster-netbird.md")
	combined := strings.ToLower(doc)
	for _, required := range []string{
		"netbird is the private control/data overlay",
		"independent ubiquity clusters",
		"do not stretch one kubernetes cluster across regions",
		"geo dns",
		"global load balancer",
		"public inference traffic must not hairpin through netbird",
		"rdma/nccl",
		"nvidia nic configuration operator",
		"nico",
		"argocd applicationset",
		"ubiquity.io/region",
		"ubiquity.io/rdma",
		"live evidence",
	} {
		if !strings.Contains(combined, strings.ToLower(required)) {
			t.Fatalf("multi-cluster NetBird architecture doc missing %q", required)
		}
	}
	for _, forbidden := range []string{"TODO", "TBD", "FIXME", "<real", "REDACTED PRIVATE KEY"} {
		if strings.Contains(doc, forbidden) {
			t.Fatalf("multi-cluster NetBird architecture doc contains unresolved or secret-like marker %q", forbidden)
		}
	}
}

func TestMultiClusterNetBirdReferenceExamplesArePlaceholderSafe(t *testing.T) {
	applicationSet := mustRead(t, "../../docs/reference/multi-cluster-netbird/application-set.yaml")
	clusterSecret := mustRead(t, "../../docs/reference/multi-cluster-netbird/cluster-secret-template.yaml")
	policyMatrix := mustRead(t, "../../docs/reference/multi-cluster-netbird/netbird-policy-matrix.yaml")

	for _, check := range []struct {
		name     string
		content  string
		required []string
	}{
		{
			name:    "application set",
			content: applicationSet,
			required: []string{
				"kind: ApplicationSet",
				"ubiquity.io/region",
				"ubiquity.io/rdma",
				"ubiquity.io/inference",
				"nvidia-nic-configuration-operator",
			},
		},
		{
			name:    "cluster secret",
			content: clusterSecret,
			required: []string{
				"argocd.argoproj.io/secret-type: cluster",
				"ubiquity.io/region",
				"ubiquity.io/site",
				"ubiquity.io/gpu",
				"https://NETBIRD_OVERLAY_IP_OR_DNS:6443",
				"PLACEHOLDER_BEARER_TOKEN_FROM_REMOTE_CLUSTER_SERVICE_ACCOUNT",
			},
		},
		{
			name:    "policy matrix",
			content: policyMatrix,
			required: []string{
				"argocd-controller",
				"regional-kube-apis",
				"platform-admins",
				"argocd-ui",
				"No broad regional east-west access by default",
			},
		},
	} {
		for _, required := range check.required {
			if !strings.Contains(check.content, required) {
				t.Fatalf("%s example missing %q", check.name, required)
			}
		}
	}

	for path, content := range map[string]string{
		"application-set.yaml":         applicationSet,
		"cluster-secret-template.yaml": clusterSecret,
		"netbird-policy-matrix.yaml":   policyMatrix,
	} {
		for _, forbidden := range []string{"nb_pat_", "BEGIN PRIVATE KEY", "REDACTED PRIVATE KEY", "eyJhbGci", "setupkey:"} {
			if strings.Contains(content, forbidden) {
				t.Fatalf("%s contains secret-like material %q", path, forbidden)
			}
		}
	}
}
