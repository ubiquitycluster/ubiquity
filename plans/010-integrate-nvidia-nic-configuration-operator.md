# Plan 010: Integrate NVIDIA NIC Configuration Operator

> **Executor instructions**: This plan reconciles the wider NICo/NVIDIA NIC Configuration Operator implementation work already present in the working tree. Keep the implementation project-native; do not copy upstream controller logic. Use the chart wrapper, CRD vendoring, safety values, tests, and docs as the acceptance boundary.
>
> **Drift check (run first)**: `git diff --stat efd46ed..HEAD -- system/nvidia-nic-configuration-operator docs/admin-guide/nvidia-infra-controller-node-management.md docs/admin-guide/runbooks/nvidia-infra-controller/nic-configuration-operator.md bootstrap/root/templates/stack.yaml bootstrap/root/values.yaml`

## Status

- **Priority**: P1
- **Effort**: L
- **Risk**: HIGH
- **Depends on**: none
- **Category**: nico/networking/platform-integration
- **Planned at**: commit `efd46ed`, 2026-06-15
- **Status**: DONE (implemented and verified)

## Why this matters

Ubiquity's bare-metal/node-lifecycle roadmap defaults to NICo for day-2 physical node management. NVIDIA NIC Configuration Operator support is required to manage NIC firmware/configuration safely without turning BMO/Metal3 into the default lifecycle path or introducing nscale-specific dependencies.

## Reconciled dirty files

- `system/nvidia-nic-configuration-operator/` — first-party Helm wrapper for NVIDIA NIC Configuration Operator CRDs, RBAC, operator/daemon resources, firmware storage opt-in, supported firmware ConfigMap, schema, and unit tests.
- `docs/admin-guide/runbooks/nvidia-infra-controller/nic-configuration-operator.md` — operator runbook.
- `docs/admin-guide/nvidia-infra-controller-node-management.md` — day-2 node-management documentation updates.
- `bootstrap/root/templates/stack.yaml` — GitOps exclusion/project handling for explicit NICo lifecycle installation.
- `bootstrap/root/values.yaml` — root-level NICo/NIC Configuration Operator opt-in boundary.

## Acceptance evidence

- `helm unittest system/nvidia-nic-configuration-operator`
- `helm lint system/nvidia-nic-configuration-operator`
- `go run ./cmd/ubiquity test --sandbox-deploy`
- `go test ./pkg/... ./cmd/... -count=1`

## Done criteria

- [x] Wrapper chart renders operator, daemon, RBAC, service account, CRDs, and safety templates.
- [x] Schema/tests cover unsafe GPUDirect/RDMA validation inputs and explicit firmware storage opt-in.
- [x] Root GitOps does not deploy NICo/NIC Configuration Operator wrappers globally unless explicitly enabled.
- [x] Documentation/runbook describes supported day-2 operation and boundaries.
- [x] Focused Helm, sandbox render, Go, and build gates pass.
