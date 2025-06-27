# Copyright 2023 Logicalis UKI. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0. Previously licensed under the Functional Source License (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/logicalisuki/ubiquity-open-open/blob/main/LICENSE
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# This software was previously licensed under the Functional Source License but has now transitioned to an Apache 2.0 License
# as of June 2025.
# See the License for the specific language governing permissions and
# limitations under the License.
data "cloudflare_zones" "domain" {
  filter {
    name   = var.domain
    status = "active"
    paused = false
  }
}

module "record_generator" {
  source         = "../record_generator"
  name           = lower(var.name)
  public_instances = var.public_instances
  domain_tag       = var.domain_tag
  vhost_tag        = var.vhost_tag
}

resource "cloudflare_record" "records" {
  count   = length(module.record_generator.records)
  zone_id = data.cloudflare_zones.domain.zones[0].id
  name    = module.record_generator.records[count.index].name
  value   = module.record_generator.records[count.index].value
  type    = module.record_generator.records[count.index].type
  dynamic "data" {
    for_each = module.record_generator.records[count.index].data != null ? [module.record_generator.records[count.index].data] : []
    content {
      algorithm   = data.value["algorithm"]
      fingerprint = data.value["fingerprint"]
      type        = data.value["type"]
    }
  }
}

module "acme" {
  source           = "../acme"
  dns_provider     = "cloudflare"
  name             = lower(var.name)
  domain           = var.domain
  email            = var.email
  sudoer_username  = var.sudoer_username
  public_instances = var.public_instances
  ssh_private_key  = var.ssh_private_key
  ssl_tags         = var.ssl_tags
  acme_key_pem     = var.acme_key_pem
}

output "hostnames" {
  value = distinct(compact([for record in module.record_generator.records : join(".", [record.name, var.domain]) if record.type == "A" ]))
}