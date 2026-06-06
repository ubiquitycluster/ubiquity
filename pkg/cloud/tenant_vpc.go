package cloud

import (
	"fmt"
	"net"
	"strings"
)

// TenantVPCRequest creates an isolated tenant network boundary without replacing Ubiquity's existing AI tenancy.
type TenantVPCRequest struct {
	Tenant      string
	CIDR        string
	Gateway     string
	Bridge      string
	CPUQuota    string
	MemoryQuota string
	GPUQuota    string
}

// RenderTenantVPC renders namespace, quota, Multus network, and deny-by-default policy primitives.
func RenderTenantVPC(req TenantVPCRequest) (string, error) {
	req = defaultTenantVPC(req)
	if err := validateTenantVPC(req); err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, `apiVersion: v1
kind: Namespace
metadata:
  name: %s
  labels:
    ubiquity.ai/tenant: %s
    ubiquity.ai/vpc-cidr: %s
---
apiVersion: v1
kind: ResourceQuota
metadata:
  name: %s-quota
  namespace: %s
spec:
  hard:
    requests.cpu: %q
    requests.memory: %s
`, req.Tenant, req.Tenant, req.CIDR, req.Tenant, req.Tenant, req.CPUQuota, req.MemoryQuota)
	if req.GPUQuota != "" {
		fmt.Fprintf(&b, "    nvidia.com/gpu: %q\n", req.GPUQuota)
	}
	fmt.Fprintf(&b, `---
apiVersion: k8s.cni.cncf.io/v1
kind: NetworkAttachmentDefinition
metadata:
  name: %s-vpc
  namespace: %s
  labels:
    ubiquity.ai/tenant: %s
    ubiquity.ai/network-isolation: multus
spec:
  config: '{"cniVersion":"0.3.1","type":"bridge","bridge":"%s","ipam":{"type":"host-local","subnet":"%s","gateway":"%s"}}'
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: %s-default-deny
  namespace: %s
spec:
  podSelector: {}
  policyTypes: ["Ingress", "Egress"]
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: %s-allow-same-tenant
  namespace: %s
spec:
  podSelector: {}
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              ubiquity.ai/tenant: %s
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              ubiquity.ai/tenant: %s
`, req.Tenant, req.Tenant, req.Tenant, req.Bridge, req.CIDR, req.Gateway, req.Tenant, req.Tenant, req.Tenant, req.Tenant, req.Tenant, req.Tenant)
	return b.String(), nil
}

func defaultTenantVPC(req TenantVPCRequest) TenantVPCRequest {
	if req.CIDR == "" {
		req.CIDR = "10.60.0.0/24"
	}
	if req.Gateway == "" {
		req.Gateway = "10.60.0.1"
	}
	if req.Bridge == "" && req.Tenant != "" {
		req.Bridge = "br-" + req.Tenant
	}
	if req.CPUQuota == "" {
		req.CPUQuota = "100"
	}
	if req.MemoryQuota == "" {
		req.MemoryQuota = "512Gi"
	}
	return req
}

func validateTenantVPC(req TenantVPCRequest) error {
	if !kubeName.MatchString(req.Tenant) {
		return fmt.Errorf("tenant %q must be DNS-compatible", req.Tenant)
	}
	if _, _, err := net.ParseCIDR(req.CIDR); err != nil {
		return fmt.Errorf("VPC CIDR %q is invalid: %w", req.CIDR, err)
	}
	if ip := net.ParseIP(req.Gateway); ip == nil {
		return fmt.Errorf("VPC gateway %q must be an IP address", req.Gateway)
	}
	if req.Bridge == "" {
		return fmt.Errorf("VPC bridge is required")
	}
	return nil
}
