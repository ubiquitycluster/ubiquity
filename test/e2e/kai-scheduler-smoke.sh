#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${UBIQUITY_KAI_SMOKE_NAMESPACE:-kai-smoke}"
QUEUE_NAME="${UBIQUITY_KAI_SMOKE_QUEUE:-default-queue}"
SMOKE_POD="${UBIQUITY_KAI_SMOKE_POD:-kai-scheduling-smoke}"
EVIDENCE_NAMESPACE="${UBIQUITY_KAI_EVIDENCE_NAMESPACE:-kai-scheduler}"
EVIDENCE_CONFIGMAP="kai-scheduling-smoke-test-passed"
TIMEOUT="${UBIQUITY_KAI_SMOKE_TIMEOUT:-180s}"

if [[ "${1:-}" == "--dry-run" ]]; then
  cat <<'MSG'
KAI Scheduler smoke dry-run:
- gated by UBIQUITY_RUN_KAI_SMOKE=true
- validates queues.scheduling.run.ai queue evidence
- applies smoke resources with kubectl apply --server-side
- waits for a scheduled/completed smoke pod with kubectl wait
- records kai-scheduling-smoke-test-passed only after live scheduling evidence
- fail closed: render/apply success alone is not KAI scheduling proof
MSG
  exit 0
fi

if [[ "${UBIQUITY_RUN_KAI_SMOKE:-false}" != "true" ]]; then
  echo "Skipping KAI Scheduler smoke; set UBIQUITY_RUN_KAI_SMOKE=true to run live proof." >&2
  exit 0
fi

for bin in kubectl; do
  if ! command -v "${bin}" >/dev/null 2>&1; then
    echo "required command ${bin} not found" >&2
    exit 1
  fi
done

if ! kubectl get crd queues.scheduling.run.ai >/dev/null; then
  echo "queues.scheduling.run.ai CRD is missing; fail closed for KAI scheduling readiness" >&2
  exit 1
fi

if ! kubectl -n "${EVIDENCE_NAMESPACE}" get deploy kai-operator kai-scheduler-default binder admission pod-grouper podgroup-controller queue-controller >/dev/null; then
  echo "KAI Scheduler controller deployment evidence is missing; fail closed" >&2
  exit 1
fi

kubectl apply --server-side -f - <<YAML
apiVersion: v1
kind: Namespace
metadata:
  name: ${NAMESPACE}
---
apiVersion: scheduling.run.ai/v1
kind: Queue
metadata:
  name: ${QUEUE_NAME}
spec: {}
---
apiVersion: v1
kind: Pod
metadata:
  name: ${SMOKE_POD}
  namespace: ${NAMESPACE}
  labels:
    app.kubernetes.io/name: kai-scheduling-smoke
  annotations:
    kueue.x-k8s.io/queue-name: ${QUEUE_NAME}
spec:
  restartPolicy: Never
  schedulerName: kai-scheduler
  containers:
    - name: smoke
      image: busybox:1.36
      command: ["/bin/sh", "-c", "echo kai scheduled && sleep 2"]
      resources:
        requests:
          cpu: 10m
          memory: 16Mi
YAML

kubectl get queues.scheduling.run.ai "${QUEUE_NAME}" >/dev/null
kubectl -n "${NAMESPACE}" wait --for=condition=PodScheduled "pod/${SMOKE_POD}" --timeout="${TIMEOUT}"
kubectl -n "${NAMESPACE}" wait --for=condition=Ready "pod/${SMOKE_POD}" --timeout="${TIMEOUT}" || true
kubectl -n "${NAMESPACE}" wait --for=jsonpath='{.status.phase}'=Succeeded "pod/${SMOKE_POD}" --timeout="${TIMEOUT}"

kubectl -n "${EVIDENCE_NAMESPACE}" create configmap "${EVIDENCE_CONFIGMAP}" \
  --from-literal=queue="${QUEUE_NAME}" \
  --from-literal=namespace="${NAMESPACE}" \
  --from-literal=pod="${SMOKE_POD}" \
  --from-literal=timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --dry-run=client -o yaml | kubectl apply --server-side -f -

echo "KAI Scheduler smoke passed and recorded ${EVIDENCE_NAMESPACE}/${EVIDENCE_CONFIGMAP}"
