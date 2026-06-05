#!/usr/bin/env bash
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

kubectl -n gpu-operator get service nvidia-dcgm-exporter || kubectl -n monitoring-system get service dcgm-exporter
kubectl get --raw /api/v1/namespaces/gpu-operator/services/nvidia-dcgm-exporter:9400/proxy/metrics | grep -i dcgm

kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.status.allocatable.nvidia\.com/rdma}{"\n"}{end}' | grep -E '[1-9][0-9]*$'
kubectl get network-attachment-definitions.k8s.cni.cncf.io -A
kubectl get network-attachment-definitions.k8s.cni.cncf.io -A | grep -E 'rdma|ipoib'
kubectl -n kube-system get pods | grep -E 'whereabouts|multus' || true
kubectl -n gpu-operator create configmap rdma-network-smoke-test-passed \
  --from-literal=resource=nvidia.com/rdma \
  --from-literal=networkAttachment=rdma-ipoib \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl -n nim-operator get pods
kubectl -n nim-service get nimservice
kubectl -n nim-operator get configmap nim-smoke-test-passed

kubectl -n kai-scheduler rollout status deploy/kai-operator --timeout=5m
kubectl -n kai-scheduler rollout status deploy/kai-scheduler-default --timeout=5m
kubectl -n kai-scheduler rollout status deploy/binder --timeout=5m
kubectl -n kai-scheduler rollout status deploy/admission --timeout=5m
kubectl -n kai-scheduler rollout status deploy/pod-grouper --timeout=5m
kubectl -n kai-scheduler rollout status deploy/podgroup-controller --timeout=5m
kubectl -n kai-scheduler rollout status deploy/queue-controller --timeout=5m
kubectl get queues.scheduling.run.ai default-queue
kubectl run kai-scheduler-smoke --rm -i --restart=Never \
  --image=busybox:1.36 \
  --overrides='{"spec":{"schedulerName":"kai-scheduler"}}' \
  --command -- sh -c 'echo scheduled-by-kai-scheduler-default'
kubectl -n kai-scheduler create configmap kai-scheduling-smoke-test-passed \
  --from-literal=scheduler=kai-scheduler-default \
  --from-literal=queue=default-queue \
  --dry-run=client -o yaml | kubectl apply -f -
