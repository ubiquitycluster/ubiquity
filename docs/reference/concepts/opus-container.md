# Opus container

The Opus container makes it easy to get all of the dependencies needed to interact with Ubiquity, and deploy it (hence the name, Opus - Or creation).

## How to open it

You can use it by deploying the container via Docker::

=== "Docker"

    ```sh
    make tools
    ```

It will open a shell like this:

```
[cjcshadowsan@bootstrap:~/ubiquity]$ echo hello
hello
```

## How it works

- All dependencies are defined in the `Dockerfile` and installed when the container is built
- The container is built using the `Makefile` in the root of the repository
- The container is run using the `Makefile` in the root of the repository
- It mounts the current directory into the container, so you can work on the files in the current directory
- It sets the current user as the user in the container, so files created in the container are owned by the current user
- It attaches SSH keys from the current user to the container, so you can use SSH to connect to other machines
- It provides you will all tools to manage or deploy a Ubiquity cluster including kubectl, helm, kustomize, terraform, ansible, K9s, clustershell, etc.

## Known issues

- If your Docker engine is not running in rootless mode, all files created by the tools container will be owned by `root`