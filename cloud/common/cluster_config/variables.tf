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

variable "instances" { }
variable "volume_devices" { }
variable "sudoer_username" { }
variable "cluster_name" { }
variable "domain_name" { }
variable "guest_passwd" { }
variable "nb_users" { }
variable "software_stack" { }
variable "cloud_provider" { }
variable "cloud_region" { }
variable "tf_ssh_key" { }
variable "ansible_vars" { }
variable "filesystems" { }
variable "public_keys" { }
variable "master_key" { }