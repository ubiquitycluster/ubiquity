# cert-manager

## Scope

This runbook covers day-2 operation of cert-manager in a Ubiquity cluster: health
checks, certificate renewal failures, issuer troubleshooting, recovery, and the
evidence to capture before escalation. It does not claim that certificates are
valid in a live cluster unless the commands below were run against that cluster
and the evidence was recorded.

## Health checks

Run these checks before changing issuers or certificate resources:

```bash
kubectl -n cert-manager get deploy,pods
kubectl get clusterissuer,issuer --all-namespaces
kubectl get certificates --all-namespaces
kubectl get certificaterequest,order,challenge --all-namespaces
```

A healthy baseline has all cert-manager Deployments available, no CrashLooping
pods, ready issuers, and `Certificate` objects with `READY=True`.

Inspect controller logs when readiness is unclear:

```bash
kubectl -n cert-manager logs deploy/cert-manager --tail=200
kubectl -n cert-manager logs deploy/cert-manager-webhook --tail=200
kubectl -n cert-manager logs deploy/cert-manager-cainjector --tail=200
```

## Certificate renewal and failure response

1. Identify failing certificates:

   ```bash
   kubectl get certificates --all-namespaces
   kubectl describe certificate -n <namespace> <certificate>
   ```

2. Follow the request chain:

   ```bash
   kubectl describe certificaterequest -n <namespace> <request>
   kubectl describe order -n <namespace> <order>
   kubectl describe challenge -n <namespace> <challenge>
   ```

3. Confirm issuer state and backing credentials:

   ```bash
   kubectl describe clusterissuer <issuer>
   kubectl describe issuer -n <namespace> <issuer>
   kubectl get secret -n cert-manager
   ```

4. For DNS-01 failures, verify DNS provider credentials, zone delegation, and
   TXT propagation. For HTTP-01 failures, verify ingress class, service routing,
   and external load balancer reachability.

5. If the problem is a stale request, delete only the stuck subordinate object
   after capturing evidence. cert-manager should recreate the next request in
   the chain:

   ```bash
   kubectl delete certificaterequest -n <namespace> <request>
   ```

Do not delete the serving `Secret` unless rollback impact is understood. Removing
a serving secret can interrupt workloads before cert-manager recreates it.

## Recovery procedure

Use the least destructive recovery path first:

1. Restart cert-manager controllers only after collecting logs:

   ```bash
   kubectl -n cert-manager rollout restart deploy/cert-manager deploy/cert-manager-webhook deploy/cert-manager-cainjector
   kubectl -n cert-manager rollout status deploy/cert-manager
   ```

2. Re-apply known-good issuer manifests from GitOps if drift is suspected:

   ```bash
   kubectl diff -f <issuer-manifest>.yaml
   kubectl apply --server-side --dry-run=server -f <issuer-manifest>.yaml
   kubectl apply --server-side -f <issuer-manifest>.yaml
   ```

3. If webhook admission is broken, verify the webhook service endpoints and CA
   bundle injection before disabling policies:

   ```bash
   kubectl -n cert-manager get svc,endpoints cert-manager-webhook
   kubectl get validatingwebhookconfiguration cert-manager-webhook -o yaml
   kubectl get mutatingwebhookconfiguration cert-manager-webhook -o yaml
   ```

4. After recovery, confirm a renewal path with a non-production test certificate
   before declaring the incident closed.

## Escalation and evidence

Attach this evidence to the incident or change record:

```bash
kubectl -n cert-manager get pods -o wide
kubectl get certificates,certificaterequest,order,challenge --all-namespaces -o wide
kubectl describe certificate -n <namespace> <certificate>
kubectl describe certificaterequest -n <namespace> <request>
kubectl -n cert-manager logs deploy/cert-manager --tail=500
```

Record:

- affected namespaces and hostnames
- issuer names and provider type
- whether serving secrets were changed or deleted
- exact remediation commands run
- post-recovery certificate `READY` state and expiry time
