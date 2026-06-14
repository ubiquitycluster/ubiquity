# ADR-005: Use Helm chart per component

**Status:** Accepted

**Date:** 2025-06-01

## Context

Need idempotent, parameterized deployment of 20+ system components.

## Decision

Use Helm for component packaging. Industry standard, integrates with ArgoCD, supports dependency management, and gives each platform or system component a versioned values interface. Use Kustomize only for environment-specific overlays or composition layers such as `platform/hpc-ubiq`, not as a replacement for per-component Helm packaging.

## Consequences

- Positive: Industry standard tooling with broad ecosystem support.
- Positive: Native ArgoCD integration for GitOps workflows.
- Positive: Dependency management for multi-component deployments.
- Negative: Chart versioning overhead. Use app-template pattern to reduce boilerplate.
