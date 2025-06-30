# Standard k3s Cluster Example for GCP
# This example demonstrates the standard pattern for deploying
# a k3s cluster on GCP with exactly 3 control plane nodes

terraform {
  required_version = ">= 0.14.2"
}

module "k3s_cluster" {
  source = "../gcp"
  
  # Standard k3s cluster configuration
  config_git_url = "https://github.com/ubiquitycluster/ubiquity.git"
  config_version = "main"

  cluster_name = "ubiquity-k3s"
  domain       = "example.com"
  image        = "ubuntu-os-cloud/ubuntu-2004-lts"
  
  # Standard k3s cluster instance configuration
  # This pattern ensures exactly 3 control plane nodes and configurable compute nodes
  instances = {
    mgmt = { 
      type = "e2-medium",
      tags = ["mgmt", "ansible", "public"]
    }
    ctrl = { 
      type = "e2-standard-2",
      tags = ["master", "k8s"]
      count = 3  # Always 3 for HA k3s control plane
    }
    worker = { 
      type = "e2-standard-4",
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
  
  # GCP-specific configuration
  project = var.project
  region  = var.region
  zone    = var.zone
  
  # Master key for cluster configuration
  master_key = var.master_key
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

variable "project" {
  type        = string
  description = "GCP project ID"
}

variable "region" {
  type        = string
  default     = "us-central1"
  description = "GCP region for deployment"
}

variable "zone" {
  type        = string
  default     = ""
  description = "GCP zone (optional, will be auto-selected if not specified)"
}

variable "master_key" {
  type        = string
  description = "Master key for cluster configuration"
  sensitive   = true
}

# Standard outputs for k3s cluster
output "cluster_info" {
  value = module.k3s_cluster.cluster_info
  description = "k3s cluster information"
}

output "kubeconfig_command" {
  value = module.k3s_cluster.kubeconfig_command
  description = "Command to retrieve kubeconfig from the cluster"
}

output "cluster_endpoints" {
  value = module.k3s_cluster.cluster_endpoints
  description = "k3s cluster API endpoints"
}

output "accounts" {
  value = module.k3s_cluster.accounts
  description = "User account information"
}

output "public_ip" {
  value = module.k3s_cluster.public_ip
  description = "Public IP addresses of cluster nodes"
}
