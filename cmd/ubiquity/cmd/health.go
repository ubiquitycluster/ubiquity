package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubiquitycluster/ubiquity/pkg/aiplatform"
	"github.com/ubiquitycluster/ubiquity/pkg/nico"
)

type healthCheck struct {
	name  string
	check func() error
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check cluster health",
	Long:  `Runs health checks against the cluster: kubectl connectivity, core components.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		aiOnly, err := cmd.Flags().GetBool("ai")
		if err != nil {
			return err
		}
		if aiOnly {
			status := collectAIReadinessSnapshot()
			fmt.Print(renderAIReadinessStatus(status))
			if !status.Ready {
				return fmt.Errorf("NVIDIA AI platform is not ready")
			}
			return nil
		}
		nicoOnly, err := cmd.Flags().GetBool("nico")
		if err != nil {
			return err
		}
		if nicoOnly {
			fmt.Print(renderNICOReadinessStatus(collectNICOReadinessSnapshot(false)))
			return nil
		}

		checks := []healthCheck{
			{"kubectl connectivity", func() error {
				return exec.Command("kubectl", "cluster-info").Run()
			}},
			{"ArgoCD server", func() error {
				return exec.Command("kubectl", "-n", "argocd", "get", "pod", "-l", "app.kubernetes.io/name=argocd-server").Run()
			}},
		}

		allPassed := true
		for _, c := range checks {
			fmt.Printf("  %s ... ", c.name)
			if err := c.check(); err != nil {
				fmt.Printf("FAIL (%v)\n", err)
				allPassed = false
			} else {
				fmt.Println("OK")
			}
		}

		if allPassed {
			fmt.Println("\nAll checks passed.")
		} else {
			fmt.Println("\nSome checks failed. Run 'ubiquity logs' for details.")
		}

		fmt.Print(renderAIReadinessStatus(collectAIReadinessSnapshot()))
		fmt.Print(renderAIStoreReadinessStatus(collectAIStoreReadinessSnapshot()))
		fmt.Print(renderNICOReadinessStatus(collectNICOReadinessSnapshot(false)))
		return nil
	},
}

func init() {
	healthCmd.Flags().Bool("ai", false, "run only NVIDIA AI platform readiness checks and fail closed when evidence is missing")
	healthCmd.Flags().Bool("nico", false, "run only NVIDIA Infra Controller readiness checks")
	rootCmd.AddCommand(healthCmd)
}

func collectNICOReadinessSnapshot(realHardware bool) nico.ReadinessResult {
	snapshot := nico.ReadinessSnapshot{Services: map[string]bool{}}
	// Fail closed by default. We only mark readiness from explicit Kubernetes evidence.
	for _, workload := range nico.ChartComponentNames() {
		ready := deploymentsAvailable("nico-system", workload) || deploymentsAvailable("nvidia-infra-controller", workload) || deploymentComponentReady(workload)
		snapshot.Workloads = append(snapshot.Workloads, nico.WorkloadStatus{Name: workload, Ready: ready})
	}
	if deploymentComponentReady("nico-rest-api") || deploymentsAvailable("nico-system", "nico-rest-api") || deploymentsAvailable("nvidia-infra-controller", "nico-rest-api") {
		snapshot.RESTAPIReady = true
	} else {
		// Backstop for older installs: CRDs alone do not prove REST API availability,
		// but preserve compatibility when wrapper labels are unavailable.
		snapshot.RESTAPIReady = exec.Command("kubectl", "get", "crd", "sites.nvidia.com").Run() == nil || exec.Command("kubectl", "get", "crd", "machines.nvidia.com").Run() == nil
	}
	if readyPodCount("nico-system", "app.kubernetes.io/component=nico-rest-site-agent") > 0 || readyPodCount("nvidia-infra-controller", "app.kubernetes.io/component=nico-rest-site-agent") > 0 || readyPodCountAllNamespaces("app.kubernetes.io/component=nico-rest-site-agent") > 0 {
		snapshot.SiteAgentReady = true
	}
	for _, svc := range nico.ChartComponentNames() {
		if exec.Command("kubectl", "-n", "nico-system", "get", "service", svc).Run() == nil || exec.Command("kubectl", "-n", "nvidia-infra-controller", "get", "service", svc).Run() == nil || serviceComponentReady(svc) {
			snapshot.Services[svc] = true
		}
	}
	return nico.EvaluateReadiness(snapshot, nico.ReadinessOptions{RealHardware: realHardware})
}

func deploymentComponentReady(component string) bool {
	return readyPodCountAllNamespaces("app.kubernetes.io/component="+component) > 0
}

func serviceComponentReady(component string) bool {
	output, err := exec.Command("kubectl", "get", "service", "-A", "-l", "app.kubernetes.io/component="+component, "-o", "name").Output()
	return err == nil && strings.TrimSpace(string(output)) != ""
}

func renderNICOReadinessStatus(status nico.ReadinessResult) string {
	var b strings.Builder
	if status.Ready {
		b.WriteString("\nNVIDIA Infra Controller bare-metal lifecycle readiness: READY\n")
	} else {
		b.WriteString("\nNVIDIA Infra Controller bare-metal lifecycle readiness: NOT READY\n")
	}
	if len(status.Failures) == 0 {
		b.WriteString("  all required NICo lifecycle signals present\n")
	} else {
		for _, failure := range status.Failures {
			b.WriteString(fmt.Sprintf("  FAIL — %s\n", failure))
		}
	}
	b.WriteString("  policy: fail closed; Ubiquity does not claim bare-metal lifecycle ready unless NICo is ready\n")
	return b.String()
}

func collectAIReadinessSnapshot() aiplatform.ReadinessStatus {
	// Conservative default: only mark a check ready when Kubernetes evidence can
	// be collected successfully. Any command failure leaves the evidence false.
	snapshot := aiplatform.ClusterSnapshot{
		GPUAllocatableByNode: map[string]int{},
		MIGAllocatableByNode: map[string]map[string]int{},
		RDMAResourcesByNode:  map[string]int{},
	}

	if deploymentsAvailable("gpu-operator", "gpu-operator") {
		snapshot.GPUOperatorReady = true
	}
	if daemonSetsReady("gpu-operator", "nvidia-device-plugin-daemonset") {
		snapshot.GPUDevicePluginReady = true
	}
	if exec.Command("kubectl", "-n", "gpu-operator", "get", "service", "nvidia-dcgm-exporter").Run() == nil ||
		exec.Command("kubectl", "-n", "monitoring-system", "get", "service", "dcgm-exporter").Run() == nil {
		snapshot.DCGMMetricsScraped = true
	}
	if output, err := exec.Command("kubectl", "get", "nodes", "-o", "json").Output(); err == nil {
		if gpusByNode, parseErr := aiplatform.ParseGPUAllocatableByNode(output); parseErr == nil {
			snapshot.GPUAllocatableByNode = gpusByNode
		}
		if migByNode, parseErr := aiplatform.ParseMIGAllocatableByNode(output); parseErr == nil {
			snapshot.MIGAllocatableByNode = migByNode
		}
		if rdmaByNode, parseErr := aiplatform.ParseAllocatableResourceByNode(output, "nvidia.com/rdma"); parseErr == nil {
			snapshot.RDMAResourcesByNode = rdmaByNode
		}
	}
	if output, err := exec.Command("kubectl", "get", "network-attachment-definitions.k8s.cni.cncf.io", "-A", "-o", "json").Output(); err == nil {
		if attachments, parseErr := aiplatform.ParseNetworkAttachments(output); parseErr == nil {
			snapshot.NetworkAttachments = attachments
		}
	}
	if exec.Command("kubectl", "-n", "gpu-operator", "get", "configmap", "rdma-network-smoke-test-passed").Run() == nil ||
		exec.Command("kubectl", "-n", "network-operator", "get", "configmap", "rdma-network-smoke-test-passed").Run() == nil {
		snapshot.LastRDMASmokeTestPassed = true
	}
	if exec.Command("kubectl", "-n", "nim-operator", "get", "configmap", "nim-smoke-test-passed").Run() == nil {
		snapshot.LastNIMSmokeTestPassed = true
		snapshot.NIMServicesReady = 1
	}
	if deploymentsAvailable("kai-scheduler", "kai-operator", "kai-scheduler-default", "binder", "admission", "pod-grouper", "podgroup-controller", "queue-controller") {
		snapshot.KAISchedulerReady = true
	}
	if exec.Command("kubectl", "get", "queues.scheduling.run.ai", "default-queue").Run() == nil {
		snapshot.KAIQueueReady = true
	}
	if exec.Command("kubectl", "-n", "kai-scheduler", "get", "configmap", "kai-scheduling-smoke-test-passed").Run() == nil {
		snapshot.LastKAISchedulingSmokeTestPassed = true
	}

	return aiplatform.EvaluateReadiness(snapshot)
}

func deploymentsAvailable(namespace string, names ...string) bool {
	if len(names) == 0 {
		return false
	}
	args := append([]string{"-n", namespace, "get", "deploy", "-o", "json"}, names...)
	output, err := exec.Command("kubectl", args...).Output()
	if err != nil {
		return false
	}
	available, err := aiplatform.ParseAvailableDeployments(output)
	if err != nil {
		return false
	}
	for _, name := range names {
		if !available[name] {
			return false
		}
	}
	return true
}

func daemonSetsReady(namespace string, names ...string) bool {
	if len(names) == 0 {
		return false
	}
	args := append([]string{"-n", namespace, "get", "daemonset", "-o", "json"}, names...)
	output, err := exec.Command("kubectl", args...).Output()
	if err != nil {
		return false
	}
	ready, err := aiplatform.ParseReadyDaemonSets(output)
	if err != nil {
		return false
	}
	for _, name := range names {
		if !ready[name] {
			return false
		}
	}
	return true
}

func readyPodCount(namespace, selector string) int {
	output, err := exec.Command("kubectl", "-n", namespace, "get", "pods", "-l", selector, "-o", "json").Output()
	if err != nil {
		return 0
	}
	count, err := aiplatform.ParseReadyPodCount(output)
	if err != nil {
		return 0
	}
	return count
}

func readyPodCountAllNamespaces(selector string) int {
	output, err := exec.Command("kubectl", "get", "pods", "-A", "-l", selector, "-o", "json").Output()
	if err != nil {
		return 0
	}
	count, err := aiplatform.ParseReadyPodCount(output)
	if err != nil {
		return 0
	}
	return count
}

func collectAIStoreReadinessSnapshot() aiplatform.ReadinessStatus {
	snapshot := aiplatform.AIStoreSnapshot{}
	if deploymentsAvailable("ais-operator-system", "ais-operator-controller-manager") ||
		deploymentsAvailable("aistore", "ais-operator-controller-manager") {
		snapshot.OperatorReady = true
	}
	if exec.Command("kubectl", "get", "crd", "aistores.ais.nvidia.com").Run() == nil ||
		exec.Command("kubectl", "get", "crd", "aisclusters.ais.nvidia.com").Run() == nil {
		snapshot.CRDsEstablished = true
	}
	if exec.Command("kubectl", "-n", "aistore", "get", "aistore").Run() == nil ||
		exec.Command("kubectl", "-n", "ais", "get", "aistore").Run() == nil {
		snapshot.ClusterReady = true
	}
	if readyPodCount("aistore", "app.kubernetes.io/component=proxy") > 0 || readyPodCount("ais", "app.kubernetes.io/component=proxy") > 0 {
		snapshot.ProxyPodsReady = true
	}
	if readyPodCount("aistore", "app.kubernetes.io/component=target") > 0 || readyPodCount("ais", "app.kubernetes.io/component=target") > 0 {
		snapshot.TargetPodsReady = true
	}
	if exec.Command("kubectl", "-n", "aistore", "get", "configmap", "aistore-target-storage-proven").Run() == nil ||
		exec.Command("kubectl", "-n", "ais", "get", "configmap", "aistore-target-storage-proven").Run() == nil {
		snapshot.TargetPVCsBound = true
	}
	if exec.Command("kubectl", "-n", "aistore", "get", "configmap", "aistore-bucket-smoke-test-passed").Run() == nil ||
		exec.Command("kubectl", "-n", "ais", "get", "configmap", "aistore-bucket-smoke-test-passed").Run() == nil {
		snapshot.BucketSmokeTestPassed = true
	}
	if exec.Command("kubectl", "-n", "aistore", "get", "configmap", "aistore-gpu-artifact-read-passed").Run() == nil ||
		exec.Command("kubectl", "-n", "ais", "get", "configmap", "aistore-gpu-artifact-read-passed").Run() == nil {
		snapshot.GPUArtifactReadPassed = true
	}
	if exec.Command("kubectl", "-n", "aistore", "get", "configmap", "aistore-metrics-proven").Run() == nil ||
		exec.Command("kubectl", "-n", "ais", "get", "configmap", "aistore-metrics-proven").Run() == nil {
		snapshot.MetricsAvailable = true
	}
	return aiplatform.EvaluateAIStoreReadiness(snapshot)
}

func renderAIStoreReadinessStatus(status aiplatform.ReadinessStatus) string {
	var b strings.Builder
	if status.Ready {
		b.WriteString("\nNVIDIA AIStore data-plane readiness: READY\n")
	} else {
		b.WriteString("\nNVIDIA AIStore data-plane readiness: NOT READY\n")
	}
	b.WriteString("  scope: optional AI dataset/cache/object path; not a generic PVC replacement\n")
	for _, check := range status.Checks {
		state := "FAIL"
		if check.Ready {
			state = "OK"
		}
		b.WriteString(fmt.Sprintf("  %s: %s — %s\n", check.Name, state, check.Message))
	}
	b.WriteString("  policy: report separately from core GPU/NIM/KAI readiness; fail closed for AIStore claims\n")
	return b.String()
}

func renderAIReadinessStatus(status aiplatform.ReadinessStatus) string {
	var b strings.Builder
	if status.Ready {
		b.WriteString("\nNVIDIA AI platform readiness: READY\n")
	} else {
		b.WriteString("\nNVIDIA AI platform readiness: NOT READY\n")
	}
	for _, check := range status.Checks {
		state := "FAIL"
		if check.Ready {
			state = "OK"
		}
		b.WriteString(fmt.Sprintf("  %s: %s — %s\n", check.Name, state, check.Message))
	}
	b.WriteString("  policy: fail closed; no NVIDIA approval/certification claim without explicit evidence\n")
	return b.String()
}
