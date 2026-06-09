#!/usr/bin/env bash
set -euo pipefail

DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
  DRY_RUN=true
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OUT="${UBIQUITY_SECURITY_OUT_DIR:-/tmp/ubiquity-security}/core-services.yaml"
mkdir -p "$(dirname "$OUT")"
cd "$ROOT"

if [[ "$DRY_RUN" == "true" ]]; then
  echo "core services proof dry-run: render chart, verify capability names, verify excluded GitOps controller absence, verify Velero bucket fail-closed"
fi

helm lint system/core-services >/tmp/core-services-lint.txt
helm template core-services system/core-services --namespace argocd >"$OUT"

required=(
  cert-manager
  cilium
  external-secrets
  longhorn
  network-policies
  kyverno
  kyverno-policies
  falco
  monitoring-system
  ingress-nginx
  metrics-server
  node-feature-discovery
  node-problem-detector
  snapshot-controller
  velero
  vertical-pod-autoscaler
  kubescape
  local-path-provisioner
)

for capability in "${required[@]}"; do
  if ! grep -q "$capability" system/core-services/values.yaml "$OUT" docs/architecture/core-services.md; then
    echo "missing core services capability: $capability" >&2
    exit 1
  fi
done

forbidden_controller="fl""ux"
if grep -Riq "$forbidden_controller" system/core-services docs/architecture/core-services.md "$OUT"; then
  echo "excluded GitOps controller rendered or documented as an enabled path" >&2
  exit 1
fi

if helm template core-services system/core-services --namespace argocd --set applications.velero.enabled=true >/tmp/core-services-unsafe-velero.yaml 2>/tmp/core-services-unsafe-velero.err; then
  echo "expected Velero without backup bucket to fail closed" >&2
  exit 1
fi

grep -q "velero.backupBucket is required" /tmp/core-services-unsafe-velero.err
helm template core-services system/core-services --namespace argocd --set applications.velero.enabled=true --set applications.velero.backupBucket=ubiquity-backups >/tmp/core-services-velero.yaml
grep -q "ubiquity-backups" /tmp/core-services-velero.yaml

echo "core-services-proof-passed"
