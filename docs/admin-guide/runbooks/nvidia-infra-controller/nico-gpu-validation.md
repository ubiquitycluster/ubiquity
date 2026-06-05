# Runbook: NICo GPU validation

Status: experimental/preview. This runbook records local GPU validation evidence for NICo-managed Machines. Passing this runbook does not certify a GPU, driver, server, cluster, or NVIDIA software stack.

## Preconditions

- The Machine belongs to a GPU-capable pool.
- NICo reports Machine inventory and, when available, Machine GPU stats.
- The OS image and Kubernetes GPU stack are intentionally selected for the site.
- GPU Operator/DCGM components are deployed when required by the site design.
- The test workload is approved for the cluster and does not use secret material.

## Variables

```sh
export NICO_MACHINE=worker-gpu-01
export GPU_OPERATOR_NAMESPACE=gpu-operator
```

## NICo evidence

```sh
nicoctl machine get "${NICO_MACHINE}" --output yaml
nicoctl machine gpu-stats "${NICO_MACHINE}" --output yaml || true
```

Record whether expected accelerators are present. If the command is unavailable in the deployed NICo version, note that Machine GPU stats are not available and rely on Kubernetes/DCGM evidence instead.

## Kubernetes node evidence

```sh
kubectl get node "${NICO_MACHINE}" -o wide
kubectl get node "${NICO_MACHINE}" -o jsonpath='{.status.allocatable.nvidia\.com/gpu}{"\n"}' || true
kubectl get nodes -o json | grep -E 'nvidia.com/(gpu|mig-)' || true
```

For MIG pools, confirm the expected `nvidia.com/mig-*` resources rather than full GPU resources.

## GPU Operator evidence

```sh
kubectl -n "${GPU_OPERATOR_NAMESPACE}" get pods
kubectl -n "${GPU_OPERATOR_NAMESPACE}" rollout status daemonset/nvidia-device-plugin-daemonset --timeout=5m
kubectl -n "${GPU_OPERATOR_NAMESPACE}" get service nvidia-dcgm-exporter || true
```

## Smoke workload

Run a short `nvidia-smi` workload only on a cluster where GPU test pods are permitted:

```sh
kubectl run nvidia-smi-smoke \
  --rm -i \
  --restart=Never \
  --image=nvcr.io/nvidia/cuda:12.6.3-base-ubuntu22.04 \
  --limits=nvidia.com/gpu=1 \
  --overrides="{\"spec\":{\"nodeName\":\"${NICO_MACHINE}\"}}" \
  --command -- nvidia-smi
```

For MIG-only pools, use the appropriate MIG resource limit instead of `nvidia.com/gpu=1`.

## DCGM evidence

```sh
kubectl -n "${GPU_OPERATOR_NAMESPACE}" get service nvidia-dcgm-exporter
kubectl get --raw /api/v1/namespaces/${GPU_OPERATOR_NAMESPACE}/services/nvidia-dcgm-exporter:9400/proxy/metrics | grep -i dcgm | head
```

If DCGM exporter runs in a different namespace or through the monitoring stack, use the site-specific service path.

## Success criteria

- NICo Machine state is healthy.
- Machine GPU stats are present, or their absence is explicitly documented for the deployed NICo version.
- Kubernetes exposes expected GPU or MIG allocatable resources.
- Device plugin is ready.
- Optional smoke workload completes.
- DCGM or site monitoring emits GPU metrics when expected.

## Failure handling

- Keep the node cordoned if GPU resources are required for production.
- Preserve NICo Task IDs, kubelet logs, GPU Operator pod logs, and DCGM evidence.
- Check OS image, kernel headers, driver version, secure boot policy, PCIe visibility, and firmware.
- Do not describe partial local validation as certification.
