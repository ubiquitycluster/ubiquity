#!/bin/sh
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

# Remove the values.yaml file if it already exists
rm -f values.yaml || true

# Combine the values-seed.yaml and defaults.yaml files together into a single values.yaml file
yq eval-all '. as $item ireduce ({}; . *+ $item)' values-seed.yaml defaults.yaml > values.yaml

# Render the argocd chart and apply it to the cluster
helm template \
    --include-crds \
    --namespace argocd \
    argocd . \
    | kubectl apply -n argocd -f -

# Wait for the argo crds to be created
kubectl -n argocd wait --timeout=60s --for condition=Established \
       crd/applications.argoproj.io \
       crd/applicationsets.argoproj.io

# Patch the argocd configmap to include the build options for the kustomize plugin
kubectl patch configmap argocd-cm --namespace argocd --patch "$(cat argocd-cm.kustomize-buildoptions.patch.yaml)"

# Patch the argocd configmap to include the tls cert for the repo git repo
kubectl patch configmap argocd-tls-certs-cm --namespace argocd --patch "$(cat argocd-repo-tls-cert.yaml)"
