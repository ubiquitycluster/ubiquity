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