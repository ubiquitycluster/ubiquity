package virtualization

import (
	"fmt"
	"regexp"
	"strings"
)

// NetworkIsolation records the VM network isolation mode.
type NetworkIsolation string

const (
	NetworkIsolationPod    NetworkIsolation = "pod"
	NetworkIsolationMultus NetworkIsolation = "multus"
)

// GPUAttachmentMode selects the KubeVirt device field used for NVIDIA GPU resources.
type GPUAttachmentMode string

const (
	// GPUAttachmentGPU renders domain.devices.gpus and is suitable for classic GPU passthrough resources.
	GPUAttachmentGPU GPUAttachmentMode = "gpu"
	// GPUAttachmentHostDevice renders domain.devices.hostDevices, the preferred shape for PCI VF/vGPU resources.
	GPUAttachmentHostDevice GPUAttachmentMode = "hostDevice"
)

// GPURequest describes GPU passthrough/vGPU resource requests for KubeVirt.
type GPURequest struct {
	Enabled      bool
	ResourceName string
	Count        int
	Mode         GPUAttachmentMode
}

// ExternalAccess describes optional Service exposure for selected VM ports.
type ExternalAccess struct {
	Enabled bool
	Ports   []int
}

// NetworkRequest describes an isolated secondary network for VM workloads.
type NetworkRequest struct {
	Isolation NetworkIsolation
	Name      string
	CIDR      string
	Gateway   string
	Bridge    string
}

// DiskAttachment describes an existing PVC attached to a VM as a persistent data disk.
type DiskAttachment struct {
	Name    string
	PVCName string
}

// VMRequest is the reviewer-visible VM provisioning contract rendered into KubeVirt resources.
type VMRequest struct {
	Name             string
	Namespace        string
	OS               string
	InstanceType     string
	Preference       string
	CPUCores         int
	Memory           string
	DiskSize         string
	StorageClass     string
	BootDisk         string
	DataDisks        []DiskAttachment
	Network          NetworkRequest
	GPU              GPURequest
	External         ExternalAccess
	SSHAuthorizedKey string
}

// OSProfile defines an operating-system image known to be renderable as a KubeVirt VM.
type OSProfile struct {
	Name        string
	DisplayName string
	SourceURL   string
	CloudInit   bool
	Notes       string
}

var dnsName = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

var osProfiles = map[string]OSProfile{
	"ubuntu-24.04": {
		Name:        "ubuntu-24.04",
		DisplayName: "Ubuntu Server 24.04 LTS",
		SourceURL:   "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img",
		CloudInit:   true,
		Notes:       "Ubuntu Noble cloud image imported by CDI DataVolume.",
	},
	"rocky-9": {
		Name:        "rocky-9",
		DisplayName: "Rocky Linux 9 GenericCloud",
		SourceURL:   "https://dl.rockylinux.org/pub/rocky/9/images/x86_64/Rocky-9-GenericCloud-Base.latest.x86_64.qcow2",
		CloudInit:   true,
		Notes:       "Rocky 9 GenericCloud image imported by CDI DataVolume.",
	},
	"windows-2022": {
		Name:        "windows-2022",
		DisplayName: "Windows Server 2022",
		SourceURL:   "https://example.invalid/replace-with-windows-2022-cloudbase-init-and-virtio-win-image.qcow2",
		CloudInit:   true,
		Notes:       "Placeholder for a licensed Windows Server 2022 image with Cloudbase-Init and virtio-win drivers; operators must supply an authorized image URL.",
	},
}

// OSProfiles returns known OS profiles in a stable order.
func OSProfiles() []OSProfile {
	return []OSProfile{osProfiles["ubuntu-24.04"], osProfiles["rocky-9"], osProfiles["windows-2022"]}
}

// RenderVM renders KubeVirt/CDI/Multus resources for a single VM.
func RenderVM(req VMRequest) (string, error) {
	req = defaultVMRequest(req)
	if err := validateVMRequest(req); err != nil {
		return "", err
	}
	profile := osProfiles[req.OS]
	var b strings.Builder
	writeNamespace(&b, req)
	if req.Network.Isolation == NetworkIsolationMultus {
		writeNAD(&b, req)
	}
	writeNetworkPolicy(&b, req)
	if req.External.Enabled {
		writeExternalService(&b, req)
	}
	if req.BootDisk == "" {
		writeDataVolume(&b, req, profile)
	}
	writeVM(&b, req, profile)
	return b.String(), nil
}

func defaultVMRequest(req VMRequest) VMRequest {
	if req.Namespace == "" {
		req.Namespace = "virtual-machines"
	}
	if req.OS == "" {
		req.OS = "ubuntu-24.04"
	}
	if req.CPUCores == 0 {
		req.CPUCores = 2
	}
	if req.Memory == "" {
		req.Memory = "4Gi"
	}
	if req.DiskSize == "" {
		req.DiskSize = "40Gi"
	}
	if req.Network.Isolation == "" {
		req.Network.Isolation = NetworkIsolationPod
	}
	if req.Network.Name == "" {
		req.Network.Name = req.Name + "-net"
	}
	if req.Network.CIDR == "" {
		req.Network.CIDR = "10.42.0.0/24"
	}
	if req.Network.Gateway == "" {
		req.Network.Gateway = "10.42.0.1"
	}
	if req.Network.Bridge == "" {
		req.Network.Bridge = "br-" + req.Namespace
	}
	if req.GPU.Enabled && req.GPU.Count == 0 {
		req.GPU.Count = 1
	}
	if req.GPU.Enabled && req.GPU.Mode == "" {
		req.GPU.Mode = GPUAttachmentGPU
	}
	return req
}

func validateVMRequest(req VMRequest) error {
	if !dnsName.MatchString(req.Name) {
		return fmt.Errorf("VM name %q must be a DNS-compatible lowercase name", req.Name)
	}
	if !dnsName.MatchString(req.Namespace) {
		return fmt.Errorf("VM namespace %q must be a DNS-compatible lowercase name", req.Namespace)
	}
	if _, ok := osProfiles[req.OS]; !ok {
		return fmt.Errorf("unknown VM OS profile %q", req.OS)
	}
	if req.Network.Isolation != NetworkIsolationPod && req.Network.Isolation != NetworkIsolationMultus {
		return fmt.Errorf("unsupported network isolation mode %q", req.Network.Isolation)
	}
	if req.Network.Isolation == NetworkIsolationMultus && !dnsName.MatchString(req.Network.Name) {
		return fmt.Errorf("network attachment name %q must be DNS-compatible", req.Network.Name)
	}
	if req.GPU.Enabled && strings.TrimSpace(req.GPU.ResourceName) == "" {
		return fmt.Errorf("GPU-enabled VMs require a KubeVirt permittedHostDevices resource name such as nvidia.com/GA100_A100_PCIE_40GB")
	}
	if req.GPU.Enabled && req.GPU.Mode != GPUAttachmentGPU && req.GPU.Mode != GPUAttachmentHostDevice {
		return fmt.Errorf("unsupported GPU attachment mode %q", req.GPU.Mode)
	}
	if req.BootDisk != "" && !dnsName.MatchString(req.BootDisk) {
		return fmt.Errorf("boot disk %q must be a DNS-compatible PVC name", req.BootDisk)
	}
	seenDisks := map[string]struct{}{}
	for _, disk := range req.DataDisks {
		if !dnsName.MatchString(disk.Name) || !dnsName.MatchString(disk.PVCName) {
			return fmt.Errorf("disk attachment names must be DNS-compatible: %+v", disk)
		}
		if disk.Name == "rootdisk" || disk.Name == "bootdisk" || disk.Name == "cloudinitdisk" {
			return fmt.Errorf("disk attachment name %q is reserved", disk.Name)
		}
		if _, ok := seenDisks[disk.Name]; ok {
			return fmt.Errorf("disk attachment name %q is duplicated", disk.Name)
		}
		seenDisks[disk.Name] = struct{}{}
	}
	if req.External.Enabled {
		if len(req.External.Ports) == 0 {
			return fmt.Errorf("external VM access requires at least one external port")
		}
		for _, port := range req.External.Ports {
			if port < 1 || port > 65535 {
				return fmt.Errorf("external port %d must be in range 1-65535", port)
			}
		}
	}
	return nil
}

func writeNamespace(b *strings.Builder, req VMRequest) {
	fmt.Fprintf(b, `apiVersion: v1
kind: Namespace
metadata:
  name: %s
  labels:
    app.kubernetes.io/part-of: ubiquity-virtual-machines
    ubiquity.ai/vm-tenant: %s
---
`, req.Namespace, req.Namespace)
}

func writeNAD(b *strings.Builder, req VMRequest) {
	fmt.Fprintf(b, `apiVersion: k8s.cni.cncf.io/v1
kind: NetworkAttachmentDefinition
metadata:
  name: %s
  namespace: %s
  labels:
    app.kubernetes.io/part-of: ubiquity-virtual-machines
    ubiquity.ai/network-isolation: multus
spec:
  config: '{"cniVersion":"0.3.1","type":"bridge","bridge":"%s","ipam":{"type":"host-local","subnet":"%s","gateway":"%s"}}'
---
`, req.Network.Name, req.Namespace, req.Network.Bridge, req.Network.CIDR, req.Network.Gateway)
}

func writeNetworkPolicy(b *strings.Builder, req VMRequest) {
	fmt.Fprintf(b, `apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: %s-default-deny
  namespace: %s
  labels:
    app.kubernetes.io/part-of: ubiquity-virtual-machines
spec:
  podSelector:
    matchLabels:
      kubevirt.io/domain: %s
  policyTypes:
    - Ingress
    - Egress
---
`, req.Name, req.Namespace, req.Name)
}

func writeExternalService(b *strings.Builder, req VMRequest) {
	fmt.Fprintf(b, `apiVersion: v1
kind: Service
metadata:
  name: %s-external
  namespace: %s
  labels:
    app.kubernetes.io/part-of: ubiquity-virtual-machines
    ubiquity.ai/external-access: port-list
spec:
  type: LoadBalancer
  selector:
    kubevirt.io/domain: %s
  ports:
`, req.Name, req.Namespace, req.Name)
	for _, port := range req.External.Ports {
		fmt.Fprintf(b, "    - name: tcp-%d\n      port: %d\n      targetPort: %d\n      protocol: TCP\n", port, port, port)
	}
	b.WriteString("---\n")
}

func writeDataVolume(b *strings.Builder, req VMRequest, profile OSProfile) {
	fmt.Fprintf(b, `apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: %s-rootdisk
  namespace: %s
  labels:
    app.kubernetes.io/part-of: ubiquity-virtual-machines
    ubiquity.ai/os-profile: %s
spec:
  source:
    http:
      url: %s
  pvc:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: %s
`, req.Name, req.Namespace, req.OS, profile.SourceURL, req.DiskSize)
	if req.StorageClass != "" {
		fmt.Fprintf(b, "    storageClassName: %s\n", req.StorageClass)
	}
	b.WriteString("---\n")
}

func writeVM(b *strings.Builder, req VMRequest, profile OSProfile) {
	fmt.Fprintf(b, `apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: %s
  namespace: %s
  labels:
    app.kubernetes.io/part-of: ubiquity-virtual-machines
    kubevirt.io/domain: %s
    ubiquity.ai/os-profile: %s
    ubiquity.ai/network-isolation: %s
  annotations:
    ubiquity.ai/source: kubevirt-cdi-multus
    ubiquity.ai/os-display-name: %q
    ubiquity.ai/gpu-note: "KubeVirt GPU access requires permittedHostDevices and NVIDIA GPU Operator/device-plugin resources on VM-capable nodes."
spec:
  runStrategy: Manual
`, req.Name, req.Namespace, req.Name, req.OS, req.Network.Isolation, profile.DisplayName)
	if req.InstanceType != "" {
		fmt.Fprintf(b, "  instancetype:\n    name: %s\n    kind: VirtualMachineClusterInstancetype\n", req.InstanceType)
	}
	if req.Preference != "" {
		fmt.Fprintf(b, "  preference:\n    name: %s\n    kind: VirtualMachineClusterPreference\n", req.Preference)
	}
	fmt.Fprintf(b, `  template:
    metadata:
      labels:
        kubevirt.io/domain: %s
        ubiquity.ai/vm-name: %s
    spec:
      domain:
`, req.Name, req.Name)
	if req.InstanceType == "" {
		fmt.Fprintf(b, `        cpu:
          cores: %d
        resources:
          requests:
            memory: %s
`, req.CPUCores, req.Memory)
	}
	fmt.Fprintf(b, `        devices:
          disks:
`)
	rootDiskName := "rootdisk"
	if req.BootDisk != "" {
		rootDiskName = "bootdisk"
	}
	fmt.Fprintf(b, "            - name: %s\n              disk:\n                bus: virtio\n", rootDiskName)
	for _, disk := range req.DataDisks {
		fmt.Fprintf(b, "            - name: %s\n              disk:\n                bus: virtio\n", disk.Name)
	}
	fmt.Fprintf(b, `            - name: cloudinitdisk
              disk:
                bus: virtio
`)
	if req.GPU.Enabled {
		heading := "gpus"
		prefix := "gpu"
		if req.GPU.Mode == GPUAttachmentHostDevice {
			heading = "hostDevices"
			prefix = "hostdev"
		}
		fmt.Fprintf(b, "          %s:\n", heading)
		for i := 0; i < req.GPU.Count; i++ {
			fmt.Fprintf(b, "            - name: %s%d\n              deviceName: %s\n", prefix, i, req.GPU.ResourceName)
		}
	}
	fmt.Fprintf(b, `      networks:
        - name: podnet
          pod: {}
`)
	if req.Network.Isolation == NetworkIsolationMultus {
		fmt.Fprintf(b, "        - name: isolated-net\n          multus:\n            networkName: %s\n", req.Network.Name)
	}
	fmt.Fprintf(b, `      volumes:
`)
	if req.BootDisk != "" {
		fmt.Fprintf(b, "        - name: bootdisk\n          persistentVolumeClaim:\n            claimName: %s\n", req.BootDisk)
	} else {
		fmt.Fprintf(b, "        - name: rootdisk\n          dataVolume:\n            name: %s-rootdisk\n", req.Name)
	}
	for _, disk := range req.DataDisks {
		fmt.Fprintf(b, "        - name: %s\n          persistentVolumeClaim:\n            claimName: %s\n", disk.Name, disk.PVCName)
	}
	fmt.Fprintf(b, `        - name: cloudinitdisk
          cloudInitNoCloud:
            userData: |-
              #cloud-config
              ssh_authorized_keys:
                - %s
              package_update: true
`, req.SSHAuthorizedKey)
}
