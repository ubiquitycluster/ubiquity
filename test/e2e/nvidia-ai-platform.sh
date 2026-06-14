#!/usr/bin/env bash
# Explicitly gated real-GPU proof path. Composed checks must include:
# - nim-smoke-test evidence via nim-gpu-serving-smoke.sh
# - rdma-network-smoke-test-passed evidence via nvidia-rdma-smoke.sh for nvidia.com/rdma and network-attachment-definitions.k8s.cni.cncf.io
# - kai-scheduling-smoke-test-passed evidence via kai-scheduler-smoke.sh, kai-scheduler-default, and default-queue
set -euo pipefail

if [ "${UBIQUITY_RUN_GPU_E2E:-}" != "true" ]; then
  echo "Skipping NVIDIA AI platform GPU E2E. Set UBIQUITY_RUN_GPU_E2E=true on a real GPU cluster to run."
  exit 0
fi

kubectl get nodes -l nvidia.com/gpu.present=true
# Accept full GPU resources or MIG-partitioned nvidia.com/mig-* resources as accelerator capacity evidence.
kubectl get nodes -o json | grep -E 'nvidia.com/(gpu|mig-)'
kubectl -n gpu-operator rollout status deploy/gpu-operator --timeout=5m
kubectl -n gpu-operator get deploy gpu-operator -o jsonpath='{.status.readyReplicas} {.status.availableReplicas}{"\n"}' | grep -E '^[1-9][0-9]* [1-9][0-9]*$'
kubectl -n gpu-operator rollout status daemonset/nvidia-device-plugin-daemonset --timeout=5m
kubectl -n gpu-operator get daemonset nvidia-device-plugin-daemonset -o jsonpath='{.status.desiredNumberScheduled} {.status.numberReady} {.status.numberAvailable}{"\n"}' | grep -E '^[1-9][0-9]* [1-9][0-9]* [1-9][0-9]*$'

kubectl run nvidia-smi-smoke --rm -i --restart=Never \
  --image=nvcr.io/nvidia/cuda:12.6.3-base-ubuntu22.04 \
  --limits=nvidia.com/gpu=1 \
  --command -- nvidia-smi

kubectl -n gpu-operator get service nvidia-dcgm-exporter
kubectl get --raw /api/v1/namespaces/gpu-operator/services/nvidia-dcgm-exporter:9400/proxy/metrics | grep -i dcgm

UBIQUITY_RUN_NVIDIA_RDMA_SMOKE=true "$(dirname "$0")/nvidia-rdma-smoke.sh"

UBIQUITY_RUN_NIM_GPU_SMOKE=true "$(dirname "$0")/nim-gpu-serving-smoke.sh"

UBIQUITY_RUN_KAI_SMOKE=true "$(dirname "$0")/kai-scheduler-smoke.sh"
