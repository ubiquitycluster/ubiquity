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

set -e
set -x

installer_cmd=""
package_type=""
sudo_cmd="sudo -EH"

#determine os type and installer
type yum &> /dev/null && installer_cmd="yum install -y" package_type='rpm'
# prefer dnf if both yum and dnf are installed e.g. fedora
type dnf &> /dev/null && installer_cmd="dnf install -y" package_type='rpm'

type apt &> /dev/null && installer_cmd="apt install -y" package_type='deb'

# common packages
packages="debootstrap curl wget python3-setuptools"

# add distro specifc packages here
if [ 'deb' == "${package_type}" ]; then
    packages="$packages qemu-utils kpartx"
else
    packages="$packages e2fsprogs xfsprogs qemu-img dosfstools kpartx gdisk podman-docker"
fi

#install basic dependecies
$sudo_cmd $installer_cmd $packages

#set up diskimage builder submodule
#git submodule init
#git submodule update

$sudo_cmd pip3.9 install diskimage-builder
