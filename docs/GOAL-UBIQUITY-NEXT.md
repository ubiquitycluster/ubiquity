# Goal: Project Ubiquity Next — The HPC Cluster Lifecycle Platform

> **One-line vision:** Transform Ubiquity from a collection of scripts, Makefiles, and manifests into a **battle-tested, CLI-driven HPC Cluster Lifecycle Platform** that a single engineer can deploy and operate with confidence.

## What This Goal Is

This is a single, unified transformation goal for the project. It addresses every critical issue identified — monolithic scripts, dead code, missing CI/CD, sparse testing, mixed tooling, security gaps, and quality of life — in one coherent push. The result is not just a "cleaner" repo; it is a **fundamentally different category of project** that can attract contributors, be verified by CI, and survive a production outage.

---

## The North Star

```
"One CLI command to deploy. One CI pipeline to validate. One framework to operate."
```

Today: `make metal bootstrap external wait post-install` with manual recovery.
Tomorrow: `ubiquity up --env prod` with integrated rollback, observability, and self-healing.

---

## Design Tenets

1. **Idempotency everywhere** — Running the same command twice produces the same result.
2. **Default secure** — Zero-trust networking, hardened nodes, admission policies, SBOM scanning baked in.
3. **Observable by default** — Every provisioning step emits structured logs and metrics.
4. **Tested in CI** — No change lands without validation at the unit, integration, and conformance level.
5. **Single toolchain** — One CLI, one testing framework, one way to express infrastructure.
6. **No generated artifacts in version control** — The repo holds source truth; everything else is built on demand.

---

## The Seven Pillars of Ubiquity Next

### Pillar 1: Unified CLI (`ubiquity`)

**Replace the 55K-line monolithic `scripts/configure` and the Makefile orchestration layer** with a single compiled CLI.

- `ubiquity init` — bootstrap configuration, generate skeleton
- `ubiquity up` — deploy full stack (detects platform: metal/cloud/sandbox)
- `ubiquity down` — tear down
- `ubiquity status` — cluster health summary
- `ubiquity logs` — structured provisioning logs
- `ubiquity test` — run the test suite

**Why this matters**: A new contributor today has to learn Makefile syntax, then Ansible, then Terraform, then Helm, then shell, then the 54K-line configure script. A unified CLI hides that complexity. The CLI internally composes the same tools, but the surface area is one binary with subcommands.

**Implementation approach**: Go CLI (single binary, cross-platform, good UX libraries like cobra + charmbracelet). It wraps terraform/ansible/helm as subprocesses and manages state.

### Pillar 2: Production CI/CD Pipeline

**Replace the `.github/` placeholder with a real CI pipeline that gates every PR.**

- **Lint**: `terraform fmt -check`, `ansible-lint`, `helm lint`, `shellcheck`, `yamllint`, `pre-commit`
- **Validate**: `terraform validate`, `kubeconform`, `conftest` (OPA policies on manifests)
- **Test**: `molecule test` for Ansible roles, `terratest` for Terraform modules, `helm unittest` for charts
- **Scan**: `trivy` for container images and IaC misconfigurations, `kube-bench` for CIS checks
- **Build**: Container image builds via ko or docker buildx, pushed to GHCR
- **Deploy**: Sandbox cluster (k3d) deployed in CI; ArgoCD syncs validating apps reach Ready state

**Why this matters**: Without CI, every change is a leap of faith. With CI, you catch regressions before they hit production and can confidently accept PRs from new contributors.

### Pillar 3: Modular Architecture (code quality)

**Break the monoliths and eliminate duplication.**

- Split `scripts/configure` and `scripts/configure-sandbox` (~113,000 combined lines) into a Go CLI with testable packages:
  - `pkg/network/` — IPAM, DNS, NAT configuration
  - `pkg/provision/` — PXE, Ironic, node lifecycle
  - `pkg/config/` — configuration loading and validation
  - `pkg/cloud/` — cloud provider abstractions
  - `pkg/storage/` — storage backend setup
- Eliminate 87 duplicate files by extracting shared Kustomize bases and Helm subcharts
- Move `disabled/` to an `archive` branch or remove entirely
- Remove `site/` (generated HTML) from version control

### Pillar 4: Unified Kubernetes Package Management

**Standardize on one approach for K8s application definitions.**

- Today: Mix of Helm charts (20+), Kustomize overlays (17+), raw YAML
- Tomorrow: All apps are **Helm charts** with **Kustomize overlays** for environment-specific patches
- Each chart has `Chart.yaml`, `values.yaml`, `templates/`, `tests/`
- All charts pass `helm lint` and `chart-testing` (ct) in CI

**Why this matters**: Helm is the de facto standard for Kubernetes. Kustomize is the overlay mechanism. Using both consistently means any Kubernetes engineer can contribute immediately.

### Pillar 5: Security Hardening Baseline

**Move from "configured but not enforced" to "enforced by default."**

- **Admission control**: Kyverno or Gatekeeper policies enforced for:
  - Pod Security Standards (restricted profile)
  - No privileged containers
  - No host network access for workloads
  - Required resource limits
- **Image scanning**: Trivy in CI on every container build; CVE gate blocks deployment
- **Network policies**: Default-deny policies for all namespaces; explicit allow rules
- **Secret rotation**: Vault integration with automatic renewal for TLS certs, database creds
- **Runtime security**: Falco rules enabled, alerting to Loki, dashboard in Grafana
- **CIS benchmarks**: kube-bench runs as a CronJob, results in Grafana

### Pillar 6: Testing Framework

**Move from "3 Go test files" to a comprehensive test pyramid.**

| Level | Tool | Covers |
|-------|------|--------|
| Unit (Terraform) | `terraform test` / terratest | Module logic, variable validation |
| Unit (Ansible) | `molecule` with verify playbooks | Role correctness on containerized targets |
| Unit (Helm) | `helm unittest` | Template rendering, value combinations |
| Integration | `kuttl` / `chainsaw` | Operator behavior, CRD reconciliation |
| Conformance | `sonobuoy` | K8s API compliance |
| E2E | `k6` + custom probes | Application health, HPC job submission |
| Security | `trivy`, `kube-bench`, `kubescape` | Vulnerability & misconfiguration detection |

**Why this matters**: Tests are not a luxury for a system that provisions bare metal and manages production HPC workloads. They are the safety net that makes refactoring possible.

### Pillar 7: Provisioning Observability & State Management

**Replace `sleep 60; ./scripts/wait-main-apps` with structured, observable provisioning.**

- Each phase writes state to a JSON file (or Kubernetes `ConfigMap`)
- State includes: phase name, start/end timestamps, logs URL, error details
- `ubiquity status` reads this state and displays a progress bar / table
- Failed phases can be retried individually (`ubiquity retry bootstrap`)
- Structured logs (JSON) shipped to Loki
- Grafana dashboard showing: deployment progress, error rates, duration trends

**Why this matters**: Today if provisioning fails at step 4 of 6, the user sees a cryptic error and has to re-read the playbook output to understand where. With state management, they see "FAILED at bootstrap/external-secrets — timeout waiting for CRD" and can retry just that step.

---

## Implementation Phases

### Phase 1: Foundation (Weeks 1-2)
- [ ] Scaffold Go CLI (`cmd/ubiquity/`, `pkg/`)
- [ ] Port `scripts/configure` network + DNS logic to `pkg/network/`
- [ ] Add `site/` to `.gitignore`; purge from git history with `git filter-repo`
- [ ] Add `.github/workflows/ci.yaml` with lint + validate jobs
- [ ] Configure pre-commit with all hooks
- [ ] Remove dead code: `disabled/`, commented Makefile targets, stale license URLs

### Phase 2: Standardization (Weeks 3-4)
- [ ] Consolidate Kustomize-only apps to Helm charts
- [ ] Extract 87 duplicate files into shared bases/subcharts
- [ ] Add `helm unittest` tests to every chart
- [ ] Add Trivy scanning to CI
- [ ] Implement `ubiquity up` sandbox flow

### Phase 3: Security (Weeks 5-6)
- [ ] Deploy Kyverno with baseline policies
- [ ] Add network policies for all namespaces
- [ ] Add kube-bench CronJob
- [ ] Enable Falco with Loki output
- [ ] Add conftest/OPA policy checks to CI

### Phase 4: Testing & Reliability (Weeks 7-8)
- [ ] Molecule tests for all Ansible roles
- [ ] Terratest for Terraform modules
- [ ] KUTTL integration tests for operators
- [ ] State management for provisioning phases
- [ ] `ubiquity status` and `ubiquity retry` commands

### Phase 5: Polish (Weeks 9-10)
- [ ] `ubiquity test` — run full test suite
- [ ] Grafana dashboard for provisioning observability
- [ ] `ubiquity logs` — structured log tailing
- [ ] Documentation refresh
- [ ] Onboarding guide: "Deploy a cluster in 10 minutes with `ubiquity up`"

---

## Success Criteria

The goal is achieved when all of the following are true:

| # | Criterion | Current | Target |
|---|-----------|---------|--------|
| 1 | Lines of Python in configure scripts | ~113,000 in 2 files | 0 (ported to Go CLI) |
| 2 | CI pipeline gating PRs | None | Lint + validate + scan + test on every PR |
| 3 | Test coverage for infrastructure code | ~0% | >70% for Go CLI, molecule for all Ansible, terratest for all Terraform |
| 4 | Duplicate files | 87 | 0 |
| 5 | Generated HTML in version control | 99 files / 96K lines | 0 (`site/` .gitignored, built in CI) |
| 6 | Dead/disabled code | `disabled/`, commented targets, stale refs | 0 |
| 7 | Package manager consistency | Helm + Kustomize + raw YAML | Helm + Kustomize only |
| 8 | Security policy enforcement | None | Kyverno + Trivy + kube-bench + Falco active |
| 9 | Deploy command | `make metal bootstrap external wait` | `ubiquity up --env prod` |
| 10 | Provisioning observability | `sleep 60; ./wait-main-apps` | Structured logs + state tracking + retry |

---

## What Success Looks Like

A new contributor clones the repo, runs:

```bash
ubiquity init
ubiquity up --sandbox
```

And 15 minutes later has a running HPC cluster in Docker (k3d). They can submit a Slurm job, browse the dashboard, and see Grafana showing cluster health. When they open a PR, CI runs in 5 minutes and tells them exactly what's wrong — no manual debugging required.

The project goes from "a working system that's hard to contribute to" to **"a system that's both powerful and approachable"** — the hallmark of infrastructure that lasts.
