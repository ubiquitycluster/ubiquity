package cloud

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestActiveChartsHaveTestsAndNonPlaceholderVersions(t *testing.T) {
	root := chartMaturityRepoRoot(t)
	versionRE := regexp.MustCompile(`(?m)^version:\s*([^\s]+)`)

	var missingTests []string
	var placeholderVersions []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		if d.IsDir() {
			switch {
			case rel == ".git", rel == "disabled", rel == "graphify-out", rel == "node_modules", rel == "vendor":
				return filepath.SkipDir
			case strings.Contains(rel, "/disabled/"):
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Base(path) != "Chart.yaml" {
			return nil
		}

		dir := filepath.Dir(path)
		relDir, err := filepath.Rel(root, dir)
		if err != nil {
			return err
		}
		relDir = filepath.ToSlash(relDir)

		if info, err := os.Stat(filepath.Join(dir, "tests")); err != nil || !info.IsDir() {
			missingTests = append(missingTests, relDir)
		}

		chart, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		match := versionRE.FindSubmatch(chart)
		if len(match) < 2 || strings.TrimSpace(string(match[1])) == "0.0.0" {
			placeholderVersions = append(placeholderVersions, relDir)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk active charts: %v", err)
	}

	sort.Strings(missingTests)
	sort.Strings(placeholderVersions)

	if len(missingTests) > 0 {
		t.Errorf("active charts missing tests directories:\n%s", strings.Join(missingTests, "\n"))
	}
	if len(placeholderVersions) > 0 {
		t.Errorf("active charts missing non-placeholder versions:\n%s", strings.Join(placeholderVersions, "\n"))
	}
}

func chartMaturityRepoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate repository root from working directory")
		}
		dir = parent
	}
}
