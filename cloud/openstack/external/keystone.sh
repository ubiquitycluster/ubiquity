#!/bin/bash
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