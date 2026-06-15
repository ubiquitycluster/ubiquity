# Plan 006: Fix NICo readiness service/workload contract

> **Executor instructions**: Follow this plan step by step. Encode the default-rendered NICo chart semantics in tests before changing readiness logic.
>
> **Drift check (run first)**: `git diff --stat efd46ed..HEAD -- pkg/nico/readiness.go pkg/nico/readiness_test.go system/nvidia-infra-controller-core/values.yaml system/nvidia-infra-controller-core/templates/services.yaml`

## Status

- **Priority**: P1
- **Effort**: S/M
- **Risk**: MED
- **Depends on**: none
- **Category**: correctness
- **Planned at**: commit `efd46ed`, 2026-06-15
- **Status**: DONE (implemented and verified)

## Why this matters

`ubiquity health --nico` should fail closed on real missing readiness evidence, not fail forever because default-rendered components intentionally do not expose Kubernetes Services. False-negative readiness weakens NCP/NICo acceptance evidence.

## Current state

- `pkg/nico/readiness.go:28` includes `nico-dhcp`, `nico-dns`, and `nico-ntp` in required Services.
- `pkg/nico/readiness.go:47-50` fails when any required Service is absent.
- `system/nvidia-infra-controller-core/templates/services.yaml:1-23` renders Services only when `component.service.enabled` is true.
- `system/nvidia-infra-controller-core/values.yaml:70-103` disables Services for `nico-dhcp`, `nico-dns`, and `nico-ntp` by default.

## Commands you will need

| Purpose | Command | Expected on success |
|---|---|---|
| Focused tests | `go test ./pkg/nico -count=1` | exits 0 |
| Chart render smoke | `helm template nico system/nvidia-infra-controller-core >/tmp/nico-core.yaml` | exits 0 |

## Scope

**In scope**:
- `pkg/nico/readiness.go`
- `pkg/nico/readiness_test.go`

**Out of scope**:
- Changing NICo core chart defaults unless readiness cannot be made faithful to the chart contract.

## Steps

1. Split required component names from required service names.
2. Keep all NICo components in `ChartComponentNames()` for status and documentation coverage.
3. Require Services only for components whose default chart values expose Services: `nico-api`, `nico-bmc-proxy`, `nico-hardware-health`, `nico-pxe`, `nico-ssh-console-rs`, `nico-rest-api`, `nico-rest-site-agent`.
4. Add a regression proving readiness can pass when DHCP/DNS/NTP workloads are ready but do not expose Services.
5. Update missing-foundation tests to expect service failures only for service-backed components.

## Done criteria

- [ ] Readiness no longer requires non-rendered DHCP/DNS/NTP Services.
- [ ] `ChartComponentNames()` still includes all default NICo components.
- [ ] `go test ./pkg/nico -count=1` passes.

## STOP conditions

Stop if upstream NICo chart semantics require Services for DHCP/DNS/NTP in production overlays and the defaults are only sandbox placeholders; report the overlay evidence.
