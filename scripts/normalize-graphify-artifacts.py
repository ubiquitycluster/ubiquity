#!/usr/bin/env python3
"""Normalize Graphify artifacts so they are portable across checkouts.

Graphify's checked-in graph is a repository knowledge bundle. Some intermediate
files produced by current Graphify versions contain machine-local absolute paths
such as `/Users/alice/work/repo/...` or `/home/ci/repo/...`. Those paths make the
bundle user-specific even though the durable graph (`graph.json`) already uses
repository-relative `source_file` values.

This script post-processes tracked Graphify artifacts after `graphify update .`:

* paths under the current repository root become repository-root-relative paths;
* other home-directory paths become `${HOME}/...` placeholders;
* `.graphify_python` is reduced to `python3` instead of a user-local tool path;
* an OKF v0.1-compatible wiki bundle is emitted under `graphify-out/wiki/`.
"""

from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
from pathlib import Path, PurePosixPath
from typing import Any

HOME_PATH_RE = re.compile(r"/(?:Users|home)/[^\s\"'<>:]+")
OKF_TIMESTAMP = "2026-06-14T00:00:00Z"


def git_root(start: Path) -> Path:
    out = subprocess.check_output(
        ["git", "rev-parse", "--show-toplevel"], cwd=start, text=True
    ).strip()
    return Path(out).resolve()


def git_ls_files(root: Path, pathspec: str) -> list[Path]:
    out = subprocess.check_output(["git", "ls-files", pathspec], cwd=root, text=True)
    return [root / line for line in out.splitlines() if line]


def normalize_string(value: str, root: Path) -> str:
    root_posix = root.as_posix().rstrip("/")
    root_name = root.name

    # First make any occurrence under this checkout repository-root-relative.
    value = value.replace(root_posix + "/", "")
    value = value.replace(root_posix, ".")
    if value == root_posix:
        return "."

    # Then anonymize any remaining home directory paths, including snippets copied
    # from docs or generated caches. Keep the suffix when it points at this repo.
    def repl(match: re.Match[str]) -> str:
        path = match.group(0)
        marker = f"/{root_name}/"
        if marker in path:
            return path.split(marker, 1)[1]
        return "${HOME}/" + PurePosixPath(path).name

    return HOME_PATH_RE.sub(repl, value)


def normalize_json(value: Any, root: Path) -> Any:
    if isinstance(value, str):
        return normalize_string(value, root)
    if isinstance(value, list):
        return [normalize_json(item, root) for item in value]
    if isinstance(value, dict):
        normalized: dict[str, Any] = {}
        for key, item in value.items():
            normalized_key = normalize_string(str(key), root)
            normalized[normalized_key] = normalize_json(item, root)
        return normalized
    return value


def normalize_file(path: Path, root: Path) -> bool:
    if path.name == ".graphify_python":
        desired = "python3\n"
        if path.read_text(errors="ignore") != desired:
            path.write_text(desired)
            return True
        return False

    try:
        raw = path.read_text()
    except UnicodeDecodeError:
        return False

    # Preserve Graphify's native file formatting. A text-level replacement is
    # enough because the portability issue is path serialization, not JSON
    # structure. This keeps diffs reviewable even for large graph/cache files.
    normalized = normalize_string(raw, root)
    if path.relative_to(root).as_posix() == "graphify-out/manifest.json":
        # Graphify's incremental save path can emit the manifest as one JSON
        # object with duplicate file-path keys when changed-file and full-corpus
        # state are combined. Normal JSON parsers keep only the last value, so
        # normalize to that explicit map contract here and let repository tests
        # guard against duplicate keys returning.
        normalized_obj = json.loads(normalized, object_pairs_hook=dict)
        normalized = json.dumps(normalized_obj, indent=2, ensure_ascii=False) + "\n"
    if path.suffix == ".py" and path.is_relative_to(root / "graphify-out"):
        normalized = re.sub(
            r"ROOT = Path\('[^']*'\)",
            "ROOT = Path(__file__).resolve().parent.parent",
            normalized,
        )

    if normalized != raw:
        path.write_text(normalized)
        return True
    return False


def write_okf_bundle(root: Path) -> list[Path]:
    wiki = root / "graphify-out" / "wiki"
    wiki.mkdir(parents=True, exist_ok=True)

    files = {
        "index.md": f"""---
type: Knowledge Bundle
title: Ubiquity Graphify knowledge bundle
description: Repository-root-relative Open Knowledge Format entrypoint for the checked-in Graphify graph.
tags: [graphify, okf, repository-knowledge]
timestamp: {OKF_TIMESTAMP}
okf_version: "0.1"
---

# Open Knowledge Format entrypoint

This directory is an OKF-compatible, repository-portable entrypoint for the
checked-in Graphify knowledge graph. It follows the Open Knowledge Format model:
plain Markdown files, YAML frontmatter, repository-root-relative links, and
permissive consumption by agents.

# Concepts

* [Repository knowledge graph](repository-knowledge-graph.md) - How agents should consume `graphify-out/graph.json` without depending on a specific user's checkout path.
* [Portable Graphify workflow](portable-graphify-workflow.md) - How to update and normalize Graphify artifacts after repository changes.

# Citations

[1] [Open Knowledge Format specification](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md)
[2] [Google Cloud: How the Open Knowledge Format can improve data sharing](https://cloud.google.com/blog/products/data-analytics/how-the-open-knowledge-format-can-improve-data-sharing/)
""",
        "repository-knowledge-graph.md": f"""---
type: Reference
title: Repository knowledge graph
description: Durable Graphify graph artifacts use repository-root-relative paths so any clone can consume them consistently.
resource: /graphify-out/graph.json
tags: [graphify, portability, agents]
timestamp: {OKF_TIMESTAMP}
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
""",
        "portable-graphify-workflow.md": f"""---
type: Playbook
title: Portable Graphify workflow
description: Steps for maintaining a Graphify bundle that works in any user's repository clone.
tags: [graphify, okf, workflow]
timestamp: {OKF_TIMESTAMP}
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
""",
    }

    written = []
    for name, content in files.items():
        path = wiki / name
        if not path.exists() or path.read_text() != content:
            path.write_text(content)
            written.append(path)
    return written


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--check",
        action="store_true",
        help="fail if tracked graphify artifacts still contain home-directory paths",
    )
    args = parser.parse_args()

    root = git_root(Path.cwd())
    changed = []
    for path in git_ls_files(root, "graphify-out"):
        if path.is_file() and normalize_file(path, root):
            changed.append(path.relative_to(root).as_posix())

    for path in write_okf_bundle(root):
        rel = path.relative_to(root).as_posix()
        if rel not in changed:
            changed.append(rel)

    if args.check:
        offenders = []
        for path in git_ls_files(root, "graphify-out"):
            if not path.is_file():
                continue
            content = path.read_text(errors="ignore")
            if "/Users/" in content or "/home/" in content:
                offenders.append(path.relative_to(root).as_posix())
        if offenders:
            for rel in offenders:
                print(f"machine-local path remains: {rel}")
            return 1

    if changed:
        print("normalized graphify artifacts:")
        for rel in sorted(changed):
            print(f"  {rel}")
    else:
        print("graphify artifacts already portable")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
