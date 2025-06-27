# Copyright 2023 Logicalis UKI. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0. Previously licensed under the Functional Source License.
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

default: dragonfly-ns namespace argocd root

argocd/charts: argocd/Chart.yaml
	cd argocd \
		&& helm dependency update

namespace:
	kubectl create namespace argocd --dry-run=client --output=yaml \
		| kubectl apply -f -

dragonfly/charts: ../dragonfly-system/Chart.yaml
	cd ../dragonfly-system \
		&& helm dependency update \
	        && cd -

dragonfly-ns:
	kubectl create namespace dragonfly-system --dry-run=client --output=yaml \
		| kubectl apply -f -

.PHONY: argocd
argocd: argocd/charts
	cd argocd && \
		./apply.sh

.PHONY: root
root:
	cd root && \
		./apply.sh

#.PHONY: dragonfly
#root: dragonfly/charts
#	cd ../dragonfly-system && \
                ./apply.sh && \
		cd -
