#!/usr/bin/env bash
set -euo pipefail

DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
  DRY_RUN=true
elif [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
  cat <<'USAGE'
Usage: sbom-and-signing-proof.sh [--dry-run]

Generates SPDX and CycloneDX SBOMs with syft and signs/verifies an image with cosign/Sigstore.
Live signing is gated by UBIQUITY_RUN_IMAGE_SIGNING=true.
USAGE
  exit 0
fi

OUT_DIR="${UBIQUITY_SECURITY_OUT_DIR:-/tmp/ubiquity-security}"
IMAGE="${UBIQUITY_IMAGE_REF:-ghcr.io/ubiquitycluster/ubiquity:ci}"
mkdir -p "$OUT_DIR"

run_or_record() {
  local output="$1"; shift
  if [[ "$DRY_RUN" == "true" ]]; then
    printf '[dry-run] %s\n' "$*" | tee "$output" >/dev/null
  else
    "$@" | tee "$output" >/dev/null
  fi
}

if [[ "$DRY_RUN" == "true" ]]; then
  echo '{"SPDXID":"SPDXRef-DOCUMENT","name":"ubiquity-dry-run"}' >"$OUT_DIR/sbom.spdx.json"
  echo '{"bomFormat":"CycloneDX","specVersion":"1.5","metadata":{"component":{"name":"ubiquity-dry-run"}}}' >"$OUT_DIR/sbom.cyclonedx.json"
  printf '[dry-run] syft dir:. -o spdx-json=%s/sbom.spdx.json\n' "$OUT_DIR" >"$OUT_DIR/sbom-generation.log"
  printf '[dry-run] syft dir:. -o cyclonedx-json=%s/sbom.cyclonedx.json\n' "$OUT_DIR" >>"$OUT_DIR/sbom-generation.log"
  printf '[dry-run] cyclonedx validate --input-file %s/sbom.cyclonedx.json\n' "$OUT_DIR" >"$OUT_DIR/cyclonedx-validation.log"
  printf '[dry-run] cosign sign --yes %s\n[dry-run] cosign verify %s\n' "$IMAGE" "$IMAGE" >"$OUT_DIR/cosign.log"
  echo "SBOM/signing dry-run artifacts written to $OUT_DIR"
  exit 0
fi

if ! command -v syft >/dev/null 2>&1; then
  echo "syft is required for SBOM generation" >&2
  exit 2
fi
syft dir:. -o "spdx-json=$OUT_DIR/sbom.spdx.json"
syft dir:. -o "cyclonedx-json=$OUT_DIR/sbom.cyclonedx.json"

if command -v cyclonedx >/dev/null 2>&1; then
  cyclonedx validate --input-file "$OUT_DIR/sbom.cyclonedx.json" | tee "$OUT_DIR/cyclonedx-validation.log"
else
  echo "cyclonedx CLI unavailable; generated CycloneDX JSON at $OUT_DIR/sbom.cyclonedx.json" | tee "$OUT_DIR/cyclonedx-validation.log"
fi

if [[ "${UBIQUITY_RUN_IMAGE_SIGNING:-}" != "true" ]]; then
  echo "Skipping cosign signing; set UBIQUITY_RUN_IMAGE_SIGNING=true for live signing." | tee "$OUT_DIR/cosign.log"
  exit 0
fi
if ! command -v cosign >/dev/null 2>&1; then
  echo "cosign is required for Sigstore signing" >&2
  exit 3
fi
cosign sign --yes "$IMAGE" | tee "$OUT_DIR/cosign.log"
cosign verify "$IMAGE" | tee -a "$OUT_DIR/cosign.log"
