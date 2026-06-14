# Cloud Live Readiness Proof Implementation Plan

> **For Hermes:** Use test-driven-development skill to implement this plan task-by-task.

**Goal:** Convert Ubiquity cloud primitive readiness from render/readiness scaffolding into service-specific, fail-closed live proof contracts and a runnable evidence bundle.

**Architecture:** Keep render/apply as intent only. Add service-specific readiness contracts that join required CRDs, controller conditions, named smoke-test markers, restore-drill completion/readability markers, and tenant-cluster kubeconfig/API/node evidence. Collection remains Kubernetes-evidence based and fails closed when markers or controller status are absent.

**Tech Stack:** Go, Cobra CLI, Kubernetes `kubectl`, JSON readiness evidence, shell E2E scripts, Helm lint/template verification.

---

### Task 1: Commit the plan

**Objective:** Preserve implementation scope before changing code.

**Files:**
- Create: `docs/plans/2026-06-09-cloud-live-readiness-proof.md`

**Verify:**
- `git add docs/plans/2026-06-09-cloud-live-readiness-proof.md && git commit -m "docs: plan cloud live readiness proof"`

### Task 2: Add service-specific readiness contracts

**Objective:** Define explicit proof resources and smoke marker names for every managed cloud service.

**Files:**
- Modify: `pkg/cloud/service_readiness.go`
- Test: `pkg/cloud/service_readiness_test.go`

**Required services:** bucket, postgres, redis, kafka, registry/Harbor, mariadb, mongodb, nats, rabbitmq, clickhouse, opensearch, qdrant, openbao, http-cache, tcp-balancer.

**Verification:**
- `go test ./pkg/cloud -run TestManagedServiceReadinessContractsRequireSpecificSmokeMarkers -count=1`
- Commit: `feat: define cloud service readiness contracts`

### Task 3: Require restore-drill completion/readability proof

**Objective:** Ensure restore proof requires controller completion and readable restored data, not only Restore object presence.

**Files:**
- Modify: `pkg/cloud/readiness.go`
- Test: `pkg/cloud/readiness_test.go`

**Required evidence:**
- `restore-drill-controller-succeeded`
- `restore-drill-readable`
- `cloud-restore-drill-smoke-passed`

**Verification:**
- `go test ./pkg/cloud -run TestCloudReadinessRequiresRestoreCompletionReadableDataAndMarker -count=1`
- Commit: `fix: require completed cloud restore drill evidence`

### Task 4: Add tenant cluster readiness evidence

**Objective:** Evaluate tenant Cluster API conditions, kubeconfig secret presence, API reachability, and node readiness.

**Files:**
- Modify: `pkg/cloud/readiness.go`
- Test: `pkg/cloud/readiness_test.go`
- Modify: `cmd/ubiquity/cmd/cloud.go`
- Test: `cmd/ubiquity/cmd/cloud_test.go`

**Required markers/resources:**
- `tenant-cluster-kubeconfig-present`
- `tenant-cluster-api-reachable`
- `tenant-cluster-nodes-ready`
- `clusters.cluster.x-k8s.io` conditions with Ready true.

**Verification:**
- `go test ./pkg/cloud ./cmd/ubiquity/cmd -run 'TestCloudReadinessRequiresTenantClusterEvidence|TestCloudCollectReadinessIncludesTenantClusterMarkers' -count=1`
- Commit: `feat: require tenant cluster live readiness evidence`

### Task 5: Add cloud readiness proof bundle script

**Objective:** Provide a single CI/operator command that produces prerequisite, provenance, dry-run, collected evidence, readiness report, and restore-drill evidence outputs.

**Files:**
- Create: `test/e2e/cloud-readiness-proof-bundle.sh`
- Test: `pkg/cloud/cloud_readiness_bundle_test.go`

**Verification:**
- `go test ./pkg/cloud -run TestCloudReadinessProofBundleScriptCoversEvidenceOutputs -count=1`
- `bash -n test/e2e/cloud-readiness-proof-bundle.sh`
- `test/e2e/cloud-readiness-proof-bundle.sh --dry-run`
- Commit: `test: add cloud readiness proof bundle`

### Task 6: Integrate scripts/CI docs

**Objective:** Wire proof bundle and service markers into CI/GitOps docs and existing smoke scripts.

**Files:**
- Modify: `test/e2e/cloud-service-smoke-tests.sh`
- Modify: `.github/workflows/ci.yaml`
- Modify: `docs/runbooks/cloud-readiness-validation.md`
- Modify: `docs/runbooks/cloud-production-readiness-audit.md`
- Test: `pkg/cloud/chart_docs_test.go`

**Verification:**
- `go test ./pkg/cloud -run 'TestCloudReadinessDocs|TestCloudCIIncludesReadinessProofFlows' -count=1`
- Commit: `docs: wire cloud readiness proof into CI and runbooks`

### Task 7: Final verification

**Objective:** Prove all changes pass project checks and leave the repo clean.

**Commands:**
- `go test ./pkg/... ./cmd/... -count=1`
- active Helm lint/template loop with explicit namespace
- `bash -n test/e2e/*.sh`
- `test/e2e/cloud-readiness-proof-bundle.sh --dry-run`
- `test/e2e/cloud-service-smoke-tests.sh --dry-run`
- `git diff --check`
- `graphify update .` when available
- `git status --short --branch`
