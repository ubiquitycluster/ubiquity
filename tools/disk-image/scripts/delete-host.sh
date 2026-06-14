#!/bin/bash
# Copyright The Ubiquity Authors.
#
# Licensed under the Apache License, Version 2.0. Previously licensed under the Functional Source License (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/ubiquitycluster/ubiquity/blob/main/LICENSE
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# This software was previously licensed under the Functional Source License but has now transitioned to an Apache 2.0 License
# as of June 2025.
# See the License for the specific language governing permissions and
# limitations under the License.

# NVIDIA Infra Controller reset: this legacy BareMetalHost delete helper is
# fallback/migration-only for sites that have not moved day-2 bare-metal
# lifecycle ownership to NVIDIA Infra Controller. This is fallback/migration-only;
# do not use for new day-2 lifecycle automation. Prefer NVIDIA Infra Controller
# Task workflows.

BMHOST=$1

if [ -z "${BMHOST}" ] ; then
    echo "Usage: delete_host.sh <BareMetalHost Name>"
    exit 1
fi

kubectl patch secret "${BMHOST}-bmc-secret" -n metal-nodes --type merge \
    -p '{"metadata":{"finalizers":[]}}'

kubectl patch baremetalhost "${BMHOST}" -n metal-nodes --type merge \
    -p '{"metadata":{"finalizers":[]}}'