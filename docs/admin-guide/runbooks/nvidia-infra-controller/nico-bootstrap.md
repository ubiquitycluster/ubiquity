# Runbook: NICo bootstrap

Status: experimental/preview. This runbook enables NVIDIA Infra Controller (NICo) after the Ubiquity management plane exists. It does not certify any hardware or software platform.

## Preconditions

- You have a maintenance/change record for enabling NICo.
- You know the target namespace, usually `nico-system` or `nvidia-infra-controller`.
- Vault or an equivalent secret store is available.
- External Secrets Operator can project approved runtime secrets.
- PostgreSQL ownership, backup, and restore procedures are documented.
- LoadBalancer, DNS, NTP, BMC, and provisioning networks are reachable.
- No secrets are stored in GitOps values.

## Procedure

1. Confirm prerequisites:

```sh
kubectl get nodes
kubectl get storageclass
kubectl get clusterissuer,issuer -A || true
kubectl get externalsecret,secretstore,clustersecretstore -A || true
```

2. Render or sync the prerequisite wrapper:

```sh
helm template nico-prereqs system/nvidia-infra-controller-prereqs
```

3. Confirm secret projection without printing secret data:

```sh
kubectl -n nico-system get externalsecret,secret --ignore-not-found
kubectl -n nvidia-infra-controller get externalsecret,secret --ignore-not-found
```

4. Deploy NICo Core through GitOps or a controlled Helm flow.

```sh
helm template nico-core system/nvidia-infra-controller-core
```

5. Deploy NICo REST and site-agent through GitOps or a controlled Helm flow.

```sh
helm template nico-rest platform/nvidia-infra-controller-rest
```

6. Wait for readiness:

```sh
kubectl -n nico-system rollout status deploy/nico-controller --timeout=10m || true
kubectl -n nico-system rollout status deploy/nico-rest --timeout=10m || true
kubectl -n nico-system rollout status deploy/site-agent --timeout=10m || true
kubectl -n nvidia-infra-controller get pods,svc --ignore-not-found
```

7. Record the bootstrap-to-NICo handoff:

```yaml
status: experimental-preview
bootstrapBoundaryCrossed: true
nicoNamespace: nico-system
sourceRepo: NVIDIA/infra-controller
certificationClaim: none
```

## Validation

- NICo workloads are ready.
- REST health endpoint responds if exposed.
- site-agent can see the intended site.
- No Kubernetes Secret manifests with literal secret data were committed.
- At least one non-production Machine is available for discovery/provisioning validation.

## Rollback

If bootstrap fails, stop GitOps sync for NICo wrappers, preserve logs, and remove only resources that were created by this change. Do not delete shared prerequisites such as cert-manager, Vault, External Secrets Operator, MetalLB, or PostgreSQL unless they were created solely for this test and the platform owner approves.
