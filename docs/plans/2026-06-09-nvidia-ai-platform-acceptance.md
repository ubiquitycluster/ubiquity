# NVIDIA AI Platform Acceptance Completion Implementation Plan

> **For Hermes:** Use test-driven-development skill to implement this plan task-by-task.

**Goal:** Close the remaining NVIDIA AI platform acceptance criteria from `docs/GOAL-NVIDIA-AI-PLATFORM.md` with fail-closed, source-backed, verifiable GitOps/readiness/demo paths.

**Architecture:** Strengthen Ubiquity-owned glue around NVIDIA upstream components rather than reimplementing operators. The CLI renders/applies a deterministic ArgoCD GitOps application set for AI platform profiles, readiness collectors require live controller/workload evidence, and gated E2E scripts prove real-GPU serving/scheduling/observability only when explicitly enabled.

**Tech Stack:** Go/Cobra CLI, Helm wrapper charts, ArgoCD Application manifests, bash E2E scripts, Kubernetes/NVIDIA GPU Operator/NIM Operator/KAI Scheduler/NVIDIA Network Operator.

---

## Acceptance mapping

1. GitOps GPU platform deployment path: render/apply ArgoCD Applications from `ubiquity ai-platform --profile ai-production render|apply`.
2. GPU Operator substrate: readiness checks for operator, driver, runtime/toolkit, device plugin, GFD, DCGM exporter, MIG manager, validators, and allocatable GPU/MIG resources.
3. DCGM replacement: remove production fallback to hand-authored `monitoring-system` DCGM exporter; keep legacy charts disabled only.
4. NIM serving: add gated real-GPU NIM smoke script that waits for NIMService and records `nim-smoke-test-passed` only after endpoint success.
5. Health/info readiness: ensure `ubiquity health --ai` and `ubiquity info` expose GPU runtime, operator state, serving, scheduler, and RDMA/network readiness signals.
6. RDMA profile evidence: add gated RDMA smoke script that fails closed and records `rdma-network-smoke-test-passed` only after resource/NAD evidence.
7. Ollama demotion: disable `apps/ollama` by default and document it as diagnostics/local only.
8. Gated GPU E2E: keep skipped-by-default behavior and compose GPU/NIM/RDMA/KAI checks when explicitly enabled.
9. Final demo path: add a gated final demo script that proves provision, reconcile, schedule, serve, observe, and validate phases.

## Tasks

### Task 1: Plan artifact
- Create this plan file.
- Commit with `docs: plan NVIDIA AI platform acceptance completion`.

### Task 2: GPU substrate readiness granularity
- Add failing tests in `pkg/aiplatform/readiness_test.go` requiring separate check names for `gpu-driver`, `gpu-runtime-toolkit`, `gpu-feature-discovery`, `gpu-dcgm-exporter`, `gpu-mig-manager`, and `gpu-validator`.
- Add booleans to `aiplatform.ClusterSnapshot` and checks to `EvaluateReadiness`.
- Update `cmd/ubiquity/cmd/health.go` collector to derive these booleans from GPU Operator-managed daemonsets/pods/services.
- Verify targeted and full Go tests.

### Task 3: GitOps application render/apply path
- Add tests in `cmd/ubiquity/cmd/ai_platform_test.go` requiring `renderAIPlatformManifest(ai-production)` to include ArgoCD `Application` resources for GPU Operator, NVIDIA Network Operator, NIM Operator, KAI Scheduler, and AI tenancy.
- Extend `renderAIPlatformManifest` to render deterministic ArgoCD Applications alongside the profile ConfigMap.
- Verify CLI render output and tests.

### Task 4: DCGM legacy demotion
- Add tests ensuring health/E2E no longer accept `monitoring-system/dcgm-exporter` as production readiness evidence.
- Update health collector and GPU E2E script to require `gpu-operator/nvidia-dcgm-exporter` evidence.
- Keep legacy chart gated/disabled only.

### Task 5: NIM and RDMA smoke scripts
- Add tests requiring gated scripts:
  - `test/e2e/nim-gpu-serving-smoke.sh`
  - `test/e2e/nvidia-rdma-smoke.sh`
- Scripts must skip by default, run only with explicit env flags, and record evidence ConfigMaps only after live checks.
- Update `test/e2e/nvidia-ai-platform.sh` to compose these scripts when enabled.

### Task 6: Ollama diagnostics-only default
- Add tests requiring `apps/ollama/values.yaml` to be disabled by default and docs/CLI to say diagnostics only.
- Change default values if necessary.

### Task 7: Final acceptance demo path
- Add `test/e2e/nvidia-ai-final-demo.sh`, gated by `UBIQUITY_RUN_NVIDIA_AI_FINAL_DEMO=true`.
- Script phases: provision, reconcile, schedule, serve, observe, validate.
- It should call the GitOps apply path dry-run/server-side proof, GPU E2E, NIM, RDMA, KAI, health/info, and write a final evidence marker only after all required live steps pass.

### Task 8: Docs and verification
- Update `docs/admin-guide/nvidia-ai-platform.md` and `docs/GOAL-NVIDIA-AI-PLATFORM.md` acceptance notes to point to the concrete commands and evidence boundaries.
- Run:
  - `go test ./pkg/... ./cmd/... -count=1`
  - active Helm lint/render over all non-disabled charts
  - bash syntax checks and dry-run/default-skip checks for all E2E scripts
  - fail-closed `ubiquity health --ai`, `ubiquity health --aistore`, and VM readiness smoke
  - `git diff --check`
  - `graphify update .` when available
- Commit final verified implementation.
