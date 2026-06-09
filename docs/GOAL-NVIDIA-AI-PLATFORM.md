# Goal: NVIDIA-backed AI Workload Platform

## Current-state assessment

Ubiquity is a broad Kubernetes/HPC automation framework, not yet a complete NVIDIA AI workload platform.

What exists now:
- A Go CLI with a six-phase deployment flow: metal, bootstrap, security, external, wait, post-install.
- GitOps bootstrap through ArgoCD and a root app-of-apps pattern.
- A large IaC surface: Helm charts, Terraform cloud modules, Ansible bare-metal roles, Metal3/BMO-oriented provisioning, monitoring, storage, identity, and workload-manager components.
- A working shared `charts/app-template` pattern and an `apps/ollama` chart for local LLM serving.
- Some NVIDIA-adjacent pieces: `system/nvidia-network-operator`, hand-authored DCGM exporter DaemonSets, Slurm GPU accounting snippets, and GPU-related monitoring labels.
- A `ubiquity ai` command that calls a local Ollama endpoint for cluster-state troubleshooting.

Main gaps for serving AI workloads:
- No first-class NVIDIA GPU Operator stack. GPU driver lifecycle, CUDA/container runtime wiring, device plugin, MIG management, GPU feature discovery, validator, and DCGM should be managed by NVIDIA GPU Operator rather than bespoke or partial manifests.
- DCGM exporter is hand-authored and pinned to an old image. It should be replaced by the exporter and ServiceMonitor behavior provided through NVIDIA GPU Operator or NVIDIA dcgm-exporter chart/config.
- NVIDIA Network Operator integration exists, but it is narrow and static: RDMA resource naming, NIC policy values, Multus/whereabouts assumptions, and fabric validation are not modeled as a platform capability.
- Ollama is the only AI serving workload. It is useful for local diagnostics, but it is not an NVIDIA-optimized inference platform and should not be the default production serving plane.
- No NIM/NeMo microservice operator integration, model-serving CRDs, model profiles, NGC secret handling, or GPU-aware serving validation.
- No NVIDIA AI runtime recipe layer that can generate validated GPU Operator, network, storage, observability, and serving manifests from one platform profile.
- No AI-specific scheduler, queueing, gang scheduling, topology-aware placement, or quota model for multi-user GPU workloads.
- No end-to-end validation that proves: GPUs are visible, drivers are healthy, container runtime is configured, RDMA is available where expected, inference workloads can schedule, and DCGM metrics are scraped.
- No production-grade operational UX for GPU platform health: `ubiquity health` only checks kubectl and ArgoCD; it does not validate GPU nodes, operators, NIM services, MIG profiles, RDMA resources, or model-serving endpoints.

## Target goal

Transform Ubiquity from a general HPC/Kubernetes GitOps framework into a fully stand-up-able NVIDIA AI workload platform by adopting NVIDIA-maintained Kubernetes components wherever they are stronger than Ubiquity's local implementation, replacing bespoke GPU/runtime/serving/telemetry manifests with source-backed NVIDIA repositories and charts, and exposing the whole stack as a small set of declarative Ubiquity AI platform profiles.

The resulting platform must be able to provision a Kubernetes cluster, install the NVIDIA GPU and network substrate, deploy validated AI runtime manifests, serve production inference workloads, schedule GPU-intensive batch/training jobs, collect GPU telemetry, and fail closed when NVIDIA platform readiness is not proven.

## NVIDIA repositories/components to evaluate and integrate

Use components from https://github.com/orgs/NVIDIA/repositories as the default source of truth when they provide better functionality than Ubiquity-local code:

1. NVIDIA/gpu-operator
   - Replace missing/bespoke GPU enablement with GPU Operator-managed drivers, container toolkit/runtime wiring, device plugin, GPU feature discovery, DCGM exporter, validator, MIG Manager, and related CRDs.
   - Ubiquity should own values/profile generation and validation, not reimplement operator internals.

2. NVIDIA/k8s-device-plugin
   - Do not maintain custom GPU device-plugin manifests. Use this directly only where GPU Operator is intentionally disabled; otherwise consume it via GPU Operator.

3. NVIDIA/dcgm-exporter
   - Replace hand-authored, stale DCGM DaemonSets in `system/monitoring-system` and `monitoring/monitoring-system` with NVIDIA-managed DCGM exporter configuration through GPU Operator or the upstream chart/config.

4. NVIDIA/k8s-nim-operator
   - Add a first-class NIM serving layer for NVIDIA NIM and NeMo microservices, with CRDs, NGC secret handling, model profiles, GPU resource requests, ingress, autoscaling, readiness checks, and smoke tests.
   - This should replace Ollama as the production inference default. Keep Ollama only as a lightweight optional diagnostics/local-lab app.

5. NVIDIA/aicr
   - Use AICR as the preferred recipe/manifest source for optimized, validated, reproducible GPU-accelerated AI runtime on Kubernetes.
   - Ubiquity should translate high-level platform profiles into AICR recipes/manifests, or vendor/consume generated outputs with provenance metadata.

6. NVIDIA/KAI-Scheduler
   - Replace priority/quota-only local scheduling for production AI workloads with GPU-aware fairness, gang scheduling, queue hierarchy, topology-aware placement, and DRA-aware scheduling where validated.
   - Add a wrapper chart and sandbox deployment proof before treating it as production-ready.

7. NVIDIA/deepops
   - Use as the source-backed bare-metal orchestration reference for DGX/GPU cluster practices, Kubernetes via Kubespray, Slurm paths, supported OS assumptions, and validation procedures.
   - Do not treat it as a drop-in replacement for Ubiquity's k3d/Cluster API sandbox path; consume its validated practices and document the handoff boundary.

8. NVIDIA/cloud-native-stack
   - Use as the source-backed component matrix and PoC reference for NVIDIA Cloud Native Stack batches: Kubernetes, CNI, GPU Operator, Network Operator, NIM Operator, KAI Scheduler, monitoring, and load balancer versions.
   - Its basic/non-HA Kubernetes installation guidance is PoC-oriented, so production orchestration should retain Ubiquity's fail-closed profile validation rather than claim CNS production certification.

9. NVIDIA/k8s-dra-driver or DRA-related NVIDIA components, if mature enough for the target Kubernetes versions
   - Add an optional next-generation resource allocation path for advanced GPU sharing/topology cases while keeping GPU Operator/device-plugin as the stable default.

10. NVIDIA/ais-k8s and NVIDIA/aistore
   - Evaluate AIStore as an AI data plane option for high-performance object storage/cache for training and inference datasets.
   - Integrate only when it is a better fit than the current generic storage choices for AI workload paths.

11. NVIDIA/nvidia-terraform-modules
   - Evaluate replacing or augmenting bespoke cloud GPU Kubernetes Terraform with NVIDIA-maintained GPU cluster infrastructure modules where they provide stronger GPU-specific correctness.

12. NVIDIA/k8s-launch-kit, NVIDIA/kubectl-nv, NVIDIA/container-canary, NVIDIA/knavigator, NVIDIA/knetscan, and NVIDIA/ISV-NCP-Validation-Suite
   - Evaluate for deployment workflows, admin UX, container/runtime validation, scheduling validation, network reachability, and lab/platform validation.
   - Integrate tools only when they materially improve correctness or operator experience versus Ubiquity-local logic.

## Required platform capabilities

### 1. Declarative AI platform profiles

Add `ubiquity ai-platform` profile support with at least these profiles:
- `sandbox`: CPU-only or simulated path; no NVIDIA readiness claims.
- `gpu-basic`: GPU Operator, DCGM metrics, one NIM sample service, and basic health checks.
- `gpu-rdma`: GPU Operator plus NVIDIA Network Operator/RDMA validation.
- `gpu-mig`: GPU Operator with MIG profiles and scheduling validation.
- `ai-production`: GPU substrate, network substrate, NIM serving, observability, storage/data plane, security, and E2E validation.

Profiles must produce deterministic Helm values/ArgoCD Applications and must record component source, version, chart repository, and validation status.

### 2. Replace local/bespoke functionality when NVIDIA-maintained components are better

Replacement targets:
- Replace hand-authored DCGM exporter YAML with NVIDIA-managed DCGM exporter integration.
- Replace any standalone/custom GPU device-plugin path with GPU Operator or k8s-device-plugin-backed profiles.
- Replace Ollama as production AI serving with NIM Operator-backed services; retain Ollama only as optional local diagnostics.
- Replace ad hoc NVIDIA Network Operator values with profile-driven, validated NIC/RDMA policy generation.
- Replace generic GPU health checks with NVIDIA validator/container-canary/kubectl-nv-compatible checks where available.

### 3. AI serving plane

Add a NIM/NVIDIA serving layer that can:
- Create NGC pull-secret/config secrets safely without committing credentials.
- Deploy at least one sample NIM service through k8s-nim-operator.
- Expose service endpoints through existing ingress and identity patterns.
- Apply GPU resource requests/limits, node selectors, tolerations, and optional MIG profile constraints.
- Support smoke tests that submit a request to the served model and assert a successful response.

### 4. GPU scheduler and multi-tenant workload readiness

Add GPU-aware workload support for:
- Batch/training jobs with Slurm/Kubernetes coexistence semantics documented.
- GPU quotas, priority classes, taints/tolerations, and namespace-level tenancy.
- Optional topology-aware scheduling/gang scheduling evaluation using NVIDIA-relevant scheduler tooling where appropriate.
- Clear separation between inference services, batch jobs, and platform/system DaemonSets.

### 5. Observability and health

Extend `ubiquity health` and `ubiquity info` to report:
- GPU Operator phase and operand health.
- Driver/container-runtime/device-plugin readiness.
- GPU allocatable/capacity per node.
- MIG configuration and allocatable slices where enabled.
- DCGM metrics scrape status.
- RDMA resources and NetworkAttachmentDefinitions where enabled.
- NIM service readiness and last smoke-test result.

Fail closed: if GPU resources or NVIDIA operators are not healthy, the command must report non-ready status and must not claim the platform can serve AI workloads.

### 6. Validation and CI

Add validation tiers:
- Unit tests for profile-to-values rendering.
- Helm template/golden tests for NVIDIA stack charts.
- K3d/KWOK or mock tests for ArgoCD/Application generation.
- Real GPU-node E2E tests gated behind an explicit environment flag.
- Runtime smoke tests for `nvidia-smi`, container runtime GPU access, DCGM scrape, RDMA resources, and a NIM inference request.

### 7. Documentation and provenance

Document:
- Supported NVIDIA component versions and chart repositories.
- Platform profiles and required hardware/network prerequisites.
- How to supply NGC credentials and model-specific configuration.
- When Ubiquity uses NVIDIA upstream behavior versus Ubiquity-generated glue.
- Known limitations and unsupported claims.

No generated docs or CLI output may claim NVIDIA approval/certification unless explicit approval evidence is attached.

## Implementation status

Current implementation status for this goal:

1. End-to-end GitOps deployment path for a GPU-enabled Kubernetes AI platform is implemented through `ubiquity ai-platform render/apply --profile ai-production`, which emits ArgoCD `Application` resources for the NVIDIA GPU Operator, NVIDIA Network Operator, NIM Operator, KAI Scheduler, and AI workload tenancy charts.
2. GPU Operator is the primary GPU substrate path. Readiness now fails closed on individual evidence for the driver, runtime/toolkit, device plugin, GPU Feature Discovery, GPU Operator managed DCGM exporter, MIG Manager, validators, and allocatable GPU or MIG resources.
3. Hand-authored DCGM exporter paths are no longer accepted as production readiness evidence. Production telemetry evidence must come through the NVIDIA GPU Operator managed `gpu-operator/nvidia-dcgm-exporter` service.
4. NIM Operator-backed serving is proven by the gated `test/e2e/nim-gpu-serving-smoke.sh` path, which requires a real GPU node, waits for a `NIMService`, calls the endpoint with `curl --fail`, and records `nim-smoke-test-passed` only after success.
5. `ubiquity health --ai` and `ubiquity info --ai` report fail-closed readiness signals for GPU runtime, operator state, NIM serving readiness, KAI scheduler readiness, and NVIDIA Network Operator / RDMA readiness.
6. NVIDIA Network Operator / RDMA profile readiness is fail-closed through allocatable `nvidia.com/rdma`, `NetworkAttachmentDefinition`, and `rdma-network-smoke-test-passed` evidence.
7. Ollama is diagnostics-only. The chart is disabled by default and the `ubiquity ai` command describes it as local diagnostics rather than production AI serving.
8. Gated real-GPU E2E remains skipped by default and runs only when explicitly enabled. Focused proofs are available for NIM, RDMA, KAI scheduling, and the composed GPU E2E path.
9. The final demo path is `UBIQUITY_RUN_NVIDIA_AI_FINAL_DEMO=true test/e2e/nvidia-ai-platform-final-demo.sh`; it proves provision, reconcile, schedule, serve, observe, and validate in order and records `nvidia-ai-final-demo-passed` only after all stages pass.

## Acceptance criteria

The goal is complete when:

1. A user can run a documented Ubiquity command/profile and deploy a GPU-enabled Kubernetes AI platform through GitOps.
2. NVIDIA GPU Operator is the primary GPU enablement path and validates drivers, runtime, device plugin, DCGM, MIG, and GPU node readiness.
3. Hand-authored DCGM exporter manifests are removed or disabled in favor of NVIDIA-managed exporter integration.
4. A NIM Operator-backed inference service can be deployed and smoke-tested successfully on a GPU node.
5. `ubiquity health` and `ubiquity info` provide AI platform readiness signals and fail closed on missing GPU/runtime/operator/serving prerequisites.
6. The NVIDIA Network Operator/RDMA path is profile-driven and validated where enabled.
7. Ollama remains available only as an optional local/diagnostic app and is no longer presented as the production AI serving layer.
8. CI includes deterministic rendering tests, profile validation tests, and gated GPU E2E tests.
9. Documentation explains the NVIDIA component source, version, role, replacement decision, and operational procedure.
10. The platform can demonstrate end-to-end AI workload service readiness: provision, reconcile, schedule, serve, observe, and validate.

## Suggested implementation phases

P0 - Inventory and source-of-truth decisions
- Map every current GPU/AI/RDMA/serving/telemetry manifest to either NVIDIA upstream replacement, Ubiquity glue, or removal.
- Write ADRs for GPU Operator, NIM Operator, AICR recipe usage, DCGM replacement, and Ollama demotion.

P1 - NVIDIA substrate
- Add GPU Operator profile/chart integration.
- Replace DCGM exporter manifests.
- Extend health/info commands for GPU substrate readiness.
- Add rendering and mocked readiness tests.

P2 - AI serving plane
- Add k8s-nim-operator integration.
- Add NGC secret flow, model profile values, ingress, and smoke tests.
- Demote Ollama to optional diagnostics.

P3 - Network/data/scheduler maturity
- Harden NVIDIA Network Operator/RDMA profile handling.
- Evaluate AIStore and NVIDIA Terraform modules for replacement/augmentation.
- Add scheduling, quota, tenancy, and topology-aware workload tests.

P4 - Production validation
- Add gated real-GPU E2E jobs.
- Add docs, provenance metadata, and readiness reports.
- Publish a demonstration path that proves the platform can serve AI workloads without unsupported claims.
