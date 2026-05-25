This goal transforms the ubiquitycluster/ubiquity repository from a collection of scripts, Makefiles, and mixed manifests into a battle-tested, CLI-driven HPC Cluster Lifecycle Platform. Single engineer, single CLI, production confidence.

Prerequisites: Go 1.22+, Docker, Helm, kubectl, k3d, pre-commit, shellcheck, yamllint, ansible-lint, trivy, task (go-task). Install anything missing.

Repo: /home/ubuntu/ubiquity

## What To Build

### 1. Unified CLI (the biggest item)

Replace the whole Makefile orchestration layer and both monolithic configure scripts (~113K lines of Python in 2 files) with a single Go CLI at cmd/ubiquity/.

- `ubiquity init` — bootstrap config, generate skeleton
- `ubiquity up` — detect platform (metal/cloud/sandbox) and deploy
- `ubiquity down` — tear down
- `ubiquity status` — cluster health summary (reads provisioning state)
- `ubiquity logs` — structured provisioning logs from Loki or local files
- `ubiquity test` — run the test suite

Structure: cmd/ubiquity/ for main.go and cobra commands, pkg/network/, pkg/provision/, pkg/config/, pkg/cloud/ for the domain logic. Use cobra + viper for CLI framework, charmbracelet/bubbletea for TUI components (status, progress). The CLI wraps terraform/ansible/helm as subprocesses internally.

Port at minimum:
- The network config logic (IPAM, DNS) from scripts/configure
- The provisioning pipeline (metal → bootstrap → external → wait → post-install) from the Makefile into the CLI state machine
- The sandbox bootstrap flow from scripts/configure-sandbox

### 2. Production CI/CD

Add .github/workflows/ci.yaml with jobs for:
- Lint: terraform fmt -check, ansible-lint, helm lint, shellcheck, yamllint, pre-commit run --all-files
- Validate: terraform validate, kubeconform on all K8s manifests, conftest (OPA policies)
- Test: molecule test on Ansible roles (target: metal/roles/*), helm unittest on all charts
- Scan: trivy fs --severity HIGH,CRITICAL . on changed paths
- Build: container image builds via ko, push to GHCR
- Deploy (optional for non-fork PRs): k3d cluster → ArgoCD sync → smoke test

Also add:
- pre-commit config (.pre-commit-config.yaml) with terraform_fmt, ansible-lint, helm-lint, yamllint, shellcheck, end-of-file-fixer, trailing-whitespace
- Renovate config tuned for Helm charts + Terraform + GitHub Actions

### 3. Clean up dead code and generated artifacts

- Add site/ to .gitignore and remove it from git tracking (it's generated mkdocs output — 99 HTML files, 96K lines)
- Remove the disabled/ directory entirely (abandoned apps: Tekton, Matrix, JupyterHub, cloudshell, system-upgrade)
- Remove commented-out Makefile targets (docker-build, docker-push, lines 43-50, 114-115)
- Fix stale license URLs — bootstrap/root/Chart.yaml line 8 references "ubiquity-open" repo
- Delete the `directory should say Dockerfiles/ not Dockerfiles/images-build` inconsistency
- Remove duplicate files (87 reported by pygount — consolidate into shared Kustomize bases and Helm subcharts)

### 4. Standardize on Helm + Kustomize for all K8s manifests

Every Kubernetes app should be a Helm chart with appropriate values.yaml and a Kustomize overlay for environment-specific patches. Convert any remaining raw YAML (e.g., coredns-autoscaler/coredns-autoscaler.yaml) to charts. Each chart must pass helm lint and have at least a basic helm unittest test.

### 5. Security hardening baseline

- Add a Kyverno or Gatekeeper deployment (in system/) with baseline policies:
  - Pod Security Standards: restricted profile
  - No privileged containers allowed
  - No hostNetwork for workloads (except system-critical)
  - Required resource limits on all pods
- Add trivy scanning to CI (block on CRITICAL)
- Add kube-bench as a CronJob in system/
- Add default-deny NetworkPolicy templates for all namespaces

### 6. Comprehensive testing

Create tests/ at each layer:
- pkg/* Go unit tests (>70% coverage on the CLI)
- molecule/ with molecule.yml + verify playbook for each Ansible role in metal/roles/
- terratest/ for each cloud provider Terraform module
- helm/tests/ with helm unittest test files for every chart
- integration/ with kuttl test suites for operator behavior

### 7. Provisioning state management

Add a JSON state file mechanism (into the CLI) that tracks each provisioning phase:
- Phase name, start time, end time, status (pending/running/success/failed), error message, log URL
- `ubiquity status` reads this and renders a progress table
- Failed phases can be retried individually: `ubiquity retry <phase>`
- State stored in .ubiquity/state.json

## Execution approach

Work in phases, making a git commit after each phase completes. Do Phase 1 first (it's the hardest and unlocks everything else). Use delegate_task for parallel work within a phase.

Phase 1: CLI scaffold + CI pipeline + dead code cleanup
Phase 2: Helm standardization + duplicate elimination
Phase 3: Security policies
Phase 4: Testing
Phase 5: State management + polish

## Success criteria (all must be true)

1. `ubiquity up --sandbox` deploys a working cluster (no more Makefile)
2. CI pipeline runs on every PR: lint → validate → scan → test, all green
3. No configure scripts remain (~113K lines ported to Go)
4. site/ directory no longer tracked in git
5. disabled/ directory removed
6. Every K8s app uses Helm chart + Kustomize overlay
7. Kyverno policies enforced in the sandbox deployment
8. Trivy scanning blocks CRITICAL vulnerabilities in CI
9. Go unit tests exist for all pkg/* packages
10. provisioning state management works: ubiquity status shows phase-level progress
