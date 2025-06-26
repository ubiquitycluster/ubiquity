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

output "public_instances" {
  value = module.cluster_config.public_instances
}

output "public_ip" {
  value = {
    for key, values in module.cluster_config.public_instances: key => values["public_ip"]
    if values["public_ip"] != ""
  }
}

output "cluster_name" {
  value = lower(var.cluster_name)
}

output "domain" {
  value = lower(var.domain)
}

output "accounts" {
  value = {
    guests = {
      usernames =   var.nb_users != 0 ? (
        "user[${format(format("%%0%dd", length(tostring(var.nb_users))), 1)}-${var.nb_users}]"
      ) : (
        "You have chosen to create user accounts yourself (`nb_users = 0`), please read the documentation on how to manage this at https://github.com/logicalisuki/ubiquity/blob/main/docs/README.md#103-add-a-user-account"
      ),
      password = module.cluster_config.guest_passwd
    }
    sudoer = {
      username = var.sudoer_username
      password = "N/A (public ssh-key auth)"
    }
  }
}

output "master_private_key" {
  value     = module.instance_config.master_key.private
  sensitive = true
}
output "ssh_private_key" {
  value     = module.instance_config.ssh_key.private
  sensitive = true
}