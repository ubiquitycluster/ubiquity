# Plan 008: Pin CI supply chain and close remaining fail-open gates

> **Executor instructions**: Make CI gates honest and deterministic. Prefer pinned versions and explicit advisory steps over hidden `|| true` masking.
>
> **Drift check (run first)**: `git diff --stat efd46ed..HEAD -- .github/workflows/ci.yaml .github/workflows/release.yaml pkg/cloud/security_ci_test.go`

## Status

- **Priority**: P1
- **Effort**: M
- **Risk**: MED
- **Depends on**: none
- **Category**: ci/supply-chain
- **Planned at**: commit `efd46ed`, 2026-06-15
- **Status**: DONE (implemented and verified)

## Why this matters

A CI pipeline that claims to gate lint, validation, scanning, and tests should not silently ignore failures or consume mutable tools/actions without reviewed changes.

## Current state

- `.github/workflows/ci.yaml:54-60` checks only the PR head commit file list for pre-commit.
- `.github/workflows/ci.yaml:68-76` masks Ansible/Terraform failures.
- `.github/workflows/ci.yaml:39-40`, `81`, `141-142`, `168`, `199`, `204` use mutable/unverified installer inputs.
- `.github/workflows/release.yaml:22-30` uses mutable GoReleaser/SBOM action versions.
- Graphify strict freshness is not invoked by `.github/workflows`.

## Commands you will need

| Purpose | Command | Expected on success |
|---|---|---|
| CI contract tests | `go test ./pkg/cloud -run 'SecurityCI|Graphify|Workflow' -count=1` | exits 0 |
| YAML syntax | `python3 -c 'import yaml, pathlib; [yaml.safe_load(p.read_text()) for p in pathlib.Path(".github/workflows").glob("*.yaml")]'` | exits 0 if PyYAML available |

## Scope

**In scope**:
- `.github/workflows/ci.yaml`
- `.github/workflows/release.yaml`
- `pkg/cloud/security_ci_test.go` or adjacent CI contract tests

**Out of scope**:
- Resolving every existing lint violation in `metal/` or Terraform directories unless required to make CI pass locally; if violations exist, encode scoped excludes with comments.

## Steps

1. Add/extend CI contract tests that reject `|| true` on Ansible/Terraform gates, reject `@master`, reject `version: latest` for GoReleaser, and require Graphify strict freshness in CI.
2. Use PR base/head diff for pre-commit file selection instead of `git diff-tree` on only the head commit.
3. Remove fail-open masking from Ansible/Terraform gates. If necessary, mark known exclusions explicitly.
4. Pin tool installer versions in env variables and replace `@latest` / `releases/latest` where practical.
5. Add `scripts/check-graphify-freshness.sh --strict` as a CI step.
6. Keep Trivy and release actions on stable pinned versions or SHAs.

## Done criteria

- [ ] CI contract tests cover remaining fail-open/mutable patterns.
- [ ] Ansible/Terraform failures are not hidden by blanket `|| true`.
- [ ] Graphify strict freshness runs in GitHub CI.
- [ ] Focused CI tests pass.

## STOP conditions

Stop if removing fail-open gates reveals a large unrelated backlog of existing Ansible/Terraform failures; report exact failures and convert those gates to explicit advisory jobs only if user approves.
