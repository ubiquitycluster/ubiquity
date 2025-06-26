Changelog
---------

**8.0.0+1.12.3**

- Cilium v1.12.3
- introduce `cilium_delegate_to` variable. Previously this was hard coded to `127.0.0.1` and it's also the default of this variable.
- introduce `cilium_helm_show_commands` variable (see README for more information)
- introduce `cilium_template_output_directory` variable (see README for more information)
- introduce `Molecule` tests
- don't use `shell` module anymore to execute `helm` command. Now `kubernetes.core.helm*` modules are used.
- use two underscores for internal variables
- ansible-lint: fix various issues like using FQDN Ansible module names now
- add `requirements.yml`
- add `collections.yml`

