package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/docker"
)

func TestToolsContainer(t *testing.T) {
	if testing.Short() || os.Getenv("UBIQUITY_RUN_TOOLS_CONTAINER_TEST") != "true" {
		t.Skip("skipping tools container test; set UBIQUITY_RUN_TOOLS_CONTAINER_TEST=true to run")
	}

	image := "nixos/nix"
	projectRoot, _ := filepath.Abs("../")

	options := &docker.RunOptions{
		Remove: true,
		Volumes: []string{
			fmt.Sprintf("%s:%s", projectRoot, projectRoot),
			"ubiquity-tools-cache:/root/.cache",
			"ubiquity-tools-nix:/nix",
		},
		OtherOptions: []string{
			"--workdir", projectRoot,
		},
		Command: []string{
			"nix-shell",
			"--command", "exit",
		},
	}

	docker.Run(t, image, options)
}
