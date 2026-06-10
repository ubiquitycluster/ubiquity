# Install pre-commit hooks

Git hook scripts catch formatting, lint, and generated-artifact issues before a
change leaves your workstation. They do not replace CI, but they shorten the
feedback loop for common mistakes.

## Install pre-commit

Choose the install method that matches your workstation:

=== "macOS / Homebrew"

    ```sh
    brew install pre-commit
    ```

=== "Python / pipx"

    ```sh
    python3 -m pip install --user pipx
    pipx install pre-commit
    ```

=== "Arch Linux"

    ```sh
    sudo pacman -S python-pre-commit
    ```

## Install repository hooks

From the repository root, run:

```sh
make git-hooks
```

If you prefer to call pre-commit directly, run:

```sh
pre-commit install
```

## Verify the hooks

Run all hooks once before opening a review:

```sh
pre-commit run --all-files
```

If a hook rewrites files, inspect the diff, rerun the command, and include the
resulting changes in the same commit as the source change.

## Troubleshooting

- `pre-commit: command not found`: install pre-commit with one of the methods
  above and confirm the installation directory is on `PATH`.
- Hook environments fail to build: delete the cached environment with
  `pre-commit clean`, then rerun `pre-commit run --all-files`.
- Hooks disagree with CI: update local dependencies, rerun from a clean working
  tree, and prefer the CI result if the hook version differs.
- Generated docs are stale: run the generator mentioned by the failing hook and
  review the generated diff before committing.
