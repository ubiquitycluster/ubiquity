# OVH

Refer to [Ubiquity Documentation](https://github.com/ubiquitycluster/ubiquity/tree/main/docs) for a complete step-by-step guide.

TL;DR:
1. Download your project OpenStack RC File v3 (v2 won't work) from the [OVH OpenStack interface](https://horizon.cloud.ovh.net/project/) at the top right corner of the page.
2. Source your project openrc file : `source _project_-openrc.sh`.
3. Download the [latest release of Ubiquity for OVH](https://github.com/ubiquitycluster/ubiquity/releases/latest).
4. Unpack the release.
5. Adapt the [parameters](https://github.com/ubiquitycluster/ubiquity/tree/main/docs#4-configuration) in `main.tf` file.
6. In a terminal, change your current directory to the directory containing `main.tf`.
7. Initiate the Terraform state : `terraform init`.
8. Verify the Terraform plan : `terraform plan`.
9. Apply the Terraform plan : `terraform apply`.

To tear down the cluster, call: `terraform destroy`.