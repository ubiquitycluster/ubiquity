# NVIDIA AI Platform operations guide

This guide explains how Ubiquity's NVIDIA-backed AI workload platform is intended to be used, what it deploys, how the sandbox proof works, and what must be true before the platform can claim it is ready to serve AI workloads.

This guide does not claim NVIDIA approval or certification. Ubiquity uses NVIDIA-maintained repositories and charts as source-backed components where they are stronger than local manifests, but readiness is still proven by Ubiquity validation and by the state of the target cluster.

## Quick start

Use this path when you want the fastest local proof that the NVIDIA AI platform manifests are coherent before touching a real GPU cluster.

```sh
# 1. Inspect the desired profile and upstream component provenance.
ubiquity ai-platform --profile ai-production

# 2. Render the NVIDIA AI sandbox deploy set without requiring NVIDIA devices.
ubiquity test --sandbox-deploy

# 3. Optional: create a local k3d sandbox and apply the sandbox-safe NVIDIA AI control planes.
ubiquity up --sandbox

# 4. On a real GPU cluster, check readiness after GitOps/platform reconciliation.
ubiquity health
ubiquity info
```

Expected local sandbox result:

- `ubiquity test --sandbox-deploy` should print `NVIDIA AI sandbox deploy components rendered cleanly`.
- A live k3d sandbox can deploy the NVIDIA control planes and CRDs, but it must not claim GPU workload readiness because k3d normally has no NVIDIA devices.
- Real serving readiness requires a GPU node, NVIDIA runtime/device evidence, telemetry evidence, NIM service readiness, and a smoke test.

## Mental model

Ubiquity owns the platform profile, values, orchestration order, validation, GitOps glue, and fail-closed readiness policy.

NVIDIA-maintained repositories own the specialized GPU, networking, scheduling, telemetry, and serving behavior where they are stronger than Ubiquity-local code.

The intended division is:

1. Ubiquity profile selected with `ubiquity ai-platform --profile <profile>`.
2. Profile maps to source-backed NVIDIA components and versions.
3. Ubiquity wrapper charts provide deterministic values and sandbox-safe overrides.
4. GitOps or sandbox deployment reconciles the charts.
5. Readiness commands verify cluster evidence and fail closed when evidence is missing.

Ubiquity should not reimplement NVIDIA operator internals. It should generate/validate the configuration around those operators and report whether the deployed platform is actually ready.

## Supported NVIDIA components

| Role | Source repository | Ubiquity integration | Replacement decision |
| --- | --- | --- | --- |
| GPU substrate | NVIDIA/gpu-operator | `system/nvidia-gpu-operator` wrapper chart. Primary path for driver, container runtime, device plugin, GPU feature discovery, MIG Manager, validator, and DCGM exporter. | Replaces missing or bespoke GPU enablement. |
| DCGM telemetry | NVIDIA/dcgm-exporter | Managed through NVIDIA GPU Operator by default. Legacy hand-authored DCGM manifests are gated behind `legacyDcgmExporter.enabled=false`. | Replaces stale hand-authored DCGM DaemonSets. |
| RDMA/networking | NVIDIA/network-operator (formerly Mellanox/network-operator) / NVIDIA Network Operator distribution | `system/nvidia-network-operator` wrapper chart. Production values model `nvidia.com/rdma` resources, NetworkAttachmentDefinitions, and NIC policy. Sandbox values deploy control plane/CRDs only. | Replaces static RDMA/secondary-network glue where RDMA profiles are enabled. |
| Production serving | NVIDIA/k8s-nim-operator | `platform/nim-operator` wrapper chart. Adds NIM CRDs and operator. Sample NIM services require NGC secret references. | Replaces Ollama as production inference default. |
| AI workload scheduler | NVIDIA/KAI-Scheduler | `platform/kai-scheduler` wrapper chart using OCI chart `oci://ghcr.io/kai-scheduler/kai-scheduler` at `v0.10.2`. | Replaces local priority/quota-only scheduling for production GPU workloads once profile readiness is proven. |
| AI runtime recipes | NVIDIA/aicr | Profile metadata records AICR as the preferred recipe source for validated GPU AI runtime manifests. | Evaluation/adoption path for recipe-backed runtime manifests. |
| Data-plane option | NVIDIA/aistore and NVIDIA/ais-k8s | Optional AIStore evaluation path. AIStore replaces Longhorn for AI dataset/cache paths when object/S3 semantics fit, but it is not a generic PVC replacement. | Prefer AIStore for model artifacts, checkpoints, training/inference datasets, sharded archives, remote bucket acceleration, and GPU-adjacent cache/object access; retain Longhorn only for generic PVCs until a stronger POSIX/RWX option is selected. |
| Bare-metal reference | NVIDIA/deepops | Source-backed reference for DGX/GPU cluster practices, Kubespray, Slurm paths, OS support, and validation procedures. | Reference, not a drop-in replacement for Ubiquity sandbox/Cluster API flow. |
| Cloud native reference | NVIDIA/cloud-native-stack | Source-backed component matrix and PoC reference for NVIDIA Cloud Native Stack batches. | Reference; do not treat its basic non-HA install as production certification. |

## Profiles

Inspect a profile with:

```sh
ubiquity ai-platform --profile sandbox
ubiquity ai-platform --profile gpu-basic
ubiquity ai-platform --profile gpu-rdma
ubiquity ai-platform --profile gpu-mig
ubiquity ai-platform --profile ai-production
```

Profile meanings:

| Profile | Purpose | Readiness claim |
| --- | --- | --- |
| `sandbox` | CPU-only/local development path. Used to prove charts and control planes render/apply without NVIDIA devices. | Makes no NVIDIA hardware readiness claim. |
| `gpu-basic` | GPU Operator, DCGM telemetry, NIM serving path, and basic fail-closed health checks. | Ready only when GPU Operator, runtime, device plugin, GPU resources, DCGM, and NIM evidence are present. |
| `gpu-rdma` | `gpu-basic` plus Network Operator/RDMA path. | Ready only when `nvidia.com/rdma` allocatable resources, NetworkAttachmentDefinition objects, and the `rdma-network-smoke-test-passed` marker are proven. |
| `gpu-mig` | GPU Operator with MIG-aware validation and KAI Scheduler evaluation. | Ready only when MIG profiles and allocatable `nvidia.com/mig-*` resources match the profile. MIG-partitioned clusters can satisfy accelerator capacity with MIG resources even when `nvidia.com/gpu` is absent. |
| `ai-production` | Full NVIDIA-backed platform profile: GPU, RDMA, NIM serving, telemetry, KAI Scheduler, tenancy, optional data-plane evaluation, and gated E2E. | Ready only after end-to-end evidence proves provision, reconcile, schedule, serve, observe, and validate. |

The profile output includes component source repositories, chart repositories, replacement decisions, and bare-metal orchestration alternatives.

## What sandbox deploy proves

Run:

```sh
ubiquity test --sandbox-deploy
```

This is a deterministic, hardware-free check. It proves that the NVIDIA AI sandbox deploy set can be dependency-resolved and rendered by Helm without physical NVIDIA devices.

It currently covers:

- `system/nvidia-gpu-operator`
- `system/nvidia-network-operator`
- `platform/nim-operator`
- `platform/kai-scheduler`
- `platform/ai-workload-tenancy`

It proves:

- Chart dependencies can be fetched from their configured repositories.
- Helm templates render valid YAML with CRDs included.
- Sandbox-safe values are selected where required.
- NGC credentials are referenced as secrets and are not embedded.
- Generated `Chart.lock` and `.tgz` dependency artifacts are cleaned after validation.

It does not prove:

- NVIDIA drivers loaded on a host.
- GPUs are visible to Kubernetes.
- Containers can run `nvidia-smi`.
- RDMA devices are present.
- A NIM model can pull from NGC or serve inference.
- The platform is NVIDIA certified or approved.

## How sandbox deploy works

`ubiquity test --sandbox-deploy` uses the same NVIDIA AI target registry that sandbox deployment uses, but it only renders charts. The flow is:

1. Discover sandbox deploy targets under `system/`, `monitoring/`, `platform/`, and `apps/`.
2. Filter to the NVIDIA AI platform target set.
3. For each target, read `Chart.yaml` dependencies.
4. Add non-OCI, non-file Helm repositories as needed.
5. Run `helm dependency build`.
6. Run `helm template --include-crds --namespace <namespace> release <chart>`.
7. If a chart has `values-sandbox.yaml`, include it with `--values <chart>/values-sandbox.yaml`.
8. Remove generated Helm dependency archives and lock files.

Live sandbox apply with `ubiquity up --sandbox` uses k3d when needed and applies the sandbox-safe charts to a local cluster. If no cluster is reachable, sandbox chart apply skips cleanly instead of hanging.

Important sandbox-specific behavior:

- `system/nvidia-network-operator/values-sandbox.yaml` disables hardware-specific secondary CNI/RDMA operands that can break k3d networking. It deploys the control plane and CRDs only.
- `platform/nim-operator` is applied with an explicit namespace because some rendered namespaced resources do not include namespace metadata.
- `platform/kai-scheduler` uses `kubectl apply --server-side --force-conflicts` because KAI Scheduler CRDs exceed the Kubernetes client-side last-applied annotation size limit.

## Production deployment flow

Use this flow for a real GPU environment. Do not skip readiness checks.

1. Confirm hardware and cluster prerequisites:
   - GPU nodes are present.
   - Kubernetes version is compatible with the selected NVIDIA components.
   - Container runtime can support NVIDIA runtime configuration.
   - NGC access is available if NIM workloads will be deployed.
   - RDMA/NIC prerequisites are available if using `gpu-rdma` or `ai-production`.

2. Inspect the selected profile:

   ```sh
   ubiquity ai-platform --profile ai-production
   ```

3. Supply secrets outside Git:
   - NGC API key secret.
   - NGC pull secret.
   - Any model-specific or registry credentials.

4. Reconcile the substrate:
   - `system/nvidia-gpu-operator`
   - `system/nvidia-network-operator` for RDMA profiles

5. Reconcile platform services:
   - `platform/nim-operator`
   - `platform/kai-scheduler`
   - `platform/ai-workload-tenancy`

6. Deploy model-serving resources only after prerequisites exist:
   - NGC secrets.
   - GPU nodes with allocatable GPU or MIG resources.
   - NIM operator CRDs and controller ready.

7. Run readiness checks:

   ```sh
   ubiquity health
   ubiquity info
   ```

8. Run gated GPU E2E only on an appropriate GPU cluster:

   ```sh
   UBIQUITY_RUN_GPU_E2E=true test/e2e/nvidia-ai-platform.sh
   ```

   The NIM smoke path calls the configured NIM endpoint and writes the `nim-smoke-test-passed` ConfigMap only after the endpoint responds successfully. The readiness collector treats that ConfigMap as evidence that serving was tested rather than merely installed.

## NGC credentials

NGC credentials must never be committed to Git.

Create Kubernetes secrets out-of-band through a secret manager, SOPS workflow, or manual cluster-scoped operation. Then reference existing secrets from `platform/nim-operator/values.yaml` with:

- `ngc.existingApiKeySecret`
- `ngc.existingPullSecret`

The chart intentionally fails when `ngc.commitCredentials=true`. This prevents accidental credential material from entering GitOps manifests.

Safe pattern:

```sh
kubectl -n nim-operator create secret generic ngc-api \
  --from-literal=NGC_API_KEY='[REDACTED]'

kubectl -n nim-operator create secret docker-registry ngc-pull \
  --docker-server=nvcr.io \
  --docker-username='$oauthtoken' \
  --docker-password='[REDACTED]'
```

Do not paste real secret values into repository files, plans, issue text, or generated docs.

## KAI Scheduler usage

KAI Scheduler is the NVIDIA-backed path for production AI scheduling semantics such as queues, fairness, gang scheduling, topology-aware placement, and GPU-aware batch/inference scheduling.

Ubiquity integrates it through:

- `platform/kai-scheduler/Chart.yaml`
- `platform/kai-scheduler/values.yaml`
- `platform/kai-scheduler/values-sandbox.yaml`

Sandbox apply uses:

```sh
kubectl apply --server-side --force-conflicts
```

Reason: KAI Scheduler CRDs are large enough that client-side apply can fail with an annotation size error. Server-side apply avoids storing the large last-applied annotation and is required for reliable sandbox/live validation.

In production, KAI Scheduler readiness is not only controller rollout. A production readiness check must also prove that workloads can schedule through the intended queues and that GPU resources are visible to the scheduler.

## Bare-metal orchestration alternatives

Ubiquity considers NVIDIA bare-metal repositories, but it separates replacement candidates from references.

- `NVIDIA/deepops`
  - Use as a reference for DGX/GPU cluster operational practices, Ansible/Kubespray, Slurm paths, OS assumptions, and validation ideas.
  - Do not treat as a drop-in replacement for Ubiquity's k3d sandbox or future Cluster API/Metal3 flow.

- `NVIDIA/cloud-native-stack`
  - Use as a reference matrix for NVIDIA Cloud Native Stack component combinations and PoC install ordering.
  - Do not claim production HA readiness solely because a Cloud Native Stack PoC path exists.

- `NVIDIA/gpu-operator`, `Mellanox/network-operator`, `NVIDIA/k8s-nim-operator`, and `NVIDIA/KAI-Scheduler`
  - Treat as adoption/replacement candidates because they directly own stronger platform functionality than local Ubiquity manifests.

## Validation

Local deterministic checks:

```sh
go test ./pkg/aiplatform ./cmd/ubiquity/cmd -v
go run ./cmd/ubiquity test --sandbox-deploy
helm lint system/nvidia-gpu-operator
helm lint system/nvidia-network-operator
helm lint platform/nim-operator
helm lint platform/kai-scheduler
helm lint platform/ai-workload-tenancy
```

Full local Go check:

```sh
go test ./pkg/... ./cmd/... -count=1
go build ./cmd/ubiquity/...
```

Real GPU-node E2E is gated:

```sh
UBIQUITY_RUN_GPU_E2E=true test/e2e/nvidia-ai-platform.sh
```

The E2E script checks GPU nodes, GPU Operator rollout, NVIDIA device plugin rollout, `nvidia-smi`, DCGM service/metrics evidence, NIM operator pods, NIM smoke-test evidence, KAI Scheduler controller rollouts, `default-queue` evidence, a pod scheduled with `spec.schedulerName: kai-scheduler`, and the `kai-scheduling-smoke-test-passed` marker ConfigMap.

## Readiness and fail-closed behavior

The platform is not ready unless evidence exists for every capability claimed by the selected profile.

Examples:

- Missing GPU nodes: not ready.
- GPU Operator not reconciled: not ready.
- Device plugin not healthy: not ready.
- No allocatable `nvidia.com/gpu` or expected MIG resources: not ready.
- DCGM metrics absent: not ready.
- RDMA resources missing for RDMA profiles: not ready.
- NIM operator ready but no model smoke test: not ready for serving claims.
- KAI Scheduler controllers ready but no queue/scheduling proof: not ready for production scheduling claims.

This fail-closed policy is intentional. Ubiquity should report exactly what evidence is missing instead of implying the cluster can serve AI workloads.

## Troubleshooting

Use these checks when sandbox or production validation fails.

Check profile output:

```sh
ubiquity ai-platform --profile ai-production
```

Check deterministic sandbox rendering:

```sh
ubiquity test --sandbox-deploy
```

Check cluster access:

```sh
kubectl cluster-info
kubectl get nodes -o wide
```

Check NVIDIA operator namespaces:

```sh
kubectl -n gpu-operator get pods,deploy
kubectl -n nvidia-network-operator get pods,deploy
kubectl -n nim-operator get pods,deploy
kubectl -n kai-scheduler get pods,deploy
```

Check NVIDIA CRDs:

```sh
kubectl get crd clusterpolicies.nvidia.com
kubectl get crd nicclusterpolicies.mellanox.com
kubectl get crd nimservices.apps.nvidia.com
kubectl get crd configs.kai.scheduler
kubectl get crd queues.scheduling.run.ai
kubectl get crd podgroups.scheduling.run.ai
```

If Network Operator breaks k3d networking, confirm sandbox values were used:

```sh
helm template --include-crds --namespace nvidia-network-operator release system/nvidia-network-operator \
  --values system/nvidia-network-operator/values-sandbox.yaml
```

If KAI Scheduler CRDs fail with an annotation-size error, use server-side apply:

```sh
kubectl apply --server-side --force-conflicts -f <rendered-kai-scheduler.yaml>
```

If NIM resources appear in `default` instead of `nim-operator`, force the namespace during apply:

```sh
kubectl apply -n nim-operator -f <rendered-nim-operator.yaml>
```

## Unsupported claims

No NVIDIA approval or certification is implied.

Do not claim the platform is ready to serve AI workloads unless the selected profile's readiness checks have passed on the target cluster.

Do not claim a CPU-only k3d sandbox proves GPU runtime, RDMA, NIM model serving, or production scheduling behavior. It proves only that the sandbox-safe control-plane manifests render and can be applied without NVIDIA hardware.
