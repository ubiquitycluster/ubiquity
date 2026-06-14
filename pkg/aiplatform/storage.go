package aiplatform

// StorageDecision records how a storage candidate should be used in the NVIDIA
// AI workload platform. The decision is intentionally scoped because AI object
// data paths, generic Kubernetes PVCs, and POSIX/RWX filesystems are different
// requirements.
type StorageDecision string

const (
	StorageDecisionAdoptForAIDataPlane  StorageDecision = "adopt-for-ai-data-plane"
	StorageDecisionRetainForGenericPVCs StorageDecision = "retain-for-generic-pvcs"
	StorageDecisionReferenceOnly        StorageDecision = "reference-only"
)

// StorageAlternative documents a source-backed storage option and what it can
// safely replace. Replacement claims are split so AI dataset/cache improvements
// do not accidentally imply replacement of generic Kubernetes PVC semantics.
type StorageAlternative struct {
	Name                      string
	Source                    string
	Decision                  StorageDecision
	Scope                     string
	Rationale                 string
	ReplacesLonghornForAIData bool
	ReplacesGenericPVCs       bool
	ReadinessEvidence         []string
	Caveats                   []string
}

// StorageAlternatives returns the evaluated storage options for the NVIDIA AI
// platform profile. NVIDIA AIStore is preferred for AI object/dataset/cache
// paths; Longhorn remains a generic PVC fallback rather than the AI data plane.
func StorageAlternatives() []StorageAlternative {
	return []StorageAlternative{
		{
			Name:                      "nvidia-aistore",
			Source:                    "https://github.com/NVIDIA/aistore + https://github.com/NVIDIA/ais-k8s",
			Decision:                  StorageDecisionAdoptForAIDataPlane,
			Scope:                     "high-throughput AI dataset, checkpoint, model artifact, S3-compatible object, cache, sharded archive, and remote-bucket acceleration paths",
			Rationale:                 "AIStore is NVIDIA-maintained storage built for AI workloads, Kubernetes production deployment via ais-k8s, S3-compatible access, PyTorch integration, ETL/data transforms, batch object fetches, elastic scale-out, and local-disk target performance close to GPU workers.",
			ReplacesLonghornForAIData: true,
			ReplacesGenericPVCs:       false,
			ReadinessEvidence: []string{
				"AIS operator rollout available",
				"AIS cluster CR ready",
				"proxy and target pods ready",
				"bucket create/list/read/write smoke test passed",
				"GPU workload can read representative model or dataset artifact through AIS/S3 endpoint",
			},
			Caveats: []string{
				"AIStore is object/data-plane storage, not a drop-in RWX/POSIX PVC replacement",
				"ais-k8s target storage still requires local PVs, host paths, or another PV provider for target disks",
				"enable only after capacity, persistence, failure-domain, and operational-readiness checks pass",
			},
		},
		{
			Name:                "longhorn",
			Source:              "system/longhorn-system plus https://github.com/longhorn/longhorn",
			Decision:            StorageDecisionRetainForGenericPVCs,
			Scope:               "generic Kubernetes PVC use cases, small stateful platform services, and non-performance-critical RWO/RWX volumes until a stronger POSIX/RWX alternative is selected",
			Rationale:           "Longhorn is useful generic replicated block storage, but it is not NVIDIA-maintained and should not be the default high-throughput AI data/cache path when AIStore is a better fit.",
			ReplacesGenericPVCs: false,
			ReadinessEvidence: []string{
				"Longhorn manager/UI healthy for generic platform storage only",
				"application PVCs bound and recoverable",
			},
			Caveats: []string{
				"do not claim Longhorn proves NVIDIA AI data-plane readiness",
				"do not use as the preferred dataset/cache layer for large GPU training or inference artifact paths",
			},
		},
		{
			Name:      "ais-local-pv-targets",
			Source:    "https://github.com/NVIDIA/ais-k8s/blob/main/docs/storage_volumes.md",
			Decision:  StorageDecisionReferenceOnly,
			Scope:     "local disk PV or hostPath backing for AIStore target pods on bare-metal AI storage/GPU nodes",
			Rationale: "ais-k8s documents target PVCs bound to existing PVs, ReadWriteOnce access, storageClass matching, XFS recommendation, node affinity, and optional hostPath bypass for target disks.",
			ReadinessEvidence: []string{
				"target PVCs bound to node-local PVs",
				"target pods scheduled on nodes that own the disks",
				"AIS resilver/rebalance and bucket smoke tests pass",
			},
			Caveats: []string{
				"local PVs are backing media for AIS targets, not a shared filesystem",
				"hostPath mode has security implications and must be explicit",
			},
		},
	}
}
