# Plan 004: Fix Graphify manifest duplicate keys

> **Executor instructions**: Follow this plan step by step. Run every verification command and confirm the expected result before moving to the next step. If anything in the STOP conditions occurs, stop and report — do not improvise. When done, update the status row for this plan in `plans/README.md`.
>
> **Drift check (run first)**: `git diff --stat efd46ed..HEAD -- graphify-out/manifest.json graphify-out/graph.json scripts/normalize-graphify-artifacts.py scripts/check-graphify-freshness.sh pkg/cloud/graphify_portability_test.go`
> If any listed files changed since this plan was written, compare the Current state excerpts against live code before proceeding; on mismatch, stop and report.

## Status

- **Priority**: P2
- **Effort**: M
- **Risk**: MED
- **Depends on**: none
- **Category**: generated-artifact-contract
- **Planned at**: commit `efd46ed`, 2026-06-15
- **Status**: DONE (implemented and verified)

## Why this matters

`graphify-out/manifest.json` currently contains duplicate JSON object keys for every file. Most JSON parsers silently keep only the last member value, making the manifest ambiguous as a review artifact and unsafe as a deterministic freshness contract. This matters because the repository treats tracked Graphify artifacts as an OKF-style knowledge bundle and uses freshness checks during development.

## Current state

- Duplicate key probe captured during planning:

```text
manifest key occurrences 2516
unique keys 1258
duplicate keys 1258
.devcontainer/devcontainer.json [2, 6292]
platform/ai-platform-console/values.schema.json [6257, 7817]
```

- The duplicate line pairs show the same path emitted twice in one JSON object:
  - `graphify-out/manifest.json:2` and `graphify-out/manifest.json:6292`
  - `graphify-out/manifest.json:6257` and `graphify-out/manifest.json:7817`

## Commands you will need

| Purpose | Command | Expected on success |
|---|---|---|
| Duplicate-key probe | `python3 scripts/check-graphify-manifest-unique-keys.py` | exits 0 after adding the script/test, or equivalent Go test passes |
| Existing portability test | `go test ./pkg/cloud -run Graphify -count=1` | exits 0 |
| Freshness check | `bash scripts/check-graphify-freshness.sh` | exits 0 |
| Graph normalize | `python3 scripts/normalize-graphify-artifacts.py` | exits 0 and preserves portable paths |
| Full Go checks | `go test ./pkg/... ./cmd/... -count=1` | exits 0 |

Run from the repo root.

## Scope

**In scope**:
- Add a duplicate-key validation check for `graphify-out/manifest.json`.
- Fix the tracked `graphify-out/manifest.json` representation so each file path appears once, or migrate to an explicit array schema if duplicate records are intentional.
- Update `scripts/normalize-graphify-artifacts.py`, `scripts/check-graphify-freshness.sh`, or `pkg/cloud/graphify_portability_test.go` if they are the right enforcement point.

**Out of scope**:
- Rewriting Graphify itself outside this repository unless explicitly requested.
- Dropping existing graph nodes/edges to make the manifest fit.
- Forcing a full destructive graph rebuild that reduces graph coverage without review.

## Git workflow

- Branch: `improve/004-graphify-manifest-unique-keys`
- Commit per logical step using the repo's commit style.
- Do not push or open a PR unless the operator instructed it.

## Steps

### Step 1: Add a duplicate-key detector

Implement a lightweight validation script or Go test that scans `graphify-out/manifest.json` while preserving duplicate key occurrences. A normal `json.load` is insufficient because it hides duplicates.

The detector should report duplicate key counts and sample paths without printing massive output.

**Verify**: the detector fails against the current manifest and reports 1258 duplicated keys or equivalent.

### Step 2: Identify why manifest entries are duplicated

Inspect the manifest generation path and the local tracked Graphify update workflow. Determine whether duplicates come from combining changed-file and full-corpus manifest sections, or from running normalizer/save-manifest twice.

**Verify**: record the root cause in code comments or the plan execution notes.

### Step 3: Fix manifest generation or normalization

Choose the smallest safe fix:
- If the manifest is intended to be a map, deduplicate by path before writing and keep the canonical latest hash/mtime data.
- If duplicate historical records are intended, change the schema to an array of records and update consumers/tests accordingly.

For this repo, prefer preserving the existing map contract unless Graphify consumers require record history.

**Verify**: duplicate-key detector exits 0.

### Step 4: Refresh and normalize Graphify artifacts safely

Run the repository Graphify update workflow with the user's semantic-extraction preference if semantic docs changed. Otherwise run the deterministic update/normalizer only as appropriate. Do not force-overwrite a graph that loses a large number of nodes.

**Verify**:
- `python3 scripts/normalize-graphify-artifacts.py`
- `bash scripts/check-graphify-freshness.sh`
- duplicate-key detector

## Test plan

- New duplicate-key validator or Go test.
- `go test ./pkg/cloud -run Graphify -count=1`
- `bash scripts/check-graphify-freshness.sh`
- `python3 scripts/normalize-graphify-artifacts.py`
- `go test ./pkg/... ./cmd/... -count=1`

## Done criteria

- [ ] `graphify-out/manifest.json` has no duplicate JSON object keys.
- [ ] The repository has an automated check preventing duplicate manifest keys from returning.
- [ ] Graphify artifacts remain portable/repo-relative.
- [ ] Freshness check exits 0.
- [ ] No graph coverage is lost without explicit review.

## STOP conditions

Stop and report back if:

- Fixing the manifest requires changing Graphify's upstream package behavior outside this repo.
- A normal Graphify merge attempts to reduce graph node count substantially.
- Consumers disagree on whether `manifest.json` is a map or historical record list.
- Semantic extraction is required and the local Codex-authenticated subagent path is unavailable.

## Maintenance notes

Do not use a standard JSON parser as the only duplicate-key check. Python's `json` module and many Go decoders will silently retain only the last duplicate object member.
