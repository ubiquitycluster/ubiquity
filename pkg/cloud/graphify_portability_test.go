package cloud

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestGraphifyArtifactsAreRepoPortable(t *testing.T) {
	tracked := gitLsFiles(t, "../../", "graphify-out")
	if len(tracked) == 0 {
		t.Fatal("expected tracked graphify-out artifacts")
	}

	for _, rel := range tracked {
		path := filepath.Join("../../", rel)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", rel, err)
		}
		if info.IsDir() {
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		if bytes.Contains(content, []byte("/Users/")) || bytes.Contains(content, []byte("/home/")) {
			t.Fatalf("%s contains machine-local home-directory paths; graphify artifacts must be repo-root-relative", rel)
		}
	}
}

func TestGraphifyWorkflowUsesPortableOKFBundle(t *testing.T) {
	for _, path := range []string{
		"../../graphify-out/wiki/index.md",
		"../../graphify-out/wiki/repository-knowledge-graph.md",
		"../../scripts/normalize-graphify-artifacts.py",
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected portable Graphify/OKF artifact %s: %v", path, err)
		}
	}

	indexBytes, err := os.ReadFile("../../graphify-out/wiki/index.md")
	if err != nil {
		t.Fatalf("read OKF index: %v", err)
	}
	index := string(indexBytes)
	for _, required := range []string{
		"okf_version: \"0.1\"",
		"Open Knowledge Format",
		"repository-root-relative",
		"repository-knowledge-graph.md",
	} {
		if !strings.Contains(index, required) {
			t.Fatalf("Graphify OKF index missing %q", required)
		}
	}

	docBytes, err := os.ReadFile("../../docs/developers/graphify-workflow.md")
	if err != nil {
		t.Fatalf("read Graphify workflow doc: %v", err)
	}
	doc := string(docBytes)
	for _, required := range []string{
		"Open Knowledge Format",
		"scripts/normalize-graphify-artifacts.py",
		"repository-root-relative",
		"graphify-out/wiki/index.md",
	} {
		if !strings.Contains(doc, required) {
			t.Fatalf("Graphify workflow doc missing portable OKF guidance %q", required)
		}
	}
}

func TestGraphifyManifestHasUniquePathKeys(t *testing.T) {
	content, err := os.ReadFile("../../graphify-out/manifest.json")
	if err != nil {
		t.Fatalf("read Graphify manifest: %v", err)
	}

	keyLine := regexp.MustCompile(`(?m)^  "([^"]+)":`)
	counts := map[string]int{}
	for _, match := range keyLine.FindAllSubmatch(content, -1) {
		counts[string(match[1])]++
	}

	var duplicates []string
	for key, count := range counts {
		if count > 1 {
			duplicates = append(duplicates, key)
		}
	}
	if len(duplicates) > 0 {
		limit := len(duplicates)
		if limit > 5 {
			limit = 5
		}
		t.Fatalf("graphify-out/manifest.json contains %d duplicate JSON object keys; examples: %s", len(duplicates), strings.Join(duplicates[:limit], ", "))
	}
}

func gitLsFiles(t *testing.T, dir string, args ...string) []string {
	t.Helper()
	cmdArgs := append([]string{"ls-files"}, args...)
	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %s: %v", strings.Join(cmdArgs, " "), err)
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files
}
