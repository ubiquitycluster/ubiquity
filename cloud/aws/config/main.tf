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
terraform {
  required_version = ">= 1.2.1"
}

variable "pool" {
  description = "Slurm pool of compute nodes"
  default = []
}

module "aws" {
  source         = "../"
  config_git_url = "https://github.com/ubiquitycluster/ubiq-playbooks.git"
  config_version = "main"

  cluster_name = "ubiquity"
  domain       = "ubiquitycluster.uk"
  
  # Rocky Linux 9.1 -  eu-west-2
  # https://rockylinux.org/cloud-images
  image        = "ami-0fee1bd6233a6b61c"

  instances = {
    mgmt  = { type = "r6id.32xlarge",  count = 1, tags = ["all", "mgmt", "ansible", "nfs", "slurm_master"] },
    ctrl  = { type = "i3en.24xlarge",  count = 3, tags = ["all", "ctrl", "master", "k8s"] },
    login = { type = "i4i.32xlarge", count = 1, tags = ["all", "login", "login_node", "public", "proxy", "worker"] },
    node  = { type = "i3en.24xlarge", count = 20, tags = ["all", "node", "compute", "worker"] }
  }
  private_vpc_id = 1
  private_vpc_private_subnet_ids = ["ubiquity-subnet"]
  # var.pool is managed by Slurm through Terraform REST API.
  # To let Slurm manage a type of nodes, add "pool" to its tag list.
  # When using Terraform CLI, this parameter is ignored.
  # Refer to Ubiquity Documentation - Enable Ubiquity Autoscaling
  pool = var.pool

  volumes = {
    nfs = {
      home     = { size = 10, type = "gp2" }
      project  = { size = 50, type = "gp2" }
      scratch  = { size = 50, type = "gp2" }
    }
  }

  public_keys = [file("~/.ssh/id_rsa.pub")]

  nb_users     = 10

  # Shared password, randomly chosen if blank
  guest_passwd = ""

  # AWS specifics
  region            = "eu-west-2"
}

output "accounts" {
  value = module.aws.accounts
}

output "public_ip" {
  value = module.aws.public_ip
}

## Uncomment to register your domain name with CloudFlare
# module "dns" {
#   source           = "git::https://github.com/ubiquitycluster/ubiquity.git/cloud/dns/cloudflare"
#   email            = "you@example.com"
#   name             = module.aws.cluster_name
#   domain           = module.aws.domain
#   public_instances = module.aws.public_instances
#   ssh_private_key  = module.aws.ssh_private_key
#   sudoer_username  = module.aws.accounts.sudoer.username
# }

## Uncomment to register your domain name with Google Cloud
# module "dns" {
#   source           = "git::https://github.com/ubiquitycluster/ubiquity.git/cloud/dns/gcloud"
#   email            = "you@example.com"
#   project          = "your-project-id"
#   zone_name        = "you-zone-name"
#   name             = module.aws.cluster_name
#   domain           = module.aws.domain
#   public_instances = module.aws.public_instances
#   ssh_private_key  = module.aws.ssh_private_key
#   sudoer_username  = module.aws.accounts.sudoer.username
# }

# output "hostnames" {
# 	value = module.dns.hostnames
# }