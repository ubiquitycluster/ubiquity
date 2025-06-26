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
output "user_data" {
  value     = local.user_data
  sensitive = true
}

output "rsa_hostkeys" {
  value = { for x, values in var.instances : x => tls_private_key.rsa_hostkeys[values["prefix"]].public_key_openssh }
}

output "master_key" {
  value = local.master_key
}

output "ed25519_hostkeys" {
  value = { for x, values in var.instances : x => tls_private_key.ed25519_hostkeys[values["prefix"]].public_key_openssh }
}

output "ssh_key" {
  value = local.ssh_key
}