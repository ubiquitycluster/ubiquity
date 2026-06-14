# Bootstrap to NICo boundary

Status: experimental/preview. This reference defines the handoff from Ubiquity bootstrap automation to NVIDIA Infra Controller (NICo) day-2 lifecycle management. It is not a certification or support statement.

## Boundary definition

The bootstrap boundary is crossed when Ubiquity has created enough management-plane capability for NICo to own ongoing bare-metal lifecycle operations.

Before the boundary, Ubiquity may prepare:

- management cluster installation;
- GitOps controllers;
- networking and DNS required by the management cluster;
- Vault or equivalent secret manager integration;
- cert-manager and External Secrets Operator;
- LoadBalancer services for bare-metal management endpoints;
- initial credentials and PostgreSQL connectivity;
- initial NICo Core, NICo REST, and site-agent deployment.

After the boundary, NICo should own:

- Machine discovery and inventory;
- BMC Redfish power and boot operations;
- Operating System assignment;
- install, reinstall, deprovision, reboot, and inventory collection Tasks;
- Machine, Instance, and Machine GPU stats state;
- day-2 node lifecycle audit evidence.

## Handoff criteria

A site is ready to hand off day-2 lifecycle operations to NICo when:

1. NICo Core and REST workloads are ready.
2. site-agent is ready and can see the intended site.
3. NICo services for REST, PXE/provisioning, BMC access, DNS/NTP dependencies, and hardware health are either ready or explicitly not used by the deployment mode.
4. Required secrets are projected from an approved secret manager.
5. PostgreSQL persistence and backups are assigned to an owner.
6. At least one non-production Machine can be discovered or represented in inventory.
7. Operators have a documented rollback path.

## Ownership rule

A physical Machine must have a single active lifecycle owner:

- `bootstrap`: temporary ownership by Ubiquity bootstrap automation.
- `nico`: ownership by NVIDIA Infra Controller.
- `legacy-metal3`: fallback/migration-only ownership by BareMetalHost/Ironic automation.
- `manual-breakglass`: time-limited human intervention recorded in a change ticket.

Do not mix NICo and legacy BareMetalHost actions for the same Machine unless the migration runbook explicitly marks the transfer point.

## Recommended handoff record

Record the following in the site operations log:

```yaml
site: example-site
handoffDate: 2026-06-05
status: experimental-preview
ownerBefore: bootstrap
ownerAfter: nico
nicoNamespace: nico-system
coreReady: true
restReady: true
siteAgentReady: true
secretSource: external-secret-store-name
postgresqlOwner: platform-database-team
firstMachine: worker-gpu-01
notes: No certification claim. Local day-2 lifecycle preview enabled.
```

## Rollback

Rollback means returning day-2 automation to a known previous owner, not running two systems at once. Before rollback:

1. Stop new NICo Tasks.
2. Wait for active Tasks to complete or document cancellation.
3. Export Machine inventory and Task evidence.
4. Disable NICo automation for the affected Machines.
5. Re-enable the previous process only after ownership is clear.

If rollback targets legacy BareMetalHost automation, treat it as fallback/migration-only and document why NICo ownership is being suspended.
