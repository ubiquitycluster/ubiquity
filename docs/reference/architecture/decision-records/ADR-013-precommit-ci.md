# ADR-013: Use pre-commit.ci for automated linting

**Status:** Accepted

**Date:** 2026-05-26

## Context

Pre-commit hooks were installed locally but not enforced in CI. PRs could bypass
linting entirely. GitHub Actions was used for linting but consumed compute credits.

## Decision

Enable pre-commit.ci service to run pre-commit on every PR automatically.
This is free for open-source projects and runs in parallel to GitHub Actions.

## Consequences

- Positive: Zero-config CI for linting — no workflow YAML needed
- Positive: Frees GitHub Actions compute for actual testing
- Positive: Auto-fix commits for trivial issues
- Neutral: Requires pre-commit.ci GitHub App installation