package nodestatus

import "testing"

func TestParseKubernetesNodeEvidenceExtractsReadyRolesCordonAndNVIDIAResources(t *testing.T) {
	input := []byte(`{"items":[
		{"metadata":{"name":"cp01","labels":{"node-role.kubernetes.io/control-plane":""}},"spec":{"unschedulable":true},"status":{"allocatable":{"nvidia.com/gpu":"8","nvidia.com/mig-1g.10gb":"7","nvidia.com/rdma":"4"},"conditions":[{"type":"Ready","status":"True"}]}},
		{"metadata":{"name":"cn01","labels":{}},"status":{"allocatable":{"nvidia.com/gpu":"0"},"conditions":[{"type":"Ready","status":"False"}]}}
	]}`)
	got, err := ParseKubernetesNodeEvidence(input)
	if err != nil {
		t.Fatalf("ParseKubernetesNodeEvidence returned error: %v", err)
	}
	cp := got["cp01"]
	if !cp.Ready || !cp.Cordoned || cp.GPUs != 8 || cp.MIGProfiles["nvidia.com/mig-1g.10gb"] != 7 || cp.RDMAResources != 4 || !cp.NVIDIAReady {
		t.Fatalf("cp01 evidence missing fields: %+v", cp)
	}
	if len(cp.Roles) != 1 || cp.Roles[0] != "control-plane" {
		t.Fatalf("expected control-plane role, got %#v", cp.Roles)
	}
	cn := got["cn01"]
	if cn.Ready || cn.NVIDIAReady || cn.GPUs != 0 || len(cn.Roles) != 1 || cn.Roles[0] != "worker" {
		t.Fatalf("expected fail-closed worker evidence, got %+v", cn)
	}
}
