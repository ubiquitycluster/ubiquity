# NVIDIA Infra Controller status aggregation

Status: experimental/preview. This reference describes how Ubiquity should aggregate NICo, Kubernetes, BMC, and GPU evidence into operator-facing status. It is not a certification, compliance, or support statement.

## Goal

Status aggregation should answer four questions without exposing secrets:

1. Is NICo itself ready?
2. Can NICo see the site and Machines?
3. Are lifecycle Tasks progressing or complete?
4. Are provisioned nodes healthy in Kubernetes and, where relevant, showing GPU/RDMA evidence?

## Inputs

Use read-only sources where possible:

- NICo Core workload status.
- NICo REST health endpoint.
- site-agent readiness and site visibility.
- NICo Machine status.
- NICo Instance status.
- NICo Task status and timestamps.
- Machine GPU stats when available.
- Kubernetes Node conditions, labels, taints, and allocatable resources.
- GPU Operator/DCGM metrics where deployed.
- BMC Redfish power state and event evidence through approved NICo paths.

## Aggregated readiness levels

| Level | Meaning | Example condition |
| --- | --- | --- |
| `Unavailable` | NICo cannot be queried reliably. | REST unavailable or site-agent down. |
| `InventoryOnly` | Machines are visible but not provisioned. | Machine `Discovered` or `Registered`. |
| `Changing` | A lifecycle Task is active. | Task `Pending`, `Running`, or waiting. |
| `Ready` | Machine and Instance are healthy for intended use. | Task `Succeeded`, Instance `Ready`, Kubernetes node `Ready`. |
| `Degraded` | Some health evidence is missing or failing. | GPU stats missing for GPU pool, kubelet not ready, BMC warning. |
| `Failed` | A terminal error requires intervention. | Task `Failed` or Machine `Failed`. |

## Recommended status fields

```yaml
machine: worker-gpu-01
site: example-site
aggregateStatus: Ready
nico:
  machineState: Ready
  instanceState: Ready
  lastTask: install-1234
  lastTaskState: Succeeded
kubernetes:
  nodeReady: true
  schedulable: true
  gpuAllocatable: 8
  rdmaAllocatable: 1
gpu:
  machineGpuStatsPresent: true
  dcgmMetricsPresent: true
bmc:
  powerState: On
  redfishReachable: true
notes: Experimental/preview local validation evidence only.
```

## Precedence

When sources disagree, report the most conservative state:

1. `Failed` if any active or latest required Task has a terminal failure.
2. `Unavailable` if NICo REST or site-agent cannot be queried.
3. `Changing` if a Task is active.
4. `Degraded` if Machine, Instance, Kubernetes, GPU, or BMC evidence is missing for the intended pool.
5. `Ready` only when all required evidence is present.

## GPU and RDMA evidence

For GPU pools, aggregate at least one of:

- NICo Machine GPU stats showing expected accelerators;
- Kubernetes allocatable `nvidia.com/gpu` or `nvidia.com/mig-*` resources;
- DCGM exporter metrics reachable through the monitoring stack.

For RDMA pools, aggregate at least one of:

- Kubernetes allocatable `nvidia.com/rdma` or site-defined RDMA resource;
- NetworkAttachmentDefinition for RDMA/IPoIB workloads;
- site-specific NIC health evidence.

Missing GPU/RDMA evidence should be `Degraded`, not `Ready`, for pools where those capabilities are required.

## Secret handling

Status output must not include:

- BMC usernames or passwords;
- API tokens;
- kubeconfig contents;
- private keys;
- unredacted bearer tokens;
- full secret object data.

Store detailed credentials only in the approved secret manager and reference secret names, not values.
