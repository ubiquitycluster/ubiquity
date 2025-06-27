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
}

resource "aws_vpc" "network" {
  cidr_block = "10.0.0.0/16"

  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = "${var.cluster_name}-vpc"
  }
}

# Internet gateway to give our VPC access to the outside world
resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.network.id
}

# Grant the VPC internet access by creating a very generic
# destination CIDR ("catch all" - the least specific possible)
# such that we route traffic to outside as a last resource for
# any route that the table doesn't know about.
resource "aws_route" "internet_access" {
  route_table_id         = aws_vpc.network.main_route_table_id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.gw.id
}

resource "aws_subnet" "subnet" {
  vpc_id     = aws_vpc.network.id
  cidr_block = "10.0.0.0/24"
  availability_zone = local.availability_zone
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.cluster_name}-subnet"
  }
}

resource "aws_security_group" "allow_out_any" {
  name   = "allow_out_any"
  vpc_id = aws_vpc.network.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "allow_in_services" {
  name = "allow_in_services"

  description = "Allows services traffic into login nodes"
  vpc_id      = aws_vpc.network.id

  # Standard k3s cluster communication rules
  dynamic "ingress" {
    for_each = local.k3s_firewall_rules
    iterator = rule
    content {
      from_port   = rule.value.from_port
      to_port     = rule.value.to_port
      protocol    = rule.value.ip_protocol
      cidr_blocks = [rule.value.cidr]
    }
  }

  # User-defined firewall rules
  dynamic "ingress" {
    for_each = var.firewall_rules
    iterator = rule
    content {
      from_port   = rule.value.from_port
      to_port     = rule.value.to_port
      protocol    = rule.value.ip_protocol
      cidr_blocks = [rule.value.cidr]
    }
  }

  tags = {
    Name = "${var.cluster_name}-allow_in_services"
  }
}

resource "aws_security_group" "allow_any_inside_vpc" {
  name = "allow_any_inside_vpc"

  vpc_id = aws_vpc.network.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["10.0.0.0/16"]
    self        = true
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["10.0.0.0/16"]
    self        = true
  }

  tags = {
    Name = "${var.cluster_name}-allow_any_inside_vpc"
  }
}

resource "aws_network_interface" "nic" {
  for_each        = module.design.instances
  subnet_id       = aws_subnet.subnet.id
  interface_type  = contains(each.value["tags"], "efa") ? "efa" : null

  security_groups = concat(
    [
      aws_security_group.allow_any_inside_vpc.id,
      aws_security_group.allow_out_any.id,
    ],
    contains(each.value["tags"], "public") ? [aws_security_group.allow_in_services.id] : []
  )

  tags = {
    Name = "${var.cluster_name}-${each.key}-if"
  }
}

resource "aws_eip" "public_ip" {
  for_each = {
    for x, values in module.design.instances : x => true if contains(values.tags, "public")
  }
  domain     = "vpc"
  instance   = aws_instance.instances[each.key].id
  depends_on = [aws_internet_gateway.gw]
  tags = {
    Name = "${var.cluster_name}-${each.key}-eip"
  }
}

resource "aws_lb" "control_plane_int" {
  name =  "${var.infrastructure_id}-int"

  load_balancer_type = "network"
  internal = "true"
  subnets  = [for subnet in aws_subnet.subnet.* : subnet.id]

tags =  merge(
  var.default_tags,
  tomap(
    {
    "kubernetes.io/cluster/${var.infrastructure_id}" = "shared",
    "Name" = "${var.infrastructure_id}-int"
    }
    )
  )
}

resource "aws_lb_listener" "control_plane_int_6443" {
  load_balancer_arn =  aws_lb.control_plane_int.arn

  port = "6443"
  protocol = "TCP"

  default_action {
    target_group_arn =  aws_lb_target_group.control_plane_int_6443.arn
    type = "forward"
  }
}

resource "aws_lb_target_group" "control_plane_int_6443" {
  name =  "${var.infrastructure_id}-6443-int-tg"
  port = 6443
  protocol = "TCP"
  tags =  var.default_tags
  target_type = "ip"
  vpc_id =  aws_vpc.network.id
  deregistration_delay = 60
  /*health_check {
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 10
    port                = 22623
    protocol            = "HTTP"
    path                = "/"
  }*/
}

locals {
  ansibleserver_ip = [
      for x, values in module.design.instances : aws_network_interface.nic[x].private_ip
      if contains(values.tags, "ansible")
  ]
}