package cloud

import (
	"strings"
	"testing"
)

func TestRenderTenantVPCCreatesNamespaceQuotaNetworkIsolationAndMultusNAD(t *testing.T) {
	manifest, err := RenderTenantVPC(TenantVPCRequest{
		Tenant:      "tenant-a",
		CIDR:        "10.60.0.0/24",
		Gateway:     "10.60.0.1",
		Bridge:      "br-tenant-a",
		CPUQuota:    "200",
		MemoryQuota: "1Ti",
		GPUQuota:    "8",
	})
	if err != nil {
		t.Fatalf("RenderTenantVPC returned error: %v", err)
	}
	for _, required := range []string{
		"kind: Namespace", "name: tenant-a",
		"kind: ResourceQuota", "requests.cpu: \"200\"", "requests.memory: 1Ti", "nvidia.com/gpu: \"8\"",
		"kind: NetworkAttachmentDefinition", "bridge", "br-tenant-a", "10.60.0.0/24",
		"kind: NetworkPolicy", "tenant-a-default-deny", "tenant-a-allow-same-tenant",
		"ubiquity.ai/vpc-cidr: 10.60.0.0/24",
	} {
		if !strings.Contains(manifest, required) {
			t.Fatalf("manifest missing %q:\n%s", required, manifest)
		}
	}
}

func TestRenderTenantVPCFailsClosedForInvalidTenantOrCIDR(t *testing.T) {
	_, err := RenderTenantVPC(TenantVPCRequest{Tenant: "Tenant_A"})
	if err == nil || !strings.Contains(err.Error(), "tenant") {
		t.Fatalf("expected tenant validation error, got %v", err)
	}
	_, err = RenderTenantVPC(TenantVPCRequest{Tenant: "tenant-a", CIDR: "not-a-cidr"})
	if err == nil || !strings.Contains(err.Error(), "CIDR") {
		t.Fatalf("expected CIDR validation error, got %v", err)
	}
}
