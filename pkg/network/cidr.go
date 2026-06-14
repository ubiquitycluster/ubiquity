// Package network provides IP address management (IPAM) and DNS utility functions.
package network

import (
	"encoding/binary"
	"fmt"
	"net"
)

// IsValidCIDR reports whether s is a valid IPv4 CIDR notation string.
func IsValidCIDR(s string) bool {
	_, _, err := net.ParseCIDR(s)
	return err == nil
}

// NetworkToNetmask converts a CIDR notation to a dotted-quad netmask.
// Example: "10.0.0.0/22" → "255.255.252.0"
func NetworkToNetmask(cidr string) (string, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", fmt.Errorf("invalid CIDR: %w", err)
	}
	mask := ipnet.Mask
	if len(mask) != 4 {
		return "", fmt.Errorf("not a valid IPv4 mask")
	}
	return fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3]), nil
}

// NetworkToBroadcast calculates the broadcast address from a CIDR.
// Example: "10.0.0.0/22" → "10.0.3.255"
func NetworkToBroadcast(cidr string) (string, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", fmt.Errorf("invalid CIDR: %w", err)
	}
	ip := ipnet.IP.To4()
	if ip == nil {
		return "", fmt.Errorf("not an IPv4 address")
	}
	ones, bits := ipnet.Mask.Size()
	mask := net.CIDRMask(ones, bits)
	broadcast := make(net.IP, 4)
	for i := 0; i < 4; i++ {
		broadcast[i] = ip[i] | ^mask[i]
	}
	return broadcast.String(), nil
}

// NetworkToGateway returns the gateway IP for a CIDR with the given host offset.
// Example: "10.0.0.0/22", 254 → "10.0.3.254"
func NetworkToGateway(cidr string, offset int) (string, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", fmt.Errorf("invalid CIDR: %w", err)
	}
	ip := ipnet.IP.To4()
	if ip == nil {
		return "", fmt.Errorf("not an IPv4 address")
	}
	ipInt := binary.BigEndian.Uint32(ip)
	ipInt += uint32(offset)
	result := make(net.IP, 4)
	binary.BigEndian.PutUint32(result, ipInt)
	return result.String(), nil
}

// NetworkToRange returns the start and end of the usable host range for a CIDR.
// Example: "10.0.0.0/24" → "10.0.0.1", "10.0.0.254"
func NetworkToRange(cidr string) (start, end string, err error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", "", fmt.Errorf("invalid CIDR: %w", err)
	}
	ip := ipnet.IP.To4()
	if ip == nil {
		return "", "", fmt.Errorf("not an IPv4 address")
	}
	ones, bits := ipnet.Mask.Size()
	mask := net.CIDRMask(ones, bits)

	// First usable address: network + 1
	first := make(net.IP, 4)
	copy(first, ip)
	first[3]++

	// Last usable address: broadcast - 1
	broadcast := make(net.IP, 4)
	for i := 0; i < 4; i++ {
		broadcast[i] = ip[i] | ^mask[i]
	}
	broadcast[3]--

	return first.String(), broadcast.String(), nil
}

// DnsmasqConfig generates a dnsmasq configuration snippet for the given domain and DNS server.
func DnsmasqConfig(domain, dnsServer string) string {
	return fmt.Sprintf("server=/%s/%s\n", domain, dnsServer)
}