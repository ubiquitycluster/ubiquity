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
variable "cluster_name" {
  type        = string
  description = "Name by which this cluster will be known as."
  validation {
    condition     = can(regex("^[a-z][0-9a-z-]*$", var.cluster_name))
    error_message = "The cluster_name value must be lowercase alphanumeric characters and start with a letter. It can include dashes."
  }
}

variable "nb_users" {
  type        = number
  default     = 0
  description = "Number of user accounts with a common password that will be created"
}

variable "instances" {
  description = "Map that defines the parameters for each type of instance of the cluster"
  validation {
    condition     = alltrue(concat([for key, values in var.instances: [contains(keys(values), "type"), contains(keys(values), "tags")]]...))
    error_message = "Each entry in var.instances needs to have at least a type and a list of tags."
  }
  validation {
    condition = sum([for key, values in var.instances: contains(values["tags"], "proxy") ? values["count"] : 0]) < 2
    error_message = "At most one instance in var.instances can have the _proxy_ tag"
  }
  validation {
    condition = sum([for key, values in var.instances: contains(values["tags"], "login") ? 1 : 0]) < 2
    error_message = "At most one type of instances in var.instances can have the _login_ tag"
  }
}

variable "image" {
  type        = any
  description = "Name of the operating system image that will be used to create a boot disk for the instances"
}

variable "volumes" {
  description = "Map that defines the volumes to be attached to the instances"
  validation {
    condition     = length(var.volumes) > 0 ? alltrue(concat([for k_i, v_i in var.volumes: [for k_j, v_j in v_i: contains(keys(v_j), "size")]]...)) : true
    error_message = "Each volume in var.volumes needs to have at least a size attribute."
  }
}

variable "filesystems" {
  type        = map
  description = "Map of cloud provider filesystems to create (i.e: AWS EFS)"
  default     = {}
  validation {
    condition     = var.filesystems != {} ? alltrue([for key, values in var.filesystems : contains(keys(values), "type")]) : true
    error_message = "Each entry in var.filesystems needs to have at least a type."
  }
}

variable "domain" {
  type        = string
  description = "String which when combined with cluster_name will formed the cluster FQDN"
}

variable "public_keys" {
  type        = list(string)
  description = "List of SSH public keys that can log in as {sudoer_username}"
}

variable "guest_passwd" {
  type        = string
  default     = ""
  description = "Guest accounts common password. If left blank, the password is randomly generated."
  validation {
    condition     = length(var.guest_passwd) == 0 || length(var.guest_passwd) >= 8
    error_message = "The guest_passwd value must at least 8 characters long or an empty string."
  }
}

variable "config_git_url" {
  type        = string
  description = "URL to the Ubiquity ansible configuration git repo"
  validation {
    condition     = can(regex("^https://.*\\.git$", var.config_git_url))
    error_message = "The config_git_url variable must be an https url to a git repo."
  }
}

variable "config_version" {
  type        = string
  description = "Tag, branch, or commit that specifies which ansible configuration revision is to be used"
}

variable "ansible_vars" {
  type        = string
  default     = "---"
  description = "String formatted as YAML defining ansible key-value pairs to be included in the ansible environment"
}

variable "sudoer_username" {
  type        = string
  default     = "ubiquity"
  description = "Username of the administrative account"
}

variable "firewall_rules" {
  type = list(
    object({
      name        = string
      from_port   = number
      to_port     = number
      ip_protocol = string
      cidr        = string
    })
  )
  default = [
    {
      "name"        = "SSH",
      "from_port"   = 22,
      "to_port"     = 22,
      "ip_protocol" = "tcp",
      "cidr"        = "0.0.0.0/0"
    },
    {
      "name"        = "HTTP",
      "from_port"   = 80,
      "to_port"     = 80,
      "ip_protocol" = "tcp",
      "cidr"        = "0.0.0.0/0"
    },
    {
      "name"        = "HTTPS",
      "from_port"   = 443,
      "to_port"     = 443,
      "ip_protocol" = "tcp",
      "cidr"        = "0.0.0.0/0"
    },
    #{
    #  "name"        = "Globus",
    #  "from_port"   = 2811,
    #  "to_port"     = 2811,
    #  "ip_protocol" = "tcp",
    #  "cidr"        = "54.237.254.192/29"
    #},
    {
      "name"        = "MyProxy",
      "from_port"   = 7512,
      "to_port"     = 7512,
      "ip_protocol" = "tcp",
      "cidr"        = "0.0.0.0/0"
    },
    {
      "name"        = "GridFTP",
      "from_port"   = 50000,
      "to_port"     = 51000,
      "ip_protocol" = "tcp",
      "cidr"        = "0.0.0.0/0"
    }
  ]
  description = "List of login external firewall rules defined as map of 5 values name, from_port, to_port, ip_protocol and cidr"
}

variable "generate_ssh_key" {
  type        = bool
  default     = false
  description = "If set to true, Terraform will generate an ssh key pair to connect to the cluster. Default: false"
}

variable "software_stack" {
  type        = string
  default     = "eessi"
  description = "Provider of research computing software stack (can be 'ubiquity' or 'eessi')"
}

variable "pool" {
  default = []
}

#variable "azure_subscription_id" {
#  type = string
#}

#variable "site" {
#  type      = string
#  default   = "ubiqcluster"
#}

#variable "azure_tenant_id" {
#  type      = string
 # sensitive = true
#}

variable "storage_account_name" {
  type      = string
  default   = "ubiqcluster"
}

variable "proximity_placement_group" {
  type = object({
    new  = bool
    name = string
  })

  default = {
    new  = true
    name = "existing-proximity-placement-group"
  }

  description = "Proximity placement group options"

  validation {
    condition     = contains([false, true], var.proximity_placement_group.new)
    error_message = "The proximity_placement_group.new value should be false or true."
  }

  validation {
    condition     = can(regex("^[0-9A-Za-z_-]{1,64}$", var.proximity_placement_group.name))
    error_message = "The proximity_placement_group.name value should be alphanumeric characters and 1-64 characters long."
  }
}

variable "storage_account" {
  type      = string
  default   = "ubiqcluster"
}
#variable "nb_public_ip" {
#  description = "Number of public IPs to assign corresponding to one IP per vm. Set to 0 to not assign any public IP addresses."
#  default     = "1"
#}

#variable "public_ip_dns" {
#  description = "Optional globally unique per datacenter region domain name label to apply to each public ip address. e.g. thisvar.varlocation.cloudapp.azure.com where you specify only thisvar here. This is an array of names which will pair up sequentially to the number of public ips defined in var.nb_public_ip. One name or empty string is required for every public ip. If no public ip is desired, then set this to an array with a single empty string."
#  default     = ["ubiqmaster1","ubiqmaster2"]
#}

variable "default_tags" {
    default = {}
}

variable "infrastructure_id" { default = "ubiquity" }

variable "private_vpc_id" {}

# Subnet Details
variable "private_vpc_private_subnet_ids" {
    description = "List of subnet ids"
    type        = list(string)
}
