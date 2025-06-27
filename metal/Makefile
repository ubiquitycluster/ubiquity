# Copyright 2023 Logicalis UKI. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0. Previously licensed under the Functional Source License (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/logicalisuki/ubiquity-open/blob/main/LICENSE
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# This software was previously licensed under the Functional Source License but has now transitioned to an Apache 2.0 License
# as of June 2025.
# See the License for the specific language governing permissions and
# limitations under the License.

.POSIX:

env ?= "prod"

default: boot cluster
sandbox: k3d

~/.ssh/id_ed25519:
	ssh-keygen -t ed25519 -f "$@"

ssh:  
	../scripts/ssh-key-setup.sh

ipmitool:
	yum -y install ipmitool

boot: ~/.ssh/id_ed25519 ssh
	ansible-playbook \
		--inventory inventories/${env}.yml \
                --key-file "~/.ssh/id_ed25519" \
		boot.yml

cluster:
	ansible-playbook \
		--inventory inventories/${env}.yml \
                --key-file "~/.ssh/id_ed25519" \
		cluster.yml

console:
	ansible-console \
		--inventory inventories/${env}.yml

k3d:
	command -v k3d || curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
	k3d cluster start ubiquity-dev || k3d cluster create --config k3d-dev.yaml
	k3d kubeconfig get ubiquity-dev > kubeconfig.yaml

clean:
	ansible-playbook \
		--inventory inventories/${env}.yml \
		clean.yml

clean-sandbox:
	k3d cluster delete ubiquity-dev
