Complete the remaining gaps in the Ubiquity CLI transformation at /home/ubuntu/ubiquity.

## What's done already

- CLI scaffold: init, up (6-phase pipeline), down, status, logs, retry, test, configure
- CI pipeline: .github/workflows/ci.yaml (5 jobs), .pre-commit-config.yaml (7 hooks), renovate.json5
- Security: Kyverno, kube-bench, NetworkPolicies — 3 Helm charts, wired as phase 3/6
- State management: JSON state at ~/.ubiquity/state.json, phase lifecycle, retry
- Config port: pkg/config/ with .env management, YAML patching, interactive wizard
- Testing: 24 tests across config, provision, network, cli
- Dead code cleanup: site/ untracked, stale license URLs fixed, dead Makefile targets removed

## What's left

### 1. Wire Viper config framework
The goal specifies cobra + viper. Currently only cobra is used. Wire Viper into the CLI root command for loading config from .ubiquity.yaml.

Tasks:
- Add `github.com/spf13/viper` to go.mod (go get)
- In cmd/ubiquity/cmd/root.go: import viper, call viper.SetConfigName(".ubiquity") and viper.AutomaticEnv() in the root init()
- In cmd/ubiquity/cmd/init.go: after creating ~/.ubiquity/, generate a skeleton .ubiquity.yaml with default config values (domain, timezone, editor, ingress provider)
- Wire viper to read --env and --config flags and bind them with viper.BindPFlag so env flags work via both --flag and UBQUITY_ENV env var
- Verify: `go build ./...` passes, `ubiquity init` creates ~/.ubiquity/.ubiquity.yaml

### 2. Remove original Python configure scripts
The Go port (pkg/config/ + `ubiquity configure`) works, but the original Python scripts are still on disk.

- Move scripts/configure to scripts/configure.py.bak (or delete)
- Move scripts/configure-sandbox to scripts/configure-sandbox.py.bak (or delete)
- Add a comment in the Makefile noting the configure target now delegates to `ubiquity configure`

### 3. Add helm unittest tests to remaining security charts
Only 2 of ~24 charts have helm unittest tests. The 3 security charts (kyverno-policies, kube-bench, network-policies) need them.

For each chart in system/kyverno-policies/ system/kube-bench/ system/network-policies/:
- Create tests/<chart>_test.yaml with basic assertions:
  - Renders without error
  - Has expected number of resources
  - Specific templates render with correct metadata
- Verify with: `helm unittest system/<chart>/`

### 4. Eliminate 40 duplicate files
pygount reports 40 duplicate files. Consolidate them:

- Find duplicates with: `fdupes -r .` or pygount's output
- Extract shared content into pkg/provision/templates/ or a shared Kustomize base
- Replace duplicates with symlinks or Kustomize resource references
- Verify pygount shows 0 duplicates

### 5. Wire Terraform subprocess wrapper in external phase
The goal says the CLI wraps terraform/ansible/helm as subprocesses. Only helm is wrapped.

In cmd/ubiquity/cmd/up.go, update provisionExternal to:
- Check if terraform is installed (exec.LookPath)
- If env==sandbox: skip (already done)
- For prod: run `terraform init && terraform apply -auto-approve` in the appropriate cloud directory
- Set working directory based on env (cloud/aws, cloud/azure, cloud/gcp, cloud/openstack, cloud/ovh)
- Pipe output to stdout so user sees terraform progress
- Add a provisionAnsible helper (for the metal phase) similarly

### 6. Add KUTTL integration test skeleton
Create integration/ directory:

- integration/kuttl-test.yaml — KUTTL test config
- integration/assertions/ — placeholder assertion files
- integration/README.md — notes on running KUTTL tests

### 7. Add molecule test content
The molecule/ directory exists but has no real test content.

- Create molecule/default/converge.yml that actually applies a simple role (e.g., metal/roles/tune)
- Create molecule/default/verify.yml that checks the role had its intended effect
- molecule/default/molecule.yml already exists, just make sure it's valid

### 8. Add Bubbletea TUI to status output (stretch goal)
If time permits, replace the plain-text status output with a charmbracelet/bubbletea TUI:

- Add `github.com/charmbracelet/bubbletea` and `github.com/charmbracelet/lipgloss` to go.mod
- Create pkg/tui/status.go with a bubbletea model that renders the provisioning state
- The model should show a table of phases with colored status indicators
- Wire it into cmd/ubiquity/cmd/status.go (fallback to plain text if TUI fails)
- Add a --plain flag to skip TUI

## Execution

Work through items 1-7 in order. Each item is independent enough for delegate_task. Commit after each item with a descriptive message.

Item 8 (Bubbletea TUI) is a stretch — do it last if there's time.

## Verification

After all items, run:
- go build ./... — must pass
- go test ./pkg/... ./cmd/... -v -count=1 — all tests green
- helm lint system/kyverno-policies/ system/kube-bench/ system/network-policies/ — all pass
- ubiquity init && ubiquity up --sandbox --skip-security — 6-phase pipeline runs
- pygount --format=summary --folders-to-skip=".git,node_modules,site" . | grep duplicate — 0 duplicates
