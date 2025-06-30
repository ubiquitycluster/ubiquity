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
# Standard k3s firewall rules for cluster communication
locals {
  # Standard k3s cluster communication ports
  k3s_firewall_rules = [
    {
      name        = "k3s-api-server"
      from_port   = 6443
      to_port     = 6443
      ip_protocol = "tcp"
      cidr        = "10.0.0.0/24"
    },
    {
      name        = "k3s-flannel-vxlan"
      from_port   = 8472
      to_port     = 8472
      ip_protocol = "udp"
      cidr        = "10.0.0.0/24"
    },
    {
      name        = "k3s-kubelet-metrics"
      from_port   = 10250
      to_port     = 10250
      ip_protocol = "tcp"
      cidr        = "10.0.0.0/24"
    },
    {
      name        = "k3s-flannel-wireguard"
      from_port   = 51820
      to_port     = 51821
      ip_protocol = "udp"
      cidr        = "10.0.0.0/24"
    },
    {
      name        = "k3s-etcd-client"
      from_port   = 2379
      to_port     = 2380
      ip_protocol = "tcp"
      cidr        = "10.0.0.0/24"
    }
  ]
  
  # Combine k3s rules with user-defined rules
  all_firewall_rules = concat(local.k3s_firewall_rules, var.firewall_rules)
}

resource "google_compute_network" "network" {
  name = "${var.cluster_name}-network"
}

resource "google_compute_subnetwork" "subnet" {
  name          = "${var.cluster_name}-subnet"
  network       = google_compute_network.network.self_link
  ip_cidr_range = "10.0.0.0/24"
  region        = var.region
}

resource "google_compute_router" "router" {
  name    = "${var.cluster_name}-router"
  region  = var.region
  network = google_compute_network.network.self_link
  bgp {
    asn = 64514
  }
}

resource "google_compute_router_nat" "nat" {
  name                               = "${var.cluster_name}-nat"
  router                             = google_compute_router.router.name
  region                             = var.region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
}

resource "google_compute_firewall" "allow_all_internal" {
  name    = format("%s-allow-all-internal", var.cluster_name)
  network = google_compute_network.network.self_link

  source_ranges = [google_compute_subnetwork.subnet.ip_cidr_range]

  allow {
    protocol = "tcp"
  }

  allow {
    protocol = "udp"
  }

  allow {
    protocol = "icmp"
  }

}

resource "google_compute_firewall" "default" {
  count   = length(local.all_firewall_rules)
  name    = format("%s-%s", var.cluster_name, lower(local.all_firewall_rules[count.index].name))
  network = google_compute_network.network.self_link

  source_ranges = [local.all_firewall_rules[count.index].cidr]

  allow {
    protocol = local.all_firewall_rules[count.index].ip_protocol
    ports = [local.all_firewall_rules[count.index].from_port != local.all_firewall_rules[count.index].to_port ?
      "${local.all_firewall_rules[count.index].from_port}-${local.all_firewall_rules[count.index].to_port}" :
      tostring(local.all_firewall_rules[count.index].from_port)
    ]
  }

  target_tags = ["public"]
}

resource "google_compute_address" "nic" {
  for_each     = module.design.instances
  name         = format("%s-%s-ipv4", var.cluster_name, each.key)
  address_type = "INTERNAL"
  subnetwork   = google_compute_subnetwork.subnet.self_link
  region       = var.region
}

resource "google_compute_address" "public_ip" {
  for_each = { for x, values in module.design.instances : x => true if contains(values.tags, "public") }
  name     = format("%s-%s-public-ipv4", var.cluster_name, each.key)
}