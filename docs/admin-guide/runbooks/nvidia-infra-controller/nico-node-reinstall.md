# Runbook: NICo node reinstall

Status: experimental/preview. This runbook reinstalls one NICo-managed node. It wipes or replaces the node OS according to the selected Operating System image policy. It is not a certification claim.

## Preconditions

- The node is safe to remove from service.
- Application owners approved disruption.
- The target Machine is owned by NICo, not legacy BareMetalHost automation.
- The replacement OS image is approved or explicitly candidate-scoped for this Machine.
- The BMC is reachable through the approved NICo path.
- You have a rollback or deprovision plan.

## Variables

```sh
export NICO_NAMESPACE=nico-system
export NICO_MACHINE=worker-gpu-01
export NICO_OS_IMAGE=ubuntu-22.04-gpu-2026-06-01
```

## Procedure

1. Confirm Machine and node identity:

```sh
nicoctl machine get "${NICO_MACHINE}" --output yaml
kubectl get node "${NICO_MACHINE}" -o wide
```

2. Cordon and drain if it is a Kubernetes node:

```sh
kubectl cordon "${NICO_MACHINE}"
kubectl drain "${NICO_MACHINE}" --delete-emptydir-data --ignore-daemonsets --force
```

3. Assign or confirm the target Operating System:

```sh
nicoctl machine assign-os "${NICO_MACHINE}" --os-image "${NICO_OS_IMAGE}"
```

4. Create the reinstall Task:

```sh
nicoctl task create reinstall --machine "${NICO_MACHINE}" --output yaml
```

5. Watch progress:

```sh
nicoctl task wait --machine "${NICO_MACHINE}" --for condition=Succeeded --timeout 90m
nicoctl machine get "${NICO_MACHINE}" --output yaml
```

6. Validate node return:

```sh
kubectl get node "${NICO_MACHINE}" -o wide
kubectl describe node "${NICO_MACHINE}" | grep -E 'Ready|nvidia.com/gpu|nvidia.com/mig-|nvidia.com/rdma' || true
```

7. Uncordon when validation is complete:

```sh
kubectl uncordon "${NICO_MACHINE}"
```

## Success criteria

- Reinstall Task is `Succeeded`.
- Machine and Instance report healthy states.
- Kubernetes node rejoins and is `Ready` if applicable.
- GPU/RDMA evidence is present for accelerator pools.
- No secrets were printed or committed.

## Failure handling

- Keep the node cordoned.
- Capture Task output, NICo controller logs, site-agent logs, BMC event logs, and console screenshots or text logs.
- Do not switch to legacy BareMetalHost tooling unless the ownership transfer is approved and recorded.
- If the Machine must be returned to service quickly, use the documented rollback image or deprovision plan.
