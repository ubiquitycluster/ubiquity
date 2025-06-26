# Getting started

Installing Ubiquity is a relatively simple and straightforward process. That is the entire goal - To make it easy to install and use.

## Pick method of installation

There are 3 supported methods to install and run Ubiquity:

- [Using laptop/workstation in a standalone sandbox mode](admin-guide/deployment/sandbox.md). Fastest stand-up time, but runs on a single system. Should be considered "for evaluation only". Lacks longhorn storage.
- [Using a cloud provider multi-node](admin-guide/deployment/cloud/index.md). For deploying on a cloud provider.
- [Using on-premises hardware, multi-node](admin-guide/deployment/on-prem.md). For deploying on-premises.

## Configure Ubiquity

You can tune Ubiquity to your needs after initial deployment. These include:

- Configuring [identity providers](admin-guide/identities/summary.md). Ubiquity supports a range of OIDC and SAML based IdPs.
- Configuring [storage providers](admin-guide/storage/summary.md). Ubiquity supports a range of storage providers.
- Configuring [platform providers](admin-guide/providers/summary.md). Ubiquity supports a range of platform providers.
- [Adding catalogue items](admin-guide/providers/adding-a-catalogue-app.md) for sharing amongst Ubiquity users as a self-serve function.

## Access Ubiquity

You are done! Simply access using the [user-guide](user-guide/index.md)! 

If you are happy and want to support the project, make sure you check the [support page](about/support.md).

!!! danger
    Before going into production, make sure you have completed
    the [go-live checklist](checklist-for-production.md).