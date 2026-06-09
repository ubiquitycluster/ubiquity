package cloud

// RequiredCloudSmokeTests returns named smoke-test markers that must pass before cloud readiness can be claimed.
func RequiredCloudSmokeTests() []string {
	markers := append([]string{}, AllManagedServiceSmokeTests()...)
	markers = append(markers,
		"restore-drill-controller-succeeded",
		"restore-drill-readable",
		"cloud-restore-drill-smoke-passed",
		"tenant-cluster-kubeconfig-present",
		"tenant-cluster-api-reachable",
		"tenant-cluster-nodes-ready",
	)
	return markers
}
