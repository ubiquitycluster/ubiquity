# Standard k3s Cluster Example for AWS
# This example demonstrates the standard pattern for deploying
# a k3s cluster on AWS with exactly 3 control plane nodes

terraform {
  required_version = ">= 0.14.2"
}

module "k3s_cluster" {
  source = "../aws"
  
  # Standard k3s cluster configuration
  config_git_url = "https://github.com/ubiquitycluster/ubiquity.git"
  config_version = "main"

  cluster_name = "ubiquity-k3s"
  domain       = "example.com"
  image        = "ami-033e6106180a626d0" # CentOS 7 - ca-central-1
  
  # Standard k3s cluster instance configuration
  # This pattern ensures exactly 3 control plane nodes and configurable compute nodes
  instances = {
    mgmt = { 
      type = "t3.medium",
      tags = ["mgmt", "ansible", "public"]
    }
    ctrl = { 
      type = "t3.large",
      tags = ["master", "k8s"]
      count = 3  # Always 3 for HA k3s control plane
    }
    worker = { 
      type = "t3.xlarge",
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
  
  # AWS-specific configuration
  region = var.region
  availability_zone = var.availability_zone
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
  default     = "us-east-1"
  description = "AWS region for deployment"
}

variable "availability_zone" {
  type        = string
  default     = ""
  description = "AWS availability zone (optional, will be auto-selected if not specified)"
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
