package aiplatform

// AIStoreSnapshot contains optional AI data-plane readiness evidence. AIStore is
// intentionally evaluated separately from core GPU/NIM/KAI readiness because it
// is preferred for AI object/data paths but is not required for every workload.
type AIStoreSnapshot struct {
	OperatorReady         bool
	CRDsEstablished       bool
	ClusterReady          bool
	ProxyPodsReady        bool
	TargetPodsReady       bool
	TargetPVCsBound       bool
	BucketSmokeTestPassed bool
	GPUArtifactReadPassed bool
	MetricsAvailable      bool
}

// EvaluateAIStoreReadiness fails closed for the optional NVIDIA AIStore data
// plane. Missing evidence means AIStore is not ready, but that should be
// reported as data-plane readiness rather than generic Kubernetes PVC readiness.
func EvaluateAIStoreReadiness(snapshot AIStoreSnapshot) ReadinessStatus {
	checks := []CheckResult{
		boolCheck("aistore-operator", snapshot.OperatorReady, "AIS operator is available", "AIS operator is missing or not ready"),
		boolCheck("aistore-crds", snapshot.CRDsEstablished, "AIS CRDs are established", "AIS CRDs are missing or not established"),
		boolCheck("aistore-cluster", snapshot.ClusterReady, "AIS cluster custom resource is ready", "AIS cluster custom resource is not proven ready"),
		boolCheck("aistore-proxies", snapshot.ProxyPodsReady, "AIS proxy pods are ready", "AIS proxy pods are missing or not ready"),
		boolCheck("aistore-targets", snapshot.TargetPodsReady, "AIS target pods are ready", "AIS target pods are missing or not ready"),
		boolCheck("aistore-target-pvcs", snapshot.TargetPVCsBound, "AIS target PVCs are bound or explicit hostPath backing is approved", "AIS target storage backing is not proven bound or approved"),
		boolCheck("aistore-bucket-smoke", snapshot.BucketSmokeTestPassed, "AIS bucket create/list/write/read/delete smoke test passed", "AIS bucket smoke test has not passed"),
		boolCheck("aistore-gpu-artifact-read", snapshot.GPUArtifactReadPassed, "GPU workload read a representative artifact through AIS/S3", "GPU workload artifact read through AIS/S3 has not passed"),
		boolCheck("aistore-metrics", snapshot.MetricsAvailable, "AIS health or metrics output is available", "AIS health or metrics output is not proven available"),
	}

	ready := true
	for _, check := range checks {
		if !check.Ready {
			ready = false
			break
		}
	}
	return ReadinessStatus{Ready: ready, Checks: checks}
}
