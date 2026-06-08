package cloud

// RequiredCloudSmokeTests returns named smoke-test markers that must pass before cloud readiness can be claimed.
func RequiredCloudSmokeTests() []string {
	return []string{
		"postgres-connectivity",
		"redis-connectivity",
		"kafka-produce-consume",
		"objectbucket-read-write",
		"restore-drill-readable",
	}
}
