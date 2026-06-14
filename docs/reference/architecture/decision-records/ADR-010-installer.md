# ADR-010: Use Go PXE installer instead of Python/Docker

**Status:** Accepted

**Date:** 2026-05-26

## Context

Original PXE server was a Docker Compose-based setup using dhcpd + tftpd + httpd (Apache).
The upstream homelab project rewrote theirs in Go using the pixiecore library, which provides
a single-binary PXE solution with DHCP proxy support.

## Decision

Rewrite the PXE server as a Go binary using pixiecore. Single static binary, no Docker
dependency for the installer, built-in DHCP proxy mode avoids conflicts with existing
network DHCP servers.

## Consequences

- Positive: Single binary, no Docker required for PXE boot
- Positive: DHCP proxy mode works alongside existing network DHCP (opt-in)
- Positive: Built-in phone-home API for installation coordination
- Negative: Must manually extract kernel/initrd from OS install media
- Neutral: MAC to host mapping is static (config map in Go source, can be externalized later)