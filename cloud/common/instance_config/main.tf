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
resource "random_string" "ansibleserver_password" {
  length  = 32
  special = false
}

resource "tls_private_key" "ssh" {
  count     = var.generate_ssh_key ? 1 : 0
  algorithm = "ED25519"
}

resource "tls_private_key" "master_ssh" {
  algorithm = "RSA"
  rsa_bits = 4096
  provisioner "local-exec" {
    when    = create
    command = "echo '${tls_private_key.master_ssh.private_key_pem}' > master_key.priv && chmod 400 master_key.priv"
  }
}

resource "tls_private_key" "rsa_hostkeys" {
  for_each  = toset([for x, values in var.instances: values["prefix"]])
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tls_private_key" "ed25519_hostkeys" {
  for_each  = toset([for x, values in var.instances: values["prefix"]])
  algorithm = "ED25519"
}

locals {
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
  ssh_key = {
    public  = try("${chomp(tls_private_key.ssh[0].public_key_openssh)} terraform@localhost", null)
    private = try(tls_private_key.ssh[0].private_key_pem, null)
  }        
  master_key = {
    public  = try("${chomp(tls_private_key.master_ssh.public_key_openssh)} terraform@localhost", null)
    private = try(tls_private_key.master_ssh.private_key_pem, null) 
  }

  user_data = {
    for key, values in var.instances : key =>
    templatefile("${path.module}/ansible/vars/main.yaml",
      {
        tags                   = values["tags"],
        node_name              = key,
        ansibleenv_git         = var.config_git_url,
        ansibleenv_rev         = var.config_version,
        ansibleserver_ip       = var.ansibleserver_ip,
        ansibleserver_password = random_string.ansibleserver_password.result,
        sudoer_username        = var.sudoer_username,
        master_key             = local.master_key.public,
        master_nodes           = local.master_nodes,
        worker_nodes           = local.worker_nodes,
        login_nodes            = local.login_nodes,
        compute_nodes          = local.compute_nodes,
        mgmt_nodes             = local.mgmt_nodes,
        gpu_nodes              = local.gpu_nodes,
        gpfs_nodes             = local.gpfs_nodes,
        nfs_nodes              = local.nfs_nodes,
        nfs_clients            = local.nfs_clients,
        nfs_server             = local.nfs_server,
        vis_nodes              = local.vis_nodes,
        k8s_nodes              = local.k8s_nodes

        ssh_authorized_keys   = local.ssh_key.public == null ? var.public_keys : concat(var.public_keys, [local.ssh_key.public])
        hostkeys = {
          rsa = {
            private = tls_private_key.rsa_hostkeys[values["prefix"]].private_key_pem
            public  = tls_private_key.rsa_hostkeys[values["prefix"]].public_key_openssh
          }
          ed25519 = {
            private = tls_private_key.ed25519_hostkeys[values["prefix"]].private_key_openssh
            public  = tls_private_key.ed25519_hostkeys[values["prefix"]].public_key_openssh
          }
        }
      }
    )
  }
}