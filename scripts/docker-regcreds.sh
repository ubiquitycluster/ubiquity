#!/usr/bin/env bash
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

set -eu

echo "Please enter your dockerhub registry username:"
read -r username
echo "Please enter your dockerhub registry password:"
read -r -s password

for namespace in $(kubectl get namespaces -o=json | jq -r ".items | .[] | .metadata.name"); do
    kubectl delete --namespace "$namespace" secret regcred || true
    kubectl create secret --namespace "$namespace" docker-registry regcred \
        --docker-server=docker.io \
        --docker-username="$username" \
        --docker-password="$password" \
        --docker-email=unused
done

for namespace in $(kubectl get namespaces -o=json | jq -r ".items | .[] | .metadata.name"); do
    kubectl apply --namespace "$namespace" -f .dockercreds.yaml
    for serviceaccount in $(kubectl get serviceaccounts --namespace "$namespace" -o=json | jq -r ".items | .[] | .metadata.name"); do
        kubectl patch serviceaccount --namespace "$namespace" "$serviceaccount" -p '{"imagePullSecrets": [{"name": "regcred"}]}'
    done
done