# ADR-014: Use NVIDIA Infra Controller for day-2 node lifecycle

## Status

Accepted

## Context

Ubiquity needs a production default for long-term physical node lifecycle management on NVIDIA accelerated clusters. The platform must inspect live node state, manage bootable Operating System image choices, and gate destructive operations such as drain, reboot, reinstall, and maintenance entry.

Metal3/BareMetal Operator remains useful for bootstrap and migration paths, but running two independent day-2 lifecycle controllers against the same physical fleet can create ownership ambiguity and safety risk.

## Decision

Use NVIDIA Infra Controller (NICo) as the default in-cluster day-2 physical node manager for NVIDIA hardware.

Keep Metal3/BMO as fallback, bootstrap, or migration-only integration unless a deployment explicitly opts into that path. Ubiquity-owned CLI and documentation must make the NICo boundary clear and must keep destructive lifecycle verbs behind explicit confirmation and drain acknowledgement gates.

## Consequences

- NICo status, Task, Machine, and Operating System evidence is the default source for node lifecycle readiness.
- BMO/Metal3 documentation must describe migration boundaries rather than implying dual ownership.
- Multi-OS image catalogs need source, checksum, rollback, architecture, boot-mode, and driver-stack metadata.
- Live proof can claim only observed local readiness. It is not NVIDIA approved or NVIDIA certified without attached external approval evidence.
