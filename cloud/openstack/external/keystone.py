#!/usr/bin/env python3
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

import json
import os

from os.path import expanduser, exists

try:
    from yaml import safe_load
except ImportError:
    yaml_avail = False
else:
    yaml_avail = True

cloudmap = { }
if 'OS_AUTH_URL' in os.environ:
    cloudmap['auth_url'] = os.environ['OS_AUTH_URL']
elif 'OS_CLOUD' in os.environ:
    clouds = os.environ.get(
        'OS_CLIENT_CONFIG_FILE',
        expanduser("~/.config/openstack/clouds.yaml")
    )
    if exists(clouds):
        cloud_name = os.environ['OS_CLOUD']
        if yaml_avail:
            with open(clouds) as f:
                content = safe_load(f)
                if cloud_name in content['clouds']:
                    cloudmap['auth_url'] = content['clouds'][cloud_name]['auth']['auth_url']
        else:
            found = False
            key = cloud_name
            search_level = 0
            with open(clouds) as file:
                for line in file:
                    line_lstripped = line.lstrip()
                    # Skip commented lines
                    if line_lstripped[:1] == '#':
                        continue
                    cur_level = len(line) - len(line_lstripped)
                    if cur_level < search_level:
                        break
                    line_stripped = line_lstripped.rstrip().rstrip(':').strip('"\'')
                    if found and line_stripped[:9] ==  'auth_url:':
                        cloudmap['auth_url'] = line_stripped[9:].strip()
                        break
                    elif line_stripped == key:
                        found = True
                        search_level = cur_level + 1

if 'auth_url' in cloudmap:
    parsed_url = cloudmap['auth_url']
    cloudmap['name'] = parsed_url[8:].split('.')[0]
else:
    cloudmap['auth_url'] = ''
    cloudmap['name'] = ''

print(json.dumps(cloudmap))