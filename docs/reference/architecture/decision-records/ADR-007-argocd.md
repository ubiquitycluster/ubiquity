# ADR-007: Use ArgoCD over Flux

**Status:** Accepted

**Date:** 2025-06-01

## Context

Need GitOps operator for cluster management.

## Decision

Use ArgoCD. ApplicationSet for multi-cluster support, mature SSO integration, better UI.

## Consequences

- Positive: ApplicationSet enables multi-cluster/multi-environment patterns.
- Positive: Mature SSO integration.
- Positive: Better UI than Flux.
- Negative: Heavier resource footprint than Flux. But ApplicationSet is critical for multi-environment pattern.
