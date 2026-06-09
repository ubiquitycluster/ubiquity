# Production CI Security Hardening Implementation Plan

> **For Hermes:** Use test-driven-development skill to implement this plan task-by-task.

**Goal:** Close remaining production CI/security hardening gaps with reviewer-visible CI jobs, scripts, policy fixtures, and docs-as-tests.

**Architecture:** Add explicit proof scripts for SBOM, image signing, runtime security, network behavior, Helm chart test/dependency freshness, and dependency reports. Wire those flows into GitHub Actions while preserving fail-closed behavior for live validation and dry-run behavior for regular CI.

**Tech Stack:** GitHub Actions, Go tests, Helm/helm-unittest, Kyverno CLI, Syft/CycloneDX, Cosign/Sigstore, Falco/Falcosidekick, Bash, Kubernetes NetworkPolicy probes.

---

### Task 1: Add the plan

**Objective:** Save the implementation plan and commit it.

**Files:**
- Create: `docs/plans/2026-06-09-production-ci-security-hardening.md`

**Verification:** `git log --oneline -1` shows the plan commit.

### Task 2: SBOM and image-signing CI contracts

**Objective:** Add failing tests that require SBOM generation with Syft/CycloneDX and image signing/verification with Cosign/Sigstore, then wire CI/scripts.

**Files:**
- Modify: `.github/workflows/ci.yaml`
- Create: `test/e2e/sbom-and-signing-proof.sh`
- Test: `pkg/cloud/security_ci_test.go`

**Acceptance:** Script supports `--dry-run`, uses `syft`, `cyclonedx`, `cosign sign`, `cosign verify`, and CI publishes artifact paths.

### Task 3: Kyverno policy-test coverage

**Objective:** Ensure Kyverno policy tests are executable via `kyverno test` and include deny/allow fixtures.

**Files:**
- Modify/Create: `system/kyverno-policies/kyverno-test/`
- Modify: `.github/workflows/ci.yaml`
- Test: `pkg/cloud/security_ci_test.go`

**Acceptance:** CI references `kyverno test system/kyverno-policies/kyverno-test` and fixtures contain allowed and denied resources.

### Task 4: Runtime security validation

**Objective:** Add Falco rule, alerting path, dashboard validation, and CI/script proof.

**Files:**
- Create: `system/falco-rules/Chart.yaml`
- Create: `system/falco-rules/values.yaml`
- Create: `system/falco-rules/templates/rules.yaml`
- Create: `system/falco-rules/tests/rules_test.yaml`
- Create: `test/e2e/runtime-security-validation.sh`
- Modify: `grafana/dashboards/cluster.json`
- Test: `pkg/cloud/security_ci_test.go`

**Acceptance:** Falco rule names alert on privileged shells, alert path references Falcosidekick/Alertmanager, dashboard includes Falco panels, script supports `--dry-run`.

### Task 5: Network policy behavioral proof

**Objective:** Upgrade network policy validation from render-only to deny/allow behavioral proof.

**Files:**
- Modify: `test/e2e/network-policy-behavior.sh`
- Test: `pkg/cloud/security_ci_test.go`

**Acceptance:** Script creates allow and deny pods/services, proves denied traffic fails and allowed traffic succeeds, and CI runs dry-run.

### Task 6: Helm unittest and dependency freshness

**Objective:** Require helm-unittest coverage across all active system/platform charts and chart dependency freshness checks.

**Files:**
- Create: `test/e2e/helm-hardening-checks.sh`
- Modify: `.github/workflows/ci.yaml`
- Test: `pkg/cloud/security_ci_test.go`

**Acceptance:** Script loops active charts, runs helm unittest where test suites exist, fails when active system/platform charts lack tests, runs `helm dependency list`, `helm dependency build`, and `helm lint/template`.

### Task 7: Scheduled dependency freshness report

**Objective:** Add scheduled workflow/report for go modules, Helm chart dependencies, GitHub Actions, and container images.

**Files:**
- Create: `.github/workflows/dependency-freshness.yml`
- Create: `test/e2e/dependency-freshness-report.sh`
- Test: `pkg/cloud/security_ci_test.go`

**Acceptance:** Workflow runs on schedule and manual dispatch, script emits a report artifact and supports `--dry-run`.

### Task 8: Full verification

**Objective:** Verify all changes and keep graph metadata current when possible.

**Commands:**
- `go test ./pkg/... ./cmd/... -count=1`
- Active chart `helm lint` / `helm template`
- `bash -n test/e2e/*.sh`
- Dry-run the new scripts
- `git diff --check`
- `graphify update .` if available

**Acceptance:** Working tree clean, all planned commits present, any live-path skips/fail-closed behavior documented.
