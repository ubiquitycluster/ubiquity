#!/bin/bash
# Copyright The Ubiquity Authors.
#
# Licensed under the Apache License, Version 2.0. Previously licensed under the Functional Source License (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/ubiquitycluster/ubiquity/blob/main/LICENSE
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# This software was previously licensed under the Functional Source License but has now transitioned to an Apache 2.0 License
# as of June 2025.
# See the License for the specific language governing permissions and
# limitations under the License.

echo "Basename: [$(basename "$0")]"
echo "executed: $0"
echo "Dirname: [$(dirname "$0")]"
echo "pwd: [$(pwd)]"

if [[ "$PWD" = *ubiquity/scripts/bmo ]]; then

  yum -y install httpd-tools gettext

  kubectl create ns baremetal-operator-system

  patch -u -b ../../baremetal-operator/ironic-deployment/base/ironic.yaml config/ironic.yaml.patch

  read -p "Please press y to edit your configurations accordingly." CONT
  if [ "$CONT" = "y" ]; then
    vi ../../baremetal-operator/ironic-deployment/components/tls/certificate.yaml
    vi config/ironic_bmo_configmap.env
    vi config/ironic.yaml.patch
    vi storage/longhorn/pvc-longhorn.yaml
  else
    echo "chickening out, nobody wants to run PXE everywhere...";
    exit 1
  fi

  kubectl -n baremetal-operator-system apply -f storage/longhorn/pvc-longhorn.yaml
  cp -rp config/ironic_bmo_configmap.env ../../baremetal-operator/ironic-deployment/components/keepalived/ironic_bmo_configmap.env
  cp -rp config/ironic_bmo_configmap.env ../../baremetal-operator/ironic-deployment/default/ironic_bmo_configmap.env
  cp -rp config/ironic_bmo_configmap.env ../../baremetal-operator/config/default/ironic.env
  cp -rp config/deploy.sh ../../baremetal-operator/tools/deploy.sh

  cd ../../baremetal-operator/
  cd config/base/crds/
  kustomize build | kubectl -ns baremetal-operator-system apply -f -
  cd -

  read -p "Please enter the ironic host name (ex: ironic.cluster.local): " PROMPT_IRONIC_HOST
  read -p "Please enter the ironic host IP you would like to listen on: " PROMPT_IRONIC_HOST_IP
  read -p "Please enter the ironic username you would like to use: " PROMPT_IRONIC_USERNAME
  read -p "Please enter the ironic password you would like to use for ${PROMPT_IRONIC_USERNAME}: " PROMPT_IRONIC_PASSWORD
  read -p "Please enger the ironic inspector username you would like to use: " PRONPT_IRONIC_INSPECTOR_USERNAME
  read -p "Please enter the ironic inspector password you would like to use for ${PROMPT_IRONIC_INSPECTOR_USERNAME}: " PROMPT_IRONIC_INSPECTOR_PASSWORD

  IRONIC_HOST=$PROMPT_IRONIC_HOST IRONIC_HOST_IP=$PROMPT_IRONIC_HOST_IP IRONIC_USERNAME=$PROMPT_IRONIC_USERNAME IRONIC_PASSWORD=$PROMPT_IRONIC_PASSWORD IRONIC_INSPECTOR_USERNAME=$PROMPT_IRONIC_INSPECTOR_USERNAME IRONIC_INSPECTOR_PASSWORD=$PROMPT_IRONIC_INSPECTOR_PASSWORD USER=root ./tools/deploy.sh -b -i -t -m -k
  #IRONIC_HOST=ironic.cluster.local IRONIC_HOST_IP=10.199.251.120 USER=root ./tools/deploy.sh  -b -i -t -m -k
else
  echo "You're in the wrong directory! run this from ubiquity/scripts/bmo for it to work properly"
fi
