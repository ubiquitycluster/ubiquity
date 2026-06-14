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

Checks that graphify-out/graph.json and graphify-out/GRAPH_REPORT.md exist,
reports whether the latest committed graphify-out update is older than HEAD, and
verifies that committed Graphify artifacts are repository-portable (no user home
directory paths). This script does not run graph extraction.

Set GRAPHIFY_FRESHNESS_STRICT=true or pass --strict to exit non-zero when graph
artifacts appear stale or non-portable.
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
require_file graphify-out/wiki/index.md

if [[ -f scripts/normalize-graphify-artifacts.py ]]; then
  if ! python3 scripts/normalize-graphify-artifacts.py --check >/tmp/graphify-portability-check.out 2>&1; then
    fail "graphify-out artifacts contain checkout-local paths; run python3 scripts/normalize-graphify-artifacts.py"
    cat /tmp/graphify-portability-check.out >&2 || true
  fi
else
  fail "scripts/normalize-graphify-artifacts.py is missing; cannot verify Graphify portability"
fi

head_commit="$(git rev-parse HEAD)"
graph_commit="$(git log -1 --format=%H -- graphify-out 2>/dev/null || true)"
graph_dirty="$(git status --porcelain -- graphify-out 2>/dev/null || true)"

if [[ -z "$graph_commit" ]]; then
  fail "graphify-out artifacts have no committed history; run graphify update . and commit them"
fi

if [[ -n "$graph_dirty" ]]; then
  echo "Graphify artifacts present with uncommitted graphify-out updates; commit them with this slice. Last graph commit: ${graph_commit:-none}; HEAD: $head_commit"
  exit 0
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
