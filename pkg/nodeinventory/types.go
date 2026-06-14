package nodeinventory

import (
	"errors"
	"fmt"
	"strings"
)

const (
	OSFamilyRocky  = "rocky"
	OSFamilyRHEL   = "rhel"
	OSFamilyUbuntu = "ubuntu"
	OSFamilyCustom = "custom"
)

// OSImage is Ubiquity's declarative source for an operating-system image that
// can be translated to a NICo Operating System object or bootstrap fallback data.
type OSImage struct {
	Name         string            `yaml:"name" json:"name"`
	Family       string            `yaml:"family" json:"family"` // rocky, rhel, ubuntu, custom
	Version      string            `yaml:"version" json:"version"`
	Architecture string            `yaml:"architecture" json:"architecture"`
	IPXEScript   string            `yaml:"ipxeScript,omitempty" json:"ipxeScript,omitempty"`
	UserData     string            `yaml:"userData,omitempty" json:"userData,omitempty"`
	ImageURL     string            `yaml:"imageURL,omitempty" json:"imageURL,omitempty"`
	Checksum     string            `yaml:"checksum,omitempty" json:"checksum,omitempty"`
	Provenance   string            `yaml:"provenance,omitempty" json:"provenance,omitempty"`
	Dev          bool              `yaml:"dev,omitempty" json:"dev,omitempty"`
	Labels       map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// BareMetalNode is Ubiquity's desired-node record. NICo owns day-2 lifecycle;
// Ansible rendering from this type is bootstrap-only fallback output.
type BareMetalNode struct {
	Name            string            `yaml:"name" json:"name"`
	Site            string            `yaml:"site" json:"site"`
	MachineSelector map[string]string `yaml:"machineSelector,omitempty" json:"machineSelector,omitempty"`
	Role            string            `yaml:"role,omitempty" json:"role,omitempty"`
	OSImageRef      string            `yaml:"osImageRef" json:"osImageRef"`
	InstanceTypeRef string            `yaml:"instanceTypeRef,omitempty" json:"instanceTypeRef,omitempty"`
	GPUProfile      string            `yaml:"gpuProfile,omitempty" json:"gpuProfile,omitempty"`
	JoinProfile     string            `yaml:"joinProfile,omitempty" json:"joinProfile,omitempty"`
	Labels          map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// NodeInventory is the single declarative source that feeds NICo Operating
// Systems, NICo instance requests, and bootstrap-only Ansible fallback output.
type NodeInventory struct {
	APIVersion string          `yaml:"apiVersion,omitempty" json:"apiVersion,omitempty"`
	Kind       string          `yaml:"kind,omitempty" json:"kind,omitempty"`
	Images     []OSImage       `yaml:"images" json:"images"`
	Nodes      []BareMetalNode `yaml:"nodes" json:"nodes"`
}

func (i NodeInventory) Validate() error {
	var problems []string
	seenImages := map[string]bool{}
	for idx, image := range i.Images {
		prefix := fmt.Sprintf("images[%d]", idx)
		if image.Name == "" {
			problems = append(problems, prefix+": name is required")
		} else if seenImages[image.Name] {
			problems = append(problems, fmt.Sprintf("image %q: duplicate name", image.Name))
		} else {
			seenImages[image.Name] = true
		}
		if !supportedFamily(image.Family) {
			problems = append(problems, fmt.Sprintf("image %q: unsupported OS family %q", image.Name, image.Family))
		}
		if image.Version == "" {
			problems = append(problems, fmt.Sprintf("image %q: version is required", image.Name))
		}
		if image.Architecture == "" {
			problems = append(problems, fmt.Sprintf("image %q: architecture is required", image.Name))
		}
		if !image.Dev {
			missing := []string{}
			if image.Checksum == "" {
				missing = append(missing, "checksum")
			}
			if image.Provenance == "" {
				missing = append(missing, "provenance")
			}
			if len(missing) > 0 {
				problems = append(problems, fmt.Sprintf("image %q: non-dev images require %s", image.Name, strings.Join(missing, " and ")))
			}
		}
	}
	for idx, node := range i.Nodes {
		prefix := fmt.Sprintf("nodes[%d]", idx)
		if node.Name == "" {
			problems = append(problems, prefix+": name is required")
		}
		if node.Site == "" {
			problems = append(problems, fmt.Sprintf("node %q: site is required", node.Name))
		}
		if node.OSImageRef == "" {
			problems = append(problems, fmt.Sprintf("node %q: osImageRef is required", node.Name))
		} else if !seenImages[node.OSImageRef] {
			problems = append(problems, fmt.Sprintf("node %q: osImageRef %q does not match any image", node.Name, node.OSImageRef))
		}
	}
	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}

func supportedFamily(family string) bool {
	switch family {
	case OSFamilyRocky, OSFamilyRHEL, OSFamilyUbuntu, OSFamilyCustom:
		return true
	default:
		return false
	}
}
