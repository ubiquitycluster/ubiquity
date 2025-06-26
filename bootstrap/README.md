# Bootstrap
This directory contains the bootstrap code for ArgoCD. The bootstrap code is responsible for creating the initial ArgoCD installation, including the `argocd-initial-admin-secret` secret, the `argocd` namespace, and the `argocd-server` deployment.

A Helm chart to simplify creation of ArgoCD AppProject and Applications for easier App of Apps deployments via one or more `values.yaml`.

## Overview

This chart simplifies the creation of a common bootstrapping and gitops maintenance pattern in ArgoCD called the app of apps pattern.  Please ensure you have a thorough understand of the [app of apps pattern](#app-of-apps-pattern) before continuing.

This chart leverages the concept of tenants throughout.

A tenant is a single ArgoCD `AppProject` that contains one or more ArgoCD `Applications`.  A tenant can deploy to the parent cluster (default), or to other clusters connected to ArgoCD in the case of a multi-tenant management cluster.  Both of these patterns are equally acceptible and depend heavily on the use case.

This relationship is illustrated below with the default tenant for all clusters, the `admin`:

```bash
                                        +-> Admin Application 0
+--------------+     +---------------+  |
| admin tenant | --> | admin project | -+-> Admin Application 1
+--------------+     +---------------+  |
                                        +-> Admin Application n
```

Although tenants are not a formal term within ArgoCD like `AppProjects` and `Applications`, they allow an encapsulation of a project/applications within a cluster using familiar terminology, and are extremely handy when talking about multi-tenancy.

For every tenant, the associating `AppProject` provides a host of options for RBAC and ABAC, many of which are re-used between projects or tenants within the same cluster.

This helm chart takes these concepts of applications, projects, and tenants, and provides an easy, robust, and declarative way to define clusters that follow a similar gitops app of apps approach.

## Example

The following example maps a `values.yaml` file to it's respective architecture to demonstrate the available configurations:

```bash


```

## Recommended Folder Structure

The recommended folder structure for a `*-bootstrap` repository in order to house `argocd-bootstrap` and any other custom applications is as follows:

```bash
*-bootstrap
|-- bootstrap
    |-- this chart loaded via cloning this repository. Applied at make time, enforced thereafter by ArgoCD.`
|-- apps
    |-- applications unique to `*-bootstrap`
|-- integrations
    |-- any integrations not tied directly to a single application
```

## Single Tenancy Example

The following example is the most common use case, where ArgoCD is deployed within the destination cluster, and all `Applications` are deployed in the same cluster.

For this example, we'll demonstrate bootstrapping an application from an external helm chart and a custom kustomize application from the bootstrap repo.

Assume a repository with the following folder structure:

```bash
st-bootstrap
|-- bootstrap
|-- apps
    |-- echoserver
        |-- kustomization.yaml
        |-- ...
|-- dev-umbrella.yaml
|-- values.yaml
```

Like all bootstraps, there exists a single environment specific umbrella application resource.

```yaml
# dev-umbrella.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: umbrella
  namespace: argocd
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: admin
  source:
    repoURL: https://sample.repo.git
    targetRevision: dev
    path: bootstrap/

    helm:
      releaseName: umbrella
      parameters:
        - name: repoURL
          value: $ARGOCD_APP_SOURCE_REPO_URL
        - name: targetRevision
          value: $ARGOCD_APP_SOURCE_TARGET_REVISION
        - name: env
          value: dev
      valueFiles:
        - ../values.yaml

  # Destination cluster and namespace to deploy the application
  destination:
    server: https://kubernetes.default.svc
    namespace: argocd

  # Sync policy
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

The above `dev-umbrella` is deployed to the cluster that has ArgoCD set up according to [best practices](#best-practices-for-deploying-argocd).

Once deployed, ArgoCD will attempt to reconcile the `Application` using the `bootstrap/` helm chart with the following `values.yaml`.

```yaml
# values.yaml
tenants:
  admin:
    create: false
    apps:
      guestbook:
        repoURL: https://github.com/argoproj/argocd-example-apps/tree/master/kustomize-guestbook
        path: kustomize-guestbook/
        targetRevision: HEAD
        wave: -4
      echoserver:
        path: apps/echoserver
        wave: 4
```

The `values.yaml` provides a simple interface to a tailored app of apps approach for defining multiple `Applications` that can be sourced through a variety of repositoriy combinations.  It acts as a simple, declarative view of _what_ (repoURL, path, targetRevision) is deployed into an ArgoCD managed cluster, as well as _when_ things are deployed (sync waves).

### App of Apps pattern

Since `Applications` simply encapsulate and organize other manifests, and are themselves just manifests, a common pattern in ArgoCD is to have parent `Applications` managing individual or groups of child `Applications`.  By leveraging this concept, along with `sync-waves`, ordered bootstrapping of sets of logical application deployments become trivial to orchestrate, and declaratively defined.

The diagram below shows a basic single layer inheritance hierarchy, where a single "umbrella app" can be used to represent a sets of "child" applications:

```bash
# umbrella syncs Child A and B, which subsequently deploy manifests
                  +--------------------------+
              +-> | Child A : sync-wave : -2 | --> App A Manifests
+----------+  |   +--------------------------+
| umbrella | -+
+----------+  |   +--------------------------+
              +-> | Child B : sync-wave : 4  | --> App B Manifests
                  +--------------------------+
# Child A will deploy first because -2 is before 4, sync waves range from -5 <= sync-wave <= 5
```

As discussed earlier, what makes ArgoCD cool is that the `Application` inheritance can chain to a deeply nested hierarchy (someone should flex how deep this goes).  An `Application` can generically encapsulate an actual deployable application (such as [Rancher](https://github.com/rancher/rancher)), a set of applications represented by multiple `Applications`, or both!

By leveraging this hierarchical `Application` concept further, we can build a more complex relationship where a cluster consists of multiple sets of applications, and is defined by nothing more than it's pointer to other sets of applications.

```bash
                                    +-------------------+
                                +-> | Grandchild App AA | --> Grandchild App AA Manifests
                 +-----------+  |   +-------------------+
             +-> | App Set A | -+
             |   +-----------+  |   +-------------------+
             |                  +-> | Grandchild App AB | --> Grandchild App AB Manifests
             |                      +-------------------+
             |
+---------+  |   +-----------+     +-------------------+
| cluster | -+-> | App Set B | --> | Grandchild App BA | --> Grandchild App BA Manifests
+---------+  |   +-----------+     +-------------------+
             |
             |
             |           +-------------+
             +---------> | Child App A | --> Child App A Manifests
                         +-------------+
```

In the above example, a "cluster" is represented simply by 3 top level `Application` CRDs, which can either be in the same repository, or external repositories.  Based off the levels of inheritance, complex relationships and tenancies can be declaratively modeled and independently managed.

### Best Practices for Deploying ArgoCD

ArgoCD offers both a `helm` and `kustomize` approach to installation, both fo which have their own strengths and weaknesses.