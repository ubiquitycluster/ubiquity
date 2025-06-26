#!/bin/bash
# Copyright 2023 Logicalis UKI. All Rights Reserved.
#
# Licensed under the Functional Source License, Version 1.0, Apache 2.0 Change License (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/logicalisuki/ubiquity/blob/main/LICENSE.md
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# It also allows for the transition of this software to an Apache 2.0 Licence
# on the second anniversary of the date we make the software available.
# See the License for the specific language governing permissions and
# limitations under the License.

# shellcheck disable=SC1091
#source lib/common.sh
# shellcheck disable=SC1091
#source lib/network.sh
# shellcheck disable=SC1091
#source lib/images.sh

IMAGE=$1
IMAGE_NAME=$(basename ${IMAGE})

if [ -z "${IMAGE}" ] ; then
    echo "Usage: cp-disk-image.sh [image-name - full path]"
    exit 1
fi

if echo "${IMAGE_NAME}" | grep -qi centos 2>/dev/null ; then
    OS_TYPE=centos
else
    OS_TYPE=unknown
fi

kubectl cp -n baremetal-operator-system ${IMAGE} $(kubectl -n baremetal-operator-system get pods | grep ironic | awk '{print $1}'):/shared/html/images/${IMAGE_NAME} -c ironic-httpd

kubectl cp -n baremetal-operator-system ${IMAGE}.md5 $(kubectl -n baremetal-operator-system get pods | grep ironic | awk '{print $1}'):/shared/html/images/${IMAGE_NAME}.md5sum -c ironic-httpd
