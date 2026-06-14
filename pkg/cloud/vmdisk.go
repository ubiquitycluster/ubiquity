package cloud

import (
	"fmt"
	"regexp"
	"strings"
)

var kubeName = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

// VMDiskSourceType identifies the source used to populate a VM disk.
type VMDiskSourceType string

const (
	VMDiskSourceBlank VMDiskSourceType = "blank"
	VMDiskSourceHTTP  VMDiskSourceType = "http"
	VMDiskSourcePVC   VMDiskSourceType = "pvc"
)

// VMDiskSource describes a blank, imported, or cloned KubeVirt/CDI disk.
type VMDiskSource struct {
	Type         VMDiskSourceType
	URL          string
	PVCName      string
	PVCNamespace string
}

// VMDiskRequest is the Ubiquity-native standalone VM disk contract.
type VMDiskRequest struct {
	Name         string
	Namespace    string
	Size         string
	StorageClass string
	AccessMode   string
	Source       VMDiskSource
}

// RenderVMDisk renders a reusable VM disk without requiring VM creation.
func RenderVMDisk(req VMDiskRequest) (string, error) {
	req = defaultVMDisk(req)
	if err := validateVMDisk(req); err != nil {
		return "", err
	}
	if req.Source.Type == VMDiskSourceBlank {
		return renderBlankPVC(req), nil
	}
	return renderDiskDataVolume(req), nil
}

func defaultVMDisk(req VMDiskRequest) VMDiskRequest {
	if req.Namespace == "" {
		req.Namespace = "virtual-machines"
	}
	if req.Size == "" {
		req.Size = "40Gi"
	}
	if req.AccessMode == "" {
		req.AccessMode = "ReadWriteOnce"
	}
	if req.Source.Type == "" {
		req.Source.Type = VMDiskSourceBlank
	}
	if req.Source.PVCNamespace == "" {
		req.Source.PVCNamespace = req.Namespace
	}
	return req
}

func validateVMDisk(req VMDiskRequest) error {
	if !kubeName.MatchString(req.Name) {
		return fmt.Errorf("VM disk name %q must be DNS-compatible", req.Name)
	}
	if !kubeName.MatchString(req.Namespace) {
		return fmt.Errorf("VM disk namespace %q must be DNS-compatible", req.Namespace)
	}
	switch req.Source.Type {
	case VMDiskSourceBlank:
	case VMDiskSourceHTTP:
		if strings.TrimSpace(req.Source.URL) == "" {
			return fmt.Errorf("http source URL is required for VM disk %q", req.Name)
		}
	case VMDiskSourcePVC:
		if !kubeName.MatchString(req.Source.PVCName) {
			return fmt.Errorf("source PVC name is required for VM disk %q", req.Name)
		}
		if !kubeName.MatchString(req.Source.PVCNamespace) {
			return fmt.Errorf("source PVC namespace %q must be DNS-compatible", req.Source.PVCNamespace)
		}
	default:
		return fmt.Errorf("unsupported VM disk source %q", req.Source.Type)
	}
	return nil
}

func renderBlankPVC(req VMDiskRequest) string {
	var b strings.Builder
	fmt.Fprintf(&b, `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: %s
  namespace: %s
  labels:
    app.kubernetes.io/part-of: ubiquity-virtual-machines
    ubiquity.ai/disk-source: blank
spec:
  accessModes:
    - %s
  resources:
    requests:
      storage: %s
`, req.Name, req.Namespace, req.AccessMode, req.Size)
	if req.StorageClass != "" {
		fmt.Fprintf(&b, "  storageClassName: %s\n", req.StorageClass)
	}
	return b.String()
}

func renderDiskDataVolume(req VMDiskRequest) string {
	var b strings.Builder
	fmt.Fprintf(&b, `apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: %s
  namespace: %s
  labels:
    app.kubernetes.io/part-of: ubiquity-virtual-machines
    ubiquity.ai/disk-source: %s
spec:
  source:
`, req.Name, req.Namespace, req.Source.Type)
	if req.Source.Type == VMDiskSourceHTTP {
		fmt.Fprintf(&b, "    http:\n      url: %s\n", req.Source.URL)
	} else {
		fmt.Fprintf(&b, "    pvc:\n      name: %s\n      namespace: %s\n", req.Source.PVCName, req.Source.PVCNamespace)
	}
	fmt.Fprintf(&b, `  pvc:
    accessModes:
      - %s
    resources:
      requests:
        storage: %s
`, req.AccessMode, req.Size)
	if req.StorageClass != "" {
		fmt.Fprintf(&b, "    storageClassName: %s\n", req.StorageClass)
	}
	return b.String()
}
