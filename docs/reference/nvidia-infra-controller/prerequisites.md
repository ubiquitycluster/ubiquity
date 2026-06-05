# NVIDIA Infra Controller prerequisites

This reference summarizes the prerequisite contract for Ubiquity's experimental/preview NVIDIA Infra Controller (NICo) wrappers:

- `system/nvidia-infra-controller-prereqs`
- `system/nvidia-infra-controller-core`
- `platform/nvidia-infra-controller-rest`

Canonical operational runbook: [NVIDIA Infra Controller prerequisites](../../admin-guide/runbooks/nvidia-infra-controller/prerequisites.md)

`system/nvidia-infra-controller-prereqs` is render/report-only. It renders a
reviewable prerequisite ConfigMap and does not install, upgrade, or configure
MetalLB, cert-manager, Vault, External Secrets Operator, PostgreSQL,
StorageClasses, or Secrets.

## Required services

NICo deployments must either reuse or explicitly deploy these capabilities before day-2 node lifecycle ownership is handed to NICo:

- LoadBalancer support such as MetalLB or an existing equivalent.
- cert-manager for serving, webhook, and internal service certificates.
- Vault or an equivalent external secret store.
- External Secrets Operator or equivalent projection from the approved secret store.
- PostgreSQL with documented ownership, persistence, backup, and restore procedures.
- DNS, NTP, BMC/Redfish, provisioning, and management-plane network reachability for the target site.

## Safety defaults

- No secret literals belong in GitOps values or rendered manifests.
- Source BMC, database, IdP, API, and image-access credentials from Vault or an
  equivalent external secret store and project them with External Secrets
  Operator or an approved equivalent.
- The NICo prerequisites wrapper must not change the default StorageClass implicitly.
- BMO/Metal3 is fallback/migration-only and is excluded from root GitOps defaults; NICo is the default day-2 lifecycle backend.
- Ubiquity bootstrap automation should hand off to a single active lifecycle owner: NICo.
- Multi-OS boot or reinstall flows must use approved NICo Operating System image
  records for each OS/version/boot-mode combination rather than ad hoc image URLs.

## Validation

Render the prerequisites wrapper before enabling NICo Core or REST:

```sh
helm template nico-prereqs system/nvidia-infra-controller-prereqs
```

Then confirm the prerequisite owners and endpoints are recorded in the site change record before syncing the NICo GitOps wrappers.
