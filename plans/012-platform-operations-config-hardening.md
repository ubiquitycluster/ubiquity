# Plan 012: Platform operations configuration hardening

> **Executor instructions**: This plan reconciles platform-service, backup/monitoring, documentation, and sandbox operational hardening already present in the working tree. Keep the changes scoped to deterministic config safety and sandbox/operator reliability.
>
> **Drift check (run first)**: `git diff --stat efd46ed..HEAD -- bootstrap/argocd docs/reference/helm-charts.md mkdocs.yml platform/gitea platform/harbor platform/onyxia system/k8up-operator system/longhorn-system`

## Status

- **Priority**: P2
- **Effort**: M
- **Risk**: MED
- **Depends on**: none
- **Category**: platform-ops/config-hardening/observability
- **Planned at**: commit `efd46ed`, 2026-06-15
- **Status**: DONE (implemented and verified)

## Why this matters

Several dirty files are operational hardening rather than core NICo/NCP implementation: sandbox Argo CD pullability, Git service metadata, secret escaping, Onyxia cleanup, K8up metrics wiring, Longhorn monitoring tests, and docs/reference navigation. These belong in an explicit platform-operations plan instead of remaining orphaned worktree noise.

## Reconciled dirty files

- `bootstrap/argocd/values-sandbox.yaml` — disables optional redis exporter/metrics paths that block fresh k3d sandbox pulls.
- `docs/reference/helm-charts.md`, `mkdocs.yml` — documentation/reference navigation updates.
- `platform/gitea/files/config/config.yaml`, `platform/gitea/files/config/main.go` — repository description handling and TODO cleanup.
- `platform/harbor/templates/harbor-config-overwrite-secret.yaml` — safer External Secrets templating/quoting for Harbor OIDC secret material.
- `platform/onyxia/values.yaml` — stale TODO/comment cleanup.
- `system/k8up-operator/kustomization.yaml`, `system/k8up-operator/schedule.yaml` — backup schedule Prometheus URL wiring and stale TODO cleanup.
- `system/longhorn-system/templates/servicemonitor.yaml`, `system/longhorn-system/tests/basic_test.yaml` — Longhorn monitoring rendering/test coverage.

## Acceptance evidence

- `helm unittest system/longhorn-system`
- `helm lint system/longhorn-system platform/harbor platform/gitea system/k8up-operator` where applicable
- `go test ./pkg/... ./cmd/... -count=1`
- `make test && make build`

## Done criteria

- [x] Sandbox Argo CD values avoid known non-pullable optional images.
- [x] Harbor OIDC secret templating safely quotes External Secrets data.
- [x] K8up schedule has concrete Prometheus URL wiring.
- [x] Longhorn monitoring resources are covered by Helm unit tests.
- [x] Platform docs/reference/navigation changes are reconciled with the workstream.
