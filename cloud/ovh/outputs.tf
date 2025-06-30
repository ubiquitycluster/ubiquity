# Copyright The Ubiquity Authors.
#
# Licensed under the Apache License, Version 2.0. Previously licensed under the Functional Source License (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/ubiquitycluster/ubiquity-open/blob/main/LICENSE
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
        "You have chosen to create user accounts yourself (`nb_users = 0`), please read the documentation on how to manage this at https://github.com/ubiquitycluster/ubiquity/blob/main/docs/README.md#103-add-a-user-account"
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

# Standard k3s cluster outputs
output "cluster_info" {
  description = "Information about the deployed cluster"
  value = {
    cluster_name     = var.cluster_name
    cloud_provider   = local.cloud_provider
    cloud_region     = local.cloud_region
    cluster_type     = module.design.cluster_type
    master_count     = module.design.master_count
    worker_count     = length([for k, v in module.design.instances : k if contains(v.tags, "worker")])
    ansible_server   = local.ansibleserver_ip
    software_stack   = lookup(var.ansible_vars, "software_stack", "unknown")
  }
}

output "kubeconfig_command" {
  description = "Command to retrieve kubeconfig for k3s cluster"
  value = lookup(var.ansible_vars, "software_stack", "null") == "k3s" ? (
    "scp ${var.sudoer_username}@${local.ansibleserver_ip}:/etc/rancher/k3s/k3s.yaml ./kubeconfig && sed -i 's/127.0.0.1/${local.ansibleserver_ip}/g' ./kubeconfig"
  ) : "Not applicable - cluster is not running k3s"
}

output "cluster_endpoints" {
  description = "Important cluster endpoints"
  value = lookup(var.ansible_vars, "software_stack", "null") == "k3s" ? {
    kubernetes_api = [
      for k, v in module.design.instances : 
      "${contains(v.tags, "public") ? openstack_compute_instance_v2.instances[k].network[1].fixed_ip_v4 : openstack_networking_port_v2.nic[k].all_fixed_ips[0]}:6443"
      if contains(v.tags, "master")
    ]
    ansible_server = "${local.ansibleserver_ip}:22"
  } : {}
}