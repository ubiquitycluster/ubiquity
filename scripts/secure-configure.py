#!/usr/bin/env python3
# Copyright The Ubiquity Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/ubiquitycluster/ubiquity/blob/main/LICENSE

"""
Secure configuration script for Ubiquity using SOPS
Updates configuration files in-place and configures ArgoCD with SOPS support
"""

import os
import sys
import json
import yaml
import subprocess
import getpass
import fileinput
import re
import base64
import bcrypt
import ipaddress
from pathlib import Path
from ruamel.yaml import YAML
from jinja2 import Environment, FileSystemLoader, StrictUndefined
from rich.prompt import Confirm, Prompt
from rich.console import Console
from rich.panel import Panel

console = Console()

class SOPSConfig:
    def __init__(self, config_dir=".ubiquity"):
        self.config_dir = Path(config_dir)
        self.config_dir.mkdir(exist_ok=True)
        self.secrets_file = self.config_dir / "secrets.yaml"
        self.config_file = self.config_dir / "config.yaml"
        self.sops_config_file = Path(".sops.yaml")
        self.age_key_file = Path.home() / ".config" / "sops" / "age" / "keys.txt"
        self.config = {}
        self.secrets = {}
        # IP pool management for MetalLB
        self.metallb_ip_pools = {}
        self.assigned_ips = set()
        
    def _check_sops_installed(self) -> bool:
        """Check if SOPS is installed and available"""
        try:
            result = subprocess.run(['sops', '--version'], 
                                  capture_output=True, text=True, check=True)
            console.print(f"✅ SOPS found: {result.stdout.strip()}", style="green")
            return True
        except (subprocess.CalledProcessError, FileNotFoundError):
            console.print("❌ SOPS not found. Please install SOPS first:", style="red")
            console.print("  - macOS: brew install sops", style="blue")
            console.print("  - Linux: https://github.com/mozilla/sops/releases", style="blue")
            return False
    
    def _setup_age_key(self) -> str:
        """Setup age key for SOPS"""
        if self.age_key_file.exists():
            # Extract existing public key
            with open(self.age_key_file, 'r') as f:
                content = f.read()
                for line in content.split('\n'):
                    if line.startswith('# public key:'):
                        public_key = line.split(': ')[1]
                        console.print(f"✅ Using existing age key: {public_key}", style="green")
                        return public_key
        
        console.print("⚠️ No age key found. Creating one...", style="yellow")
        
        try:
            # Create age key directory
            self.age_key_file.parent.mkdir(parents=True, exist_ok=True)
            
            # Generate age key
            result = subprocess.run(['age-keygen'], 
                                  capture_output=True, text=True, check=True)
 
            # Save key to file
            self.age_key_file.write_text(result.stdout)
            self.age_key_file.chmod(0o600)
            
            # Extract public key
            lines = result.stdout.strip().split('\n')
            for line in lines:
                if line.startswith('# public key:'):
                    public_key = line.split(': ')[1]
                    console.print(f"✅ Generated new age key: {public_key}", style="green")
                    return public_key
                    
        except (subprocess.CalledProcessError, FileNotFoundError) as e:
            console.print(f"❌ Failed to generate age key: {e}", style="red")
            console.print("Please install age: https://github.com/FiloSottile/age", style="blue")
        
        return None
    
    def _update_file_in_place(self, file_path: str, updates: dict, comment_char='#'):
        """Update configuration file in-place preserving formatting"""
        if not Path(file_path).exists():
            console.print(f"⚠️ File {file_path} not found, skipping",
                          style="yellow")
            return
        
        console.print(f"📝 Updating {file_path}...", style="blue")
        
        # Read file and apply updates
        with fileinput.input(file_path, inplace=True) as file:
            for line in file:
                # Skip comments and empty lines
                stripped_line = line.strip()
                if (stripped_line.startswith(comment_char) or
                    stripped_line.startswith('##') or
                    not stripped_line):
                    print(line, end='')
                    continue
                
                updated = False
                for key, value in updates.items():
                    # Only update if the value is not empty/None
                    if not value:
                        continue
                    
                    # Match YAML format preserving exact indentation
                    yaml_pattern = rf'^(\s*){re.escape(key)}(\s*:\s*)(.*)$'
                    match = re.match(yaml_pattern, line)
                    
                    if match:
                        # Preserve exact original indentation
                        original_indent = match.group(1)
                        colon_and_spaces = match.group(2)
                        
                        if isinstance(value, bool):
                            value_str = str(value).lower()
                        elif (isinstance(value, str) and
                              (' ' in value or value.startswith('v'))):
                            # Quote strings with spaces or version strings
                            if not value.startswith('"'):
                                value_str = f'"{value}"'
                            else:
                                value_str = value
                        else:
                            value_str = str(value)
                        
                        # Build output preserving original formatting
                        output_line = (f'{original_indent}{key}'
                                     f'{colon_and_spaces}{value_str}')
                        print(output_line)
                        updated = True
                        break
                
                if not updated:
                    print(line, end='')
    
    def _update_yaml_file_in_place(self, file_path: str, updates: dict):
        """Update YAML file in-place preserving structure"""
        if not Path(file_path).exists():
            console.print(f"⚠️ File {file_path} not found, skipping",
                          style="yellow")
            return
        
        console.print(f"📝 Updating YAML {file_path}...", style="blue")
        
        yaml_handler = YAML()
        yaml_handler.preserve_quotes = True
        yaml_handler.map_indent = 2
        yaml_handler.sequence_indent = 2  # Changed from 4 to 2
        yaml_handler.sequence_dash_offset = 0
        yaml_handler.width = 4096  # Prevent line wrapping
        yaml_handler.default_flow_style = False
        yaml_handler.allow_unicode = True
        
        with open(file_path, 'r') as f:
            data = yaml_handler.load(f) or {}
        
        # Apply nested updates
        self._apply_nested_updates(data, updates)
        
        with open(file_path, 'w') as f:
            yaml_handler.dump(data, f)
    
    def _apply_nested_updates(self, data: dict, updates: dict):
        """Apply nested dictionary updates"""
        for key, value in updates.items():
            if '.' in key:
                # Handle nested keys like 'server.config.url'
                keys = key.split('.')
                current = data
                for k in keys[:-1]:
                    if k not in current:
                        current[k] = {}
                    current = current[k]
                current[keys[-1]] = value
            else:
                data[key] = value
    
    def _decrypt_secrets(self) -> dict:
        """Decrypt secrets using SOPS"""
        if not self.secrets_file.exists():
            return {}
        
        try:
            result = subprocess.run([
                'sops', '--decrypt', str(self.secrets_file)
            ], capture_output=True, text=True, check=True)
            
            return yaml.safe_load(result.stdout) or {}
            
        except subprocess.CalledProcessError as e:
            console.print(f"❌ Failed to decrypt secrets: {e}", style="red")
            return {}
    
    def _encrypt_secrets(self, data: dict):
        """Encrypt secrets using SOPS"""
        try:
            # Ensure the secrets file exists in the right location
            console.print(f"📁 Encrypting secrets to: {self.secrets_file}", style="blue")
            
            # Write plaintext secrets directly to final location
            yaml_handler = YAML()
            yaml_handler.preserve_quotes = True
            yaml_handler.explicit_start = True
            yaml_handler.default_flow_style = False
            
            with open(self.secrets_file, 'w') as f:
                yaml_handler.dump(data, f)
            
            # Debug: Check if .sops.yaml exists and show its contents
            if self.sops_config_file.exists():
                console.print("📋 SOPS configuration found", style="green")
                with open(self.sops_config_file, 'r') as f:
                    sops_content = f.read()
                    console.print(f"SOPS config:\n{sops_content}", style="dim")
            else:
                console.print("❌ No .sops.yaml found", style="red")
            
            # Try to encrypt with verbose output
            result = subprocess.run([
                'sops', '--encrypt', '--in-place', str(self.secrets_file)
            ], capture_output=True, text=True)
            
            if result.returncode != 0:
                console.print(f"SOPS error output: {result.stderr}", style="red")
                console.print(f"SOPS stdout: {result.stdout}", style="dim")
                raise subprocess.CalledProcessError(result.returncode, result.args)
            
            console.print("🔐 Secrets encrypted with SOPS", style="green")
            
        except subprocess.CalledProcessError as e:
            console.print(f"❌ Failed to encrypt secrets: {e}", style="red")
            # Clean up the plaintext file if encryption failed
            if self.secrets_file.exists():
                self.secrets_file.unlink()
            raise
    
    def configure_argocd_sops(self):
        """Configure ArgoCD with SOPS support"""
        console.print("\n🔧 Configuring ArgoCD with SOPS support...", style="bold")
        
        # 1. Create age key secret for ArgoCD
        self._create_argocd_age_secret()
        
        # 2. Update ArgoCD configuration to use SOPS
        self._update_argocd_config_for_sops()
        
        # 3. Create SOPS configuration in ArgoCD namespace
        self._create_argocd_sops_config()
    
    def _create_argocd_age_secret(self):
        """Create Kubernetes secret with age key for ArgoCD"""
        if not self.age_key_file.exists():
            console.print("❌ Age key file not found", style="red")
            return
        
        # Read age private key
        with open(self.age_key_file, 'r') as f:
            age_key_content = f.read()
        
        # Create secret manifest
        secret_manifest = {
            "apiVersion": "v1",
            "kind": "Secret",
            "metadata": {
                "name": "sops-age-key",
                "namespace": "argocd"
            },
            "type": "Opaque",
            "stringData": {
                "keys.txt": age_key_content
            }
        }
        
        # Save to bootstrap directory
        secrets_dir = Path("bootstrap/argocd/secrets")
        secrets_dir.mkdir(parents=True, exist_ok=True)
        
        yaml_handler = YAML()
        yaml_handler.preserve_quotes = True
        
        with open(secrets_dir / "sops-age-key.yaml", 'w') as f:
            yaml_handler.dump(secret_manifest, f)
        
        console.print("✅ Created ArgoCD age key secret", style="green")
    
    def _update_argocd_config_for_sops(self):
        """Update ArgoCD configuration to support SOPS"""
        argocd_config_files = [
            "bootstrap/argocd/defaults.yaml",
            "bootstrap/argocd/values.yaml"
        ]
        
        for config_file in argocd_config_files:
            if Path(config_file).exists():
                self._add_sops_to_argocd_repo_server(config_file)
    
    def _add_sops_to_argocd_repo_server(self, file_path: str):
        """Add SOPS configuration to ArgoCD repo server preserving config"""
        yaml_handler = YAML()
        yaml_handler.preserve_quotes = True
        yaml_handler.map_indent = 2
        yaml_handler.sequence_indent = 2
        yaml_handler.sequence_dash_offset = 0
        yaml_handler.width = 4096
        yaml_handler.default_flow_style = False
        yaml_handler.allow_unicode = True
        
        with open(file_path, 'r') as f:
            data = yaml_handler.load(f) or {}
        
        # Ensure argo-cd structure exists
        if 'argo-cd' not in data:
            data['argo-cd'] = {}
        
        argocd = data['argo-cd']
        
        # Ensure repoServer structure exists
        if 'repoServer' not in argocd:
            argocd['repoServer'] = {}
        
        repo_server = argocd['repoServer']
        
        # Add SOPS environment variable (preserve existing env vars)
        if 'env' not in repo_server:
            repo_server['env'] = []
        
        # Check if SOPS_AGE_KEY_FILE already exists
        sops_env_exists = any(
            isinstance(env, dict) and env.get('name') == 'SOPS_AGE_KEY_FILE'
            for env in repo_server['env']
        )
        
        if not sops_env_exists:
            repo_server['env'].append({
                'name': 'SOPS_AGE_KEY_FILE',
                'value': '/sops-age-keys/keys.txt'
            })
        
        # Check if ARGOCD_EXEC_TIMEOUT already exists, add default if not
        timeout_env_exists = any(
            isinstance(env, dict) and env.get('name') == 'ARGOCD_EXEC_TIMEOUT'
            for env in repo_server['env']
        )
        
        if not timeout_env_exists:
            repo_server['env'].append({
                'name': 'ARGOCD_EXEC_TIMEOUT',
                'value': '300'
            })
        
        # Add volumes (preserve existing volumes)
        if 'volumes' not in repo_server:
            repo_server['volumes'] = []
        
        # Check if custom-tools volume exists
        custom_tools_exists = any(
            isinstance(vol, dict) and vol.get('name') == 'custom-tools'
            for vol in repo_server['volumes']
        )
        
        if not custom_tools_exists:
            repo_server['volumes'].append({
                'name': 'custom-tools',
                'emptyDir': {}
            })
        
        # Check if sops-age-keys volume exists
        sops_volume_exists = any(
            isinstance(vol, dict) and vol.get('name') == 'sops-age-keys'
            for vol in repo_server['volumes']
        )
        
        if not sops_volume_exists:
            repo_server['volumes'].append({
                'name': 'sops-age-keys',
                'secret': {
                    'secretName': 'sops-age-key'
                }
            })
        
        # Add volume mounts (preserve existing mounts)
        if 'volumeMounts' not in repo_server:
            repo_server['volumeMounts'] = []
        
        # Check if SOPS mount exists
        sops_mount_exists = any(
            (isinstance(mount, dict) and
             mount.get('mountPath') == '/usr/local/bin/sops')
            for mount in repo_server['volumeMounts']
        )
        
        if not sops_mount_exists:
            repo_server['volumeMounts'].append({
                'mountPath': '/usr/local/bin/sops',
                'name': 'custom-tools',
                'subPath': 'sops'
            })
        
        # Check if age keys mount exists
        age_keys_mount_exists = any(
            (isinstance(mount, dict) and
             mount.get('mountPath') == '/sops-age-keys')
            for mount in repo_server['volumeMounts']
        )
        
        if not age_keys_mount_exists:
            repo_server['volumeMounts'].append({
                'name': 'sops-age-keys',
                'mountPath': '/sops-age-keys',
                'readOnly': True
            })
        
        # Add init containers (preserve existing init containers)
        if 'initContainers' not in repo_server:
            repo_server['initContainers'] = []
        
        # Check if install-sops init container exists
        sops_init_exists = any(
            (isinstance(container, dict) and
             container.get('name') == 'install-sops')
            for container in repo_server['initContainers']
        )
        
        if not sops_init_exists:
            repo_server['initContainers'].append({
                'name': 'install-sops',
                'image': 'alpine:latest',
                'command': ['sh', '-c'],
                'args': [
                    'apk add --no-cache curl && '
                    'curl -Lo /custom-tools/sops https://github.com/mozilla/sops/releases/download/v3.8.1/sops-v3.8.1.linux.amd64 && '
                    'chmod +x /custom-tools/sops'
                ],
                'volumeMounts': [
                    {
                        'mountPath': '/custom-tools',
                        'name': 'custom-tools'
                    }
                ]
            })
        
        # Save the updated configuration
        with open(file_path, 'w') as f:
            yaml_handler.dump(data, f)
        
        console.print(f"✅ Added SOPS configuration to {file_path}", style="green")
    
    def _create_argocd_sops_config(self):
        """Create SOPS configuration in ArgoCD namespace"""
        sops_config = Path("bootstrap/argocd/.sops.yaml")
        
        if self.age_key_file.exists():
            with open(self.age_key_file, 'r') as f:
                content = f.read()
                for line in content.split('\n'):
                    if line.startswith('# public key:'):
                        public_key = line.split(': ')[1]
                        
                        config_content = {
                            "creation_rules": [
                                {
                                    "path_regex": r".*\.yaml$",
                                    "age": public_key
                                }
                            ]
                        }
                        
                        with open(sops_config, 'w') as f:
                            yaml.dump(config_content, f, default_flow_style=False)
                        
                        console.print("✅ Created ArgoCD SOPS configuration", style="green")
                        break
    
    def _initialize_ip_pools(self):
        """Initialize MetalLB IP pools from configuration"""
        # Get IP pools from config
        network_config = self.config.get("networking", {})
        metallb_config = network_config.get("metallb", {})
        
        self.metallb_ip_pools = metallb_config.get("pools", {})
        
        # If no pools configured, prompt for them
        if not self.metallb_ip_pools:
            console.print("⚠️ No MetalLB IP pools configured. Please configure them.", 
                         style="yellow")
            self._configure_metallb_pools()
        
        # Validate that each pool has at least 16 IPs
        for pool_name, ip_range in self.metallb_ip_pools.items():
            if not self._validate_ip_pool_size(ip_range):
                console.print(f"❌ IP pool '{pool_name}' ({ip_range}) has fewer than 16 IPs", 
                            style="red")
                raise ValueError(f"IP pool '{pool_name}' must have at least 16 IPs")
        
        # Scan existing configurations to find already assigned IPs
        self._scan_assigned_ips()
        
        console.print(f"✅ Initialized {len(self.metallb_ip_pools)} IP pools", 
                     style="green")
        for pool_name, ip_range in self.metallb_ip_pools.items():
            available_count = len(self._get_available_ips_in_range(ip_range))
            console.print(f"  - {pool_name}: {ip_range} ({available_count} available)", 
                         style="blue")
    
    def _configure_metallb_pools(self):
        """Configure MetalLB IP pools interactively"""
        console.print("\n📡 MetalLB Load Balancer IP Pool Configuration", style="bold")
        
        pools = {}
        
        # Internal pool
        console.print("Internal pool is for services inside the cluster network")
        while True:
            internal_pool = Prompt.ask(
                "Internal IP pool (format: start-end or CIDR)",
                default="10.0.3.220-10.0.3.235"
            )
            if self._validate_ip_pool_size(internal_pool):
                pools["internal"] = internal_pool
                break
            console.print("❌ Pool must have at least 16 IPs", style="red")
        
        # External pool
        console.print("External pool is for services needing external access")
        while True:
            external_pool = Prompt.ask(
                "External IP pool (format: start-end or CIDR)",
                default="10.148.121.20-10.148.121.35"
            )
            if self._validate_ip_pool_size(external_pool):
                pools["external"] = external_pool
                break
            console.print("❌ Pool must have at least 16 IPs", style="red")
        
        # Update configuration
        if "networking" not in self.config:
            self.config["networking"] = {}
        if "metallb" not in self.config["networking"]:
            self.config["networking"]["metallb"] = {}
        
        self.config["networking"]["metallb"]["pools"] = pools
        self.metallb_ip_pools = pools
        
        console.print("✅ Configured IP pools:", style="green")
        for name, pool in pools.items():
            console.print(f"  - {name}: {pool}", style="blue")
    
    def _validate_ip_pool_size(self, ip_range: str) -> bool:
        """Validate that an IP range has at least 16 IPs"""
        try:
            if '-' in ip_range:
                start_ip, end_ip = ip_range.split('-')
                start = ipaddress.IPv4Address(start_ip.strip())
                end = ipaddress.IPv4Address(end_ip.strip())
                return int(end) - int(start) + 1 >= 16
            else:
                # Single IP or CIDR notation
                network = ipaddress.IPv4Network(ip_range, strict=False)
                return network.num_addresses >= 16
        except (ipaddress.AddressValueError, ValueError):
            return False
    
    def _get_available_ips_in_range(self, ip_range: str) -> list:
        """Get list of available IPs in a range"""
        try:
            if '-' in ip_range:
                start_ip, end_ip = ip_range.split('-')
                start = ipaddress.IPv4Address(start_ip.strip())
                end = ipaddress.IPv4Address(end_ip.strip())
                all_ips = [str(ipaddress.IPv4Address(i)) 
                          for i in range(int(start), int(end) + 1)]
            else:
                # CIDR notation
                network = ipaddress.IPv4Network(ip_range, strict=False)
                all_ips = [str(ip) for ip in network.hosts()]
            
            # Return only unassigned IPs
            return [ip for ip in all_ips if ip not in self.assigned_ips]
        except (ipaddress.AddressValueError, ValueError):
            console.print(f"❌ Invalid IP range: {ip_range}", style="red")
            return []
    
    def _get_all_pool_ips(self) -> set:
        """Get all IPs from all configured pools"""
        all_ips = set()
        for pool_name, ip_range in self.metallb_ip_pools.items():
            try:
                if '-' in ip_range:
                    start_ip, end_ip = ip_range.split('-')
                    start = ipaddress.IPv4Address(start_ip.strip())
                    end = ipaddress.IPv4Address(end_ip.strip())
                    pool_ips = [str(ipaddress.IPv4Address(i)) 
                              for i in range(int(start), int(end) + 1)]
                else:
                    # CIDR notation
                    network = ipaddress.IPv4Network(ip_range, strict=False)
                    pool_ips = [str(ip) for ip in network.hosts()]
                all_ips.update(pool_ips)
            except (ipaddress.AddressValueError, ValueError):
                continue
        return all_ips
    
    def _scan_assigned_ips(self):
        """Scan existing configuration files for assigned MetalLB IPs"""
        self.assigned_ips.clear()
        
        # Search for loadBalancerIPs in YAML files, excluding docs and site
        for yaml_file in Path(".").rglob("*.yaml"):
            if (yaml_file.is_file() and 
                not str(yaml_file).startswith('.') and
                'docs/' not in str(yaml_file) and
                'site/' not in str(yaml_file)):
                try:
                    with open(yaml_file, 'r') as f:
                        content = f.read()
                        
                        # Look for metallb.universe.tf/loadBalancerIPs annotations
                        ip_matches = re.findall(
                            r'metallb\.universe\.tf/loadBalancerIPs:\s*([0-9.]+)',
                            content)
                        for ip in ip_matches:
                            self.assigned_ips.add(ip.strip())
                        
                        # Look for direct loadBalancerIP field
                        lb_ip_matches = re.findall(
                            r'loadBalancerIP:\s*([0-9.]+)',
                            content)
                        for ip in lb_ip_matches:
                            self.assigned_ips.add(ip.strip())
                            
                        # Look for service_annotations in AWX manifests
                        service_annotation_matches = re.findall(
                            r'service_annotations:.*?metallb\.universe\.tf/'
                            r'loadBalancerIPs:\s*([0-9.]+)',
                            content, re.DOTALL)
                        for ip in service_annotation_matches:
                            self.assigned_ips.add(ip.strip())
                            
                except (UnicodeDecodeError, PermissionError):
                    # Skip binary or protected files
                    continue
        
        if self.assigned_ips:
            console.print(f"📋 Found {len(self.assigned_ips)} already "
                         f"assigned IPs:", style="blue")
            for ip in sorted(self.assigned_ips):
                console.print(f"  - {ip}", style="dim")
    
    def _assign_load_balancer_ip(self, service_name: str, pool_name: str = "internal") -> str:
        """Assign a unique load balancer IP from the specified pool"""
        if pool_name not in self.metallb_ip_pools:
            console.print(f"❌ Unknown IP pool: {pool_name}", style="red")
            return None
        
        available_ips = self._get_available_ips_in_range(self.metallb_ip_pools[pool_name])
        
        if not available_ips:
            console.print(f"❌ No available IPs in pool '{pool_name}'", style="red")
            return None
        
        # Use the first available IP
        assigned_ip = available_ips[0]
        self.assigned_ips.add(assigned_ip)
        
        console.print(f"📍 Assigned IP {assigned_ip} to {service_name} from pool '{pool_name}'", 
                     style="green")
        return assigned_ip
    
    def _extract_service_name_from_file(self, file_path: Path) -> str:
        """Extract service name from file path"""
        # Get service name from file path - usually the parent directory
        parts = file_path.parts
        if 'system' in parts:
            # system/ingress-nginx/values.yaml -> ingress-nginx
            idx = parts.index('system')
            if idx + 1 < len(parts):
                return parts[idx + 1]
        elif 'apps' in parts:
            # apps/hajimari/values.yaml -> hajimari
            idx = parts.index('apps')
            if idx + 1 < len(parts):
                return parts[idx + 1]
        elif 'platform' in parts:
            # platform/awx/values.yaml -> awx
            idx = parts.index('platform')
            if idx + 1 < len(parts):
                return parts[idx + 1]
        
        # Fallback to parent directory name
        return file_path.parent.name
    
    def _determine_ip_pool(self, service_name: str, file_path: Path) -> str:
        """Determine which IP pool to use for a service"""
        # Define service-to-pool mapping
        external_services = {
            'ingress-nginx', 'nginx-ingress', 'traefik', 
            'haproxy', 'envoy', 'gateway'
        }
        
        # Services that typically need external access
        if service_name.lower() in external_services:
            return "external"
        
        # Check if file path suggests external service
        if 'system' in file_path.parts and any(
            ext in service_name.lower() 
            for ext in ['ingress', 'gateway', 'proxy']
        ):
            return "external"
        
        # Default to internal pool
        return "internal"
    
    def _file_needs_metallb_ip(self, file_path: Path) -> bool:
        """Check if a values file should have MetalLB IP assignment"""
        # Services that commonly need LoadBalancer IPs
        metallb_services = {
            'ingress-nginx', 'nginx-ingress', 'traefik', 
            'awx', 'tower', 'jenkins', 'gitlab', 'harbor'
        }
        
        service_name = self._extract_service_name_from_file(file_path)
        if service_name.lower() in metallb_services:
            return True
        
        # Also check if the file already contains service.type: LoadBalancer
        try:
            with open(file_path, 'r') as f:
                content = f.read()
                if re.search(r'type:\s*LoadBalancer', content):
                    return True
        except (UnicodeDecodeError, PermissionError):
            pass
        
        return False
    
    def _add_metallb_annotations(self):
        """Add MetalLB annotations to services that need them"""
        console.print("\n🔧 Adding MetalLB annotations where needed...",
                     style="bold")
        
        # Find all values files that might need MetalLB annotations
        for values_file in Path(".").rglob("*values*.yaml"):
            if (values_file.is_file() and 
                'docs/' not in str(values_file) and
                'site/' not in str(values_file)):
                
                service_name = self._extract_service_name_from_file(values_file)
                
                # Check if this service should have MetalLB IP
                if self._file_needs_metallb_ip(values_file):
                    console.print(f"📝 Checking {values_file} for MetalLB "
                                 f"annotation...", style="blue")
                    
                    # Read file to check if annotation exists
                    with open(values_file, 'r') as f:
                        content = f.read()
                    
                    # Check if metallb annotation already exists
                    if "metallb.universe.tf/loadBalancerIPs:" in content:
                        console.print(f"✅ MetalLB annotation already exists "
                                     f"in {values_file}", style="green")
                        continue
                    
                    # Check if service.annotations section exists
                    if ("service:" in content and "annotations:" in content and
                        re.search(r'service:\s*\n.*?annotations:', content,
                                 re.DOTALL)):
                        # Add to existing service annotations
                        self._add_to_existing_service_annotations(values_file)
                    elif "service:" in content:
                        # Add annotations section to service
                        self._add_service_annotations_section(values_file)
                    else:
                        console.print(f"⚠️ No service section found in "
                                     f"{values_file}", style="yellow")
        
        # Also handle special service files and AWX manifests
        self._update_special_service_files()
    
    def _update_special_service_files(self):
        """Update special service files like AWX manifests and standalone 
        service files"""
        # Handle AWX platform manifests
        awx_files = list(Path(".").rglob("**/awx*.yaml"))
        for awx_file in awx_files:
            if (awx_file.is_file() and 
                'docs/' not in str(awx_file) and
                'site/' not in str(awx_file)):
                self._update_awx_service_annotations(awx_file)
        
        # Handle standalone service files
        service_files = list(Path(".").rglob("**/service.yaml"))
        for service_file in service_files:
            if (service_file.is_file() and 
                'docs/' not in str(service_file) and
                'site/' not in str(service_file)):
                self._update_standalone_service_file(service_file)
    
    def _update_awx_service_annotations(self, file_path: Path):
        """Update AWX service_annotations with MetalLB IP"""
        with open(file_path, 'r') as f:
            content = f.read()
        
        # Check if this is an AWX manifest
        if "kind: AWX" not in content:
            return
        
        service_name = "awx"
        pool = self._determine_ip_pool(service_name, file_path)
        
        # Check if already has MetalLB annotation and extract existing IP
        existing_ip_match = re.search(
            r'metallb\.universe\.tf/loadBalancerIPs:\s*([0-9.]+)', content)
        if existing_ip_match:
            existing_ip = existing_ip_match.group(1)
            # Check if existing IP is in our configured pools
            if existing_ip in self._get_all_pool_ips():
                console.print(f"✅ AWX already has valid IP {existing_ip} "
                             f"from pool", style="green")
                self.assigned_ips.add(existing_ip)
                return
            else:
                console.print(f"⚠️ AWX has IP {existing_ip} outside "
                             f"configured pools, updating...", style="yellow")
        
        assigned_ip = self._assign_load_balancer_ip(service_name, pool)
        if not assigned_ip:
            return
        
        # Update service_annotations field
        lines = content.split('\n')
        updated_lines = []
        found_service_annotations = False
        
        for line in lines:
            if line.strip().startswith('service_annotations:'):
                found_service_annotations = True
                if existing_ip_match:
                    # Replace existing IP in multi-line annotations
                    updated_lines.append(line)
                else:
                    # Add MetalLB annotation to service_annotations
                    updated_lines.append(line)
                    if '|' in line:
                        # Multi-line service_annotations
                        updated_lines.append(f"    metallb.universe.tf/"
                                            f"loadBalancerIPs: {assigned_ip}")
                    else:
                        # Single line - convert to multi-line
                        indent_match = re.match(r'^(\s*)', line)
                        indent = indent_match.group(1) if indent_match else ""
                        updated_lines[-1] = f"{indent}service_annotations: |"
                        updated_lines.append(f"{indent}  metallb.universe.tf/"
                                            f"loadBalancerIPs: {assigned_ip}")
            elif (found_service_annotations and existing_ip_match and
                  "metallb.universe.tf/loadBalancerIPs:" in line):
                # Replace the IP in the existing line
                indent_match = re.match(r'^(\s*)', line)
                indent = indent_match.group(1) if indent_match else ""
                updated_lines.append(f"{indent}metallb.universe.tf/"
                                    f"loadBalancerIPs: {assigned_ip}")
            else:
                updated_lines.append(line)
        
        # Write back the file
        with open(file_path, 'w') as f:
            f.write('\n'.join(updated_lines))
        
        if existing_ip_match:
            console.print(f"✅ Updated MetalLB IP from {existing_ip} to "
                         f"{assigned_ip} in AWX manifest {file_path}",
                         style="green")
        else:
            console.print(f"✅ Added MetalLB IP {assigned_ip} to AWX manifest "
                         f"{file_path}", style="green")
    
    def _update_standalone_service_file(self, file_path: Path):
        """Update standalone service.yaml files with loadBalancerIP"""
        with open(file_path, 'r') as f:
            content = f.read()
        
        # Check if this is a LoadBalancer service
        if "type: LoadBalancer" not in content:
            return
        
        # Check if already has loadBalancerIP or MetalLB annotation
        if ("loadBalancerIP:" in content or 
            "metallb.universe.tf/loadBalancerIPs:" in content):
            return
        
        service_name = self._extract_service_name_from_file(file_path)
        pool = self._determine_ip_pool(service_name, file_path)
        assigned_ip = self._assign_load_balancer_ip(service_name, pool)
        
        if not assigned_ip:
            return
        
        # Add loadBalancerIP to spec section
        lines = content.split('\n')
        updated_lines = []
        in_spec = False
        
        for line in lines:
            if line.strip() == "spec:":
                in_spec = True
                updated_lines.append(line)
                # Add loadBalancerIP after spec:
                indent_match = re.match(r'^(\s*)', line)
                indent = indent_match.group(1) if indent_match else ""
                updated_lines.append(f"{indent}  loadBalancerIP: {assigned_ip}")
            else:
                updated_lines.append(line)
        
        # Write back the file
        with open(file_path, 'w') as f:
            f.write('\n'.join(updated_lines))
        
        console.print(f"✅ Added loadBalancerIP {assigned_ip} to service "
                     f"{file_path}", style="green")
    
    def _add_to_existing_service_annotations(self, file_path: Path):
        """Add MetalLB annotation to existing service annotations"""
        service_name = self._extract_service_name_from_file(file_path)
        pool = self._determine_ip_pool(service_name, file_path)
        assigned_ip = self._assign_load_balancer_ip(service_name, pool)
        
        if not assigned_ip:
            return
        
        temp_lines = []
        annotation_added = False
        
        with open(file_path, 'r') as f:
            lines = f.readlines()
        
        i = 0
        while i < len(lines):
            line = lines[i]
            temp_lines.append(line)
            
            # Look for service: followed by annotations:
            if (re.match(r'^\s*service:\s*$', line) and 
                i + 1 < len(lines)):
                
                # Look ahead for annotations section within the service block
                j = i + 1
                service_indent = len(line) - len(line.lstrip())
                
                while j < len(lines):
                    next_line = lines[j]
                    next_indent = len(next_line) - len(next_line.lstrip())
                    
                    # If we find a line at same or less indentation, we've left the service block
                    if (next_line.strip() and 
                        next_indent <= service_indent and 
                        not next_line.strip().startswith('#')):
                        break
                    
                    # Found annotations within service block
                    if re.match(r'^\s*annotations:\s*$', next_line):
                        # Add all lines up to annotations
                        for k in range(i + 1, j + 1):
                            temp_lines.append(lines[k])
                        
                        # Add MetalLB annotation
                        annotation_indent = " " * (next_indent + 2)
                        temp_lines.append(f"{annotation_indent}metallb.universe.tf/loadBalancerIPs: {assigned_ip}\n")
                        annotation_added = True
                        i = j
                        break
                    
                    j += 1
            
            i += 1
        
        # Write back the file if annotation was added
        if annotation_added:
            with open(file_path, 'w') as f:
                f.writelines(temp_lines)
            console.print(f"✅ Added MetalLB IP {assigned_ip} to {file_path}", style="green")
    
    def _add_service_annotations_section(self, file_path: Path):
        """Add annotations section to existing service"""
        service_name = self._extract_service_name_from_file(file_path)
        pool = self._determine_ip_pool(service_name, file_path)
        assigned_ip = self._assign_load_balancer_ip(service_name, pool)
        
        if not assigned_ip:
            return
        
        temp_lines = []
        annotation_added = False
        
        with open(file_path, 'r') as f:
            lines = f.readlines()
        
        for i, line in enumerate(lines):
            temp_lines.append(line)
            
            # Add annotations after service: line
            if re.match(r'^\s*service:\s*$', line):
                # Get indentation from service line
                indent_match = re.match(r'^(\s*)', line)
                indent = indent_match.group(1) if indent_match else ""
                
                # Add annotations section
                temp_lines.append(f"{indent}  annotations:\n")
                temp_lines.append(f"{indent}    metallb.universe.tf/loadBalancerIPs: {assigned_ip}\n")
                annotation_added = True
        
        # Write back the file if annotation was added
        if annotation_added:
            with open(file_path, 'w') as f:
                f.writelines(temp_lines)
            console.print(f"✅ Added MetalLB IP {assigned_ip} to {file_path}", style="green")
    
    def update_configurations_in_place(self):
        """Update all configuration files in-place with current settings"""
        console.print("\n📝 Updating configuration files in-place...", style="bold")
        
        # Initialize IP pools for MetalLB
        self._initialize_ip_pools()
        
        # Add MetalLB annotations where needed
        self._add_metallb_annotations()
        
        # Update Ansible variables
        self._update_ansible_vars()
        
        # Update ArgoCD configurations
        self._update_argocd_configs()
        
        # Update Kubernetes manifests
        self._update_k8s_manifests()
        
        # Update cert-manager annotations
        cert_provider = self.config.get("services", {}).get("cert_provider", "pebble-issuer")
        self._update_cert_manager_annotations(cert_provider)
        
        # Update other configuration files
        self._update_other_configs()
    
    def _update_ansible_vars(self):
        """Update Ansible variable files using regex to preserve formatting"""
        cluster_cfg = self.config.get("cluster", {})
        network_cfg = self.config.get("networking", {})
        k8s_cfg = self.config.get("kubernetes", {})
        
        ansible_updates = {
            "ubiquity_user": self.secrets.get("ubiquity_reg_user", ""),
            "ubiquity_pass": self.secrets.get("ubiquity_reg_token", ""),
            "cluster_name": cluster_cfg.get("name", "ubiquity"),
            "cluster_domain": cluster_cfg.get("domain", "ubiquitycluster.local"),
            "timezone": cluster_cfg.get("timezone", "Europe/London"),
            "internal_ipv4_interface": network_cfg.get("internal_ipv4_interface", 
                                                       "ens4f0"),
            "internal_ipv4_address": network_cfg.get("internal_ipv4_address", 
                                                     "10.0.3.253"),
            "internal_ipv4_network": network_cfg.get("internal_ipv4_network", 
                                                     "10.0.0.0/22"),
            "keepalived_ipv4_vip": network_cfg.get("keepalived_ipv4_vip", 
                                                   "10.0.3.250"),
            "k3s_version": k8s_cfg.get("version", "v1.23.17+k3s1"),
            "cluster_cidr": k8s_cfg.get("cluster_cidr", "10.46.0.0/22"),
            "service_cidr": k8s_cfg.get("service_cidr", "10.48.0.0/22")
        }
        
        ansible_files = [
            "metal/group_vars/all.yml",
            "metal/group_vars/metal.yml"
        ]
        
        for file_path in ansible_files:
            if Path(file_path).exists():
                self._update_file_in_place(file_path, ansible_updates)
    
    def _update_argocd_configs(self):
        """Update ArgoCD configuration files"""
        argocd_updates = {}
        
        # Add repository configuration if credentials exist
        if self.secrets.get("seed_repo_username") and self.secrets.get("seed_repo_password"):
            argocd_updates["argo-cd.configs.repositories"] = [
                {
                    "url": "https://github.com/ubiquitycluster/ubiquity",
                    "name": "ubiquity-seed",
                    "username": self.secrets["seed_repo_username"],
                    "password": self.secrets["seed_repo_password"]
                }
            ]
        
        # Note: Domain updates are now handled by _update_k8s_manifests()
        # Note: cert-manager annotations are handled by 
        # _update_cert_manager_annotations()
        
        argocd_files = [
            "bootstrap/argocd/values.yaml",
            "bootstrap/argocd/values-seed.yaml"
        ]
        
        for file_path in argocd_files:
            if Path(file_path).exists():
                self._update_yaml_file_in_place(file_path, argocd_updates)
    
    def _update_cert_manager_annotations(self, cert_provider):
        """Update all cert-manager.io/cluster-issuer annotations to the specified provider"""
        console.print(f"🔐 Updating cert-manager annotations to: {cert_provider}", style="cyan")
        
        # Find all YAML files (excluding only docs and site directories)
        yaml_files = []
        for yaml_file in Path(".").rglob("*.yaml"):
            if (yaml_file.is_file() and 
                'docs/' not in str(yaml_file) and
                'site/' not in str(yaml_file)):
                yaml_files.append(yaml_file)
        
        console.print(f"📋 Found {len(yaml_files)} YAML files to check", style="blue")
        
        updated_files = []
        for yaml_file in yaml_files:
            try:
                # Read file content first
                with open(yaml_file, 'r', encoding='utf-8') as f:
                    content = f.read()
                
                # Check if file contains cert-manager annotations
                if 'cert-manager.io/cluster-issuer:' not in content:
                    continue
                
                console.print(f"📝 Processing {yaml_file}...", style="blue")
                
                # Process file line by line
                lines = content.splitlines(keepends=True)
                updated_lines = []
                updated = False
                
                for line in lines:
                    # Match only complete cert-manager.io/cluster-issuer: lines
                    if (re.match(r'^\s*cert-manager\.io/cluster-issuer:\s*\S+', line) and
                        not line.strip().startswith('#')):
                        # Extract indentation and current value
                        indent_match = re.match(r'^(\s*)', line)
                        indent = indent_match.group(1) if indent_match else ""
                        
                        current_value_match = re.search(
                            r'cert-manager\.io/cluster-issuer:\s*(.+?)$', 
                            line.strip())
                        current_value = (current_value_match.group(1).strip() 
                                       if current_value_match else "")
                        
                        if current_value != cert_provider:
                            console.print(f"  🔄 Updating from '{current_value}' to '{cert_provider}'", style="yellow")
                            updated_lines.append(f"{indent}cert-manager.io/cluster-issuer: {cert_provider}\n")
                            updated = True
                        else:
                            updated_lines.append(line)
                    else:
                        updated_lines.append(line)
                
                # Write back to file only if changes were made
                if updated:
                    # Create backup first
                    backup_path = yaml_file.with_suffix('.yaml.bak')
                    with open(backup_path, 'w', encoding='utf-8') as backup_f:
                        backup_f.write(content)
                    
                    # Write updated content
                    with open(yaml_file, 'w', encoding='utf-8') as f:
                        f.writelines(updated_lines)
                    
                    # Remove backup if write was successful
                    backup_path.unlink()
                    
                    updated_files.append(yaml_file)
                    console.print(f"  ✅ Updated {yaml_file}", style="green")
                
            except PermissionError as e:
                console.print(f"  ❌ Permission denied accessing {yaml_file}: {e}", style="red")
            except UnicodeDecodeError as e:
                console.print(f"  ⚠️ Encoding error in {yaml_file}: {e}", style="yellow")
            except OSError as e:
                console.print(f"  ❌ I/O error processing {yaml_file}: {e}", style="red")
            except Exception as e:
                console.print(f"  ⚠️ Unexpected error processing {yaml_file}: {e}", style="red")
        
        if updated_files:
            console.print(f"🔐 Updated cert-manager annotations in {len(updated_files)} files", style="green")
        else:
            console.print("🔐 No cert-manager annotations needed updating", style="blue")
    
    def _update_k8s_manifests(self):
        """Update Kubernetes manifest files"""
        # Update ingress hostnames based on cluster domain
        cluster_domain = self.config.get("cluster", {}).get("domain", "ubiquitycluster.local")
        cert_provider = self.config.get("services", {}).get(
            "cert_provider", "pebble-issuer")
        
        console.print(f"📋 Using cert provider: {cert_provider}", style="cyan")
        console.print(f"📋 Target cluster domain: {cluster_domain}", style="cyan")
        
        # Get the old domain patterns to replace (exclude current target domain)
        all_possible_domains = [
            "ubiquitycluster.local",     # Development default
            "ubiquitycluster.uk",        # Production default
            ".local",                    # Any .local domain
            ".nip.io"                   # Dynamic DNS domains
        ]
        
        # Only include domains that are different from the target
        old_domain_patterns = []
        for domain in all_possible_domains:
            if domain != cluster_domain and not cluster_domain.endswith(domain):
                old_domain_patterns.append(domain)
        
        console.print(f"📋 Will replace domains: {old_domain_patterns}",
                      style="blue")
        
        # Update ingress files
        self._update_ingress_files(old_domain_patterns, cluster_domain)
        
        # Update values and defaults files
        self._update_yaml_config_files(old_domain_patterns, cluster_domain)
    
    def _update_ingress_files(self, old_domain_patterns, cluster_domain):
        """Update ingress files with new domain"""
        for ingress_file in Path(".").rglob("*ingress*.yaml"):
            if (ingress_file.is_file() and
                    'docs/' not in str(ingress_file) and
                    'site/' not in str(ingress_file)):
                console.print(f"📝 Updating ingress {ingress_file}...",
                              style="blue")
                
                self._update_yaml_file_domains(
                    ingress_file, old_domain_patterns, cluster_domain)
    
    def _update_yaml_config_files(self, old_domain_patterns, cluster_domain):
        """Update values.yaml, cr.yaml and defaults.yaml files with new domain"""
        # Process both *values*.yaml, *cr.yaml and *defaults*.yaml files
        patterns = ["*values*.yaml", "*defaults*.yaml", "*cr.yaml"]
        
        for pattern in patterns:
            for config_file in Path(".").rglob(pattern):
                if (config_file.is_file() and
                        'docs/' not in str(config_file) and
                        'site/' not in str(config_file)):
                    console.print(f"📝 Updating config {config_file}...",
                                  style="blue")
                    
                    self._update_yaml_file_domains(
                        config_file, old_domain_patterns, cluster_domain)
    
    def _update_yaml_file_domains(self, yaml_file, old_domain_patterns,
                                  cluster_domain):
        """Update a YAML file in place with new domain patterns"""
        with fileinput.input(str(yaml_file), inplace=True) as file:
            for line in file:
                # Skip comments and empty lines - never modify them
                stripped_line = line.strip()
                if (stripped_line.startswith('#') or
                        stripped_line.startswith('##') or
                        not stripped_line):
                    print(line, end='')
                    continue
                
                # Update hostnames preserving structure
                if ("host:" in line and
                        not line.strip().startswith('#')):
                    updated_line = self._update_host_line(
                        line, old_domain_patterns, cluster_domain)
                    print(updated_line, end='')
                # Update hostname list entries (- hostname or URL)
                elif (line.strip().startswith('-') and
                      not line.strip().startswith('# -') and
                      'host:' not in line and
                      not line.strip().startswith('- /') and  # Exclude paths
                      any(pattern in line for pattern in old_domain_patterns)):
                    updated_line = self._update_list_item_line(
                        line, old_domain_patterns, cluster_domain)
                    print(updated_line, end='')
                # Update URL fields (domain, root_url, auth_url, etc.)
                elif self._is_url_field_line(line, old_domain_patterns):
                    updated_line = self._update_url_field_line(
                        line, old_domain_patterns, cluster_domain)
                    print(updated_line, end='')
                # Handle MetalLB LoadBalancer IP assignment
                elif ("metallb.universe.tf/loadBalancerIPs:" in line and
                      not line.strip().startswith('#')):
                    updated_line = self._update_metallb_line(line, yaml_file)
                    print(updated_line, end='')
                # Handle direct loadBalancerIP field
                elif ("loadBalancerIP:" in line and
                      not line.strip().startswith('#')):
                    updated_line = self._update_load_balancer_ip_line(
                        line, yaml_file)
                    print(updated_line, end='')
                else:
                    print(line, end='')
    
    def _update_host_line(self, line, old_domain_patterns, cluster_domain):
        """Update a host: line with new domain"""
        # Check if line contains any old domain pattern
        should_update = any(pattern in line for pattern in old_domain_patterns)
        
        if should_update:
            # Extract indentation and service name
            indent_match = re.match(r'^(\s*)', line)
            indent = indent_match.group(1) if indent_match else ""
            
            # Check if this is a list item (has dash)
            is_list_item = '- host:' in line
            
            # Extract service name (before first dot)
            host_part = line.split('host:')[1].strip()
            service_name = host_part.split('.')[0].strip()
            
            if is_list_item:
                return f"{indent}- host: {service_name}.{cluster_domain}\n"
            else:
                return f"{indent}host: {service_name}.{cluster_domain}\n"
        else:
            return line
    
    def _update_list_item_line(self, line, old_domain_patterns,
                               cluster_domain):
        """Update a list item line (- hostname) with new domain"""
        indent_match = re.match(r'^(\s*)', line)
        indent = indent_match.group(1) if indent_match else ""
        
        # Extract content after the dash
        content_part = line.split('-', 1)[1].strip()
        
        # Check if this is a URL (contains protocol)
        is_url = content_part.startswith(('http://', 'https://'))
        if is_url:
            # Parse URL to update only the hostname part
            import urllib.parse
            parsed = urllib.parse.urlparse(content_part)
            
            # Extract service name from hostname
            service_name = parsed.hostname.split('.')[0]
            
            # Rebuild URL with new domain
            new_hostname = f"{service_name}.{cluster_domain}"
            new_url = parsed._replace(netloc=new_hostname).geturl()
            
            return f"{indent}- {new_url}\n"
        else:
            # Simple hostname - extract service name
            service_name = content_part.split('.')[0].strip()
            return f"{indent}- {service_name}.{cluster_domain}\n"
    
    def _is_url_field_line(self, line, old_domain_patterns):
        """Check if line contains a URL field that should be updated"""
        url_fields = [
            'domain:', 'root_url:', 'auth_url:', 'token_url:', 'api_url:',
            'baseURL:', 'url:', 'redirectURI:', 'ROOT_URL:', 'hostname:',
            'externalURL:', 'KEYCLOAK_URL:', 'MINIO_URL:', 'ONYXIA_API_URL:',
            'VAULT_URL:', 'hajimari.io/url:', 'issuer:', 'servername:'
        ]
        
        return (any(url_field in line for url_field in url_fields) and
                not line.strip().startswith('#') and
                any(pattern in line for pattern in old_domain_patterns))
    
    def _update_url_field_line(self, line, old_domain_patterns,
                               cluster_domain):
        """Update a URL field line with new domain"""
        # Extract the field name and current value
        field_match = re.match(r'^(\s*)([^:]+):\s*(.*)$', line)
        if not field_match:
            return line
        
        indent = field_match.group(1)
        field_name = field_match.group(2)
        current_value = field_match.group(3)
        
        # Only update if it's actually a domain/URL, not a file path
        should_update = (not current_value.startswith('/') and
                         any(pattern in line
                             for pattern in old_domain_patterns))
        
        if not should_update:
            return line
        
        # Handle different field types
        if field_name.strip() == 'domain':
            # Simple domain field
            service_name = current_value.split('.')[0].strip()
            new_value = f"{service_name}.{cluster_domain}"
        else:
            # URL fields - preserve protocol and path
            if current_value.startswith(('http://', 'https://')):
                import urllib.parse
                parsed = urllib.parse.urlparse(current_value)
                service_name = parsed.hostname.split('.')[0]
                new_hostname = f"{service_name}.{cluster_domain}"
                new_value = parsed._replace(netloc=new_hostname).geturl()
            else:
                # Fallback for non-URL values
                service_name = current_value.split('.')[0].strip()
                new_value = f"{service_name}.{cluster_domain}"
        
        return f"{indent}{field_name}: {new_value}\n"
    
    def _update_metallb_line(self, line, yaml_file):
        """Update MetalLB LoadBalancer IP assignment"""
        # Extract the original indentation
        indent_match = re.match(r'^(\s*)', line)
        indent = indent_match.group(1) if indent_match else ""
        
        # Extract current IP if any
        current_ip_match = re.search(r':\s*([0-9.]+)', line)
        if current_ip_match:
            current_ip = current_ip_match.group(1)
            # Keep existing IP if it's already assigned
            if current_ip in self.assigned_ips:
                return line
            else:
                # IP exists but not tracked, add to assigned
                self.assigned_ips.add(current_ip)
                return line
        else:
            # No IP assigned, assign one from pool
            service_name = self._extract_service_name_from_file(yaml_file)
            pool = self._determine_ip_pool(service_name, yaml_file)
            assigned_ip = self._assign_load_balancer_ip(service_name, pool)
            if assigned_ip:
                return (f"{indent}metallb.universe.tf/"
                        f"loadBalancerIPs: {assigned_ip}\n")
            else:
                return line
    
    def _update_load_balancer_ip_line(self, line, yaml_file):
        """Update direct loadBalancerIP field"""
        # Extract the original indentation
        indent_match = re.match(r'^(\s*)', line)
        indent = indent_match.group(1) if indent_match else ""
        
        # Extract current IP if any
        current_ip_match = re.search(r':\s*([0-9.]+)', line)
        if current_ip_match:
            current_ip = current_ip_match.group(1)
            # Keep existing IP if it's already assigned
            if current_ip in self.assigned_ips:
                return line
            else:
                # IP exists but not tracked, add to assigned
                self.assigned_ips.add(current_ip)
                return line
        else:
            # No IP assigned, assign one from pool
            service_name = self._extract_service_name_from_file(yaml_file)
            pool = self._determine_ip_pool(service_name, yaml_file)
            assigned_ip = self._assign_load_balancer_ip(service_name, pool)
            if assigned_ip:
                return f"{indent}loadBalancerIP: {assigned_ip}\n"
            else:
                return line
    
    def _update_other_configs(self):
        """Update other configuration files"""
        # Update any .env files
        env_files = [".env", "metal/.env", "bootstrap/.env"]
        
        env_updates = {}
        
        # Add cluster configuration
        for key, value in self.config.get("cluster", {}).items():
            env_updates[f"CLUSTER_{key.upper()}"] = value
        
        # Add networking configuration
        for key, value in self.config.get("networking", {}).items():
            env_updates[f"NETWORK_{key.upper()}"] = value
        
        for env_file in env_files:
            if Path(env_file).exists():
                self._update_file_in_place(env_file, env_updates)
    
    def initialize(self) -> bool:
        """Initialize SOPS configuration"""
        if not self._check_sops_installed():
            return False
        
        # Setup age key
        public_key = self._setup_age_key()
        if not public_key:
            return False
        
        # Create/update .sops.yaml if needed
        if not self.sops_config_file.exists():
            sops_config = {
                "creation_rules": [
                    {
                        "path_regex": r"\.ubiquity/secrets\.yaml$",
                        "age": public_key
                    },
                    {
                        "path_regex": r"bootstrap/.*secrets.*\.yaml$",
                        "age": public_key
                    },
                    {
                        "path_regex": r".*secret.*\.ya?ml$",
                        "age": public_key
                    }
                ]
            }
            
            with open(self.sops_config_file, 'w') as f:
                yaml.dump(sops_config, f, default_flow_style=False)
            
            console.print("✅ Created .sops.yaml configuration", style="green")
        
        # Load existing configuration
        self._load_config()
        self.secrets = self._decrypt_secrets()
        
        # Initialize secrets with defaults if empty
        if not self.secrets:
            self.secrets = self._get_default_secrets()
        
        # Initialize IP pools
        self._initialize_ip_pools()
        
        console.print("✅ SOPS configuration initialized", style="green")
        return True
    
    def _load_config(self):
        """Load non-sensitive configuration"""
        if self.config_file.exists():
            with open(self.config_file, 'r') as f:
                self.config = yaml.safe_load(f) or {}
        else:
            self.config = self._get_default_config()
    
    def _get_default_config(self) -> dict:
        """Get default configuration template"""
        return {
            "cluster": {
                "name": "ubiquity",
                "domain": "ubiquitycluster.local",
                "timezone": "Europe/London"
            },
            "networking": {
                "internal_ipv4_interface": "ens4f0",
                "internal_ipv4_address": "10.0.3.253",
                "internal_ipv4_network": "10.0.0.0/22",
                "external_ipv4_interface": "ens4f0",
                "external_ipv4_address": "10.0.7.253",
                "external_ipv4_network": "10.0.4.0/22",
                "keepalived_ipv4_vip": "10.0.3.250",
                "metallb": {
                    "enabled": True,
                    "pools": {
                        "internal": "10.0.3.220-10.0.3.235",
                        "external": "10.148.121.20-10.148.121.35"
                    }
                }
            },
            "kubernetes": {
                "k3s_version": "v1.33.1+k3s1",
                "cluster_cidr": "10.46.0.0/22",
                "service_cidr": "10.48.0.0/22",
                "cilium_enabled": False,
                "ciliumlb_external_ip_range": "10.0.1.200-10.0.1.220",
                "ciliumlb_internal_ip_range": "10.0.1.200-10.0.1.220"
            },
            "services": {
                "cert_provider": "pebble-issuer",
                "dns_server": "8.8.8.8",
                "ntp_server": "8.8.8.8",
                "kubeless_enabled": False
            },
            "system": {
                "base_os": "Rocky",
                "base_os_version": "9.4",
                "use_ofed": False,
                "ofed_version": "23.10-3.2.2.0-LTS",
                "use_doca": False,
                "doca_version": "v3.0.0"
            }
        }
    
    def _get_default_secrets(self) -> dict:
        """Get default secrets template"""
        return {
            "seed_repo_username": "",
            "seed_repo_password": "",
            "seed_repo_ssh_key": "",
            "ubiquity_reg_user": "",
            "ubiquity_reg_token": "",
            "dockerhub_reg_user": "",
            "dockerhub_reg_token": "",
            "k3s_encryption_secret": "",
            "cloudcmd_password": "",
            "argocd_admin_password": "",
            "grafana_admin_password": "",
            "minio_root_password": ""
        }
    
    def configure_interactive(self):
        """Interactive configuration wizard"""
        console.print(Panel.fit("🚀 Ubiquity Secure Configuration with SOPS", style="bold blue"))
        
        # Configure non-sensitive settings
        self._configure_cluster()
        self._configure_networking() 
        self._configure_kubernetes()
        self._configure_services()
        self._configure_system()
        
        # Configure MetalLB IP pools if not already configured
        if ("networking" not in self.config or 
            "metallb" not in self.config["networking"] or 
            "pools" not in self.config["networking"]["metallb"] or 
            not self.config["networking"]["metallb"]["pools"]):
            self._configure_metallb_pools()
        
        # Configure secrets
        self._configure_secrets()
        
        # Save configuration
        self._save_config()
        self._encrypt_secrets(self.secrets)
        
        # Configure ArgoCD with SOPS
        self.configure_argocd_sops()
        
        # Update all files in-place
        self.update_configurations_in_place()
    
    def _configure_cluster(self):
        """Configure basic cluster settings"""
        console.print("\n📋 Cluster Configuration", style="bold")
        
        self.config["cluster"]["name"] = Prompt.ask(
            "Cluster name", 
            default=self.config["cluster"]["name"]
        )
        
        self.config["cluster"]["domain"] = Prompt.ask(
            "Domain name", 
            default=self.config["cluster"]["domain"]
        )
        
        self.config["cluster"]["timezone"] = Prompt.ask(
            "Timezone", 
            default=self.config["cluster"]["timezone"]
        )
    
    def _configure_networking(self):
        """Configure networking settings"""
        console.print("\n🌐 Network Configuration", style="bold")
        
        self.config["networking"]["internal_ipv4_interface"] = Prompt.ask(
            "Internal IPv4 interface",
            default=self.config["networking"]["internal_ipv4_interface"]
        )
        
        self.config["networking"]["internal_ipv4_address"] = Prompt.ask(
            "Internal IPv4 address",
            default=self.config["networking"]["internal_ipv4_address"]
        )
        
        self.config["networking"]["keepalived_ipv4_vip"] = Prompt.ask(
            "Keepalived VIP",
            default=self.config["networking"]["keepalived_ipv4_vip"]
        )
        
        # Always prompt for MetalLB IP pools
        console.print("\n📡 MetalLB Load Balancer IP Pools", style="bold")
        
        # Initialize metallb config if not exists
        if "metallb" not in self.config["networking"]:
            self.config["networking"]["metallb"] = {"pools": {}}
        
        metallb_config = self.config["networking"]["metallb"]
        
        # Always prompt for internal pool
        current_internal = metallb_config.get("pools", {}).get("internal", "10.0.3.220-10.0.3.235")
        console.print("Internal pool is for services inside the cluster network")
        while True:
            internal_pool = Prompt.ask(
                "Internal IP pool (format: start-end or CIDR)",
                default=current_internal
            )
            if self._validate_ip_pool_size(internal_pool):
                metallb_config.setdefault("pools", {})["internal"] = internal_pool
                break
            console.print("❌ Pool must have at least 16 IPs", style="red")
        
        # Always prompt for external pool
        current_external = metallb_config.get("pools", {}).get("external", "10.148.121.20-10.148.121.35")
        console.print("External pool is for services needing external access")
        while True:
            external_pool = Prompt.ask(
                "External IP pool (format: start-end or CIDR)",
                default=current_external
            )
            if self._validate_ip_pool_size(external_pool):
                metallb_config.setdefault("pools", {})["external"] = external_pool
                break
            console.print("❌ Pool must have at least 16 IPs", style="red")
        
        # Store the metallb config
        self.config["networking"]["metallb"] = metallb_config
        
        console.print("✅ Configured IP pools:", style="green")
        console.print(f"  Internal: {internal_pool}", style="blue")
        console.print(f"  External: {external_pool}", style="blue")
    
    def _configure_kubernetes(self):
        """Configure Kubernetes settings"""
        console.print("\n☸️ Kubernetes Configuration", style="bold")
        
        self.config["kubernetes"]["version"] = Prompt.ask(
            "K3s version",
            default=self.config["kubernetes"]["version"]
        )
    
    def _configure_services(self):
        """Configure services settings"""
        console.print("\n⚙️ Services Configuration", style="bold")
        
        cert_options = ["pebble-issuer", "letsencrypt-staging", "letsencrypt-prod"]
        console.print("Certificate providers: " + ", ".join(cert_options))
        
        self.config["services"]["cert_provider"] = Prompt.ask(
            "Certificate provider",
            default=self.config["services"]["cert_provider"],
            choices=cert_options
        )
    
    def _configure_system(self):
        """Configure system settings"""
        console.print("\n🖥️ System Configuration", style="bold")
        
        self.config["system"]["use_ofed"] = Confirm.ask(
            "Use OFED drivers?",
            default=self.config["system"]["use_ofed"]
        )
    
    def _configure_secrets(self):
        """Configure sensitive information"""
        console.print("\n🔐 Secrets Configuration", style="bold")
        
        # Git repository credentials
        if Confirm.ask("Configure Git repository credentials?"):
            self.secrets["seed_repo_username"] = Prompt.ask(
                "Git username",
                default=self.secrets.get("seed_repo_username", "")
            )
            
            if Confirm.ask("Update Git password?"):
                self.secrets["seed_repo_password"] = getpass.getpass("Git password: ")
        
        # Container registry credentials
        if Confirm.ask("Configure container registry credentials?"):
            self.secrets["ubiquity_reg_user"] = Prompt.ask(
                "Ubiquity registry username",
                default=self.secrets.get("ubiquity_reg_user", "")
            )
            
            if Confirm.ask("Update Ubiquity registry token?"):
                self.secrets["ubiquity_reg_token"] = getpass.getpass("Ubiquity registry token: ")
        
        # Generate missing passwords
        if not self.secrets.get("argocd_admin_password"):
            if Confirm.ask("Generate ArgoCD admin password?"):
                import secrets as sec
                self.secrets["argocd_admin_password"] = sec.token_urlsafe(16)
                console.print("✅ Generated ArgoCD admin password", style="green")
        
        # K3s encryption secret
        if not self.secrets.get("k3s_encryption_secret"):
            if Confirm.ask("Generate K3s encryption secret?"):
                import secrets as sec
                self.secrets["k3s_encryption_secret"] = sec.token_urlsafe(32)
                console.print("✅ Generated K3s encryption secret", style="green")
    
    def _save_config(self):
        """Save non-sensitive configuration"""
        yaml_handler = YAML()
        yaml_handler.preserve_quotes = True
        yaml_handler.explicit_start = True
        yaml_handler.default_flow_style = False
        
        with open(self.config_file, 'w') as f:
            yaml_handler.dump(self.config, f)
        
        console.print("💾 Configuration saved", style="green")

def main():
    """Main function"""
    config = SOPSConfig()
    
    try:
        # Initialize SOPS
        if not config.initialize():
            sys.exit(1)
        
        if len(sys.argv) > 1:
            command = sys.argv[1]
            
            if command == "configure":
                config.configure_interactive()
            elif command == "update":
                config.update_configurations_in_place()
            elif command == "argocd-sops":
                config.configure_argocd_sops()
            elif command == "edit":
                subprocess.run(['sops', str(config.secrets_file)])
            else:
                console.print(f"❌ Unknown command: {command}", style="red")
                console.print("Available commands: configure, update, argocd-sops, edit", style="blue")
                sys.exit(1)
        else:
            # Interactive mode
            config.configure_interactive()
            
    except KeyboardInterrupt:
        console.print("\n👋 Configuration cancelled", style="yellow")
    except Exception as e:
        console.print(f"❌ Error: {e}", style="red")
        sys.exit(1)

if __name__ == '__main__':
    main()
