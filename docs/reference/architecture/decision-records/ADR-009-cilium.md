# ADR-009: Use Cilium + Hubble for advanced networking and observability

**Status:** Accepted

**Date:** 2026-05-26

## Context

Ubiquity uses k3s with its built-in flannel CNI and MetalLB for Layer 2 load balancer IP
announcement. This works well for basic deployments but has limitations:

- **No network policy enforcement** (k3s disables its network policy controller when flannel is used)
  without installing a third-party CNI that supports it
- **No network observability** — flannel provides no visibility into traffic flows,
  dropped packets, or service dependencies
- **MetalLB is a separate component** — adds operational overhead and its own CRDs
- **kube-proxy overhead** — k3s runs kube-proxy by default, adding iptables rules
  for every Service. Cilium can replace kube-proxy entirely

The upstream homelab project migrated to Cilium for these reasons, and the Ubiquity
community has expressed interest in eBPF-based networking for better performance
and observability in HPC environments.

Alternatives considered:

1. **Keep flannel + MetalLB** — Lowest complexity. No new components to learn.
   But no observability, no network policies, no kube-proxy replacement.

2. **Calico** — Mature CNI with network policies and observability. However, it uses
   iptables/ipset under the hood, not eBPF. No built-in L2 load balancer IP announcement
   (requires separate MetalLB or similar). Heavier than Cilium on resource-constrained nodes.

3. **Cilium** — eBPF-based CNI that replaces flannel, kube-proxy, and MetalLB with
   a single component. Built-in Hubble for flow observability. L2/L4 load balancer
   IP announcement via CiliumL2AnnouncementPolicy. Kubernetes-native API (CRDs for
   IP pools, network policies). Active community, CNCF graduated.

4. **OVN-Kubernetes** — Open vSwitch-based CNI. Powerful but complex to operate.
   Less commonly used in the k3s ecosystem. No built-in L2 announcement.

## Decision

Use **Cilium** as the advanced CNI option with **Hubble** for network observability.
Cilium is optional — flannel + MetalLB remains the default for simplicity.
Users opt into Cilium via the documented migration path.

Key configuration for k3s compatibility:
- `k8sServiceHost: 127.0.0.1` — k3s exposes the API server on localhost
- `k8sServicePort: 6444` — k3s's default API server port (not 6443 like standard K8s)
- `kubeProxyReplacement: true` — Cilium replaces kube-proxy entirely
- `flannel-backend: none` — k3s must be configured to not start flannel
- `disable-kube-proxy: true` — k3s must be configured to not start kube-proxy

Hubble is deployed as part of Cilium (relay + UI), not as a separate component.

## Consequences

- Positive: Single component replaces flannel + kube-proxy + MetalLB.
- Positive: Hubble provides network flow visualization (service maps, dropped packet
  monitoring) useful for HPC workload debugging.
- Positive: eBPF-based networking is significantly more performant than iptables-based
  alternatives for high-throughput HPC traffic.
- Positive: Kubernetes-native API — CiliumLoadBalancerIPPool and
  CiliumL2AnnouncementPolicy are standard CRDs manageable via ArgoCD GitOps.
- Negative: Requires k3s reconfiguration (Ansible role change). Not a drop-in replacement.
- Negative: eBPF requires a recent Linux kernel (5.10+). Some older HPC nodes may not
  be compatible.
- Negative: Cilium's feature surface is larger than flannel, increasing the
  troubleshooting surface area.
- Neutral: Migration is opt-in and documented. Existing clusters can stay on flannel.
- Neutral: Hubble UI must be accessed via port-forward or Ingress (no default exposure).