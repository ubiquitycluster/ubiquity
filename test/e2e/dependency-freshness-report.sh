#!/usr/bin/env bash
set -euo pipefail

DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
  DRY_RUN=true
fi

OUT_DIR="${UBIQUITY_SECURITY_OUT_DIR:-/tmp/ubiquity-security}"
REPORT="$OUT_DIR/dependency-freshness-report.md"
mkdir -p "$OUT_DIR"

{
  echo "# Dependency freshness report"
  echo
  echo "Generated: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo
  echo "## Go modules"
  if [[ "$DRY_RUN" == "true" ]]; then
    echo '
```sh
go list -m -u all
```'
  else
    go list -m -u all || true
  fi
  echo
  echo "## Helm chart dependencies"
  echo '
```sh
helm dependency list <chart>
```'
  if [[ "$DRY_RUN" == "true" ]]; then
    find system platform -name Chart.yaml -not -path '*/disabled/*' | sort | sed 's#^#- helm dependency list #'
  else
    while IFS= read -r chart; do
      dir="$(dirname "$chart")"
      echo "### $dir"
      helm dependency list "$dir" || true
    done < <(find system platform -name Chart.yaml -not -path '*/disabled/*' | sort)
  fi
  echo
  echo "## GitHub actions"
  echo "actions: inspect .github/workflows for pinned versions and update cadence"
  grep -R "uses:" .github/workflows || true
  echo
  echo "## Container images"
  echo "container images: inspect Helm values, manifests, and Dockerfiles for image tags"
  grep -R "image:" system platform apps bootstrap 2>/dev/null | head -200 || true
  echo
  echo "## Automation"
  echo "dependabot: .github/dependabot.yml"
  echo "renovate: renovate.json"
} >"$REPORT"

echo "dependency freshness report written to $REPORT"
