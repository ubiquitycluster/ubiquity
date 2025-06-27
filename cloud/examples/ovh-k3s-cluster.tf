# OVH k3s cluster deployment example
# This example demonstrates deploying a high-availability k3s cluster on OVH Public Cloud
# following Ubiquity's standardized infrastructure patterns.

terraform {
  required_providers {
    openstack = {
      source  = "terraform-provider-openstack/openstack"
      version = "~> 1.53"
    }
  }
}

# Standard k3s cluster configuration
# This creates:
# - 3 control plane nodes (for HA)
# - 2 worker nodes (configurable)
# - 1 ansible management node
# - Standard networking with k3s ports configured

module "ovh_k3s_cluster" {
  source = "../ovh"

  # Basic cluster configuration
  cluster_name = "ovh-k3s-demo"
  domain       = "example.com"

  # Standard k3s instance configuration
  instances = {
    # Control plane nodes (exactly 3 required for k3s HA)
    master1 = { prefix = "master", type = "s1-4", tags = ["master", "public"] }
    master2 = { prefix = "master", type = "s1-4", tags = ["master"] }
    master3 = { prefix = "master", type = "s1-4", tags = ["master"] }
    
    # Worker nodes (scalable)
    worker1 = { prefix = "worker", type = "s1-4", tags = ["worker"] }
    worker2 = { prefix = "worker", type = "s1-4", tags = ["worker"] }
    
    # Ansible management node
    mgmt = { prefix = "mgmt", type = "s1-2", tags = ["ansible", "public"] }
  }

  # k3s software stack configuration
  ansible_vars = {
    software_stack = "k3s"
    k3s_version    = "v1.28.5+k3s1"
    # Additional k3s configuration can be added here
  }

  # OVH-specific settings
  image = "Ubuntu 22.04"
  
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
  value       = module.ovh_k3s_cluster.cluster_info
}

output "kubeconfig_command" {
  description = "Command to retrieve kubeconfig"
  value       = module.ovh_k3s_cluster.kubeconfig_command
}

output "cluster_endpoints" {
  description = "Cluster endpoints"
  value       = module.ovh_k3s_cluster.cluster_endpoints
}

output "ssh_access" {
  description = "SSH access information"
  value = {
    ansible_server = "ssh ${module.ovh_k3s_cluster.cluster_info.ansible_server}"
    master_nodes   = [
      for endpoint in module.ovh_k3s_cluster.cluster_endpoints.kubernetes_api :
      "ssh ubuntu@${split(":", endpoint)[0]}"
    ]
  }
}
