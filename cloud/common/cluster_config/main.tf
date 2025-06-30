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
resource "random_pet" "guest_passwd" {
  count     = var.guest_passwd != "" ? 0 : 1
  length    = 4
  separator = "."
}

locals {
  public_instances = { for key, values in var.instances : key => values if contains(values["tags"], "public") }
  ansibleserver_id = try(element([for key, values in var.instances: values["id"] if contains(values["tags"], "ansible")], 0), "")
  all_tags = toset(flatten([for key, values in var.instances : values["tags"]]))
  all_nodes = toset(flatten([for key, values in var.instances : [key] ]))

  master_nodes = { for key, values in var.instances : key => values if contains(values["tags"], "master") }
  worker_nodes = { for key, values in var.instances : key => values if contains(values["tags"], "worker") }
  login_nodes = { for key, values in var.instances : key => values if contains(values["tags"], "login") }
  compute_nodes = { for key, values in var.instances : key => values if contains(values["tags"], "compute") }
  mgmt_nodes = { for key, values in var.instances : key => values if contains(values["tags"], "mgmt") }
  gpu_nodes = { for key, values in var.instances : key => values if contains(values["tags"], "gpu") }
  gpfs_nodes = { for key, values in var.instances : key => values if contains(values["tags"], "cesgpfs") }
  nfs_nodes = { for key, values in var.instances : key => values if contains(values["tags"], "nfs") }
  nfs_clients = { for key, values in var.instances : key => values if contains(values["tags"], "nfs_client") }
  nfs_server = { for key, values in var.instances : key => values if contains(values["tags"], "nfs_server") }
  vis_nodes = { for key, values in var.instances : key => values if contains(values["tags"], "vis") }
  k8s_nodes = { for key, values in var.instances : key => values if contains(values["tags"], "k8s") }

  tag_ip = { for tag in local.all_tags :
    tag => [for key, values in var.instances : values["local_ip"] if contains(values["tags"], tag)]
  }
  inventory_data = templatefile("${path.module}/ansible/inventory/prod.tftpl",
    {
      instances             = yamlencode(var.instances)
      tags                  = yamlencode(local.all_tags)
      tag_ip                = yamlencode(local.tag_ip)
      master_nodes          = local.master_nodes
      worker_nodes          = local.worker_nodes
      login_nodes           = local.login_nodes
      compute_nodes         = local.compute_nodes
      mgmt_nodes            = local.mgmt_nodes
      gpu_nodes             = local.gpu_nodes
      gpfs_nodes            = local.gpfs_nodes
      nfs_nodes             = local.nfs_nodes
      nfs_clients           = local.nfs_clients
      nfs_server            = local.nfs_server
      vis_nodes             = local.vis_nodes
      k8s_nodes             = local.k8s_nodes

      nodes = { for node in local.all_nodes :
      
        node => [ 
          for key, values in var.instances : 
            {
              hostname = key
              ip       = values["local_ip"]
              //tags = values["tags"]
              // Add extra fields here
              extra_field1 = "value1"
              extra_field2 = "value2"
            } if node == key
            //if contains(values["tags"], node)
        ]
      }
      sudoer_username       = var.sudoer_username
    })

  ansible_vars = templatefile("${path.module}/terraform_data.yaml",
    {
      instances   = yamlencode(var.instances)
      tag_ip      = yamlencode(local.tag_ip)
      volumes     = yamlencode(var.volume_devices)
      filesystems = yamlencode(var.filesystems)
      data        = yamlencode({
        sudoer_username = var.sudoer_username
        public_keys     = var.tf_ssh_key.public == null ? var.public_keys : concat(var.public_keys, [var.tf_ssh_key.public])
        cluster_name    = lower(var.cluster_name)
        domain_name     = var.domain_name
        guest_passwd    = var.guest_passwd != "" ? var.guest_passwd : try(random_pet.guest_passwd[0].id, "")
        nb_users        = var.nb_users
      })
  })
  facts = {
    software_stack = var.software_stack
    cloud          = {
      provider = var.cloud_provider
      region = var.cloud_region
    }
  }
}

resource "null_resource" "deploy_ansible_vars" {
  count = contains(local.all_tags, "ansible") && contains(local.all_tags, "public") ? 1 : 0

  connection {
    type                = "ssh"
    bastion_host        = local.public_instances[keys(local.public_instances)[0]]["public_ip"]
    bastion_user        = var.sudoer_username
    bastion_private_key = var.tf_ssh_key.private
    user                = var.sudoer_username
    host                = "ansible"
    private_key         = var.tf_ssh_key.private
  }

  triggers = {
    user_data         = md5(var.ansible_vars)
    ansible_vars      = md5(local.ansible_vars)
    inventory_data    = md5(local.inventory_data)
    facts             = md5(yamlencode(local.facts))
    ansibleserver     = local.ansibleserver_id
  }

  provisioner "file" {
    content     = local.ansible_vars
    destination = "terraform_data.yaml"
  }
  provisioner "file" {
    content     = var.master_key.private
    destination = "~/.ssh/private_key.pem"
  }
  provisioner "remote-exec" {
    when    = create
    inline = ["ls ~/.ssh/private_key.pem || echo '${var.master_key.private}' > ~/.ssh/private_key.pem && chmod 400 ~/.ssh/private_key.pem"]
  }
  provisioner "file" {
    content     = local.inventory_data
    destination = "inventory.yaml"
  }

  provisioner "file" {
    content     = yamlencode(local.facts)
    destination = "terraform_facts.yaml"
  }

  provisioner "file" {
    content     = var.ansible_vars
    destination = "user_data.yaml"
  }



  provisioner "remote-exec" {
    inline = [
      "id -u ansible &> /dev/null || sudo useradd -m -s /bin/bash ansible",
      "which yum && sudo yum -y install make git || echo 'No yum' && true",
      "which apt && sudo apt-get -y install make git || echo 'No apt-get' && true",
      "#git clone https://github.com/ubiquitycluster/ubiquity.git",
      "#cd ubiquity && make tools",
      "#sudo install -m 650 terraform_facts.yaml /etc/ansible/facts/",
      # These chgrp commands do nothing if the ansible group does not yet exist
      # so these are also handled by ansible.yaml
      "#sudo chgrp ansible /etc/ansible/data/terraform_data.yaml /etc/ansible/data/user_data.yaml &> /dev/null || true",
      "#sudo chgrp ansible /etc/ansible/facts/terraform_facts.yaml &> /dev/null || true",
      "#rm -f terraform_data.yaml user_data.yaml terraform_facts.yaml",
      "#[ -f /usr/local/bin/consul ] && [ -f /usr/bin/jq ] && consul event -token=$(sudo jq -r .acl.tokens.agent /etc/consul/config.json) -name=ansible $(date +%s) || true",
    ]
  }
}
