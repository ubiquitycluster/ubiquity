# ArgoCD

# ArgoCD

## Scope

The `argocd` role installs [ArgoCD](https://argo-cd.readthedocs.io/), a GitOps
tool, in the cluster.

## Accessing ArgoCD

The ArgoCD `"admin"` password is available from a secret:

```console
$ ARGOCD_INITIAL_ADMIN_PASSWORD="$(kubectl get secrets -n argocd argocd-initial-admin-secret -ojson | jq .data.password -r | base64 -d)"
$ echo $ARGOCD_INITIAL_ADMIN_PASSWORD
AbCdEf
```

This password can be used for full-access on the ArgoCD web interface,
or for the CLI in opus:

```console
$ argocd login
Username: admin
Password:
'admin:login' logged in successfully
Context 'port-forward' updated
$ argocd app list
[...]
```

## ArgoCD application

Do not call this directly, use [GitOps application instead](gitops#gitops-application).
