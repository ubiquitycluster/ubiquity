package cloud

import (
	"strings"
	"testing"
)

func TestRenderTenantKubernetesClusterComplementsNICOWithCAPIWorkloadCluster(t *testing.T) {
	manifest, err := RenderTenantKubernetesCluster(TenantClusterRequest{
		Name:              "tenant-a-dev",
		Namespace:         "tenant-a",
		KubernetesVersion: "v1.31.4",
		ControlPlaneClass: "kamaji",
		NodePoolClass:     "nico-managed-workers",
		WorkerReplicas:    3,
	})
	if err != nil {
		t.Fatalf("RenderTenantKubernetesCluster returned error: %v", err)
	}
	for _, required := range []string{
		"kind: Cluster", "apiVersion: cluster.x-k8s.io/v1beta1", "name: tenant-a-dev",
		"ubiquity.ai/node-lifecycle: nico-primary",
		"kind: TenantControlPlane", "controlPlaneClass: kamaji", "version: v1.31.4",
		"kind: MachineDeployment", "replicas: 3", "nico-managed-workers",
	} {
		if !strings.Contains(manifest, required) {
			t.Fatalf("manifest missing %q:\n%s", required, manifest)
		}
	}
}

func TestRenderTenantKubernetesClusterFailsClosedForInvalidVersionOrReplicas(t *testing.T) {
	_, err := RenderTenantKubernetesCluster(TenantClusterRequest{Name: "tenant-a-dev", KubernetesVersion: "1.31.4"})
	if err == nil || !strings.Contains(err.Error(), "version") {
		t.Fatalf("expected version validation error, got %v", err)
	}
	_, err = RenderTenantKubernetesCluster(TenantClusterRequest{Name: "tenant-a-dev", KubernetesVersion: "v1.31.4", WorkerReplicas: -1})
	if err == nil || !strings.Contains(err.Error(), "replicas") {
		t.Fatalf("expected replicas validation error, got %v", err)
	}
}
