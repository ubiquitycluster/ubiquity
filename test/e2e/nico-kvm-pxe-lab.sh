#!/usr/bin/env bash
set -euo pipefail

# Opt-in KVM/QEMU PXE validation harness for NVIDIA Infra Controller lifecycle.
# This starts a qemu-bmc/containerlab topology, checks Redfish/IPMI reachability,
# and exercises non-destructive Ubiquity/NICo node flows. Destructive power/reset
# checks require UBIQUITY_NICO_KVM_LAB_DESTRUCTIVE=1 and exact confirmation.
# Use --dry-run for CI-safe validation of qemu-bmc, Redfish, IPMI, PXE, and NICo command flow.

DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
  DRY_RUN=true
elif [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
  sed -n '1,24p' "$0"
  exit 0
fi

if [[ "$DRY_RUN" == "true" ]]; then
  echo "[dry-run] validate qemu-bmc/containerlab fixture, Redfish/IPMI endpoints, PXE path, and NICo day-2 composition"
  echo "[dry-run] containerlab deploy -t test/fixtures/nico-kvm-pxe/containerlab.yml"
  echo "[dry-run] wait_for_redfish 127.0.0.1 8443 admin [REDACTED]"
  echo "[dry-run] wait_for_ipmi 127.0.0.1 8623 admin [REDACTED]"
  echo "[dry-run] go run ./cmd/ubiquity nodes os apply rocky-9 --site kvm-lab -o json"
  echo "[dry-run] go run ./cmd/ubiquity nodes power qemu-node-01 reset --confirm qemu-node-01 --site kvm-lab -o json"
  echo "[dry-run] UBIQUITY_RUN_NICO_DAY2=true test/e2e/nico-day2-lifecycle-proof.sh --dry-run"
  exit 0
fi

if [[ "${UBIQUITY_NICO_KVM_LAB:-}" != "1" ]]; then
  echo "SKIP: set UBIQUITY_NICO_KVM_LAB=1 to run the NICo KVM/QEMU PXE lab"
  exit 0
fi

missing=()
for bin in docker qemu-system-x86_64 kubectl helm go curl; do
  if ! command -v "$bin" >/dev/null 2>&1; then
    missing+=("$bin")
  fi
done
if ! command -v containerlab >/dev/null 2>&1 && ! command -v clab >/dev/null 2>&1; then
  missing+=("containerlab-or-clab")
fi
if [[ "${UBIQUITY_NICO_KVM_CHECK_IPMI:-1}" == "1" ]] && ! command -v ipmitool >/dev/null 2>&1; then
  missing+=("ipmitool")
fi
if [[ ${#missing[@]} -gt 0 ]]; then
  echo "FAIL: missing prerequisites: ${missing[*]}" >&2
  exit 1
fi

if [[ ! -e /dev/kvm ]]; then
  echo "WARN: /dev/kvm is not present; qemu-bmc may fall back to slow TCG if supported" >&2
fi

if [[ -z "${UBIQUITY_NICO_BASE_URL:-}" ]]; then
  echo "FAIL: UBIQUITY_NICO_BASE_URL must point at the NICo REST endpoint for live validation" >&2
  exit 1
fi

if [[ -z "${UBIQUITY_NICO_TOKEN:-}" && -z "${UBIQUITY_NICO_TOKEN_COMMAND:-}" && "${UBIQUITY_NICO_ALLOW_UNAUTHENTICATED:-}" != "true" ]]; then
  echo "FAIL: provide UBIQUITY_NICO_TOKEN, UBIQUITY_NICO_TOKEN_COMMAND, or explicitly set UBIQUITY_NICO_ALLOW_UNAUTHENTICATED=true" >&2
  exit 1
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FIXTURE_DIR="$ROOT_DIR/test/fixtures/nico-kvm-pxe"
TOPOLOGY="$FIXTURE_DIR/containerlab.yml"

if [[ ! -f "$TOPOLOGY" ]]; then
  echo "FAIL: missing topology fixture: $TOPOLOGY" >&2
  exit 1
fi

CLAB_BIN="containerlab"
if ! command -v containerlab >/dev/null 2>&1; then
  CLAB_BIN="clab"
fi

cleanup() {
  if [[ "${UBIQUITY_NICO_KVM_LAB_KEEP:-}" != "1" ]]; then
    "$CLAB_BIN" destroy -t "$TOPOLOGY" --cleanup || true
  fi
}
trap cleanup EXIT

wait_for_redfish() {
  local host="$1" port="$2" user="$3" pass="$4"
  local url="https://${host}:${port}/redfish/v1"
  for _ in $(seq 1 60); do
    if curl -kfsS -u "${user}:${pass}" "$url" >/dev/null 2>&1; then
      echo "Redfish ready at $url"
      return 0
    fi
    sleep 2
  done
  echo "FAIL: Redfish did not become ready at $url" >&2
  return 1
}

wait_for_ipmi() {
  local host="$1" port="$2" user="$3" pass="$4"
  for _ in $(seq 1 60); do
    if ipmitool -I lanplus -H "$host" -p "$port" -U "$user" -P "$pass" chassis power status >/dev/null 2>&1; then
      echo "IPMI ready at ${host}:${port}"
      return 0
    fi
    sleep 2
  done
  echo "FAIL: IPMI did not become ready at ${host}:${port}" >&2
  return 1
}

"$CLAB_BIN" deploy -t "$TOPOLOGY"

NODE="${UBIQUITY_NICO_KVM_NODE:-qemu-node-01}"
SITE="${UBIQUITY_NICO_SITE:-kvm-lab}"
OS_IMAGE="${UBIQUITY_NICO_KVM_OS_IMAGE:-rocky-9}"
BMC_HOST="${UBIQUITY_NICO_KVM_BMC_HOST:-127.0.0.1}"
REDFISH_PORT="${UBIQUITY_NICO_KVM_REDFISH_PORT:-8443}"
IPMI_PORT="${UBIQUITY_NICO_KVM_IPMI_PORT:-8623}"
BMC_USER="${QEMU_BMC_USER:-admin}"
BMC_PASS="${QEMU_BMC_PASSWORD:-password}"

wait_for_redfish "$BMC_HOST" "$REDFISH_PORT" "$BMC_USER" "$BMC_PASS"
if [[ "${UBIQUITY_NICO_KVM_CHECK_IPMI:-1}" == "1" ]]; then
  wait_for_ipmi "$BMC_HOST" "$IPMI_PORT" "$BMC_USER" "$BMC_PASS"
fi

export UBIQUITY_NICO_MODE=live

go run "$ROOT_DIR/cmd/ubiquity" nodes list --site "$SITE" -o json
go run "$ROOT_DIR/cmd/ubiquity" nodes os apply "$OS_IMAGE" --site "$SITE" -o json
go run "$ROOT_DIR/cmd/ubiquity" nodes status "$NODE" --site "$SITE" -o json || true

if [[ "${UBIQUITY_NICO_KVM_LAB_DESTRUCTIVE:-}" == "1" ]]; then
  go run "$ROOT_DIR/cmd/ubiquity" nodes power "$NODE" reset --confirm "$NODE" --site "$SITE" -o json
else
  echo "SKIP: destructive reset skipped; set UBIQUITY_NICO_KVM_LAB_DESTRUCTIVE=1 to run nodes power with --confirm $NODE"
fi

echo "KVM/QEMU PXE lab preflight completed. Use the physical hardware gate for final GPU/RDMA acceptance."
UBIQUITY_RUN_NICO_DAY2=true "$ROOT_DIR/test/e2e/nico-day2-lifecycle-proof.sh" --dry-run
