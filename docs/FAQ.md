# Frequently Asked Questions

## General
**Q: What is Ubiquity?**
A: An HPC cluster lifecycle platform using Infrastructure as Code and GitOps.

**Q: What does HPC mean?**
A: High Performance Computing — clusters of powerful nodes for compute-intensive workloads.

## Installation
**Q: What hardware do I need?**
A: See README. Minimum tested: 3× Lenovo ThinkCentre M700 Tiny (i5-6600T, 16GB RAM, 500GB SSD).

**Q: Can I try it without hardware?**
A: Yes: `ubiquity up --sandbox` creates a local k3d cluster in Docker.

## Configuration
**Q: How do I change the domain?**
A: `ubiquity configure --domain mydomain.com`

**Q: How do I add a worker node?**
A: Update the Ansible inventory in metal/inventories/ then run `ubiquity retry metal`.

**Q: How do I change the Kubernetes version?**
A: Edit the KUBERNETES_VERSION in .env or run `ubiquity configure --interactive`.

## Troubleshooting
**Q: A phase failed. How do I retry?**
A: Run `ubiquity status` to see failed phases, then `ubiquity retry <phase>` (e.g., `ubiquity retry bootstrap`).

**Q: How do I tear down and start fresh?**
A: `ubiquity down` then `ubiquity up --sandbox`.

**Q: Where are logs?**
A: `ubiquity logs` reads from provisioning state at ~/.ubiquity/state.json.
