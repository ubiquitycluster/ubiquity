package nodestatus

import (
	"errors"
	"testing"
)

func TestSafetyRequiresExactConfirmation(t *testing.T) {
	_, err := EvaluateSafety(SafetyRequest{Operation: OperationRemove, NodeName: "cn01", Confirm: "wrong", DrainConfirmed: true}, SafetyClusterState{Node: NodeStatus{Name: "cn01"}})
	if !errors.Is(err, ErrSafetyGateDenied) {
		t.Fatalf("err = %v, want ErrSafetyGateDenied", err)
	}
}

func TestSafetyBlocksControlPlaneQuorumLoss(t *testing.T) {
	_, err := EvaluateSafety(SafetyRequest{Operation: OperationPowerOff, NodeName: "cp01", Confirm: "cp01", DrainConfirmed: true}, SafetyClusterState{Node: NodeStatus{Name: "cp01", Roles: []string{"control-plane"}}, ControlPlaneReady: 2, ControlPlaneTotal: 3})
	if !errors.Is(err, ErrSafetyGateDenied) {
		t.Fatalf("err = %v, want quorum denial", err)
	}
}

func TestSafetyRequiresDrainForReadyNodesUnlessForcedWithReason(t *testing.T) {
	_, err := EvaluateSafety(SafetyRequest{Operation: OperationPowerOff, NodeName: "cn01", Confirm: "cn01"}, SafetyClusterState{Node: NodeStatus{Name: "cn01", KubernetesReady: true}})
	if !errors.Is(err, ErrSafetyGateDenied) {
		t.Fatalf("err = %v, want drain denial", err)
	}
	_, err = EvaluateSafety(SafetyRequest{Operation: OperationPowerOff, NodeName: "cn01", Confirm: "cn01", Force: true}, SafetyClusterState{Node: NodeStatus{Name: "cn01", KubernetesReady: true}})
	if !errors.Is(err, ErrSafetyGateDenied) {
		t.Fatalf("err = %v, want force reason denial", err)
	}
	decision, err := EvaluateSafety(SafetyRequest{Operation: OperationPowerOff, NodeName: "cn01", Confirm: "cn01", Force: true, ForceReason: "emergency hardware isolation"}, SafetyClusterState{Node: NodeStatus{Name: "cn01", KubernetesReady: true}})
	if err != nil || !decision.Allowed {
		t.Fatalf("decision=%+v err=%v, want allowed forced operation with reason", decision, err)
	}
}

func TestSafetyRequiresStorageAndAIStoreAcknowledgement(t *testing.T) {
	_, err := EvaluateSafety(SafetyRequest{Operation: OperationRemove, NodeName: "cn01", Confirm: "cn01", DrainConfirmed: true}, SafetyClusterState{Node: NodeStatus{Name: "cn01"}, HasLocalPV: true})
	if !errors.Is(err, ErrSafetyGateDenied) {
		t.Fatalf("err = %v, want storage denial", err)
	}
	decision, err := EvaluateSafety(SafetyRequest{Operation: OperationRemove, NodeName: "cn01", Confirm: "cn01", DrainConfirmed: true, StorageAcknowledged: true, AIStoreAcknowledged: true}, SafetyClusterState{Node: NodeStatus{Name: "cn01"}, HasLocalPV: true, HasAIStoreTargetData: true})
	if err != nil || !decision.Allowed || len(decision.Warnings) != 2 {
		t.Fatalf("decision=%+v err=%v, want allowed with storage/aistore warnings", decision, err)
	}
}
