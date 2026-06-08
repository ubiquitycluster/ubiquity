package cloud

import (
	"fmt"
	"strings"
)

// RenderCloudProductionAuditChecklist returns the reviewer-facing checklist for production cloud readiness.
func RenderCloudProductionAuditChecklist() string {
	var b strings.Builder
	fmt.Fprintln(&b, "# Ubiquity cloud production readiness audit checklist")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Use this checklist before marking any cloud primitive, tenant service, VM, or backup policy production-ready. Render/apply proof is intent only; readiness must be backed by live evidence and must not object existence as readiness.")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Required evidence")
	for _, item := range []string{
		"Run Helm schema validation and server-side dry-run for the relevant primitive.",
		"Render and review the operator install-plan contract, including controller ownership and expected air-gap artifacts.",
		"Install prerequisite CRDs/controllers before applying service intent; record required CRDs and present CRDs.",
		"Run `ubiquity cloud collect-readiness` and then `ubiquity cloud readiness --readiness-file <file>` until the report contains `ready: true`.",
		"Prove persistent services with a restore drill; rendered Restore objects are not restore proof.",
		"Capture required smoke tests: " + strings.Join(RequiredCloudSmokeTests(), ", ") + ".",
		"For KubeVirt, review the KubeVirt image catalog, standalone disk attachments, CDI import, guest boot, and guest health separately.",
		"For air-gap environments, verify every operator/chart/image artifact referenced by the install-plan is mirrored and checksummed.",
	} {
		fmt.Fprintf(&b, "- [ ] %s\n", item)
	}
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Commands")
	fmt.Fprintln(&b, "```sh")
	fmt.Fprintln(&b, "ubiquity cloud render prerequisites")
	fmt.Fprintln(&b, "ubiquity cloud render operator-bundles")
	fmt.Fprintln(&b, "test/e2e/cloud-primitives-server-dry-run.sh")
	fmt.Fprintln(&b, "ubiquity cloud collect-readiness > /tmp/cloud-readiness-evidence.json")
	fmt.Fprintln(&b, "ubiquity cloud readiness --readiness-file /tmp/cloud-readiness-evidence.json")
	fmt.Fprintln(&b, "ubiquity cloud render restore-drill")
	fmt.Fprintln(&b, "ubiquity virtual-machines image-catalog")
	fmt.Fprintln(&b, "```")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## managed service readiness resources")
	fmt.Fprintln(&b, "These resource APIs are collected by default and evaluated via controller status conditions:")
	for _, resource := range AllManagedServiceReadinessResources() {
		fmt.Fprintf(&b, "- %s\n", resource)
	}
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Fail-closed boundaries")
	for _, boundary := range []string{
		"A successful render is not readiness.",
		"A successful apply is not readiness.",
		"A Kubernetes object existing is not object existence readiness.",
		"A KubeVirt image catalog entry does not prove CDI import or guest boot.",
		"A backup Schedule or Restore object does not prove recoverability without a completed restore drill and smoke test.",
	} {
		fmt.Fprintf(&b, "- %s\n", boundary)
	}
	return b.String()
}
