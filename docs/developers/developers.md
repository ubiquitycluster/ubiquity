# Ubiquity Developer Documentation

## Table of Content

1. [Setup](#1-setup)
2. [Where to start](#2-where-to-start)
4. [Troubleshooting](#4-troubleshooting)
5. [Release](#5-release)


## 1. Setup

To develop for Ubiquity you will need:
* Terraform (>= 1.2.1)
* git
* Access to a Cloud provider (AWS, GCP, Azure, OpenStack, etc.) OR hardware to develop on - See the development rig that we use [here](../../Rig.jpg)
* Ability to communicate with the cloud provider API from your computer
* A cloud project with enough room for the resource described in section [Ubiquity Doc 1.1](README.md#11-quotas).

## 2. Where to start

The Ubiquity project is defined by an infrastructure-as-code component that is responsible for generating a cluster architecture in a cloud and a Ansible environment that configures the cluster instances based on their role.

If you wish to add device, an instance, add a new networking interface or a filesystem, you will most likely need to develop some Terraform code if you're working on a cloud provider. 
The project structure for Terraform code is described in the [reference design document](../architecture/cloud/design.md). 
The document also describes how one could work with current Ubiquity code to add support for another cloud provider.

If you wish to add a service to one of the Ansible roles, install new software, modify an instance configuration or role, you will most likely need to develop some Ansible
code. The following section provides more details on the Ansible environments available and how to develop them.

## 3. Ansible environment

Ubiquity Terraform code initialises every instance to be a cloud-init agent and an instance
with the tag `awx.ubiquitycluster.uk` (by default, configurable) as the AWX main server. 

On the AWX main server, there is a folder containing the configuration code for the instances of the cluster, this folder is called an Ansible environment and it is pulled from GitHub 
during the initial configuration of the AWX main server.

The source of that environment is provided to Terraform using the variable `config_git_url`.

A repository describing an AWX environment must contain at the least the following files 
and folders:
```
ubiq-playbooks
┗ collections
  ┗ requirements.yml
┣ ansible.cfg
┣ ubiquity.yml
┗ inventory
  ┗ mail.yml
┗ vars/
  ┗ main.yml
```

- [`collections/requirements.yml`](https://docs.ansible.com/ansible/5/user_guide/collections_using.html) specifies the Ansible collections that need to be installed in the environment.
- [`ansible.cfg`](https://docs.ansible.com/ansible/latest/reference_appendices/config.html) overrides the primary server default settings for the environment. Note that these are often present per-role as well for role-specific overrides.
- [`ubiquity.yml`](https://docs.ansible.com/ansible/latest/playbook_guide/index.html) configures a playbook that runs an ordered list of playbooks with their own roles inside to be applied to the instances. This structure is to enable individual playbooks to be applied to instances as needed (and allow testing of individual playbooks).
- `vars/main.yml` is common data source to apply variables as an override to the roles. You can also apply this inside AWX.
- `inventory/main.yml` defines how each instance will be configured based on their hostname and/or groups.

An example of a bare-bone Ubiquity environment is available on GitHub:
[Ubiquity/ubiq-playbooks](https://github.com/CJCShadowsan/ubiq-playbooks), while the
Ubiquity AWX environment that replicates a Ubiquity HPC cluster is named
[ubiquitycluster/ubiq-playbooks](https://github.com/ubiquitycluster/ubiq-playbooks).

### terraform_data.yaml: a bridge between Terraform and AWX

To provide information on the deployed resources and the value of the input parameters,
Ubiquity Terraform code uploads to the AWX main server a file named `terraform_data.yaml`,
in the folder `/etc/puppetlabs/data/`. There is also a symlink created in
`/etc/puppetlabs/code/environment/production/data/` to ease its usage inside the Puppet
environment.

When included in the data hierarchy (`hiera.yaml`), `terraform_data.yaml` can provide
information about the instances, the volumes and the variables set by the user
through the `main.tf` file. The file has the following structure:
```yaml
---
terraform:
  instances:
    hostname1:
      local_ip: "10.0.0.x"
      public_ip: ""
      tags: ["tag_1"]
      hostkeys:
        rsa: ""
  volumes:
    volume_tag1:
      volume_1:
        - "/dev/disk/by-id/123-*"
      volume_2:
        - "/dev/disk/by-id/123-abc-*"
  tag_ip:
    tag_1:
      - 10.0.0.x
  data:
    cluster_name: ""
    consul_token: ""
    domain_name: ""
    freeipa_passwd: ""
    guest_passwd: ""
    munge_key: ""
    nb_users: ""
    sudoer_username: ""
```

The values provided by `terraform_data.yaml` can be accessed in Puppet by using the
`lookup()` function. For example, to access an instance's list of tags:
```puppet
lookup("terraform.instances.${::hostname}.tags")
```
The data source can also be used to define a key in another data source YAML file by using the
`alias()` function. For example, to define the number of guest accounts using the value of `nb_user`,
we could add this to `common.yaml`
```yaml
profile::accounts::guests::nb_accounts: "%{alias('terraform.data.nb_users')}"
```

### Configuring instances: site.pp and classes

The configuration of each instance is defined in `manifests/site.pp` file of the Puppet environment.
In this file, it is possible to define a configuration based on an instance hostname
```
node "mgmt1" { }
```
or using the instance tags by defining the configuration for the `default` node :
```
node default {
  $instance_tags = lookup("terraform.instances.${::hostname}.tags")
  if 'tag_1' in $instances_tags { }
}
```

It is possible to define [Puppet resource](https://puppet.com/docs/puppet/6/type.html) directly
in `site.pp`. However, above a certain level of complexity, which can be reach fairly quickly, it
is preferable to define classes and include these classes in `site.pp` based on the node hostname
or tags.

Classes can be defined in the Puppet environment under the following path:
`site/profile/manifests`. These classes are named profile classes and the philosophy
behind it is explained in [Puppet documentation](https://puppet.com/docs/pe/2019.8/osp/the_roles_and_profiles_method.html). Because these classes are defined in `site/profile`,
their name has to start with the prefix `profile::`.

It is also possible to include classes defined externally and installed using the `Puppetfile`.
These classes installed by [librarian-puppet](https://github.com/voxpupuli/librarian-puppet) can be found in the `modules` folder of the
Puppet environment.

## 4. Troubleshooting

### 4.1 cloud-init

To test new additions to OS images and Ansible playbooks, it is possible to
execute cloud-init phases manually. There are four steps that can be executed sequentially: init local, init
modules config and modules final. Here are the corresponding commands to execute each step:
```
cloud-init init --local
cloud-init init
cloud-init modules --mode=config
cloud-init modules --mode=final
```

It is also possible to clean a cloud-init execution and have it execute again at next reboot. To do so, enter
the following command:
```
cloud-init clean
```
Add `-r` to the previous command to reboot the instance once cloud-init has finishing cleaning.

### 4.2 SELinux

SELinux is enabled on every instances of a Ubiquity cluster. Some applications do not provide
SELinux policies which can lead to their malfunctionning when SELinux is enabled. It is possible
to track down the reasons why SELinux is preventing an application to work properly using
the command-line tool `ausearch`.

If you suspect application `app-a` to be denied by SELinux to work properly, run the following
command as root:
```
ausearch -c app-a --raw | grep denied
```

To see all requests denied by SELinux:
```
ausearch --raw | grep denied
```

Sometime, the denials are hidden from regular logging. To display all denials, run the following
command as root:
```
semodule --disable_dontaudit --build
```
then re-execute the application that is not working properly.

Once you have found the denials that are the cause of the problem, you can create a new policy
to allow the requests that were previously denied with the following command:
```
ausearch -c app-a --raw | grep denied | audit2allow -a -M app-a
```

Finally, you can install the generated policy using the command provided by `auditallow`.

#### Building the policy package file (.pp) from the enforcement file (.te)

If you need to tweak an existing enforcement file and you want to recompile the policy package,
you can with the following commands:
```
checkmodule -M -m -o my_policy.mod my_policy.te
semodule_package -o my_policy.pp -m my_policy.mod
```

#### References
- https://wiki.gentoo.org/wiki/SELinux
- https://wiki.gentoo.org/wiki/SELinux/Tutorials/Where_to_find_SELinux_permission_denial_details


## 5. Release

To build a release, use the script `release.sh` located in the scripts folder of the Ubiquity git repo.
```
Usage: release.sh VERSION [provider ...]
```
The script creates a folder named `releases` where it was called.

The `VERSION` argument is expected to correspond to git tag in the `ubiq-playbooks` repo.
It could also be a branch name or a commit. If the provider optional argument is left blank,
release files will be built for all providers currently supported by Ubiquity.

Examples:

- Building a release for OpenStack with the ansible repo main branch:
    ```
    $ ./release.sh main openstack
    ```
- Building a release for GCP with the latest Terraform and cloud-init, and version 5.8 of ansible playbooks
Ubiquity:
    ``` 
    $ ./release.sh 5.8 gcp
    ```
- Building a release for Azure and OVH with the latest Terraform and cloud-init, and version 5.7 of ansible playbooks
Ubiquity:
    ```
    $ ./release.sh 5.7 azure ovh
    ```