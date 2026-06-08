#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" == "--dry-run" ]]; then
  echo "network policy behavior dry-run: validates default-deny-ingress default-deny-egress allow-dns command contract"
  exit 0
fi

if [[ "${UBIQUITY_RUN_NETWORK_POLICY_E2E:-}" != "true" ]]; then
  cat <<'MSG'
Skipping network policy behavior E2E.
Set UBIQUITY_RUN_NETWORK_POLICY_E2E=true to create test namespaces/pods and prove allow/deny behavior.
MSG
  exit 0
fi

NS="${UBIQUITY_NETWORK_POLICY_E2E_NAMESPACE:-ubiquity-netpol-e2e}"
trap 'kubectl delete namespace "$NS" --ignore-not-found=true >/dev/null 2>&1 || true' EXIT

kubectl create namespace "$NS"
helm upgrade --install ubiquity-network-policies system/network-policies --namespace "$NS" --set namespace="$NS"

kubectl apply -n "$NS" -f - <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: echo-server
spec:
  replicas: 1
  selector:
    matchLabels:
      app: echo-server
  template:
    metadata:
      labels:
        app: echo-server
    spec:
      containers:
        - name: http
          image: busybox:1.36
          command: ["sh", "-c", "while true; do { echo -e 'HTTP/1.1 200 OK\r\n\r\nok'; } | nc -l -p 8080; done"]
          ports:
            - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: echo-server
spec:
  selector:
    app: echo-server
  ports:
    - name: http
      port: 8080
      targetPort: 8080
---
apiVersion: v1
kind: Pod
metadata:
  name: netpol-client
spec:
  restartPolicy: Never
  containers:
    - name: client
      image: busybox:1.36
      command: ["sleep", "3600"]
YAML

kubectl wait -n "$NS" --for=condition=Available deploy/echo-server --timeout=90s
kubectl wait -n "$NS" --for=condition=Ready pod/netpol-client --timeout=90s
kubectl wait -n "$NS" --for=jsonpath='{.metadata.name}' networkpolicy/default-deny-ingress --timeout=30s
kubectl wait -n "$NS" --for=jsonpath='{.metadata.name}' networkpolicy/default-deny-egress --timeout=30s
kubectl wait -n "$NS" --for=jsonpath='{.metadata.name}' networkpolicy/allow-dns --timeout=30s

kubectl exec -n "$NS" netpol-client -- nslookup kubernetes.default.svc.cluster.local
if kubectl exec -n "$NS" netpol-client -- wget --timeout=3 -qO- http://echo-server:8080; then
  echo "expected default-deny-ingress/default-deny-egress to block HTTP traffic" >&2
  exit 1
fi

echo "network policy behavior proof passed: DNS allowed, HTTP denied by default-deny policies"
