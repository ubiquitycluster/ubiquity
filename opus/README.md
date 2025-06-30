# Docker image for `OPUS`

[![Build Status](https://github.com/ubiquitycluster/ubiquity/workflows/lint/badge.svg)](https://github.com/ubiquitycluster/ubiquity/actions?workflow=lint)
[![Build Status](https://github.com/ubiquitycluster/ubiquity/workflows/build/badge.svg)](https://github.com/ubiquitycluster/ubiquity/actions?workflow=build)
[![Build Status](https://github.com/ubiquitycluster/ubiquity/workflows/build-kops/badge.svg)](https://github.com/ubiquitycluster/ubiquity/actions?workflow=build-kops)
[![Build Status](https://github.com/ubiquitycluster/ubiquity/workflows/build-helm/badge.svg)](https://github.com/ubiquitycluster/ubiquity/actions?workflow=build-helm)

## 🚀 GitHub Container Registry

Opus images are now available on **GitHub Container Registry (GHCR)** for improved accessibility and integration:

```bash
# Pull the latest base image
docker pull ghcr.io/ubiquity/opus:latest

# Pull specific flavours
docker pull ghcr.io/ubiquity/opus:latest-aws
docker pull ghcr.io/ubiquity/opus:latest-tools
docker pull ghcr.io/ubiquity/opus:latest-opus-all-helm3.10
```

**Available registries:**
- 🆕 **GitHub Container Registry**: `ghcr.io/ubiquity/opus:*` (Public, Free)

**Benefits of GHCR:**
- ✅ Free public hosting
- ✅ Integrated with GitHub permissions
- ✅ Automated builds via GitHub Actions
- ✅ Security scanning with Trivy
- ✅ Multi-architecture support

> See [GHCR-SETUP.md](./GHCR-SETUP.md) for detailed setup and usage instructions.


> #### All [#ubiquity-hpc](https://github.com/topics/ubiquity-hpc) Docker images

> #### All [#ubiquity-hpc](https://github.com/topics/ubiquity-hpc) Makefile able to b/eploying ur c

View **[Dockerfile](https://github.com/ubiquitycluster/ubiquity/blob/main/opus/Dockerfile)** on GitHub.

Opus is a RockyLinux-based multistage-build dockerized version of the setup/administration tools for Ubiquity<sup>[1]</sup> in many different flavours.

Opus is the "creation" container for Ubiquity, a means of deploying and managing Ubiquity environments in a consistent manner. All functions are included within the container, and it is designed to be run from a bastion host or a jump box.

It contains Ansible, Terraform, Rust, Golang, K9s, and both AZ cli and awscli tools as well as ArgoCD, Helm, Kops, Kubectl, and Openshift CLI tools.

It comes with **[Mitogen](https://github.com/dw/mitogen)**<sup>[2]</sup> to speed up your runs by up to **600%**<sup>[3][4]</sup> (see [Examples](#run-ansible-playbook-with-mitogen)).
The image is built weekly against multiple stable versions and pushed to Azure CR (for enterprise Logicalis UKI customers).

* <sup>[1] Official project: https://github.com/localhost/ubiquity</sup>
* <sup>[2] Official project: https://github.com/dw/mitogen</sup>
* <sup>[3] [How to Speed Up Your Ansible Playbooks Over 600%](https://www.toptechskills.com/ansible-tutorials-courses/speed-up-ansible-playbooks-pipelining-mitogen/)</sup>
* <sup>[4] [Mitogen for Ansible](https://networkgenomics.com/ansible/)</sup>


## Available Docker image versions

This repository provides many different Ansible flavours (each flavour also divided into different Ansible versions).

The following tree shows how the different flavours derive from each other (each child has all the tools and features of its parent plus its own additions).

For Ubiquity however, simply get the latest version of the `opus` flavour.
```css
       base                    #docker-tag:  :latest
         |                                   :<version>
         |
       tools                   #docker-tag:  :latest-tools
      /  |  \                                :<version>-tools
     /   |   \
infra  azure  aws              #docker-tag:  :latest-infra     :latest-azure     :latest-aws
  |     /      |                             :<version>-infra  :<version>-azure  :<version>-aws
  |    /       |
   \  /     awsk8s            #docker-tag:  :latest-awsk8s
    \/        /  \                           :<version>-awsk8s
     \       /    \
      \  awskops  awshelm     #docker-tag   :latest-awskops     :latest-awshelm
       \             \                      :<version>-awskops  :<version>-awshelm
        \             \
         \___________ opus    #docker-tag   :latest-opus
                                            :<version>-opus
                                            
```
> <sub>`<version>` refers to the latest<sup>\[1\],</sup> patch-level version of Ansible. E.g.: `2.9`, `2.10`, `2.11`, ...</sub><br/>
> <sub>\[1\]: latest as docker images are (re)built every night via CI against the latest available patch level version of Ansible</sub>


The following table shows a quick overview of provided libraries and tools for each flavour. For more details see further down below.

| Flavour | Based on | Additional Python libs | Additional binaries |
|---------|---------------|------------------------|---------------------|
| base    | -        | `cffi`, `cryptography`, `Jinja2`, `junit-xml`, `lxml`, `paramiko`, `PyYAML` | - |
| tools   | base     | `firewall`, `dnspython`, `mitogen` | `bash`, `git`, `gpg`, `jq`, `ssh`, `yq` | `firewalld` |
| infra   | tools    | `docker`, `docker-compose`, `jsondiff`, `netaddr`, `pexpect`, `psycopg2`, `pypsexec`, `pymongo`, `PyMySQL`, `smbprotocol` | `rsync`,`sshpass`,`proxychains` |
| azure   | tools    | `azure-*`              | - |
| aws     | tools    | `awscli`, `botocore`, `boto`, `boto3` | `aws` |
| awsk8s  | aws      | `openshift`            | `kubectl`, `oc` |
| awskops | awsk8s   | -                      | `kops` |
| awshelm | awsk8s   | -                      | `helm` |
| opus    | awshelm  | -                      | `k9s`,`clustershell`,`conman`  |


### Opus base

The following Opus Docker images are as small as possible and only contain Ansible itself.

| Docker tag | Build from                           |
|------------|--------------------------------------|
| `latest`   | Latest stable Ansible version        |
| `2.15`     | Latest stable Ansible 2.15.x version |

### Opus tools

The following Opus Docker images contain everything from `Ansible base` and additionally: `bash`, `git`, `gpg`, `jq`, `ssh` and `dnspython` and Ansible **`mitogen`** strategy plugin (see [Examples](#run-ansible-playbook-with-mitogen)).

| Docker tag     | Build from                           |
|----------------|--------------------------------------|
| `latest-tools` | Latest stable Ansible version        |
| `2.15-tools`   | Latest stable Ansible 2.15.x version |

### Opus infra

The following Ansible Docker images contain everything from `Opus tools` and additionally: `docker`, `pexpect`, `psycopg2`, `pypsexec`, `pymongo`, `PyMySQL` and `smbprotocol` Python libraries.

| Docker tag     | Build from                           |
|----------------|--------------------------------------|
| `latest-infra` | Latest stable Ansible version        |
| `2.15-infra`   | Latest stable Ansible 2.15.x version |

### Opus azure

The following Ansible Docker images contain everything from `Opus tools` and additionally: `azure`.

| Docker tag     | Build from                           |
|----------------|--------------------------------------|
| `latest-azure` | Latest stable Ansible version        |
| `2.15-azure`   | Latest stable Ansible 2.15.x version |

### Opus aws

The following Opus Docker images contain everything from `Opus tools` and additionally: `aws-cli`, `boto`, `boto3` and `botocore`.

| Docker tag   | Build from                           |
|--------------|--------------------------------------|
| `latest-aws` | Latest stable Ansible version        |
| `2.15-aws`   | Latest stable Ansible 2.11.x version |

### Opus awsk8s

The following Opus Docker images contain everything from `Opus aws` and additionally: `openshift` and `kubectl`.

| Docker tag      | Build from                           |
|-----------------|--------------------------------------|
| `latest-awsk8s` | Latest stable Ansible version        |
| `2.15-awsk8s`   | Latest stable Ansible 2.11.x version |

### Opus awskops
The following Ansible Docker images contain everything from `Opus awsk8s` and additionally: `kops` in its latest patch level version.

> https://github.com/kubernetes/kops/releases

#### Kops 1.25 (latest 1.25.x)

| Docker tag           | Build from                                          |
|----------------------|-----------------------------------------------------|
| `latest-awskops1.25` | Latest stable Ansible version with kops 1.25        |
| `2.15-awskops1.25`   | Latest stable Ansible 2.11.x version with kops 1.25 |

### Opus awshelm
The following Ansible Docker images contain everything from `Opus awsk8s` and additionally: `helm` in its latest patch level version.

> https://github.com/helm/helm/releases

#### Helm 3.10 (latest 3.10.x)

| Docker tag           | Build from                                           |
|----------------------|------------------------------------------------------|
| `latest-awshelm3.10`  | Latest stable Ansible version with Helm 3.10        |
| `2.15-awshelm3.10`    | Latest stable Ansible 2.11.x version with Helm 3.10 |


## Docker environment variables

Environment variables are available for all flavours except for `Opus base`.

| Variable        | Default | Allowed values | Description |
|-----------------|---------|----------------|-------------|
| `USER`          | ``      | `ansible`      | Set this to `ansible` to have everything run inside the container by the user `ansible` instead of `root` |
| `UID`           | `1000`  | integer        | If your local uid is not `1000` set it to your uid to syncronize file/dir permissions during mounting |
| `GID`           | `1000`  | integer        | If your local gid is not `1000` set it to your gid to syncronize file/dir permissions during mounting |
| `INIT_GPG_KEY`  | ``      | string         | If your gpg key requires a password you can initialize it during startup and cache the password (requires `INIT_GPG_PASS` as well) |
| `INIT_GPG_PASS` | ``      | string         | If your gpg key requires a password you can initialize it during startup and cache the password (requires `INIT_GPG_KEY` as well) |
| `INIT_GPG_CMD`  | ``      | string         | A custom command which will initialize the GPG key password. This allows for interactive mode to enter your password manually during startup. (Mutually exclusive to `INIT_GPG_KEY` and `INIT_GPG_PASS`) |


## Docker mounts

The working directory inside the Docker container is **`/data/`** and should be mounted locally to
the root of your project where your Ansible playbooks are.


## Examples

### Run Ansible playbook
```bash
docker run --rm -v $(pwd):/data localhost/ansible ansible-playbook playbook.yml
```

### Run Ansible playbook with Mitogen

> [Mitogen](https://github.com/dw/mitogen) updates Ansible’s slow and wasteful shell-centric implementation with pure-Python equivalents, invoked via highly efficient remote procedure calls to persistent interpreters tunnelled over SSH.

> No changes are required to target hosts. The extension is considered stable and real-world use is encouraged.

**Configuration (option 1)**

`ansible.cfg`
```ini
[defaults]
strategy_plugins = /usr/lib/python3.8/site-packages/ansible_mitogen/plugins/strategy
strategy         = mitogen_linear
```

**Configuration (option 2)**
```bash
# Instead of hardcoding it via ansible.cfg,  you could also add the
# option on-the-fly via environment variables.
export ANSIBLE_STRATEGY_PLUGINS=/usr/lib/python3.8/site-packages/ansible_mitogen/plugins/strategy
export ANSIBLE_STRATEGY=mitogen_linear
```

**Invocation**

```bash
docker run --rm -v $(pwd):/data localhost/ansible:latest-tools ansible-playbook playbook.yml
```

**Further reading:**

* [Mitogen on GitHub](https://github.com/dw/mitogen)
* [Mitogen Documentation](https://networkgenomics.com/ansible/)
* [How to Speed Up Your Ansible Playbooks Over 600%](https://www.toptechskills.com/ansible-tutorials-courses/speed-up-ansible-playbooks-pipelining-mitogen/)
* [Speed up Ansible with Mitogen](https://dev.to/sshnaidm/speed-up-ansible-with-mitogen-2c3j)


### Run Ansible playbook with non-root user
```bash
# Use 'ansible' user inside Docker container
docker run --rm \
  -e USER=ansible \
  -v $(pwd):/data \
  localhost/opus:latest-tools ansible-playbook playbook.yml
```
```bash
# Use 'ansible' user inside Docker container
# Use custom uid/gid for 'ansible' user inside Docker container
docker run --rm \
  -e USER=ansible \
  -e MY_UID=1000 \
  -e MY_GID=1000 \
  -v $(pwd):/data \
  localhost/opus:latest-tools ansible-playbook playbook.yml
```

### Run Ansible playbook with local ssh keys mounted
```bash
# Ensure to set same uid/gid as on your local system for Docker user
# to prevent permission issues during docker mounts
docker run --rm \
  -e USER=ansible \
  -e MY_UID=1000 \
  -e MY_GID=1000 \
  -v ${HOME}/.ssh/:/home/ansible/.ssh/:ro \
  -v $(pwd):/data \
  localhost/opus:latest-tools ansible-playbook playbook.yml
```

### Run Ansible playbook with local password-less gpg keys mounted
```bash
# Ensure to set same uid/gid as on your local system for Docker user
# to prevent permission issues during docker mounts
docker run --rm \
  -e USER=ansible \
  -e MY_UID=1000 \
  -e MY_GID=1000 \
  -v ${HOME}/.gnupg/:/home/ansible/.gnupg/ \
  -v $(pwd):/data \
  localhost/opus:latest-tools ansible-playbook playbook.yml
```

### Run Ansible playbook with local gpg keys mounted and automatically initialized
This is required in case your GPG key itself is encrypted with a password.
Note that the password needs to be in *single quotes*.
```bash
# Ensure to set same uid/gid as on your local system for Docker user
# to prevent permission issues during docker mounts
docker run --rm \
  -e USER=ansible \
  -e MY_UID=1000 \
  -e MY_GID=1000 \
  -e INIT_GPG_KEY=user@domain.tld \
  -e INIT_GPG_PASS='my gpg password' \
  -v ${HOME}/.gnupg/:/home/ansible/.gnupg/ \
  -v $(pwd):/data \
  localhost/opus:latest-tools ansible-playbook playbook.yml
```
Alternatively you can also export your GPG key and password to the shell's environment:
```bash
# Ensure to write the password in single quotes
export MY_GPG_KEY='user@domain.tld'
export MY_GPG_PASS='my gpg password'
```
```bash
# Ensure to set same uid/gid as on your local system for Docker user
# to prevent permission issues during docker mounts
docker run --rm \
  -e USER=ansible \
  -e MY_UID=1000 \
  -e MY_GID=1000 \
  -e INIT_GPG_KEY=${MY_GPG_KEY} \
  -e INIT_GPG_PASS=${MY_GPG_PASS} \
  -v ${HOME}/.gnupg/:/home/ansible/.gnupg/ \
  -v $(pwd):/data \
  localhost/opus:latest-tools ansible-playbook playbook.yml
```

### Run Ansible playbook with local gpg keys mounted and interactively interactively
The following will work with password-less and password-set GPG keys.
In case it requires a password, it will ask for the password and you need to enter it.
```bash
# Ensure to set same uid/gid as on your local system for Docker user
# to prevent permission issues during docker mounts
docker run --rm \
  -e USER=ansible \
  -e MY_UID=1000 \
  -e MY_GID=1000 \
  -e INIT_GPG_CMD='echo test | gpg --encrypt -r user@domain.tld | gpg --decrypt --pinentry-mode loopback' \
  -v ${HOME}/.gnupg/:/home/ansible/.gnupg/ \
  -v $(pwd):/data \
  localhost/opus:latest-tools ansible-playbook playbook.yml
```

### Run Ansible Galaxy
```bash
# Ensure to set same uid/gid as on your local system for Docker user
# to prevent permission issues during docker mounts
docker run --rm \
  -e USER=ansible \
  -e MY_UID=1000 \
  -e MY_GID=1000 \
  -v $(pwd):/data \
  localhost/opus:latest-tools ansible-galaxy install -r requirements.yml
```

### Run Ansible playbook with AWS credentials
```bash
# Basic
docker run --rm \
  -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
  -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
  -v $(pwd):/data \
  localhost/opus:latest-aws ansible-playbook playbook.yml
```
```bash
# With AWS Session Token
docker run --rm \
  -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
  -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
  -e AWS_SESSION_TOKEN=$AWS_SESSION_TOKEN \
  -v $(pwd):/data \
  localhost/opus:latest-aws ansible-playbook playbook.yml
```
```bash
# With ~/.aws/ config and credentials directories mounted (read/only)
# If you want to make explicit use of aws profiles, use this variant
# Ensure to set same uid/gid as on your local system for Docker user
# to prevent permission issues during docker mounts
docker run --rm \
  -e USER=ansible \
  -e MY_UID=1000 \
  -e MY_GID=1000 \
  -v ${HOME}/.aws/config:/home/ansible/.aws/config:ro \
  -v ${HOME}/.aws/credentials:/home/ansible/.aws/credentials:ro \
  -v $(pwd):/data \
  localhost/opus:latest-aws ansible-playbook playbook.yml
```

### Run Ansible playbook against AWS with gpg vault initialization
```bash
# Ensure to set same uid/gid as on your local system for Docker user
# to prevent permission issues during docker mounts
docker run --rm \
  -e USER=ansible \
  -e MY_UID=1000 \
  -e MY_GID=1000 \
  -e INIT_GPG_KEY=user@domain.tld \
  -e INIT_GPG_PASS='my gpg password' \
  -v ${HOME}/.aws/config:/home/ansible/.aws/config:ro \
  -v ${HOME}/.aws/credentials:/home/ansible/.aws/credentials:ro \
  -v ${HOME}/.gnupg/:/home/ansible/.gnupg/ \
  -v $(pwd):/data \
  localhost/opus:latest-aws \
  ansible-playbook playbook.yml
```
As the command is getting pretty long, you could wrap it into a Makefile.
```make
ifneq (,)
.error This Makefile requires GNU Make.
endif

.PHONY: dry run

CURRENT_DIR = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
ANSIBLE = 2.8
UID = 1000
GID = 1000

# Ansible check mode uses mitogen_linear strategy for much faster roll-outs
dry:
ifndef GPG_PASS
	docker run --rm it \
		-e ANSIBLE_STRATEGY_PLUGINS=/usr/lib/python3.8/site-packages/ansible_mitogen/plugins/strategy \
		-e ANSIBLE_STRATEGY=mitogen_linear \
		-e USER=ansible \
		-e MY_UID=$(UID) \
		-e MY_GID=$(GID) \
		-v $${HOME}/.aws/config:/home/ansible/.aws/config:ro \
		-v $${HOME}/.aws/credentials:/home/ansible/.aws/credentials:ro \
		-v $${HOME}/.gnupg/:/home/ansible/.gnupg/ \
		-v $(CURRENT_DIR):/data \
		localhost/opus:$(ANSIBLE)-aws ansible-playbook playbook.yml --check
else
	docker run --rm it \
		-e ANSIBLE_STRATEGY_PLUGINS=/usr/lib/python3.8/site-packages/ansible_mitogen/plugins/strategy \
		-e ANSIBLE_STRATEGY=mitogen_linear \
		-e USER=ansible \
		-e MY_UID=$(UID) \
		-e MY_GID=$(GID) \
		-e INIT_GPG_KEY=$${GPG_KEY} \
		-e INIT_GPG_PASS=$${GPG_PASS} \
		-v $${HOME}/.aws/config:/home/ansible/.aws/config:ro \
		-v $${HOME}/.aws/credentials:/home/ansible/.aws/credentials:ro \
		-v $${HOME}/.gnupg/:/home/ansible/.gnupg/ \
		-v $(CURRENT_DIR):/data \
		localhost/opus:$(ANSIBLE)INIT_GPG_KEY` -aws \
		ansible-playbook playbook.yml --check
endif

# Ansible real run uses default strategy
run:
ifndef GPG_PASS
	docker run --rm it \
		-e USER=ansible \
		-e MY_UID=$(UID) \
		-e MY_GID=$(GID) \
		-v $${HOME}/.aws/config:/home/ansible/.aws/config:ro \
		-v $${HOME}/.aws/credentials:/home/ansible/.aws/credentials:ro \
		-v $${HOME}/.gnupg/:/home/ansible/.gnupg/ \
		-v $(CURRENT_DIR):/data \
		localhost/opus:$(ANSIBLE)-aws ansible-playbook playbook.yml
else
	docker run --rm it \
		-e USER=ansible \
		-e MY_UID=$(UID) \
		-e MY_GID=$(GID) \
		-e INIT_GPG_KEY=$${GPG_KEY} \
		-e INIT_GPG_PASS=$${GPG_PASS} \
		-v $${HOME}/.aws/config:/home/ansible/.aws/config:ro \
		-v $${HOME}/.aws/credentials:/home/ansible/.aws/credentials:ro \
		-v $${HOME}/.gnupg/:/home/ansible/.gnupg/ \
		-v $(CURRENT_DIR):/data \
		localhost/opus:$(ANSIBLE)-aws \
		ansible-playbook playbook.yml
endif
```
**Important:**

THE `GPG_KEY` and `GPG_PASS` will not be echo'ed out by the Make command and you are advised to
export those values via your shell's `export` command to the env in order to hide it.

If you still want to specify them on the command line via `make dry GPG_KEY='pass'`
and your pass or key contains one or more `$` characters
then they must all be escaped with an additional `$` in front. This is not necessary if you export
them.

**Example:** If your password is `test$5`, then you must use `make dry GPG_PASS='test$$5'`.


Then you can call it easily:
```bash
# With GPG password from the env
export GPG_KEY='user@domain.tld'
export GPG_PASS='THE_GPG_PASSWORD_HERE'
make dry
make run

# With GPG password on the cli
make dry GPG_KEY='user@domain.tld' GPG_PASS='THE_GPG_PASSWORD_HERE'
make run GPG_KEY='user@domain.tld' GPG_PASS='THE_GPG_PASSWORD_HERE'

# Without GPG password
make dry
make run

# With different Ansible version
make dry ANSIBLE=2.6
make run ANSIBLE=2.6

# With different uid/gid
make dry UID=1001 GID=1001
make run UID=1001 GID=1001
```


## Build locally

To build locally you require GNU Make to be installed. Instructions as  shown below.

### Ansible base
```bash
# Build latest Ansible base
# image: localhost/ansible:latest
make build

# Build Ansible 2.6 base
# image: localhost/ansible:2.6
make build ANSIBLE=2.6
```
### Ansible tools
```bash
# Build latest Ansible tools
# image: localhost/ansible:latest-tools
make build ANSIBLE=latest FLAVOUR=tools

# Build Ansible 2.6 tools
# image: localhost/ansible:2.6-tools
make build ANSIBLE=2.6 FLAVOUR=tools
```

### Ansible infra
```bash
# Build latest Ansible infra
# image: localhost/ansible:latest-infra
make build ANSIBLE=latest FLAVOUR=infra

# Build Ansible 2.6 infra
# image: localhost/ansible:2.6-infra
make build ANSIBLE=2.6 FLAVOUR=infra
```

### Ansible azure
```bash
# Build latest Ansible azure
# image: localhost/ansible:latest-azure
make build ANSIBLE=latest FLAVOUR=azure

# Build Ansible 2.6 azure
# image: localhost/ansible:2.6-azure
make build ANSIBLE=2.6 FLAVOUR=azure
```

### Ansible aws
```bash
# Build latest Ansible aws
# image: localhost/ansible:latest-aws
make build ANSIBLE=latest FLAVOUR=aws

# Build Ansible 2.6 aws
# image: localhost/ansible:2.6-aws
make build ANSIBLE=2.6 FLAVOUR=aws
```

### Ansible awsk8s
```bash
# Build latest Ansible awsk8s
# image: localhost/ansible:latest-awsk8s
make build ANSIBLE=latest FLAVOUR=awsk8s

# Build Ansible 2.6 awsk8s
# image: localhost/ansible:2.6-awsk8s
make build ANSIBLE=2.6 FLAVOUR=awsk8s
```

### Ansible awskops
```bash
# Build latest Ansible with Kops 1.8
# image: localhost/ansible:latest-awskops1.8
make build ANSIBLE=latest FLAVOUR=awskops KOPS=1.8

# Build Ansible 2.6 with Kops 1.8
# image: localhost/ansible:2.6-awskops1.8
make build ANSIBLE=2.6 FLAVOUR=awskops KOPS=1.8
```

### Ansible awshelm
```bash
# Build latest Ansible with Helm 2.14
# image: localhost/ansible:latest-awshelm1.8
make build ANSIBLE=latest FLAVOUR=awshelm HELM=2.14

# Build Ansible 2.6 with Kops 1.8
# image: localhost/ansible:2.6-awshelm1.8
make build ANSIBLE=2.6 FLAVOUR=awshelm HELM=2.14
```

## License

**[Apache-2.0 License](../LICENSE)**
Copyright (c) 2025 The Ubiquity Authors