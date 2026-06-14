#!/usr/bin/env bash
set -euo pipefail

if [ "${UBIQUITY_RUN_NICO_BARE_METAL_E2E:-}" != "true" ]; then
  echo "Skipping NICo bare-metal node lifecycle E2E. Set UBIQUITY_RUN_NICO_BARE_METAL_E2E=true on a dedicated hardware test Machine to run."
  exit 0
fi

if [ "${UBIQUITY_ACK_NICO_BARE_METAL_DESTRUCTIVE:-}" != "true" ]; then
  echo "Refusing to run destructive NICo bare-metal E2E. Set UBIQUITY_ACK_NICO_BARE_METAL_DESTRUCTIVE=true after dedicating the target Machine."
  exit 2
fi

NICOCTL_BIN="${NICOCTL_BIN:-nicoctl}"
NICO_NAMESPACE="${NICO_NAMESPACE:-nico-system}"
NICO_SITE="${NICO_SITE:?set NICO_SITE}"
NICO_MACHINE="${NICO_MACHINE:?set NICO_MACHINE to the dedicated test Machine name}"
NICO_OS_IMAGE="${NICO_OS_IMAGE:?set NICO_OS_IMAGE to the approved test Operating System image}"
NICO_TASK_TIMEOUT="${NICO_TASK_TIMEOUT:-90m}"
GPU_OPERATOR_NAMESPACE="${GPU_OPERATOR_NAMESPACE:-gpu-operator}"
RUN_GPU_VALIDATION="${NICO_RUN_GPU_VALIDATION:-false}"
RUN_DEPROVISION="${NICO_RUN_DEPROVISION:-false}"

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Required command not found: $1" >&2
    exit 127
  }
}

run_nico() {
  "$NICOCTL_BIN" "$@"
}

require_cmd "$NICOCTL_BIN"
require_cmd kubectl

cleanup_note() {
  echo "If this script failed mid-run, leave ${NICO_MACHINE} cordoned and inspect NICo Task logs before retrying."
}
trap cleanup_note ERR

echo "Running destructive NICo bare-metal E2E for Machine ${NICO_MACHINE} in site ${NICO_SITE}."
echo "This records local validation evidence only and is not a certification claim."

kubectl -n "$NICO_NAMESPACE" get pods,svc
run_nico site get "$NICO_SITE" --output json
run_nico machine get "$NICO_MACHINE" --output json

if kubectl get node "$NICO_MACHINE" >/dev/null 2>&1; then
  kubectl cordon "$NICO_MACHINE"
  kubectl drain "$NICO_MACHINE" --delete-emptydir-data --ignore-daemonsets --force
else
  echo "Kubernetes node ${NICO_MACHINE} not present before install/reinstall; continuing."
fi

run_nico machine assign-os "$NICO_MACHINE" --os-image "$NICO_OS_IMAGE"
run_nico task create reinstall --machine "$NICO_MACHINE" --output json
run_nico task wait --machine "$NICO_MACHINE" --for condition=Succeeded --timeout "$NICO_TASK_TIMEOUT"
run_nico machine get "$NICO_MACHINE" --output json
run_nico instance get --machine "$NICO_MACHINE" --output json || true

kubectl wait --for=condition=Ready "node/${NICO_MACHINE}" --timeout=20m
kubectl get node "$NICO_MACHINE" -o wide

if [ "$RUN_GPU_VALIDATION" = "true" ]; then
  kubectl get node "$NICO_MACHINE" -o json | grep -E 'nvidia.com/(gpu|mig-)'
  kubectl -n "$GPU_OPERATOR_NAMESPACE" rollout status daemonset/nvidia-device-plugin-daemonset --timeout=5m
  run_nico machine gpu-stats "$NICO_MACHINE" --output json || true
  kubectl run nico-nvidia-smi-smoke \
    --rm -i \
    --restart=Never \
    --image=nvcr.io/nvidia/cuda:12.6.3-base-ubuntu22.04 \
    --limits=nvidia.com/gpu=1 \
    --overrides="{\"spec\":{\"nodeName\":\"${NICO_MACHINE}\"}}" \
    --command -- nvidia-smi
else
  echo "Skipping GPU validation. Set NICO_RUN_GPU_VALIDATION=true for GPU-capable dedicated hardware."
fi

if [ "$RUN_DEPROVISION" = "true" ]; then
  kubectl cordon "$NICO_MACHINE" || true
  kubectl drain "$NICO_MACHINE" --delete-emptydir-data --ignore-daemonsets --force || true
  run_nico task create deprovision --machine "$NICO_MACHINE" --output json
  run_nico task wait --machine "$NICO_MACHINE" --for condition=Succeeded --timeout "$NICO_TASK_TIMEOUT"
else
  kubectl uncordon "$NICO_MACHINE" || true
  echo "Skipping deprovision. Set NICO_RUN_DEPROVISION=true only for disposable hardware."
fi

echo "NICo bare-metal node lifecycle E2E completed for ${NICO_MACHINE}."
