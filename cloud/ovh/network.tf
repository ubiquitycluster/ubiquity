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

# External network data source
data "openstack_networking_network_v2" "ext_network" {
  external = true
}

# Internal network
resource "openstack_networking_network_v2" "int_network" {
  name = "${var.cluster_name}_network"
}

# Subnet
resource "openstack_networking_subnet_v2" "subnet" {
  name        = "${var.cluster_name}_subnet"
  network_id  = openstack_networking_network_v2.int_network.id
  ip_version  = 4
  cidr        = "10.0.1.0/24"
  no_gateway  = true
  enable_dhcp = true
}

# Security group with k3s support
resource "openstack_compute_secgroup_v2" "secgroup" {
  name        = "${var.cluster_name}-secgroup"
  description = "${var.cluster_name} security group"

  # ICMP
  rule {
    from_port   = -1
    to_port     = -1
    ip_protocol = "icmp"
    self        = true
  }

  # All TCP traffic within cluster
  rule {
    from_port   = 1
    to_port     = 65535
    ip_protocol = "tcp"
    self        = true
  }

  # All UDP traffic within cluster
  rule {
    from_port   = 1
    to_port     = 65535
    ip_protocol = "udp"
    self        = true
  }

  # k3s_firewall_rules - k3s-specific rules for external access
  
  # Kubernetes API server (external access)
  rule {
    from_port   = 6443
    to_port     = 6443
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  # flannel VXLAN (cluster internal)
  rule {
    from_port   = 8472
    to_port     = 8472
    ip_protocol = "udp"
    self        = true
  }

  # kubelet API (cluster internal)
  rule {
    from_port   = 10250
    to_port     = 10250
    ip_protocol = "tcp"
    self        = true
  }

  # Wireguard (if using wireguard flannel backend)
  rule {
    from_port   = 51820
    to_port     = 51821
    ip_protocol = "udp"
    self        = true
  }

  # etcd (for HA clusters - cluster internal)
  rule {
    from_port   = 2379
    to_port     = 2380
    ip_protocol = "tcp"
    self        = true
  }

  # Custom firewall rules
  dynamic "rule" {
    for_each = var.firewall_rules
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
  for_each              = module.design.instances
  name                  = format("%s-%s-port", var.cluster_name, each.key)
  network_id            = local.network.id
  security_group_ids    = [openstack_compute_secgroup_v2.secgroup.id]
  port_security_enabled = true
  fixed_ip {
    subnet_id = local.subnet.id
  }
}

# Local values for network configuration
locals {
  network   = openstack_networking_network_v2.int_network
  subnet    = openstack_networking_subnet_v2.subnet
  public_ip = { 
    for x, values in module.design.instances : x => openstack_compute_instance_v2.instances[x].network[1].fixed_ip_v4 
    if contains(values.tags, "public") 
  }
  ext_networks = [{
    access_network = true,
    name           = data.openstack_networking_network_v2.ext_network.name
  }]
}
