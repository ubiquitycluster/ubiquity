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

.POSIX:

default: apply

cloud: azure

azure:
	./scripts/ovhcli.sh
	./scripts/ovhlogin.sh

~/.terraform.d/credentials.tfrc.json:
	terraform login

terraform.tfvars:
	ansible-playbook tfvars.yml

.terraform.lock.hcl: ~/.terraform.d/credentials.tfrc.json versions.tf terraform.tfvars
	terraform init
	touch .terraform.lock.hcl

plan: .terraform.lock.hcl
	terraform plan

apply: azure .terraform.lock.hcl
	terraform apply -auto-approve
	terraform apply -auto-approve
	./scripts/prepareAnsible.sh

clean:
	terraform destroy -auto-approve
