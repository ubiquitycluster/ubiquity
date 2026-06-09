package nodestatus

import (
	"errors"
	"strings"
)

type Operation string

const (
	OperationRemove      Operation = "remove"
	OperationReinstall   Operation = "reinstall"
	OperationPowerOff    Operation = "power-off"
	OperationReset       Operation = "reset"
	OperationReimage     Operation = "reimage"
	OperationDrain       Operation = "drain"
	OperationEvict       Operation = "evict"
	OperationMaintenance Operation = "maintenance"
	OperationAdd         Operation = "add"
)

type SafetyRequest struct {
	Operation           Operation
	NodeName            string
	Confirm             string
	Force               bool
	ForceReason         string
	DrainConfirmed      bool
	StorageAcknowledged bool
	AIStoreAcknowledged bool
}

type SafetyClusterState struct {
	Node                 NodeStatus
	ControlPlaneReady    int
	ControlPlaneTotal    int
	HasLocalPV           bool
	HasAIStoreTargetData bool
}

type SafetyDecision struct {
	Allowed  bool
	Warnings []string
}

var ErrSafetyGateDenied = errors.New("safety gate denied")

func EvaluateSafety(req SafetyRequest, state SafetyClusterState) (SafetyDecision, error) {
	warnings := []string{}
	if req.NodeName == "" {
		req.NodeName = state.Node.Name
	}
	if requiresConfirmation(req.Operation) && req.Confirm != req.NodeName {
		return SafetyDecision{}, errors.Join(ErrSafetyGateDenied, errors.New("--confirm must exactly match node name"))
	}
	if req.Force && strings.TrimSpace(req.ForceReason) == "" {
		return SafetyDecision{}, errors.Join(ErrSafetyGateDenied, errors.New("--force requires --reason"))
	}
	if isControlPlane(state.Node) && !quorumSafeAfterOneLoss(state.ControlPlaneReady, state.ControlPlaneTotal) {
		return SafetyDecision{}, errors.Join(ErrSafetyGateDenied, errors.New("operation would risk control-plane quorum"))
	}
	if state.Node.KubernetesReady && requiresDrain(req.Operation) && !req.DrainConfirmed && !req.Force {
		return SafetyDecision{}, errors.Join(ErrSafetyGateDenied, errors.New("Kubernetes Ready nodes must be cordoned/drained before lifecycle operation"))
	}
	if state.HasLocalPV && !req.StorageAcknowledged {
		return SafetyDecision{}, errors.Join(ErrSafetyGateDenied, errors.New("local persistent volume data detected; acknowledge storage risk"))
	}
	if state.HasAIStoreTargetData && !req.AIStoreAcknowledged {
		warnings = append(warnings, "AIStore target data detected; require explicit acknowledgement before destructive lifecycle operations")
		return SafetyDecision{Warnings: warnings}, errors.Join(ErrSafetyGateDenied, errors.New("AIStore target data acknowledgement required"))
	}
	if state.HasLocalPV {
		warnings = append(warnings, "local persistent volume data acknowledged")
	}
	if state.HasAIStoreTargetData {
		warnings = append(warnings, "AIStore target data acknowledged")
	}
	return SafetyDecision{Allowed: true, Warnings: warnings}, nil
}

func requiresConfirmation(op Operation) bool {
	switch op {
	case OperationRemove, OperationReinstall, OperationReset, OperationPowerOff, OperationReimage, OperationDrain, OperationEvict, OperationMaintenance:
		return true
	default:
		return false
	}
}

func requiresDrain(op Operation) bool {
	switch op {
	case OperationRemove, OperationReinstall, OperationPowerOff, OperationReset, OperationReimage, OperationDrain, OperationEvict, OperationMaintenance:
		return true
	default:
		return false
	}
}

func isControlPlane(n NodeStatus) bool {
	for _, role := range n.Roles {
		r := strings.ToLower(role)
		if r == "control-plane" || r == "master" || strings.Contains(r, "control-plane") {
			return true
		}
	}
	return false
}

func quorumSafeAfterOneLoss(ready, total int) bool {
	if total <= 0 {
		return false
	}
	remaining := ready - 1
	needed := total/2 + 1
	return remaining >= needed
}
