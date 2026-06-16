# NVIDIA Infra Controller node management

Status: experimental/preview. This page documents Ubiquity integration guidance for NVIDIA Infra Controller (NICo) day-2 bare-metal lifecycle operations. It is not a hardware, software, compliance, support, or certification claim. In short: no certification is implied by this documentation or the smoke tests.

Use this document only after the initial Ubiquity bootstrap boundary has been crossed and a management cluster exists. Do not commit BMC passwords, API tokens, private keys, kubeconfigs, or bearer tokens in Git. Use Vault, External Secrets Operator, or an equivalent approved secret manager.

## Scope

NICo is the preferred day-2 path for new Ubiquity bare-metal node lifecycle automation:

- discover and inventory physical Machines;
- register BMC Redfish access without exposing secrets in Git;
- assign an Operating System image to a Machine;
- create and track lifecycle Tasks;
- observe Machine, Instance, and Machine GPU stats status;
- reinstall, reboot, deprovision, and remove Machines in a controlled way.

Legacy Metal3 BareMetalHost and Ironic helper scripts are fallback/migration-only paths for sites that already depend on them. Do not operate the same Machine from both NICo and legacy BareMetalHost automation unless an explicit migration plan transfers ownership.

## Components and namespaces

Ubiquity wraps the upstream NVIDIA/infra-controller project with GitOps charts:

- `system/nvidia-infra-controller-prereqs`: documentation-as-code report for prerequisite ownership.
- `system/nvidia-infra-controller-core`: NICo Core controller services.
- `system/nvidia-nic-configuration-operator`: NVIDIA NIC Configuration Operator CRDs, manager, daemon, RBAC, and optional `NicConfigurationTemplate` objects for NIC-level firmware/configuration workflows.
- `platform/nvidia-infra-controller-rest`: NICo REST and site-agent services.

Default examples use namespace `nico-system`. Some deployments may use `nvidia-infra-controller`. Confirm the namespace in your GitOps values before running commands.

## Prerequisites

Before managing nodes with NICo, verify:

1. The management cluster is healthy.
2. MetalLB or an equivalent LoadBalancer is available where NICo services require stable VIPs.
3. cert-manager is installed or intentionally reused.
4. Vault or an equivalent secret store holds BMC, database, and API credentials.
5. External Secrets Operator projects only approved runtime secrets into NICo namespaces.
6. PostgreSQL is reachable and has a documented backup/restore owner.
7. Management, BMC, PXE/provisioning, and node networks are documented.
8. NVIDIA Network Operator and Maintenance Operator are installed before enabling NIC-level configuration via `system/nvidia-nic-configuration-operator`.
9. Operators understand that this integration is experimental/preview.

## Readiness checks

Use read-only checks first:

```sh
kubectl get ns nico-system nvidia-infra-controller --ignore-not-found
kubectl -n nico-system get deploy,sts,svc,pods || true
kubectl -n nvidia-infra-controller get deploy,sts,svc,pods || true
ubiquity health --nico || true
```

If your site exposes a NICo REST endpoint, query health without embedding secrets in command history:

```sh
: "${NICO_API:?set NICO_API to the NICo REST base URL}"
curl --fail --silent --show-error "${NICO_API}/healthz"
```

## Node lifecycle workflow

1. Confirm the node is physically cabled, reachable on the BMC network, and safe to operate.
2. Store BMC Redfish credentials in the approved secret manager.
3. Register or discover the Machine through NICo.
4. Confirm inventory fields: serial number, BMC address, boot MAC, platform, CPU, memory, disks, NICs, and accelerators.
5. Assign an Operating System image from the approved NICo OS image catalog.
6. Create an install or reinstall Task.
7. Watch Task status until the Machine and Instance reach a terminal healthy state.
8. Confirm Kubernetes node admission, labels, taints, GPU resources, and workload readiness where applicable.
9. Record the Task ID, Machine ID, image ID, and validation evidence in the change ticket.

## Add a node

Example sequence using placeholder commands. Adapt command names to the NICo CLI/API version deployed at the site:

```sh
export NICO_NAMESPACE=nico-system
export NICO_SITE=example-site
export NICO_MACHINE=worker-gpu-01
export NICO_OS_IMAGE=ubuntu-22.04-gpu

kubectl -n "${NICO_NAMESPACE}" get pods
nicoctl site get "${NICO_SITE}"
nicoctl machine discover --site "${NICO_SITE}"
nicoctl machine get "${NICO_MACHINE}" --output yaml
nicoctl machine assign-os "${NICO_MACHINE}" --os-image "${NICO_OS_IMAGE}"
nicoctl task create install --machine "${NICO_MACHINE}" --output yaml
nicoctl task wait --machine "${NICO_MACHINE}" --for condition=Succeeded --timeout 90m
```

Validation:

```sh
kubectl get node "${NICO_MACHINE}" -o wide
kubectl describe node "${NICO_MACHINE}" | grep -E 'nvidia.com/gpu|nvidia.com/mig-|nvidia.com/rdma' || true
nicoctl machine get "${NICO_MACHINE}" --output yaml
nicoctl machine gpu-stats "${NICO_MACHINE}" --output yaml || true
```

## Remove or deprovision a node

Remove one Machine at a time unless a site-specific maintenance plan explicitly allows parallel changes.

```sh
export NICO_MACHINE=worker-gpu-01
kubectl drain "${NICO_MACHINE}" --delete-emptydir-data --ignore-daemonsets --force
kubectl delete node "${NICO_MACHINE}"
nicoctl task create deprovision --machine "${NICO_MACHINE}" --output yaml
nicoctl task wait --machine "${NICO_MACHINE}" --for condition=Succeeded --timeout 60m
nicoctl machine get "${NICO_MACHINE}" --output yaml
```

Only delete inventory after confirming the Machine is powered down or returned to the expected pool state:

```sh
nicoctl machine delete "${NICO_MACHINE}"
```

## Reinstall a node

Use reinstall when ownership remains with NICo and the intended outcome is a clean OS on the same physical Machine.

For multi-OS boot or mixed-pool environments, choose an approved Operating System image record for the exact OS/version/architecture/boot-mode/driver stack. Treat switching OS families as a destructive reinstall that requires a change record, a rollback image, and secret references sourced from Vault/External Secrets rather than image-embedded credentials.

```sh
export NICO_MACHINE=worker-gpu-01
export NICO_OS_IMAGE=ubuntu-22.04-gpu
kubectl cordon "${NICO_MACHINE}"
kubectl drain "${NICO_MACHINE}" --delete-emptydir-data --ignore-daemonsets --force
nicoctl machine assign-os "${NICO_MACHINE}" --os-image "${NICO_OS_IMAGE}"
nicoctl task create reinstall --machine "${NICO_MACHINE}" --output yaml
nicoctl task wait --machine "${NICO_MACHINE}" --for condition=Succeeded --timeout 90m
kubectl uncordon "${NICO_MACHINE}" || true
```

## Operational safety

- Prefer read-only inspection before mutating Tasks.
- Keep all BMC credentials in a secret manager; never place secret literals in docs, Git, shell scripts, or tickets.
- Require maintenance windows for power, boot order, disk wipe, or reinstall actions.
- Keep NICo Task IDs in change records.
- Do not claim that a successful smoke test certifies a platform. It only records local validation evidence.
- Treat mock E2E scripts as API and process checks, not hardware proof.
- Treat bare-metal E2E scripts as destructive unless the target Machine has been dedicated for testing.

## Related documentation

- `docs/architecture/on-prem/nvidia-infra-controller-node-lifecycle.md`
- `docs/admin-guide/runbooks/nvidia-infra-controller/prerequisites.md`
- `docs/reference/nvidia-infra-controller/bootstrap-to-nico.md`
- `docs/reference/nvidia-infra-controller/node-state-machine.md`
- `docs/reference/nvidia-infra-controller/os-images.md`
- `docs/reference/nvidia-infra-controller/status-aggregation.md`
- `docs/admin-guide/runbooks/nvidia-infra-controller/nico-bootstrap.md`
- `docs/admin-guide/runbooks/nvidia-infra-controller/nico-machine-provisioning.md`
- `docs/admin-guide/runbooks/nvidia-infra-controller/nico-node-reinstall.md`
- `docs/admin-guide/runbooks/nvidia-infra-controller/nico-bmc-redfish.md`
- `docs/admin-guide/runbooks/nvidia-infra-controller/nico-gpu-validation.md`
- `docs/admin-guide/runbooks/nvidia-infra-controller/nic-configuration-operator.md`
