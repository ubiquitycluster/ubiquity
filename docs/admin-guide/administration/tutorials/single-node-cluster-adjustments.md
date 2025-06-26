# Single node cluster adjustments

Update the following changes, then commit and push.

## Reduce Longhorn replica count

Set the `defaultClassReplicaCount` to 1:

```yaml title="system/longhorn-system/values.yaml" hl_lines="6"
--8<--
system/longhorn-system/values.yaml
--8<--
```

## Disable automatic upgrade for OS and k3s

Because they will try to drain the only node, the pods will have no place to go.
Remove them entirely:

```sh
mv system/kured disabled/system/kured
mv system/system-upgrade disabled/system/system-upgrade
```

Commit and push the change.
You can revert it later when you add more nodes by just moving them back:

```sh
mv disabled/system/kured system/kured
mv disabled/system/system-upgrade system/system-upgrade
```

This process of enabling/disabling is the same for all the other components and makes it easy to add/remove them.
