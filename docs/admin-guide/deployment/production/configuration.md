# Configuration

Open the [tools container](../../runbooks/tools-container.md), which includes the
required deployment tools.

=== "Docker"

    ```sh
    make tools
    ```

=== "Nix"

    ```sh
    nix-shell
    ```

!!! note

     It will take a while to build the tools container the first time.

Run the configuration workflow:

```sh
make configure
```

!!! example "example input"

    ```text
    Text editor (nvim): nvim
    Enter seed repo (github.com/ubiquitycluster/ubiquity): github.com/example/platform-gitops
    Enter your domain (ubiquity.example.com): example.com
    Enter cluster name (ubiquity): prod-a
    Enter external DNS provider (cloudflare): cloudflare
    Enter lifecycle backend (nico): nico
    ```

The workflow prompts you to edit inventory and environment inputs. Review these
values before any live apply:

- IP address: desired static address for each machine, not a temporary install
  address
- Disk: the target installation disk such as `sda`, `sdb`, or `nvme0n1`
- Network interface: the production NIC name such as `eth0`, `eno1`, or `ens4f0`
- External address: the address used by ingress, management, or load balancer
  traffic
- External interface: optional interface used for external traffic
- Wake on LAN / BMC: whether power operations use WoL, IPMI, Redfish, or NICo
- MAC address: lowercase, colon-separated MAC address for PXE and inventory
  matching
- Credential references: Vault, sealed-secret, or password-manager references;
  do not place raw passwords or tokens in Git

!!! example "inventory excerpt"

    ```yaml title="metal/inventories/prod.yml"
    --8<--
    metal/inventories/prod.yml
    --8<--
    ```

## Production-lite caveats

Production-lite environments may share operator workstations, smaller control
planes, or simplified external dependencies. Treat them as validation
installations, not as proof that the production environment is ready. Record
which production requirements are intentionally absent, such as HA load balancer,
offsite backup, multi-operator credential custody, or hardware redundancy.

## Proof boundary

Dry-run/local proof confirms that configuration files render and that commands
can execute without mutating infrastructure. Live production proof requires the
actual environment to show Kubernetes readiness, NICo lifecycle evidence, DNS and
certificate issuance, backup target reachability, and workload smoke tests.

At the end of configuration, examine the diff, confirm that no raw credentials
were written, then commit and push the reviewed changes through the normal change
process.
