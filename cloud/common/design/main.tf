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
locals {
  domain_name = "${lower(var.cluster_name)}.${lower(var.domain)}"
  
  # Enhanced instance generation with k3s cluster standards
  # Automatically ensure 3 control plane nodes if any instance has master tag
  has_master_nodes = length([for prefix, attrs in var.instances : prefix if contains(attrs.tags, "master")]) > 0
  
  # Standard k3s cluster configuration with enforced 3 control plane nodes
  standard_instances = local.has_master_nodes ? merge(var.instances, {
    ctrl = {
      type = lookup([for prefix, attrs in var.instances : attrs if contains(attrs.tags, "master")][0], "type", "t3.medium")
      tags = ["master", "k8s"]
      count = 3
    }
  }) : var.instances
  
  instances = merge(
    flatten([
      for prefix, attrs in local.standard_instances : [
        for i in range(lookup(attrs, "count", 1)) : {
          (format("%s%d", prefix, i + 1)) = merge(
            { for attr, value in attrs : attr => value if attr != "count" },
            { prefix = prefix }
          )
        }
      ]
    ])...
  )

  # Add validation for k3s cluster requirements
  master_count = length([for key, values in local.instances : key if contains(values.tags, "master")])
  
  instance_per_volume = merge([
    for ki, vi in var.volumes : {
      for kj, vj in vi :
      "${ki}-${kj}" => merge({
        instances = [for x, values in local.instances : x if contains(values.tags, ki)]
      }, vj)
    }
  ]...)

  volumes = merge([
    for key, values in local.instance_per_volume : {
      for instance in values["instances"] :
      "${instance}-${key}" => merge(
        { for key, value in values : key => value if key != "instances" },
      { instance = instance })
    }
  ]...)

  volume_per_instance = transpose({ for key, value in local.instance_per_volume : key => value["instances"] })
}