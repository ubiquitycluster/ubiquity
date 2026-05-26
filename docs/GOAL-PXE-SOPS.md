Implement a Go-based PXE installer and SOPS secrets management for Ubiquity at /home/ubuntu/ubiquity. These are equivalent implementations of features from the upstream homelab project, built from scratch to work within Ubiquity's environment (no NixOS, no nixos-anywhere).

## What we're building

### Feature A: Go-based PXE installer
homelab uses a NixOS installer image + `nixos-anywhere`. Ubiquity's equivalent is a Go binary that handles the initial PXE boot orchestration via the `pixiecore` library, and then hands off to Ubiquity's existing Ansible playbooks for OS installation and configuration (Rocky Linux / Fedora / Ubuntu via kickstart/preseed).

### Feature B: SOPS secrets management
homelab uses SOPS with Age keys for GitOps-compatible secrets. Ubiquity will adopt the same approach, replacing/integrating with the existing Vault setup.

---

## P0 Items (must do)

### P0-A1: Create the Go PXE installer binary

Create `tools/cmd/ubiquity-install/main.go` — a standalone Go binary that:
- Listens on port 67 (DHCP proxy) and 69 (TFTP) using the `pixiecore` library
- Serves a PXE boot image (kernel + initrd) to machines on the network
- Has a MAC→hostname mapping to identify machines
- Exposes a phone-home HTTP API that machines call after booting
- Coordinates the installation flow: PXE → phone home → trigger Ansible

Add `go.universe.tf/netboot` as a dependency.

Structure:

```go
// tools/cmd/ubiquity-install/main.go
package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/exec"
    "os/signal"
    "sync"
    "syscall"

    "go.universe.tf/netboot/pixiecore"
)

// hostMap maps MAC addresses to hostnames/kickstart files
var hostMap = map[string]string{
    "bc:24:11:d0:28:34": "node1",
    "bc:24:11:0d:2f:20": "node2",
}

type PhoneHomePayload struct {
    MAC  string `json:"mac"`
    IP   string `json:"ip"`
    Host string `json:"host,omitempty"`
}

var (
    installed = make(map[string]bool)
    inFlight  = make(map[string]bool)
    mu        sync.Mutex
    kernel    = flag.String("kernel", "", "Path to kernel bzImage")
    initrd    = flag.String("initrd", "", "Path to initrd")
    address   = flag.String("address", "0.0.0.0", "Address to listen on")
    apiPort   = flag.Int("api-port", 8080, "Phone-home API port")
)

type bootHandler struct {
    kernel, initrd string
}

func (b bootHandler) BootSpec(m pixiecore.Machine) (*pixiecore.Spec, error) {
    mac := m.MAC.String()
    mu.Lock()
    inFlight[mac] = true
    mu.Unlock()
    log.Printf("PXE boot request from %s (%s)", m.MAC, m.IP)
    return &pixiecore.Spec{
        Kernel: pixiecore.ID("kernel"),
        Initrd: []pixiecore.ID{"initrd"},
        Cmdline: "console=ttyS0,115200n8 console=tty0",
    }, nil
}

func phoneHomeHandler(w http.ResponseWriter, r *http.Request) {
    var p PhoneHomePayload
    if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
        http.Error(w, err.Error(), 400)
        return
    }
    log.Printf("Phone home from %s (%s)", p.MAC, p.IP)
    mu.Lock()
    installed[p.MAC] = true
    delete(inFlight, p.MAC)
    host, ok := hostMap[p.MAC]
    mu.Unlock()
    if ok {
        // Trigger Ansible playbook for this host
        go runAnsible(host, p.IP)
    }
    w.WriteHeader(200)
}

func runAnsible(host, ip string) {
    cmd := exec.Command("ansible-playbook",
        "-i", fmt.Sprintf("%s,", ip),
        "metal/boot.yml",
        "--limit", host,
    )
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        log.Printf("Ansible failed for %s: %v", host, err)
    }
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
    mu.Lock()
    defer mu.Unlock()
    json.NewEncoder(w).Encode(map[string]interface{}{
        "installed": installed,
        "in_flight": inFlight,
    })
}

func main() {
    flag.Parse()
    if *kernel == "" || *initrd == "" {
        log.Fatal("--kernel and --initrd are required")
    }

    // PXE server
    pxe, err := pixiecore.NewServer(pixiecore.Config{
        Handler:   bootHandler{kernel: *kernel, initrd: *initrd},
        Address:   *address,
        DHCPProxy: true, // Act as DHCP proxy, not full DHCP server
    })
    if err != nil {
        log.Fatalf("Failed to start PXE server: %v", err)
    }
    go pxe.Serve()

    // Phone-home API
    http.HandleFunc("/phone-home", phoneHomeHandler)
    http.HandleFunc("/status", statusHandler)
    apiAddr := fmt.Sprintf("%s:%d", *address, *apiPort)
    log.Printf("Phone-home API listening on %s", apiAddr)
    go http.ListenAndServe(apiAddr, nil)

    // Wait for shutdown
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
    <-sig
}
```

Also scaffold:
- tools/go.mod — Go module for the installer
- tools/go.sum — Run go mod tidy

Add a Makefile target:
```makefile
# Build the PXE installer binary
installer:
	go build -o ubiquity-installer ./tools/cmd/ubiquity-install/
```

Commit: git add tools/ Makefile && git commit -m "feat: add Go-based PXE installer binary with pixiecore and phone-home API"

### P0-A2: Create installer kickstart templates

Ubiquity installs Rocky Linux/Fedora/Ubuntu via PXE using kickstart/preseed files, not NixOS. Create the kickstart templates that the PXE installer serves.

Create metal/kickstart/ directory with:

metal/kickstart/rocky9.ks:
```kickstart
#version=RHEL9
text
network --bootproto=dhcp --device=link --activate
url --url="http://{{ .MirrorURL }}/rocky/9/BaseOS/x86_64/os/"
lang en_US.UTF-8
keyboard us
timezone {{ .Timezone }} --utc
rootpw --iscrypted {{ .RootPasswordHash }}
user --name=admin --groups=wheel --iscrypted --password={{ .AdminPasswordHash }}
selinux --enforcing
firewall --enabled --service=ssh
services --enabled=NetworkManager,sshd
skipx
zerombr
clearpart --all --initlabel
autopart --type=lvm
bootloader --location=mbr
%packages
@^server-product-environment
kexec-tools
%end
reboot
```

metal/kickstart/ubuntu24.04.user-data:
```yaml
#cloud-config
autoinstall:
  version: 1
  identity:
    hostname: {{ .Hostname }}
    password: {{ .PasswordHash }}
    username: ubuntu
  ssh:
    install-server: true
    authorized-keys:
      - {{ .SSHKey }}
  storage:
    layout:
      name: lvm
  packages:
    - qemu-guest-agent
  late-commands:
    - curtin in-target -- /usr/bin/systemctl enable ssh
```

metal/Makefile should have a target:
```makefile
# Generate kickstart files from templates
kickstart:
	./scripts/gen-kickstart
```

git add metal/kickstart/ && git commit -m "feat: add kickstart/preseed templates for PXE OS installation"

### P0-A3: Add SOPS secrets management

homelab uses SOPS with Age keys for encrypting secrets in Git. Add equivalent support.

Install sops: go install github.com/getsops/sops/v3/cmd/sops@latest

Create .sops.yaml at repo root:
```yaml
creation_rules:
  - path_regex: secrets/.*\.enc\.yaml
    age: age1abc123...  # placeholder, user generates their own Age key
```

Create secrets/ directory:
- secrets/README.md — explaining how to use SOPS
- secrets/.gitkeep

Create a helper script scripts/sops-edit:
```bash
#!/bin/bash
# Edit an encrypted secret file
# Usage: ./scripts/sops-edit secrets/example.enc.yaml
set -euo pipefail
sops "$@"
```

Create a helper script scripts/sops-decrypt:
```bash
#!/bin/bash
# Decrypt secret to stdout (for piping to kubectl etc)
# Usage: ./scripts/sops-decrypt secrets/example.enc.yaml
set -euo pipefail
exec sops --decrypt "$1"
```

Make both executable: chmod +x scripts/sops-*

Create secrets/secrets.enc.yaml example:
```yaml
# Encrypted with SOPS + Age
# Edit with: ./scripts/sops-edit secrets/secrets.enc.yaml
# Decrypt with: ./scripts/sops-decrypt secrets/secrets.enc.yaml
apiVersion: v1
kind: Secret
metadata:
  name: cluster-secrets
  namespace: global-secrets
type: Opaque
stringData:
  # Actual values encrypted with sops
  cloudflare_api_token: ENC[AES256_GCM,data:...,iv:...,tag:...]
  admin_password: ENC[AES256_GCM,data:...,iv:...,tag:...]
sops:
  kms: []
  gcp_kms: []
  azure_kv: []
  hc_vault: []
  age:
    - recipient: age1...  # placeholder
      enc: |
        ---
        # encrypted key
  lastmodified: "2026-05-26T00:00:00Z"
  mac: ENC[AES256_GCM,data:...]
```

Create docs/how-to-guides/secrets-management.md:
```markdown
# Secrets Management with SOPS

## Overview
SOPS (Secrets OPerationS) encrypts secrets using Age keys.
Encrypted files are committed to Git; decryption happens at deploy time.

## Setup
1. Generate an Age key: `age-keygen -o ~/.config/sops/age/keys.txt`
2. Export the public key: `age-keygen -y ~/.config/sops/age/keys.txt`
3. Add the public key to .sops.yaml

## Usage
Edit a secret:
```
./scripts/sops-edit secrets/mysecret.enc.yaml
```

Decrypt a secret:
```
./scripts/sops-decrypt secrets/mysecret.enc.yaml
```

Decrypt and pipe to kubectl:
```
./scripts/sops-decrypt secrets/mysecret.enc.yaml | kubectl apply -f -
```

## Integration with External Secrets Operator
SOPS-encrypted files can be used with the External Secrets Operator
via the `sops` provider (or by pre-decrypting with a GitOps tool like
ArgoCD's SOPS plugin).
```

Add sops-note to .pre-commit-config.yaml (optional, no strict enforcement):
```yaml
  # Warn if unencrypted secrets are committed
  - repo: local
    hooks:
      - id: sops-encrypted
        name: check secrets are encrypted
        entry: ./scripts/check-sops
        language: script
        files: ^secrets/
```

Commit: git add .sops.yaml secrets/ scripts/sops-edit scripts/sops-decrypt docs/how-to-guides/secrets-management.md && git commit -m "feat: add SOPS secrets management with Age encryption"

### P0-A4: Write ADR-010 for Go PXE installer and ADR-011 for SOPS secrets

ADR-010: Use Go PXE installer instead of Python/Docker

Context: Original PXE server was a Docker Compose-based setup using dhcpd + tftpd + httpd (Apache). The upstream homelab project rewrote theirs in Go using the pixiecore library, which provides a single-binary PXE solution with DHCP proxy support.
Decision: Rewrite the PXE server as a Go binary using pixiecore. Single static binary, no Docker dependency for the installer, built-in DHCP proxy mode avoids conflicts with existing network DHCP servers.
Consequences:
- Positive: Single binary, no Docker required for PXE boot
- Positive: DHCP proxy mode works alongside existing network DHCP (opt-in)
- Positive: Built-in phone-home API for installation coordination
- Negative: Must manually extract kernel/initrd from OS install media (no longer using Docker Compose)
- Neutral: MAC→host mapping is static (config map in Go source, can be externalized later)

ADR-011: Use SOPS for secrets management

Context: Need to manage secrets (API tokens, passwords, TLS certs) in a GitOps-compatible way. Previously used Vault (complex to operate) and Sealed Secrets (Kubernetes-specific). The upstream homelab project uses SOPS with Age keys.
Decision: Use SOPS with Age keys for encrypting secrets at rest. Encrypted files are committed to Git; decryption happens at deploy time or via ArgoCD's SOPS plugin.
Consequences:
- Positive: Works offline (no dependency on Vault cluster)
- Positive: Encrypted files are standard YAML — Git diff shows encrypted content has changed
- Positive: Age keys are simpler than PGP/GPG
- Negative: Key management is manual (Age private key must be distributed to team members)
- Neutral: Coexists with Vault (can use both for different use cases)

Update README.md to link to both new ADRs.

Commit: git add docs/reference/architecture/decision-records/ADR-010-installer.md docs/reference/architecture/decision-records/ADR-011-sops.md && git commit -m "docs: add ADR-010 for Go PXE installer and ADR-011 for SOPS secrets"

---

## P1 Items (high priority)

### P1-A1: Wire the PXE installer into ubiquity up

In cmd/ubiquity/cmd/up.go, add a `provisionPXE` function that:
- Checks if the installer binary exists at `tools/ubiquity-installer`
- If yes: runs it in the background with `--kernel` and `--initrd` pointing to extracted installer media
- If no: falls back to the existing Docker-based PXE server (via Ansible)

Add a `--pxe-installer` flag to the up command to explicitly enable/disable.

git commit -m "feat: wire Go PXE installer into ubiquity up --pxe-installer" -a

### P1-A2: Add sops decrypt step to provisioning pipeline

In the bootstrap phase (provisionBootstrap), add a step that decrypts SOPS secrets before applying:
```go
// Decrypt SOPS secrets for the cluster
cmd := exec.Command("./scripts/sops-decrypt", "secrets/secrets.enc.yaml")
decrypted, err := cmd.Output()
if err == nil {
    // Pipe to kubectl
    applyCmd := exec.Command("kubectl", "apply", "-f", "-")
    applyCmd.Stdin = bytes.NewReader(decrypted)
    applyCmd.Run()
}
```

git commit -m "feat: add sops decrypt step to provisioning pipeline" -a

### P1-A3: Add installer status to ubiquity status

The status command should show the installer phone-home status if the installer is running.

git commit -m "feat: show PXE installer status in ubiquity status" -a

---

## Verification

After all items:
- go build ./... — PASS
- go test ./pkg/... ./cmd/... -count=1 — all green
- go build -o /dev/null ./tools/cmd/ubiquity-install/... — installer builds
- ls tools/cmd/ubiquity-install/main.go — exists
- ls .sops.yaml — exists
- ls secrets/secrets.enc.yaml — exists
- ls docs/how-to-guides/secrets-management.md — exists
- ls docs/reference/architecture/decision-records/ADR-010-installer.md — exists
- ls docs/reference/architecture/decision-records/ADR-011-sops.md — exists
- ubiquity up --help — shows --pxe-installer flag
