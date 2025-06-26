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
variable "location" {
  type        = string
  description = "Label of the Azure location where the cluster will be created"
}

variable "azure_resource_group" {
  type        = string
  default     = ""
  description = "Name of an existing resource group that will be used when creating the computing resources. If left empty, terraform will create a new resource group."
}

variable "azurerm_storage_account" {
  type        = string
  default     = ""
  description = "Name of an existing storage account that will be used when creating storage resources. If left empty, terraform will create a new storage account."
}

variable "plan" {
  default = {
    name      = null
    product   = null
    publisher = null
  }
}

locals {
  cloud_provider = "azure"
  cloud_region   = var.location
}