---
type: Playbook
title: Portable Graphify workflow
description: Steps for maintaining a Graphify bundle that works in any user's repository clone.
tags: [graphify, okf, workflow]
timestamp: 2026-06-14T00:00:00Z
---

# Steps

1. Run `graphify update .` after code or documentation changes.
2. Run `python3 scripts/normalize-graphify-artifacts.py` to strip checkout-local
   home directory paths and refresh this OKF entrypoint.
3. Run `go test ./pkg/cloud -run Graphify -count=1` and
   `scripts/check-graphify-freshness.sh --strict`.
4. Commit code, docs, and material graph updates in the same slice.

# Rationale

OKF defines a knowledge bundle as plain files that can be cloned, diffed, and
consumed without a proprietary platform or machine-specific runtime. The
Graphify bundle follows that model by keeping durable paths repository-root
relative.

# Citations

[1] [Google Cloud OKF blog](https://cloud.google.com/blog/products/data-analytics/how-the-open-knowledge-format-can-improve-data-sharing/)
