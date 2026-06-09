#!/usr/bin/env bash
set -euo pipefail

EVIDENCE_NAMESPACE="${UBIQUITY_NVIDIA_AI_DEMO_NAMESPACE:-ubiquity-system}"
EVIDENCE_CONFIGMAP="nvidia-ai-final-demo-passed"
UBIQUITY_BIN="${UBIQUITY_BIN:-go run ./cmd/ubiquity}"

if [[ "${1:-}" == "--dry-run" ]]; then
  cat <<'MSG'
NVIDIA AI platform final demo dry-run:
- gated by UBIQUITY_RUN_NVIDIA_AI_FINAL_DEMO=true
- provision: render/prove the ai-production platform profile and prerequisites
- reconcile: apply GitOps ArgoCD Applications server-side
- schedule: run KAI Scheduler smoke through kai-scheduler-smoke.sh
- serve: run NIM GPU serving smoke through nim-gpu-serving-smoke.sh
- observe: prove NVIDIA GPU Operator managed DCGM metrics and ubiquity info --ai
- validate: require ubiquity health --ai and record nvidia-ai-final-demo-passed
- fail closed: skipped by default and records evidence only after all stages pass
MSG
  exit 0
fi

if [[ "${UBIQUITY_RUN_NVIDIA_AI_FINAL_DEMO:-false}" != "true" ]]; then
  echo "Skipping NVIDIA AI platform final demo. Set UBIQUITY_RUN_NVIDIA_AI_FINAL_DEMO=true on a real GPU cluster to run."
  exit 0
fi

if ! command -v kubectl >/dev/null 2>&1; then
  echo "required command kubectl not found" >&2
  exit 1
fi

run_ubiquity() {
  # shellcheck disable=SC2086
  ${UBIQUITY_BIN} "$@"
}

stage() {
  printf '\n== %s ==\n' "$1"
}

stage provision
run_ubiquity ai-platform render --profile ai-production >/tmp/ubiquity-ai-platform-render.yaml
grep -q 'name: ai-platform-nvidia-gpu-operator' /tmp/ubiquity-ai-platform-render.yaml
grep -q 'name: ai-platform-nim-operator' /tmp/ubiquity-ai-platform-render.yaml
grep -q 'name: ai-platform-kai-scheduler' /tmp/ubiquity-ai-platform-render.yaml

stage reconcile
run_ubiquity ai-platform apply --profile ai-production --server-side
kubectl -n argocd get applications.argoproj.io -l app.kubernetes.io/part-of=ubiquity-ai-platform

stage schedule
UBIQUITY_RUN_KAI_SMOKE=true "$(dirname "$0")/kai-scheduler-smoke.sh"

stage serve
UBIQUITY_RUN_NIM_GPU_SMOKE=true "$(dirname "$0")/nim-gpu-serving-smoke.sh"

stage observe
kubectl -n gpu-operator get service nvidia-dcgm-exporter
kubectl get --raw /api/v1/namespaces/gpu-operator/services/nvidia-dcgm-exporter:9400/proxy/metrics | grep -i dcgm >/dev/null
run_ubiquity info --ai

stage validate
UBIQUITY_RUN_NVIDIA_RDMA_SMOKE=true "$(dirname "$0")/nvidia-rdma-smoke.sh"
run_ubiquity health --ai

kubectl -n "${EVIDENCE_NAMESPACE}" create configmap "${EVIDENCE_CONFIGMAP}" \
  --from-literal=provision=passed \
  --from-literal=reconcile=passed \
  --from-literal=schedule=passed \
  --from-literal=serve=passed \
  --from-literal=observe=passed \
  --from-literal=validate=passed \
  --from-literal=timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --dry-run=client -o yaml | kubectl apply --server-side -f -

echo "NVIDIA AI platform final demo passed and recorded ${EVIDENCE_NAMESPACE}/${EVIDENCE_CONFIGMAP}"
