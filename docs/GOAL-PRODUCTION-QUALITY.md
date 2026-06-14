Take the Ubiquity CLI to production quality at /home/ubuntu/ubiquity. All 15 items below must be completed. Commit after each item with a descriptive message.

## Prerequisites
Go 1.22+, git, goreleaser (install via `go install github.com/goreleaser/goreleaser/v2@latest`), docker (for multi-platform build verification).

## Items

### 1. Add version command with ldflags
Wire build-time version info into the binary so `ubiquity version` and `ubiquity --version` both work.

- Create cmd/ubiquity/cmd/version.go with a cobra command that prints Version, Commit, Date
- Add a `var (Version = "dev" Commit = "none" Date = "unknown")` block in root.go or main.go
- Wire rootCmd.Flags().Bool("version", false, "print version") in root.go's init() that prints version on `--version` flag
- Add `-ldflags "-X github.com/ubiquitycluster/ubiquity/cmd/ubiquity/cmd.Version=$(VERSION) -X github.com/ubiquitycluster/ubiquity/cmd/ubiquity/cmd.Commit=$(COMMIT) -X github.com/ubiquitycluster/ubiquity/cmd/ubiquity/cmd.Date=$(DATE)"` to the build target in the Makefile
- Verify: `go build -ldflags "-X github.com/ubiquitycluster/ubiquity/cmd/ubiquity/cmd.Version=1.0.0 -X github.com/ubiquitycluster/ubiquity/cmd/ubiquity/cmd.Commit=abc123" ./cmd/ubiquity/... && go run ./cmd/ubiquity/... version` prints "1.0.0"
- git commit: `git add cmd/ubiquity/cmd/version.go cmd/ubiquity/cmd/root.go Makefile && git commit -m "feat: add version command with build-time ldflags"`

### 2. Fix .bak files in git
The deprecated Python configure scripts are `.py.bak` files still tracked in git.

- Run: `git rm scripts/configure.py.bak scripts/configure-sandbox.py.bak`
- Verify they're gone with `git ls-files scripts/*.bak`
- git commit: `git commit -m "chore: remove deprecated Python configure script backups"`

### 3. Wire ubiquity down to actually tear down
Currently `ubiquity down` just prints a message. Wire it to tear down based on provisioning state.

- Read the provisioning state from `provision.LoadState()`
- If environment == "sandbox": run `k3d cluster delete ubiquity-dev` (gracefully if k3d not available)
- If environment is a cloud provider (aws, azure, gcp, openstack, ovh): run `terraform destroy -auto-approve` in the corresponding cloud/ directory
- Delete the state file at the end: `os.Remove(provision.StatePath())`
- Show a progress message for each step
- git commit: `git add cmd/ubiquity/cmd/down.go && git commit -m "feat: wire ubiquity down to tear down cluster and cloud resources"`

### 4. Wire ubiquity logs to read real state/log files
Currently `ubiquity logs` prints a static message. Wire it to read from the provisioning state.

- Change logs.go RunE to load the provisioning state
- If no phase arg: print all phases with their error messages and log URLs from the state
- If phase arg provided: print that phase's details (status, duration, error, log_url)
- If no state exists: print "No provisioning state found. Run 'ubiquity up' first."
- Keep the `logs [phase]` positional argument
- git commit: `git add cmd/ubiquity/cmd/logs.go && git commit -m "feat: wire ubiquity logs to read from provisioning state"`

### 5. Increase CLI test coverage to >50%
Current coverage: cmd/ubiquity/cmd = 8%, pkg/network = 0%.

For cmd/ubiquity/cmd/:
- Add root_test.go tests for:
  - TestVersionFlag: verifies --version flag is registered
  - TestVersionCmd: verifies version command is registered
  - TestConfigureCmdHelp: verifies configure command has expected Use and Flags
  - TestRetryCmd: verifies retry command registered with valid args
- Refactor phase executors into a Provider interface in pkg/provision/exec.go:
  - type Provider interface { Metal(env string) error; Bootstrap(env string) error; ... }
  - Create a default RealProvider that uses exec.Command
  - Make up.go use the Provider interface
  - Create a MockProvider for testing
- Add cmd/ubiquity/cmd/up_test.go with tests using MockProvider
- Target: cmd/ubiquity/cmd coverage >50%, pkg/provision >85%

Run `go test ./cmd/ubiquity/cmd/... -coverprofile=cov.out && go tool cover -func=cov.out` to verify.

git commit: `git add cmd/ubiquity/cmd/ cmd/ubiquity/cmd/*_test.go pkg/provision/ && git commit -m "test: refactor for testability, add Provider interface, boost CLI coverage >50%"`

### 6. Add goreleaser config for binary releases
Users should be able to download pre-built binaries for their platform.

- Create .goreleaser.yaml at repo root:
  - Build for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64
  - Use the ldflags from item 1 for version injection
  - Create .tar.gz archives
  - Create homebrew tap formula (brew tap ubiquitycluster/tap)
  - Add a checksums file
- Add a GitHub Actions workflow `.github/workflows/release.yaml`:
  - Trigger on tags matching `v*`
  - Run goreleaser release --clean
  - Upload artifacts to GitHub release
- Verify with `goreleaser check` and `goreleaser build --snapshot --clean`

git commit: `git add .goreleaser.yaml .github/workflows/release.yaml && git commit -m "ci: add goreleaser config and release workflow"`

### 7. Update Makefile default target
The `make` command should guide users to the CLI, not run the old pipeline.

- Change the `default:` target to print a message and delegate to the CLI
- Add a comment at the top of the Makefile: "# The ubiquity CLI is the recommended entry point. Run 'ubiquity up' instead."
- Keep existing targets for backward compatibility
- git commit: `git commit -m "chore: update Makefile default to guide users to ubiquity CLI" -a`

### 8. Populate pkg/network with real IPAM/DNS logic
The network package is just doc.go. Port the IP address calculation and DNS config logic.

From the original configure scripts, extract:
- CIDR to netmask conversion: take `10.0.0.0/22` and produce `255.255.252.0`
- CIDR to broadcast address: take `10.0.0.0/22` and produce `10.0.3.255`
- Gateway calculation: network base + offset
- Provisioner IP: first usable or specific offset
- DNS config: generate dnsmasq configuration based on domain + search domain

Each function should be pure (no I/O), take strings, return strings, with unit tests.

Create pkg/network/cidr.go with:
- func NetworkToNetmask(cidr string) (string, error) — e.g. "10.0.0.0/22" → "255.255.252.0"
- func NetworkToBroadcast(cidr string) (string, error) — e.g. "10.0.0.0/22" → "10.0.3.255"
- func NetworkToGateway(cidr string, offset int) (string, error) — e.g. "10.0.0.0/22", 254 → "10.0.3.254"
- func IsValidCIDR(cidr string) bool — already exists in pkg/config, move here or keep both

Create pkg/network/cidr_test.go with tests for each function.

git commit: `git add pkg/network/ && git commit -m "feat: add IPAM/DNS logic to pkg/network with CIDR calculations"`

### 9. Add CLI Dockerfile
Containerized CLI build for CI pipelines.

- Create Dockerfile.cli at repo root:
  - Stage 1: golang:1.23 AS build — build with ldflags
  - Stage 2: gcr.io/distroless/static-debian12 — copy binary, entrypoint
  - Binary at /usr/local/bin/ubiquity
- git commit: `git add Dockerfile.cli && git commit -m "build: add multi-stage Dockerfile for CLI binary"`

### 10. Add ubiquity binary to .gitignore
Prevent accidentally committing the Go build artifact.

- Add to .gitignore:
  - `ubiquity` (the binary)
  - `ubiquity-cli` (builder artifact)
  - `dist/` (goreleaser output)
- git commit: `git commit -m "chore: add ubiquity binary and dist/ to .gitignore" -a`

### 11. Update README to show the CLI
The README.md still references Makefile commands. Update it.

- Find the "Get Started" section and add a note:
  "### Quick Start with the CLI
   ```
   ubiquity init
   ubiquity configure --domain mycluster.example.com
   ubiquity up --sandbox
   ```
   See the CLI help: `ubiquity --help`"
- Keep existing documentation for backward compatibility
- git commit: `git commit -m "docs: update README with CLI quick start commands" -a`

### 12. Remove orphan scripts/test.py
This 24-line Python test script is leftover from the old Python era.

- Run: `git rm scripts/test.py`
- git commit: `git commit -m "chore: remove orphan scripts/test.py"`

### 13. Add --json flag to version command
Machine-readable version output for CI scripts.

- Add a --json flag to the version command
- When --json is passed, print JSON with fields: version, commit, date, go_version, os, arch
- git commit: `git commit -m "feat: add --json flag to version command" -a`

### 14. Generate shell completion files
Ship completion files for bash, zsh, fish.

- Add a Makefile target: `completions` that runs `ubiquity completion bash > completions/ubiquity.bash` and same for zsh, fish
- Create completions/ directory
- Add a CI step or note in release workflow to generate and ship these
- git commit: `git commit -m "build: generate shell completion files for bash/zsh/fish" -a`

### 15. Wire version ldflags into CI build
Connect the goreleaser config from item 6 with the ldflags from item 1.

- Ensure .goreleaser.yaml uses ldflags with {{ .Version }}, {{ .Commit }}, {{ .Date }}
- Ensure `make build` passes ldflags
- The install target should use the same ldflags
- git commit: `git commit -m "build: wire version ldflags into Makefile and goreleaser builds" -a`

## Verification

After all 15 items:

- go build ./... — PASS
- go test ./pkg/... ./cmd/... -count=1 — all green
- ubiquity version — prints version info
- ubiquity version --json — prints valid JSON
- ubiquity down — reads state, calls k3d or terraform destroy
- ubiquity logs — reads from provisioning state
- go test ./cmd/ubiquity/cmd/... -cover | grep -o '[0-9.]*%'  — >50%
- goreleaser check — PASS
- git ls-files scripts/*.bak — empty
- ls scripts/test.py — should not exist
- git log --oneline -15 — 15 commits, one per item
