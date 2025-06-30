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

# Create virtual network
resource "azurerm_virtual_network" "network" {
  name                = "${var.cluster_name}_vnet"
  address_space       = ["10.0.0.0/16"]
  location            = var.location
  resource_group_name = local.resource_group_name
}

# Create subnet
resource "azurerm_subnet" "subnet" {
  name                 = "${var.cluster_name}_subnet"
  resource_group_name  = local.resource_group_name
  virtual_network_name = azurerm_virtual_network.network.name
  address_prefixes     = ["10.0.1.0/24"]
}

# Create public IPs
resource "azurerm_public_ip" "public_ip" {
  for_each            = module.design.instances
  name                = format("%s-%s-public-ipv4", var.cluster_name, each.key)
  location            = var.location
  resource_group_name = local.resource_group_name
  allocation_method   = contains(each.value.tags, "public") ? "Static" : "Dynamic"
}

# Create Network Security Group and rule
resource "azurerm_network_security_group" "public" {
  name                = "${var.cluster_name}_public-firewall"
  location            = var.location
  resource_group_name = local.resource_group_name

  dynamic "security_rule" {
    for_each = local.all_firewall_rules
    iterator = rule
    content {
      name                       = rule.value.name
      priority                   = (100 + rule.value.from_port) % 4096
      direction                  = "Inbound"
      access                     = "Allow"
      protocol                   = title(rule.value.ip_protocol)
      source_port_range          = "*"
      destination_port_range     = "${rule.value.from_port}-${rule.value.to_port}"
      source_address_prefix      = "*"
      destination_address_prefix = rule.value.cidr
    }
  }
}

# Create network interface
resource "azurerm_network_interface" "nic" {
  for_each            = module.design.instances
  name                = format("%s-%s-nic", var.cluster_name, each.key)
  location            = var.location
  resource_group_name = local.resource_group_name
  enable_accelerated_networking = local.accelerated_network

  ip_configuration {
    name                          = format("%s-%s-nic_config", var.cluster_name, each.key)
    subnet_id                     = azurerm_subnet.subnet.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.public_ip[each.key].id
  }
}

resource "azurerm_network_interface_security_group_association" "public" {
  for_each                  = { for x, values in module.design.instances : x => true if contains(values.tags, "public") }
  network_interface_id      = azurerm_network_interface.nic[each.key].id
  network_security_group_id = azurerm_network_security_group.public.id
}

locals {
  ansibleserver_ip = [
      for x, values in module.design.instances : azurerm_network_interface.nic[x].private_ip_address
      if contains(values.tags, "ansible")
  ]
  accelerated_network = "true"
}