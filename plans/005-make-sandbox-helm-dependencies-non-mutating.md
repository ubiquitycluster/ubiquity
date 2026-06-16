# Plan 005: Make sandbox Helm dependency rendering non-mutating

> **Executor instructions**: Follow this plan step by step. Add regression tests before implementation. Preserve unrelated dirty files. If cited code has drifted, compare live code and update this plan before proceeding.
>
> **Drift check (run first)**: `git diff --stat efd46ed..HEAD -- cmd/ubiquity/cmd/up.go cmd/ubiquity/cmd/up_test.go cmd/ubiquity/cmd/sandbox_deploy_test.go`

## Status

- **Priority**: P1
- **Effort**: M
- **Risk**: MED
- **Depends on**: none
- **Category**: correctness/dx
- **Planned at**: commit `efd46ed`, 2026-06-15
- **Status**: DONE (implemented and verified)

## Why this matters

Sandbox render/apply paths must be safe in a dirty developer checkout. Today they can run Helm dependency commands in source chart directories and delete `Chart.lock` / `charts/` artifacts, which makes validation non-reproducible and can destroy tracked dependency artifacts.

## Current state

- `cmd/ubiquity/cmd/up.go` renders sandbox targets by running `helm dependency build` directly against the source chart directory and defers cleanup:
  - `renderSandboxHelmTarget` around lines 531-543.
- `cleanupSandboxDependencyArchives` removes `charts/*.tgz`, `charts/`, and `Chart.lock` from source chart directories around lines 624-635.
- `runHelmTemplateAndApply` runs `helm dependency update` and ignores its error around lines 1008-1011.

## Commands you will need

| Purpose | Command | Expected on success |
|---|---|---|
| Focused tests | `go test ./cmd/ubiquity/cmd -run 'TestRenderSandboxHelmTargetDoesNotMutateSourceChart|TestRunHelmTemplateAndApplyReturnsDependencyFailure' -count=1` | exits 0 |
| Sandbox render proof | `go run ./cmd/ubiquity test --sandbox-deploy` | exits 0 |
| Full Go focused package | `go test ./cmd/ubiquity/cmd -count=1` | exits 0 |

## Scope

**In scope**:
- `cmd/ubiquity/cmd/up.go`
- `cmd/ubiquity/cmd/up_test.go` or `cmd/ubiquity/cmd/sandbox_deploy_test.go`

**Out of scope**:
- Changing Helm chart dependencies themselves.
- Committing generated `charts/*.tgz` artifacts.

## Steps

1. Add a regression proving sandbox render does not remove an existing `Chart.lock` or existing `charts/*.tgz` in a source chart directory.
2. Refactor sandbox render to copy charts to a temporary work directory before `helm dependency build` and `helm template`.
3. Remove source-tree cleanup of dependency artifacts from sandbox render paths.
4. Make `runHelmTemplateAndApply` return a clear error when dependency update/build fails.
5. Add a focused regression for dependency-update failure propagation using the existing command seams or a small test seam if needed.

## Test plan

- Add focused Go tests for source-tree non-mutation and dependency failure propagation.
- Run `go run ./cmd/ubiquity test --sandbox-deploy`.
- Run `go test ./cmd/ubiquity/cmd -count=1`.

## Done criteria

- [ ] Sandbox Helm render uses a temp workdir for dependency materialization.
- [ ] No code path deletes `Chart.lock` from source chart directories during sandbox validation.
- [ ] Helm dependency errors are returned, not ignored.
- [ ] Focused and package tests pass.

## STOP conditions

Stop if Helm dependency build cannot work from a copied chart because of file:// dependencies that require repo-relative paths; report the specific chart and dependency.
