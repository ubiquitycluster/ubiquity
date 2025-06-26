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

# Install cloud-init

rm -f /etc/dhcp/dhclient.d/google_hostname.sh
if [ ! -f /etc/cloud/cloud-init.disabled ]; then
    if  [ ! -f /usr/bin/cloud-init ]; then
        # Try to install cloud-init every 5 seconds for 12 times.
        for i in $(seq 12); do
            yum -y install cloud-init && break
            sleep 5
        done
    fi
    # Verify installation was successful
    if  [ -f /usr/bin/cloud-init ]; then
        systemctl disable cloud-init
        touch /etc/cloud/cloud-init.disabled
        cloud-init clean --logs
        cloud-init init --local
        cloud-init init
        cloud-init modules --mode=config
        cloud-init modules --mode=final
    else
        echo "Problem installing cloud-init. Verify network connectivity and reboot."
    fi
elif [ -f /usr/bin/cloud-init ]; then
    yum -y remove cloud-init
fi