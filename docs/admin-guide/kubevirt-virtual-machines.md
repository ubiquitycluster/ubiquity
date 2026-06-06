# KubeVirt virtual machines on Ubiquity

Ubiquity can create virtual machines on bootstrapped Kubernetes nodes with KubeVirt. The VM layer is intentionally declarative: Ubiquity renders or applies KubeVirt `VirtualMachine`, CDI `DataVolume`, Multus `NetworkAttachmentDefinition`, and Kubernetes `NetworkPolicy` resources. This is not a GPU readiness or NVIDIA certification claim; production readiness still requires live KubeVirt, CDI, Multus, NVIDIA GPU Operator, and hardware evidence.

## Components

- KubeVirt provides the Kubernetes `VirtualMachine` API.
- Containerized Data Importer imports operating-system cloud images into PVC-backed `DataVolume` disks.
- Multus attaches optional isolated secondary networks to VM pods.
- NetworkPolicy defaults can fail closed for tenant VM traffic.
- NVIDIA GPU Operator and KubeVirt `permittedHostDevices` are required before a GPU resource such as `nvidia.com/GA100_A100_PCIE_40GB` can be passed into a VM.

## Supported OS profiles

The initial OS profiles are:

- `ubuntu-24.04`: Ubuntu Noble cloud image from `cloud-images.ubuntu.com`.
- `rocky-9`: Rocky Linux 9 GenericCloud image.
- `windows-2022`: operator-supplied licensed Windows Server 2022 image with Cloudbase-Init and virtio-win drivers. The default URL is intentionally a placeholder; do not commit licensed or credentialed image URLs.

## CLI usage

Render the VM image catalog before creating disks or VMs:

```sh
ubiquity virtual-machines image-catalog
```

The image catalog records supported OS profiles and the readiness boundary `import-and-guest-boot-not-proven-by-catalog`; operators must still prove CDI import, VM boot, and guest-level health before claiming an image is usable.

Render a CPU-only Ubuntu VM:

```sh
ubiquity virtual-machines render \
  --name ubuntu-dev \
  --namespace tenant-a \
  --os ubuntu-24.04
```

Render a Rocky Linux VM with a KubeVirt GPU device, Multus network isolation, a reusable KubeVirt instance type/profile, and selected external ports:

```sh
ubiquity virtual-machines render \
  --name rocky-gpu \
  --namespace tenant-a \
  --os rocky-9 \
  --instance-type gx-a100-medium \
  --preference ubuntu-server \
  --network-isolation multus \
  --network-name tenant-a-rdma \
  --network-bridge br-tenant-a \
  --network-cidr 10.44.0.0/24 \
  --network-gateway 10.44.0.1 \
  --gpu-resource nvidia.com/GA100_A100_PCIE_40GB \
  --gpu-attachment-mode gpu \
  --gpu-count 1 \
  --external \
  --external-port 22 \
  --external-port 443
```

When `--instance-type` is present, CPU and memory are taken from the referenced KubeVirt `VirtualMachineClusterInstancetype` instead of being duplicated in the VM manifest. For PCI VF / vGPU resources, set `--gpu-attachment-mode hostDevice`; for classic KubeVirt GPU resources, keep the default `gpu` mode.

Render a VM that boots from an existing standalone disk and attaches reusable data disks created by `ubiquity cloud render vm-disk`:

```sh
ubiquity virtual-machines render \
  --name ai-notebook \
  --namespace tenant-a \
  --boot-disk golden-ubuntu \
  --attach-disk datasets:datasets-pvc \
  --attach-disk checkpoints:checkpoints-pvc
```

When `--boot-disk` is set, Ubiquity references the existing PVC as `bootdisk` and skips rendering a new root CDI `DataVolume`. Each `--attach-disk` value uses `name:pvc` form and renders a KubeVirt disk plus `persistentVolumeClaim` volume.

Apply uses server-side dry-run by default:

```sh
ubiquity virtual-machines apply --name ubuntu-dev --os ubuntu-24.04
```

Use `--dry-run=false` only against a cluster with the KubeVirt, CDI, and any required Multus/NVIDIA GPU prerequisites already installed.

## Helm usage

The same contract is available as a chart:

```sh
helm template kubevirt-vms platform/kubevirt-vms \
  --namespace virtual-machines \
  --set vm.os=ubuntu-24.04
```

For network isolation:

```sh
helm template kubevirt-vms platform/kubevirt-vms \
  --namespace tenant-a \
  --set vm.namespace=tenant-a \
  --set vm.networkIsolation=multus \
  --set vm.network.name=tenant-a-rdma
```

## GPU VM requirements

KubeVirt GPU assignment is expressed under `spec.template.spec.domain.devices.gpus` with a `deviceName` matching a resource exposed to Kubernetes. On NVIDIA clusters this requires:

1. NVIDIA GPU Operator/device-plugin exposing GPU or vGPU resources on VM-capable nodes.
2. KubeVirt configured with `permittedHostDevices` for the same resource name.
3. Node placement policies that keep VM workloads on nodes that can satisfy the requested GPU resource.
4. A live `nvidia-smi` or equivalent in-guest validation before readiness is claimed.

## Network isolation requirements

When `networkIsolation=multus`, Ubiquity renders a Multus `NetworkAttachmentDefinition` and attaches it to the VM as `isolated-net`. A default-deny `NetworkPolicy` is also rendered to make pod-network ingress/egress fail closed until operators add explicit tenant policy. For RDMA or SR-IOV data planes, replace the bridge CNI example with the Network Operator/SR-IOV/RDMA network attachment appropriate to the site.

## What local sandbox validation proves

A CPU-only sandbox can prove deterministic rendering, Helm lint/template success, server-side schema acceptance after CRDs are installed, and that VM manifests include the right KubeVirt/CDI/Multus/GPU fields. It does not prove guest boot, in-guest GPU access, RDMA readiness, Windows licensing, or NVIDIA certification.
