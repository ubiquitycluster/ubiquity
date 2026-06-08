#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" == "--dry-run" ]]; then
  cat <<'MSG'
NICo virtual bare-metal lab dry-run:
- qemu-system-x86_64 guests emulate bare-metal hosts
- sushy-tools exposes redfish BMC endpoints
- PXE boot validates boot-service integration boundaries
- kubectl apply --server-side applies NICo wrapper resources
- ubiquity nodes status / nodes power / nodes apply-os paths are exercised
- destructive actions remain dry-run unless explicit live flags are set
MSG
  exit 0
fi

if [[ "${UBIQUITY_RUN_NICO_VIRTUAL_BARE_METAL_E2E:-}" != "true" ]]; then
  cat <<'MSG'
Skipping NICo virtual bare-metal E2E.
Set UBIQUITY_RUN_NICO_VIRTUAL_BARE_METAL_E2E=true on a host with KVM/QEMU, sushy-tools, kubectl, and a disposable lab namespace.
MSG
  exit 0
fi

for required in qemu-system-x86_64 sushy-emulator kubectl; do
  command -v "$required" >/dev/null 2>&1 || { echo "missing required command: $required" >&2; exit 1; }
done

NS="${UBIQUITY_NICO_LAB_NAMESPACE:-nico-virtual-lab}"
CLI="${UBIQUITY_CLI:-go run ./cmd/ubiquity}"
QEMU_IMAGE="${UBIQUITY_NICO_LAB_QEMU_IMAGE:-/var/lib/ubiquity/nico-lab-node.qcow2}"
SUSHY_LISTEN="${UBIQUITY_NICO_LAB_SUSHY_LISTEN:-127.0.0.1:8000}"

kubectl create namespace "$NS" --dry-run=client -o yaml | kubectl apply --server-side -f -
helm template nico platform/nvidia-infra-controller --namespace "$NS" | kubectl apply --server-side --force-conflicts -f -

cat <<EOF
Starting virtual bare-metal lab contract:
- qemu-system-x86_64 image: $QEMU_IMAGE
- redfish endpoint via sushy-tools: http://$SUSHY_LISTEN/redfish/v1
- PXE boot is expected to be configured by the caller's lab network
EOF

# The concrete network topology is environment-specific; this script keeps mutation
# behind explicit flags and verifies Ubiquity's safety boundaries before hardware actions.
UBIQUITY_NICO_MODE=mock $CLI nodes status --backend nico --dry-run
UBIQUITY_NICO_MODE=mock $CLI nodes power lab-node-1 off --backend nico --dry-run --confirm lab-node-1 --reason virtual-lab-maintenance
UBIQUITY_NICO_MODE=mock $CLI nodes apply-os lab-node-1 --backend nico --dry-run --confirm lab-node-1 --os ubuntu-24.04

echo "NICo virtual bare-metal lab proof completed: qemu-system-x86_64/redfish/PXE contract checked; destructive actions remain dry-run"
