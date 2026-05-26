# ADR-012: Use devcontainer for reproducible development environments

**Status:** Accepted

**Date:** 2026-05-26

## Context

New contributors spent significant time installing Go, Helm, kubectl, pre-commit,
shellcheck, etc. before they could make their first contribution. Different team
members had different tool versions, causing "works on my machine" issues.

## Decision

Use a VS Code devcontainer with pinned tool versions. "Reopen in Container"
automatically provisions the full environment with all tools at known versions.

## Consequences

- Positive: Reduces onboarding from hours to minutes
- Positive: All team members use identical tool versions
- Negative: Requires VS Code and Docker
- Neutral: Can be bypassed by installing tools manually