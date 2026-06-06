# Cloud readiness validation runbook

This runbook defines the minimum evidence required before a Ubiquity cloud primitive is called ready.

## Required order

1. Render the manifest or Helm chart.
2. Verify CRD presence for every rendered custom resource.
3. Run Kubernetes server-side dry-run with `kubectl apply --dry-run=server -f -`.
4. Apply through the intended GitOps or administrative path.
5. Check controller reconciliation using status condition fields and controller-specific health signals.
6. Run workload smoke tests.
7. Run a restore drill for persistent services before claiming durability.

## Fail-closed rules

- Object existence is not readiness.
- A successful render is not readiness.
- A successful apply is not readiness.
- Readiness is not object existence; use controller status condition evidence.
- Missing CRD presence, missing status condition, or unknown controller state must fail closed.
- Backup policies are incomplete until a restore drill has succeeded.
- VM disks are incomplete until CDI import/clone conditions and PVC binding prove success.
- Tenant networking is incomplete until a positive allowed-path test and a negative denied-path test both pass.

## Script

Use `test/e2e/cloud-primitives-server-dry-run.sh` against a CRD-enabled sandbox to prove schemas and API compatibility before live reconciliation testing.

## Readiness evidence evaluation

After collecting evidence, run:

```sh
ubiquity cloud readiness --readiness-file /tmp/cloud-readiness-evidence.json
```

To collect baseline CRD and controller condition evidence from a live cluster, run:

```sh
ubiquity cloud collect-readiness > /tmp/cloud-readiness-evidence.json
ubiquity cloud readiness --readiness-file /tmp/cloud-readiness-evidence.json
```

Use repeated `--readiness-resource <resource.api.group>` flags when validating a scoped primitive rather than the default cloud resource set.

`ubiquity cloud readiness` fails closed if a required CRD is missing, a resource has no status conditions, no positive Ready/Available/Bound/Reconciled/Succeeded condition exists, or a smoke test is false.
