# Runbook: NICo BMC Redfish validation

Status: experimental/preview. This runbook validates BMC Redfish reachability for NICo-managed Machines without exposing credentials. It is not a hardware certification claim.

## Preconditions

- The Machine is dedicated to NICo or under a documented migration plan.
- BMC network routing and firewall rules are approved.
- BMC credentials are stored in Vault or an equivalent secret manager.
- Operators must not paste BMC passwords into shell history, Git, docs, or tickets.

## What to validate

- The BMC address belongs to the intended physical Machine.
- Redfish endpoint is reachable from the site-agent or approved BMC proxy path.
- Power state can be read.
- Boot override and power operations are permitted only during maintenance.
- Firmware and event logs are available for troubleshooting if the site policy allows.

## Read-only procedure

1. Confirm projected secret names without printing values:

```sh
kubectl -n nico-system get externalsecret,secret --ignore-not-found | grep -i bmc || true
kubectl -n nvidia-infra-controller get externalsecret,secret --ignore-not-found | grep -i bmc || true
```

2. Confirm Machine inventory:

```sh
export NICO_MACHINE=worker-gpu-01
nicoctl machine get "${NICO_MACHINE}" --output yaml
```

3. Request BMC status through NICo or the approved site tool:

```sh
nicoctl bmc status --machine "${NICO_MACHINE}" --output yaml
```

4. Record only non-secret evidence:

```yaml
machine: worker-gpu-01
bmcReachable: true
redfishService: reachable
powerState: On
credentialSource: external-secret-name-only
secretsExposed: false
```

## Optional controlled power test

Run only in a maintenance window on a Machine that is safe to reboot:

```sh
nicoctl task create power-cycle --machine "${NICO_MACHINE}" --output yaml
nicoctl task wait --machine "${NICO_MACHINE}" --for condition=Succeeded --timeout 20m
```

Validate the Machine returns to the expected state:

```sh
nicoctl machine get "${NICO_MACHINE}" --output yaml
kubectl get node "${NICO_MACHINE}" -o wide || true
```

## Failure handling

- If Redfish is unreachable, verify network path, DNS, firewall, BMC VLAN, and site-agent location.
- If authentication fails, rotate credentials in the secret manager; do not print or copy secret values.
- If power state is wrong, stop destructive actions and involve the hardware owner.
- Preserve BMC event logs and NICo Task IDs in the change record.

## Safety notes

- Redfish power actions can interrupt workloads.
- Do not use vendor default credentials.
- Do not embed credentials in URLs.
- Do not commit BMC inventory files containing passwords.
