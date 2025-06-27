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
.PHONY: *
.EXPORT_ALL_VARIABLES:

# -------------------------------------------------------------------------------------------------
# OS Info
# -------------------------------------------------------------------------------------------------
DISTRO_VER := $(shell egrep '^ID=' /etc/os-release | awk -F= '{print $2}' | tr -d '"' | tr -d 'ID=')

# -------------------------------------------------------------------------------------------------
# Docker configuration
# -------------------------------------------------------------------------------------------------
DIR = Dockerfiles/images-build
FILE = Dockerfile
IMAGE = cjcshadowsan/hpc-ubiq
TAG = latest
NO_CACHE =

# Version & Flavour
BUILDVER = latest
FLAVOUR = slurm

# --------------------------------------
# Docker Targets
# --------------------------------------

#.PHONY: docker-build
#docker-build: generate manifests ## Build the docker image
#	docker build . -t ${IMG} --build-arg http_proxy=$(http_proxy) --build-arg https_proxy=$(https_proxy)

# Push the docker image
#.PHONY: docker-push
#docker-push:
#	docker push ${IMG}

# -------------------------------------------------------------------------------------------------
# Default Target
# -------------------------------------------------------------------------------------------------

KUBECONFIG = $(shell pwd)/metal/kubeconfig.yaml
KUBE_CONFIG_PATH = $(KUBECONFIG)

test1: metal bootstrap wait post-install
default: metal bootstrap external wait post-install
ucl: cluster bootstrap external wait post-install
demo_onprem: metal bootstrap external wait post-install
demo_azure: azure cluster bootstrap external wait post-install
demo_aws: aws cluster bootstrap external wait post-install
sandbox: sandbox-boot bootstrap

aws:
	make -C cloud awscloud

aws-clean:
	make -C cloud awsclean

azure:
	make -C cloud azurecloud

azure-clean:
	make -C cloud azureclean

configure-sandbox:
	./scripts/configure-sandbox

configure:
	./scripts/configure

sandbox-boot:
	make -C metal sandbox

metal:
	make -C metal

cluster: 
	make -C metal cluster

bmo:
	make -C baremetal-operator

bootstrap:
	make -C bootstrap

storage:
	make -C storage

external:
	make -C external

wait:
	sleep 60; ./scripts/wait-main-apps

post-install:
	./scripts/hacks
	./scripts/bmo/bmo-create.sh

docker:
#	./scripts/get-docker.sh
#	./scripts/docker-perms.sh

podman:
	sudo yum -y install podman
	sudo systemctl enable --now podman.socket
	sudo test -L /var/run/docker.sock && echo "link exists" || ln -s /run/podman/podman.sock /var/run/docker.sock
nat:
	./scripts/create-nat.sh

ifeq ($(DISTRO_VER),rhel)
RUNTIME=podman
tools: podman opus
else ifeq ($(DISTRO_VER),rocky)
RUNTIME=podman
tools: podman opus
else ifeq ($(DISTRO_VER),almalinux)
RUNTIME=podman
tools: podman opus
else ifeq ($(DISTRO_VER),ubuntu)
RUNTIME=docker
tools: docker opus
else ifeq ($(DISTRO_VER),debian)
RUNTIME=docker
tools: docker opus
else
RUNTIME=docker
tools: docker opus
endif

opus:
	mkdir -p ${HOME}/.terraform.d
	$(RUNTIME) run \
		--rm \
		--interactive \
		--tty \
		--privileged \
		--network host \
		--env "KUBECONFIG=${KUBECONFIG}" \
		--volume $(shell pwd):$(shell pwd) \
		--volume ${HOME}/.ssh:/root/.ssh \
		--volume "/var/run/docker.sock:/var/run/docker.sock" \
		--volume ${HOME}/.terraform.d:/root/.terraform.d \
		--volume ubiquity-tools-cache:/root/.cache \
		--volume ubiquity-tools-aws:/data \
		--workdir $(shell pwd) \
		ubiquity.azurecr.io/opus:latest-opus-main /bin/bash

nixos:  
	#export PATH=/usr/bin:$PATH
	#export DOCKER_HOST=unix:///run/user/1000/docker.sock
	docker run \
		--rm \
		--interactive \
		--tty \
		--privileged \
		--network host \
		--env "KUBECONFIG=${KUBECONFIG}" \
		--volume $(shell pwd):$(shell pwd) \
		--volume ${HOME}/.ssh:/root/.ssh \
		--volume ${HOME}/.terraform.d:/root/.terraform.d \
		--volume ubiquity-tools-cache:/root/.cache \
		--volume ubiquity-tools-nix:/nix \
		--workdir $(shell pwd) \
		nixos/nix nix-shell

test:
	make -C test

dev: metal bootstrap wait post-install

docs:
	docker run \
		--rm \
		--interactive \
		--tty \
		--publish 8000:8000 \
		--volume $(shell pwd):/docs \
		squidfunk/mkdocs-material

git-hooks:
	pre-commit install

clean:
	make -C metal clean

clean-sandbox:
	make -C metal clean-sandbox
# -------------------------------------------------------------------------------------------------
# Build Targets
# -------------------------------------------------------------------------------------------------

build:
	@ \
	if [ "$(FLAVOUR)" = "slurm" ]; then \
		docker build \
			$(NO_CACHE) \
			--label "org.opencontainers.image.created"="$$(date --rfc-3339=s)" \
			--label "org.opencontainers.image.revision"="$$(git rev-parse HEAD)" \
			--label "org.opencontainers.image.version"="${VERSION}" \
			--build-arg VERSION=$(BUILDVER) \
			-t $(IMAGE):slurmprobe-$(BUILDVER) -f $(DIR)/slurmprobe/$(FILE) $(DIR)/slurmprobe; \
		docker build \
                        $(NO_CACHE) \
                        --label "org.opencontainers.image.created"="$$(date --rfc-3339=s)" \
                        --label "org.opencontainers.image.revision"="$$(git rev-parse HEAD)" \
                        --label "org.opencontainers.image.version"="${VERSION}" \
                        --build-arg VERSION=$(BUILDVER) \
                        -t $(IMAGE):slurmmunge-$(BUILDVER) -f $(DIR)/slurmmunge/$(FILE) $(DIR)/slurmmunge; \
		docker build \
                        $(NO_CACHE) \
                        --label "org.opencontainers.image.created"="$$(date --rfc-3339=s)" \
                        --label "org.opencontainers.image.revision"="$$(git rev-parse HEAD)" \
                        --label "org.opencontainers.image.version"="${VERSION}" \
                        --build-arg VERSION=$(BUILDVER) \
                        -t $(IMAGE):slurmconf-$(BUILDVER) -f $(DIR)/slurmconf/$(FILE) $(DIR)/slurmconf; \
                docker build \
                        $(NO_CACHE) \
                        --label "org.opencontainers.image.created"="$$(date --rfc-3339=s)" \
                        --label "org.opencontainers.image.revision"="$$(git rev-parse HEAD)" \
                        --label "org.opencontainers.image.version"="${VERSION}" \
                        --build-arg VERSION=$(BUILDVER) \
                        -t $(IMAGE):slurmcontainer-$(BUILDVER) -f $(DIR)/slurmcontainer/$(FILE) $(DIR)/slurmcontainer; \
	elif [ "$(FLAVOUR)" = "openpbs" ]; then \
                docker build \
                        $(NO_CACHE) \
                        --label "org.opencontainers.image.created"="$$(date --rfc-3339=s)" \
                        --label "org.opencontainers.image.revision"="$$(git rev-parse HEAD)" \
                        --label "org.opencontainers.image.version"="${VERSION}" \
                        --build-arg VERSION=$(BUILDVER) \
                        -t $(IMAGE):openpbsconf-$(BUILDVER) -f $(DIR)/openpbsconf/$(FILE) $(DIR)/openpbsconf; \
                docker build \
			$(NO_CACHE) \
			--label "org.opencontainers.image.created"="$$(date --rfc-3339=s)" \
			--label "org.opencontainers.image.revision"="$$(git rev-parse HEAD)" \
                        --label "org.opencontainers.image.version"="${VERSION}" \
                        --build-arg VERSION=$(BUILDVER) \
                        -t $(IMAGE):openpbsinit-$(BUILDVER) -f $(DIR)/openpbsinit/$(FILE) $(DIR)/openpbsinit; \
		docker build \
			$(NO_CACHE) \
			--label "org.opencontainers.image.created"="$$(date --rfc-3339=s)" \
			--label "org.opencontainers.image.revision"="$$(git rev-parse HEAD)" \
			--label "org.opencontainers.image.version"="${VERSION}" \
			--build-arg VERSION=$(BUILDVER) \
			-t $(IMAGE):openpbs-$(BUILDVER) -f $(DIR)/openpbs/$(FILE) $(DIR)/openpbs; \
	elif [ "$(FLAVOUR)" = "oar" ]; then \
		docker build \
			$(NO_CACHE) \
			--label "org.opencontainers.image.created"="$$(date --rfc-3339=s)" \
			--label "org.opencontainers.image.revision"="$$(git rev-parse HEAD)" \
			--label "org.opencontainers.image.version"="${VERSION}" \
			--build-arg VERSION=$(BUILDVER) \
			-t $(IMAGE):oarconf-$(BUILDVER) -f $(DIR)/oarconf/$(FILE) $(DIR)/oarconf; \
		docker build \
			$(NO_CACHE) \
			--label "org.opencontainers.image.created"="$$(date --rfc-3339=s)" \
			--label "org.opencontainers.image.revision"="$$(git rev-parse HEAD)" \
			--label "org.opencontainers.image.version"="${VERSION}" \
			--build-arg VERSION=$(BUILDVER) \
			-t $(IMAGE):oarinit-$(BUILDVER) -f $(DIR)/oarinit/$(FILE) $(DIR)/oarinit; \
		docker build \
			$(NO_CACHE) \
			--label "org.opencontainers.image.created"="$$(date --rfc-3339=s)" \
			--label "org.opencontainers.image.revision"="$$(git rev-parse HEAD)" \
			--label "org.opencontainers.image.version"="${VERSION}" \
			--build-arg VERSION=$(BUILDVER) \
			-t $(IMAGE):oar-$(BUILDVER) -f $(DIR)/oar/$(FILE) $(DIR)/oar; \
	else \
		echo "not building anything"; \
	fi

rebuild: NO_CACHE=--no-cache
rebuild: pull-base-image
rebuild: build
