package cloud

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var fiveFieldCron = regexp.MustCompile(`^\S+\s+\S+\s+\S+\s+\S+\s+\S+$`)
var dayRetention = regexp.MustCompile(`^([1-9][0-9]*)d$`)

// BackupPolicyRequest describes backup/snapshot/resource-preset operations for a tenant.
type BackupPolicyRequest struct {
	Name                 string
	Namespace            string
	Schedule             string
	Retention            string
	RepositorySecretName string
	SnapshotClass        string
	PresetName           string
	PresetCPU            string
	PresetMemory         string
	PresetGPU            string
}

// RenderBackupPolicy renders K8up, VolumeSnapshotClass, and resource preset primitives.
func RenderBackupPolicy(req BackupPolicyRequest) (string, error) {
	req = defaultBackupPolicy(req)
	if err := validateBackupPolicy(req); err != nil {
		return "", err
	}
	keepDays := dayRetention.FindStringSubmatch(req.Retention)[1]
	var b strings.Builder
	fmt.Fprintf(&b, `apiVersion: k8up.io/v1
kind: Schedule
metadata:
  name: %s
  namespace: %s
  labels:
    ubiquity.ai/backup-policy: tenant
spec:
  backend:
    repositorySecretRef:
      name: %s
      key: password
  backup:
    schedule: %q
  check:
    schedule: %q
  prune:
    schedule: %q
    retention:
      keepDaily: %s
---
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotClass
metadata:
  name: %s
  labels:
    ubiquity.ai/snapshot-policy: retained
  annotations:
    ubiquity.ai/immutability-note: "Retain snapshots and use immutable object storage where configured; rendering is not restore proof."
driver: driver.longhorn.io
deletionPolicy: Retain
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: %s-resource-preset
  namespace: %s
  labels:
    ubiquity.ai/resource-preset: %s
data:
  cpu: %q
  memory: %s
`, req.Name, req.Namespace, req.RepositorySecretName, req.Schedule, req.Schedule, req.Schedule, keepDays, req.SnapshotClass, req.PresetName, req.Namespace, req.PresetName, req.PresetCPU, req.PresetMemory)
	if req.PresetGPU != "" {
		fmt.Fprintf(&b, "  nvidia.com/gpu: %q\n", req.PresetGPU)
	}
	return b.String(), nil
}

func defaultBackupPolicy(req BackupPolicyRequest) BackupPolicyRequest {
	if req.Namespace == "" {
		req.Namespace = "tenant-a"
	}
	if req.Schedule == "" {
		req.Schedule = "0 2 * * *"
	}
	if req.Retention == "" {
		req.Retention = "30d"
	}
	if req.RepositorySecretName == "" && req.Name != "" {
		req.RepositorySecretName = req.Name + "-repo"
	}
	if req.SnapshotClass == "" {
		req.SnapshotClass = "longhorn-snapshots"
	}
	if req.PresetName == "" {
		req.PresetName = "gpu-medium"
	}
	if req.PresetCPU == "" {
		req.PresetCPU = "16"
	}
	if req.PresetMemory == "" {
		req.PresetMemory = "128Gi"
	}
	return req
}

func validateBackupPolicy(req BackupPolicyRequest) error {
	if !kubeName.MatchString(req.Name) {
		return fmt.Errorf("backup policy name %q must be DNS-compatible", req.Name)
	}
	if !kubeName.MatchString(req.Namespace) || !kubeName.MatchString(req.RepositorySecretName) || !kubeName.MatchString(req.SnapshotClass) || !kubeName.MatchString(req.PresetName) {
		return fmt.Errorf("backup policy namespace, repository secret, snapshot class, and preset name must be DNS-compatible")
	}
	if !fiveFieldCron.MatchString(req.Schedule) {
		return fmt.Errorf("backup policy schedule %q must be a five-field cron expression", req.Schedule)
	}
	match := dayRetention.FindStringSubmatch(req.Retention)
	if match == nil {
		return fmt.Errorf("backup policy retention %q must be expressed as Nd, for example 30d", req.Retention)
	}
	days, _ := strconv.Atoi(match[1])
	if days < 1 || days > 3650 {
		return fmt.Errorf("backup policy retention must be between 1d and 3650d")
	}
	return nil
}
