# NVIDIA NIC Configuration Operator

Status: experimental/preview. This page documents Ubiquity's wrapper for the upstream NVIDIA NIC Configuration Operator at `https://github.com/Mellanox/nic-configuration-operator`. It is not a certification or support claim.

The operator configures NVIDIA NIC firmware and firmware-related settings through Kubernetes CRDs. It is separate from NVIDIA Infra Controller (NICo): NICo remains Ubiquity's preferred day-2 physical node lifecycle backend, while the NIC Configuration Operator handles NIC-level configuration and firmware workflows on already managed nodes.

## Ubiquity chart

Ubiquity vendors the upstream Helm chart as:

```sh
system/nvidia-nic-configuration-operator
```

The chart installs the upstream CRDs, manager Deployment, configuration DaemonSet, RBAC, and supported firmware ConfigMap. It is excluded from the root GitOps ApplicationSet unless `bootstrap/root` is rendered with `nico.enabled=true`, matching the rest of the NICo/NVIDIA node-lifecycle wrappers.

## Prerequisites

Before enabling this chart, verify:

1. The NVIDIA Network Operator is installed and healthy.
2. NVIDIA/Mellanox NIC nodes are labelled by Node Feature Discovery.
3. The NVIDIA Maintenance Operator CRD/API is installed, because the NIC Configuration Operator uses `maintenance.nvidia.com/NodeMaintenance` for coordinated changes.
4. Operators have an approved maintenance window. NIC configuration and firmware updates can reboot or temporarily drain nodes.
5. Firmware-update storage ownership is explicit. Ubiquity does not create NIC firmware PVCs against an implicit default StorageClass.

## Safe defaults

The wrapper intentionally differs from the upstream defaults in two places:

- `nicFirmwareStorage.create=false` and `nicFirmwareStorage.pvcName=""` by default. Inventory/configuration can run without firmware-update storage, and firmware storage must be opted in explicitly.
- If `nicFirmwareStorage.create=true`, `nicFirmwareStorage.storageClassName` is required. This prevents accidental binding to a cluster default StorageClass.

## Example values

```yaml
nicFirmwareStorage:
  create: true
  pvcName: nic-fw-storage-pvc
  storageClassName: nico-firmware-rwx
  availableStorageSize: 10Gi

nicConfigurationTemplates:
  - name: connectx6dx-roce
    nodeSelector:
      feature.node.kubernetes.io/network-sriov.capable: "true"
    nicSelector:
      nicType: "101d"
      partNumbers: []
      pciAddresses: []
      serialNumbers: []
    resetToDefault: false
    template:
      numVfs: 0
      linkType: Ethernet
      pciPerformanceOptimized:
        enabled: true
        maxReadRequest: 4096
      roceOptimized:
        enabled: true
        qos:
          trust: dscp
          pfc: "0,0,0,1,0,0,0,0"
          tos: 0
      gpuDirectOptimized:
        enabled: true
        env: Baremetal
```

## Validation commands

Render locally before enabling GitOps:

```sh
helm lint system/nvidia-nic-configuration-operator
helm template nvidia-nic-configuration-operator system/nvidia-nic-configuration-operator --namespace network-operator
helm unittest system/nvidia-nic-configuration-operator
```

Cluster checks after deployment:

```sh
kubectl -n network-operator get deploy,ds,pods -l app.kubernetes.io/part-of=nic-configuration-operator
kubectl get crd | grep configuration.net.nvidia.com
kubectl -n network-operator get nicdevices.configuration.net.nvidia.com
kubectl -n network-operator get nicconfigurationtemplates.configuration.net.nvidia.com
```

## Operational cautions

- Prefer PCI addresses or part numbers for tightly scoped templates; serial numbers can be unsafe on embedded systems with shared VPD images.
- Do not apply overlapping templates. Upstream reports an error when more than one template matches one device, and none are applied.
- `gpuDirectOptimized` requires `pciPerformanceOptimized.enabled=true`.
- `spectrumXOptimized` should not be combined with explicit `roceOptimized`; Spectrum-X settings include RoCE behavior upstream.
- Treat `resetToDefault=true`, firmware templates, and firmware sources as destructive maintenance actions requiring explicit change control.
