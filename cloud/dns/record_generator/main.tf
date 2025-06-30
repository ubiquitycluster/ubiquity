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
variable "name" {
}

variable "vhosts" {
    type    = list(string)
    default = ["ipa", "jupyter", "mokey", "explore"]
}

variable "public_instances" {}

variable "domain_tag" {}
variable "vhost_tag" {}

data "external" "key2fp" {
  for_each = var.public_instances
  program = ["bash", "${path.module}/key2fp.sh"]
  query = {
    rsa = each.value["hostkeys"]["rsa"]
    ed25519 = each.value["hostkeys"]["ed25519"]
  }
}

locals {
    records = concat(
    [
        for key, values in var.public_instances: {
            type = "A"
            name = join(".", [key, var.name])
            value = values["public_ip"]
            data = null
        }
    ],
    flatten([
        for key, values in var.public_instances: [
            for vhost in var.vhosts:
            {
                type  = "A"
                name  = join(".", [vhost, var.name])
                value = values["public_ip"]
                data  = null
            }
        ]
        if contains(values["tags"], var.vhost_tag)
    ]),
    [
        for key, values in var.public_instances: {
            type  = "A"
            name  = var.name
            value = values["public_ip"]
            data  = null
        }
        if contains(values["tags"], var.domain_tag)
    ],
    [
        for key, values in var.public_instances: {
            type  = "SSHFP"
            name  = join(".", [key, var.name])
            value = null
            data  = {
                algorithm   = data.external.key2fp[key].result["rsa_algorithm"]
                type        = 2
                fingerprint = data.external.key2fp[key].result["rsa_sha256"]
            }
        }
    ],
    [
        for key, values in var.public_instances: {
            type  = "SSHFP"
            name  = join(".", [key, var.name])
            value = null
            data  = {
                algorithm   = data.external.key2fp[key].result["ed25519_algorithm"]
                type        = 2
                fingerprint = data.external.key2fp[key].result["ed25519_sha256"]
            }
        }
    ],
    [
         {
            type  = "SSHFP"
            name  = var.name
            value = null
            data  = {
                algorithm   = try(coalesce([for key, values in var.public_instances: data.external.key2fp[key].result["rsa_algorithm"] if contains(values["tags"], var.domain_tag)]...), 0)
                type        = 2
                fingerprint = try(coalesce([for key, values in var.public_instances: data.external.key2fp[key].result["rsa_sha256"] if contains(values["tags"], var.domain_tag)]...), 0)
            }
        }
    ],
    [
         {
            type  = "SSHFP"
            name  = var.name
            value = null
            data  = {
                algorithm   = try(coalesce([for key, values in var.public_instances: data.external.key2fp[key].result["ed25519_algorithm"] if contains(values["tags"], var.domain_tag)]...), 0)
                type        = 2
                fingerprint = try(coalesce([for key, values in var.public_instances: data.external.key2fp[key].result["ed25519_sha256"] if contains(values["tags"], var.domain_tag)]...), 0)
            }
        }
    ])
}

output "records" {
    value = local.records
}