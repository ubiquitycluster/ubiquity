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
variable "project" {
}

variable "zone_name" {
}

variable "name" {
}

variable "domain" {
}

variable "email" {
}

variable "acme_key_pem" {
  type = string
  default = ""
}

variable "sudoer_username" {
}

variable "domain_tag" {
  description = "Indicate which tag the instances that will be pointed by the domain name A record has to have."
  default     = "login"
}

variable "vhost_tag" {
  description = "Indicate which tag the instances that will be pointed by the vhost A record has to have."
  default = "proxy"
}

variable "ssl_tags" {
  description = "Indicate which tag the instances that will receive a copy of the wildcard SSL certificate has to have."
  default = ["proxy", "ssl"]
}

variable "public_instances" { }

variable "ssh_private_key" {
  type = string
}