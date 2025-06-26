# Updating documentation (this website)

This project uses the [Diátaxis](https://diataxis.fr) technical documentation framework.
The website is generated using [Material for MkDocs](https://squidfunk.github.io/mkdocs-material) and can be viewed at [docs.ubiquitycluster.uk](https://docs.ubiquitycluster.uk).

There are 4 main parts:

- [Getting started (tutorials)](https://diataxis.fr/tutorials): learning-oriented
- [Concepts (explanation)](https://diataxis.fr/explanation): understanding-oriented
- [How-to guides](https://diataxis.fr/how-to-guides): goal-oriented
- [Reference](https://diataxis.fr/reference): information-oriented

## Local development

To edit and view locally, run:

```sh
make docs
```

Then visit [localhost:8000](http://localhost:8000)

## Deployment

It's running on my other cluster in the [cjcshadowsan/ubiquity](https://github.com/cjcshadowsan/ubiquity) project
(so if Devin goes down everyone can still read the documentation).

<!-- TODO -->
<!-- This website is running in both my own cluster and on my other cluster in the [cjcshadowsan/ubiquity](https://github.com/cjcshadowsan/ubiquity) project (both in `apps/ubiquity-docs`), -->
<!-- with manual DNS switch over in case I want to rebuild either of them (this is the most cost effective way to do this that I can think of). -->

<!-- You don't have to do this, you can host it on 1 cluster just fine. -->
<!-- But for 0.000000000001% of you who have 2 clusters like me, here's how to switch between them: -->

<!-- - Add the following annotation to the Ingress on the new cluster: `TODO` -->
<!-- - Go to DNS config on Cloudflare dashboard -->
<!-- - Find the TXT record for `docs.ubiquitycluster.uk` and switch the `ownerID` between `ubiquity` and `cjcshadowsan` -->
<!-- - Wait for the matching CNAME or A record to change -->
<!-- - Check if you can still access the website -->
<!-- - Do what ever you want to do -->
<!-- - (Optional) Switch back to the previous cluster -->