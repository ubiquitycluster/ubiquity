package main

import "testing"

func TestMainPackageCompiles(t *testing.T) {
	// The executable package intentionally stays thin: main delegates command
	// behavior to cmd/ubiquity/cmd, where the Cobra command registry and command
	// behavior are covered. This smoke test keeps the root package under `go test`
	// so regressions in executable wiring are visible to package-level checks.
}
