package network

import "testing"

func TestIsValidCIDR(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"10.0.0.0/22", true},
		{"192.168.1.0/24", true},
		{"10.0.0.0/33", false},
		{"not-a-cidr", false},
		{"", false},
		{"10.0.0.1/32", true},
	}
	for _, tt := range tests {
		if got := IsValidCIDR(tt.input); got != tt.want {
			t.Errorf("IsValidCIDR(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestNetworkToNetmask(t *testing.T) {
	tests := []struct {
		cidr string
		want string
	}{
		{"10.0.0.0/22", "255.255.252.0"},
		{"192.168.1.0/24", "255.255.255.0"},
		{"10.0.0.0/16", "255.255.0.0"},
		{"10.0.0.0/32", "255.255.255.255"},
	}
	for _, tt := range tests {
		got, err := NetworkToNetmask(tt.cidr)
		if err != nil {
			t.Errorf("NetworkToNetmask(%q) unexpected error: %v", tt.cidr, err)
			continue
		}
		if got != tt.want {
			t.Errorf("NetworkToNetmask(%q) = %q, want %q", tt.cidr, got, tt.want)
		}
	}
}

func TestNetworkToNetmaskInvalid(t *testing.T) {
	_, err := NetworkToNetmask("invalid")
	if err == nil {
		t.Error("expected error for invalid CIDR")
	}
}

func TestNetworkToBroadcast(t *testing.T) {
	tests := []struct {
		cidr string
		want string
	}{
		{"10.0.0.0/22", "10.0.3.255"},
		{"192.168.1.0/24", "192.168.1.255"},
		{"10.0.0.0/16", "10.0.255.255"},
	}
	for _, tt := range tests {
		got, err := NetworkToBroadcast(tt.cidr)
		if err != nil {
			t.Errorf("NetworkToBroadcast(%q) unexpected error: %v", tt.cidr, err)
			continue
		}
		if got != tt.want {
			t.Errorf("NetworkToBroadcast(%q) = %q, want %q", tt.cidr, got, tt.want)
		}
	}
}

func TestNetworkToGateway(t *testing.T) {
	tests := []struct {
		cidr   string
		offset int
		want   string
	}{
		{"10.0.0.0/22", 254, "10.0.0.254"},
		{"10.0.0.0/22", 1, "10.0.0.1"},
		{"10.0.0.0/22", 1022, "10.0.3.254"},
		{"192.168.1.0/24", 1, "192.168.1.1"},
	}
	for _, tt := range tests {
		got, err := NetworkToGateway(tt.cidr, tt.offset)
		if err != nil {
			t.Errorf("NetworkToGateway(%q, %d) unexpected error: %v", tt.cidr, tt.offset, err)
			continue
		}
		if got != tt.want {
			t.Errorf("NetworkToGateway(%q, %d) = %q, want %q", tt.cidr, tt.offset, got, tt.want)
		}
	}
}

func TestNetworkToRange(t *testing.T) {
	start, end, err := NetworkToRange("10.0.0.0/24")
	if err != nil {
		t.Fatal(err)
	}
	if start != "10.0.0.1" {
		t.Errorf("start = %q, want %q", start, "10.0.0.1")
	}
	if end != "10.0.0.254" {
		t.Errorf("end = %q, want %q", end, "10.0.0.254")
	}
}

func TestDnsmasqConfig(t *testing.T) {
	got := DnsmasqConfig("example.com", "8.8.8.8")
	if got != "server=/example.com/8.8.8.8\n" {
		t.Errorf("got %q", got)
	}
}