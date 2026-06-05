# NVIDIA Infra Controller node lifecycle

Status: experimental/preview documentation reset for Ubiquity day-2 bare-metal lifecycle management.

Ubiquity is aligning on NVIDIA Infra Controller as the preferred day-2 lifecycle manager for on-premises bare-metal nodes. The target upstream is NVIDIA/infra-controller. This page describes the architecture boundary and vocabulary Ubiquity should use while integrating with that project.

This document is intentionally descriptive rather than a certification statement. It does not claim hardware, software, or support certification.

## Role in Ubiquity

NVIDIA Infra Controller owns day-2 bare-metal lifecycle operations after the initial Ubiquity bootstrap boundary. In this model, Ubiquity can still prepare a site, networking, credentials, and the first management cluster, but ongoing node lifecycle work should flow through NVIDIA Infra Controller rather than ad-hoc Metal3/Ironic helper scripts.

The bootstrap boundary is the line between:

- bootstrap-time work: enough PXE, image, Kubernetes, and credential setup to bring the management plane online; and
- day-2 work: discover, inventory, provision, reinstall, power-cycle, deprovision, monitor, and audit machines after the management plane exists.

Legacy BareMetalHost, baremetal-operator, and disk-image scripts remain fallback/migration-only paths for existing environments that have not moved to NVIDIA Infra Controller yet.

## Components

- NICo Core: the controller logic that reconciles inventory, lifecycle state, requested operations, and Task progress for the bare-metal fleet.
- NICo REST: the API surface used by Ubiquity or operators to request lifecycle operations and read state.
- site-agent: the per-site connector that reports local hardware and executes site-scoped operations from the management plane.
- NVIDIA/infra-controller: the upstream repository identity used by this documentation when referring to NVIDIA Infra Controller.

## Vocabulary mapping

Ubiquity documentation should use NVIDIA Infra Controller terms when describing the day-2 lifecycle:

- Operating System: the bootable OS image, version, configuration, and install intent assigned to a machine.
- Machine: a physical bare-metal server under inventory and lifecycle control.
- Instance: the realized allocation or installed runtime of a Machine with an Operating System and workload placement context.
- Task: an auditable operation such as discover, install, reinstall, deprovision, reboot, or collect inventory.
- Machine GPU stats: GPU inventory and utilization evidence associated with a Machine, including accelerator presence and health signals when available.

## Lifecycle flow

1. Ubiquity bootstraps the management plane and crosses the bootstrap boundary.
2. A site-agent registers site-local inventory with NICo Core.
3. Operators or Ubiquity integrations call NICo REST to request Operating System assignment, Machine lifecycle operations, or status queries.
4. NICo Core records and reconciles a Task for each lifecycle operation.
5. Machine status, Instance status, and Machine GPU stats are reported back through NVIDIA Infra Controller APIs.

## Migration posture

For new day-2 lifecycle automation, prefer NVIDIA Infra Controller. Metal3 BareMetalHost references and helper scripts in this repository are retained only as fallback/migration-only documentation for clusters already depending on them.

Because this integration is experimental/preview, keep operational runbooks explicit about which path is active at a site. Do not mix legacy BareMetalHost automation and NVIDIA Infra Controller automation for the same Machine unless a migration plan says how ownership is transferred.
