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
data "google_dns_managed_zone" "domain" {
  name    = var.zone_name
  project = var.project
}

module "record_generator" {
  source         = "../record_generator"
  name           = lower(var.name)
  public_instances = var.public_instances
  domain_tag       = var.domain_tag
  vhost_tag        = var.vhost_tag
}

resource "google_dns_record_set" "records" {
  count        = length(module.record_generator.records)
  managed_zone = data.google_dns_managed_zone.domain.name
  project      = var.project
  name         = join(".", [module.record_generator.records[count.index].name, var.domain, ""])
  type         = module.record_generator.records[count.index].type
  rrdatas      = [module.record_generator.records[count.index].type != "SSHFP" ?
                  module.record_generator.records[count.index].value : 
                  join(" ", [module.record_generator.records[count.index].data["algorithm"],
                             module.record_generator.records[count.index].data["type"],
                             module.record_generator.records[count.index].data["fingerprint"]])
                 ]
  ttl          = 300
}

module "acme" {
  source              = "../acme"
  dns_provider        = "gcloud"
  dns_provider_config = {
    GCE_PROJECT = var.project
  }
  name                = lower(var.name)
  domain              = var.domain
  email               = var.email
  sudoer_username     = var.sudoer_username
  public_instances    = var.public_instances
  ssh_private_key     = var.ssh_private_key
  ssl_tags            = var.ssl_tags
  acme_key_pem        = var.acme_key_pem
}

output "hostnames" {
  value = distinct(compact([for record in module.record_generator.records : join(".", [record.name, var.domain]) if record.type == "A" ]))
}