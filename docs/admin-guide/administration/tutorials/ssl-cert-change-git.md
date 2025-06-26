# How to Change an SSL cert for your upstream git repository

## Situation
Your git repo that you host your cluster configuration on needs to have an SSL certificate change. As all of the GitOps nature of the project requires that ArgoCD can read a git repo, it needs to be aware that an SSL certificate is needed - And if it's a self-cert SSL key then it needs to be created in the right format and applied.

### Task
This article provides instructions for safely applying a key to the Ubiquity ArgoCD instance.


## Requirements
- A Ubiquity cluster
- A (valid) SSL key for your git repo domain (see next for detail)

## Background
If you have a need to replace a key to your git infrastructure that is read by a Ubiquity cluster, this guide will provide steps in the proper order to ensure a painless process.

Please ensure you complete a backup of your old SSL key before continuing this process.

## Solution
The following steps should be undertaken:

### Create a valid SSL key
An extension to the x509 standard, the `subjectAltName` is used by ArgoCD to validate that a server matches the host. Please ensure this is present. In this example, we are creating a self-signed certificate using an existing CSR. For more information on creating an SSL certificate, CSRs etc please see the official OpenSSL documentation.

N.B. it is important to note that you need an up-to-date version of OpenSSL. Version 1.1.1 or greater is appropriate as of this guide.

To generate the key in this example (an example of updating a gitlab cert), we specify the `-addext` functionality present inside OpenSSL 1.1.1:

```
openssl req -new -nodes -x509 -subj "/C=UK/ST=Example/L=Example/O=ExampleOrg/CN=example.com" -days 3650 -keyout /etc/gitlab/ssl/gitlab.key -out /etc/gitlab/ssl/gitlab.crt -extensions v3_ca -addext "subjectAltName = DNS:example.com"
```

This process will give you a certificate with a subjectAltName field defined. You can verify this using:

```
openssl x509 -noout -text -in /etc/gitlab/ssl/gitlab.crt | grep DNS:
```

And you should see your DNS for example.com in that DNS field.

### Applying the key
Once it is applied to GitLab (other git repository hosting is available), then from there you only need to update the yaml in the git repository, and apply the new certificate live. To do this:

Edit the `bootstrap/argocd/argocd-<repo>-tls-cert.yaml` file - This file has a copy of the pem present that is stored in a configmap:

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-tls-certs-cm
  namespace: argocd
  labels:
    app.kubernetes.io/name: argocd-cm
    app.kubernetes.io/part-of: argocd
data:
  example.com: |
    -----BEGIN CERTIFICATE-----
    <cert information>
    -----END CERTIFICATE-----
```

Edit this and put the appropriate key in-place for the correct site that you generated.

Then:

```
git commit -am "Updated TLS cert for git repository"
git push <remote that you use, if origin you can simply use git push>
```

Once this is completed, to help it take effect faster you can then apply that configmap straight away. 

### OPUS actions

To apply straight away, start up an OPUS container, go into your `bootstrap/argocd/` directory and apply the manifest manually:

```
kubectl -n argocd apply -f argocd-<repo>-tls-cert.yaml
```

This will apply live on the system.

Then the final stage is to restart the applicationset controller in order for it to re-read all of the git repositories (we made our change to the git repo before this to prevent argocd reverting the live change to a configmap we just did). Using K9s, restarting a pod is a simple process. We go into the pods view, filter for argocd-applicationset-controller, and then ctrl-d to delete the pod - Because a statefulset exists for it, it will simply restart the pod:

```
k9s

:pods

/argocd-applicationset-controller

ctrl-d
```

Once the pod restarts, you can then also go and tail the logs of that pod by pressing the `l` hotkey.