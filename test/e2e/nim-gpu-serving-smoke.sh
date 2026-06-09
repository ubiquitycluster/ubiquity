#!/usr/bin/env bash
set -euo pipefail

NIM_NAMESPACE="${UBIQUITY_NIM_NAMESPACE:-nim-service}"
NIM_OPERATOR_NAMESPACE="${UBIQUITY_NIM_OPERATOR_NAMESPACE:-nim-operator}"
NIM_SERVICE="${UBIQUITY_NIM_SERVICE:-meta-llama-3-2-1b-instruct}"
ENDPOINT="${UBIQUITY_NIM_ENDPOINT:-http://${NIM_SERVICE}.${NIM_NAMESPACE}.svc.cluster.local:8000/v1/models}"
TIMEOUT="${UBIQUITY_NIM_SMOKE_TIMEOUT:-15m}"
EVIDENCE_CONFIGMAP="nim-smoke-test-passed"

if [[ "${1:-}" == "--dry-run" ]]; then
  cat <<'MSG'
NIM GPU serving smoke dry-run:
- gated by UBIQUITY_RUN_NIM_GPU_SMOKE=true
- requires a real GPU node and NIMService
- waits for NIMService readiness with kubectl wait
- uses curl --fail against the NIM endpoint
- records nim-smoke-test-passed only after endpoint success
- fail closed: NIM Operator installation alone is not serving readiness
MSG
  exit 0
fi

if [[ "${UBIQUITY_RUN_NIM_GPU_SMOKE:-false}" != "true" ]]; then
  echo "Skipping NIM GPU serving smoke. Set UBIQUITY_RUN_NIM_GPU_SMOKE=true on a real GPU cluster to run."
  exit 0
fi

for bin in kubectl curl; do
  if ! command -v "${bin}" >/dev/null 2>&1; then
    echo "required command ${bin} not found" >&2
    exit 1
  fi
done

kubectl get nodes -o json | grep -E 'nvidia.com/(gpu|mig-)' >/dev/null || {
  echo "no GPU or MIG allocatable evidence found; fail closed for NIM serving" >&2
  exit 1
}

kubectl -n "${NIM_OPERATOR_NAMESPACE}" get pods >/dev/null
kubectl -n "${NIM_NAMESPACE}" get nimservice "${NIM_SERVICE}" >/dev/null
kubectl -n "${NIM_NAMESPACE}" wait --for=condition=Ready "nimservice/${NIM_SERVICE}" --timeout="${TIMEOUT}"

curl --fail --silent --show-error "${ENDPOINT}" >/tmp/ubiquity-nim-smoke-response.json

kubectl -n "${NIM_OPERATOR_NAMESPACE}" create configmap "${EVIDENCE_CONFIGMAP}" \
  --from-literal=namespace="${NIM_NAMESPACE}" \
  --from-literal=nimService="${NIM_SERVICE}" \
  --from-literal=endpoint="${ENDPOINT}" \
  --from-literal=timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --dry-run=client -o yaml | kubectl apply --server-side -f -

echo "NIM GPU serving smoke passed and recorded ${NIM_OPERATOR_NAMESPACE}/${EVIDENCE_CONFIGMAP}"
