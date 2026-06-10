# Ubiquity Project Completion Goal

## Goal

Execute the committed project completion plan at:

`docs/plans/2026-06-10-project-completion-plan-excluding-pr-packaging-and-live-thinkcentre-proof.md`

The objective is to finish the remaining local, CI-safe maturity work for Ubiquity while explicitly excluding PR packaging/publishing and live ThinkCentre/bare-metal deployment proof.

## Explicitly out of scope

- Packaging, splitting, or publishing the current local commit stack into reviewable PRs.
- Running live ThinkCentre or other physical bare-metal deployment proof.

## In scope

### 1. Helm chart maturity

- Add a chart maturity contract test.
- Add Helm unittest coverage for the remaining active charts without test directories.
- Replace remaining `version: 0.0.0` placeholder chart versions.
- Enforce chart maturity in local/CI-safe verification.

### 2. TODO/FIXME debt reduction

- Inventory outstanding TODO/FIXME debt.
- Resolve or explicitly categorize critical production/configuration TODOs.
- Pay special attention to secret-handling TODOs and avoid committing raw credentials.

### 3. CLI/root command coverage

- Add root package smoke coverage so `cmd/ubiquity` no longer reports `[no test files]`.
- Add command registry coverage.
- Add provisioning executor safety/regression coverage.

### 4. NICo day-2 lifecycle proof without live hardware execution

- Codify NICo as the default day-2 physical node lifecycle path.
- Keep BMO/Metal3 fallback-only unless explicitly requested.
- Add CI-safe NICo lifecycle contract tests and dry-run proof script.

### 5. Production operations documentation

- Replace cert-manager and Vault runbook TODOs with actionable operations guidance.
- Complete production deployment, post-installation, external resource, and PXE/admin tutorial docs.
- Keep readiness wording evidence-bounded and avoid approval/certification claims without evidence.

### 6. Graphify maintenance

- Document Graphify-first workflow in the repo.
- Add a lightweight Graphify freshness check.
- Keep `graphify-out/` tracked and update it after code/docs changes.

### 7. Generated/reference docs freshness

- Enforce generated chart/reference docs freshness.
- Refresh references and Graphify artifacts at the end of implementation.

## Execution rules

- Use Graphify before project/codebase decisions: `graphify query`, `graphify explain`, or `graphify path` against `graphify-out/graph.json`.
- Use TDD for code and testable documentation contracts.
- Commit each completed feature slice independently.
- Leave unrelated dirty files unstaged.
- Never commit raw credentials, tokens, kubeconfigs, or private keys; redact with `[REDACTED]` if encountered.
- After code/docs changes, run `graphify update .` and commit material graph updates with the slice.
- Do not generate HTML visualization for graphs over 5,000 nodes without explicit approval.

## Acceptance criteria

This goal is complete when the final gate from the implementation plan passes:

```bash
graphify query "project completion chart maturity CLI coverage NICo production operations docs Graphify maintenance" --budget 3000
go test ./pkg/... ./cmd/... -count=1
scripts/generate-helm-chart-reference.sh --check
test/e2e/core-services-proof.sh --dry-run
bash test/e2e/nico-lifecycle-proof.sh --dry-run
git diff --check
git status --short --branch
```

Expected outcome: all in-scope plan phases are implemented, verified, committed, and reflected in `graphify-out/`.
