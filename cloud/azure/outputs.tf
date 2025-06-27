# Copyright 2023 Logicalis UKI. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0. Previously licensed under the Functional Source License (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/logicalisuki/ubiquity-open-open/blob/main/LICENSE
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# This software was previously licensed under the Functional Source License but has now transitioned to an Apache 2.0 License
# as of June 2025.
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
        "You have chosen to create user accounts yourself (`nb_users = 0`), please read the documentation on how to manage this at https://github.com/logicalisuki/ubiquity-open/blob/main/docs/README.md#103-add-a-user-account"
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

# Standard cluster information output
output "cluster_info" {
  value = {
    cluster_name   = var.cluster_name
    cluster_type   = module.design.cluster_type
    master_count   = module.design.master_count
    domain_name    = module.design.domain_name
    cloud_provider = local.cloud_provider
    cloud_region   = local.cloud_region
  }
  description = "Standard cluster information across all providers"
}

# k3s specific outputs
output "kubeconfig_command" {
  value = length([for key, values in module.cluster_config.public_instances : key if contains(values.tags, "ansible")]) > 0 ? "ssh -i ~/.ssh/id_rsa ${var.sudoer_username}@${[for key, values in module.cluster_config.public_instances : values.public_ip if contains(values.tags, "ansible")][0]} 'sudo cat /etc/rancher/k3s/k3s.yaml'" : "No ansible server found"
  description = "Command to retrieve kubeconfig from the k3s cluster"
}

output "cluster_endpoints" {
  value = length([for key, values in module.cluster_config.public_instances : key if contains(values.tags, "ansible")]) > 0 ? {
    api_server = "https://${[for key, values in module.cluster_config.public_instances : values.public_ip if contains(values.tags, "ansible")][0]}:6443"
    dashboard  = "https://${[for key, values in module.cluster_config.public_instances : values.public_ip if contains(values.tags, "ansible")][0]}"
  } : {}
  description = "k3s cluster API endpoints"
}