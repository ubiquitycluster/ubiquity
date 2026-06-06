package cmd

import "testing"

func TestParseDiskAttachmentsAcceptsNamePVCList(t *testing.T) {
	disks, err := parseDiskAttachments([]string{"datasets:datasets-pvc", "checkpoints:checkpoints-pvc"})
	if err != nil {
		t.Fatalf("parseDiskAttachments returned error: %v", err)
	}
	if len(disks) != 2 || disks[0].Name != "datasets" || disks[0].PVCName != "datasets-pvc" || disks[1].Name != "checkpoints" {
		t.Fatalf("unexpected parsed disks: %#v", disks)
	}
}

func TestParseDiskAttachmentsRejectsMalformedInput(t *testing.T) {
	_, err := parseDiskAttachments([]string{"datasets"})
	if err == nil {
		t.Fatalf("expected malformed attach disk error")
	}
}
