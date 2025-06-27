# Ubiquity Standard Network Template
# This file provides a standard template for networking across all cloud providers
# to ensure consistency for k3s cluster deployment

# Copyright 2025 Logicalis UKI. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 
# See the License for the specific language governing permissions and
# limitations under the License.

# Standard network configuration requirements:
# 1. Create VPC/Network with appropriate CIDR (default: 10.0.0.0/16)
# 2. Create subnet for instances (default: 10.0.1.0/24)  
# 3. Create security groups/firewall rules for k3s
# 4. Create network interfaces for each instance
# 5. Create public IPs for instances tagged with "public"
# 6. Ensure proper routing and NAT for internet access

locals {
  # Standard network configuration
  vpc_cidr    = "10.0.0.0/16"
  subnet_cidr = "10.0.1.0/24"
  
  # Standard k3s firewall rules - required for cluster communication
  k3s_firewall_rules = [
    {
      name        = "SSH"
      from_port   = 22
      to_port     = 22
      ip_protocol = "tcp"
      cidr        = "0.0.0.0/0"
    },
    {
      name        = "HTTP"
      from_port   = 80
      to_port     = 80
      ip_protocol = "tcp"
      cidr        = "0.0.0.0/0"
    },
    {
      name        = "HTTPS"
      from_port   = 443
      to_port     = 443
      ip_protocol = "tcp"
      cidr        = "0.0.0.0/0"
    },
    {
      name        = "k3s-api-server"
      from_port   = 6443
      to_port     = 6443
      ip_protocol = "tcp"
      cidr        = local.vpc_cidr
    },
    {
      name        = "k3s-flannel-vxlan"
      from_port   = 8472
      to_port     = 8472
      ip_protocol = "udp"
      cidr        = local.vpc_cidr
    },
    {
      name        = "k3s-kubelet-metrics"
      from_port   = 10250
      to_port     = 10250
      ip_protocol = "tcp"
      cidr        = local.vpc_cidr
    },
    {
      name        = "k3s-flannel-wireguard"
      from_port   = 51820
      to_port     = 51821
      ip_protocol = "udp"
      cidr        = local.vpc_cidr
    },
    {
      name        = "k3s-etcd-client"
      from_port   = 2379
      to_port     = 2380
      ip_protocol = "tcp"
      cidr        = local.vpc_cidr
    }
  ]
  
  # Combine standard k3s rules with user-defined rules
  all_firewall_rules = concat(local.k3s_firewall_rules, var.firewall_rules)
  
  # Standard ansible server IP detection from network interfaces
  ansibleserver_ip = [
    for x, values in module.design.instances : 
    PROVIDER_NIC_PRIVATE_IP_REFERENCE # Replace with provider-specific NIC private IP
    if contains(values.tags, "ansible")
  ]
}

# Provider-specific network resources should include:
# 1. VPC/Network resource
# 2. Subnet resource  
# 3. Internet Gateway/Router
# 4. NAT Gateway (if needed)
# 5. Security Group/Firewall with k3s rules
# 6. Network Interface for each instance
# 7. Public IP for instances with "public" tag
# 8. Route tables and associations

# Example structure (customize for each provider):
/*
resource "provider_vpc" "network" {
  cidr = local.vpc_cidr
  tags = {
    Name = "${var.cluster_name}-vpc"
  }
}

resource "provider_subnet" "subnet" {
  vpc_id = provider_vpc.network.id
  cidr   = local.subnet_cidr
  tags = {
    Name = "${var.cluster_name}-subnet"
  }
}

resource "provider_security_group" "k3s_firewall" {
  name   = "${var.cluster_name}-k3s-sg"
  vpc_id = provider_vpc.network.id
  
  dynamic "ingress" {
    for_each = local.all_firewall_rules
    content {
      from_port   = ingress.value.from_port
      to_port     = ingress.value.to_port
      protocol    = ingress.value.ip_protocol
      cidr_blocks = [ingress.value.cidr]
    }
  }
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "provider_network_interface" "nic" {
  for_each    = module.design.instances
  subnet_id   = provider_subnet.subnet.id
  security_groups = [provider_security_group.k3s_firewall.id]
  
  tags = {
    Name = "${var.cluster_name}-${each.key}-nic"
  }
}

resource "provider_public_ip" "public_ip" {
  for_each = {
    for key, values in module.design.instances : key => values
    if contains(values.tags, "public")
  }
  
  tags = {
    Name = "${var.cluster_name}-${each.key}-public-ip"
  }
}
*/
