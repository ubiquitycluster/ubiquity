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
resource "openstack_compute_secgroup_v2" "secgroup" {
  name        = "${var.cluster_name}-secgroup"
  description = "${var.cluster_name} security group"

  rule {
    from_port   = -1
    to_port     = -1
    ip_protocol = "icmp"
    self        = true
  }

  rule {
    from_port   = 1
    to_port     = 65535
    ip_protocol = "tcp"
    self        = true
  }

  rule {
    from_port   = 1
    to_port     = 65535
    ip_protocol = "udp"
    self        = true
  }

  # k3s-specific rules
  # Kubernetes API server
  rule {
    from_port   = 6443
    to_port     = 6443
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  # flannel VXLAN
  rule {
    from_port   = 8472
    to_port     = 8472
    ip_protocol = "udp"
    self        = true
  }

  # kubelet API
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

  # etcd (for HA clusters)
  rule {
    from_port   = 2379
    to_port     = 2380
    ip_protocol = "tcp"
    self        = true
  }

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