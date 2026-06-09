# Core Services Capability Parity Implementation Plan

> **For Hermes:** Use test-driven-development skill to implement this plan task-by-task.

**Goal:** Add a local, verifiable core-services deployment capability that covers the useful infrastructure functions found in the reference repository while preserving Ubiquity's existing per-component Helm/GitOps model.

**Architecture:** Do not copy a monolithic chart. Ubiquity already has a demonstrably better model for reviewability: one chart per component plus GitOps orchestration. Add a small `system/core-services` chart that renders ArgoCD `Application` objects for local component charts and vetted public upstream charts only when Ubiquity does not already carry a local wrapper. Do not add Flux or any proprietary repository dependency. Keep storage defaults on existing Longhorn/local-path choices and make backup/live readiness fail closed when required values are missing.

**Tech Stack:** Helm, ArgoCD Application CRs, Go docs-as-code/contract tests, helm-unittest, Bash dry-run proof script.

---

### Task 1: Add executable capability contract

**Objective:** Convert the investigated capability map into tests before implementing the chart.

**Files:**
- Create: `pkg/cloud/core_services_contract_test.go`

**Verification:**
`go test ./pkg/cloud -run TestCoreServicesCapabilityContract -count=1`

Expected first result: FAIL because `system/core-services` does not exist.

### Task 2: Implement core-services GitOps orchestration chart

**Objective:** Render Application resources for core components without Flux or proprietary repository dependencies.

**Files:**
- Create: `system/core-services/Chart.yaml`
- Create: `system/core-services/values.yaml`
- Create: `system/core-services/templates/_helpers.tpl`
- Create: `system/core-services/templates/applications.yaml`
- Create: `system/core-services/tests/applications_test.yaml`

**Components:**
- Existing/local path apps: cert-manager, cilium, external-secrets, longhorn, network-policies, kyverno, kyverno-policies, falco, monitoring-system, ingress-nginx.
- Public upstream chart apps when local wrappers do not already exist: metrics-server, node-feature-discovery, node-problem-detector, snapshot-controller, velero, vertical-pod-autoscaler, kubescape, local-path-provisioner.
- Explicit exclusions: Flux and any repository needing a private/proprietary product.

**Verification:**
`go test ./pkg/cloud -run TestCoreServicesCapabilityContract -count=1`
`helm lint system/core-services`
`helm template core-services system/core-services --namespace argocd`
`helm unittest system/core-services`

### Task 3: Add CI-safe proof script

**Objective:** Prove the bundle renders all expected capabilities, forbids Flux, and requires Velero backup configuration before enabling it.

**Files:**
- Create: `test/e2e/core-services-proof.sh`
- Modify: `.github/workflows/ci.yaml`

**Verification:**
`bash -n test/e2e/core-services-proof.sh`
`test/e2e/core-services-proof.sh --dry-run`

### Task 4: Document local architecture and excluded capabilities

**Objective:** Explain why Ubiquity keeps per-component Helm charts and GitOps orchestration instead of adopting a monolithic chart.

**Files:**
- Create: `docs/architecture/core-services.md`
- Modify: `docs/architecture/overview.md`

**Verification:**
`go test ./pkg/cloud -run TestCoreServicesCapabilityContract -count=1`

### Task 5: Final verification and graph update

**Objective:** Prove all changes work and leave the repository clean.

**Commands:**
- `go test ./pkg/... ./cmd/... -count=1`
- `helm lint system/core-services`
- `helm template core-services system/core-services --namespace argocd >/tmp/core-services.yaml`
- `helm unittest system/core-services`
- `test/e2e/core-services-proof.sh --dry-run`
- active Helm lint/render loop
- `git diff --check`
- `graphify update .` if available
