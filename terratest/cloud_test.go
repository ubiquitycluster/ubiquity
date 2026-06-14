package terratest

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
)

// TestCloudModule is a skeleton test for Terraform cloud modules.
// It is skipped by default because it requires actual cloud credentials.
func TestCloudModule(t *testing.T) {
	t.Skip("Skipping: requires cloud credentials")

	terraformOptions := &terraform.Options{
		TerraformDir: "../cloud/aws",
		Vars: map[string]interface{}{
			"cluster_name": "test-cluster",
		},
	}

	defer terraform.Destroy(t, terraformOptions)
	terraform.InitAndApply(t, terraformOptions)
}