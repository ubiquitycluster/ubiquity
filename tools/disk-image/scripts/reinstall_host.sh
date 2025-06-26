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

BMHOST=$1

if [ -z "${BMHOST}" ] ; then
    echo "Usage: reinstall_host.sh <BareMetalHost Name>"
    exit 1
fi

OSIMG=$(kubectl -n metal-nodes get bmh -o custom-columns=OSIMG:.spec.image.url ${BMHOST} | tail -1)
echo "Current OS image is $OSIMG"

kubectl patch bmh "${BMHOST}" -n metal-nodes --type merge \
    -p '{"spec":{"image":{"url":"http://change.me"}}}'

sleep 2

kubectl patch bmh "${BMHOST}" -n metal-nodes --type merge \
     -p '{"spec":{"image":{"url":"'${OSIMG}'"}}}'

OSIMG=$(kubectl -n metal-nodes get bmh -o custom-columns=OSIMG:.spec.image.url ${BMHOST} | tail -1)
echo "Updated OS image is $OSIMG"