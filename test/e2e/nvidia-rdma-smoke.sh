#!/usr/bin/env bash
set -euo pipefail

EVIDENCE_NAMESPACE="${UBIQUITY_RDMA_EVIDENCE_NAMESPACE:-gpu-operator}"
EVIDENCE_CONFIGMAP="rdma-network-smoke-test-passed"
RESOURCE_NAME="${UBIQUITY_RDMA_RESOURCE_NAME:-nvidia.com/rdma}"
NETWORK_ATTACHMENT_REGEX="${UBIQUITY_RDMA_NETWORK_ATTACHMENT_REGEX:-rdma|ipoib}"

if [[ "${1:-}" == "--dry-run" ]]; then
  cat <<'MSG'
NVIDIA RDMA smoke dry-run:
- gated by UBIQUITY_RUN_NVIDIA_RDMA_SMOKE=true
- validates nvidia.com/rdma allocatable resource evidence
- validates network-attachment-definitions.k8s.cni.cncf.io evidence
- records rdma-network-smoke-test-passed only after live evidence
- uses kubectl apply --server-side for evidence marker
- fail closed: NVIDIA Network Operator install alone is not RDMA readiness
MSG
  exit 0
fi

if [[ "${UBIQUITY_RUN_NVIDIA_RDMA_SMOKE:-false}" != "true" ]]; then
  echo "Skipping NVIDIA RDMA smoke. Set UBIQUITY_RUN_NVIDIA_RDMA_SMOKE=true on an RDMA-capable GPU cluster to run."
  exit 0
fi

if ! command -v kubectl >/dev/null 2>&1; then
  echo "required command kubectl not found" >&2
  exit 1
fi

rdma_nodes=$(kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.status.allocatable.nvidia\.com/rdma}{"\n"}{end}' | awk '$2 ~ /^[1-9][0-9]*$/ {print $1":"$2}')
if [[ -z "${rdma_nodes}" ]]; then
  echo "no ${RESOURCE_NAME} allocatable evidence found; fail closed for RDMA readiness" >&2
  exit 1
fi

network_attachments=$(kubectl get network-attachment-definitions.k8s.cni.cncf.io -A)
echo "${network_attachments}" | grep -E "${NETWORK_ATTACHMENT_REGEX}" >/dev/null || {
  echo "no RDMA/IPOIB NetworkAttachmentDefinition evidence found; fail closed for RDMA readiness" >&2
  exit 1
}

kubectl -n "${EVIDENCE_NAMESPACE}" create configmap "${EVIDENCE_CONFIGMAP}" \
  --from-literal=resource="${RESOURCE_NAME}" \
  --from-literal=nodes="${rdma_nodes}" \
  --from-literal=networkAttachmentRegex="${NETWORK_ATTACHMENT_REGEX}" \
  --from-literal=timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --dry-run=client -o yaml | kubectl apply --server-side -f -

echo "NVIDIA RDMA smoke passed and recorded ${EVIDENCE_NAMESPACE}/${EVIDENCE_CONFIGMAP}"
