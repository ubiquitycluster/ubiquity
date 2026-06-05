#!/usr/bin/env bash
set -euo pipefail

if [ "${UBIQUITY_RUN_NICO_MOCK_E2E:-}" != "true" ]; then
  echo "Skipping NICo mock node lifecycle E2E. Set UBIQUITY_RUN_NICO_MOCK_E2E=true to run against a mock NICo API/CLI target."
  exit 0
fi

NICOCTL_BIN="${NICOCTL_BIN:-nicoctl}"
NICO_NAMESPACE="${NICO_NAMESPACE:-nico-system}"
NICO_SITE="${NICO_SITE:-mock-site}"
NICO_MACHINE="${NICO_MACHINE:-mock-machine-01}"
NICO_OS_IMAGE="${NICO_OS_IMAGE:-mock-os-image}"
NICO_TASK_TIMEOUT="${NICO_TASK_TIMEOUT:-15m}"

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Required command not found: $1" >&2
    exit 127
  }
}

run_nico() {
  "$NICOCTL_BIN" "$@"
}

require_cmd "$NICOCTL_BIN"

if command -v kubectl >/dev/null 2>&1; then
  kubectl -n "$NICO_NAMESPACE" get pods,svc --ignore-not-found || true
else
  echo "kubectl not found; continuing with NICo CLI-only mock lifecycle."
fi

echo "Checking mock site visibility: ${NICO_SITE}"
run_nico site get "$NICO_SITE" --output json

echo "Refreshing mock Machine inventory for site: ${NICO_SITE}"
run_nico machine discover --site "$NICO_SITE" --output json
run_nico machine get "$NICO_MACHINE" --output json

echo "Assigning mock Operating System image ${NICO_OS_IMAGE} to ${NICO_MACHINE}"
run_nico machine assign-os "$NICO_MACHINE" --os-image "$NICO_OS_IMAGE"

echo "Creating mock install Task"
run_nico task create install --machine "$NICO_MACHINE" --output json
run_nico task wait --machine "$NICO_MACHINE" --for condition=Succeeded --timeout "$NICO_TASK_TIMEOUT"
run_nico machine get "$NICO_MACHINE" --output json
run_nico instance get --machine "$NICO_MACHINE" --output json || true
run_nico machine gpu-stats "$NICO_MACHINE" --output json || true

echo "Creating mock reinstall Task"
run_nico task create reinstall --machine "$NICO_MACHINE" --output json
run_nico task wait --machine "$NICO_MACHINE" --for condition=Succeeded --timeout "$NICO_TASK_TIMEOUT"

if [ "${NICO_MOCK_DEPROVISION:-false}" = "true" ]; then
  echo "Creating mock deprovision Task"
  run_nico task create deprovision --machine "$NICO_MACHINE" --output json
  run_nico task wait --machine "$NICO_MACHINE" --for condition=Succeeded --timeout "$NICO_TASK_TIMEOUT"
else
  echo "Skipping mock deprovision. Set NICO_MOCK_DEPROVISION=true to include it."
fi

echo "NICo mock node lifecycle E2E completed. This is mock validation only and is not a certification claim."
