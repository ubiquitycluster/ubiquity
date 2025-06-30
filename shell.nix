# https://status.nixos.org
# Copyright The Ubiquity Authors.
#
# Licensed under the Functional Source License, Version 1.0, Apache 2.0 Change License (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/ubiquitycluster/ubiquity/blob/main/LICENSE
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# It also allows for the transition of this software to an Apache 2.0 Licence
# on the second anniversary of the date we make the software available.
# See the License for the specific language governing permissions and
# limitations under the License.

{ pkgs ? import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/refs/tags/22.05.tar.gz") {} }:

let
  python-packages = pkgs.python3.withPackages (p: with p; [
    jinja2
    kubernetes
    netaddr
    python-ipmi
    rich
  ]);
in
pkgs.mkShell {
  buildInputs = with pkgs; [
    ansible
    ansible-lint
    bmake
    diffutils
    docker
    docker-compose
    firefox
    git
    go
    grc
    iproute2
    k9s
    kube3d
    vim
    kubectl
    kubernetes-helm
    kustomize
    libisoburn
    neovim
    openssh
    p7zip
    pre-commit
    tzdata
    shellcheck
    terraform
    yamllint
    yq-go
    asciinema
    envsubst
    apacheHttpd
    ipmitool
    python-packages
  ];
}
