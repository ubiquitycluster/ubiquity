package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubiquitycluster/ubiquity/pkg/cloud"
)

type cloudOptions struct {
	DryRun             bool
	VMDisk             cloud.VMDiskRequest
	Tenant             cloud.TenantVPCRequest
	Cluster            cloud.TenantClusterRequest
	Service            cloud.ManagedServiceRequest
	Backup             cloud.BackupPolicyRequest
	ReadinessFile      string
	ReadinessResources []string
}

var cloudOpts = cloudOptions{
	DryRun: true,
	VMDisk: cloud.VMDiskRequest{
		Name:      "data-disk",
		Namespace: "virtual-machines",
		Size:      "40Gi",
		Source:    cloud.VMDiskSource{Type: cloud.VMDiskSourceBlank},
	},
	Tenant: cloud.TenantVPCRequest{
		Tenant:      "tenant-a",
		CIDR:        "10.60.0.0/24",
		Gateway:     "10.60.0.1",
		Bridge:      "br-tenant-a",
		CPUQuota:    "100",
		MemoryQuota: "512Gi",
	},
	Cluster: cloud.TenantClusterRequest{
		Name:              "tenant-a-dev",
		Namespace:         "tenant-a",
		KubernetesVersion: "v1.31.4",
		ControlPlaneClass: "kamaji",
		NodePoolClass:     "nico-managed-workers",
		WorkerReplicas:    3,
	},
	Service: cloud.ManagedServiceRequest{
		Name:         "datasets",
		Namespace:    "tenant-a",
		Type:         cloud.ServiceBucket,
		StorageClass: "standard",
		Size:         "100Gi",
		Replicas:     3,
	},
	Backup: cloud.BackupPolicyRequest{
		Name:                 "tenant-a-daily",
		Namespace:            "tenant-a",
		Schedule:             "0 2 * * *",
		Retention:            "30d",
		RepositorySecretName: "tenant-a-daily-repo",
		SnapshotClass:        "longhorn-snapshots",
		PresetName:           "gpu-medium",
		PresetCPU:            "16",
		PresetMemory:         "128Gi",
		PresetGPU:            "1",
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
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.Tenant.Tenant, "tenant", cloudOpts.Tenant.Tenant, "tenant name for VPC resources")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.Tenant.CIDR, "cidr", cloudOpts.Tenant.CIDR, "tenant VPC CIDR")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.Tenant.Gateway, "gateway", cloudOpts.Tenant.Gateway, "tenant VPC gateway")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.Tenant.Bridge, "bridge", cloudOpts.Tenant.Bridge, "tenant VPC Multus bridge")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.Tenant.CPUQuota, "cpu-quota", cloudOpts.Tenant.CPUQuota, "tenant CPU quota")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.Tenant.MemoryQuota, "memory-quota", cloudOpts.Tenant.MemoryQuota, "tenant memory quota")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.Tenant.GPUQuota, "gpu-quota", cloudOpts.Tenant.GPUQuota, "tenant GPU quota")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.Cluster.KubernetesVersion, "kubernetes-version", cloudOpts.Cluster.KubernetesVersion, "tenant cluster Kubernetes version")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.Cluster.ControlPlaneClass, "control-plane-class", cloudOpts.Cluster.ControlPlaneClass, "tenant cluster control-plane class")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.Cluster.NodePoolClass, "node-pool-class", cloudOpts.Cluster.NodePoolClass, "tenant cluster node-pool class")
	cloudCmd.PersistentFlags().IntVar(&cloudOpts.Cluster.WorkerReplicas, "worker-replicas", cloudOpts.Cluster.WorkerReplicas, "tenant cluster worker replicas")
	cloudCmd.PersistentFlags().StringVar((*string)(&cloudOpts.Service.Type), "service-type", string(cloudOpts.Service.Type), "managed service type (bucket, postgres, redis, kafka, registry)")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.Service.StorageClass, "service-storage-class", cloudOpts.Service.StorageClass, "managed service storage class")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.Service.Size, "service-size", cloudOpts.Service.Size, "managed service storage size/limit")
	cloudCmd.PersistentFlags().IntVar(&cloudOpts.Service.Replicas, "service-replicas", cloudOpts.Service.Replicas, "managed service replicas")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.Backup.Schedule, "backup-schedule", cloudOpts.Backup.Schedule, "backup five-field cron schedule")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.Backup.Retention, "backup-retention", cloudOpts.Backup.Retention, "backup retention such as 30d")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.Backup.RepositorySecretName, "backup-repository-secret", cloudOpts.Backup.RepositorySecretName, "backup repository secret name")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.Backup.SnapshotClass, "snapshot-class", cloudOpts.Backup.SnapshotClass, "VolumeSnapshotClass name")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.Backup.PresetName, "preset-name", cloudOpts.Backup.PresetName, "resource preset name")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.Backup.PresetCPU, "preset-cpu", cloudOpts.Backup.PresetCPU, "resource preset CPU")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.Backup.PresetMemory, "preset-memory", cloudOpts.Backup.PresetMemory, "resource preset memory")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.Backup.PresetGPU, "preset-gpu", cloudOpts.Backup.PresetGPU, "resource preset GPU count")
	cloudCmd.PersistentFlags().StringVar(&cloudOpts.ReadinessFile, "readiness-file", cloudOpts.ReadinessFile, "JSON cloud readiness evidence file")
	cloudCmd.PersistentFlags().StringSliceVar(&cloudOpts.ReadinessResources, "readiness-resource", nil, "resource API to collect readiness conditions from; repeat for multiple resources")

	renderCmd := &cobra.Command{Use: "render RESOURCE", Short: "Render a cloud primitive", Args: cobra.ExactArgs(1), RunE: runCloudRender}
	applyCmd := &cobra.Command{Use: "apply RESOURCE", Short: "Apply a cloud primitive", Args: cobra.ExactArgs(1), RunE: runCloudApply}
	readinessCmd := &cobra.Command{Use: "readiness", Short: "Evaluate fail-closed cloud readiness evidence", Args: cobra.NoArgs, RunE: runCloudReadiness}
	collectReadinessCmd := &cobra.Command{Use: "collect-readiness", Short: "Collect cloud readiness evidence JSON from the current cluster", Args: cobra.NoArgs, RunE: runCloudCollectReadiness}
	auditChecklistCmd := &cobra.Command{Use: "audit-checklist", Short: "Render the production cloud readiness audit checklist", Args: cobra.NoArgs, RunE: runCloudAuditChecklist}
	cloudCmd.AddCommand(renderCmd, applyCmd, readinessCmd, collectReadinessCmd, auditChecklistCmd)
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

func runCloudReadiness(cmd *cobra.Command, args []string) error {
	if cloudOpts.ReadinessFile == "" {
		return fmt.Errorf("--readiness-file is required")
	}
	content, err := os.ReadFile(cloudOpts.ReadinessFile)
	if err != nil {
		return fmt.Errorf("read readiness evidence: %w", err)
	}
	var evidence cloud.CloudReadinessEvidence
	if err := json.Unmarshal(content, &evidence); err != nil {
		return fmt.Errorf("parse readiness evidence JSON: %w", err)
	}
	fmt.Fprint(cmd.OutOrStdout(), cloud.RenderCloudReadinessReport(cloud.EvaluateCloudReadiness(evidence)))
	return nil
}

func runCloudCollectReadiness(cmd *cobra.Command, args []string) error {
	evidence, err := collectCloudReadinessEvidence(cmd.Context())
	if err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(encoded))
	return nil
}

func runCloudAuditChecklist(cmd *cobra.Command, args []string) error {
	fmt.Fprint(cmd.OutOrStdout(), cloud.RenderCloudProductionAuditChecklist())
	return nil
}

func collectCloudReadinessEvidence(ctx context.Context) (cloud.CloudReadinessEvidence, error) {
	evidence := cloud.CloudReadinessEvidence{
		RequiredCRDs: cloud.RequiredCloudCRDs(),
		SmokeTests:   map[string]bool{},
		Metadata:     map[string]string{"collector": "ubiquity cloud collect-readiness"},
	}
	out, err := runCloudKubectl(ctx, []string{"get", "crd", "-o", "json"}, nil)
	if err != nil {
		return evidence, fmt.Errorf("collect CRD readiness evidence: %w", err)
	}
	present, err := parseKubectlCRDNames(out)
	if err != nil {
		return evidence, err
	}
	evidence.PresentCRDs = present
	resources := cloudOpts.ReadinessResources
	if len(resources) == 0 {
		resources = defaultCloudReadinessResources()
	}
	for _, resource := range resources {
		resource = strings.TrimSpace(resource)
		if resource == "" {
			continue
		}
		out, err := runCloudKubectl(ctx, []string{"get", resource, "-A", "-o", "json"}, nil)
		if err != nil {
			evidence.Metadata["skipped/"+resource] = err.Error()
			continue
		}
		items, err := parseKubectlResourceEvidence(out)
		if err != nil {
			return evidence, fmt.Errorf("parse %s readiness evidence: %w", resource, err)
		}
		evidence.Resources = append(evidence.Resources, items...)
	}
	return evidence, nil
}

type kubectlList struct {
	Items []struct {
		Kind     string `json:"kind"`
		Metadata struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"metadata"`
		Status struct {
			Conditions []cloud.CloudCondition `json:"conditions"`
		} `json:"status"`
	} `json:"items"`
}

func parseKubectlCRDNames(content []byte) ([]string, error) {
	var list struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
		} `json:"items"`
	}
	if err := json.Unmarshal(content, &list); err != nil {
		return nil, fmt.Errorf("parse CRD JSON: %w", err)
	}
	names := make([]string, 0, len(list.Items))
	for _, item := range list.Items {
		if item.Metadata.Name != "" {
			names = append(names, item.Metadata.Name)
		}
	}
	return names, nil
}

func parseKubectlResourceEvidence(content []byte) ([]cloud.CloudResourceEvidence, error) {
	var list kubectlList
	if err := json.Unmarshal(content, &list); err != nil {
		return nil, err
	}
	resources := make([]cloud.CloudResourceEvidence, 0, len(list.Items))
	for _, item := range list.Items {
		resources = append(resources, cloud.CloudResourceEvidence{
			Kind:       item.Kind,
			Namespace:  item.Metadata.Namespace,
			Name:       item.Metadata.Name,
			Conditions: item.Status.Conditions,
		})
	}
	return resources, nil
}

func defaultCloudReadinessResources() []string {
	resources := []string{
		"datavolumes.cdi.kubevirt.io",
		"virtualmachines.kubevirt.io",
		"clusters.cluster.x-k8s.io",
		"schedules.k8up.io",
		"restores.k8up.io",
	}
	resources = append(resources, cloud.AllManagedServiceReadinessResources()...)
	return resources
}

func renderCloudResource(resource string) (string, error) {
	switch resource {
	case "vm-disk", "vmdisk", "disk":
		return cloud.RenderVMDisk(cloudOpts.VMDisk)
	case "vpc", "tenant-vpc", "tenant":
		if cloudOpts.Tenant.Tenant == "tenant-a" && cloudOpts.VMDisk.Name != "data-disk" {
			cloudOpts.Tenant.Tenant = cloudOpts.VMDisk.Name
		}
		return cloud.RenderTenantVPC(cloudOpts.Tenant)
	case "tenant-cluster", "kubernetes", "kubernetes-cluster":
		if cloudOpts.Cluster.Name == "tenant-a-dev" && cloudOpts.VMDisk.Name != "data-disk" {
			cloudOpts.Cluster.Name = cloudOpts.VMDisk.Name
		}
		if cloudOpts.Cluster.Namespace == "tenant-a" && cloudOpts.VMDisk.Namespace != "virtual-machines" {
			cloudOpts.Cluster.Namespace = cloudOpts.VMDisk.Namespace
		}
		return cloud.RenderTenantKubernetesCluster(cloudOpts.Cluster)
	case "service", "managed-service", "catalog":
		if cloudOpts.Service.Name == "datasets" && cloudOpts.VMDisk.Name != "data-disk" {
			cloudOpts.Service.Name = cloudOpts.VMDisk.Name
		}
		if cloudOpts.Service.Namespace == "tenant-a" && cloudOpts.VMDisk.Namespace != "virtual-machines" {
			cloudOpts.Service.Namespace = cloudOpts.VMDisk.Namespace
		}
		return cloud.RenderManagedService(cloudOpts.Service)
	case "backup", "backup-policy", "ops-policy", "platform-ops":
		if cloudOpts.Backup.Name == "tenant-a-daily" && cloudOpts.VMDisk.Name != "data-disk" {
			cloudOpts.Backup.Name = cloudOpts.VMDisk.Name
		}
		if cloudOpts.Backup.Namespace == "tenant-a" && cloudOpts.VMDisk.Namespace != "virtual-machines" {
			cloudOpts.Backup.Namespace = cloudOpts.VMDisk.Namespace
		}
		return cloud.RenderBackupPolicy(cloudOpts.Backup)
	case "restore-drill", "backup-restore", "restore":
		if cloudOpts.Backup.Name == "tenant-a-daily" && cloudOpts.VMDisk.Name != "data-disk" {
			cloudOpts.Backup.Name = cloudOpts.VMDisk.Name
		}
		if cloudOpts.Backup.Namespace == "tenant-a" && cloudOpts.VMDisk.Namespace != "virtual-machines" {
			cloudOpts.Backup.Namespace = cloudOpts.VMDisk.Namespace
		}
		return cloud.RenderBackupRestoreDrill(cloud.BackupRestoreDrillRequest{Name: cloudOpts.Backup.Name, Namespace: cloudOpts.Backup.Namespace, RepositorySecretName: cloudOpts.Backup.RepositorySecretName})
	case "prerequisites", "prereqs", "requirements":
		return cloud.RenderCloudPrerequisites(cloud.CloudPrerequisitesRequest{Name: "cloud-prereqs", Namespace: "ubiquity-system"})
	case "operator-bundles", "operators", "install-plan":
		return cloud.RenderCloudOperatorBundles(cloud.CloudOperatorBundlesRequest{Name: "cloud-operators", Namespace: "ubiquity-system"})
	case "governance", "policy-bundle", "cloud-governance":
		return cloud.RenderCloudGovernance(cloud.CloudGovernanceRequest{Name: "tenant-a-governance", Namespace: "tenant-a"})
	default:
		return "", fmt.Errorf("unsupported cloud resource %q", resource)
	}
}
