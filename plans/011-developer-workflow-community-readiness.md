# Plan 011: Developer workflow and community readiness

> **Executor instructions**: This plan reconciles developer-experience, community-readiness, CLI/TUI ergonomics, and local-test reliability work already present in the working tree. Keep changes project-native and avoid adding live-cluster requirements to default tests.
>
> **Drift check (run first)**: `git diff --stat efd46ed..HEAD -- CONTRIBUTING.md README.md Makefile cmd/ubiquity/cmd/init.go pkg/tui test pkg/cloud/community_readiness_test.go go.mod go.sum`

## Status

- **Priority**: P2
- **Effort**: M
- **Risk**: MED
- **Depends on**: none
- **Category**: developer-experience/tests/community-readiness
- **Planned at**: commit `efd46ed`, 2026-06-15
- **Status**: DONE (implemented and verified)

## Why this matters

The repository's default contributor path should be runnable on fresh developer machines and CI runners without hidden tools, live clusters, Docker image directories, or terminal-only behavior. The working tree includes changes that make local builds/tests deterministic and document the expected flow.

## Reconciled dirty files

- `CONTRIBUTING.md`, `README.md` — Markdown cleanup and contributor/readme coverage.
- `Makefile` — safer OS detection, Go CLI-backed configure targets, and non-fatal legacy image build behavior when image contexts are absent.
- `cmd/ubiquity/cmd/init.go` — default environment/config generation alignment.
- `pkg/tui/status.go`, `pkg/tui/status_test.go`, `go.mod`, `go.sum` — Bubble Tea-backed status rendering with fallback and dependency updates.
- `test/Makefile`, `test/external_test.go`, `test/integration_test.go`, `test/tools_test.go` — short-mode/default-test skips for live Terraform, cluster, and container tests.
- `pkg/cloud/community_readiness_test.go` — docs/devcontainer/pre-commit/community artifact coverage.

## Acceptance evidence

- `go test ./pkg/... ./cmd/... -count=1`
- `make test`
- `make build`

## Done criteria

- [x] Default `make test` avoids live-cluster/container/Terraform requirements unless explicitly enabled.
- [x] `make build` builds the CLI and skips absent legacy HPC image contexts instead of failing spuriously.
- [x] Contributor docs and Markdown fences are test-covered.
- [x] TUI status path remains covered by Go tests and has a non-interactive fallback.
