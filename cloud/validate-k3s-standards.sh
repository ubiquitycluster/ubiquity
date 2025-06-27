#!/bin/bash
# Ubiquity k3s Standards Validation Script
# This script validates that all cloud providers follow the standardized k3s deployment patterns

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLOUD_DIR="$SCRIPT_DIR"

echo "🔍 Validating Ubiquity k3s Standards Compliance..."
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Validation results
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0

# Function to check if a file exists and contains required content
check_file_content() {
    local provider=$1
    local file=$2
    local pattern=$3
    local description=$4
    
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    
    if [[ -f "$CLOUD_DIR/$provider/$file" ]]; then
        if grep -q "$pattern" "$CLOUD_DIR/$provider/$file"; then
            echo -e "  ${GREEN}✓${NC} $description"
            PASSED_CHECKS=$((PASSED_CHECKS + 1))
        else
            echo -e "  ${RED}✗${NC} $description - Pattern not found: $pattern"
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
        fi
    else
        echo -e "  ${RED}✗${NC} $description - File not found: $provider/$file"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
    fi
}

# Function to validate a provider
validate_provider() {
    local provider=$1
    echo -e "\n${BLUE}📁 Validating $provider provider${NC}"
    echo "----------------------------------------"
    
    # Check for required files
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    if [[ -d "$CLOUD_DIR/$provider" ]]; then
        echo -e "  ${GREEN}✓${NC} Provider directory exists"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
    else
        echo -e "  ${RED}✗${NC} Provider directory missing"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
        return
    fi
    
    # Check infrastructure.tf standards
    check_file_content "$provider" "infrastructure.tf" "module \"design\"" "Uses common design module"
    check_file_content "$provider" "infrastructure.tf" "module \"instance_config\"" "Uses common instance_config module"
    check_file_content "$provider" "infrastructure.tf" "module \"cluster_config\"" "Uses common cluster_config module"
    check_file_content "$provider" "infrastructure.tf" "cloud_provider.*=.*\"$provider\"" "Defines correct cloud_provider"
    check_file_content "$provider" "infrastructure.tf" "ansibleserver_ip" "Defines ansibleserver_ip"
    check_file_content "$provider" "infrastructure.tf" "k3s_cluster_validation" "Includes k3s cluster validation"
    
    # Check network.tf standards
    check_file_content "$provider" "network.tf" "k3s_firewall_rules" "Defines k3s firewall rules"
    check_file_content "$provider" "network.tf" "6443" "Includes k3s API server port"
    check_file_content "$provider" "network.tf" "8472" "Includes Flannel VXLAN port"
    check_file_content "$provider" "network.tf" "10250" "Includes Kubelet metrics port"
    
    # Check outputs.tf standards
    check_file_content "$provider" "outputs.tf" "cluster_info" "Provides cluster_info output"
    check_file_content "$provider" "outputs.tf" "kubeconfig_command" "Provides kubeconfig_command output"
    check_file_content "$provider" "outputs.tf" "cluster_endpoints" "Provides cluster_endpoints output"
    check_file_content "$provider" "outputs.tf" "cluster_type.*module.design.cluster_type" "Uses standardized cluster_type"
    check_file_content "$provider" "outputs.tf" "master_count.*module.design.master_count" "Uses standardized master_count"
}

# Validate common design module
echo -e "\n${BLUE}🔧 Validating Common Design Module${NC}"
echo "----------------------------------------"
check_file_content "common/design" "main.tf" "has_master_nodes" "Detects k3s master nodes"
check_file_content "common/design" "main.tf" "master_count" "Counts master nodes"
check_file_content "common/design" "outputs.tf" "cluster_type" "Outputs cluster type"
check_file_content "common/design" "outputs.tf" "master_count" "Outputs master count"

# Validate common variables
echo -e "\n${BLUE}📝 Validating Common Variables${NC}"
echo "----------------------------------------"
check_file_content "common" "variables.tf" "master.*ansible" "Validates k3s requirements"

# List of providers to validate
PROVIDERS=("aws" "gcp" "azure" "openstack" "ovh")

# Validate each provider
for provider in "${PROVIDERS[@]}"; do
    validate_provider "$provider"
done

# Validate examples
echo -e "\n${BLUE}📚 Validating k3s Examples${NC}"
echo "----------------------------------------"
check_file_content "examples" "aws-k3s-cluster.tf" "count = 3.*Always 3 for HA k3s" "AWS example enforces 3 control plane nodes"
check_file_content "examples" "gcp-k3s-cluster.tf" "count = 3.*Always 3 for HA k3s" "GCP example enforces 3 control plane nodes"

# Check for template files
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
if [[ -f "$CLOUD_DIR/common/infrastructure_template.tf" ]]; then
    echo -e "  ${GREEN}✓${NC} Infrastructure template provided"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
else
    echo -e "  ${RED}✗${NC} Infrastructure template missing"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
fi

TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
if [[ -f "$CLOUD_DIR/common/network_template.tf" ]]; then
    echo -e "  ${GREEN}✓${NC} Network template provided"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
else
    echo -e "  ${RED}✗${NC} Network template missing"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
fi

# Summary
echo -e "\n${BLUE}📊 Validation Summary${NC}"
echo "===================="
echo -e "Total checks: $TOTAL_CHECKS"
echo -e "${GREEN}Passed: $PASSED_CHECKS${NC}"
echo -e "${RED}Failed: $FAILED_CHECKS${NC}"

if [[ $FAILED_CHECKS -eq 0 ]]; then
    echo -e "\n${GREEN}🎉 All k3s standards validation checks passed!${NC}"
    echo -e "The cloud providers are properly standardized for k3s deployment."
    exit 0
else
    echo -e "\n${RED}❌ $FAILED_CHECKS validation checks failed.${NC}"
    echo -e "Please review the failed checks and ensure all providers follow k3s standards."
    exit 1
fi
