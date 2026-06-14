package aiplatform

import "testing"

func TestStorageAlternativesPreferAIStoreForAIDataPlaneButNotGenericPVCs(t *testing.T) {
	alternatives := StorageAlternatives()
	if len(alternatives) < 3 {
		t.Fatalf("expected storage alternatives to compare AIStore, Longhorn, and local PV paths, got %d", len(alternatives))
	}

	byName := make(map[string]StorageAlternative, len(alternatives))
	for _, alternative := range alternatives {
		byName[alternative.Name] = alternative
		if alternative.Source == "" {
			t.Fatalf("storage alternative %q must have source provenance", alternative.Name)
		}
		if alternative.Decision == "" || alternative.Scope == "" || alternative.Rationale == "" {
			t.Fatalf("storage alternative %q must have decision, scope, and rationale", alternative.Name)
		}
	}

	ais := byName["nvidia-aistore"]
	if ais.Decision != StorageDecisionAdoptForAIDataPlane {
		t.Fatalf("AIStore should be the preferred AI data-plane option, got %q", ais.Decision)
	}
	if !ais.ReplacesLonghornForAIData {
		t.Fatal("AIStore should replace Longhorn for high-throughput AI dataset/cache object paths")
	}
	if ais.ReplacesGenericPVCs {
		t.Fatal("AIStore must not claim to replace generic Kubernetes PVC/shared filesystem use cases")
	}

	longhorn := byName["longhorn"]
	if longhorn.Decision != StorageDecisionRetainForGenericPVCs {
		t.Fatalf("Longhorn should be retained only for generic PVCs until a better POSIX/RWX option is selected, got %q", longhorn.Decision)
	}
	if longhorn.ReplacesLonghornForAIData {
		t.Fatal("Longhorn cannot replace itself for AI data-plane paths")
	}
}

func TestProductionProfileIncludesAIStoreAsEvaluatedStorageCandidate(t *testing.T) {
	profile, err := GetProfile("ai-production")
	if err != nil {
		t.Fatalf("GetProfile(ai-production): %v", err)
	}
	component := profile.ComponentsByName()["aistore"]
	if component.SourceRepo != "https://github.com/NVIDIA/aistore" {
		t.Fatalf("AIStore must be sourced from NVIDIA/aistore, got %q", component.SourceRepo)
	}
	if component.ProductionDefault {
		t.Fatal("AIStore should not become unconditional production default before persistence/capacity/readiness proof")
	}
	if !component.Optional {
		t.Fatal("AIStore should remain optional/evaluated until storage fit checks pass")
	}
}
