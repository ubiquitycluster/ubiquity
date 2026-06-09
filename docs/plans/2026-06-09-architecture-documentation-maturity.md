# Architecture Documentation Maturity Implementation Plan

> **For Hermes:** Use test-driven-development skill to implement this plan task-by-task.

**Goal:** Close remaining architecture/documentation maturity gaps with docs-as-code tests for ADR coverage, networking cleanup, Helm chart reference generation, live-proof/non-certification language, and Kustomize/Helm boundaries.

**Architecture:** Add executable documentation tests under `pkg/cloud` because that package already owns repository-level documentation compliance tests. Implement the smallest documentation and script changes needed to satisfy each test, then verify with Go tests, markdown/script checks, Helm reference generation, and clean git status.

**Tech Stack:** Go tests, Markdown docs, Bash helper scripts, Helm metadata, repository search validation.

---

### Task 1: ADR audit contract

**Objective:** Prove the required ADR set exists and is linked from the ADR index.

**Files:**
- Modify: `pkg/cloud/architecture_docs_test.go`
- Modify: `docs/reference/architecture/decision-records/README.md`
- Create/modify ADR files only if a required decision is missing.

**Verification:**
`go test ./pkg/cloud -run TestArchitectureADRAuditCoversRequiredDecisions -count=1`

### Task 2: Networking architecture cleanup

**Objective:** Remove unresolved TODO language from `docs/architecture/networking.md` and replace it with concrete architecture guidance.

**Files:**
- Modify: `pkg/cloud/architecture_docs_test.go`
- Modify: `docs/architecture/networking.md`

**Verification:**
`go test ./pkg/cloud -run TestArchitectureNetworkingDocHasNoTODOs -count=1`

### Task 3: Helm chart reference generation

**Objective:** Add an auto-generated Helm chart reference and a reproducible generation script.

**Files:**
- Modify: `pkg/cloud/architecture_docs_test.go`
- Create: `scripts/generate-helm-chart-reference.sh`
- Create: `docs/reference/helm-charts.md`

**Verification:**
`go test ./pkg/cloud -run TestHelmChartReferenceIsGeneratedAndCurrent -count=1`
`bash -n scripts/generate-helm-chart-reference.sh`
`scripts/generate-helm-chart-reference.sh --check`

### Task 4: Cloud/NVIDIA live-proof language

**Objective:** Keep live-proof boundaries explicit and prevent NVIDIA approved/certified claims without attached approval evidence.

**Files:**
- Modify: `pkg/cloud/architecture_docs_test.go`
- Modify cloud/NVIDIA readiness docs as needed.

**Verification:**
`go test ./pkg/cloud -run TestCloudNVIDIAReadinessDocsUseEvidenceBoundaries -count=1`

### Task 5: Kustomize/Helm relationship docs

**Objective:** Explain why `platform/hpc-ubiq` uses Kustomize while platform/system components are generally Helm charts.

**Files:**
- Modify: `pkg/cloud/architecture_docs_test.go`
- Create: `docs/architecture/kustomize-helm.md`
- Modify: `docs/architecture/overview.md`

**Verification:**
`go test ./pkg/cloud -run TestKustomizeHelmRelationshipIsDocumented -count=1`

### Task 6: Final verification

**Objective:** Run full regression checks, update graphify if available, and leave the tree clean.

**Commands:**
- `go test ./pkg/... ./cmd/... -count=1`
- `bash -n scripts/generate-helm-chart-reference.sh`
- `scripts/generate-helm-chart-reference.sh --check`
- active Helm lint/template loop
- `git diff --check`
- `graphify update .` if available
