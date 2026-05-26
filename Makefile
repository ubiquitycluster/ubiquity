# The ubiquity CLI is the recommended entry point. Run 'ubiquity up' instead.
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

# -------------------------------------------------------------------------------------------------
# Default Target
# -------------------------------------------------------------------------------------------------

KUBECONFIG = $(shell pwd)/metal/kubeconfig.yaml
KUBE_CONFIG_PATH = $(KUBECONFIG)

test1: metal bootstrap wait post-install
default:
	@echo "============================================"
	@echo "  Ubiquity - HPC Cluster Lifecycle Platform"
	@echo "  The CLI is the recommended entry point."
	@echo "  Run: ubiquity up --sandbox"
	@echo "============================================"
	@exit 0
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

# configure: delegates to 'ubiquity configure' (Go CLI)
configure-sandbox:
	./scripts/configure-sandbox.py.bak

# configure: delegates to 'ubiquity configure' (Go CLI)
configure:
	./scripts/configure.py.bak

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
		ghcr.io/ubiquitycluster/opus:latest-opus-all-helm3.10 /bin/bash

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

# Development workflow
dev: cli
	@echo "============================================"
	@echo "  Development environment ready"
	@echo "  Run: ./ubiquity-cli up --sandbox"
	@echo "============================================"

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
# CLI Targets
# -------------------------------------------------------------------------------------------------

# Build the ubiquity CLI binary with version info
cli:
	go build -ldflags "-X github.com/ubiquitycluster/ubiquity/cmd/ubiquity/cmd.Version=latest -X github.com/ubiquitycluster/ubiquity/cmd/ubiquity/cmd.Commit=$$(git rev-parse --short HEAD 2>/dev/null) -X github.com/ubiquitycluster/ubiquity/cmd/ubiquity/cmd.Date=$$(date -u +%Y-%m-%d)" -o ubiquity-cli ./cmd/ubiquity/
	@echo "Built: ubiquity-cli"

install: cli
	install -m 755 ubiquity-cli /usr/local/bin/ubiquity
# -------------------------------------------------------------------------------------------------
# Build Targets
# -------------------------------------------------------------------------------------------------

build:
	@VERSION=$(BUILDVER) COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "none") DATE=$$(date --rfc-3339=s 2>/dev/null || date -u +%Y-%m-%dT%H:%M:%SZ); \
	export VERSION COMMIT DATE; \
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

# Build the PXE installer binary
installer:
	cd tools && go build -o ../ubiquity-installer ./cmd/ubiquity-install/

# Record a demo asciicast
demo:
	@echo "Recording demo... Press Ctrl+D when done."
	asciinema rec ubiquity-demo.cast -c "./ubiquity-cli up --sandbox --skip-security 2>&1 | head -20"
	@echo "Demo saved to ubiquity-demo.cast"
	@echo "Upload to https://asciinema.org or replay locally with 'asciinema play ubiquity-demo.cast'"

# ── Multi-version k3s test matrix ──────────────────────────────────────────
.PHONY: test-k3d-matrix test-k3d-v1.30 test-k3d-v1.31 test-k3d-v1.32

# Run the full k3s version matrix test
test-k3d-matrix:
	bash test/k3d-matrix.sh

# Test a specific k3s version
test-k3d-v1.30:
	K3S_IMAGE=rancher/k3s:v1.30.14-k3s2 bash test/k3d-matrix.sh

test-k3d-v1.31:
	K3S_IMAGE=rancher/k3s:v1.31.14-k3s1 bash test/k3d-matrix.sh

test-k3d-v1.32:
	K3S_IMAGE=rancher/k3s:v1.32.13-k3s1 bash test/k3d-matrix.sh

# Show available targets
help:
	@echo "Available targets:"
	@echo "  cli         Build the ubiquity CLI binary"
	@echo "  installer   Build the PXE installer binary"
	@echo "  completions Generate shell completion files"
	@echo "  test        Run tests"
	@echo "  dev         Prepare development environment"
	@echo "  install     Install CLI to /usr/local/bin"
