package nodeinventory

import "fmt"

// NICoOperatingSystem is a small, JSON/YAML-postable representation of NICo's
// Operating System API payload. It intentionally avoids generated-client
// coupling while preserving apiVersion/kind/metadata/spec shape.
type NICoOperatingSystem struct {
	APIVersion string                  `json:"apiVersion" yaml:"apiVersion"`
	Kind       string                  `json:"kind" yaml:"kind"`
	Metadata   NICoObjectMetadata      `json:"metadata" yaml:"metadata"`
	Spec       NICoOperatingSystemSpec `json:"spec" yaml:"spec"`
}

type NICoObjectMetadata struct {
	Name   string            `json:"name" yaml:"name"`
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

type NICoOperatingSystemSpec struct {
	Family       string            `json:"family" yaml:"family"`
	Version      string            `json:"version" yaml:"version"`
	Architecture string            `json:"architecture" yaml:"architecture"`
	ImageURL     string            `json:"imageURL,omitempty" yaml:"imageURL,omitempty"`
	Checksum     string            `json:"checksum" yaml:"checksum"`
	Provenance   string            `json:"provenance" yaml:"provenance"`
	IPXEScript   string            `json:"ipxeScript" yaml:"ipxeScript"`
	UserData     string            `json:"userData" yaml:"userData"`
	Labels       map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

func RenderNICoOperatingSystems(inventory NodeInventory) ([]NICoOperatingSystem, error) {
	if err := inventory.Validate(); err != nil {
		return nil, err
	}
	objects := make([]NICoOperatingSystem, 0, len(inventory.Images))
	for _, image := range inventory.Images {
		ipxe, userData, err := renderBootData(image)
		if err != nil {
			return nil, err
		}
		objects = append(objects, NICoOperatingSystem{
			APIVersion: "infra.nvidia.com/v1alpha1",
			Kind:       "OperatingSystem",
			Metadata: NICoObjectMetadata{
				Name:   image.Name,
				Labels: image.Labels,
			},
			Spec: NICoOperatingSystemSpec{
				Family:       image.Family,
				Version:      image.Version,
				Architecture: image.Architecture,
				ImageURL:     image.ImageURL,
				Checksum:     image.Checksum,
				Provenance:   image.Provenance,
				IPXEScript:   ipxe,
				UserData:     userData,
				Labels:       image.Labels,
			},
		})
	}
	return objects, nil
}

func renderBootData(image OSImage) (string, string, error) {
	switch image.Family {
	case OSFamilyRocky, OSFamilyRHEL:
		return renderKickstartIPXE(image), renderKickstartUserData(image), nil
	case OSFamilyUbuntu:
		return renderUbuntuIPXE(image), renderUbuntuAutoinstallUserData(image), nil
	case OSFamilyCustom:
		if image.IPXEScript == "" || image.UserData == "" {
			return "", "", fmt.Errorf("custom image %q requires IPXEScript and UserData", image.Name)
		}
		return image.IPXEScript, image.UserData, nil
	default:
		return "", "", fmt.Errorf("image %q: unsupported OS family %q", image.Name, image.Family)
	}
}

func renderKickstartIPXE(image OSImage) string {
	return fmt.Sprintf(`#!ipxe
echo Booting %s %s (%s)
set image-url %s
kernel ${image-url}/images/pxeboot/vmlinuz initrd=initrd.img inst.stage2=${image-url} inst.ks=${user-data-url} ip=dhcp
initrd ${image-url}/images/pxeboot/initrd.img
boot
`, image.Family, image.Version, image.Architecture, image.ImageURL)
}

func renderKickstartUserData(image OSImage) string {
	return fmt.Sprintf(`# Kickstart for %s
text
lang en_US.UTF-8
keyboard us
timezone UTC --utc
network --bootproto=dhcp --activate
rootpw --lock
reboot
%%packages
@^minimal-environment
%%end
`, image.Name)
}

func renderUbuntuIPXE(image OSImage) string {
	return fmt.Sprintf(`#!ipxe
echo Booting Ubuntu %s autoinstall (%s)
set image-url %s
kernel ${image-url}/casper/vmlinuz ip=dhcp url=${image-url} autoinstall ds=nocloud-net;s=${user-data-base-url}/
initrd ${image-url}/casper/initrd
boot
`, image.Version, image.Architecture, image.ImageURL)
}

func renderUbuntuAutoinstallUserData(image OSImage) string {
	return fmt.Sprintf(`#cloud-config
autoinstall:
  version: 1
  identity:
    hostname: %s
    username: ubuntu
    password: "!"
  ssh:
    install-server: true
  storage:
    layout:
      name: direct
`, image.Name)
}
