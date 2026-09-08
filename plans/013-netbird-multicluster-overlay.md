# Plan 013: Add NetBird multi-cluster overlay architecture

> **Executor instructions**: Follow this plan step by step. Run every verification command and confirm the expected result before moving to the next step. If anything in the STOP conditions occurs, stop and report — do not improvise. When done, update the status row for this plan in `plans/README.md`.
>
> **Drift check (run first)**: `git diff --stat 2e92d4a..HEAD -- docs/architecture/networking.md docs/architecture/overview.md docs/index.md mkdocs.yml bootstrap/README.md pkg/cloud/architecture_docs_test.go`
> If any listed files changed since this plan was written, compare the Current state excerpts against live code before proceeding; on mismatch, stop and report.

## Status

- **Priority**: P1
- **Effort**: M
- **Risk**: MED
- **Depends on**: none
- **Category**: architecture/docs/tests/gitops
- **Status**: DONE — implemented and verified on `improve/netbird-multicluster-overlay`
- **Planned at**: commit `2e92d4a`, 2026-06-17

## Why this matters

Ubiquity already has ArgoCD, ApplicationSet, NVIDIA GPU/RDMA/NICo, fail-closed readiness, and deny-first networking contracts, but it does not yet document or test the intended geographic multi-cluster operating model. The requested design must make the boundary explicit: NetBird connects independent Ubiquity clusters for private management and optional private service access; it must not be represented as one stretched Kubernetes cluster or as the public inference edge.

## Current state

- `docs/architecture/networking.md` documents local cluster ingress, NetworkPolicy, AI/RDMA evidence, and cloudflared boundaries, but has no NetBird/geographic multi-cluster section.
- `docs/architecture/overview.md` documents one-cluster provisioning flow through metal/cloud, bootstrap, system, platform, and apps.
- `bootstrap/README.md` mentions that tenants can deploy to other clusters connected to ArgoCD, but has no concrete fleet/ApplicationSet pattern for region labels and NetBird-reachable cluster APIs.
- `docs/runbooks/cloud-readiness-validation.md` defines fail-closed readiness evidence, but not the additional multi-cluster evidence required before routing traffic to a regional endpoint.
- `mkdocs.yml` has no navigation entry for a multi-cluster NetBird architecture page.

## Commands you will need

| Purpose | Command | Expected on success |
|---|---|---|
| Focused docs contract | `go test ./pkg/cloud -run TestMultiClusterNetBirdArchitectureContract -count=1` | exits 0; new contract test passes |
| Package tests | `go test ./pkg/cloud -count=1` | exits 0 |
| Markdown/navigation smoke | `python3 - <<'PY'\nfrom pathlib import Path\nfor p in ['docs/architecture/multi-cluster-netbird.md','docs/reference/multi-cluster-netbird/application-set.yaml','docs/reference/multi-cluster-netbird/cluster-secret-template.yaml']:\n    assert Path(p).exists(), p\nprint('multi-cluster docs present')\nPY` | exits 0 and prints presence marker |
| Graph refresh | `graphify update . && python3 scripts/normalize-graphify-artifacts.py` | exits 0; graph artifacts normalized |
| Full verification | `make test` | exits 0 |
| Build verification | `make build` | exits 0 |

Run from the repo root.

## Scope

**In scope**:
- `docs/architecture/multi-cluster-netbird.md` — primary architecture and operating model.
- `docs/architecture/networking.md` — short boundary/link section so networking readers find the model.
- `docs/architecture/overview.md` — high-level topology mention.
- `docs/runbooks/cloud-readiness-validation.md` — multi-cluster routing/readiness evidence extension.
- `docs/index.md` and `mkdocs.yml` — navigation/linkage.
- `bootstrap/README.md` — GitOps/ApplicationSet fleet note.
- `docs/reference/multi-cluster-netbird/*.yaml` — non-secret example manifests for ApplicationSet, cluster secret template, and NetBird policy matrix.
- `pkg/cloud/*_test.go` — regression tests proving the docs and examples preserve the intended contract.
- `graphify-out/*` — refreshed tracked knowledge graph artifacts after changes.
- `plans/README.md` and this plan — status tracking.

**Out of scope**:
- Installing live NetBird or creating a NetBird account/PAT.
- Committing real setup keys, PATs, kubeconfigs, bearer tokens, or cluster CA data.
- Stretching Kubernetes control planes, etcd, CNI, RDMA, NCCL, or storage across regions.
- Replacing existing Ubiquity ingress, NetworkPolicy, GPU Operator, Network Operator, or NICo charts.
- Publishing/pushing/PR creation unless the operator asks.

## Git workflow

- Branch: `improve/netbird-multicluster-overlay`, created from freshly pulled `origin/main`.
- Commit style: `docs: add netbird multi-cluster overlay architecture` after verification.
- Do not push or open a PR unless explicitly requested.

## Steps

### Step 1: Add a failing docs contract test

Create `pkg/cloud/multicluster_netbird_test.go`. The test must require:

- `docs/architecture/multi-cluster-netbird.md` exists.
- The doc says NetBird is a private control/data overlay between independent Ubiquity clusters.
- The doc says not to stretch one Kubernetes cluster across regions.
- The doc distinguishes public Geo DNS/global load balancing from NetBird.
- The doc preserves local RDMA/NCCL and per-cluster NICo boundaries.
- Example manifests include `ApplicationSet`, `argocd.argoproj.io/secret-type: cluster`, `ubiquity.io/region`, `ubiquity.io/rdma`, and placeholder-only secret material.

**Verify**: `go test ./pkg/cloud -run TestMultiClusterNetBirdArchitectureContract -count=1` should fail because the files do not exist yet.

### Step 2: Write the architecture doc and examples

Add `docs/architecture/multi-cluster-netbird.md` with:

- The central management cluster + regional workload clusters topology.
- Phased rollout from management mesh through first remote cluster, NVIDIA/NICo stack, inference routing, and fleet hardening.
- NetBird policy model for ArgoCD, platform admins, SRE observability, CI, and regional clusters.
- ApplicationSet label taxonomy and targeting rules.
- Geo DNS/global LB health checks for inference.
- Data locality guidance for stateless inference, batch training, and synchronous distributed training.
- Ubiquity-specific NVIDIA/NICo readiness rules.
- Explicit non-goals and security boundaries.

Add `docs/reference/multi-cluster-netbird/application-set.yaml`, `cluster-secret-template.yaml`, and `netbird-policy-matrix.yaml` as placeholder-safe examples.

**Verify**: `go test ./pkg/cloud -run TestMultiClusterNetBirdArchitectureContract -count=1` should pass.

### Step 3: Link the model from existing docs

Patch `docs/architecture/networking.md`, `docs/architecture/overview.md`, `docs/runbooks/cloud-readiness-validation.md`, `bootstrap/README.md`, `docs/index.md`, and `mkdocs.yml` so operators can discover the multi-cluster model from architecture, readiness, GitOps, and docs navigation entry points.

**Verify**: `go test ./pkg/cloud -count=1` should pass.

### Step 4: Refresh graph artifacts

Run `graphify update .` and `python3 scripts/normalize-graphify-artifacts.py`. If Graphify reports code-only updates for documentation changes, state that semantic documentation extraction remains pending rather than overstating semantic coverage. Keep tracked artifacts portable.

**Verify**: `python3 scripts/normalize-graphify-artifacts.py` exits 0 and `git diff --check` exits 0.

### Step 5: Run full gates and finalize plan status

Run `make test` and `make build`. If unavailable tooling blocks either command, run the focused checks and report the blocker with exact output. Mark this plan DONE in `plans/README.md` and update this status block only after implementation and verification are complete.

## Test plan

- RED/GREEN contract: `go test ./pkg/cloud -run TestMultiClusterNetBirdArchitectureContract -count=1`.
- Package regression: `go test ./pkg/cloud -count=1`.
- Repo gates: `make test`, `make build`.
- Graph/artifact hygiene: `graphify update .`, `python3 scripts/normalize-graphify-artifacts.py`, `git diff --check`.

## Done criteria

- [x] Multi-cluster NetBird architecture doc exists and is linked from docs navigation.
- [x] Placeholder-safe GitOps/cluster/NetBird policy examples exist.
- [x] Tests fail before docs/examples exist and pass after implementation.
- [x] Readiness docs require per-region live evidence before traffic routing.
- [x] Docs clearly state not to stretch Kubernetes/RDMA/NCCL/storage across geography.
- [x] `go test ./pkg/cloud -count=1` exits 0.
- [x] `make test` exits 0 or blocker is reported exactly.
- [x] `make build` exits 0 or blocker is reported exactly.
- [x] Graphify artifacts are refreshed/normalized or semantic-refresh limitation is explicitly reported.

## STOP conditions

Stop and report back if:

- Live NetBird credentials, setup keys, kubeconfigs, or bearer tokens are required.
- Implementing this requires changing production deployment behavior rather than adding documented architecture/reference examples.
- Existing docs contradict the independent-cluster boundary and need a larger architecture decision.
- Full verification fails for unrelated repository failures after focused tests pass.

## Maintenance notes

Future implementation can convert the reference YAML into a chart or scaffold command once real NetBird operator CRDs and cluster-registration workflows are selected. Until then, examples must remain placeholder-safe and documentation-only.
