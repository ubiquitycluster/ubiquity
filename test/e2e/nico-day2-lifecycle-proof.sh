#!/usr/bin/env bash
set -euo pipefail

RUN_FLAG="${UBIQUITY_RUN_NICO_DAY2:-false}"
NODE="${UBIQUITY_NICO_DAY2_NODE:-nico-virtual-node-1}"
IMAGE="${UBIQUITY_NICO_DAY2_IMAGE:-ubuntu-24.04-gpu}"
NAMESPACE="${UBIQUITY_NICO_DAY2_NAMESPACE:-ubiquity-system}"
EVIDENCE_CONFIGMAP="nico-day2-lifecycle-proof-passed"
UBIQUITY_BIN="${UBIQUITY_BIN:-go run ./cmd/ubiquity}"

usage() {
  cat <<'USAGE'
Usage: nico-day2-lifecycle-proof.sh [--dry-run]

Gated live NICo/NVIDIA Infra Controller day-2 proof.
Runs only when UBIQUITY_RUN_NICO_DAY2=true unless --dry-run is used.

Required live evidence:
  - ubiquity health --nico
  - ubiquity nodes enroll
  - ubiquity nodes inspect
  - ubiquity nodes image
  - ubiquity nodes reboot
  - ubiquity nodes cordon
  - ubiquity nodes drain
  - ubiquity nodes maintenance
  - ubiquity nodes status reconcile
  - long-term status fields: bmcStatus kubeletStatus gpuStatus rdmaStatus firmwareStatus imageStatus maintenanceState
  - Kubernetes evidence marker: nico-day2-lifecycle-proof-passed
USAGE
}

DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
  DRY_RUN=true
elif [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
  usage
  exit 0
fi

run() {
  if [[ "$DRY_RUN" == "true" ]]; then
    printf '[dry-run] %s\n' "$*"
  else
    eval "$@"
  fi
}

if [[ "$DRY_RUN" != "true" && "$RUN_FLAG" != "true" ]]; then
  echo "Skipping NICo day-2 lifecycle proof; set UBIQUITY_RUN_NICO_DAY2=true to run live validation."
  exit 0
fi

run "$UBIQUITY_BIN health --nico"
run "$UBIQUITY_BIN nodes enroll '$NODE' --os '$IMAGE' --output json"
run "$UBIQUITY_BIN nodes inspect '$NODE' --output json"
run "$UBIQUITY_BIN nodes image '$IMAGE' --output json"
run "$UBIQUITY_BIN nodes cordon '$NODE' --confirm '$NODE' --output json"
run "$UBIQUITY_BIN nodes drain '$NODE' --confirm '$NODE' --drain-confirmed --output json"
run "$UBIQUITY_BIN nodes reboot '$NODE' --confirm '$NODE' --drain-confirmed --reason 'NICo day-2 proof reboot gate' --output json"
run "$UBIQUITY_BIN nodes maintenance '$NODE' --confirm '$NODE' --drain-confirmed --reason 'NICo day-2 proof maintenance gate' --output json"
run "$UBIQUITY_BIN nodes status reconcile '$NODE' --output json | tee /tmp/nico-day2-status.json"

if [[ "$DRY_RUN" == "true" ]]; then
  printf '[dry-run] verify status fields: bmcStatus kubeletStatus gpuStatus rdmaStatus firmwareStatus imageStatus maintenanceState\n'
  printf '[dry-run] kubectl -n %s create configmap %s --from-literal=node=%s --dry-run=client -o yaml | kubectl apply -f -\n' "$NAMESPACE" "$EVIDENCE_CONFIGMAP" "$NODE"
  exit 0
fi

for field in bmcStatus kubeletStatus gpuStatus rdmaStatus firmwareStatus imageStatus maintenanceState; do
  if ! grep -q "\"$field\"" /tmp/nico-day2-status.json; then
    echo "missing required status field: $field" >&2
    exit 1
  fi
done

kubectl get namespace "$NAMESPACE" >/dev/null 2>&1 || kubectl create namespace "$NAMESPACE"
kubectl -n "$NAMESPACE" create configmap "$EVIDENCE_CONFIGMAP" \
  --from-literal=node="$NODE" \
  --from-literal=image="$IMAGE" \
  --from-literal=operations="enroll,inspect,image,reboot,cordon,drain,maintenance,status-reconcile" \
  --dry-run=client -o yaml | kubectl apply -f -
