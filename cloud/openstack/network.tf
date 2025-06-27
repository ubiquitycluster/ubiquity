# OpenStack Network Configuration for k3s Clusters
# Copyright 2025 Logicalis UKI. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 
# See the License for the specific language governing permissions and
# limitations under the License.

# Standard k3s firewall rules for cluster communication
locals {
  # k3s_firewall_rules - Standard k3s cluster communication ports
  k3s_firewall_rules = [
    {
      name        = "k3s-api-server"
      from_port   = 6443
      to_port     = 6443
      ip_protocol = "tcp"
      cidr        = "10.0.0.0/16"
    },
    {
      name        = "k3s-flannel-vxlan"
      from_port   = 8472
      to_port     = 8472
      ip_protocol = "udp"
      cidr        = "10.0.0.0/16"
    },
    {
      name        = "k3s-kubelet-metrics"
      from_port   = 10250
      to_port     = 10250
      ip_protocol = "tcp"
      cidr        = "10.0.0.0/16"
    },
    {
      name        = "k3s-flannel-wireguard"
      from_port   = 51820
      to_port     = 51821
      ip_protocol = "udp"
      cidr        = "10.0.0.0/16"
    },
    {
      name        = "k3s-etcd-client"
      from_port   = 2379
      to_port     = 2380
      ip_protocol = "tcp"
      cidr        = "10.0.0.0/16"
    }
  ]
  
  # Combine k3s rules with user-defined rules
  all_firewall_rules = concat(local.k3s_firewall_rules, var.firewall_rules)
}

# External network configuration
data "openstack_networking_network_v2" "external" {
  external = true
}

data "openstack_networking_subnet_v2" "subnet" {
  count     = var.subnet_id != null ? 1 : 0
  subnet_id = var.subnet_id
}

locals {
  ext_networks = var.subnet_id != null ? [] : length(var.os_floating_ips) > 0 ? [
    for network in var.os_floating_ips : {
      name           = network
      access_network = false
    }
  ] : [{
    name           = var.os_ext_network != null ? var.os_ext_network : data.openstack_networking_network_v2.external.name
    access_network = false
  }]
}

# Security group for k3s cluster communication
resource "openstack_compute_secgroup_v2" "k3s_firewall" {
  name        = "${var.cluster_name}-k3s-secgroup"
  description = "Security group for k3s cluster communication"

  # Allow all internal communication
  rule {
    from_port   = 1
    to_port     = 65535
    ip_protocol = "tcp"
    cidr        = "10.0.0.0/16"
  }

  rule {
    from_port   = 1
    to_port     = 65535
    ip_protocol = "udp"
    cidr        = "10.0.0.0/16"
  }

  # External access rules (combined k3s and user-defined)
  dynamic "rule" {
    for_each = local.all_firewall_rules
    content {
      from_port   = rule.value.from_port
      to_port     = rule.value.to_port
      ip_protocol = rule.value.ip_protocol
      cidr        = rule.value.cidr
    }
  }
}

# Network ports for instances
resource "openstack_networking_port_v2" "nic" {
  for_each       = module.design.instances
  name           = "${var.cluster_name}-${each.key}-port"
  network_id     = var.subnet_id != null ? data.openstack_networking_subnet_v2.subnet[0].network_id : data.openstack_networking_network_v2.external.id
  admin_state_up = true
  
  security_group_ids = [openstack_compute_secgroup_v2.k3s_firewall.id]

  dynamic "fixed_ip" {
    for_each = var.subnet_id != null ? [var.subnet_id] : []
    content {
      subnet_id = fixed_ip.value
    }
  }
}

# Floating IPs for public instances
resource "openstack_networking_floatingip_v2" "public_ip" {
  for_each = {
    for x, values in module.design.instances : x => true if contains(values.tags, "public")
  }
  pool = var.os_ext_network != null ? var.os_ext_network : data.openstack_networking_network_v2.external.name
}

resource "openstack_compute_floatingip_associate_v2" "public_ip" {
  for_each    = openstack_networking_floatingip_v2.public_ip
  floating_ip = each.value.address
  instance_id = openstack_compute_instance_v2.instances[each.key].id
}

locals {
  public_ip = {
    for x, values in module.design.instances : 
    x => contains(values.tags, "public") ? openstack_networking_floatingip_v2.public_ip[x].address : ""
  }
}
