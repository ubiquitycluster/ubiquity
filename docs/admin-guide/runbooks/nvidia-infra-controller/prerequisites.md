# NVIDIA Infra Controller prerequisites

This runbook documents the prerequisites for the Ubiquity NICo Helm/GitOps wrappers:

- `system/nvidia-infra-controller-core`
- `platform/nvidia-infra-controller-rest`
- `system/nvidia-infra-controller-prereqs`

All three wrappers are marked as experimental/preview and track source repository `NVIDIA/infra-controller`.

## Required platform services

### MetalLB or an equivalent LoadBalancer

NICo Core services such as PXE, DHCP/DNS helpers, BMC proxying, and REST/API entry points may need stable management-plane VIPs on bare metal. Reuse an existing L2/BGP load balancer if the cluster already has one. Deploy the Ubiquity `system/metallb-system` wrapper only when no cluster LoadBalancer implementation exists.

### cert-manager

Use cert-manager for TLS serving certificates, webhook certificates, and internal service certificates. Reuse existing ClusterIssuers or Issuers when available. Do not create a second cert-manager installation in the same cluster.

### Vault or equivalent secret store

NICo needs credentials for BMC access, database access, IdP integration, and potentially image/provisioning endpoints. Store those values in Vault or an equivalent external secret manager. The NICo Helm wrappers do not render Kubernetes Secret manifests and should not contain secret literals in Git.

### External Secrets Operator

Use External Secrets Operator to project approved values from Vault or another backend into the namespaces consumed by NICo. This keeps GitOps manifests declarative without committing secret material.

### PostgreSQL

NICo Core/REST require PostgreSQL for persistent state. Prefer reusing a managed or already operated PostgreSQL instance. Deploy a PostgreSQL chart only when the platform team explicitly chooses in-cluster database ownership and backup/restore responsibilities.

## Reuse vs deploy guidance

Use `mode: reuse` for existing cluster capabilities and document the namespace/endpoint in values. Use `mode: deploy` only when Ubiquity owns that prerequisite in the target environment and the GitOps ordering ensures the prerequisite is healthy before NICo Core/REST sync.

The prereqs wrapper is render/report-only documentation-as-code: it renders a ConfigMap report and fails rendering if a default StorageClass takeover is requested without an explicit name. It does not install, upgrade, or configure MetalLB, cert-manager, Vault, External Secrets Operator, PostgreSQL, StorageClasses, or Kubernetes Secrets.

## StorageClass safety

NICo prerequisites must not change the cluster default StorageClass implicitly. The prereqs chart defaults to:

```yaml
storageClass:
  setDefault: false
  explicitName: ""
```

Only set `storageClass.setDefault: true` when `storageClass.explicitName` is populated and the platform team has approved that StorageClass as the default for the whole cluster. This prevents accidental takeover by Longhorn, local-path, or any other storage provider.

## Suggested GitOps waves

1. Reuse or deploy prerequisites: MetalLB, cert-manager, Vault, External Secrets, PostgreSQL.
2. Confirm secrets are projected by External Secrets and database connectivity exists.
3. Deploy `system/nvidia-infra-controller-core`.
4. Deploy `platform/nvidia-infra-controller-rest` with either external IdP configuration or explicit bundled IdP preview mode.

## OS image and boot prerequisites

For multi-OS boot, reinstall, or mixed-pool operations, pre-approve a separate NICo Operating System image record for each OS, version, architecture, boot mode, driver stack, and Machine pool. Keep image-access credentials in Vault or another approved external secret store and project only the required runtime references with External Secrets Operator.
