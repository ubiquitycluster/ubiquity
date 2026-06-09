#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" == "--dry-run" ]]; then
  cat <<'MSG'
network policy behavior dry-run: creates deny-client and allow-client, proves expected denied traffic to fail, applies allow-netpol-client-to-echo, then proves expected allowed traffic to succeed and emits network-policy-deny-allow-proof-passed.
MSG
  exit 0
fi

if [[ "${UBIQUITY_RUN_NETWORK_POLICY_E2E:-}" != "true" ]]; then
  cat <<'MSG'
Skipping network policy behavior E2E.
Set UBIQUITY_RUN_NETWORK_POLICY_E2E=true to create test namespaces/pods and prove deny/allow behavior.
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
  name: deny-client
  labels:
    app: deny-client
spec:
  restartPolicy: Never
  containers:
    - name: client
      image: busybox:1.36
      command: ["sleep", "3600"]
---
apiVersion: v1
kind: Pod
metadata:
  name: allow-client
  labels:
    app: allow-client
spec:
  restartPolicy: Never
  containers:
    - name: client
      image: busybox:1.36
      command: ["sleep", "3600"]
YAML

kubectl wait -n "$NS" --for=condition=Available deploy/echo-server --timeout=90s
kubectl wait -n "$NS" --for=condition=Ready pod/deny-client --timeout=90s
kubectl wait -n "$NS" --for=condition=Ready pod/allow-client --timeout=90s
kubectl wait -n "$NS" --for=jsonpath='{.metadata.name}' networkpolicy/default-deny-ingress --timeout=30s
kubectl wait -n "$NS" --for=jsonpath='{.metadata.name}' networkpolicy/default-deny-egress --timeout=30s
kubectl wait -n "$NS" --for=jsonpath='{.metadata.name}' networkpolicy/allow-dns --timeout=30s

kubectl exec -n "$NS" deny-client -- nslookup kubernetes.default.svc.cluster.local
if kubectl exec -n "$NS" deny-client -- wget --timeout=3 -qO- http://echo-server:8080; then
  echo "expected denied traffic to fail before allow policy" >&2
  exit 1
fi

kubectl apply -n "$NS" -f - <<'YAML'
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-netpol-client-to-echo
spec:
  podSelector:
    matchLabels:
      app: echo-server
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: allow-client
      ports:
        - protocol: TCP
          port: 8080
  policyTypes:
    - Ingress
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-client-egress-to-echo
spec:
  podSelector:
    matchLabels:
      app: allow-client
  egress:
    - to:
        - podSelector:
            matchLabels:
              app: echo-server
      ports:
        - protocol: TCP
          port: 8080
  policyTypes:
    - Egress
YAML

if ! kubectl exec -n "$NS" allow-client -- wget --timeout=10 -qO- http://echo-server:8080 | grep -q ok; then
  echo "expected allowed traffic to succeed after explicit allow policy" >&2
  exit 1
fi
if kubectl exec -n "$NS" deny-client -- wget --timeout=3 -qO- http://echo-server:8080; then
  echo "expected denied traffic to fail for deny-client after allow policy" >&2
  exit 1
fi

echo "network-policy-deny-allow-proof-passed"
