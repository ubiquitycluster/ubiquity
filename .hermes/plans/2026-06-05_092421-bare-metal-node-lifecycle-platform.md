# NVIDIA Infra Controller Bare Metal Node Lifecycle Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task. Use test-driven-development for every code path. Do not claim NVIDIA approval/certification without explicit approval evidence. NVIDIA Infra Controller is currently described upstream as experimental/preview; Ubiquity must surface that status and require explicit validation before production claims.

**Goal:** Replace Ubiquity's long-term bare-metal node lifecycle design with NVIDIA Infra Controller (NICo) as the primary in-cluster infrastructure control plane, while retaining only the minimum external bootstrap needed to create the first Kubernetes management cluster.

**Architecture:** Bootstrap starts with Ubiquity's existing Ansible/PXE path only until a management Kubernetes cluster exists. After bootstrap, Ubiquity installs NICo Core + NICo REST + site-agent into the cluster and uses NICo APIs for site-local bare-metal inventory, Redfish/BMC power control, DHCP/PXE/DNS/NTP, OS image provisioning, machine/instance lifecycle, status history, hardware health, and DPU-enforced isolation where supported. Ubiquity owns the opinionated wrapper: declarative inventory/profile translation, GitOps packaging, CLI UX, safety gates, Kubernetes/NVIDIA AI readiness correlation, live node view, docs, and fail-closed validation.

**Tech Stack:** Go CLI, Kubernetes, Helm/Helmfile, ArgoCD GitOps, NVIDIA/infra-controller NICo Core, NICo REST API, NICo site-agent, Temporal, Keycloak/OIDC, Vault, External Secrets, PostgreSQL, MetalLB, cert-manager, Redfish/BMC, DHCP/PXE/iPXE, NVIDIA GPU Operator, NVIDIA k8s-launch-kit, DCGM, NIM Operator, AIStore.

---

## Correction to prior plan

The prior plan incorrectly centered Metal3/Bare Metal Operator/Ironic as the day-2 lifecycle manager. Update the target architecture:

- **Primary day-2 manager:** `NVIDIA/infra-controller` (NVIDIA Infra Controller / NICo).
- **Metal3/BMO status:** fallback/migration-only path, not the target default, unless NICo is unavailable or explicitly disabled.
- **Ubiquity role:** install/configure NICo, translate Ubiquity intent into NICo resources/API calls, aggregate live status, enforce safety gates, and correlate NICo machine/instance status with Kubernetes node and NVIDIA AI workload readiness.

NICo upstream capabilities discovered from `https://github.com/NVIDIA/infra-controller`:

- Site-local zero-trust bare-metal lifecycle management.
- API-based microservice architecture.
- DPU-enforced isolation where supported.
- Hardware inventory management and orchestration.
- Redfish-based BMC/hardware management.
- Hardware testing and firmware updates.
- IPAM and DNS services.
- Power control: on/off/reset.
- DHCP/PXE-based OS provisioning.
- Machine wipe/release orchestration.
- Machine trust enforcement during tenant switching.
- Helm deployment for NICo Core.
- NICo REST API, Temporal, Keycloak, site-agent.
- `nicocli` CLI generated from the OpenAPI spec.
- OpenAPI resources for Site, Machine, Machine status history, Operating System, Instance, Instance status history, Allocation, SKU, InfiniBand Partition, NVLink Logical Partition, Task, and Machine GPU stats.

---

## Current-state assessment

### What exists in Ubiquity today

- `metal/Makefile` runs `boot` then `cluster`.
- `metal/boot.yml` starts an ephemeral PXE service and wakes/powers nodes with WoL/IPMI.
- `metal/roles/pxe_server/tasks/main.yml` downloads/extracts a globally selected OS ISO and generates MAC-addressed Kickstart configs.
- `metal/cluster.yml` installs and joins k3s nodes, generates kubeconfig, configures MetalLB, and labels nodes.
- `cmd/ubiquity/cmd/up.go` describes a production lifecycle, but `provisionMetal()` is still not wired to a full production lifecycle manager.
- `system/baremetal-operator-system/kustomization.yaml` references old Metal3/BMO/Ironic assets. This should no longer be the target default.
- `docs/admin-guide/administration/tutorials/add-or-remove-nodes.md` documents manual inventory-driven add/remove/reinstall.
- `tools/disk-image/scripts/reinstall_host.sh` and `delete-host.sh` are unsafe BMO-era shell snippets and should be deprecated.
- AI readiness work already validates NVIDIA GPU/MIG/RDMA/DCGM/NIM/KAI/AIStore fail-closed state.

### Gaps to plug

1. Ubiquity lacks an NVIDIA-native long-term bare-metal lifecycle manager.
2. Initial bootstrap and day-2 management are not cleanly separated.
3. `ubiquity up --env prod` does not install/configure NICo.
4. Node add/remove/reinstall is manual and shell-script-driven.
5. There is no live view that joins physical machine, OS image, NICo instance, Kubernetes node, GPU/RDMA/MIG, and AI-readiness state.
6. Multi-OS boot image management is not declarative.
7. OS images are not represented through NICo's Operating System API.
8. The current PXE path is mostly Kickstart/RHEL-family oriented and not per-node OS-family driven.
9. There is no Ubiquity wrapper around NICo auth, tenant/site selection, or OpenAPI/nicocli usage.
10. There are no fail-closed validations that NICo Core, NICo REST, site-agent, DHCP, DNS, NTP, PXE, BMC proxy, and hardware-health services are actually ready.
11. Documentation does not explain the full bootstrap -> NICo -> Kubernetes node -> NVIDIA AI workload lifecycle.

---

## Replacement map

| Ubiquity current/bespoke function | Replace with NVIDIA Infra Controller | Ubiquity still owns |
| --- | --- | --- |
| Day-2 BMC/power control through scripts/IPMI snippets | NICo Redfish/BMC management and machine APIs | CLI safety policy, wrapper UX, credential redaction |
| Ephemeral day-2 PXE/DHCP/TFTP/HTTP | NICo `nico-dhcp`, `nico-pxe`, `nico-dns`, `nico-ntp` | Bootstrap-only PXE fallback, site values generation |
| Manual OS reinstall script | NICo Operating System + Instance lifecycle | OS catalog validation, reinstall command, post-provision checks |
| BMO/Metal3 target design | NICo Core + NICo REST + site-agent | Migration/fallback docs only |
| Inventory-only add/remove nodes | NICo Site/Machine/Instance lifecycle | Declarative inventory translation and safety gates |
| No live physical status view | NICo Machine/Instance/status-history/Task/GPU stats APIs | Unified `ubiquity nodes` view with Kubernetes and AI readiness |
| Manual power/release operations | NICo tasks and lifecycle operations | Quorum/drain/storage/GPU readiness gates |
| Local hardware-health assumptions | NICo hardware-health service | Policy interpretation and docs |
| Local network boot assumptions | NICo IPAM/DNS/DHCP/NTP/PXE | Site network profile generation and validation |
| Optional DPU handling | NICo DPU-enforced isolation + `NVIDIA/doca-platform` when needed | Hardware gating and no unsupported claims |

Other NVIDIA repositories remain part of the AI platform layer, not the physical lifecycle controller:

- `NVIDIA/gpu-operator` for GPU driver/runtime/device-plugin/DCGM/MIG validation.
- `NVIDIA/k8s-device-plugin` through GPU Operator or explicit fallback only.
- `NVIDIA/k8s-driver-manager` for driver upgrade/drain behavior.
- `NVIDIA/dcgm-exporter` for GPU telemetry provenance.
- `NVIDIA/k8s-launch-kit` for NIC/RDMA/SR-IOV discovery/profile generation patterns.
- `NVIDIA/k8s-nim-operator` for NIM inference serving validation.
- `NVIDIA/ais-k8s` for AI data-plane storage.
- `NVIDIA/k8s-dra-driver-gpu` for optional future DRA GPU profiles.
- `NVIDIA/cloud-native-stack` as source-backed component matrix/reference.

---

## Target operating model

### Bootstrap boundary

NICo runs in Kubernetes, so the very first cluster still needs an external bootstrap step.

```text
operator/provisioner host
  -> Ubiquity bootstrap inventory
  -> minimal PXE/Ansible path to install first management cluster
  -> ArgoCD/root app install
  -> NICo prerequisites: MetalLB, cert-manager, Vault, External Secrets, PostgreSQL
  -> NICo Core: nico-api, nico-bmc-proxy, nico-dhcp, nico-dns, nico-ntp, nico-pxe, nico-hardware-health, nico-ssh-console-rs
  -> NICo REST: REST API, Temporal, Keycloak/OIDC, site-manager, site-agent
  -> day-2 lifecycle moves to NICo APIs
```

### Day-2 lifecycle through NICo

```text
Ubiquity NodeInventory + OSImage catalog
  -> render NICo site/network/OS/machine intent
  -> create/update NICo Operating System objects
  -> discover/register machines through NICo site-agent/BMC/Redfish
  -> create NICo Instances for nodes that should join Kubernetes
  -> NICo handles DHCP/PXE/DNS/NTP/provisioning/power/wipe/release
  -> node boots and joins Kubernetes
  -> Ubiquity applies labels/taints/roles
  -> NVIDIA operators reconcile GPU/network/runtime
  -> Ubiquity marks lifecycle complete only after NICo + Kubernetes + NVIDIA readiness passes
```

### Live node view

`ubiquity nodes list` must aggregate from NICo and Kubernetes:

```text
NAME   SITE  MACHINE      INSTANCE      POWER  NICO_STATE  OS_IMAGE        K8S     ROLE     GPU      RDMA  TASK       AGE
cn01   sf01  mach-123     inst-456      on     provisioned rocky-9.4-gpu   Ready   worker   8xH100   yes   none       14d
cn02   sf01  mach-124     inst-789      on     provisioning ubuntu-24.04   None    worker   unknown  n/a   reinstall  6m
cn03   sf01  mach-125     none          off    available   none            None    spare    unknown  n/a   none       42d
```

Sources:

- NICo Site API.
- NICo Machine API.
- NICo Machine status history.
- NICo Machine GPU stats API.
- NICo Operating System API.
- NICo Instance API.
- NICo Instance status history.
- NICo Task API.
- Kubernetes Node API.
- Existing Ubiquity GPU/MIG/RDMA/DCGM/NIM/KAI readiness parsers.

---

## Implementation plan

### Phase 0: Provenance and architectural reset

#### Task 0.1: Add NICo architecture ADR

**Objective:** Replace the BMO-centered day-2 design with NICo as the target lifecycle manager.

**Files:**
- Create: `docs/architecture/on-prem/nvidia-infra-controller-node-lifecycle.md`
- Modify: `docs/architecture/on-prem/openstack-bmo-node-discovery.md` to mark BMO as fallback/legacy.
- Modify: `mkdocs.yml` if docs nav exists.
- Create/modify docs-as-code test: `pkg/aiplatform/production_validation_test.go` or `pkg/docs/nico_docs_test.go`.

**Test first:** Require the docs to contain:

```text
NVIDIA Infra Controller
NVIDIA/infra-controller
NICo Core
NICo REST
site-agent
Operating System
Machine
Instance
Task
Machine GPU stats
bootstrap boundary
experimental/preview
```

**Acceptance:** Docs clearly say NICo is primary day-2 lifecycle manager and Ubiquity uses external bootstrap only to create the first cluster.

#### Task 0.2: Deprecate BMO target docs/scripts

**Objective:** Prevent future work from continuing down the wrong default path.

**Files:**
- Modify: `docs/reference/metal3/baremetalhost-states.md`
- Modify: `docs/reference/metal3/remove-host.md`
- Modify or deprecate: `tools/disk-image/scripts/reinstall_host.sh`
- Modify or deprecate: `tools/disk-image/scripts/delete-host.sh`

**Acceptance:** Each file states BMO/Metal3 is fallback/migration-only and target default is NICo.

---

### Phase 1: Add NICo integration package

#### Task 1.1: Add NICo typed config

**Objective:** Model NICo connection and site configuration without leaking credentials.

**Files:**
- Create: `pkg/nico/config.go`
- Create: `pkg/nico/config_test.go`

**Types:**

```go
type Config struct {
    BaseURL      string
    Org          string
    APIName      string
    SiteName     string
    TokenCommand string
    TokenEnv     string
    ConfigPath   string
}
```

**Tests:**
- Defaults `APIName` to `nico`.
- Redacts tokens and token commands from string/debug output.
- Rejects empty BaseURL for live mode.
- Allows mock/offline mode for tests.

#### Task 1.2: Add NICo REST client interface

**Objective:** Wrap NICo OpenAPI endpoints behind a testable interface.

**Files:**
- Create: `pkg/nico/client.go`
- Create: `pkg/nico/client_test.go`

**Interface:**

```go
type Client interface {
    ListSites(ctx context.Context) ([]Site, error)
    ListMachines(ctx context.Context, siteID string) ([]Machine, error)
    GetMachine(ctx context.Context, machineID string) (Machine, error)
    ListOperatingSystems(ctx context.Context) ([]OperatingSystem, error)
    CreateOperatingSystem(ctx context.Context, os OperatingSystemSpec) (OperatingSystem, error)
    ListInstances(ctx context.Context, siteID string) ([]Instance, error)
    CreateInstance(ctx context.Context, req CreateInstanceRequest) (Instance, error)
    DeleteInstance(ctx context.Context, instanceID string) (Task, error)
    GetTask(ctx context.Context, taskID string) (Task, error)
    GetMachineGPUStats(ctx context.Context) (MachineGPUStats, error)
}
```

**Implementation choice:** Start with a minimal HTTP client using the documented OpenAPI paths. Later replace with generated SDK if the upstream `rest-api/sdk` stabilizes enough.

#### Task 1.3: Add `nicocli` adapter

**Objective:** Allow Ubiquity to use upstream `nicocli` when installed, while retaining the HTTP client for CI.

**Files:**
- Create: `pkg/nico/nicocli.go`
- Create: `pkg/nico/nicocli_test.go`

**Acceptance:**
- Detects `nicocli` on PATH.
- Supports `NICO_CONFIG`, `NICO_BASE_URL`, `NICO_ORG`, `NICO_TOKEN_COMMAND`.
- Parses JSON output.
- Redacts secrets.
- Falls back to HTTP client if disabled.

---

### Phase 2: Package NICo through Ubiquity GitOps

#### Task 2.1: Add NICo Core Helm wrapper

**Objective:** Make NICo Core installable via Ubiquity profiles.

**Files:**
- Create: `system/nvidia-infra-controller-core/Chart.yaml`
- Create: `system/nvidia-infra-controller-core/values.yaml`
- Create: `system/nvidia-infra-controller-core/templates/application.yaml` or use existing app-template pattern.
- Create: `system/nvidia-infra-controller-core/README.md`

**Source:** `NVIDIA/infra-controller/helm`.

**Must model:**
- `nico-api`
- `nico-bmc-proxy`
- `nico-dhcp`
- `nico-dns`
- `nico-hardware-health`
- `nico-ntp`
- `nico-pxe`
- `nico-ssh-console-rs`
- optional `unbound`

**Acceptance:** Helm render includes provenance annotations:

```yaml
ubiquity.dev/source-repo: NVIDIA/infra-controller
ubiquity.dev/source-component: nico-core
ubiquity.dev/source-status: experimental-preview
```

#### Task 2.2: Add NICo REST Helm wrapper

**Objective:** Install NICo REST API, site-agent, Temporal, Keycloak/OIDC integration.

**Files:**
- Create: `platform/nvidia-infra-controller-rest/Chart.yaml`
- Create: `platform/nvidia-infra-controller-rest/values.yaml`
- Create: `platform/nvidia-infra-controller-rest/README.md`

**Source:** `NVIDIA/infra-controller/rest-api/helm`.

**Acceptance:** Wrapper supports:
- external or bundled IdP mode
- site-agent deployment
- Temporal namespace/site UUID configuration
- image registry/tag values
- no committed secrets

#### Task 2.3: Add NICo prerequisites profile

**Objective:** Manage NICo prerequisite services through Ubiquity, not ad hoc scripts.

**Files:**
- Create: `system/nvidia-infra-controller-prereqs/Chart.yaml`
- Create: `system/nvidia-infra-controller-prereqs/values.yaml`
- Create: docs: `docs/reference/nvidia-infra-controller/prerequisites.md`

**Prereqs:**
- MetalLB
- cert-manager
- Vault
- External Secrets
- PostgreSQL

**Acceptance:** Docs explain when to reuse existing Ubiquity services versus deploy NICo-specific ones. No duplicate default StorageClass takeover without explicit profile setting.

---

### Phase 3: Declarative inventory and multi-OS image catalog

#### Task 3.1: Add Ubiquity NodeInventory and OSImage model

**Objective:** Keep one Ubiquity source of truth that can translate into NICo Operating Systems and Instance requests.

**Files:**
- Create: `pkg/nodeinventory/types.go`
- Create: `pkg/nodeinventory/types_test.go`
- Create: `examples/node-inventory/nico-prod.yaml`

**Types:**

```go
type OSImage struct {
    Name         string
    Family       string // rocky, rhel, ubuntu, custom
    Version      string
    Architecture string
    IPXEScript   string
    UserData     string
    ImageURL     string
    Checksum     string
    Labels       map[string]string
}

type BareMetalNode struct {
    Name            string
    Site            string
    MachineSelector map[string]string
    Role            string
    OSImageRef      string
    InstanceTypeRef string
    GPUProfile      string
    JoinProfile     string
    Labels          map[string]string
}
```

**Acceptance:** Supported image families include Rocky/RHEL, Ubuntu, and custom. Every non-dev image requires checksum/provenance.

#### Task 3.2: Render NICo Operating System objects

**Objective:** Support multiple bootable OS image types through NICo's Operating System API.

**Files:**
- Create: `pkg/nodeinventory/render_nico_os.go`
- Create: `pkg/nodeinventory/render_nico_os_test.go`

**Acceptance:**
- Rocky/RHEL renders iPXE + Kickstart user data.
- Ubuntu renders iPXE + autoinstall/cloud-init user data.
- Custom renders user-supplied iPXE/user-data.
- All rendered specs can be posted to NICo Operating System create/update API.

#### Task 3.3: Generate bootstrap fallback inventory from same source

**Objective:** Preserve initial bootstrap using the same inventory.

**Files:**
- Create: `pkg/nodeinventory/render_ansible.go`
- Create: `pkg/nodeinventory/render_ansible_test.go`

**Acceptance:** The initial cluster can still be bootstrapped externally, but day-2 nodes are not managed through Ansible unless explicitly using fallback mode.

---

### Phase 4: Wire `ubiquity up` to NICo

#### Task 4.1: Add deployment backend flags

**Objective:** Make backend selection explicit.

**Files:**
- Modify: `cmd/ubiquity/cmd/up.go`
- Modify/create: `cmd/ubiquity/cmd/up_test.go`

**Flags:**

```text
--metal-bootstrap-backend ansible|none
--node-lifecycle-backend nico|bmo|none
--nico-values <path>
--nico-rest-values <path>
--nico-site <name>
```

**Defaults:**
- sandbox: `none` / `none`
- production: `ansible` / `nico`

**Acceptance:** `ubiquity up --env prod` can bootstrap the first cluster and then install NICo through GitOps/rendered manifests.

#### Task 4.2: Add NICo readiness checks to `ubiquity health`

**Objective:** Fail closed unless NICo services are ready.

**Files:**
- Modify: `cmd/ubiquity/cmd/health.go`
- Create: `pkg/nico/readiness.go`
- Create: `pkg/nico/readiness_test.go`

**Checks:**
- `nico-core` workloads ready.
- `nico-rest` API health endpoint ready.
- `site-agent` connected/healthy.
- `nico-dhcp`, `nico-dns`, `nico-ntp`, `nico-pxe` workloads ready.
- `nico-bmc-proxy` ready.
- `nico-hardware-health` ready.
- At least one Site object visible.
- At least one Machine object visible in real hardware mode.

**Acceptance:** No `ubiquity health` output may claim bare-metal lifecycle readiness unless NICo readiness passes.

---

### Phase 5: Implement `ubiquity nodes` using NICo APIs

#### Task 5.1: Add node status aggregation model

**Objective:** Build the live view from NICo + Kubernetes + NVIDIA AI readiness.

**Files:**
- Create: `pkg/nodestatus/status.go`
- Create: `pkg/nodestatus/collect_nico.go`
- Create: `pkg/nodestatus/collect_nico_test.go`

**Model:**

```go
type NodeStatus struct {
    Name string
    Site string
    MachineID string
    InstanceID string
    PowerState string
    MachineStatus string
    InstanceStatus string
    OSImage string
    KubernetesNodeName string
    KubernetesReady bool
    Cordoned bool
    Roles []string
    GPUs int
    MIGProfiles map[string]int
    RDMAResources int
    NVIDIAReady bool
    ActiveTaskID string
    LastAction string
    Reason string
}
```

#### Task 5.2: Add CLI commands

**Objective:** Give operators a first-class live node view and lifecycle interface.

**Files:**
- Create: `cmd/ubiquity/cmd/nodes.go`
- Create: `cmd/ubiquity/cmd/nodes_test.go`

**Commands:**

```sh
ubiquity nodes list [-o table|json]
ubiquity nodes status <name> [-o json]
ubiquity nodes os list
ubiquity nodes os apply --inventory examples/node-inventory/nico-prod.yaml
ubiquity nodes add <name> --inventory examples/node-inventory/nico-prod.yaml
ubiquity nodes drain <name>
ubiquity nodes remove <name> --confirm <name>
ubiquity nodes reinstall <name> --os-image ubuntu-24.04-gpu --confirm <name>
ubiquity nodes power <name> on|off|reset
ubiquity nodes task <task-id>
```

**Acceptance:**
- `list` joins NICo Machine/Instance with Kubernetes node and GPU/MIG/RDMA evidence.
- `-o json` redacts credentials.
- Commands use NICo by default, not BMO.

#### Task 5.3: Implement task polling

**Objective:** Long operations must show NICo Task progress.

**Files:**
- Create: `pkg/nico/tasks.go`
- Create: `pkg/nico/tasks_test.go`

**Acceptance:**
- Polls NICo Task API with timeout.
- Shows current state/reason.
- Exits nonzero on failed/cancelled task.

---

### Phase 6: Safe lifecycle operations

#### Task 6.1: Implement add/provision

**Objective:** Create NICo Instance requests from Ubiquity inventory.

**Flow:**
1. Validate OSImage exists.
2. Ensure NICo Operating System exists or create it.
3. Select Machine by selector or explicit machine ID.
4. Create Instance with selected OS/network/profile.
5. Poll NICo Task/Instance status.
6. Wait for Kubernetes Node Ready if join profile expects cluster join.
7. Run NVIDIA post-provision checks for GPU profiles.

#### Task 6.2: Implement remove/release

**Objective:** Safely release a machine from Kubernetes and NICo.

**Safety gates:**
- Cordon first.
- Drain with timeout and blocker report.
- Prevent unsafe control-plane quorum loss.
- Detect local PV/AIStore target data and require explicit acknowledgement.
- Delete or release NICo Instance.
- Poll NICo Task.
- Verify Machine returns to expected available/released state.

#### Task 6.3: Implement reinstall/reimage

**Objective:** Reinstall using a same or different OS image through NICo Operating System + Instance lifecycle.

**Flow:**
1. Validate target OS image.
2. Drain and delete/release old instance safely.
3. Create new NICo Instance on same Machine using target Operating System.
4. Poll task/instance.
5. Wait for Kubernetes join.
6. Reapply labels/taints.
7. Run NVIDIA GPU/RDMA/MIG/DCGM checks if selected profile requires them.

#### Task 6.4: Implement power operations

**Objective:** Wrap NICo BMC/Redfish power control safely.

**Acceptance:**
- `reset` requires `--confirm <name>`.
- Powering off a Kubernetes-ready node requires drain or `--force --reason`.
- Control-plane nodes require quorum check.

---

### Phase 7: NVIDIA AI integration after NICo provisioning

#### Task 7.1: Correlate NICo GPU stats with Kubernetes resources

**Objective:** Detect mismatches between physical GPU inventory and Kubernetes allocatable GPU/MIG resources.

**Files:**
- Modify: `pkg/aiplatform/readiness.go`
- Create: `pkg/nodestatus/gpu_correlation.go`
- Create: `pkg/nodestatus/gpu_correlation_test.go`

**Acceptance:**
- NICo says GPU machine exists but Kubernetes exposes no `nvidia.com/gpu`/`nvidia.com/mig-*` => node lifecycle incomplete.
- Kubernetes exposes GPUs but NICo machine status unhealthy => node lifecycle degraded.

#### Task 7.2: Use `NVIDIA/k8s-launch-kit` for network/RDMA discovery where available

**Objective:** Tie NICo physical fabric/site info to Kubernetes RDMA readiness.

**Files:**
- Create: `pkg/nvidia/launchkit.go`
- Create: `pkg/nvidia/launchkit_test.go`

**Acceptance:**
- `ubiquity nodes status` reports network/RDMA provenance: `nico`, `k8s-launch-kit`, or `local-kubectl`.
- `ai-production` fails closed if RDMA is expected but missing.

#### Task 7.3: Add NIM/AI post-provision smoke hooks

**Objective:** Ensure the node can actually serve AI workloads after provisioning.

**Acceptance:**
- GPU node lifecycle completion may require DCGM metrics and a NIM smoke marker when profile says `ai-serving`.
- Batch/training node may require KAI scheduling smoke marker.

---

### Phase 8: In-cluster live view

#### Task 8.1: Add node lifecycle exporter

**Objective:** Expose live NICo/Kubernetes/NVIDIA node status inside the cluster.

**Files:**
- Create: `platform/node-lifecycle-exporter/Chart.yaml`
- Create: `platform/node-lifecycle-exporter/templates/deployment.yaml`
- Create: `platform/node-lifecycle-exporter/templates/servicemonitor.yaml`

**Metrics:**

```text
ubiquity_nico_machine_state
ubiquity_nico_instance_state
ubiquity_nico_task_active
ubiquity_node_kubernetes_ready
ubiquity_node_gpu_ready
ubiquity_node_rdma_ready
ubiquity_node_lifecycle_ready
```

#### Task 8.2: Add Grafana dashboard

**Files:**
- Create: `monitoring/monitoring-system/dashboards/ubiquity-nico-node-lifecycle.json`

**Panels:**
- machine count by state
- instance count by state
- active tasks
- failed provisioning operations
- GPU mismatch nodes
- RDMA mismatch nodes
- nodes not Kubernetes Ready

---

### Phase 9: Documentation overhaul

#### Task 9.1: Create admin guide

**Files:**
- Create: `docs/admin-guide/nvidia-infra-controller-node-management.md`
- Modify: `docs/admin-guide/administration/tutorials/add-or-remove-nodes.md`

**Must explain:**
- Why bootstrap is external but day-2 is NICo.
- NICo Core vs NICo REST vs site-agent.
- Prerequisites: MetalLB, cert-manager, Vault, External Secrets, PostgreSQL, Keycloak/OIDC, Temporal.
- How Ubiquity generates values for site networks, VIPs, DHCP/DNS/NTP/PXE.
- How OS images become NICo Operating Systems.
- How Machines become Instances.
- How Instances join Kubernetes.
- How GPU/RDMA/MIG validation works after join.
- How to add, remove, reinstall, and power-cycle nodes.
- That NICo is experimental/preview unless upstream status changes.

#### Task 9.2: Create OS image guide

**Files:**
- Create: `docs/reference/nvidia-infra-controller/os-images.md`

**Must include:**
- Rocky/RHEL Kickstart path.
- Ubuntu autoinstall/cloud-init path.
- Custom iPXE/user-data path.
- Checksums/provenance.
- GPU-ready image requirements.
- How to avoid committing credentials.

#### Task 9.3: Create lifecycle diagrams

**Files:**
- Create: `docs/reference/nvidia-infra-controller/bootstrap-to-nico.md`
- Create: `docs/reference/nvidia-infra-controller/node-state-machine.md`
- Create: `docs/reference/nvidia-infra-controller/status-aggregation.md`

#### Task 9.4: Create runbooks

**Files:**
- Create: `docs/admin-guide/runbooks/nico-bootstrap.md`
- Create: `docs/admin-guide/runbooks/nico-machine-provisioning.md`
- Create: `docs/admin-guide/runbooks/nico-node-reinstall.md`
- Create: `docs/admin-guide/runbooks/nico-bmc-redfish.md`
- Create: `docs/admin-guide/runbooks/nico-gpu-validation.md`

---

### Phase 10: Validation and CI

#### Task 10.1: Unit tests

Run:

```sh
go test ./pkg/... ./cmd/... -count=1
```

Must cover:
- NICo config redaction.
- NICo REST client path construction.
- `nicocli` adapter parsing.
- NodeInventory/OSImage validation.
- Operating System rendering for Rocky/RHEL/Ubuntu/custom.
- Node status aggregation.
- Safety gates.
- CLI output.

#### Task 10.2: Render tests

Run:

```sh
helm lint system/nvidia-infra-controller-core
helm lint platform/nvidia-infra-controller-rest
go run ./cmd/ubiquity test --sandbox-deploy
```

#### Task 10.3: Mock NICo E2E

Use upstream NICo local/mock capabilities where possible:

- `NVIDIA/infra-controller` local development with DevSpace/mock systems.
- NICo REST local Kind path: `make kind-reset` in upstream for reference only; Ubiquity CI should not vendor long-running upstream CI but should support a gated mock job.

Create:

```text
test/e2e/nico-mock-node-lifecycle.sh
```

Requires explicit flag:

```sh
UBIQUITY_NICO_MOCK_E2E=1 test/e2e/nico-mock-node-lifecycle.sh
```

#### Task 10.4: Real hardware E2E

Create:

```text
test/e2e/nico-bare-metal-node-lifecycle.sh
```

Requires explicit flag and sacrificial node:

```sh
UBIQUITY_NICO_BAREMETAL_E2E=1 \
UBIQUITY_NICO_SACRIFICIAL_NODE=cn-test-01 \
test/e2e/nico-bare-metal-node-lifecycle.sh
```

Proof points:
- NICo Core ready.
- NICo REST ready.
- Site-agent connected.
- Machine visible.
- Operating System created for at least one image family.
- Instance provisioning task starts and completes.
- Kubernetes Node Ready after join.
- GPU/MIG/RDMA readiness passes when selected profile requires it.
- Reinstall dry-run validates target OS image.

---

## Acceptance criteria

This implementation is complete when:

1. Ubiquity explicitly uses `NVIDIA/infra-controller` as the default day-2 bare-metal lifecycle controller.
2. BMO/Metal3 docs and scripts are marked fallback/migration-only.
3. `ubiquity up --env prod` supports `--node-lifecycle-backend nico` and installs/configures NICo through Ubiquity GitOps wrappers.
4. Ubiquity can deploy NICo Core and NICo REST prerequisites with provenance/version metadata and no committed secrets.
5. Ubiquity has a multi-OS image catalog that renders NICo Operating System objects for Rocky/RHEL, Ubuntu, and custom image types.
6. `ubiquity nodes list/status` provides a live view of all physical nodes from NICo Machine/Instance/Task/GPU APIs plus Kubernetes/NVIDIA readiness.
7. `ubiquity nodes add/remove/reinstall/power` execute through NICo APIs or `nicocli`, not shell scripts or BMO by default.
8. Safety gates prevent unsafe control-plane, storage, GPU, and forced power/reinstall operations.
9. GPU nodes are not considered lifecycle-ready until NICo, Kubernetes, and NVIDIA GPU/RDMA/MIG/DCGM profile validation all pass.
10. Documentation clearly explains initial bootstrap, NICo installation, OS images, add/remove/reinstall/power flows, live status, and troubleshooting.
11. CI includes unit tests, Helm/render tests, sandbox tests, mock NICo E2E, and gated real-hardware NICo E2E.
12. All docs and CLI output disclose NICo's upstream experimental/preview status unless NVIDIA upstream changes that status.
13. No docs or CLI output claims NVIDIA certification/approval; only source-backed integration and locally validated readiness evidence are claimed.

---

## Execution goal

Implement NVIDIA Infra Controller-backed bare-metal node lifecycle management for Ubiquity by making `NVIDIA/infra-controller` the default in-cluster day-2 control plane for physical infrastructure. Retain Ubiquity's existing Ansible/PXE path only for the bootstrap boundary required to create the initial Kubernetes management cluster, then install NICo Core, NICo REST, and site-agent through Ubiquity GitOps with explicit provenance, versions, prerequisites, and fail-closed readiness checks.

Build a Ubiquity wrapper around NICo that provides declarative site/node/OS intent, multi-OS boot image support, NICo Operating System rendering for Rocky/RHEL, Ubuntu, and custom images, Machine/Instance lifecycle operations, Task polling, and a live `ubiquity nodes` view that joins NICo physical state with Kubernetes node readiness and NVIDIA AI workload readiness. Implement `ubiquity nodes list/status/os/add/remove/reinstall/power/task` so operators can manage nodes long-term from within the Ubiquity environment while preserving control-plane quorum, storage, drain, and GPU/RDMA safety.

Use NVIDIA repositories where they provide better source-backed functionality: use `NVIDIA/infra-controller` for bare-metal lifecycle, Redfish/BMC, DHCP/PXE/DNS/NTP, machine/instance/task/status, hardware health, and DPU-enforced isolation; use `NVIDIA/k8s-launch-kit` patterns for network/RDMA discovery/profile generation; use `NVIDIA/gpu-operator`, `NVIDIA/k8s-device-plugin`, `NVIDIA/k8s-driver-manager`, `NVIDIA/dcgm-exporter`, `NVIDIA/k8s-nim-operator`, `NVIDIA/ais-k8s`, `NVIDIA/k8s-dra-driver-gpu`, and `NVIDIA/cloud-native-stack` for the AI platform layer and post-provision validation. Deprecate the BMO-centered plan as fallback/migration only.

Deliver the feature with TDD, complete admin/reference/runbook documentation, redacted secret handling, provenance metadata, mock and gated hardware E2E tests, and explicit experimental/preview disclaimers for NICo until upstream status changes.
