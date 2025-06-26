#!/usr/bin/env bash
set -eux -o pipefail

#yum -y install golang sshpass

# shellcheck disable=SC1091
#source lib/logging.sh
# shellcheck disable=SC1091
#source lib/common.sh

eval "$(go env)"
export GOPATH
DEPLOYDIR="$(dirname "$PWD")"
BMODIR=$DEPLOYDIR/metal3/scripts/bmo
OUTDIR=$DEPLOYDIR/scripts/output

# Environment variables
# M3PATH : Path to clone the metal3 dev env repo
# BMOPATH : Path to clone the baremetal operator repo
#
# BMOREPO : Baremetal operator repository URL
# BMOBRANCH : Baremetal operator repository branch to checkout
# FORCE_REPO_UPDATE : discard existing directories
#
# BMO_RUN_LOCAL : run the baremetal operator locally (not in Kubernetes cluster)

NODES_FILE=nodelist
domain=ubiquitycluster.local
M3PATH="${GOPATH}/src/github.com/metal3-io"
#BMOPATH="${M3PATH}/baremetal-operator"
BMOPATH="${PWD}/../../../baremetal-operator"
BMOREPO="${BMOREPO:-https://github.com/metal3-io/baremetal-operator.git}"
BMOBRANCH="${BMOBRANCH:-10eb5aa3e614d0fdc6315026ebab061cbae6b929}"
FORCE_REPO_UPDATE="${FORCE_REPO_UPDATE:-true}"

AWX_SITE=http://10.1.0.85
AWX_TOKEN="ZrUPBf5eQuIlcAl9tquwVHnHVO6Hki"
BMO_RUN_LOCAL="${BMO_RUN_LOCAL:-false}"
COMPUTE_NODE_PASSWORD="${COMPUTE_NODE_PASSWORD:-mypasswd}"
BM_IMAGE=${BM_IMAGE:-"compute-rl8.qcow2"}
IMAGE_URL=http://10.1.3.200:6180/images/${BM_IMAGE}
IMAGE_CHECKSUM=http://10.1.3.200:6180/images/${BM_IMAGE}.md5sum

#NODETYPE=xcc

function clone_repos {
    mkdir -p "${M3PATH}"
    if [[ -d ${BMOPATH} && "${FORCE_REPO_UPDATE}" == "true" ]]; then
      rm -rf "${BMOPATH}"
    fi
    if [ ! -d "${BMOPATH}" ] ; then
        pushd "${M3PATH}"
        git clone "${BMOREPO}"
        popd
    fi
    pushd "${BMOPATH}"
    git checkout "${BMOBRANCH}"
    git pull -r || true
    popd
}

function launch_baremetal_operator {
    docker pull $IRONIC_BAREMETAL_IMAGE
    kubectl apply -f $BMODIR/namespace/namespace.yaml
    kubectl apply -f $BMODIR/rbac/service_account.yaml -n metal3
    kubectl apply -f $BMODIR/rbac/role.yaml -n metal3
    kubectl apply -f $BMODIR/rbac/role_binding.yaml
    kubectl apply -f $BMODIR/crds/metal3.io_baremetalhosts_crd.yaml
    kubectl apply -f $BMODIR/operator/no_ironic/operator.yaml -n metal3
}

# documentation for the values below may be found at
# https://cloudinit.readthedocs.io/en/latest/topics/modules.html
create_userdata() {
    name="$1"
    COMPUTE_NODE_FQDN="$name.$domain"
    printf "## template: jinja\n" > $name-userdata.yaml
    printf "#cloud-config\n" >> $name-userdata.yaml
    if [ -n "$COMPUTE_NODE_PASSWORD" ]; then
        printf "password: ""%s" "$COMPUTE_NODE_PASSWORD" >>  $name-userdata.yaml
        printf "\nchpasswd: {expire: False}\n" >>  $name-userdata.yaml
        printf "ssh_pwauth: True\n" >>  $name-userdata.yaml
    fi

    if [ -n "$COMPUTE_NODE_FQDN" ]; then
        printf "fqdn: ""%s" "$COMPUTE_NODE_FQDN" >>  $name-userdata.yaml
        printf "\n" >>  $name-userdata.yaml
    fi
    printf "disable_root: false\n" >> $name-userdata.yaml
    printf "ssh_authorized_keys:\n  - " >> $name-userdata.yaml

    if [ ! -f $HOME/.ssh/id_rsa.pub ]; then
        yes y | ssh-keygen -t rsa -N "" -f $HOME/.ssh/id_rsa
    fi

    cat $HOME/.ssh/id_rsa.pub >> $name-userdata.yaml
    printf "  - " >> $name-userdata.yaml
    cat $HOME/.ssh/id_ed25519.pub >> $name-userdata.yaml
    printf "\n" >> $name-userdata.yaml
    cat >> $name-userdata.yaml << EOF
write_files:
  - path: /etc/ansible/inventory.json
    content: |
      {
        "name": "{{local_hostname}}",
        "description": "{{local_hostname}}",
        "enabled": true,
        "instance_id": "",
        "variables": "ansible_host: {{ds.network_json.networks[0].ip_address}}"
      }
    permissions: '0644'

runcmd:
{% set tower_api_token = "$AWX_TOKEN" %}
  - [curl, -k, -H, "Authorization: Bearer {{tower_api_token}}", -H, "Content-Type: application/json", -X, POST, -d, '{"name":"Ubiquity","organization":1}', "$AWX_SITE/api/v2/inventories/"]
  - sleep 5  # wait for the inventory to be created
  - [curl, -k, -H, "Authorization: Bearer {{tower_api_token}}", -H, "Content-Type: application/json", -X, POST, -d, "@/etc/ansible/inventory.json", "$AWX_SITE/api/v2/inventories/2/hosts/"]
EOF
    mv $name-userdata.yaml $OUTDIR/
}

apply_userdata_credential() {
    name="$1"
    cat <<EOF > ./$name-user-data-credential.yaml
apiVersion: v1
data:
  userData: $(base64 -w 0 $OUTDIR/$name-userdata.yaml)
kind: Secret
metadata:
  name: $name-user-data
  namespace: metal-nodes
type: Opaque
EOF
    mv $name-user-data-credential.yaml $OUTDIR/
    #kubectl apply -n metal-nodes -f $OUTDIR/$name-user-data-credential.yaml
}

prep_nodefile() {
    cp $NODES_FILE $OUTDIR/$NODES_FILE
}

prep_nodes() {
    address=${1}
    ssh-keyscan -H ${address} >> ~/.ssh/known_hosts
}

find_bm_mac_xcc() {
    name=${1}
    address=${2}
    ipmiuser=${3}
    ipmipass=${4}
    pxedev=${5}
    mac=${6}
    nodemac=$(for i in $(sshpass -p ${ipmipass} ssh -o stricthostkeychecking=no -l ${ipmiuser} ${address} 'adapter -list' | awk '{print $1}' | grep -v system); do tempmac=$(sshpass -p ${ipmipass} ssh -o stricthostkeychecking=no -l ${ipmiuser} ${address} "adapter -show $i ports" | grep -A5 "${pxedev}" | awk -F "[: ]+" '/Permanent/{print tolower(substr($NF,1))}' | fold -w2 | paste -sd: | tr -d '\r'); [[ -n $tempmac ]] && { echo "$tempmac"; break; }; done)
    sed -i "s/,${mac},/,${nodemac},/g" $OUTDIR/${NODES_FILE}
}

find_bm_mac_ilom() {
    name=${1}
    address=${2}
    ipmiuser=${3}
    ipmipass=${4}
    pxedev=${5}
    mac=${6}
nodemac=$(for i in $(sshpass -p ${ipmipass} ssh -o stricthostkeychecking=no -l ${ipmiuser} ${address} 'show /map1' | awk '/name=/{print $1}' | cut -d= -f2); do mac=$(sshpass -p ${ipmipass} ssh -o stricthostkeychecking=no -l ${ipmiuser} ${address} "show /system1/network1/Integrated_NICs1/NIC.$i" | awk -F "=" '/Permanent_MAC_Address/{print tolower($NF)}' | tr -d '\r'); [[ -n $mac ]] && { echo "$mac"; break; }; done)
    sed -i "s/,${mac},/,${nodemac},/g" $OUTDIR/${NODES_FILE}
}

find_bm_mac_idrac() {
    name=${1}
    address=${2}
    ipmiuser=${3}
    ipmipass=${4}
    pxedev=${5}
    mac=${6}
nodemac=$(sshpass -p ${ipmipass} ssh -o stricthostkeychecking=no -l ${ipmiuser} ${address} 'racadm nicstatistics' | grep "${pxedev}" | awk '/^NIC\.[A-Za-z]+\.([0-9]+(-[0-9]+)+)/{print tolower($NF)}' | tr -d '\r')
    sed -i "s/,${mac},/,${nodemac},/g" $OUTDIR/${NODES_FILE}
}

find_bm_mac_sbmc() {
    name=${1}
    address=${2}
    ipmiuser=${3}
    ipmipass=${4}
    pxedev=${5}
    mac=${6}
nodemac=$(for i in $(sshpass -p ${ipmipass} ssh -o stricthostkeychecking=no -l ${ipmiuser} ${address} 'ipmitool lan print' | awk -F ': ' '/MAC Address/{print $2}'); do mac=$(sshpass -p ${ipmipass} ssh -o stricthostkeychecking=no -l ${ipmiuser} ${address} "ipmitool lan print $i" | awk '/MAC Address/{print tolower($4)}' | tr -d '\r'); [[ -n $mac ]] && { echo "$mac"; break; }; done)
    sed -i "s/,${mac},/,${nodemac},/g" $OUTDIR/${NODES_FILE}
}

find_bm_mac_irmc() {
    name=${1}
    address=${2}
    ipmiuser=${3}
    ipmipass=${4}
    pxedev=${5}
    mac=${6}
#nodemac=$(for i in $(sshpass -p ${ipmipass} ssh -o stricthostkeychecking=no -l ${ipmiuser} ${address} 'ipmcget -d System -g NIC | grep "MAC Address" | awk -F " : " '{print $2}''); do mac=$(sshpass -p ${ipmipass} ssh -o stricthostkeychecking=no -l ${ipmiuser} ${address} "ipmcget -d System -g NIC -n $i | grep "MAC Address" | awk -F " : " '{print tolower($2)}' | tr -d '\r'); [[ -n $mac ]] && { echo "$mac"; break; }; done)
    sed -i "s/,${mac},/,${nodemac},/g" $OUTDIR/${NODES_FILE}
}

node_networkdata() {
    name=${1}
    dns1=${2}
    dns2=${3}
    net1=${4}
    ip1=${5}
    netmask1=${6}
    netname1=${7}
    gateway1=${8}
    net2=${9}
    ip2=${10}
    netmask2=${11}
    netname2=${12}
    gateway2=${13}
    ./lib/networkData.sh -hostname $name -dns1 $dns1 -dns2 $dns2 -net1 $net1 $ip1 $netmask1 $netname1 $gateway1 -net2 $net2 $ip2 $netmask2 $netname2 $gateway2
}


create_networkdata() {
    name=${1}
    dns1=${2}
    dns2=${3}
    net1=${4}
    ip1=${5}
    netmask1=${6}
    netname1=${7}
    gateway1=${8}
    net2=${9}
    ip2=${10}
    netmask2=${11}
    netname2=${12}
    gateway2=${13}
    node_networkdata $name $dns1 $dns2 $net1 $ip1 $netmask1 $netname1 $gateway1 $net2 $ip2 $netmask2 $netname2 $gateway2 > $name-networkdata.json
    mv $name-networkdata.json $OUTDIR/
}

apply_networkdata_credential() {
    name="$1"
    cat <<EOF > ./$name-network-data-credential.yaml
apiVersion: v1
data:
  networkData: $(base64 -w 0 $OUTDIR/$name-networkdata.json)
kind: Secret
metadata:
  name: $name-network-data
  namespace: metal-nodes
type: Opaque
EOF
    mv $name-network-data-credential.yaml $OUTDIR/
    #kubectl apply -n metal-nodes -f $OUTDIR/$name-network-data-credential.yaml
}

function make_bm_hosts {
exec < nodelist
read header
while read line
do
    echo $line |while IFS=',' read -r nodetype name address ipmiuser ipmipass pxedev user password mac dns1 dns2 net1 ip1 netmask1 netname1 gateway1 net2 ip2 netmask2 netname2 gateway2; do
        prep_nodes $address
        case $nodetype in
            xcc)
                find_bm_mac_xcc $name $address $ipmiuser $ipmipass $pxedev $mac
                ;;
            idrac)
               find_bm_mac_idrac $name $address $ipmiuser $ipmipass $pxedev $mac
               ;;
            ilom)
               find_bm_mac_ilom $name $address $ipmiuser $ipmipass $pxedev $mac
               ;;
            sbmc)
               find_bm_mac_sbmc $name $address $ipmiuser $ipmipass $pxedev $mac
               ;;
            irmc)
               find_bm_mac_irmc $name $address $ipmiuser $ipmipass $pxedev $mac
               ;;
            *)
               echo "Invalid BMC type. Please specify one of: xcc, idrac, ilom, sbmc, or irmc"
               exit 1
               ;;
        esac
    done
done
}

function make_nodedata {
mv $OUTDIR/$NODES_FILE $NODES_FILE
exec < $NODES_FILE
read header
while read line
do
    echo $line |while IFS=',' read -r nodetype name address ipmiuser ipmipass pxedev user password mac dns1 dns2 net1 ip1 netmask1 netname1 gateway1 net2 ip2 netmask2 netname2 gateway2; do
        create_userdata $name
        apply_userdata_credential $name
        create_networkdata $name $dns1 $dns2 $net1 $ip1 $netmask1 $netname1 $gateway1 $net2 $ip2 $netmask2 $netname2 $gateway2
        apply_networkdata_credential $name
        cd "${BMOPATH}"/cmd/make-bm-worker/
        GO111MODULE=auto go run "${BMOPATH}"/cmd/make-bm-worker/main.go \
           -address "$address" \
           -password "$ipmipass" \
           -user "$ipmiuser" \
           -boot-mac "$mac" \
           -disableCertificateVerification \
           -automatedCleaningMode "disabled" \
           -image-url "${IMAGE_URL}" \
           -image-checksum "${IMAGE_CHECKSUM}" \
           -image-checksum-type "md5" \
           -image-format "qcow2" \
           "$name" > /tmp/$name-bm-node.yaml
        cd -
        mv /tmp/$name-bm-node.yaml .
        #printf "  image:" >> $name-bm-node.yaml
        #printf "\n    url: ""%s" "${IMAGE_URL}" >> $name-bm-node.yaml
        #printf "\n    checksum: ""%s" "${IMAGE_CHECKSUM}" >> $name-bm-node.yaml
        printf "\n  userData:" >> $name-bm-node.yaml
        printf "\n    name: ""%s" "$name""-user-data" >> $name-bm-node.yaml
        printf "\n    namespace: metal-nodes" >> $name-bm-node.yaml
        printf "\n  networkData:" >> $name-bm-node.yaml
        printf "\n    name: ""%s" "$name""-network-data" >> $name-bm-node.yaml
        printf "\n    namespace: metal-nodes" >> $name-bm-node.yaml
        printf "\n  rootDeviceHints:" >> $name-bm-node.yaml
        printf "\n    minSizeGigabytes: 48\n" >> $name-bm-node.yaml
        mv $name-bm-node.yaml $OUTDIR/
        #kubectl apply -f $OUTDIR/$name-bm-node.yaml -n metal-nodes
    done
done
}

function make_namespace {
    kubectl create ns metal-nodes || true
}

function apply_bm_hosts {
     mkdir -p output
     prep_nodefile
     make_bm_hosts
     make_namespace
     make_nodedata
}

#clone_repos
#launch_baremetal_operator
apply_bm_hosts
#make_bm_hosts
