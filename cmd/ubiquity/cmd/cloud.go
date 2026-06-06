package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/ubiquitycluster/ubiquity/pkg/cloud"
)

type cloudOptions struct {
	DryRun bool
	VMDisk cloud.VMDiskRequest
}

var cloudOpts = cloudOptions{
	DryRun: true,
	VMDisk: cloud.VMDiskRequest{
		Name:      "data-disk",
		Namespace: "virtual-machines",
		Size:      "40Gi",
		Source:    cloud.VMDiskSource{Type: cloud.VMDiskSourceBlank},
	},
}

var cloudCmd = &cobra.Command{
	Use:     "cloud",
	Aliases: []string{"self-service"},
	Short:   "Render and apply Ubiquity cloud service primitives",
}

var runCloudKubectl = func(ctx context.Context, args []string, stdin []byte) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Stdin = bytes.NewReader(stdin)
	return cmd.CombinedOutput()
}

func init() {
	cloudCmd.PersistentFlags().BoolVar(&cloudOpts.DryRun, "dry-run", true, "use kubectl server-side dry-run for apply")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.VMDisk.Name, "name", cloudOpts.VMDisk.Name, "resource name")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.VMDisk.Namespace, "namespace", cloudOpts.VMDisk.Namespace, "target namespace/tenant")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.VMDisk.Size, "size", cloudOpts.VMDisk.Size, "disk size")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.VMDisk.StorageClass, "storage-class", cloudOpts.VMDisk.StorageClass, "storage class name")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.VMDisk.AccessMode, "access-mode", "ReadWriteOnce", "PVC access mode")
	cloudCmd.PersistentFlags().StringVar((*string)(&cloudOpts.VMDisk.Source.Type), "source", string(cloud.VMDiskSourceBlank), "VM disk source (blank, http, pvc)")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.VMDisk.Source.URL, "source-url", "", "HTTP source URL for imported disks")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.VMDisk.Source.PVCName, "source-pvc", "", "source PVC name for cloned disks")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.VMDisk.Source.PVCNamespace, "source-pvc-namespace", "", "source PVC namespace for cloned disks")

	renderCmd := &cobra.Command{Use: "render RESOURCE", Short: "Render a cloud primitive", Args: cobra.ExactArgs(1), RunE: runCloudRender}
	applyCmd := &cobra.Command{Use: "apply RESOURCE", Short: "Apply a cloud primitive", Args: cobra.ExactArgs(1), RunE: runCloudApply}
	cloudCmd.AddCommand(renderCmd, applyCmd)
	rootCmd.AddCommand(cloudCmd)
}

func runCloudRender(cmd *cobra.Command, args []string) error {
	manifest, err := renderCloudResource(args[0])
	if err != nil {
		return err
	}
	fmt.Fprint(cmd.OutOrStdout(), manifest)
	return nil
}

func runCloudApply(cmd *cobra.Command, args []string) error {
	manifest, err := renderCloudResource(args[0])
	if err != nil {
		return err
	}
	kubectlArgs := []string{"apply"}
	if cloudOpts.DryRun {
		kubectlArgs = append(kubectlArgs, "--dry-run=server")
	}
	kubectlArgs = append(kubectlArgs, "-f", "-")
	out, err := runCloudKubectl(cmd.Context(), kubectlArgs, []byte(manifest))
	fmt.Fprint(cmd.OutOrStdout(), string(out))
	if err != nil {
		return fmt.Errorf("kubectl %v failed: %w", kubectlArgs, err)
	}
	return nil
}

func renderCloudResource(resource string) (string, error) {
	switch resource {
	case "vm-disk", "vmdisk", "disk":
		return cloud.RenderVMDisk(cloudOpts.VMDisk)
	default:
		return "", fmt.Errorf("unsupported cloud resource %q", resource)
	}
}
