# NICo Day-2 Node Lifecycle Proof Implementation Plan

> **For Hermes:** Use test-driven-development skill to implement this plan task-by-task.

**Goal:** Close NICo/NVIDIA Infra Controller production-grade proof gaps with live cluster validation gates, explicit day-2 lifecycle commands, multi-OS image compatibility checks, long-term status aggregation, destructive-action safety gates, virtual bare-metal validation, and BMO/Metal3 fallback boundaries.

**Architecture:** Keep NICo as the default physical node lifecycle backend. Add reviewer-visible contracts at three layers: pure Go evaluators/parsers, CLI/live-operation wrappers that fail closed, and gated scripts/docs that prove real or virtual Kubernetes/NICo/BMC/PXE evidence without running by default.

**Tech Stack:** Go/Cobra CLI, existing `pkg/nico`, `pkg/nodestatus`, `pkg/nodeinventory`, Helm/kubectl smoke checks, qemu-bmc/containerlab virtual bare-metal fixtures, markdown docs-as-code tests.

---

### Task 1: Plan artifact and baseline discovery

**Objective:** Save this implementation plan and verify the working tree baseline before code changes.

**Files:**
- Create: `docs/plans/2026-06-09-nico-day2-lifecycle-proof.md`

**Verification:**
- `git status --short --branch`
- Commit: `docs: plan NICo day-2 lifecycle proof`

### Task 2: CLI surface for day-2 lifecycle verbs

**Objective:** Make the operator-facing `ubiquity nodes` command expose the required day-2 terms directly: enroll, inspect, image, reboot, cordon/drain, maintenance, and status reconcile.

**Files:**
- Modify: `cmd/ubiquity/cmd/nodes.go`
- Test: `cmd/ubiquity/cmd/nodes_test.go`

**TDD:**
1. Add a failing test requiring subcommands `enroll`, `inspect`, `image`, `reboot`, `cordon`, `maintenance`, and `status reconcile`.
2. Wire aliases/wrappers to existing safe implementations:
   - enroll -> add
   - inspect -> status
   - image -> os apply
   - reboot -> power reset
   - cordon -> non-evict cordon command
   - maintenance -> drain/cordon guarded operation
   - status reconcile -> live joined status output
3. Verify dry-run/mock output and destructive gates remain fail-closed.

### Task 3: Long-term node status aggregation fields

**Objective:** Extend joined node status output so it can serve as durable day-2 status aggregation for BMC, kubelet, GPU, RDMA, firmware/image, and maintenance state.

**Files:**
- Modify: `pkg/nodestatus/status.go`
- Modify: `pkg/nodestatus/collect_nico.go`
- Test: `pkg/nodestatus/collect_nico_test.go`

**TDD:**
1. Add a failing test that expects fields `bmcStatus`, `kubeletStatus`, `gpuStatus`, `rdmaStatus`, `firmwareStatus`, `imageStatus`, and `maintenanceState` in collected `NodeStatus`.
2. Populate fields from existing NICo/Kubernetes evidence conservatively.
3. Fail closed/degraded when evidence is absent or inconsistent.

### Task 4: Multi-OS image compatibility and safe fallback

**Objective:** Prove bootable Ubuntu/RHEL/Rocky/custom OS profiles include provenance, compatibility, and explicit fallback behavior.

**Files:**
- Modify: `pkg/nodeinventory/render_nico_os.go`
- Test: `pkg/nodeinventory/render_nico_os_test.go`
- Possibly modify: `examples/node-inventory/nico-prod.yaml`

**TDD:**
1. Add failing tests for OS compatibility labels and fallback behavior on unsupported/unsafe images.
2. Require provenance/checksum/architecture and boot mode metadata on every rendered NICo OS object.
3. Reject unsupported image families unless custom boot data is explicitly supplied.

### Task 5: Destructive-action safety gates at CLI boundary

**Objective:** Ensure reboot, power cycle, wipe/reimage, drain/evict, and maintenance operations resolve live NICo/Kubernetes status before mutating and require exact confirmation/acknowledgements.

**Files:**
- Modify: `pkg/nodestatus/safety.go`
- Modify: `cmd/ubiquity/cmd/nodes.go`
- Test: `pkg/nodestatus/safety_test.go`
- Test: `cmd/ubiquity/cmd/nodes_live_test.go`

**TDD:**
1. Add tests for reset/reboot, reimage/wipe, power off, drain/evict, and maintenance mode requiring exact `--confirm <node>`.
2. Add tests that Kubernetes Ready nodes require cordon/drain acknowledgement unless operation itself is the drain/cordon step.
3. Verify ambiguous/missing targets fail before client mutation.

### Task 6: Gated live NICo controller validation script

**Objective:** Add a live validation script that proves NICo controller readiness and day-2 operations against a real or virtual Kubernetes cluster while skipped by default.

**Files:**
- Create/modify: `test/e2e/nico-day2-lifecycle-proof.sh`
- Test: `pkg/aiplatform/nico_day2_lifecycle_test.go`

**TDD:**
1. Add docs/script test requiring the script to be gated by `UBIQUITY_RUN_NICO_DAY2=true` and support `--dry-run`.
2. Require evidence markers for enroll, inspect, image, reboot, cordon/drain, maintenance, and status reconcile.
3. The script must call `ubiquity health --nico`, `ubiquity nodes status`, and gated day-2 commands.
4. It must record `nico-day2-lifecycle-proof-passed` only after live proof succeeds.

### Task 7: Virtual bare-metal validation tier hardening

**Objective:** Ensure qemu/KVM, emulated Redfish/IPMI, PXE, and CI-safe validation artifacts are explicitly present and integrated into the day-2 proof path.

**Files:**
- Modify: `test/e2e/nico-kvm-pxe-lab.sh`
- Modify/add fixtures under `test/fixtures/nico-kvm-pxe/`
- Test: `pkg/aiplatform/kvm_lab_docs_test.go`

**TDD:**
1. Add failing tests requiring qemu-bmc/containerlab, Redfish, IPMI, PXE MAC, OS image, and secret references.
2. Verify the script is skipped by default and uses `UBIQUITY_NICO_KVM_LAB=1`.
3. Verify the day-2 proof can compose this lab before physical hardware.

### Task 8: BMO/Metal3 fallback boundary docs-as-code

**Objective:** Keep NICo as default and document BMO/Metal3 only as fallback/migration-only.

**Files:**
- Modify: `docs/architecture/on-prem/nvidia-infra-controller-node-lifecycle.md`
- Modify: `docs/admin-guide/runbooks/nico-bootstrap.md`
- Test: `pkg/aiplatform/nico_docs_test.go`

**TDD:**
1. Add tests requiring docs to state NICo default and BMO/Metal3 fallback/migration-only.
2. Ensure CLI `--backend bmo` fails closed and does not imply live BMO lifecycle support.

### Task 9: Full verification and graph update

**Objective:** Verify all code/docs/scripts against the plan.

**Commands:**
- `go test ./pkg/... ./cmd/... -count=1`
- Active Helm lint/render loop for all non-disabled charts
- `bash -n test/e2e/*.sh`
- Dry-run scripts:
  - `test/e2e/nico-day2-lifecycle-proof.sh --dry-run`
  - `test/e2e/nico-kvm-pxe-lab.sh --dry-run`
  - `test/e2e/nico-virtual-bare-metal-lab.sh --dry-run`
- Fail-closed commands:
  - `go run ./cmd/ubiquity health --nico`
  - `go run ./cmd/ubiquity nodes reboot node-a --confirm node-a --dry-run=false` without NICo config must fail closed
- `git diff --check`
- `graphify update .` if available; otherwise record skip.

**Completion:** all tests pass, all scripts are syntax-valid and gated, graph update attempted, git status clean after commits.
