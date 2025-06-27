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
variable "os_floating_ips" {
  default = {}
}

variable "os_ext_network" {
  type    = string
  default = null
}

variable "subnet_id" {
  type    = string
  default = null
}

data "external" "keystone" {
  program = ["bash", "${path.module}/external/keystone.sh"]
}

locals {
  cloud_provider = "openstack"
  cloud_region   = data.external.keystone.result.name
}