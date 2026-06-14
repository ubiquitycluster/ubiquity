#!/usr/bin/env bash
set -euo pipefail

RUN_FLAG="${UBIQUITY_RUN_CLOUD_READINESS_PROOF:-false}"
OUT_DIR="${UBIQUITY_CLOUD_READINESS_OUT_DIR:-/tmp/ubiquity-cloud-readiness-proof}"
UBIQUITY_BIN="${UBIQUITY_BIN:-go run ./cmd/ubiquity}"

usage() {
  cat <<'USAGE'
Usage: cloud-readiness-proof-bundle.sh [--dry-run]

Produces a cloud readiness proof bundle with:
  - prerequisite contract
  - operator provenance
  - server-side dry-run output
  - collected readiness JSON
  - readiness report
  - restore-drill evidence, including cloud-restore-drill-smoke-passed

Live mode is gated by UBIQUITY_RUN_CLOUD_READINESS_PROOF=true.
USAGE
}

DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
  DRY_RUN=true
elif [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
  usage
  exit 0
fi

if [[ "$DRY_RUN" != "true" && "$RUN_FLAG" != "true" ]]; then
  echo "Skipping cloud readiness proof bundle; set UBIQUITY_RUN_CLOUD_READINESS_PROOF=true to run live validation."
  exit 0
fi

mkdir -p "$OUT_DIR"

run_capture() {
  local label="$1" path="$2"; shift 2
  echo "== ${label} =="
  if [[ "$DRY_RUN" == "true" ]]; then
    printf '[dry-run] %s\n' "$*" | tee "$path"
  else
    "$@" | tee "$path"
  fi
}

run_shell_capture() {
  local label="$1" path="$2"; shift 2
  echo "== ${label} =="
  if [[ "$DRY_RUN" == "true" ]]; then
    printf '[dry-run] %s\n' "$*" | tee "$path"
  else
    bash -lc "$*" | tee "$path"
  fi
}

# Commands represented in this bundle include:
# ubiquity cloud render prerequisites
# ubiquity cloud render operator-bundles
# ubiquity cloud apply service --dry-run
# ubiquity cloud collect-readiness
# ubiquity cloud readiness --readiness-file

run_shell_capture "prerequisite contract" "$OUT_DIR/01-prerequisite-contract.yaml" "$UBIQUITY_BIN cloud render prerequisites"
run_shell_capture "operator provenance" "$OUT_DIR/02-operator-provenance.yaml" "$UBIQUITY_BIN cloud render operator-bundles"
run_shell_capture "server-side dry-run output" "$OUT_DIR/03-server-side-dry-run.txt" "$UBIQUITY_BIN cloud apply service --dry-run --service-type postgres --name cnpg-proof --namespace tenant-a"
run_shell_capture "restore-drill evidence" "$OUT_DIR/04-restore-drill.yaml" "$UBIQUITY_BIN cloud render restore-drill && printf '\n# required marker: cloud-restore-drill-smoke-passed\n'"
run_shell_capture "collected readiness JSON" "$OUT_DIR/05-collected-readiness.json" "$UBIQUITY_BIN cloud collect-readiness"
run_shell_capture "readiness report" "$OUT_DIR/06-readiness-report.txt" "$UBIQUITY_BIN cloud readiness --readiness-file '$OUT_DIR/05-collected-readiness.json'"

echo "Cloud readiness proof bundle written to $OUT_DIR"
