#!/bin/sh
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

node=${1}
if [ -n "${node}" ]; then
    nodeSelector='"nodeSelector": { "kubernetes.io/hostname": "'${node:?}'" },'
else
    nodeSelector=""
fi
set -x
 ./artifacts/k3s-binary-v1.20.5+k3s1 kubectl --kubeconfig ./artifacts/k3s-kube-config --as=system:serviceaccount:unprivileged-user:fake-user -n unprivileged-user run ${USER+${USER}-}sudo --restart=Never -it  --image overriden --overrides '
{
  "spec": {
    "hostPID": true,
    "hostNetwork": true,
    '"${nodeSelector?}"'
    "containers": [
      {
        "name": "busybox",
        "image": "alpine:3.7",
        "command": ["nsenter", "--mount=/proc/1/ns/mnt", "--", "sh", "-c", "hostname sudo--$(cat /etc/hostname); exec /bin/bash"],
        "stdin": true,
        "tty": true,
        "resources": {"requests": {"cpu": "10m"}},
        "securityContext": {
          "privileged": true
        }
      }
    ]
  }
}' --rm --attach
