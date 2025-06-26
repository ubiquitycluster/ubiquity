# Back up and restore K8s objects in an on-premises environment

This tutorial describes how you can use Velero, an open-source tool, to back up and restore Kubernetes cluster resources and persistent volumes in an on-premises environment. This is helpful when you destroyed some Kubernetes resources for whatever reason, for example, when you delete the suite namespace accidentally. 

**Note** This tool doesn't back up database data and NFS data. 

## Export an NFS directory
On your NFS server, export one NFS directory.

Create one directory under the base directory. For example, if your existing NFS directories share the base directory /var/vols/itom, run the following command:

```bash
mkdir -p  /var/vols/itom/minio
```

Change the permission of the directory:

```bash
chmod -R 755 /var/vols/itom/minio
```

Change the ownership of the directory (change 1999:1999 to your own UID:GID if you use custom values):

```bash
chown -R 1999:1999 /var/vols/itom/minio
```

In the /etc/exports file, export the NFS directory by adding one line (change 1999 to your own UID or GID if you use a custom value for them): 

`/var/vols/itom/minio *(rw,sync,anonuid=1999,anongid=1999,root_squash)`

Run the following command:

```bash
exportfs -ra
```

Run the following command to check that the directory is exported:

```bash
showmount -e | grep minio
```

## Download the minio images
If your control plane nodes (formerly known as "master nodes") have Internet access, download the image from a control plane node; otherwise, download the image from another Linux machine that has Internet access and then transfer the image to the control plane node.

On the download machine, navigate to a directory where you want to download the images, and then run the following commands :

```bash
docker pull minio/minio:latest 
docker pull  minio/mc:latest
```

If the control plane node has no Internet access, transfer the images to the control plane node. 

## Obtain the image IDs.
Run the following command: 

```bash
docker images |grep minio
```

In the output, find the IDs of the images. In the following example, it's `8dbf9ff992d5`.

```bash
docker.io/minio/minio latest 8dbf9ff992d5 30 hours ago 183 MB
```

Run the following command to tag one image:

```bash
docker tag <image ID> <image registry URL>/<organization name>/minio:<tag>
```

The following are two examples:

```bash
docker tag 8dbf9ff992d5 myregistry.azurecr.io/sandbox/minio:test
docker tag 8dbf9ff992d5 localhost:5000/hpeswitom/minio:test
```

```
<image ID>: the image ID you obtained in the previous step.

<image registry URL>/<organizaition name>: your image registry URL/organization name. If using the local registry, it's localhost:5000/hpeswitom; if using an external registry, ask your registry administrator for it.

<tag>: specify any value you like. 
```

Repeat the step above to tag the other image (minio/mc:latest) into your image registry. 

## Push the images into your image registry:

```bash
docker push <image registry URL>/<organization name>/minio:<tag>
docker push <image registry URL>/<organization name>/mc:<tag>

# Example 1:

docker push myregistry.azurecr.io/sandbox/minio:test
docker push myregistry.azurecr.io/sandbox/mc:test

# Example 2:

docker push localhost:5000/hpeswitom/minio:test
docker push localhost:5000/hpeswitom/mc:test
```

## Install Velero on a control plane node
Perform the following steps to install Velero on the control plane node.

Download the tarball of the latest Velero release to a temporary directory on the control plane node. The download URL is https://github.com/vmware-tanzu/velero/releases/.
Extract the package:

```bash
tar -xvf <release-tarball-name>.tar.gz
```

The directory you extracted is called the “Velero directory” in subsequent steps. 

Move the Velero binary from the Velero directory to somewhere in your PATH. For example:

```bash
cp velero /usr/local/bin/
```

Create a Velero-specific credentials file in your local directory. For example, in the Velero directory:

```bash
cat <<ENDFILE > ./credentials-velero
[default]
aws_access_key_id = minio
aws_secret_access_key = minio123
ENDFILE
```

Navigate to the Velero directory, and create a backup copy of the examples/minio/00-minio-deployment.yaml file. This is because in the next steps, you will need to edit this file. 
Edit the `00-minio-deployment.yaml` file as follows:

Add PVs/PVCs to the `00-minio-deployment.yaml` file by appending the following code lines to the end of the file (replace $minio-NFS-path and $NFS-server-FQDN with the exported NFS folder path and hostname of the NFS server):

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: minio-pv-claim
  namespace: velero
spec:
  capacity:
    storage: 10Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteMany
  persistentVolumeReclaimPolicy: Retain
  nfs:
    path: $minio-NFS-path
    server: $NFS-server-FQDN
   
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: minio-pv-claim
  namespace: velero
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 10Gi
  volumeName: minio-pv-claim
```

In the Deployment section, make the following change:

From:

```yaml
volumes:
      - name: storage
        emptyDir: {}
      - name: config
        emptyDir: {}
```

To:

```yaml
volumes:
      - name: storage
        persistentVolumeClaim:
          claimName: minio-pv-claim
```

Remove the last two lines in the Deployment section below:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: velero
  name: minio
  labels:
    component: minio
spec:
  strategy:
    type: Recreate
  selector:
    matchLabels:
      component: minio
  template:
    metadata:
      labels:
        component: minio
    spec:
      volumes:
      - name: storage
        emptyDir: {}
      - name: config
        emptyDir: {}
      containers:
      - name: minio
        image: minio/minio:latest
        imagePullPolicy: IfNotPresent
        args:
        - server
        - /storage
        - --config-dir=/config
        env:
        - name: MINIO_ACCESS_KEY
          value: "minio"
        - name: MINIO_SECRET_KEY
          value: "minio123"
        ports:
        - containerPort: 9000
        volumeMounts:
        - name: storage
          mountPath: "/storage"
        - name: config
          mountPath: "/config"
```

Replace the image values in the following lines with the images in your image registry:

```yaml
image: minio/minio:latest 
image: minio/mc:latest 
```

For example, change them as shown below (this example uses an external registry):

```yaml
image: myregistry.azurecr.io/sandbox/minio:test 
image: myregistry.azurecr.io/sandbox/mc:test
```

Run the following commands:

```bash
kubectl apply -f examples/minio/00-minio-deployment.yaml
velero install \
    --provider aws \
    --plugins velero/velero-plugin-for-aws:v1.0.0 \
    --bucket velero \
    --secret-file ./credentials-velero \
    --use-volume-snapshots=false \
    --backup-location-config region=minio,s3ForcePathStyle="true",s3Url=http://minio.velero.svc:9000
```

## Back up and restore k8s objects
Perform the following steps by using Velero.

### Backup
To back up objects, run the following command on the control plane node on which Velero is installed:

```bash
velero backup create <backup name> --include-namespaces <namespace> --wait
```

The following are examples to back up objects for the suite and CDF, respectively:

```bash
velero backup create itsma-backup --include-namespaces itsma-fghnd --wait
velero backup create core-backup --include-namespaces core --wait
```

### Restore
The following procedures assume that you have removed the suite namespace or the CDF core namespace. To test the procedure, you can run kubectl delete ns <namespace> to delete a namespace.

#### Restore the objects for the suite
Perform the following steps:

Once the suite namespace is deleted, the associated PVs are released. Before the restore, delete the released PVs by running the following command on the control plane node:

```bash
kubectl get pv |grep -i release | awk '{print $1}' | xargs kubectl delete pv
```

Run the following command to restore:

```bash
velero restore create --from-backup <suite backup name> --wait
```

For example:

```bash
velero restore create --from-backup itsma-backup --wait
```

#### Change the nodePort in itom-nginx-ingress-svc to 443:
Run the following command:

```bash
kubectl edit svc itom-nginx-ingress-svc -n <suite namespace>
```

#### Change nodePort to 443 as shown below:

```yaml
- name: https-port
    nodePort: 443
    port: 443
    protocol: TCP
    targetPort: 443
```

#### Restore the objects for CDF
Delete the PVs associated with the core namespace:

```bash
kubectl delete pv db-single itom-logging itom-vol
```

Restore with Velero:

```bash
velero restore create --from-backup <core backup> –wait
```

For example:

```bash
velero restore create --from-backup core-backup –wait
```
