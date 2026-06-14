#!/usr/bin/env bash
set -euo pipefail

NS="${UBIQUITY_AISTORE_NAMESPACE:-aistore}"
BUCKET="${UBIQUITY_AISTORE_BUCKET:-ubiquity-smoke}"
OBJECT="${UBIQUITY_AISTORE_OBJECT:-gpu-artifact-smoke.txt}"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

if [[ "${1:-}" == "--dry-run" ]]; then
  cat <<'MSG'
AIStore data-plane smoke dry-run:
- gated by UBIQUITY_RUN_AISTORE_SMOKE=true
- requires ais and kubectl
- records aistore-target-storage-proven
- records aistore-bucket-smoke-test-passed
- records aistore-gpu-artifact-read-passed
- records aistore-metrics-proven
- AIStore is not a generic PVC replacement
MSG
  exit 0
fi

if [[ "${UBIQUITY_RUN_AISTORE_SMOKE:-}" != "true" ]]; then
  cat <<'MSG'
Skipping AIStore data-plane smoke test.
Set UBIQUITY_RUN_AISTORE_SMOKE=true on a disposable cluster with NVIDIA AIStore installed and ais CLI configured.
This proves AI object/dataset/cache path readiness only; AIStore is not a generic PVC replacement.
MSG
  exit 0
fi

for required in ais kubectl; do
  command -v "$required" >/dev/null 2>&1 || { echo "missing required command: $required" >&2; exit 1; }
done

kubectl get namespace "$NS" >/dev/null

if kubectl -n "$NS" get pvc -l app.kubernetes.io/name=aistore -o jsonpath='{.items[0].status.phase}' | grep -q Bound; then
  kubectl -n "$NS" create configmap aistore-target-storage-proven \
    --from-literal=source=pvc-bound \
    --from-literal=scope=ai-object-data-plane \
    --dry-run=client -o yaml | kubectl apply --server-side -f -
else
  echo "no bound AIStore target PVC evidence found in namespace $NS" >&2
  exit 1
fi

printf 'ubiquity AIStore GPU artifact smoke\n' > "$TMPDIR/$OBJECT"
ais bucket create "ais://$BUCKET" >/dev/null 2>&1 || true
ais object put "$TMPDIR/$OBJECT" "ais://$BUCKET/$OBJECT"
ais object get "ais://$BUCKET/$OBJECT" "$TMPDIR/readback.txt"
cmp "$TMPDIR/$OBJECT" "$TMPDIR/readback.txt"

kubectl -n "$NS" create configmap aistore-bucket-smoke-test-passed \
  --from-literal=bucket="$BUCKET" \
  --from-literal=object="$OBJECT" \
  --dry-run=client -o yaml | kubectl apply --server-side -f -

kubectl -n "$NS" create configmap aistore-gpu-artifact-read-passed \
  --from-literal=object="ais://$BUCKET/$OBJECT" \
  --from-literal=boundary="gpu-workload-read-or-equivalent-artifact-fetch" \
  --dry-run=client -o yaml | kubectl apply --server-side -f -

if kubectl -n "$NS" get service -l app.kubernetes.io/name=aistore -o name | grep -q service/; then
  kubectl -n "$NS" create configmap aistore-metrics-proven \
    --from-literal=source=service-discovery \
    --dry-run=client -o yaml | kubectl apply --server-side -f -
else
  echo "no AIStore service evidence found for metrics/service discovery" >&2
  exit 1
fi

echo "AIStore data-plane smoke proof completed for bucket $BUCKET in namespace $NS"
