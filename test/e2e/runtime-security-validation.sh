#!/usr/bin/env bash
set -euo pipefail

DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
  DRY_RUN=true
fi

OUT_DIR="${UBIQUITY_SECURITY_OUT_DIR:-/tmp/ubiquity-security}"
mkdir -p "$OUT_DIR"

if [[ "$DRY_RUN" == "true" ]]; then
  cat >"$OUT_DIR/runtime-security-validation.txt" <<'EOF'
[dry-run] helm lint system/falco-rules
[dry-run] helm template falco-rules system/falco-rules --namespace falco
[dry-run] kubectl -n falco get pods -l app.kubernetes.io/name=falco
[dry-run] kubectl -n falco get pods -l app.kubernetes.io/name=falcosidekick
[dry-run] verify Falco -> Falcosidekick -> Alertmanager alerting path
[dry-run] query Prometheus metric falco_events_total
[dry-run] validate Grafana dashboard contains Falco panels
EOF
  echo "runtime security validation dry-run wrote $OUT_DIR/runtime-security-validation.txt"
  exit 0
fi

if [[ "${UBIQUITY_RUN_RUNTIME_SECURITY_VALIDATION:-}" != "true" ]]; then
  echo "Skipping runtime security validation; set UBIQUITY_RUN_RUNTIME_SECURITY_VALIDATION=true for live Falco proof."
  exit 0
fi

helm lint system/falco-rules
helm template falco-rules system/falco-rules --namespace falco >/tmp/falco-rules.yaml
kubectl -n falco get pods -l app.kubernetes.io/name=falco
kubectl -n falco get pods -l app.kubernetes.io/name=falcosidekick
kubectl -n monitoring get prometheusrule,alertmanager 2>/dev/null || kubectl -A get alertmanager 2>/dev/null
kubectl -n monitoring exec deploy/prometheus -- wget -qO- 'http://localhost:9090/api/v1/query?query=falco_events_total' | tee "$OUT_DIR/falco_events_total.json"
grep -q 'Falco' grafana/dashboards/cluster.json
