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

# Configure the Microsoft Azure Provider
provider "azurerm" {
  features {}
  # Turn off provider registration if you have imported the provider already
  #skip_provider_registration = true
  # subscription_id: The subscription ID to use.
  subscription_id   = var.azure_subscription_id
  # tenant_id: The tenant ID to use.
  tenant_id         = var.azure_tenant_id
}

# Standard locals for k3s cluster deployment
locals {
  cloud_provider = "azure"
  cloud_region   = var.location
  
  # Standard ansible server IP detection
  ansibleserver_ip = [
    for x, values in module.design.instances : azurerm_network_interface.nic[x].private_ip_address
    if contains(values.tags, "ansible")
  ]
}

module "design" {
  source       = "../common/design"
  cluster_name = var.cluster_name
  domain       = var.domain
  instances    = var.instances
  volumes      = var.volumes
}

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

# Check if user provided resource group is valid
data "azurerm_resource_group" "example" {
  count = var.azure_resource_group == "" ? 0 : 1
  name  = var.azure_resource_group
}

# Create a resource group
resource "azurerm_resource_group" "group" {
  count    = var.azure_resource_group == "" ? 1 : 0
  #name     = var.azure_resource_group
  name     = "${var.cluster_name}_resource_group"
  location = var.location
}

data "azurerm_proximity_placement_group" "ppg" {
  count               = var.proximity_placement_group.new ? 0 : 1
  name                = var.proximity_placement_group.name
  resource_group_name = azurerm_resource_group.group[count.index].name
}

# Create an availability set for the execution nodes
resource "azurerm_availability_set" "avset" {
  name                = "${var.cluster_name}_availability_set"
  location            = var.location
  resource_group_name = local.resource_group_name
  platform_update_domain_count = 1
  platform_fault_domain_count  = 1
}

locals {
  to_build_instances = {
    for key, values in module.design.instances: key => values
    if ! contains(values.tags, "pool") || contains(var.pool, key)
   }
}

# Create virtual machine
resource "azurerm_linux_virtual_machine" "instances" {
  for_each              = local.to_build_instances
  size                  = each.value.type
  name                  = format("%s-%s", var.cluster_name, each.key)
  location              = var.location
  resource_group_name   = local.resource_group_name
  network_interface_ids = [azurerm_network_interface.nic[each.key].id]
  availability_set_id = contains(each.value["tags"], "node") ? azurerm_availability_set.avset.id : null

  os_disk {
    name                 = format("%s-%s-disk", var.cluster_name, each.key)
    caching              = "ReadWrite"
    storage_account_type = lookup(each.value, "disk_type", "Premium_LRS")
    disk_size_gb         = lookup(each.value, "disk_size", 30)
  }

  dynamic "plan" {
    for_each = var.plan["name"] != null ? [var.plan] : []
    iterator = plan
    content {
      name      = plan.value["name"]
      product   = plan.value["product"]
      publisher = plan.value["publisher"]
    }
  }

  dynamic "source_image_reference" {
    for_each = can(tomap(lookup(each.value, "image", var.image))) ? [lookup(each.value, "image", var.image)] : []
    iterator = key
    content {
      publisher = key.value["publisher"]
      offer     = key.value["offer"]
      sku       = key.value["sku"]
      version   = lookup(key.value, "version", "latest")
    }
  }
  source_image_id = can(tomap(lookup(each.value, "image", var.image))) ? null : tostring(lookup(each.value, "image", var.image))

  computer_name  = each.key
  admin_username = "azure"
  custom_data    = base64gzip(module.instance_config.user_data[each.key])

  disable_password_authentication = true
  dynamic "admin_ssh_key" {
    for_each = var.public_keys
    iterator = key
    content {
      username   = "azure"
      public_key = key.value
    }

  }

  priority = contains(each.value["tags"], "spot") ? "Spot" : "Regular"
  # Spot instances specifics
  max_bid_price   = contains(each.value["tags"], "spot") ? lookup(each.value, "max_bid_price", null) : null
  eviction_policy = contains(each.value["tags"], "spot") ? lookup(each.value, "eviction_policy", "Deallocate") : null

  lifecycle {
    ignore_changes = [
      source_image_reference,
      source_image_id,
      custom_data,
    ]
  }
}

resource "azurerm_storage_account" "ubiq" {
  name                     = format("%stgacct", var.storage_account_name)
  resource_group_name      = local.resource_group_name
  location                 = var.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    Environment = "Production"
  }
}

resource "azurerm_storage_share" "azure_files" {
  for_each = { for key, values in var.filesystems : key => values if lower(values["type"]) == "azurefiles" }

  name                 = "${var.cluster_name}-${each.key}-files"
  quota                = lookup(each.value, "quota", null)
  metadata             = lookup(each.value, "metadata", {})
  enabled_protocol     = lookup(each.value, "enabled_protocol", "SMB")

  storage_account_name = var.storage_account_name

  depends_on = [azurerm_storage_account.ubiq]
}

#resource "azurerm_storage_share_mount" "azure_files_mount" {
#  for_each = { for key, values in var.filesystems : key => values if lower(values["type"]) == "azurefiles" }
#
#  storage_share_name    = azurerm_storage_share.azure_files[each.key].name
#  mount_point           = "/mnt/${var.cluster_name}/${each.key}"
#  driver_letter         = null
#  driver_options        = null
#
#  depends_on = [azurerm_storage_share.azure_files]
#}


resource "azurerm_managed_disk" "volumes" {
  for_each             = module.design.volumes
  name                 = format("%s-%s", var.cluster_name, each.key)
  location             = var.location
  resource_group_name  = local.resource_group_name
  storage_account_type = lookup(each.value, "type", "Premium_LRS")
  create_option        = "Empty"
  disk_size_gb         = each.value.size
}

resource "azurerm_virtual_machine_data_disk_attachment" "attachments" {
  for_each           = module.design.volumes
  managed_disk_id    = azurerm_managed_disk.volumes[each.key].id
  virtual_machine_id = azurerm_linux_virtual_machine.instances[each.value.instance].id
  lun                = index(module.design.volume_per_instance[each.value.instance], replace(each.key, "${each.value.instance}-", ""))
  caching            = "ReadWrite"
}

locals {
  volume_devices = {
    for ki, vi in var.volumes :
    ki => {
      for kj, vj in vi :
      kj => [for key, volume in module.design.volumes :
        "/dev/disk/azure/scsi1/lun${index(module.design.volume_per_instance[volume.instance], replace(key, "${volume.instance}-", ""))}"
        if key == "${volume["instance"]}-${ki}-${kj}"
      ]
    }
  }
}

locals {
  resource_group_name = var.azure_resource_group == "" ? azurerm_resource_group.group[0].name : var.azure_resource_group


  vmsizes = jsondecode(file("${path.module}/vmsizes.json"))
  all_instances = { for x, values in module.design.instances :
    x => {
      public_ip = azurerm_public_ip.public_ip[x].ip_address
      local_ip  = azurerm_network_interface.nic[x].private_ip_address
      prefix    = values["prefix"]
      tags      = values["tags"]
      id        = try(azurerm_linux_virtual_machine.instances[x].id, "")
      hostkeys  = {
        rsa = module.instance_config.rsa_hostkeys[x]
        ed25519 = module.instance_config.ed25519_hostkeys[x]
      }
      specs = {
        cpus = local.vmsizes[values["type"]]["vcpus"]
        ram  = local.vmsizes[values["type"]]["ram"]
        gpus = local.vmsizes[values["type"]]["gpus"]
      }
    }
  }
  all_filesystems = {
    azure_files = {
      for key, values in var.filesystems:
        key => {
          ip = "woo"
          #ip = azurerm_storage_share_mount.azure_files_mount[key].ip_address
        } if lower(values["type"]) == "azurefiles"
    }
    #lustre = {
    #  for key, values in var.filesystems:
    #    key => {
    #      host = azurerm_storage_lustre_file_system.fsx_lustre[key].dns_name
    #    } if lower(values["type"]) == "lustre"
    #}
  }
}