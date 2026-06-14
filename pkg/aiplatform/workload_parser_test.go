package aiplatform

import "testing"

func TestParseAvailableDeploymentsOnlyCountsAvailableReplicas(t *testing.T) {
	input := []byte(`{
		"items": [
			{"metadata":{"name":"ready"},"status":{"readyReplicas":1,"availableReplicas":1}},
			{"metadata":{"name":"created-but-unavailable"},"status":{"readyReplicas":1,"availableReplicas":0}},
			{"metadata":{"name":"missing-status"},"status":{}}
		]
	}`)
	available, err := ParseAvailableDeployments(input)
	if err != nil {
		t.Fatalf("ParseAvailableDeployments returned error: %v", err)
	}
	if !available["ready"] {
		t.Fatal("expected ready deployment to be available")
	}
	if available["created-but-unavailable"] || available["missing-status"] {
		t.Fatalf("unavailable deployments must not be counted ready: %#v", available)
	}
}

func TestParseAvailableDeploymentsHandlesSingleDeploymentObject(t *testing.T) {
	input := []byte(`{"metadata":{"name":"gpu-operator"},"status":{"readyReplicas":1,"availableReplicas":1}}`)
	available, err := ParseAvailableDeployments(input)
	if err != nil {
		t.Fatalf("ParseAvailableDeployments returned error: %v", err)
	}
	if !available["gpu-operator"] {
		t.Fatalf("expected single deployment object to be counted available: %#v", available)
	}
}

func TestParseReadyDaemonSetsRequiresDesiredScheduledReadyAndAvailable(t *testing.T) {
	input := []byte(`{
		"items": [
			{"metadata":{"name":"ready-ds"},"status":{"desiredNumberScheduled":2,"numberReady":2,"numberAvailable":2}},
			{"metadata":{"name":"not-available"},"status":{"desiredNumberScheduled":2,"numberReady":2,"numberAvailable":1}},
			{"metadata":{"name":"no-desired"},"status":{"desiredNumberScheduled":0,"numberReady":0,"numberAvailable":0}}
		]
	}`)
	ready, err := ParseReadyDaemonSets(input)
	if err != nil {
		t.Fatalf("ParseReadyDaemonSets returned error: %v", err)
	}
	if !ready["ready-ds"] {
		t.Fatal("expected daemonset with desired=ready=available to be ready")
	}
	if ready["not-available"] || ready["no-desired"] {
		t.Fatalf("daemonsets without full ready/available coverage must not be counted ready: %#v", ready)
	}
}

func TestParseReadyDaemonSetsHandlesSingleDaemonSetObject(t *testing.T) {
	input := []byte(`{"metadata":{"name":"nvidia-device-plugin-daemonset"},"status":{"desiredNumberScheduled":2,"numberReady":2,"numberAvailable":2}}`)
	ready, err := ParseReadyDaemonSets(input)
	if err != nil {
		t.Fatalf("ParseReadyDaemonSets returned error: %v", err)
	}
	if !ready["nvidia-device-plugin-daemonset"] {
		t.Fatalf("expected single daemonset object to be counted ready: %#v", ready)
	}
}
