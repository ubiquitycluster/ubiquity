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
module "record_generator" {
  source         = "../record_generator"
  name           = lower(var.name)
  public_instances = var.public_instances
  domain_tag       = var.domain_tag
  vhost_tag        = var.vhost_tag
}

resource "local_file" "dns_record" {
    content     = <<EOT
; Import this file in ${var.domain} DNS zone
%{ for record in module.record_generator.records ~}
${record.name}.${var.domain}.   1   IN  ${record.type}  %{ if record.value != null }${record.value}%{ else }${record.data["algorithm"]} ${record.data["type"]}  ${record.data["fingerprint"]}%{ endif }
%{ endfor ~}
EOT
    filename = "${var.name}.${var.domain}.txt"
    file_permission = "0600"
}

output "hostnames" {
  value = distinct(compact([for record in module.record_generator.records : join(".", [record.name, var.domain]) if record.type == "A" ]))
}