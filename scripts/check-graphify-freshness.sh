#!/usr/bin/env bash
set -euo pipefail

strict="${GRAPHIFY_FRESHNESS_STRICT:-false}"
for arg in "$@"; do
  case "$arg" in
    --strict)
      strict="true"
      ;;
    -h|--help)
      cat <<'USAGE'
Usage: scripts/check-graphify-freshness.sh [--strict]

Checks that graphify-out/graph.json and graphify-out/GRAPH_REPORT.md exist and
reports whether the latest committed graphify-out update is older than HEAD. This
script does not run graph extraction.

Set GRAPHIFY_FRESHNESS_STRICT=true or pass --strict to exit non-zero when graph
artifacts appear stale.
USAGE
      exit 0
      ;;
    *)
      echo "unknown argument: $arg" >&2
      exit 2
      ;;
  esac
done

fail() {
  if [[ "$strict" == "true" ]]; then
    echo "ERROR: $*" >&2
    exit 1
  fi
  echo "WARNING: $*" >&2
}

require_file() {
  local path="$1"
  if [[ ! -f "$path" ]]; then
    fail "$path is missing; run graphify update ."
  fi
}

require_file graphify-out/graph.json
require_file graphify-out/GRAPH_REPORT.md

head_commit="$(git rev-parse HEAD)"
graph_commit="$(git log -1 --format=%H -- graphify-out 2>/dev/null || true)"

if [[ -z "$graph_commit" ]]; then
  fail "graphify-out artifacts have no committed history; run graphify update . and commit them"
fi

if [[ -n "$graph_commit" && "$graph_commit" != "$head_commit" ]]; then
  changed_since_graph="$(git diff --name-only "$graph_commit"..HEAD -- ':!graphify-out/**' || true)"
  if [[ -n "$changed_since_graph" ]]; then
    fail "graphify-out last changed at $graph_commit but HEAD is $head_commit; run graphify update . after code/docs changes"
  fi
fi

if [[ -n "${graph_commit:-}" ]]; then
  echo "Graphify artifacts present; last graph commit: $graph_commit; HEAD: $head_commit"
else
  echo "Graphify artifacts present but no committed graph commit was found"
fi
