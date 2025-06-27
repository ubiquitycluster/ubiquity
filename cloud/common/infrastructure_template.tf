# Ubiquity Standard Infrastructure Template
# This file provides a standard template for cloud provider implementations
# to ensure consistency across all providers for k3s cluster deployment

# Copyright 2025 Logicalis UKI. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 
# See the License for the specific language governing permissions and
# limitations under the License.

# Provider configuration (customize per provider)
# provider "provider_name" {
#   region = var.region
# }

# Standard design module - consistent across all providers
module "design" {
  source       = "../common/design"
  cluster_name = var.cluster_name
  domain       = var.domain
  instances    = var.instances
  volumes      = var.volumes
}

# Standard instance configuration - consistent across all providers
module "instance_config" {
  source           = "../common/instance_config"
  instances        = module.design.instances
  config_git_url   = var.config_git_url
  config_version   = var.config_version
  ansibleserver_ip = local.ansibleserver_ip
  sudoer_username  = var.sudoer_username
  public_keys      = var.public_keys
  generate_ssh_key = var.generate_ssh_key
}

# Standard cluster configuration - consistent across all providers
module "cluster_config" {
  source          = "../common/cluster_config"
  instances       = local.all_instances
  nb_users        = var.nb_users
  ansible_vars    = var.ansible_vars
  software_stack  = var.software_stack
  cloud_provider  = local.cloud_provider
  cloud_region    = local.cloud_region
  sudoer_username = var.sudoer_username
  public_keys     = var.public_keys
  guest_passwd    = var.guest_passwd
  domain_name     = module.design.domain_name
  cluster_name    = var.cluster_name
  volume_devices  = local.volume_devices
  filesystems     = local.all_filesystems
  tf_ssh_key      = module.instance_config.ssh_key
  master_key      = module.instance_config.master_key 
}

# Standard locals that should be defined in each provider
locals {
  cloud_provider = "PROVIDER_NAME" # Replace with actual provider name
  cloud_region   = var.region
  
  # Standard ansible server IP detection
  ansibleserver_ip = [
    for x, values in module.design.instances : 
    PROVIDER_SPECIFIC_IP_REFERENCE # Replace with provider-specific IP reference
    if contains(values.tags, "ansible")
  ]
  
  # Standard instance consolidation - customize IP references per provider
  all_instances = { 
    for x, values in module.design.instances :
    x => {
      public_ip   = contains(values["tags"], "public") ? PROVIDER_PUBLIC_IP_REFERENCE : "" # Replace with provider-specific public IP
      local_ip    = PROVIDER_PRIVATE_IP_REFERENCE # Replace with provider-specific private IP
      tags        = values["tags"]
      id          = PROVIDER_INSTANCE_ID_REFERENCE # Replace with provider-specific instance ID
      hostkeys    = {
        rsa = module.instance_config.rsa_hostkeys[x]
      }
    }
  }
  
  # Standard volume device mapping - customize per provider
  volume_devices = {
    for ki, vi in var.volumes :
    ki => {
      for kj, vj in vi :
      kj => [for key, volume in module.design.volumes :
        "PROVIDER_SPECIFIC_DEVICE_PATH" # Replace with provider-specific device path logic
        if key == "${volume["instance"]}-${ki}-${kj}"
      ]
    }
  }
  
  # Standard filesystem handling - customize per provider if supported
  all_filesystems = {
    # Add provider-specific filesystem resources here if supported
  }
}

# Standard validation for k3s cluster deployment
resource "null_resource" "k3s_cluster_validation" {
  count = module.design.cluster_type == "k3s" ? 1 : 0
  
  lifecycle {
    precondition {
      condition     = module.design.master_count == 3
      error_message = "k3s clusters require exactly 3 control plane nodes for high availability."
    }
  }
  
  triggers = {
    cluster_type = module.design.cluster_type
    master_count = module.design.master_count
  }
}

# Standard output format - consistent across all providers
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
