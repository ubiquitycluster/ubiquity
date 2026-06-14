---
type: Reference
title: Repository knowledge graph
description: Durable Graphify graph artifacts use repository-root-relative paths so any clone can consume them consistently.
resource: /graphify-out/graph.json
tags: [graphify, portability, agents]
timestamp: 2026-06-14T00:00:00Z
---

# Purpose

`graphify-out/graph.json` is the durable repository knowledge graph. Consumers
should treat `source_file`, manifest keys, chunk file entries, and wiki links as
paths relative to the repository root, not as absolute paths tied to a developer
home directory.

# Agent contract

* Use [`graphify query`](../graph.json), `graphify explain`, and `graphify path`
  for scoped codebase navigation.
* Resolve graph file references from the repository root.
* Tolerate unknown OKF frontmatter keys and broken links, matching OKF's
  permissive consumer model.
* Do not require a specific username, home directory, or checkout location.

# Citations

[1] [Open Knowledge Format specification](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md)
