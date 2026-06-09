#!/usr/bin/env bash
set -euo pipefail

DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
  DRY_RUN=true
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

charts=()
while IFS= read -r chart; do
  charts+=("$(dirname "$chart")")
done < <(find system platform -name Chart.yaml -not -path '*/disabled/*' | sort)

missing=()
for chart in "${charts[@]}"; do
  if ! compgen -G "$chart/tests/*.yaml" >/dev/null; then
    missing+=("$chart")
  fi
  if [[ "$DRY_RUN" == "true" ]]; then
    echo "[dry-run] helm dependency list $chart"
    echo "[dry-run] helm dependency build $chart"
    echo "[dry-run] helm lint $chart"
    echo "[dry-run] helm template $(basename "$chart" | tr '[:upper:]_' '[:lower:]-') $chart --namespace helm-hardening"
    echo "[dry-run] helm unittest $chart"
  else
    helm dependency list "$chart" >/tmp/helm-dependency-list.txt || true
    helm dependency build "$chart" >/tmp/helm-dependency-build.txt
    helm lint "$chart"
    helm template "$(basename "$chart" | tr '[:upper:]_' '[:lower:]-')" "$chart" --namespace helm-hardening >/tmp/helm-hardening-template.yaml
    helm unittest "$chart"
  fi
done

if (( ${#missing[@]} > 0 )); then
  printf 'missing helm unittest coverage:\n' >&2
  printf ' - %s\n' "${missing[@]}" >&2
  exit 1
fi

echo "helm hardening checks passed for ${#charts[@]} active system/platform charts"
