# Runbook: NICo Machine provisioning

Status: experimental/preview. This runbook provisions one bare-metal Machine through NVIDIA Infra Controller (NICo). It is destructive to the selected Machine and is not a certification claim.

## Preconditions

- The target Machine is dedicated to this change window.
- The Machine has one lifecycle owner: NICo.
- BMC Redfish credentials are stored in the approved secret manager.
- The boot MAC, BMC address, serial number, rack location, and intended OS image are known.
- The OS image is approved for this site or explicitly marked as candidate for this test.
- A rollback image or deprovision plan exists.

## Variables

```sh
export NICO_NAMESPACE=nico-system
export NICO_SITE=example-site
export NICO_MACHINE=worker-gpu-01
export NICO_OS_IMAGE=ubuntu-22.04-gpu-2026-06-01
```

## Procedure

1. Confirm NICo readiness:

```sh
kubectl -n "${NICO_NAMESPACE}" get pods,svc
ubiquity health --nico || true
```

2. Discover or refresh inventory:

```sh
nicoctl site get "${NICO_SITE}"
nicoctl machine discover --site "${NICO_SITE}"
nicoctl machine get "${NICO_MACHINE}" --output yaml
```

3. Review inventory. Confirm the target is the intended physical server before continuing.

4. Assign the Operating System:

```sh
nicoctl machine assign-os "${NICO_MACHINE}" --os-image "${NICO_OS_IMAGE}"
```

5. Start provisioning:

```sh
nicoctl task create install --machine "${NICO_MACHINE}" --output yaml
```

6. Watch the Task:

```sh
nicoctl task list --machine "${NICO_MACHINE}"
nicoctl task wait --machine "${NICO_MACHINE}" --for condition=Succeeded --timeout 90m
```

7. Validate status:

```sh
nicoctl machine get "${NICO_MACHINE}" --output yaml
nicoctl instance get --machine "${NICO_MACHINE}" --output yaml || true
kubectl get node "${NICO_MACHINE}" -o wide || true
```

## Success criteria

- Latest install Task is `Succeeded`.
- Machine state is `Ready` or site-equivalent healthy state.
- Instance state is `Ready` when an Instance is expected.
- Kubernetes node is `Ready` when the Machine is a cluster node.
- GPU/RDMA status is validated separately for GPU/RDMA pools.

## Failure handling

1. Do not immediately rerun destructive Tasks.
2. Capture Task details, NICo logs, BMC event logs, and console output.
3. Mark the Machine `Cordoned` or site-equivalent maintenance state.
4. Escalate to platform and hardware owners with the Task ID and Machine ID.

## Evidence to record

- Machine ID and hostname.
- OS image ID and checksum.
- Task ID.
- Start and end time.
- Final Machine and Instance states.
- Any local validation output.
