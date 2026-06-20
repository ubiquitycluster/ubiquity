# Plan 014: Implement NetBird multi-cluster overlay functionality

> **Executor instructions**: Execute this plan on branch `improve/netbird-multicluster-overlay` and add the work to PR #38. Follow TDD: add failing tests first, implement the smallest production functionality that passes, run verification, refresh Graphify, commit, push, and re-check the PR.

## Status

- **Priority**: P1
- **Effort**: M
- **Risk**: MED
- **Depends on**: [013](013-netbird-multicluster-overlay.md)
- **Category**: cloud/gitops/netbird/multicluster
- **Status**: DONE — implemented and verified on `improve/netbird-multicluster-overlay`

## Goal

Turn the NetBird multi-cluster blueprint into executable Ubiquity functionality. Operators should be able to render a placeholder-safe NetBird multi-cluster overlay bundle from the CLI and apply it through the existing `ubiquity cloud render/apply` workflow.

The functionality must preserve the architecture contract from Plan 013:

- NetBird is the private control/data overlay between independent Ubiquity clusters.
- It must not stretch one Kubernetes cluster across regions.
- Public inference traffic uses Geo DNS/global load balancing, not NetBird hairpinning.
- Each regional cluster has its own GPU/RDMA/NICo readiness boundary.
- No NetBird PATs, setup keys, kubeconfigs, bearer tokens, private keys, or CA data are committed or rendered from defaults.

## Scope

### In scope

- Add a `cloud.RenderNetBirdMultiClusterOverlay` renderer with validated request defaults.
- Add CLI support under `ubiquity cloud render netbird-overlay` and `ubiquity cloud apply netbird-overlay`.
- Render a multi-document Kubernetes bundle containing:
  - a policy/readiness ConfigMap,
  - a placeholder-safe ArgoCD cluster Secret template,
  - an ArgoCD ApplicationSet for regional AI platform apps,
  - an RDMA readiness ApplicationSet gated by `ubiquity.io/rdma=true`.
- Add package tests for fail-closed validation and placeholder safety.
- Add CLI tests proving the resource is available and rendered with operator-supplied region/site labels.
- Update docs and Plan 013 references to point from the blueprint to the executable command.
- Refresh Graphify tracked artifacts.
- Commit and push to the existing PR #38 branch.

### Out of scope

- Calling NetBird APIs or requiring a NetBird account.
- Rendering real secrets, setup keys, kubeconfigs, or CA data.
- Installing the NetBird Kubernetes operator live.
- Changing production deployment behavior beyond adding a render/apply resource.

## TDD steps

1. Add `pkg/cloud/netbird_multicluster_test.go` with tests that initially fail because the renderer does not exist.
2. Add `cmd/ubiquity/cmd/cloud_netbird_test.go` with tests that initially fail because `cloud render netbird-overlay` is unsupported.
3. Implement `pkg/cloud/netbird_multicluster.go`.
4. Wire `NetBirdMultiClusterOverlayRequest` through `cloudOptions`, flags, and `renderCloudResource`.
5. Update docs with the exact render/apply commands.
6. Mark this plan DONE after verification passes.

## Validation commands

Run these from the repo root:

```sh
go test ./pkg/cloud -run 'TestRenderNetBird|TestNetBird' -count=1
go test ./cmd/ubiquity/cmd -run 'TestCloudRenderNetBird|TestCloudApplyNetBird|TestCloudCommand' -count=1
go test ./pkg/cloud ./cmd/ubiquity/cmd -count=1
go test ./pkg/cloud -count=1
make test
make build
graphify update .
python3 scripts/normalize-graphify-artifacts.py
git diff --check
scripts/check-graphify-freshness.sh
```

## Done criteria

- [x] `ubiquity cloud render netbird-overlay` emits a valid, placeholder-safe multi-document bundle.
- [x] The bundle includes ApplicationSet, ArgoCD cluster Secret template, NetBird policy/readiness data, and RDMA/NICo readiness gates.
- [x] Validation fails closed for missing region/site/cluster identity and unsupported public-hairpin routing.
- [x] Tests cover package renderer and CLI resource wiring.
- [x] Docs show the command and preserve the blueprint boundaries.
- [x] Graphify artifacts are refreshed and normalized.
- [x] Changes are committed and pushed to PR #38.

## STOP conditions

Stop and report if implementing this requires real NetBird credentials, live NetBird API calls, or a decision to vendor unstable upstream CRDs.
