package provision

import "testing"

func TestMockProviderRecordsCalls(t *testing.T) {
	m := &MockProvider{}
	if err := m.Metal("sandbox"); err != nil {
		t.Fatal(err)
	}
	if err := m.Bootstrap("sandbox"); err != nil {
		t.Fatal(err)
	}
	if err := m.Security("sandbox"); err != nil {
		t.Fatal(err)
	}
	if err := m.External("sandbox"); err != nil {
		t.Fatal(err)
	}
	if err := m.Wait("sandbox"); err != nil {
		t.Fatal(err)
	}
	if err := m.PostInstall("sandbox"); err != nil {
		t.Fatal(err)
	}
	if len(m.Calls) != 6 {
		t.Errorf("expected 6 calls, got %d", len(m.Calls))
	}
	if m.Calls[0] != "metal:sandbox" {
		t.Errorf("expected metal:sandbox, got %s", m.Calls[0])
	}
}

func TestExecutePhaseDispatchesCorrectly(t *testing.T) {
	m := &MockProvider{}
	ExecutePhase(m, "metal", "prod")
	if m.Calls[0] != "metal:prod" {
		t.Errorf("expected metal:prod, got %s", m.Calls[0])
	}
}

func TestExecutePhaseUnknown(t *testing.T) {
	m := &MockProvider{}
	ExecutePhase(m, "unknown", "test")
	if len(m.Calls) != 0 {
		t.Errorf("expected no calls for unknown phase, got %d", len(m.Calls))
	}
}

func TestRealProviderReturnsNoError(t *testing.T) {
	p := &RealProvider{}
	if err := p.Metal("test"); err != nil {
		t.Errorf("RealProvider.Metal returned error: %v", err)
	}
	if err := p.Bootstrap("test"); err != nil {
		t.Errorf("RealProvider.Bootstrap returned error: %v", err)
	}
	if err := p.External("test"); err != nil {
		t.Errorf("RealProvider.External returned error: %v", err)
	}
}