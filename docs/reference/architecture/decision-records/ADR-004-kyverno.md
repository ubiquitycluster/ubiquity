# ADR-004: Use Kyverno over OPA/Gatekeeper

**Status:** Accepted

**Date:** 2025-06-01

## Context

Need Kubernetes admission control for security policies.

## Decision

Use Kyverno. Kubernetes-native (no custom DSL like OPA's Rego), supports policy mutation, simpler YAML-only policies.

## Consequences

- Positive: No custom DSL to learn — policies are standard Kubernetes resources.
- Positive: Supports policy mutation in addition to validation.
- Negative: Locked to K8s-specific policies. But for cluster-only policies, Kyverno is simpler.
