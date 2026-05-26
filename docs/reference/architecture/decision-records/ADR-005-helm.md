# ADR-005: Use Helm chart per component

**Status:** Accepted

**Date:** 2025-06-01

## Context

Need idempotent, parameterized deployment of 20+ system components.

## Decision

Use Helm. Industry standard, integrates with ArgoCD, supports dependency management.

## Consequences

- Positive: Industry standard tooling with broad ecosystem support.
- Positive: Native ArgoCD integration for GitOps workflows.
- Positive: Dependency management for multi-component deployments.
- Negative: Chart versioning overhead. Use app-template pattern to reduce boilerplate.
