# Plan 002: Keep NVIDIA AI sandbox targets in sync

> **Executor instructions**: Follow this plan step by step. Run every verification command and confirm the expected result before moving to the next step. If anything in the STOP conditions occurs, stop and report — do not improvise. When done, update the status row for this plan in `plans/README.md`.
>
> **Drift check (run first)**: `git diff --stat efd46ed..HEAD -- cmd/ubiquity/cmd/ai_platform.go cmd/ubiquity/cmd/up.go cmd/ubiquity/cmd/up_test.go cmd/ubiquity/cmd/ai_platform_test.go`
> If any listed files changed since this plan was written, compare the Current state excerpts against live code before proceeding; on mismatch, stop and report.

## Status

- **Priority**: P1
- **Effort**: S
- **Risk**: MED
- **Depends on**: none
- **Category**: correctness/tests
- **Planned at**: commit `efd46ed`, 2026-06-15
- **Status**: DONE (implemented and verified)

## Why this matters

The AI production GitOps plan now includes the NVIDIA NIC Configuration Operator and the unified AI platform console, but the NVIDIA AI sandbox validation path does not include those charts. That means `ubiquity test --sandbox-deploy` can pass while skipping first-party AI platform components that should be render-validated before claiming NCP-style readiness.

## Current state

- `cmd/ubiquity/cmd/ai_platform.go` — AI production targets include the new charts:

```go
232| add("gpu-operator", "nvidia-gpu-operator", "system/nvidia-gpu-operator", "gpu-operator")
233| add("nvidia-network-operator", "nvidia-network-operator", "system/nvidia-network-operator", "nvidia-network-operator")
234| add("nvidia-nic-configuration-operator", "nvidia-nic-configuration-operator", "system/nvidia-nic-configuration-operator", "network-operator")
...
237| if profile.HasCapability(aiplatform.CapabilityUnifiedFrontend) {
238|     targets = append(targets, aiPlatformGitOpsTarget{Name: "ai-platform-console", Path: "platform/ai-platform-console", Namespace: "ai-platform"})
239| }
```

- `cmd/ubiquity/cmd/up.go` — sandbox namespace overrides omit both new charts:

```go
415| namespaceOverrides := map[string]string{
416|     "nvidia-gpu-operator":     "gpu-operator",
417|     "nim-operator":            "nim-operator",
418|     "kai-scheduler":           "kai-scheduler",
419|     "ai-workload-tenancy":     "ai-workload-tenancy",
420|     "nvidia-network-operator": "nvidia-network-operator",
421| }
```

- `cmd/ubiquity/cmd/up.go` — NVIDIA AI sandbox filter omits both new charts:

```go
516| func isNvidiaAISandboxDeployTarget(target sandboxDeployTarget) bool {
517|     included := map[string]bool{
518|         "system/nvidia-gpu-operator":     true,
519|         "system/nvidia-network-operator": true,
520|         "platform/nim-operator":          true,
521|         "platform/kai-scheduler":         true,
522|         "platform/ai-workload-tenancy":   true,
523|     }
524|     return included[target.ChartDir]
525| }
```

## Commands you will need

| Purpose | Command | Expected on success |
|---|---|---|
| Focused Go tests | `go test ./cmd/ubiquity/cmd -run 'Test.*(Sandbox|AI|Platform)' -count=1` | exits 0 |
| Full Go package test | `go test ./cmd/ubiquity/cmd -count=1` | exits 0 |
| Render validation | `go run ./cmd/ubiquity test --sandbox-deploy` | includes/render-validates all first-party AI charts and exits 0 |
| Relevant Helm tests | `helm unittest platform/ai-platform-console system/nvidia-nic-configuration-operator platform/ai-workload-tenancy` | exits 0 |

Run from the repo root.

## Scope

**In scope**:
- `cmd/ubiquity/cmd/up.go` for NVIDIA AI sandbox target filtering and namespace overrides.
- `cmd/ubiquity/cmd/up_test.go` and/or `cmd/ubiquity/cmd/ai_platform_test.go` for regression coverage.

**Out of scope**:
- Adding real GPU/RDMA hardware checks to sandbox render validation.
- Changing GitOps target generation semantics outside first-party AI chart parity.
- Modifying chart manifests except if a render bug is discovered.

## Git workflow

- Branch: `improve/002-ai-sandbox-target-sync`
- Commit per logical step using the repo's commit style.
- Do not push or open a PR unless the operator instructed it.

## Steps

### Step 1: Add a failing regression for target parity

Add or extend a unit test that compares first-party AI GitOps target paths from `aiPlatformGitOpsTargets(aiplatform.ProductionProfile())` against `isNvidiaAISandboxDeployTarget`, allowing only documented exceptions for external/hardware-only charts.

The test should fail before the implementation because these are absent:
- `system/nvidia-nic-configuration-operator`
- `platform/ai-platform-console`

**Verify**: focused test fails for the expected missing target(s).

### Step 2: Include the new charts in sandbox target filtering

Update `isNvidiaAISandboxDeployTarget` so it includes:
- `system/nvidia-nic-configuration-operator`
- `platform/ai-platform-console`

**Verify**: focused target-parity test passes.

### Step 3: Add namespace overrides for direct sandbox rendering

Update `namespaceOverrides` so direct render/apply behavior matches the AI GitOps plan:
- `nvidia-nic-configuration-operator` -> `network-operator`
- `ai-platform-console` -> `ai-platform`

**Verify**: add/update a unit test for collected targets showing the expected namespaces.

### Step 4: Run focused render and chart gates

Run `go run ./cmd/ubiquity test --sandbox-deploy`, relevant Go tests, and relevant Helm unittest commands.

**Verify**: all commands exit 0.

## Test plan

- New/updated regression covering AI GitOps target vs NVIDIA AI sandbox target parity.
- New/updated namespace override test for the two new charts.
- `go test ./cmd/ubiquity/cmd -count=1`
- `go run ./cmd/ubiquity test --sandbox-deploy`
- `helm unittest platform/ai-platform-console system/nvidia-nic-configuration-operator platform/ai-workload-tenancy`

## Done criteria

- [ ] NVIDIA AI sandbox filter includes `system/nvidia-nic-configuration-operator`.
- [ ] NVIDIA AI sandbox filter includes `platform/ai-platform-console`.
- [ ] Sandbox namespace overrides match AI GitOps namespaces for those charts.
- [ ] Regression tests fail before and pass after the fix.
- [ ] Focused render and test commands exit 0.
- [ ] No files outside in-scope list are modified unless a STOP condition is resolved with operator approval.

## STOP conditions

Stop and report back if:

- A chart cannot render in sandbox without live hardware dependencies.
- Product intent is to intentionally exclude one of these AI GitOps targets from sandbox validation.
- Existing tests reveal broader sandbox target drift that would expand this plan substantially.

## Maintenance notes

Prefer a shared source of truth for first-party AI chart target paths if possible. A regression that compares generated GitOps targets to sandbox targets is more durable than duplicating another static list without tests.
