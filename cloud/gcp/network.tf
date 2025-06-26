# Copyright 2023 Logicalis UKI. All Rights Reserved.
#
# Licensed under the Functional Source License, Version 1.0, Apache 2.0 Change License (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/logicalisuki/ubiquity/blob/main/LICENSE.md
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# It also allows for the transition of this software to an Apache 2.0 Licence
# on the second anniversary of the date we make the software available.
# See the License for the specific language governing permissions and
# limitations under the License.
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
  count   = length(var.firewall_rules)
  name    = format("%s-%s", var.cluster_name, lower(var.firewall_rules[count.index].name))
  network = google_compute_network.network.self_link

  source_ranges = [var.firewall_rules[count.index].cidr]

  allow {
    protocol = var.firewall_rules[count.index].ip_protocol
    ports = [var.firewall_rules[count.index].from_port != var.firewall_rules[count.index].to_port ?
      "${var.firewall_rules[count.index].from_port}-${var.firewall_rules[count.index].to_port}" :
      var.firewall_rules[count.index].from_port
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

locals {
  ansibleserver_ip = [
      for x, values in module.design.instances : google_compute_address.nic[x].address
      if contains(values.tags, "ansible")
  ]
}