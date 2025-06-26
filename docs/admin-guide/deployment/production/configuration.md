# Configuration

Open the [tools container](../../runbooks/tools-container.md), which includes all the tools needed:

=== "Docker"

    ```sh
    make tools
    ```

=== "Nix"

    ```sh
    nix-shell
    ```

!!! note

     It will take a while to build the tools container on the first time

Run the following script to configure the environment:

```sh
make configure
```

!!! example

    <!-- TODO update example input -->
    ```
    Text editor (nvim):
    Enter seed repo (github.com/cjcshadowsan/ubiquity): github.com/my-cluster/ubiquity
    Enter your domain (ubiquitycluster.uk): example.com
    ```

It will prompt you to edit the inventory:

- IP address: the desired one, not the current one, since your servers have no operating system installed yet
- Disk: based on `/dev/$DISK`, in my case it's `sda`, but yours can be `sdb`, `nvme0n1`...
- Network interface: usually it's `eth0`, mine is `eno1`, could be `en<s for slot, number><f for function number starting from 0> - so ens4f0`
- External address: an address. Can be the same as the internal IP address
- External interface (optional): usually it's `eth0`, mine is `eno1`, could be `en<s for slot, number><f for function number starting from 0> - so ens4f0`
- Wake on Lan: true or false, or otherwise whether to use IPMI or not
- MAC address: the **lowercase, colon separated** MAC address of the above network interface

!!! example

    ```yaml title="metal/inventories/prod.yml"
    --8<--
    metal/inventories/prod.yml
    --8<--
    ```

At the end it will show what has changed. After examining the diff, commit and push the changes.
