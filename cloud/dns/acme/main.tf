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

# Setup ACME DNS Provider
provider "acme" {
  server_url = "https://acme-v02.api.letsencrypt.org/directory"
}

variable "name" {
}

variable "domain" {
}

variable "email" {
}

variable "sudoer_username" {
}

variable "dns_provider" {
}

variable "dns_provider_config" {
  default = {}
}

variable "ssl_tags" { }

variable "public_instances" {}

variable "ssh_private_key" {
  type = string
}

variable "acme_key_pem" {
  type = string
  default = ""
}

resource "tls_private_key" "private_key" {
  count = var.acme_key_pem == "" ? 1 : 0
  algorithm = "RSA"
}

resource "acme_registration" "reg" {
  count           = var.acme_key_pem == "" ? 1 : 0
  account_key_pem = tls_private_key.private_key[0].private_key_pem
  email_address   = var.email
}

resource "acme_certificate" "certificate" {
  account_key_pem           = var.acme_key_pem == "" ? acme_registration.reg[0].account_key_pem : var.acme_key_pem
  common_name               = "${var.name}.${var.domain}"
  subject_alternative_names = ["*.${var.name}.${var.domain}"]

  dns_challenge {
    provider = var.dns_provider
    config   = var.dns_provider_config
  }
}

resource "null_resource" "deploy_certs" {
  for_each = { for key, values in var.public_instances: key => values if length(setintersection(var.ssl_tags, values.tags)) > 0 }

  triggers = {
    instance_id    = each.value["id"]
    certificate_id = acme_certificate.certificate.id
  }

  connection {
    type        = "ssh"
    user        = var.sudoer_username
    host        = each.value["public_ip"]
    host_key    = each.value["hostkeys"]["rsa"]
    private_key = var.ssh_private_key
  }

  provisioner "file" {
    content     = acme_certificate.certificate.private_key_pem
    destination = "privkey.pem"
  }

  provisioner "file" {
    content     = "${acme_certificate.certificate.certificate_pem}${acme_certificate.certificate.issuer_pem}"
    destination = "fullchain.pem"
  }

  provisioner "file" {
    content     = acme_certificate.certificate.certificate_pem
    destination = "cert.pem"
  }

  provisioner "file" {
    content     = acme_certificate.certificate.issuer_pem
    destination = "chain.pem"
  }

  provisioner "file" {
    destination = "renewal.conf"
    content     = <<EOF
version = 0.34.2
archive_dir = /etc/letsencrypt/archive/${var.name}.${var.domain}
cert = /etc/letsencrypt/live/${var.name}.${var.domain}/cert.pem
privkey = /etc/letsencrypt/live/${var.name}.${var.domain}/privkey.pem
chain = /etc/letsencrypt/live/${var.name}.${var.domain}/chain.pem
fullchain = /etc/letsencrypt/live/${var.name}.${var.domain}/fullchain.pem

[renewalparams]
authenticator = nginx
installer = nginx
account = ${basename(acme_certificate.certificate.id)}
server = https://acme-v02.api.letsencrypt.org/directory

EOF

  }

  provisioner "remote-exec" {
    inline = [
      "sudo mkdir -p /etc/letsencrypt/{archive,live}/${var.name}.${var.domain}/",
      "sudo mkdir -p /etc/letsencrypt/renewal/",
      "sudo install -m 644 -o root -g root fullchain.pem /etc/letsencrypt/archive/${var.name}.${var.domain}/fullchain1.pem",
      "sudo install -m 644 -o root -g root chain.pem /etc/letsencrypt/archive/${var.name}.${var.domain}/chain1.pem",
      "sudo install -m 644 -o root -g root cert.pem /etc/letsencrypt/archive/${var.name}.${var.domain}/cert1.pem",
      "sudo install -m 640 -o root -g root privkey.pem /etc/letsencrypt/archive/${var.name}.${var.domain}/privkey1.pem",
      "sudo install -m 644 -o root -g root renewal.conf /etc/letsencrypt/renewal/${var.name}.${var.domain}.conf",
      "sudo ln -sf /etc/letsencrypt/archive/${var.name}.${var.domain}/privkey1.pem /etc/letsencrypt/live/${var.name}.${var.domain}/privkey.pem",
      "sudo ln -sf /etc/letsencrypt/archive/${var.name}.${var.domain}/fullchain1.pem /etc/letsencrypt/live/${var.name}.${var.domain}/fullchain.pem",
      "sudo ln -sf /etc/letsencrypt/archive/${var.name}.${var.domain}/cert1.pem /etc/letsencrypt/live/${var.name}.${var.domain}/cert.pem",
      "sudo ln -sf /etc/letsencrypt/archive/${var.name}.${var.domain}/chain1.pem /etc/letsencrypt/live/${var.name}.${var.domain}/chain.pem",
      "rm cert.pem chain.pem fullchain.pem privkey.pem renewal.conf",
      "id -u caddy &> /dev/null && sudo chgrp caddy /etc/letsencrypt/archive/${var.name}.${var.domain}/privkey1.pem",
      "test -f /usr/lib/systemd/system/caddy.service && sudo systemctl restart caddy || true",
    ]
  }
}