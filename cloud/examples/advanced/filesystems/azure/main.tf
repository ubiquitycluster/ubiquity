terraform {
  required_version = ">= 1.2.1"
}

# resource "azurerm_resource_group" "group" {
#   name     = "ubiquity-rg"
#   location = "UK South"
# }

variable "pool" {
  description = "Slurm pool of compute nodes"
  default = []
}

module "azure" {
  source         = "../../azure"
  config_git_url = "https://github.com/ubiquitycluster/ubiq-playbooks.git"
  config_version = "main"
  azure_subscription_id   = "21b068cd-884d-4a7c-bc01-2558c717712d"
  azure_tenant_id         = "782e3213-b567-477d-8d64-77a942024aff"
  #azure_resource_group = "ubiquity-rg"
  cluster_name = "ubiq-azure"
  domain       = "ubiquitycluster.uk"

  # Visit https://azuremarketplace.microsoft.com/en-us/marketplace/apps/almalinux.almalinux
  # Or for the HPC image: https://azuremarketplace.microsoft.com/en-us/marketplace/apps/almalinux.almalinux-hpc
  # to contract the free AlmaLinux plan and be able to use the image.
  plan = {
    name      = "8_5"
    product   = "almalinux"
    publisher = "almalinux"
  }
  image        = {
    publisher = "almalinux",
    offer     = "almalinux",
    sku       = "8_5",
    version   = "8.5.20211118"
  }

  instances = {
    mgmt  = { type = "Standard_DS2_v2",  count = 1, tags = ["mgmt", "ansible", "nfs"] },
    ctrl = { type = "Standard_DS1_v2", count = 3, tags = ["ctrl", "master"] },
    login = { type = "Standard_DS1_v2", count = 1, tags = ["login", "public", "proxy", "worker"] },
    node  = { type = "Standard_DS1_v2",  count = 1, tags = ["node", "compute", "worker"] }
  }

  # var.pool is managed by Slurm through Terraform REST API.
  # To let Slurm manage a type of nodes, add "pool" to its tag list.
  # When using Terraform CLI, this parameter is ignored.
  # Refer to Ubiquity Documentation - Enable Ubiquity Autoscaling
  pool = var.pool

  volumes = {
    nfs = {
      home     = { size = 10 }
      project  = { size = 50 }
      scratch  = { size = 50 }
    }
  }
  filesystems = {
#    home    = { type = "azurefiles" }
#    project = { type = "lustro", size = 1200 }
#    scratch = { type = "lustro", size = 1200 }
  }
  public_keys = [file("~/.ssh/id_rsa.pub")]

  nb_users     = 10
  # Shared password, randomly chosen if blank
  guest_passwd = ""

  # Azure specifics
  location = "uksouth"
}

output "accounts" {
  value = module.azure.accounts
}

output "public_ip" {
  value = module.azure.public_ip
}

## Uncomment to register your domain name with CloudFlare
#module "dns" {
#  source           = "../../dns/cloudflare"
#  email            = "christopher.james.coates@gmail.com"
#  name             = module.azure.cluster_name
#  domain           = module.azure.domain
#  public_instances = module.azure.public_instances
#  ssh_private_key  = module.azure.ssh_private_key
#  sudoer_username  = module.azure.accounts.sudoer.username
#}

## Uncomment to register your domain name with Google Cloud
# module "dns" {
#   source           = "git::https://github.com/ubiquitycluster/ubiquity.git/cloud/dns/gcloud"
#   email            = "you@example.com"
#   project          = "your-project-id"
#   zone_name        = "you-zone-name"
#   name             = module.azure.cluster_name
#   domain           = module.azure.domain
#   public_instances = module.azure.public_instances
#   ssh_private_key  = module.azure.ssh_private_key
#   sudoer_username  = module.azure.accounts.sudoer.username
# }

# output "hostnames" {
# 	value = module.dns.hostnames
# }
