# Updating documentation (this website)

This project uses the [Diátaxis](https://diataxis.fr) technical documentation
framework. The website is generated using [Material for MkDocs](https://squidfunk.github.io/mkdocs-material) and can be viewed at
[docs.ubiquitycluster.uk](https://docs.ubiquitycluster.uk).

There are four main parts:

- [Getting started (tutorials)](https://diataxis.fr/tutorials): learning-oriented
- [Concepts (explanation)](https://diataxis.fr/explanation): understanding-oriented
- [How-to guides](https://diataxis.fr/how-to-guides): goal-oriented
- [Reference](https://diataxis.fr/reference): information-oriented

## Local development

To edit and view locally, run:

```sh
make docs
```

Then visit [localhost:8000](http://localhost:8000).

Before committing documentation changes, run the relevant generated-doc checks.
For Helm chart reference changes, use:

```sh
scripts/generate-helm-chart-reference.sh --check
```

A dry-run/local proof confirms the documentation site builds and generated
references are current. It does not prove that the public website, DNS, ingress,
or CDN path is healthy.

## Deployment

Production documentation deployment should follow the same GitOps and evidence
rules as application changes:

1. Build or preview the site locally.
2. Commit the Markdown and generated-reference changes together.
3. Let the GitOps controller reconcile the documentation application.
4. Verify the deployed workload and ingress:

   ```sh
   kubectl get applications -n argocd
   kubectl get ingress --all-namespaces | grep docs
   kubectl get certificates --all-namespaces | grep docs
   ```

5. Confirm DNS resolves to the active ingress or tunnel endpoint.
6. Open the public URL and check the changed page.

## DNS failover and rollback

If the documentation site is hosted by more than one cluster, keep the active DNS
owner explicit in the change record. For ExternalDNS-style ownership, switch the
record only after the standby site is reconciled and serving the expected
content.

Rollback options:

- revert the documentation commit and let GitOps reconcile
- switch DNS back to the previous cluster if the new cluster is unhealthy
- restore the previous generated reference file if a generator regression caused
  broken navigation

Record the rollback command, DNS record changed, and post-rollback verification.
Do not treat a local MkDocs preview as live production proof; live production
proof requires DNS, certificate, ingress, and public page evidence.
