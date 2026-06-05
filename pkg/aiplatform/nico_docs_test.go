package aiplatform

import (
	"strings"
	"testing"
)

func TestNvidiaInfraControllerArchitectureDocs(t *testing.T) {
	doc := mustRead(t, "../../docs/architecture/on-prem/nvidia-infra-controller-node-lifecycle.md")
	for _, required := range []string{
		"NVIDIA Infra Controller",
		"NVIDIA/infra-controller",
		"NICo Core",
		"NICo REST",
		"site-agent",
		"Operating System",
		"Machine",
		"Instance",
		"Task",
		"Machine GPU stats",
		"bootstrap boundary",
		"experimental/preview",
	} {
		if !strings.Contains(doc, required) {
			t.Fatalf("NVIDIA Infra Controller architecture doc must include %q", required)
		}
	}
}

func TestNvidiaInfraControllerAdminReferenceAndRunbookDocs(t *testing.T) {
	for path, required := range map[string][]string{
		"../../docs/admin-guide/nvidia-infra-controller-node-management.md": {
			"experimental/preview",
			"bootstrap boundary",
			"Machine GPU stats",
			"no certification",
			"multi-OS boot",
			"Vault/External Secrets",
		},
		"../../docs/admin-guide/administration/tutorials/add-or-remove-nodes.md": {
			"NVIDIA Infra Controller",
			"fallback/migration-only",
			"experimental/preview",
		},
		"../../docs/reference/nvidia-infra-controller/os-images.md": {
			"Operating System",
			"checksum",
			"No secrets",
			"not a certification claim",
			"multi-OS boot",
			"Vault plus External Secrets Operator",
		},
		"../../docs/reference/nvidia-infra-controller/bootstrap-to-nico.md": {
			"bootstrap boundary",
			"single active lifecycle owner",
			"experimental/preview",
		},
		"../../docs/reference/nvidia-infra-controller/prerequisites.md": {
			"NVIDIA Infra Controller",
			"PostgreSQL",
			"fallback/migration-only",
			"render/report-only",
			"does not install",
			"Multi-OS boot",
		},
		"../../docs/reference/nvidia-infra-controller/node-state-machine.md": {
			"Machine",
			"Instance",
			"Task",
			"experimental/preview",
		},
		"../../docs/reference/nvidia-infra-controller/status-aggregation.md": {
			"Machine GPU stats",
			"Status output must not include",
			"not a certification",
		},
		"../../docs/admin-guide/runbooks/nvidia-infra-controller/nico-bootstrap.md": {
			"experimental/preview",
			"No secrets",
			"bootstrap-to-NICo handoff",
		},
		"../../docs/admin-guide/runbooks/nvidia-infra-controller/nico-machine-provisioning.md": {
			"experimental/preview",
			"Operating System",
			"Task ID",
		},
		"../../docs/admin-guide/runbooks/nvidia-infra-controller/nico-node-reinstall.md": {
			"experimental/preview",
			"legacy BareMetalHost",
			"No secrets",
		},
		"../../docs/admin-guide/runbooks/nvidia-infra-controller/nico-bmc-redfish.md": {
			"Redfish",
			"without exposing credentials",
			"not a hardware certification claim",
		},
		"../../docs/admin-guide/runbooks/nvidia-infra-controller/nico-gpu-validation.md": {
			"Machine GPU stats",
			"does not certify",
			"GPU Operator",
		},
		"../../docs/admin-guide/runbooks/nico-prerequisites.md": {
			"compatibility path",
			"nvidia-infra-controller/prerequisites.md",
			"fallback/migration-only",
		},
		"../../docs/admin-guide/runbooks/nico-bootstrap.md": {
			"compatibility path",
			"nvidia-infra-controller/nico-bootstrap.md",
			"fallback/migration-only",
		},
		"../../docs/admin-guide/runbooks/nico-machine-provisioning.md": {
			"compatibility path",
			"nvidia-infra-controller/nico-machine-provisioning.md",
			"fallback/migration-only",
		},
		"../../docs/admin-guide/runbooks/nico-node-reinstall.md": {
			"compatibility path",
			"nvidia-infra-controller/nico-node-reinstall.md",
			"fallback/migration-only",
		},
		"../../docs/admin-guide/runbooks/nico-bmc-redfish.md": {
			"compatibility path",
			"nvidia-infra-controller/nico-bmc-redfish.md",
			"fallback/migration-only",
		},
		"../../docs/admin-guide/runbooks/nico-gpu-validation.md": {
			"compatibility path",
			"nvidia-infra-controller/nico-gpu-validation.md",
			"fallback/migration-only",
		},
	} {
		doc := mustRead(t, path)
		for _, term := range required {
			if !strings.Contains(doc, term) {
				t.Fatalf("%s must include %q", path, term)
			}
		}
	}
}

func TestMetal3DocsDescribeMigrationOnlyFallback(t *testing.T) {
	for path, required := range map[string][]string{
		"../../docs/architecture/on-prem/openstack-bmo-node-discovery.md": {
			"fallback/migration-only",
			"NVIDIA Infra Controller",
			"bootstrap boundary",
		},
		"../../docs/reference/metal3/baremetalhost-states.md": {
			"fallback/migration-only",
			"NVIDIA Infra Controller",
		},
		"../../docs/reference/metal3/remove-host.md": {
			"fallback/migration-only",
			"NVIDIA Infra Controller",
		},
	} {
		doc := mustRead(t, path)
		for _, term := range required {
			if !strings.Contains(doc, term) {
				t.Fatalf("%s must include %q", path, term)
			}
		}
	}
}

func TestBareMetalFallbackScriptsAreCommentedAsMigrationOnly(t *testing.T) {
	for _, path := range []string{
		"../../tools/disk-image/scripts/reinstall_host.sh",
		"../../tools/disk-image/scripts/delete-host.sh",
	} {
		script := mustRead(t, path)
		for _, term := range []string{
			"fallback/migration-only",
			"NVIDIA Infra Controller",
			"do not use for new day-2 lifecycle automation",
		} {
			if !strings.Contains(script, term) {
				t.Fatalf("%s must include comment term %q", path, term)
			}
		}
	}
}
