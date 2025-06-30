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

# Configure the AWS Provider
provider "aws" {
  region = var.region
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

data "aws_availability_zones" "available" {
  state = "available"
}

resource "random_shuffle" "random_az" {
  input = data.aws_availability_zones.available.names
  result_count = 1
}

locals {
  cloud_provider = "aws"
  cloud_region   = var.region
  
  availability_zone = (
    ( var.availability_zone != "" &&
      contains(data.aws_availability_zones.available.names,
               var.availability_zone)
      ?
      var.availability_zone : random_shuffle.random_az.result[0]
    )
  )
  
  # Standard ansible server IP detection
  ansibleserver_ip = [
    for x, values in module.design.instances : aws_network_interface.nic[x].private_ip
    if contains(values.tags, "ansible")
  ]
}

resource "aws_placement_group" "efa_group" {
  name     = "${var.cluster_name}-efa-placement_group"
  strategy = "cluster"
}

resource "aws_key_pair" "key" {
  key_name   = "${var.cluster_name}-key"
  public_key = var.public_keys[0]
}

# Instances
locals {
  regular_instances = {for key, values in module.design.instances: key => values if !contains(values["tags"], "spot")}
  spot_instances    = {for key, values in module.design.instances: key => values if contains(values["tags"], "spot")}
}

data "aws_ec2_instance_type" "instance_type" {
  for_each      = var.instances
  instance_type = each.value.type
}

locals {
  to_build_regular_instances = {
    for key, values in local.regular_instances: key => values
    if ! contains(values.tags, "pool") || contains(var.pool, key)
   }
  to_build_spot_instances = {
    for key, values in local.spot_instances: key => values
    if ! contains(values.tags, "pool") || contains(var.pool, key)
   }
}

resource "aws_instance" "instances" {
  for_each          = local.to_build_regular_instances
  instance_type     = each.value.type
  ami               = lookup(each.value, "image", var.image)
  user_data         = base64gzip(module.instance_config.user_data[each.key])
  availability_zone = local.availability_zone
  placement_group   = contains(each.value.tags, "efa") ? aws_placement_group.efa_group.id : null

  key_name          = aws_key_pair.key.key_name

  network_interface {
    network_interface_id = aws_network_interface.nic[each.key].id
    device_index         = 0
  }

  ebs_optimized = true
  root_block_device {
    volume_type = lookup(each.value, "disk_type", "gp2")
    volume_size = lookup(each.value, "disk_size", 10)
  }

  tags = {
    Name = format("%s-%s", var.cluster_name, each.key)
  }

  lifecycle {
    ignore_changes = [
      ami,
      user_data,
    ]
  }

  depends_on = [aws_internet_gateway.gw]
}

resource "aws_spot_instance_request" "spot_instances" {
  for_each          = local.to_build_spot_instances
  instance_type     = each.value.type
  ami               = lookup(each.value, "image", var.image)
  user_data         = base64gzip(module.instance_config.user_data[each.key])
  availability_zone = local.availability_zone
  placement_group   = contains(each.value.tags, "efa") ? aws_placement_group.efa_group.id : null

  key_name          = aws_key_pair.key.key_name

  network_interface {
    network_interface_id = aws_network_interface.nic[each.key].id
    device_index         = 0
  }

  ebs_optimized = true
  root_block_device {
    volume_type = lookup(each.value, "disk_type", "gp2")
    volume_size = lookup(each.value, "disk_size", 10)
  }

  tags = {
    Name = format("%s-%s", var.cluster_name, each.key)
  }

  lifecycle {
    ignore_changes = [
      ami
    ]
  }

  depends_on = [aws_internet_gateway.gw]

  # spot specific variables
  wait_for_fulfillment           = lookup(each.value, "wait_for_fulfillment", true)
  spot_type                      = lookup(each.value, "spot_type", "persistent")
  instance_interruption_behavior = lookup(each.value, "instance_interruption_behavior", "stop")
  spot_price                     = lookup(each.value, "spot_price", null)
  block_duration_minutes         = lookup(each.value, "block_duration_minutes", null)

}

resource "aws_efs_file_system" "efs" {
  for_each = { for key, values in var.filesystems: key => values if lower(values["type"]) == "efs" }

  performance_mode                = lookup(each.value, "performance_mode", null)
  provisioned_throughput_in_mibps = lookup(each.value, "provisioned_throughput_in_mibps", null)
  throughput_mode                 = lookup(each.value, "throughput_mode", null)

  tags = {
    Name = "${var.cluster_name}-${each.key}-efs"
  }
}

resource "aws_efs_mount_target" "efs_target" {
  for_each = { for key, values in var.filesystems: key => values if lower(values["type"]) == "efs" }

  file_system_id = aws_efs_file_system.efs[each.key].id
  subnet_id      = aws_subnet.subnet.id
}

resource "aws_fsx_lustre_file_system" "fsx_lustre" {
  for_each = { for key, values in var.filesystems: key => values if lower(values["type"]) == "lustre" }

  storage_capacity = max(each.value["size"], 1200)
  storage_type     = lookup(each.value, "storage_type", null)
  deployment_type  = lookup(each.value, "deployment_type", null)

  subnet_ids       = [aws_subnet.subnet.id]

  tags = {
    Name = "${var.cluster_name}-${each.key}-lustre"
  }
}

resource "aws_ebs_volume" "volumes" {
  for_each          = module.design.volumes
  availability_zone = local.availability_zone
  size              = each.value.size
  type              = lookup(each.value, "type", null)
  snapshot_id       = lookup(each.value, "snapshot", null)

  tags = {
    Name = "${var.cluster_name}-${each.key}"
  }
}

locals {
  device_names = [
    "/dev/sdf", "/dev/sdg", "/dev/sdh", "/dev/sdi", "/dev/sdj",
    "/dev/sdk", "/dev/sdl", "/dev/sdm", "/dev/sdn", "/dev/sdp"
  ]
}

resource "aws_volume_attachment" "attachments" {
  for_each     = module.design.volumes
  device_name  = local.device_names[index(module.design.volume_per_instance[each.value.instance], replace(each.key, "${each.value.instance}-", ""))]
  volume_id    = aws_ebs_volume.volumes[each.key].id
  instance_id  = aws_instance.instances[each.value.instance].id
  skip_destroy = true
}

locals {
  volume_devices = {
    for ki, vi in var.volumes :
    ki => {
      for kj, vj in vi :
      kj => [ for key, volume in module.design.volumes:
        "/dev/disk/by-id/*${replace(aws_ebs_volume.volumes["${volume["instance"]}-${ki}-${kj}"].id, "-", "")}"
        if key == "${volume["instance"]}-${ki}-${kj}"
      ]
    }
  }
}


locals {
  all_instances = { for x, values in module.design.instances :
    x => {
      public_ip   = contains(values["tags"], "public") ? aws_eip.public_ip[x].public_ip : ""
      local_ip    = aws_network_interface.nic[x].private_ip
      prefix    = values["prefix"]
      tags        = values["tags"]
      id          = try(! contains(values["tags"], "spot") ? aws_instance.instances[x].id : aws_spot_instance_request.spot_instances[x].spot_instance_id, "")
      hostkeys    = {
        rsa = module.instance_config.rsa_hostkeys[x]
        ed25519 = module.instance_config.ed25519_hostkeys[x]
      }
      specs = {
        cpus = data.aws_ec2_instance_type.instance_type[values["prefix"]].default_vcpus
        ram  = data.aws_ec2_instance_type.instance_type[values["prefix"]].memory_size
        gpus = try(one(data.aws_ec2_instance_type.instance_type[values["prefix"]].gpus).count, 0)
      }
    }
  }
  all_filesystems = {
    efs = {
      for key, values in var.filesystems:
        key => {
          ip = aws_efs_mount_target.efs_target[key].ip_address
        } if lower(values["type"]) == "efs"
    }
    lustre = {
      for key, values in var.filesystems:
        key => {
          host = aws_fsx_lustre_file_system.fsx_lustre[key].dns_name
        } if lower(values["type"]) == "lustre"
    }
  }
}