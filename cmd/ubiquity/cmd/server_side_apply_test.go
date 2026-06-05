package cmd

import "testing"

func TestKAISchedulerUsesServerSideApplyForLargeCRDs(t *testing.T) {
	if !shouldServerSideApply("/repo/platform/kai-scheduler") {
		t.Fatal("KAI Scheduler sandbox apply should use server-side apply to avoid oversized CRD last-applied annotations")
	}
	if shouldServerSideApply("/repo/platform/nim-operator") {
		t.Fatal("NIM Operator should keep default apply behavior")
	}
}
