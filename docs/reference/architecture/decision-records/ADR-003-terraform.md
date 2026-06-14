# ADR-003: Use Terraform for cloud provisioning

**Status:** Accepted

**Date:** 2025-06-01

## Context

Need to provision infrastructure across 5 cloud providers (AWS, Azure, GCP, OpenStack, OVH).

## Decision

Use Terraform. Mature multi-provider ecosystem, HCL is declarative, large community.

## Consequences

- Positive: Single tool works across all target cloud providers.
- Positive: Declarative HCL syntax is well understood by the team.
- Negative: State management needed (backends).
- Neutral: License change concerns — OpenTofu migration path noted as contingency.
