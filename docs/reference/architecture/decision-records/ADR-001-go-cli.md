# ADR-001: Use Go CLI instead of Python

**Status:** Accepted

**Date:** 2025-06-01

## Context

Originally configured via Python scripts (scripts/configure, scripts/configure-sandbox). Needed a cross-platform, single-binary experience with no runtime dependencies.

## Decision

Rewrite as a Go CLI using cobra + viper. Go produces static binaries, has excellent CLI libraries.

## Consequences

- Positive: No Python/runtime dependencies required.
- Positive: Reduced maintenance burden from 1,672 lines of Python to 834 lines of Go.
- Negative: Larger binary size (~8MB).
