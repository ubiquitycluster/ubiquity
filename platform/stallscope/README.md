# Stallscope GPU workload telemetry

This chart deploys [Stallscope](https://github.com/nshinde/stallscope) as a GPU-node DaemonSet for workload performance telemetry. It complements DCGM exporter: DCGM provides device-level GPU metrics, while Stallscope classifies active workload health as FAST, SLOW, FAIL_RISK, or UNKNOWN and correlates GPU utilization, host pressure, network errors, RDMA counters, PFC pause counters, and optional NCCL smoke evidence.

## What it collects

- NVIDIA GPU utilization and throttling evidence through `nvidia-smi`.
- Host network counters from `/proc/net/dev` via a read-only `/host/proc` mount.
- Host RDMA counters from `/sys/class/infiniband` via a read-only `/host/sys` mount.
- PFC pause counters from `ethtool -S` when available in the runtime image.
- Prometheus textfile metrics at `/var/lib/node_exporter/textfile_collector/stallscope.prom` for node-exporter collection.

## Deployment notes

The default chart pins upstream source archive commit `84b17513b9230ce66c2838bb4e6fe95f196a044c` from `nshinde/stallscope` and installs it into an ephemeral Python target directory at container start. For production, build and scan an internal image from the upstream Dockerfile, then set:

```yaml
image:
  repository: registry.example.com/observability/stallscope
  tag: "<pinned-build>"
```

The DaemonSet is restricted to GPU nodes with `nodeSelector.nvidia.com/gpu.present: "true"`. It does not request `nvidia.com/gpu` by default, so it should not reserve GPUs away from workloads. Set `runtimeClassName` or image/runtime settings as needed for your NVIDIA container runtime so `nvidia-smi` is visible inside the pod.

## Alerts

When `prometheusRule.enabled=true`, the chart emits alerts for:

- `StallscopeGPUWorkloadSlow`: Stallscope profile label is SLOW.
- `StallscopeGPUWorkloadFailRisk`: Stallscope profile label is FAIL_RISK.
- `StallscopeTelemetryUnavailable`: Stallscope is reporting alerts but no GPU utilization metric is present.

No credentials or webhook URLs are configured by default.
