# Plan 009: Align AI production GitOps target registry with profile evidence

> **Executor instructions**: Make `ubiquity ai-platform render/apply` honest about what it deploys. Every local production/evidence component must either render an Argo CD Application or have an explicit documented exclusion.
>
> **Drift check (run first)**: `git diff --stat efd46ed..HEAD -- cmd/ubiquity/cmd/ai_platform.go cmd/ubiquity/cmd/ai_platform_test.go pkg/aiplatform docs/admin-guide/nvidia-ai-platform.md`

## Status

- **Priority**: P1
- **Effort**: M
- **Risk**: MED
- **Depends on**: none
- **Category**: deployment-correctness/docs
- **Planned at**: commit `efd46ed`, 2026-06-15
- **Status**: DONE (implemented and verified)

## Why this matters

`ai-production` declares a full NVIDIA-backed profile with virtualization, tenant isolation, and unified frontend evidence. The GitOps renderer must not imply those capabilities are deployed unless it emits Applications or documents a deliberate external deployment path.

## Current state

- `pkg/aiplatform/profile.go:251-255` includes KubeVirt, CDI, Multus, and AI/NICo components in `ai-production`.
- `pkg/aiplatform/ncp_requirements.go:67-74` lists `platform/tenant-kubernetes-cluster` and `platform/tenant-vpc` as tenant evidence.
- `cmd/ubiquity/cmd/ai_platform.go:224-243` emits Applications only for GPU/network/NICo/NIM/KAI/frontend/ai-workload-tenancy.
- `docs/admin-guide/nvidia-ai-platform.md:97-103` lists a stale sandbox target set.

## Commands you will need

| Purpose | Command | Expected on success |
|---|---|---|
| Focused tests | `go test ./cmd/ubiquity/cmd ./pkg/aiplatform -run 'AIPlatform|NvidiaAISandbox|Production' -count=1` | exits 0 |
| Render probe | `go run ./cmd/ubiquity ai-platform --profile ai-production render` | output includes honest Application/exclusion metadata |
| Sandbox docs/target proof | `go test ./cmd/ubiquity/cmd -run TestSandboxDeployTargetsIncludeNvidiaAIComponents -count=1` | exits 0 |

## Scope

**In scope**:
- `cmd/ubiquity/cmd/ai_platform.go`
- `cmd/ubiquity/cmd/ai_platform_test.go`
- `docs/admin-guide/nvidia-ai-platform.md`
- profile/evidence tests under `pkg/aiplatform`

**Out of scope**:
- Creating new full KubeVirt/Tenant VPC charts if they do not exist.

## Steps

1. Create a single registry for AI platform GitOps targets and intentional exclusions.
2. Add tests that all local chart evidence paths in the production profile/requirements are either rendered as Applications or listed as exclusions with reasons.
3. Parameterize `targetRevision` or align it with root GitOps revision instead of hardcoded `HEAD`.
4. Update docs to include `platform/ai-platform-console` and `system/nvidia-nic-configuration-operator` in sandbox-deploy coverage.
5. Render `ai-production` and verify output is explicit about deployed targets and exclusions.

## Done criteria

- [ ] Production render/apply has a tested target/exclusion contract.
- [ ] `targetRevision` is configurable or aligned with root GitOps defaults.
- [ ] Sandbox docs match target registry.
- [ ] Focused tests pass.

## STOP conditions

Stop if a referenced local evidence path does not exist and cannot be represented as an exclusion without weakening the readiness claim; report the missing artifact.
