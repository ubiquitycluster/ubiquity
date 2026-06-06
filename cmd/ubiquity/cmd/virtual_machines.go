package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

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
	applyCmd.Flags().BoolVar(&vmApplyDryRun, "dry-run", true, "use kubectl server-side dry-run instead of mutating the cluster")
	virtualMachinesCmd.AddCommand(renderCmd, applyCmd)
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

func renderVirtualMachinesManifest() (string, error) {
	req := vmOpts
	if req.GPU.ResourceName != "" {
		req.GPU.Enabled = true
	}
	return virtualization.RenderVM(req)
}
