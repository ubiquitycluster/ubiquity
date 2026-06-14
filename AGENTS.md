## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

When the user types `/graphify`, invoke the `skill` tool with `skill: "graphify"` before doing anything else.

Rules:
- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- Dirty graphify-out/ files are expected after hooks or incremental updates; dirty graph files are not a reason to skip graphify. Only skip graphify if the task is about stale or incorrect graph output, or the user explicitly says not to use it.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Treat graphify-out/wiki/ as an Open Knowledge Format (OKF) entrypoint: Markdown plus YAML frontmatter, repository-root-relative links, and permissive consumers that tolerate unknown fields or missing links.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` and then `python3 scripts/normalize-graphify-artifacts.py` to keep the graph current and portable (AST-only, no API cost).
- Do not generate HTML visualization for graphs over 5,000 nodes without explicit approval; prefer no-viz graph updates for large graphs.
- Graphify reports should show token cost whenever semantic extraction or delegated semantic processing was used.
- See `docs/developers/graphify-workflow.md` and `scripts/check-graphify-freshness.sh` for the repository workflow and freshness check.
