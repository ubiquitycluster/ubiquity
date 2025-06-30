#!/bin/sh
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

agent_ok() { ssh-add -l > /dev/null 2>&1 || [ $? -eq 1 ]; }
if ! agent_ok; then
    if [ -e ~/.ssh/agent.info ]; then
       eval $(< ~/.ssh/agent.info)
    fi
fi
if ! agent_ok; then
    eval $(ssh-agent | tee ~/.ssh/agent.info)
fi
cat ~/.ssh/agent.info

ssh-add ~/.ssh/id_ed25519
