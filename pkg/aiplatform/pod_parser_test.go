package aiplatform

import "testing"

func TestParseReadyPodCountFailsClosedForEmptyList(t *testing.T) {
	count, err := ParseReadyPodCount([]byte(`{"items":[]}`))
	if err != nil {
		t.Fatalf("ParseReadyPodCount returned unexpected error: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no ready pod evidence for empty list, got %d", count)
	}
}

func TestParseReadyPodCountRequiresReadyConditionAndRunningPhase(t *testing.T) {
	input := []byte(`{
		"items": [
			{"metadata":{"name":"pending"},"status":{"phase":"Pending","conditions":[{"type":"Ready","status":"True"}]}},
			{"metadata":{"name":"not-ready"},"status":{"phase":"Running","conditions":[{"type":"Ready","status":"False"}]}},
			{"metadata":{"name":"ready"},"status":{"phase":"Running","conditions":[{"type":"Ready","status":"True"}]}}
		]
	}`)
	count, err := ParseReadyPodCount(input)
	if err != nil {
		t.Fatalf("ParseReadyPodCount returned unexpected error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly one ready running pod, got %d", count)
	}
}
