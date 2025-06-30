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

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

if [ -x "$(command -v "python3")" ]; then
    python3 ${SCRIPT_DIR}/keystone.py
elif [ -x "$(command -v "python2")" ]; then
    python2 ${SCRIPT_DIR}/keystone.py
elif [ -x "$(command -v "python")" ]; then
    python ${SCRIPT_DIR}/keystone.py
else
    # Python could not be found that's... unfortunate
    # We will have to do the job in shell script instead, ugh!
    local s="${AUTH_URL/#*:\/\/}"
    name="${s%%.*}"
    echo "{ \"auth_url\" : \"${OS_AUTH_URL}\" , \"name\":  \"${name}\" }"
fi