# Deployment

Open the tools container if you haven't already:

=== "Docker"

    ```sh
    make tools
    ```

=== "Nix"

    ```sh
    nix-shell
    ```

Build the cluster:

```sh
make
```

Yes it's that simple!

!!! example

    <script id="asciicast-??????" src="https://asciinema.org/a/??????.js" async></script>

It will take a while to download everything,
you can read the [architecture document](../../architecture/overview.md) while waiting for the deployment to complete.
