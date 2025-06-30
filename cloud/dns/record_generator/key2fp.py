#!/usr/bin/env python3
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

import base64
import hashlib
import json
import sys

ALGORITHMS = {
    'ssh-rsa' : '1',
    'ssh-dss' : '2',
    'ssh-ecdsa' : '3',
    'ssh-ed25519' : '4'
}

outputs = {}

inputs = json.load(sys.stdin)

for alg, ssh_key in inputs.items():
    key_type, key = ssh_key.split()
    key_bytes = base64.b64decode(key)
    alg_index = ALGORITHMS[key_type]
    fp_sha256 = hashlib.sha256(key_bytes).hexdigest()

    outputs['{alg}_algorithm'.format(alg=alg)] = alg_index
    outputs['{alg}_sha256'.format(alg=alg)] = fp_sha256

print(json.dumps(outputs))