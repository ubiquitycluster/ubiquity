# Graphify workflow

This repository tracks a project knowledge graph in `graphify-out/graph.json`.
Use it as the first navigation tool for architecture, codebase, and relationship
questions before falling back to raw source browsing.

## When to query the graph

For project or codebase questions, start with a scoped query:

```sh
graphify query "current remaining implementation gaps" --budget 2500
```

Use focused commands when the question is about a concept or relationship:

```sh
graphify explain "NICo lifecycle safety gates"
graphify path "ubiquity nodes" "pkg/nodestatus/safety.go"
```

`graphify query`, `graphify explain`, and `graphify path` usually return a much
smaller and more relevant subgraph than reading `graphify-out/GRAPH_REPORT.md` or
running broad text searches.

## Updating the graph

After code or documentation changes, run:

```sh
graphify update .
```

commit material graph updates with the same slice that changed the code or docs.
It is normal for `graphify-out/` to become dirty after hooks or incremental
updates; dirty graph files are not a reason to skip Graphify unless the task is
specifically about stale or incorrect graph output.

## Visualization and size limits

Do not generate HTML visualization for graphs over 5,000 nodes without explicit
approval. Large graphs should use `graphify update .` or `graphify cluster-only .
--no-viz` so the repository keeps `graph.json` and `GRAPH_REPORT.md` current
without producing an oversized `graph.html`.

## Token cost reporting

Graphify reports should show token cost when semantic extraction is performed.
If a report is manually assembled or semantic chunks are produced by delegated
agents, update `graphify-out/GRAPH_REPORT.md` with the input and output token
cost so reviewers can see extraction cost.

## Freshness check

Use the lightweight freshness check before final verification:

```sh
scripts/check-graphify-freshness.sh
scripts/check-graphify-freshness.sh --strict
```

The script checks that `graphify-out/graph.json` and
`graphify-out/GRAPH_REPORT.md` exist and compares their last commit with the
current `HEAD`. It does not run extraction; it is a cheap signal that prompts a
human or agent to run `graphify update .` when graph artifacts look stale.
