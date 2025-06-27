# Standard k3s Cluster Example
# This example demonstrates the standard pattern for deploying
# a k3s cluster across any supported cloud provider

# Copyright 2025 Logicalis UKI. All Rights Reserved.

terraform {
  required_version = ">= 0.14.2"
}

module "k3s_cluster" {
  source = "./PROVIDER_FOLDER" # Replace with actual provider folder (aws, gcp, azure, etc.)
  
  # Standard k3s cluster configuration
  config_git_url = "https://github.com/logicalisuki/ubiquity-open.git"
  config_version = "main"

  cluster_name = "ubiquity-k3s"
  domain       = "example.com"
  
  # Standard k3s cluster instance configuration
  # This pattern ensures exactly 3 control plane nodes and configurable compute nodes
  instances = {
    mgmt = { 
      type = "PROVIDER_SMALL_INSTANCE_TYPE",  # e.g., t3.medium for AWS
      tags = ["mgmt", "ansible", "public"]
    }
    ctrl = { 
      type = "PROVIDER_MEDIUM_INSTANCE_TYPE", # e.g., t3.large for AWS
      tags = ["master", "k8s"]
      count = 3  # Always 3 for HA k3s control plane
    }
    worker = { 
      type = "PROVIDER_COMPUTE_INSTANCE_TYPE", # e.g., t3.xlarge for AWS
      tags = ["worker", "k8s", "compute"]
      count = var.worker_count  # Configurable number of compute nodes
    }
  }

  # Standard volume configuration for k3s clusters
  volumes = {
    # Shared storage for NFS (attached to mgmt node)
    nfs = {
      home     = { size = 100 }
      project  = { size = 500 }
      scratch  = { size = 500 }
    }
  }

  # Standard security and access configuration
  public_keys = [file("~/.ssh/id_rsa.pub")]
  
  # User management
  nb_users     = 10
  guest_passwd = "" # Auto-generated if blank
  
  # Provider-specific configuration (customize per provider)
  region = var.region # Define in provider-specific variables
  
  # Additional provider-specific variables as needed
  # availability_zone = var.availability_zone
  # image = var.image
  # etc.
}

# Standard variables for k3s deployment
variable "worker_count" {
  type        = number
  default     = 2
  description = "Number of worker nodes for the k3s cluster"
  validation {
    condition     = var.worker_count >= 1
    error_message = "At least 1 worker node is required for a functional k3s cluster."
  }
}

variable "region" {
  type        = string
  description = "Cloud provider region for deployment"
}

# Standard outputs for k3s cluster
output "cluster_info" {
  value = module.k3s_cluster.cluster_info
  description = "k3s cluster information"
}

output "kubeconfig_command" {
  value = "ssh -i ~/.ssh/id_rsa ${module.k3s_cluster.sudoer_username}@${module.k3s_cluster.public_ip} 'sudo cat /etc/rancher/k3s/k3s.yaml'"
  description = "Command to retrieve kubeconfig from the cluster"
}

output "ansible_access" {
  value = "ssh -i ~/.ssh/id_rsa ${module.k3s_cluster.sudoer_username}@${module.k3s_cluster.public_ip}"
  description = "SSH command to access the Ansible management server"
}

output "cluster_endpoints" {
  value = {
    api_server = "https://${module.k3s_cluster.public_ip}:6443"
    dashboard  = "https://${module.k3s_cluster.public_ip}"
  }
  description = "k3s cluster API endpoints"
}
