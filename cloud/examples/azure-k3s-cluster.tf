# Azure k3s cluster deployment example
# This example demonstrates deploying a high-availability k3s cluster on Azure
# following Ubiquity's standardized infrastructure patterns.

terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}

provider "azurerm" {
  features {}
}

# Standard k3s cluster configuration
# This creates:
# - 3 control plane nodes (for HA)
# - 2 worker nodes (configurable)
# - 1 ansible management node
# - Standard networking with k3s ports configured

module "azure_k3s_cluster" {
  source = "../azure"

  # Basic cluster configuration
  cluster_name = "azure-k3s-demo"
  domain       = "example.com"

  # Standard k3s instance configuration
  instances = {
    # Control plane nodes (exactly 3 required for k3s HA)
    master1 = { prefix = "master", type = "Standard_B2ms", tags = ["master", "public"] }
    master2 = { prefix = "master", type = "Standard_B2ms", tags = ["master"] }
    master3 = { prefix = "master", type = "Standard_B2ms", tags = ["master"] }
    
    # Worker nodes (scalable)
    worker1 = { prefix = "worker", type = "Standard_B2ms", tags = ["worker"] }
    worker2 = { prefix = "worker", type = "Standard_B2ms", tags = ["worker"] }
    
    # Ansible management node
    mgmt = { prefix = "mgmt", type = "Standard_B1ms", tags = ["ansible", "public"] }
  }

  # k3s software stack configuration
  ansible_vars = {
    software_stack = "k3s"
    k3s_version    = "v1.28.5+k3s1"
    # Additional k3s configuration can be added here
  }

  # Azure-specific settings
  location            = "East US"
  resource_group_name = "rg-k3s-demo"
  
  # SSH and user configuration
  sudoer_username = "ubuntu"
  public_keys     = [file("~/.ssh/id_rsa.pub")]
  
  # Standard volumes (optional)
  volumes = {}
  
  # Network security
  firewall_rules = [
    {
      from_port   = 22
      to_port     = 22
      ip_protocol = "tcp"
      cidr        = "0.0.0.0/0"
    },
    {
      from_port   = 80
      to_port     = 80
      ip_protocol = "tcp"
      cidr        = "0.0.0.0/0"
    },
    {
      from_port   = 443
      to_port     = 443
      ip_protocol = "tcp"
      cidr        = "0.0.0.0/0"
    }
  ]
}

# Outputs for cluster management
output "cluster_info" {
  description = "k3s cluster information"
  value       = module.azure_k3s_cluster.cluster_info
}

output "kubeconfig_command" {
  description = "Command to retrieve kubeconfig"
  value       = module.azure_k3s_cluster.kubeconfig_command
}

output "cluster_endpoints" {
  description = "Cluster endpoints"
  value       = module.azure_k3s_cluster.cluster_endpoints
}

output "ssh_access" {
  description = "SSH access information"
  value = {
    ansible_server = "ssh ${module.azure_k3s_cluster.cluster_info.ansible_server}"
    master_nodes   = [
      for endpoint in module.azure_k3s_cluster.cluster_endpoints.kubernetes_api :
      "ssh ubuntu@${split(":", endpoint)[0]}"
    ]
  }
}
