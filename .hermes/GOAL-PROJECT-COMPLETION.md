# Active Goal: Ubiquity Project Completion

Set as the active execution goal for Hermes.

Canonical project goal file:

`docs/GOAL-PROJECT-COMPLETION.md`

Implementation plan:

`docs/plans/2026-06-10-project-completion-plan-excluding-pr-packaging-and-live-thinkcentre-proof.md`

## Goal

Finish the remaining local, CI-safe project maturity work while explicitly excluding:

- PR packaging/publishing/splitting of the existing local commit stack.
- Live ThinkCentre/bare-metal deployment proof.

## Execute in this order

1. Helm chart maturity.
2. CLI/root command coverage.
3. NICo dry-run lifecycle proof.
4. Production operations docs.
5. TODO/FIXME debt reduction.
6. Graphify maintenance docs/checks.
7. Generated docs and Graphify refresh.

## Required execution discipline

- Check `graphify-out/graph.json` first for project/codebase decisions.
- Use TDD for code and executable contracts.
- Commit each completed feature slice independently.
- Keep destructive/live hardware actions behind explicit safety gates.
- Do not commit secrets; redact real values as `[REDACTED]`.
- Run `graphify update .` after code/docs changes and commit material graph updates.

## Current next slice

Start with Phase A, Task A1 from the plan:

Add `pkg/cloud/chart_maturity_contract_test.go` as a RED test proving active charts must have `tests/` directories and non-placeholder versions.
