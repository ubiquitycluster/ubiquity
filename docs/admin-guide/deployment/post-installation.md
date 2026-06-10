# Post-installation

## Credential ownership and rotation

Save bootstrap credentials in an approved password manager or Vault-backed
secret store immediately after installation. Do not commit raw secret values or
root passwords to Git.

Back up these operator-owned files and record their owner:

- `~/.ssh/id_ed25519` and `~/.ssh/id_ed25519.pub` for bootstrap SSH access
- `./metal/kubeconfig.yaml` for Kubernetes administration
- `~/.terraform.d/credentials.tfrc.json` for Terraform backend access
- `./external/terraform.tfvars` for external resource inputs
- inventory-specific BMC/IPMI credentials and any generated root-password file,
  if your deployment creates one

Rotation expectations:

- rotate bootstrap SSH and Terraform credentials after handoff to the operations
  team
- rotate Vault root/recovery material according to the Vault runbook
- rotate service credentials immediately after suspected exposure or operator
  departure
- record credential IDs, owners, and rotation dates, but never record secret
  values in tickets or documentation

## Admin credentials

Retrieve admin credentials through the supported scripts or Vault paths:

- ArgoCD:
    - Username: `admin`
    - Password: run `./scripts/argocd-admin-password`
- Vault:
    - Root token: run `./scripts/vault-root-token` only during bootstrap or
      break-glass operations, then store/rotate according to the Vault runbook
- Grafana:
    - Username: `admin`
    - Password: retrieve from Vault or External Secrets; do not rely on static
      chart defaults for production
- Gitea:
    - Username: `gitea_admin`
    - Password: retrieve from Vault

## Post-install health checks

Run these checks before handing the environment to users:

```sh
kubectl get nodes -o wide
kubectl get pods --all-namespaces
kubectl get applications -n argocd
kubectl get certificates --all-namespaces
kubectl get externalsecret --all-namespaces
ubiquity health --nico
```

All required workloads should be available, certificates should be ready, and
external secret reconciliation should be healthy. Object existence alone is not
readiness; inspect status conditions and controller logs when a result is
ambiguous.

## Backup and restore

Before production use, confirm that backup jobs target off-cluster storage and
that restore procedures are documented for Vault, Git repositories, Kubernetes
state, and any persistent workloads.

Minimum evidence:

- backup destination and retention policy recorded
- one non-production restore drill completed or explicitly scheduled
- Vault snapshot process verified through the Vault runbook
- GitOps repository remote and branch protection confirmed

## Upgrade and rollback

Use staged changes for upgrades:

1. Render or dry-run the change locally.
2. Run server-side dry-run where Kubernetes APIs are available.
3. Apply through GitOps or the documented production change path.
4. Watch controller reconciliation and application health.
5. Keep rollback manifests, chart versions, and restore points available until
   post-upgrade checks pass.

## Proof boundary

A dry-run/local proof validates command wiring, manifests, and deterministic
checks. Live production proof requires evidence from the actual cluster: node
readiness, controller status, certificates, external DNS/load balancer access,
backup target reachability, and successful application smoke tests.

## Next steps

- [User onboarding](../../user-guide/onboarding.md)
