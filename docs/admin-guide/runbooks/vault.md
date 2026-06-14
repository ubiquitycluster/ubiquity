# Vault

## Scope

This runbook covers day-2 operation of Vault in a Ubiquity cluster: health
checks, seal/unseal response, credential rotation, backup/restore, and escalation
evidence. It assumes Vault is managed through the repository's GitOps manifests
and that unseal keys/root tokens are stored outside the repository in the
approved operator-controlled secret store.

## Health checks

Start with Kubernetes state and Vault status. The examples use `kubectl exec`
through the Vault namespace to avoid relying on local Vault network access:

```bash
kubectl -n vault get pods,svc,statefulset,pvc
kubectl -n vault logs statefulset/vault --tail=200
kubectl -n vault exec vault-0 -- vault status
```

A healthy Vault reports initialized, unsealed, active or standby HA mode as
expected, and no CrashLooping pods. If `vault status` fails, check service DNS,
network policy, and pod readiness before attempting recovery.

Check dependent secret generation jobs and consumers:

```bash
kubectl -n vault get jobs,cronjobs
kubectl -n external-secrets get clustersecretstore,externalsecret --all-namespaces
kubectl get events --all-namespaces --field-selector involvedObject.kind=ExternalSecret
```

## Unseal and recovery

If Vault is sealed, recover with quorum-held unseal keys. Never paste unseal keys
into shell history or commit them to Git.

1. Confirm seal state:

   ```bash
   kubectl -n vault exec vault-0 -- vault status
   ```

2. Unseal each required replica using the approved break-glass terminal/session:

   ```bash
   kubectl -n vault exec -it vault-0 -- vault operator unseal
   kubectl -n vault exec -it vault-1 -- vault operator unseal
   kubectl -n vault exec -it vault-2 -- vault operator unseal
   ```

3. Verify HA state after quorum is restored:

   ```bash
   kubectl -n vault exec vault-0 -- vault status
   kubectl -n vault exec vault-0 -- vault operator raft list-peers
   ```

If a pod cannot rejoin Raft, collect logs and peer state before removing it from
the cluster. Peer removal is destructive and should be reviewed by an operator
who holds the recovery material.

## Credential rotation

Rotate credentials through Vault first, then allow External Secrets or consuming
controllers to reconcile.

1. Identify the secret path and consumers:

   ```bash
   vault kv metadata get <mount>/<path>
   kubectl get externalsecret --all-namespaces | grep <secret-name>
   ```

2. Write the new value from a secure shell, not from checked-in files:

   ```bash
   vault kv put <mount>/<path> <key>=<new-value>
   ```

3. Confirm reconciliation and consuming workload rollout:

   ```bash
   kubectl -n external-secrets logs deploy/external-secrets --tail=200
   kubectl get externalsecret --all-namespaces
   kubectl rollout status deploy/<consumer> -n <namespace>
   ```

4. Record the rotated path, consumer, operator, and validation evidence. Do not
   record the secret value.

## Backup and restore

Use Raft snapshots for Vault data-plane backup. Store snapshots in the approved
encrypted backup location and verify restore in a non-production environment
before relying on the artifact for disaster recovery.

Create a snapshot:

```bash
kubectl -n vault exec vault-0 -- vault operator raft snapshot save /tmp/vault.snap
kubectl -n vault cp vault-0:/tmp/vault.snap ./vault-$(date +%Y%m%d%H%M%S).snap
```

Verify snapshot handling:

```bash
sha256sum vault-*.snap
ls -lh vault-*.snap
```

Restore is disruptive. Run only during an approved incident window:

```bash
kubectl -n vault cp ./vault.snap vault-0:/tmp/vault.snap
kubectl -n vault exec vault-0 -- vault operator raft snapshot restore -force /tmp/vault.snap
kubectl -n vault exec vault-0 -- vault status
```

After restore, verify External Secrets reconciliation and at least one dependent
application rollout before closing the incident.

## Escalation and evidence

Attach this evidence to the incident or change record:

```bash
kubectl -n vault get pods -o wide
kubectl -n vault describe statefulset vault
kubectl -n vault logs statefulset/vault --tail=500
kubectl -n vault exec vault-0 -- vault status
kubectl -n vault exec vault-0 -- vault operator raft list-peers
kubectl -n external-secrets get clustersecretstore,externalsecret --all-namespaces
```

Record:

- seal state and HA leader before and after remediation
- whether unseal, peer removal, snapshot restore, or credential rotation occurred
- operators involved in quorum/recovery approval
- affected secret paths and consumers without secret values
- post-recovery reconciliation evidence
