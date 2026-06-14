package nvidia

import (
	"errors"
	"testing"
)

func TestParseLaunchKitRDMAInventory(t *testing.T) {
	got, err := ParseLaunchKitRDMAInventory([]byte(`{"nodes":[{"name":"cn01","interfaces":["mlx5_0"],"rdmaResources":2}]}`))
	if err != nil {
		t.Fatalf("ParseLaunchKitRDMAInventory returned error: %v", err)
	}
	if got["cn01"].Provenance != RDMAProvenanceLaunchKit || got["cn01"].Resources != 2 || len(got["cn01"].Interfaces) != 1 {
		t.Fatalf("unexpected discovery: %+v", got["cn01"])
	}
}

func TestSelectRDMAProvenancePrefersNICoThenLaunchKitThenLocalKubectl(t *testing.T) {
	launch := map[string]RDMADiscovery{"cn01": {NodeName: "cn01", Resources: 2, Provenance: RDMAProvenanceLaunchKit}}
	local := map[string]int{"cn01": 1}
	if got := SelectRDMAProvenance("cn01", 4, launch, local); got.Provenance != RDMAProvenanceNICo || got.Resources != 4 {
		t.Fatalf("got %+v, want nico", got)
	}
	if got := SelectRDMAProvenance("cn01", 0, launch, local); got.Provenance != RDMAProvenanceLaunchKit || got.Resources != 2 {
		t.Fatalf("got %+v, want launch-kit", got)
	}
	if got := SelectRDMAProvenance("cn02", 0, nil, map[string]int{"cn02": 1}); got.Provenance != RDMAProvenanceLocalKubectl || got.Resources != 1 {
		t.Fatalf("got %+v, want local-kubectl", got)
	}
}

func TestValidateRDMAExpectedFailsClosed(t *testing.T) {
	var missing *MissingRDMAError
	if err := ValidateRDMAExpected("cn01", true, RDMADiscovery{}); !errors.As(err, &missing) {
		t.Fatalf("err=%v, want MissingRDMAError", err)
	}
	if err := ValidateRDMAExpected("cn01", true, RDMADiscovery{Resources: 1}); err != nil {
		t.Fatalf("err=%v, want nil", err)
	}
}
