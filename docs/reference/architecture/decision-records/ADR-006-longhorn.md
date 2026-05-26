# ADR-006: Use Longhorn as primary storage

**Status:** Accepted

**Date:** 2025-06-01

## Context

Need distributed block storage for K8s workloads. Rook-Ceph was the alternative.

## Decision

Use Longhorn. Simpler operations (UI, no CRUSH map), built-in backup/DR, lighter resource requirements.

## Consequences

- Positive: Simpler operations with a built-in UI and no CRUSH map to manage.
- Positive: Built-in backup and disaster recovery capabilities.
- Positive: Lighter resource requirements than Ceph.
- Negative: Less mature than Ceph for very large clusters. But for typical ubiquity deployments (5-20 nodes), Longhorn is sufficient.
