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

# Standard locals for k3s cluster deployment - extending existing locals
# Note: cloud_provider and cloud_region are already defined in openstack.tf
locals {
  # Standard cloud provider identification
  cloud_provider = "openstack"
  cloud_region   = "openstack"
  
  # Standard ansible server IP detection
  ansibleserver_ip = [
    for x, values in module.design.instances : openstack_networking_port_v2.nic[x].all_fixed_ips[0]
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
  ansibleserver_ip  = local.ansibleserver_ip
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

data "openstack_images_image_v2" "image" {
  for_each    = var.instances
  name_regex  = lookup(each.value, "image", var.image)
  most_recent = true
}

data "openstack_compute_flavor_v2" "flavors" {
  for_each = var.instances
  name     = each.value.type
}

resource "openstack_compute_keypair_v2" "keypair" {
  name       = "${var.cluster_name}-key"
  public_key = var.public_keys[0]
}

locals {
  to_build_instances = {
    for key, values in module.design.instances: key => values
    if ! contains(values.tags, "pool") || contains(var.pool, key)
   }
}

resource "openstack_compute_instance_v2" "instances" {
  for_each = local.to_build_instances
  name     = format("%s-%s", var.cluster_name, each.key)
  image_id = lookup(each.value, "disk_size", 10) > data.openstack_compute_flavor_v2.flavors[each.value.prefix].disk ? null : data.openstack_images_image_v2.image[each.value.prefix].id

  flavor_name  = each.value.type
  key_pair     = openstack_compute_keypair_v2.keypair.name
  user_data    = base64gzip(module.instance_config.user_data[each.key])
  metadata     = {}
  force_delete = true

  network {
    port = openstack_networking_port_v2.nic[each.key].id
  }
  dynamic "network" {
    for_each = local.ext_networks
    content {
      access_network = network.value.access_network
      name           = network.value.name
    }
  }

  dynamic "block_device" {
    for_each = lookup(each.value, "disk_size", 10) > data.openstack_compute_flavor_v2.flavors[each.value.prefix].disk ? [{ volume_size = lookup(each.value, "disk_size", 10) }] : []
    content {
      uuid                  = data.openstack_images_image_v2.image[each.value.prefix].id
      source_type           = "image"
      destination_type      = "volume"
      boot_index            = 0
      delete_on_termination = true
      volume_size           = block_device.value.volume_size
      volume_type           = lookup(each.value, "disk_type", null)
    }
  }

  lifecycle {
    ignore_changes = [
      image_id,
      block_device[0].uuid,
      user_data,
    ]
  }
}

resource "openstack_blockstorage_volume_v3" "volumes" {
  for_each    = module.design.volumes
  name        = "${var.cluster_name}-${each.key}"
  description = "${var.cluster_name} ${each.key}"
  size        = each.value.size
  volume_type = lookup(each.value, "type", null)
  snapshot_id = lookup(each.value, "snapshot", null)
}

resource "openstack_compute_volume_attach_v2" "attachments" {
  for_each    = module.design.volumes
  instance_id = openstack_compute_instance_v2.instances[each.value.instance].id
  volume_id   = openstack_blockstorage_volume_v3.volumes[each.key].id
}

locals {
  volume_devices = {
    for ki, vi in var.volumes :
    ki => {
      for kj, vj in vi :
      kj => [for key, volume in module.design.volumes :
        "/dev/disk/by-id/*${substr(openstack_blockstorage_volume_v3.volumes["${volume["instance"]}-${ki}-${kj}"].id, 0, 20)}"
        if key == "${volume["instance"]}-${ki}-${kj}"
      ]
    }
  }
}

locals {
  all_instances = { for x, values in module.design.instances :
    x => {
      public_ip = contains(values["tags"], "public") ? local.public_ip[x] : ""
      local_ip  = openstack_networking_port_v2.nic[x].all_fixed_ips[0]
      prefix    = values["prefix"]
      tags      = values["tags"]
      id        = ! contains(values["tags"], "pool") || contains(var.pool, x) ? openstack_compute_instance_v2.instances[x].id : ""
      hostkeys = {
        rsa = module.instance_config.rsa_hostkeys[x]
        ed25519 = module.instance_config.ed25519_hostkeys[x]
      }
      specs = {
        cpus = data.openstack_compute_flavor_v2.flavors[values["prefix"]].vcpus
        ram  = data.openstack_compute_flavor_v2.flavors[values["prefix"]].ram
        gpus = sum([
          parseint(lookup(data.openstack_compute_flavor_v2.flavors[values["prefix"]].extra_specs, "resources:VGPU", "0"), 10),
          parseint(split(":", lookup(data.openstack_compute_flavor_v2.flavors[values["prefix"]].extra_specs, "pci_passthrough:alias", "gpu:0"))[1], 10)
        ])
      }
    }
  }
  all_filesystems = {}
}