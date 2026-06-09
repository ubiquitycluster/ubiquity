# Ubiquity cloud production readiness audit checklist

Use this runbook before marking any cloud primitive, tenant service, VM, or backup policy production-ready. Render/apply proof is intent only; readiness must be backed by live evidence and must not use object existence as readiness.

## Required evidence

- Helm schema validation and server-side dry-run for the relevant primitive.
- Operator install-plan contract, including controller ownership and expected air-gap artifacts.
- Required CRDs and present CRDs captured before service intent is treated as reconcilable.
- `ubiquity cloud collect-readiness` evidence evaluated by `ubiquity cloud readiness --readiness-file <file>` with `ready: true`.
- Persistent services proven by a restore drill; rendered Restore objects are not restore proof.
- KubeVirt image catalog, standalone disk attachments, CDI import, guest boot, and guest health reviewed separately.
- Air-gap artifacts mirrored and checksummed for every referenced operator/chart/image.

## Commands

```sh
ubiquity cloud audit-checklist
ubiquity cloud render prerequisites
ubiquity cloud render operator-bundles
test/e2e/cloud-primitives-server-dry-run.sh
ubiquity cloud collect-readiness > /tmp/cloud-readiness-evidence.json
ubiquity cloud readiness --readiness-file /tmp/cloud-readiness-evidence.json
ubiquity cloud render restore-drill
ubiquity virtual-machines image-catalog
```

## Fail-closed boundaries

- A successful render is not readiness.
- A successful apply is not readiness.
- Kubernetes object existence is not readiness.
- A KubeVirt image catalog entry does not prove CDI import or guest boot.
- A backup Schedule or Restore object does not prove recoverability without a completed restore drill and smoke test.

## Cloud readiness proof bundle

Use `test/e2e/cloud-readiness-proof-bundle.sh` as the reviewer-facing collection path. The bundle must preserve:

- prerequisite contract
- operator provenance
- server-side dry-run output
- collected readiness JSON
- readiness report
- restore-drill evidence

A rendered Restore object is not recoverability proof. The audit is incomplete until restore-drill evidence includes controller success, readable restored data, and the `cloud-restore-drill-smoke-passed` marker.
