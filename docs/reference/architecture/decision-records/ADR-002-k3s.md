# ADR-002: Use k3s instead of full Kubernetes

**Status:** Accepted

**Date:** 2025-06-01

## Context

HPC clusters need lightweight K8s for resource-constrained nodes. Full K8s requires more RAM, more components.

## Decision

Use k3s (Rancher's lightweight distribution). Embedded etcd, single binary, supports standard K8s API.

## Consequences

- Positive: Standard K8s API compatibility means most tools work unmodified.
- Positive: Significantly lower resource footprint than full Kubernetes.
- Negative: Some edge features (certain CSI drivers, advanced networking) may need extra configuration.
