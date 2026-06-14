package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubiquitycluster/ubiquity/pkg/virtualization"
)

var vmOpts = virtualization.VMRequest{
	Name:      "ubuntu-gpu-dev",
	Namespace: "virtual-machines",
	OS:        "ubuntu-24.04",
	CPUCores:  4,
	Memory:    "16Gi",
	DiskSize:  "80Gi",
	Network: virtualization.NetworkRequest{
		Isolation: virtualization.NetworkIsolationPod,
		Name:      "vm-isolated-net",
		CIDR:      "10.43.0.0/24",
		Gateway:   "10.43.0.1",
		Bridge:    "br-vm-isolated",
	},
	GPU: virtualization.GPURequest{Enabled: false, Count: 1, ResourceName: "nvidia.com/GA100_A100_PCIE_40GB"},
}
var vmApplyDryRun = true
var vmAttachDisks []string

var runVirtualMachinesKubectl = func(ctx context.Context, args []string, stdin []byte) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Stdin = bytes.NewReader(stdin)
	return cmd.CombinedOutput()
}

var virtualMachinesCmd = &cobra.Command{
	Use:     "virtual-machines",
	Aliases: []string{"vms", "vm"},
	Short:   "Render or apply KubeVirt virtual machines",
	Long: `Render or apply KubeVirt VirtualMachine, CDI DataVolume, Multus NetworkAttachmentDefinition,
and NetworkPolicy resources. GPU VM access requires KubeVirt permittedHostDevices plus
NVIDIA GPU Operator/device-plugin resources on VM-capable nodes.`,
}

func init() {
	virtualMachinesCmd.PersistentFlags().StringVar(&vmOpts.Name, "name", vmOpts.Name, "VM name")
	virtualMachinesCmd.PersistentFlags().StringVar(&vmOpts.Namespace, "namespace", vmOpts.Namespace, "VM namespace/tenant")
	virtualMachinesCmd.PersistentFlags().StringVar(&vmOpts.OS, "os", vmOpts.OS, "OS profile (ubuntu-24.04, rocky-9, windows-2022)")
	virtualMachinesCmd.PersistentFlags().StringVar(&vmOpts.InstanceType, "instance-type", vmOpts.InstanceType, "optional KubeVirt VirtualMachineClusterInstancetype reference")
	virtualMachinesCmd.PersistentFlags().StringVar(&vmOpts.Preference, "preference", vmOpts.Preference, "optional KubeVirt VirtualMachineClusterPreference reference")
	virtualMachinesCmd.PersistentFlags().IntVar(&vmOpts.CPUCores, "cpu", vmOpts.CPUCores, "VM CPU cores")
	virtualMachinesCmd.PersistentFlags().StringVar(&vmOpts.Memory, "memory", vmOpts.Memory, "VM memory request")
	virtualMachinesCmd.PersistentFlags().StringVar(&vmOpts.DiskSize, "disk-size", vmOpts.DiskSize, "VM root disk size")
	virtualMachinesCmd.PersistentFlags().StringVar(&vmOpts.BootDisk, "boot-disk", vmOpts.BootDisk, "existing standalone PVC to use as boot disk; skips root DataVolume rendering")
	virtualMachinesCmd.PersistentFlags().StringArrayVar(&vmAttachDisks, "attach-disk", nil, "attach existing PVC as data disk in name:pvc form; repeat for multiple disks")
	virtualMachinesCmd.PersistentFlags().StringVar(&vmOpts.StorageClass, "storage-class", vmOpts.StorageClass, "CDI DataVolume storage class")
	virtualMachinesCmd.PersistentFlags().StringVar((*string)(&vmOpts.Network.Isolation), "network-isolation", string(vmOpts.Network.Isolation), "network isolation mode (pod, multus)")
	virtualMachinesCmd.PersistentFlags().StringVar(&vmOpts.Network.Name, "network-name", vmOpts.Network.Name, "Multus NetworkAttachmentDefinition name")
	virtualMachinesCmd.PersistentFlags().StringVar(&vmOpts.Network.CIDR, "network-cidr", vmOpts.Network.CIDR, "isolated network CIDR")
	virtualMachinesCmd.PersistentFlags().StringVar(&vmOpts.Network.Gateway, "network-gateway", vmOpts.Network.Gateway, "isolated network gateway")
	virtualMachinesCmd.PersistentFlags().StringVar(&vmOpts.Network.Bridge, "network-bridge", vmOpts.Network.Bridge, "bridge name for Multus bridge CNI")
	virtualMachinesCmd.PersistentFlags().StringVar(&vmOpts.GPU.ResourceName, "gpu-resource", "", "KubeVirt GPU deviceName/resourceName such as nvidia.com/GA100_A100_PCIE_40GB; enables GPU when set")
	virtualMachinesCmd.PersistentFlags().StringVar((*string)(&vmOpts.GPU.Mode), "gpu-attachment-mode", string(virtualization.GPUAttachmentGPU), "KubeVirt GPU attachment mode (gpu, hostDevice)")
	virtualMachinesCmd.PersistentFlags().IntVar(&vmOpts.GPU.Count, "gpu-count", vmOpts.GPU.Count, "number of GPU devices to attach")
	virtualMachinesCmd.PersistentFlags().BoolVar(&vmOpts.External.Enabled, "external", vmOpts.External.Enabled, "render a LoadBalancer Service for selected VM ports")
	virtualMachinesCmd.PersistentFlags().IntSliceVar(&vmOpts.External.Ports, "external-port", []int{22}, "external TCP port to expose; repeat for multiple ports")
	virtualMachinesCmd.PersistentFlags().StringVar(&vmOpts.SSHAuthorizedKey, "ssh-key", "replace-with-authorized-key", "SSH authorized key for cloud-init")

	renderCmd := &cobra.Command{Use: "render", Short: "Render KubeVirt VM manifests", Args: cobra.NoArgs, RunE: runVirtualMachinesRender}
	applyCmd := &cobra.Command{Use: "apply", Short: "Apply KubeVirt VM manifests", Args: cobra.NoArgs, RunE: runVirtualMachinesApply}
	imageCatalogCmd := &cobra.Command{Use: "image-catalog", Short: "Render supported VM image catalog", Args: cobra.NoArgs, RunE: runVirtualMachinesImageCatalog}
	readinessCmd := &cobra.Command{Use: "readiness", Short: "Collect fail-closed KubeVirt VM readiness evidence", Args: cobra.NoArgs, RunE: runVirtualMachinesReadiness}
	applyCmd.Flags().BoolVar(&vmApplyDryRun, "dry-run", true, "use kubectl server-side dry-run instead of mutating the cluster")
	virtualMachinesCmd.AddCommand(renderCmd, applyCmd, imageCatalogCmd, readinessCmd)
	rootCmd.AddCommand(virtualMachinesCmd)
}

func runVirtualMachinesRender(cmd *cobra.Command, args []string) error {
	manifest, err := renderVirtualMachinesManifest()
	if err != nil {
		return err
	}
	fmt.Print(manifest)
	return nil
}

func runVirtualMachinesImageCatalog(cmd *cobra.Command, args []string) error {
	manifest, err := virtualization.RenderVMImageCatalog(virtualization.VMImageCatalogRequest{Name: "vm-image-catalog", Namespace: "ubiquity-system"})
	if err != nil {
		return err
	}
	fmt.Print(manifest)
	return nil
}

func runVirtualMachinesApply(cmd *cobra.Command, args []string) error {
	manifest, err := renderVirtualMachinesManifest()
	if err != nil {
		return err
	}
	kubectlArgs := []string{"apply"}
	if vmApplyDryRun {
		kubectlArgs = append(kubectlArgs, "--dry-run=server")
	}
	kubectlArgs = append(kubectlArgs, "-f", "-")
	out, err := runVirtualMachinesKubectl(cmd.Context(), kubectlArgs, []byte(manifest))
	if len(out) > 0 {
		fmt.Print(string(out))
	}
	return err
}

func runVirtualMachinesReadiness(cmd *cobra.Command, args []string) error {
	evidence := collectVirtualMachinesReadinessEvidence(cmd.Context(), vmOpts)
	status := virtualization.EvaluateVMReadiness(evidence)
	fmt.Print(renderVirtualMachinesReadinessStatus(status))
	if !status.Ready {
		return fmt.Errorf("KubeVirt VM %s/%s is not ready", defaultName(vmOpts.Namespace, "default"), defaultName(vmOpts.Name, "unknown"))
	}
	return nil
}

func collectVirtualMachinesReadinessEvidence(ctx context.Context, req virtualization.VMRequest) virtualization.VMReadinessEvidence {
	bootDisk := req.BootDisk
	if bootDisk == "" {
		bootDisk = req.Name + "-root"
	}
	return virtualization.VMReadinessEvidence{
		Namespace:                  req.Namespace,
		Name:                       req.Name,
		BootDiskName:               bootDisk,
		DataVolumeImportReady:      kubectlOutputEquals(ctx, []string{"-n", req.Namespace, "get", "datavolume", bootDisk, "-o", "jsonpath={.status.conditions[?(@.type==\"Ready\")].status}"}, "True"),
		PersistentVolumeClaimBound: kubectlOutputEquals(ctx, []string{"-n", req.Namespace, "get", "pvc", bootDisk, "-o", "jsonpath={.status.phase}"}, "Bound"),
		VirtualMachineReady:        kubectlOutputEquals(ctx, []string{"-n", req.Namespace, "get", "virtualmachine", req.Name, "-o", "jsonpath={.status.conditions[?(@.type==\"Ready\")].status}"}, "True"),
		VirtualMachineRunning:      kubectlOutputEquals(ctx, []string{"-n", req.Namespace, "get", "virtualmachineinstance", req.Name, "-o", "jsonpath={.status.phase}"}, "Running"),
		GuestAgentReady:            kubectlCommandSucceeds(ctx, []string{"-n", req.Namespace, "get", "configmap", req.Name + "-guest-health-passed"}),
	}
}

func renderVirtualMachinesReadinessStatus(status virtualization.VMReadinessStatus) string {
	var b strings.Builder
	fmt.Fprintf(&b, "ready: %t\n", status.Ready)
	if status.Ready {
		b.WriteString("reasons:\n  []\n")
		b.WriteString("guest health evidence present\n")
		return b.String()
	}
	b.WriteString("reasons:\n")
	for _, reason := range status.Reasons {
		fmt.Fprintf(&b, "  - %s\n", reason)
	}
	b.WriteString("policy: fail closed; render/catalog proof is not CDI import, PVC, VM boot, or guest health proof\n")
	return b.String()
}

func kubectlOutputEquals(ctx context.Context, args []string, expected string) bool {
	out, err := runVirtualMachinesKubectl(ctx, args, nil)
	return err == nil && strings.TrimSpace(string(out)) == expected
}

func kubectlCommandSucceeds(ctx context.Context, args []string) bool {
	_, err := runVirtualMachinesKubectl(ctx, args, nil)
	return err == nil
}

func defaultName(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func renderVirtualMachinesManifest() (string, error) {
	req := vmOpts
	if req.GPU.ResourceName != "" {
		req.GPU.Enabled = true
	}
	if len(vmAttachDisks) > 0 {
		disks, err := parseDiskAttachments(vmAttachDisks)
		if err != nil {
			return "", err
		}
		req.DataDisks = append(req.DataDisks, disks...)
	}
	return virtualization.RenderVM(req)
}

func parseDiskAttachments(values []string) ([]virtualization.DiskAttachment, error) {
	disks := make([]virtualization.DiskAttachment, 0, len(values))
	for _, value := range values {
		parts := strings.Split(value, ":")
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			return nil, fmt.Errorf("--attach-disk must use name:pvc form, got %q", value)
		}
		disks = append(disks, virtualization.DiskAttachment{Name: strings.TrimSpace(parts[0]), PVCName: strings.TrimSpace(parts[1])})
	}
	return disks, nil
}
