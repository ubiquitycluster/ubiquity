# Amazon Web Services

Refer to [Ubiquity Documentation](https://github.com/ubiquitycluster/ubiquity/tree/main/docs) for a complete step-by-step guide.

TL;DR:
1. Export the following environment variables : `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`.
2. Download the [latest release of Ubiquity for AWS](https://github.com/ubiquitycluster/ubiquity/releases/latest).
3. Unpack the release.
4. Adapt the [parameters](https://github.com/ubiquitycluster/ubiquity/tree/main/docs#4-configuration) in `main.tf` file.
5. In a terminal, change your current directory to the directory containing `main.tf`.
6. Initiate the Terraform state : `terraform init`.
7. Verify the Terraform plan : `terraform plan`.
8. Apply the Terraform plan : `terraform apply`.

To tear down the cluster, call: `terraform destroy`.