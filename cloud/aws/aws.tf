# Copyright 2023 Logicalis UKI. All Rights Reserved.
#
# Licensed under the Functional Source License, Version 1.0, Apache 2.0 Change License (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/logicalisuki/ubiquity/blob/main/LICENSE.md
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# It also allows for the transition of this software to an Apache 2.0 Licence
# on the second anniversary of the date we make the software available.
# See the License for the specific language governing permissions and
# limitations under the License.
variable "region" {
  type        = string
  description = "Label for the AWS physical location where the cluster will be created"
}

variable "availability_zone" {
  default     = ""
  description = "Label of the datacentre inside the AWS region where the cluster will be created. If left blank, it chosen at random amongst the zones that are available."
}

locals {
  cloud_provider = "aws"
  cloud_region   = var.region
}