# Deep Assessment: Ubiquity Cluster

## Dimensions assessed

- Security
- Documentation & ADRs
- Architecture
- Ease of use
- Testing
- Functionality
- Comparison: khuedoan/homelab

---

## 1. Security

### What's good
- **Kyverno baseline policies**: restricted pod security, privileged container denial, label requirements. Chart exists at `system/kyverno-policies/`.
- **kube-bench CronJob**: CIS benchmark runs every 6h. Chart at `system/kube-bench/`.
- **Default-deny NetworkPolicies**: ingress/egress deny, allow-dns. Chart at `system/network-policies/`.
- **Trivy scanning**: in CI pipeline, blocks HIGH/CRITICAL.
- **Security phase**: wired as phase 3/6 in the provisioning pipeline.

### What's missing

| Gap | Severity | Details |
|-----|----------|---------|
| **No SBOM generation** | Medium | No Software Bill of Materials generated during CI. No `syft` or `cyclonedx` integration. |
| **No image signing** | High | Container images are built but never signed with cosign. No Sigstore integration. |
| **No admission-controller webhook testing** | High | Kyverno policies exist as YAML but are never actually tested against real K8s objects to verify they catch violations. |
| **No runtime security validation** | Medium | Falco is configured but there are no test rules, no alerting pipeline, no dashboard. |
| **No secrets scan in CI** | High | No `gitleaks` or `trufflehog` in pre-commit or CI. Secrets could be committed. |
| **No dependency vulnerability audit** | Medium | Go modules and npm packages are scanned by Renovate but there's no `govulncheck` or `npm audit` in CI. |
| **No IaC misconfiguration scanning beyond Trivy** | Low | Trivy covers some IaC, but no `checkov` or `tfsec` for deeper Terraform policy checks. |
| **Network policies defined but never verified** | Medium | `system/network-policies/` chart exists but there's no test that verifies policies actually block/allow expected traffic. |

### Fixes needed
1. Add `gitleaks` to `.pre-commit-config.yaml`
2. Add `govulncheck ./...` to CI test job
3. Add Kyverno policy tests — use `kyverno-json` or `kyverno test` to validate policies against sample resources
4. Add `cosign` to release workflow for image signing
5. Add `syft` to generate SBOM in CI (attach to GitHub release)

---

## 2. Documentation & ADRs

### What's good
- **MkDocs site**: configured at `mkdocs.yml`, generates a searchable web UI
- **Architecture overview**: `docs/architecture/overview.md` has mermaid diagrams
- **Deployment guides**: sandbox, on-prem, cloud
- **Runbooks**: AWX backup/restore, troubleshooting
- **CLI help**: all commands have `--help` with descriptions

### What's missing

| Gap | Severity | Details |
|-----|----------|---------|
| **Zero ADRs** | **Critical** | No Architecture Decision Records exist. homelab has 8+ ADRs covering base OS choice, tooling decisions, secret management, etc. Every major decision in ubiquity is undocumented. |
| **No CLI user guide** | High | README has a "Quick Start with the CLI" section but no detailed CLI reference. `ubiquity configure`, `ubiquity retry`, `ubiquity status --plain` are all undocumented. |
| **No architecture changelog** | Medium | `docs/changelog.md` exists but only tracks release versions. No record of architectural decisions and their dates. |
| **No design tenets documented** | Medium | The goal file has "Design Tenets" but they're not in the docs. A new contributor doesn't know the project's principles. |
| **No Helm chart reference** | Medium | 24+ Helm charts in `system/` but no documentation listing what they deploy, what values they accept, or how to customize them. |
| **No CLI command reference** | Medium | No man page or `docs/reference/cli.md`. Users must use `--help` every time. |
| **No FAQ** | Low | homelab has a FAQ. Ubiquity doesn't. Common questions (How do I add a node? How do I change the domain?) have no canonical answer. |
| **No architecture decision record template** | Low | homelab has a template in their ADRs doc. Ubiquity has nothing. |

### ADRs that should exist (minimum viable set)

| ADR | Topic | Rationale |
|-----|-------|-----------|
| ADR-001 | **Why Go CLI instead of Python** | Covers performance, single-binary distribution, type safety |
| ADR-002 | **Why k3s instead of full K8s** | Lightweight, embedded etcd, simpler operations |
| ADR-003 | **Why Terraform over Pulumi/CDKTF** | Mature multi-cloud ecosystem, community modules |
| ADR-004 | **Why Kyverno over OPA/Gatekeeper** | Kubernetes-native policy, mutation support |
| ADR-005 | **Why Helm chart per component** | Idempotent deployments, parameterization |
| ADR-006 | **Why Longhorn over Rook-Ceph** | Simpler operations, block storage focus |
| ADR-007 | **Why ArgoCD over Flux** | ApplicationSet multi-cluster support, mature SSO |
| ADR-008 | **Why Bubbletea over plain text** | Rich terminal UX, cross-platform TUI |

### Fixes needed
1. Create `docs/reference/architecture/decision-records/` directory with template
2. Write ADR-001 through ADR-008 (minimum)
3. Create `docs/reference/cli.md` with command reference
4. Add design tenets to `docs/architecture/overview.md`
5. Generate CLI reference automatically using `cobra` doc generation

---

## 3. Architecture

### What's good
- **Clean layer separation**: metal → bootstrap → system → platform → apps (plus cloud)
- **Provider interface**: `provision.Provider` allows mock testing and future backends
- **State machine**: JSON state file tracks phase lifecycle
- **Package structure**: `cmd/`, `pkg/`, `system/`, `platform/`, `apps/`, `metal/`, `cloud/`
- **CLI as entry point**: `ubiquity up` replaces `make` as the primary interface
- **Viper config**: supports `--env`, `UBQUITY_ENV`, `.ubiquity.yaml`

### What's missing

| Gap | Severity | Details |
|-----|----------|---------|
| **No app-template Helm chart** | **Critical** | Every app in `apps/` currently is its own complete chart. homelab uses a shared `app-template` that all apps depend on — 5 lines per app vs 50+. |
| **No central values repository** | High | Values are scattered across individual chart directories. No single source-of-truth for cluster-wide settings (domain, version overrides). |
| **No Kustomize-to-Helm bridge documented** | Medium | `platform/hpc-ubiq/` uses Kustomize while `system/` uses Helm. The relationship between these two deployment strategies is undocumented. |
| **No auto-generated Helm docs** | Medium | No `helm-docs` generation from values.yaml. Chart values are undocumented without reading raw YAML. |
| **No kubectl plugin** | Low | Could expose the CLI as a `kubectl-ubiquity` plugin for `kubectl ubiquity up`. |
| **No health check endpoint** | Low | CLI has no `ubiquity health` command to verify cluster health independently. |

### Fixes needed
1. Create a shared `charts/app-template/` chart that all apps depend on (takes templated name, image, ingress, resources)
2. Centralize global values into `values-global.yaml` at repo root
3. Add `helm-docs` to pre-commit or CI to auto-generate README.md for each chart
4. Add `ubiquity health` command that checks kubectl connectivity, ArgoCD status, storage health

---

## 4. Ease of Use

### What's good
- **Single binary**: `ubiquity` is a Go binary with no Python/Ruby/node dependencies
- **`make cli`**: builds with version ldflags
- **Shell completions**: bash, zsh, fish
- **`--help` on every command**: descriptive help text
- **Colored TUI**: `ubiquity status` has lipgloss colored output
- **`ubiquity up --sandbox`**: single command to get started
- **`ubiquity configure --domain`**: no more editing YAML by hand

### What's missing

| Gap | Severity | Details |
|-----|----------|---------|
| **No first-run wizard** | High | Running `ubiquity init` creates skeleton config but doesn't walk the user through anything. Should chain into `ubiquity configure -i`. |
| **No progress bar during up** | Medium | `ubiquity up` prints phase names but no spinner or real-time progress. Bubbletea is imported but used only for status, not for streaming progress during deployment. |
| **No `ubiquity info`** | Medium | No single command that shows: cluster version, K8s version, installed apps, resource usage. |
| **No dry-run mode** | Medium | `ubiquity up --dry-run` would show what would happen without doing it. |
| **No uninstall/cleanup** | Low | `ubiquity down` tears down but doesn't remove the CLI tool itself or clean up config. |
| **No `ubiquity docs`** | Low | Could open the docs site or generate a man page. |
| **No auto-completion install** | Low | Completions are generated but there's no `ubiquity completion install` that puts them in the right place. |

### Fixes needed
1. Chain `ubiquity init` into `ubiquity configure -i` when no `.env` exists
2. Add `ubiquity info` command that aggregates cluster state
3. Add `--dry-run` to `ubiquity up`
4. Add `ubiquity completion install --shell bash` that installs to the right directory
5. Add real-time streaming output for `ubiquity up` phases

---

## 5. Testing

### What's good
- **36 tests** across all Go packages
- **Provider interface** for mocking phase executors
- **Helm unittest**: 20 tests across 3 security charts
- **Terratest scaffold**: `terratest/` directory exists
- **Molecule scaffold**: `molecule/` directory exists
- **KUTTL scaffold**: `integration/` directory exists

### What's missing

| Gap | Severity | Details |
|-----|----------|---------|
| **No Kyverno policy tests** | High | Policies exist but have never been tested against actual K8s resources. Should use `kyverno test .` or `kyverno-json`. |
| **No E2E test** | High | No end-to-end test that provisions a sandbox cluster, deploys ArgoCD, and verifies apps reach Ready. |
| **No load testing** | Medium | HPC clusters need to handle workload. No `k6` or `vegeta` benchmarks. |
| **No upgrade test** | Medium | No test that verifies upgrading from version N to N+1 doesn't break the cluster. |
| **No Helm chart smoke tests** | Medium | Charts render but are never deployed to a real K8s (even k3d in CI) to verify they actually work. |
| **No conftest/OPA tests** | Medium | CI pipeline mentions conftest but there are no OPA policies written. |
| **No test for the down/logs/retry commands** | Low | These RunE functions are executed in tests but their return values aren't thoroughly asserted. |
| **No golden file tests for Helm** | Low | No snapshot/approval testing for rendered Helm output. A change that breaks rendering silently passes. |

### Fixes needed
1. Write Kyverno policy tests: `kyverno test system/kyverno-policies/`
2. Write one E2E test: provision k3d → wait for bootstrap → verify ArgoCD
3. Add `helm unittest` to CI for all charts that have tests
4. Add golden file tests for critical Helm charts using `helm template | diff`
5. Write conftest policies for Kubernetes manifest validation

---

## 6. Functionality

### What's good
- **6-phase provisioning pipeline**: metal → bootstrap → security → external → wait → post-install
- **Cloud provider support**: AWS, Azure, GCP, OpenStack, OVH
- **HPC workload managers**: Slurm, OpenPBS, OAR, HTCondor
- **Distributed storage**: Longhorn, Ceph, NFS, Lustre, BeeGFS, GPFS
- **SSO**: Keycloak, Dex, OpenLDAP
- **CI/CD**: Gitea, Argo Workflows, ArgoCD
- **Observability**: Prometheus, Grafana, Loki, Alertmanager

### What's missing vs homelab

| Feature | homelab | Ubiquity | Notes |
|---------|---------|----------|-------|
| **WireGuard VPN** | ✅ `apps/wireguard` | ❌ | Secure remote access to the cluster |
| **Tailscale VPN** | ✅ `apps/tailscale` | ❌ | Alternative to WireGuard, simpler setup |
| **Paperless** | ✅ `apps/paperless` | ❌ | Document management |
| **Jellyfin** | ✅ `apps/jellyfin` | ❌ | Media streaming |
| **Ollama** | ✅ `apps/ollama` | ❌ | Local LLM inference |
| **Woodpecker CI** | ✅ Self-hosted CI | ❌ | Uses GitHub Actions instead (cloud-dependent) |
| **Nix dev shell** | ✅ `flake.nix` | ❌ | Reproducible toolchain via Docker only |
| **app-template chart** | ✅ Shared chart | ❌ | 12 apps each have their own full Chart.yaml |
| **Self-hosted identity** | ✅ Authelia | ❌ | Uses Keycloak (heavier) |
| **OpenTofu** | ✅ `tofu_fmt` | ❌ | Uses Terraform (license concerns) |
| **Mixed-line-ending check** | ✅ Pre-commit hook | ❌ | No cross-platform line-ending enforcement |
| **CI helm-diff** | ✅ PR change detection | ❌ | No way to see what changes a PR makes to Helm values |

### Fixes needed
1. Create `charts/app-template/` — shared Helm chart all apps depend on (simplifies adding new apps to 5 lines)
2. Add Ollama for local AI inference (relevant for HPC environment)
3. Add Tailscale or WireGuard option for VPN access to the cluster
4. Consider OpenTofu migration or at minimum add a pre-commit hook for `terraform fmt`
5. Add `helm-diff` CI pipeline to show Helm value changes in PRs
6. Add `check-shebang-scripts-are-executable` and `mixed-line-ending` pre-commit hooks

---

## 7. Comparison: khuedoan/homelab — Deep Analysis

### What homelab does better

| Dimension | homelab | Ubiquity | Gap |
|-----------|---------|----------|-----|
| **ADRs** | 8 decision records | **0** | Critical — every architectural choice is unjustified |
| **App template chart** | Shared `app-template` | Each app has its own chart | Ubiquity: 50+ lines per app vs homelab: 5 lines |
| **Self-hosted CI** | Woodpecker (runs on cluster) | GitHub Actions (cloud) | Ubiquity can't run CI without GitHub |
| **Nix reproducibility** | `flake.nix` with pinned deps | Docker container | Ubiquity's tool versions aren't pinned |
| **Docs conciseness** | Clean, example-driven | Verbose, copy of README | Ubiquity docs need more examples, less prose |
| **Helm diff in CI** | `.woodpecker/helm-diff.yaml` | **None** | Can't audit chart changes in PRs |
| **Pre-commit hooks** | 9 hooks | 7 hooks | Missing `mixed-line-ending`, `check-shebang` |

### What ubiquity does better

| Dimension | ubiquity | homelab | Advantage |
|-----------|----------|---------|-----------|
| **CLI** | 10-command Go CLI | `make` only | Ubiquity: installable binary, no Python needed |
| **HPC support** | Slurm, OpenPBS, OAR, HTCondor | **None** | Ubiquity is purpose-built for HPC |
| **Cloud providers** | AWS, Azure, GCP, OpenStack, OVH | **None** | Ubiquity supports 5 clouds + bare metal |
| **Security hardening** | Kyverno, kube-bench, NSPolicies | **None** | Ubiquity: 3 security Helm charts + CI scanning |
| **State management** | JSON state, phase retry | **None** | Ubiquity: `ubiquity status`, `ubiquity retry <phase>` |
| **Test coverage** | 36 tests, Provider interface | Minimal tests | Ubiquity: Go unit tests, helm unittest |
| **goreleaser releases** | `.goreleaser.yaml`, 4 platforms | **None** | Ubiquity: downloadable binaries |
| **Shell completions** | bash, zsh, fish | **None** | Ubiquity: tab completion |
| **Bubbletea TUI** | Colored status table | Plain text | Ubiquity: richer terminal UX |
| **Multi-env** | sandbox, dev, prod | dev, prod | Ubiquity: sandbox mode for testing |
| **Viper config** | `.ubiquity.yaml`, env vars | `.envrc` only | Ubiquity: structured config |

### Shared weaknesses (both projects)

| Weakness | Details |
|----------|---------|
| **No secrets scanning in CI** | Neither project has gitleaks/trufflehog |
| **No SBOM generation** | Neither generates software bills of materials |
| **No image signing** | Neither signs container images with cosign |
| **No upgrade testing** | Neither tests version upgrades |
| **No load testing** | Neither benchmarks cluster performance |
| **No golden file Helm tests** | Neither snapshots Helm output for regression detection |
| **No Kyverno policy testing** | Neither validates policies against test resources |
| **No auto-generated CLI docs** | Neither generates man pages or CLI references from code |

---

## Summary: Priority Fix List

| Priority | Item | Category | Effort |
|----------|------|----------|--------|
| **P0** | Write ADR-001 through ADR-008 | Documentation | 2h |
| **P0** | Create `charts/app-template/` shared chart | Architecture | 1h |
| **P0** | Add gitleaks to pre-commit | Security | 10min |
| **P1** | Write Kyverno policy tests | Testing | 1h |
| **P1** | Add `helm-diff` CI pipeline | CI/CD | 1h |
| **P1** | Chain `ubiquity init` → `ubiquity configure -i` | Ease of use | 30min |
| **P1** | Add `ubiquity info` command | Functionality | 30min |
| **P2** | Add `mixed-line-ending`/`check-shebang` pre-commit hooks | Quality | 5min |
| **P2** | Create `docs/reference/cli.md` | Documentation | 1h |
| **P2** | Add `govulncheck` to CI | Security | 10min |
| **P3** | Add Ollama app for local LLM | Functionality | 30min |
| **P3** | Add Tailscale or WireGuard VPN | Functionality | 1h |
| **P3** | Add `ubiquity health` command | Ease of use | 30min |
