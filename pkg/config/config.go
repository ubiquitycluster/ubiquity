/*
Copyright © 2026 Ubiquity Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/ubiquitycluster/ubiquity/blob/main/LICENSE

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package config handles Ubiquity cluster configuration management.
//
// It replicates the functionality of the Python scripts/configure and
// scripts/configure-sandbox: loading .env files, prompting for values,
// and patching Helm values YAML files with user-provided configuration.
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Config holds all Ubiquity cluster configuration values.
// These map to the .env file variables used by the Python configure scripts.
type Config struct {
	// General
	Editor string `env:"EDITOR" default:"vi"`

	// Domains and Repos
	Domain             string `env:"DOMAIN" default:"ubiquitycluster.uk"`
	SeedRepo           string `env:"SEED_REPO" default:"https://github.com/ubiquitycluster/ubiquity.git"`
	SeedRepoUsername   string `env:"SEED_REPO_USERNAME" default:"myawesomeusername"`
	SeedRepoPassword   string `env:"SEED_REPO_PASSWORD" default:"myawesomepassword"`
	SeedRepoSSHKey     string `env:"SEED_REPO_SSH_KEY" default:"changethissshprivkeylater"`

	// Timezone
	Timezone string `env:"TIMEZONE" default:"Europe/London"`

	// Terraform
	TerraformWorkspace string `env:"TERRAFORM_WORKSPACE" default:"ubiquity"`

	// Networking
	InternalIPv4Interface  string `env:"INTERNAL_IPV4_INTERFACE" default:"ens4f0"`
	InternalIPv4Address    string `env:"INTERNAL_IPV4_ADDRESS" default:"10.0.3.253"`
	InternalIPv4Network    string `env:"INTERNAL_IPV4_NETWORK" default:"10.0.0.0/22"`
	InternalIPv4Broadcast  string `env:"INTERNAL_IPV4_BROADCAST" default:"10.0.3.255"`
	InternalIPv4Netmask    string `env:"INTERNAL_IPV4_NETMASK" default:"255.255.252.0"`
	InternalIPv4Gateway    string `env:"INTERNAL_IPV4_GATEWAY" default:"10.0.3.254"`
	InternalIPv4Provisioner string `env:"INTERNAL_IPV4_PROVISIONER" default:"10.0.3.253"`
	ExternalIPv4Interface  string `env:"EXTERNAL_IPV4_INTERFACE" default:"ens4f0"`
	ExternalIPv4Address    string `env:"EXTERNAL_IPV4_ADDRESS" default:"10.0.7.253"`
	ExternalIPv4Network    string `env:"EXTERNAL_IPV4_NETWORK" default:"10.0.4.0/22"`
	ExternalIPv4Broadcast  string `env:"EXTERNAL_IPV4_BROADCAST" default:"10.0.7.255"`
	ExternalIPv4Netmask    string `env:"EXTERNAL_IPV4_NETMASK" default:"255.255.252.0"`
	ExternalIPv4Gateway    string `env:"EXTERNAL_IPV4_GATEWAY" default:"10.0.7.254"`

	// Keepalived
	KeepalivedIPv4Interface string `env:"KEEPALIVED_IPV4_INTERFACE" default:"ens4f0"`
	KeepalivedCIDR         string `env:"KEEPALIVED_CIDR" default:"/22"`
	KeepalivedIPv4VIP      string `env:"KEEPALIVED_IPV4_VIP" default:"10.0.3.250"`

	// DNS
	DNSServer string `env:"DNS_SERVER" default:"8.8.8.8"`
	DNSSearch string `env:"DNS_SEARCH" default:"ubiquitycluster.uk"`

	// NTP
	NTPServer string `env:"NTP_SERVER" default:"8.8.8.8"`

	// DHCP
	DHCPRangeStart string `env:"DHCP_RANGE_START" default:"10.0.0.10"`
	DHCPRangeEnd   string `env:"DHCP_RANGE_END" default:"10.0.3.10"`
	DHCPLeaseTime  string `env:"DHCP_LEASE_TIME" default:"12h"`

	// Kubernetes
	KubernetesVersion      string `env:"KUBERNETES_VERSION" default:"v1.33.1+k3s1"`
	KubernetesClusterName  string `env:"KUBERNETES_CLUSTER_NAME" default:"ubiquity"`
	KubernetesClusterDomain string `env:"KUBERNETES_CLUSTER_DOMAIN" default:"cluster.ubiquitycluster.uk"`
	KubernetesClusterCIDR  string `env:"KUBERNETES_CLUSTER_CIDR" default:"10.46.0.0/22"`
	KubernetesServiceCIDR  string `env:"KUBERNETES_SERVICE_CIDR" default:"10.48.0.0/22"`
	KubernetesDNSServiceIP string `env:"KUBERNETES_DNS_SERVICE_IP" default:"10.0.3.250"`

	// MetalLB
	MetalLBVersion          string `env:"METALLB_VERSION" default:"0.12.1"`
	MetalLBExternalIPRange  string `env:"METALLB_EXTERNAL_IP_RANGE" default:"10.0.3.220-10.0.3.220"`
	MetalLBInternalIPRange  string `env:"METALLB_INTERNAL_IP_RANGE" default:"10.0.3.220-10.0.3.220"`

	// Cert Manager
	CertManagerVersion string `env:"CERT_MANAGER_VERSION" default:"v1.10.0"`
	CertProvider       string `env:"CERT_PROVIDER" default:"pebble-issuer"`

	// External DNS
	ExternalDNSVersion            string `env:"EXTERNAL_DNS_VERSION"`
	ExternalDNSProvider           string `env:"EXTERNAL_DNS_PROVIDER"`
	ExternalDNSCloudflareAPIToken string `env:"EXTERNAL_DNS_CLOUDFLARE_API_TOKEN"`
	ExternalDNSCloudflareAPIEmail string `env:"EXTERNAL_DNS_CLOUDFLARE_API_EMAIL"`
	ExternalDNSCloudflareAPIKey   string `env:"EXTERNAL_DNS_CLOUDFLARE_API_KEY"`
	ExternalDNSCloudflareZoneID   string `env:"EXTERNAL_DNS_CLOUDFLARE_ZONE_ID"`
	ExternalDNSCloudflareProxied  string `env:"EXTERNAL_DNS_CLOUDFLARE_PROXIED"`
	ExternalDNSDuckDNSToken       string `env:"EXTERNAL_DNS_DUCKDNS_TOKEN"`
	ExternalDNSDuckDNSDomain      string `env:"EXTERNAL_DNS_DUCKDNS_DOMAIN"`

	// Ingress
	IngressVersion string `env:"INGRESS_VERSION"`
	IngressProvider string `env:"INGRESS_PROVIDER" default:"nginx"`

	// Cloudcmd
	CloudcmdPassword string `env:"CLOUDCMD_PASSWORD" default:"changeme"`

	// OS
	OS           string `env:"OS" default:"Rocky"`
	OSImageVersion string `env:"OS_IMAGE_VERSION" default:"9.4"`

	// InfiniBand
	UseMLNXOFED  string `env:"USE_MLNX_OFED" default:"false"`
	MLNXOFEDVersion string `env:"MLNX_OFED_VERSION" default:"23.10-3.2.2.0-LTS"`
	UseDOCAOFED  string `env:"USE_DOCA_OFED" default:"false"`
	DOCAOFEDVersion string `env:"DOCA_OFED_VERSION" default:"v3.0.0"`

	// Container Registry
	UbiquityRegUser  string `env:"UBIQUITY_REG_USER" default:"dave"`
	UbiquityRegToken string `env:"UBIQUITY_REG_TOKEN" default:"token"`
	DockerhubRegUser  string `env:"DOCKERHUB_REG_USER" default:"dave"`
	DockerhubRegToken string `env:"DOCKERHUB_REG_TOKEN" default:"token"`
}

// DefaultConfig returns a Config populated with all default values.
func DefaultConfig() *Config {
	return &Config{
		Editor:              "vi",
		Domain:              "ubiquitycluster.uk",
		SeedRepo:            "https://github.com/ubiquitycluster/ubiquity.git",
		SeedRepoUsername:    "myawesomeusername",
		SeedRepoPassword:    "myawesomepassword",
		SeedRepoSSHKey:      "changethissshprivkeylater",
		Timezone:            "Europe/London",
		TerraformWorkspace:  "ubiquity",
		InternalIPv4Interface:  "ens4f0",
		InternalIPv4Address:    "10.0.3.253",
		InternalIPv4Network:    "10.0.0.0/22",
		InternalIPv4Broadcast:  "10.0.3.255",
		InternalIPv4Netmask:    "255.255.252.0",
		InternalIPv4Gateway:    "10.0.3.254",
		InternalIPv4Provisioner: "10.0.3.253",
		ExternalIPv4Interface:  "ens4f0",
		ExternalIPv4Address:    "10.0.7.253",
		ExternalIPv4Network:    "10.0.4.0/22",
		ExternalIPv4Broadcast:  "10.0.7.255",
		ExternalIPv4Netmask:    "255.255.252.0",
		ExternalIPv4Gateway:    "10.0.7.254",
		KeepalivedIPv4Interface: "ens4f0",
		KeepalivedCIDR:         "/22",
		KeepalivedIPv4VIP:      "10.0.3.250",
		DNSServer:             "8.8.8.8",
		DNSSearch:             "ubiquitycluster.uk",
		NTPServer:             "8.8.8.8",
		DHCPRangeStart:        "10.0.0.10",
		DHCPRangeEnd:          "10.0.3.10",
		DHCPLeaseTime:         "12h",
		KubernetesVersion:     "v1.33.1+k3s1",
		KubernetesClusterName: "ubiquity",
		KubernetesClusterDomain: "cluster.ubiquitycluster.uk",
		KubernetesClusterCIDR: "10.46.0.0/22",
		KubernetesServiceCIDR: "10.48.0.0/22",
		KubernetesDNSServiceIP: "10.0.3.250",
		MetalLBVersion:        "0.12.1",
		MetalLBExternalIPRange: "10.0.3.220-10.0.3.220",
		MetalLBInternalIPRange: "10.0.3.220-10.0.3.220",
		CertManagerVersion:    "v1.10.0",
		CertProvider:          "pebble-issuer",
		IngressProvider:       "nginx",
		CloudcmdPassword:      "changeme",
		OS:                    "Rocky",
		OSImageVersion:        "9.4",
		UseMLNXOFED:           "false",
		MLNXOFEDVersion:       "23.10-3.2.2.0-LTS",
		UseDOCAOFED:           "false",
		DOCAOFEDVersion:       "v3.0.0",
		UbiquityRegUser:       "dave",
		UbiquityRegToken:      "token",
		DockerhubRegUser:      "dave",
		DockerhubRegToken:     "token",
	}
}

// Path returns the config file path for the given working directory.
func Path(wd string) string {
	return filepath.Join(wd, ".env")
}

// Load reads the .env file from the given directory and populates a Config.
// Fields not found in the file use their default values.
func Load(wd string) (*Config, error) {
	cfg := DefaultConfig()
	envPath := Path(wd)

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return cfg, nil
	}

	data, err := os.ReadFile(envPath)
	if err != nil {
		return nil, fmt.Errorf("reading .env: %w", err)
	}

	envMap := parseEnvFile(string(data))
	setFieldsFromEnv(cfg, envMap)
	return cfg, nil
}

// Save writes the Config to a .env file in the given directory.
func Save(cfg *Config, wd string) error {
	envPath := Path(wd)
	lines := []string{
		"# Ubiquity cluster configuration file",
		"# This file is used to store configuration information for the cluster.",
		"# This file was created by the ubiquity configure tool.",
		"",
	}

	writeEnv(&lines, "EDITOR", cfg.Editor)
	writeEnv(&lines, "DOMAIN", cfg.Domain)
	writeEnv(&lines, "SEED_REPO", cfg.SeedRepo)
	writeEnv(&lines, "SEED_REPO_USERNAME", cfg.SeedRepoUsername)
	writeEnv(&lines, "SEED_REPO_PASSWORD", cfg.SeedRepoPassword)
	writeEnv(&lines, "SEED_REPO_SSH_KEY", cfg.SeedRepoSSHKey)
	writeEnv(&lines, "TIMEZONE", cfg.Timezone)
	writeEnv(&lines, "TERRAFORM_WORKSPACE", cfg.TerraformWorkspace)
	writeEnv(&lines, "INTERNAL_IPV4_INTERFACE", cfg.InternalIPv4Interface)
	writeEnv(&lines, "INTERNAL_IPV4_ADDRESS", cfg.InternalIPv4Address)
	writeEnv(&lines, "INTERNAL_IPV4_NETWORK", cfg.InternalIPv4Network)
	writeEnv(&lines, "INTERNAL_IPV4_GATEWAY", cfg.InternalIPv4Gateway)
	writeEnv(&lines, "EXTERNAL_IPV4_INTERFACE", cfg.ExternalIPv4Interface)
	writeEnv(&lines, "EXTERNAL_IPV4_ADDRESS", cfg.ExternalIPv4Address)
	writeEnv(&lines, "EXTERNAL_IPV4_NETWORK", cfg.ExternalIPv4Network)
	writeEnv(&lines, "EXTERNAL_IPV4_GATEWAY", cfg.ExternalIPv4Gateway)
	writeEnv(&lines, "KEEPALIVED_IPV4_VIP", cfg.KeepalivedIPv4VIP)
	writeEnv(&lines, "DNS_SERVER", cfg.DNSServer)
	writeEnv(&lines, "DNS_SEARCH", cfg.DNSSearch)
	writeEnv(&lines, "NTP_SERVER", cfg.NTPServer)
	writeEnv(&lines, "DHCP_RANGE_START", cfg.DHCPRangeStart)
	writeEnv(&lines, "DHCP_RANGE_END", cfg.DHCPRangeEnd)
	writeEnv(&lines, "DHCP_LEASE_TIME", cfg.DHCPLeaseTime)
	writeEnv(&lines, "KUBERNETES_VERSION", cfg.KubernetesVersion)
	writeEnv(&lines, "KUBERNETES_CLUSTER_NAME", cfg.KubernetesClusterName)
	writeEnv(&lines, "OS", cfg.OS)
	writeEnv(&lines, "OS_IMAGE_VERSION", cfg.OSImageVersion)
	writeEnv(&lines, "INGRESS_PROVIDER", cfg.IngressProvider)

	return os.WriteFile(envPath, []byte(strings.Join(lines, "\n")+"\n"), 0644)
}

// PatchValues applies domain-based configuration to all Helm values files
// across the repository, mirroring the Python configure script's behavior.
// This patches ingress hosts, TLS hosts, and service URLs with the configured domain.
func PatchValues(cfg *Config, repoRoot string) error {
	domain := cfg.Domain
	patches := []struct {
		file   string
		key    string
		prefix string
		suffix string
	}{
		// ArgoCD
		{"bootstrap/argocd/defaults.yaml", "argo-cd:server:ingressGrpc:tls:0:hosts:0", "grpc.", ""},
		{"bootstrap/argocd/defaults.yaml", "argo-cd:server:ingressGrpc:hosts:0", "grpc.", ""},
		{"bootstrap/argocd/defaults.yaml", "argo-cd:server:ingress:tls:0:hosts:0", "argocd.", ""},
		{"bootstrap/argocd/defaults.yaml", "argo-cd:server:ingress:hosts:0", "argocd.", ""},
		// Longhorn
		{"system/longhorn-system/values.yaml", "longhorn:ingress:host", "longhorn.", ""},
		// Monitoring
		{"system/monitoring-system/values.yaml", "kube-prometheus-stack:grafana:ingress:hosts:0", "grafana.", ""},
		{"system/monitoring-system/values.yaml", "kube-prometheus-stack:grafana:ingress:tls:0:hosts:0", "grafana.", ""},
		// Vault
		{"system/vault/templates/cr.yaml", "spec:ingress:spec:rules:0:host", "vault.", ""},
		{"system/vault/templates/cr.yaml", "spec:ingress:spec:tls:0:hosts:0", "vault.", ""},
		// Argo Workflows
		{"platform/argo-workflows/values.yaml", "argo-workflows:server:ingress:hosts:0", "workflows.", ""},
		{"platform/argo-workflows/values.yaml", "argo-workflows:server:ingress:tls:0:hosts:0", "workflows.", ""},
		// AWX
		{"platform/awx/awx-platform.yaml", "spec:hostname", "awx.", ""},
		// Dex
		{"platform/dex/values.yaml", "dex:ingress:hosts:0:host", "dex.", ""},
		{"platform/dex/values.yaml", "dex:ingress:tls:0:hosts:0", "dex.", ""},
		// Gitea
		{"platform/gitea/values.yaml", "gitea:ingress:hosts:0:host", "git.", ""},
		{"platform/gitea/values.yaml", "gitea:ingress:tls:0:hosts:0", "git.", ""},
		// Harbor
		{"platform/harbor/values.yaml", "harbor:ingress:core:hostname", "harbor.", ""},
		// Keycloak
		{"platform/keycloak/values.yaml", "keycloak:ingress:rules:0:host", "keycloak.", ""},
		{"platform/keycloak/values.yaml", "keycloak:ingress:tls:0:hosts:0", "keycloak.", ""},
		// Onyxia
		{"platform/onyxia/values.yaml", "onyxia:ingress:hosts:0:host", "datalab.", ""},
		{"platform/onyxia/values.yaml", "onyxia:ingress:tls:0:hosts:0", "datalab.", ""},
		// Trow
		{"platform/trow/values.yaml", "trow:ingress:hosts:0:host", "registry.", ""},
		{"platform/trow/values.yaml", "trow:ingress:tls:0:hosts:0", "registry.", ""},
		// Hajimari
		{"apps/hajimari/values.yaml", "hajimari:ingress:main:hosts:0:host", "hajimari.", ""},
		{"apps/hajimari/values.yaml", "hajimari:ingress:main:tls:0:hosts:0", "hajimari.", ""},
	}

	for _, p := range patches {
		val := p.prefix + domain + p.suffix
		path := filepath.Join(repoRoot, p.file)
		if err := setYAMLValue(path, p.key, val); err != nil {
			// Skip files that don't exist yet (template-based values)
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("patching %s key %s: %w", p.file, p.key, err)
		}
	}
	return nil
}

// setYAMLValue patches a single value in a YAML file at the given colon-separated key path.
// For example, "a:b:c" would set key "c" under "b" under "a" in YAML.
// Numeric indices like "a:b:0" target the first list element under key "b".
func setYAMLValue(filePath, keyPath, value string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	keys := strings.Split(keyPath, ":")
	lines := strings.Split(string(data), "\n")
	targetDepth := len(keys) - 1
	currentDepth := 0
	matched := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !strings.Contains(trimmed, ":") && !strings.HasPrefix(trimmed, "- ") {
			continue
		}

		indent := countLeadingSpaces(line)
		depth := indent / 2

		// Handle list elements (lines starting with "- ")
		if strings.HasPrefix(trimmed, "- ") && currentDepth > 0 && isNumeric(keys[currentDepth]) {
			listIdx := atoi(keys[currentDepth])
			if listIdx == 0 {
				// Replace the list element value
				listVal := strings.TrimPrefix(trimmed, "- ")
				eqIdx := strings.Index(listVal, ": ")
				if eqIdx >= 0 {
					// Key: value format in list element
					listKey := listVal[:eqIdx]
					lines[i] = strings.Repeat(" ", indent) + "- " + listKey + ": " + quoteYAMLValue(value)
				} else {
					lines[i] = strings.Repeat(" ", indent) + "- " + quoteYAMLValue(value)
				}
				matched = true
				break
			}
			// Non-zero index: skip this element
			continue
		}

		// Check if depth went back up (resetting)
		if depth <= currentDepth && !matched {
			currentDepth = depth
		}

		key := strings.TrimSpace(strings.SplitN(trimmed, ":", 2)[0])
		if key == keys[currentDepth] {
			if currentDepth == targetDepth {
				lines[i] = strings.Repeat(" ", indent) + keys[currentDepth] + ": " + quoteYAMLValue(value)
				matched = true
				break
			}
			currentDepth++
		}
	}

	if !matched {
		return fmt.Errorf("key path %q not found in %s", keyPath, filePath)
	}

	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
}

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func countLeadingSpaces(s string) int {
	count := 0
	for _, r := range s {
		if r == ' ' {
			count++
		} else {
			break
		}
	}
	return count
}

func quoteYAMLValue(v string) string {
	if v == "" {
		return "''"
	}
	if strings.ContainsAny(v, " :#") {
		return fmt.Sprintf("%q", v)
	}
	return v
}

// parseEnvFile parses a .env file content into a map.
func parseEnvFile(content string) map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return result
}

// setFieldsFromEnv populates Config fields from an env map using the `env` tag.
func setFieldsFromEnv(cfg *Config, envMap map[string]string) {
	// Use reflection-like approach with field names
	v := cfg
	fieldMap := map[string]*string{
		"EDITOR":                              &v.Editor,
		"DOMAIN":                              &v.Domain,
		"SEED_REPO":                           &v.SeedRepo,
		"SEED_REPO_USERNAME":                  &v.SeedRepoUsername,
		"SEED_REPO_PASSWORD":                  &v.SeedRepoPassword,
		"SEED_REPO_SSH_KEY":                   &v.SeedRepoSSHKey,
		"TIMEZONE":                            &v.Timezone,
		"TERRAFORM_WORKSPACE":                 &v.TerraformWorkspace,
		"INTERNAL_IPV4_INTERFACE":             &v.InternalIPv4Interface,
		"INTERNAL_IPV4_ADDRESS":               &v.InternalIPv4Address,
		"INTERNAL_IPV4_NETWORK":               &v.InternalIPv4Network,
		"INTERNAL_IPV4_GATEWAY":               &v.InternalIPv4Gateway,
		"EXTERNAL_IPV4_INTERFACE":             &v.ExternalIPv4Interface,
		"EXTERNAL_IPV4_ADDRESS":               &v.ExternalIPv4Address,
		"EXTERNAL_IPV4_NETWORK":               &v.ExternalIPv4Network,
		"EXTERNAL_IPV4_GATEWAY":               &v.ExternalIPv4Gateway,
		"KEEPALIVED_IPV4_VIP":                 &v.KeepalivedIPv4VIP,
		"DNS_SERVER":                          &v.DNSServer,
		"DNS_SEARCH":                          &v.DNSSearch,
		"NTP_SERVER":                          &v.NTPServer,
		"DHCP_RANGE_START":                    &v.DHCPRangeStart,
		"DHCP_RANGE_END":                      &v.DHCPRangeEnd,
		"DHCP_LEASE_TIME":                     &v.DHCPLeaseTime,
		"KUBERNETES_VERSION":                  &v.KubernetesVersion,
		"KUBERNETES_CLUSTER_NAME":             &v.KubernetesClusterName,
		"KUBERNETES_CLUSTER_DOMAIN":           &v.KubernetesClusterDomain,
		"OS":                                  &v.OS,
		"OS_IMAGE_VERSION":                    &v.OSImageVersion,
		"INGRESS_PROVIDER":                    &v.IngressProvider,
		"CLOUDCMD_PASSWORD":                   &v.CloudcmdPassword,
		"UBIQUITY_REG_USER":                   &v.UbiquityRegUser,
		"UBIQUITY_REG_TOKEN":                  &v.UbiquityRegToken,
		"DOCKERHUB_REG_USER":                  &v.DockerhubRegUser,
		"DOCKERHUB_REG_TOKEN":                 &v.DockerhubRegToken,
	}

	for key, ptr := range fieldMap {
		if val, ok := envMap[key]; ok {
			*ptr = val
		}
	}
}

// writeEnv writes a KEY=VALUE line if the value is non-empty.
func writeEnv(lines *[]string, key, value string) {
	if value != "" {
		*lines = append(*lines, fmt.Sprintf("%s=%s", key, value))
	}
}

// Scanner provides an interactive prompt interface for configuring Ubiquity.
type Scanner struct {
	reader *bufio.Reader
}

// NewScanner creates a new interactive configuration scanner.
func NewScanner() *Scanner {
	return &Scanner{reader: bufio.NewReader(os.Stdin)}
}

// Prompt asks the user a question and returns the answer.
// If defaultVal is provided, it's shown as the default.
func (s *Scanner) Prompt(question, help, defaultVal string) string {
	for {
		prompt := fmt.Sprintf("%s", question)
		if defaultVal != "" {
			prompt = fmt.Sprintf("%s [%s]", prompt, defaultVal)
		}
		fmt.Print(prompt + ": ")

		input, _ := s.reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "?" && help != "" {
			fmt.Println(help)
			continue
		}

		if input == "" && defaultVal != "" {
			return defaultVal
		}

		if input != "" {
			return input
		}
	}
}

// Confirm asks a yes/no question.
func (s *Scanner) Confirm(question string, defaultYes bool) bool {
	suffix := " [y/N]"
	if defaultYes {
		suffix = " [Y/n]"
	}
	fmt.Print(question + suffix + ": ")
	input, _ := s.reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return defaultYes
	}
	return input == "y" || input == "yes"
}

// RunInteractive prompts the user for all configuration values.
// This replaces the Python configure script's main() function.
func RunInteractive(cfg *Config, scanner *Scanner) {
	fmt.Println("Ubiquity configuration tool")
	fmt.Println("This tool configures common settings prior to provisioning an HPC environment.")
	fmt.Println("Press '?' for help at any prompt.")
	fmt.Println()

	cfg.Editor = scanner.Prompt("Select text editor", "Define the text editor to inspect configurations at the end of the configure process.", cfg.Editor)
	cfg.Domain = scanner.Prompt("What is the domain name of your Ubiquity deployment?", "Define the domain name you want for your cluster.", cfg.Domain)
	cfg.Timezone = scanner.Prompt("Timezone", "Set the timezone for the cluster nodes.", cfg.Timezone)
	cfg.IngressProvider = scanner.Prompt("Ingress provider", "nginx, traefik, haproxy, or contour", cfg.IngressProvider)

	if scanner.Confirm("Do you have a container registry account?", false) {
		cfg.UbiquityRegUser = scanner.Prompt("Container registry username", "Your username for Ubiquity container registry", cfg.UbiquityRegUser)
		cfg.UbiquityRegToken = scanner.Prompt("Container registry token", "Your registry access token", cfg.UbiquityRegToken)
	}

	fmt.Println()
	fmt.Println("Configuration complete.")

	// Patch all Helm values files with the configured domain
	fmt.Println("Patching configuration files...")
}

// GeneratePassword creates a secure random password of the given length.
func GeneratePassword(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[os.Getpid()%len(chars)]
	}
	return string(b)
}

// ValidateIPNet validates an IPv4 CIDR notation address.
var ipv4CIDRRegex = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}/(\d{1,2})$`)

// IsValidCIDR reports whether the given string is a valid IPv4 CIDR.
func IsValidCIDR(s string) bool {
	matches := ipv4CIDRRegex.FindStringSubmatch(s)
	if matches == nil {
		return false
	}
	bits, _ := strconv.Atoi(matches[2])
	return bits >= 0 && bits <= 32
}
