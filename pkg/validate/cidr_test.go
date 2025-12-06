package validate

import (
	"testing"
)

func TestCIDRValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		version int
		want    bool
	}{
		// IPv4 valid cases
		{"valid ipv4", "192.168.1.0/24", 4, true},
		{"valid ipv4 single ip", "192.168.1.1/32", 4, true},
		{"valid ipv4 any", "0.0.0.0/0", 4, true},

		// IPv6 valid cases
		{"valid ipv6", "fe80::/64", 6, true},
		{"valid ipv6 localhost", "::1/128", 6, true},

		// Invalid formats
		{"multiple slashes", "192.168.1.0/24/32", 0, false},
		{"no slash", "192.168.1.0", 0, false},
		{"empty string", "", 0, false},
		{"invalid ip", "300.168.1.0/24", 0, false},
		{"invalid prefix", "192.168.1.0/33", 4, false},  // IPv4 max prefix is 32
		{"invalid prefix ipv6", "fe80::/129", 6, false}, // IPv6 max prefix is 128

		// Version mismatch
		{"ipv4 as ipv6", "192.168.1.0/24", 6, false},
		{"ipv6 as ipv4", "fe80::/64", 4, false},

		// Using generic version check
		{"ipv4 generic", "192.168.1.0/24", 0, true},
		{"ipv6 generic", "fe80::/64", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateCIDR(tt.input, tt.version)
			if got != tt.want {
				t.Errorf("ValidateCIDR(%q, %d) = %v, want %v", tt.input, tt.version, got, tt.want)
			}
		})
	}
}

func TestCIDRv4v6Wrappers(t *testing.T) {
	// Test CIDRv4 wrapper
	if !CIDRv4("192.168.1.0/24") {
		t.Error("CIDRv4 should return true for valid IPv4 CIDR")
	}
	if CIDRv4("fe80::/64") {
		t.Error("CIDRv4 should return false for IPv6 CIDR")
	}

	// Test CIDRv6 wrapper
	if !CIDRv6("fe80::/64") {
		t.Error("CIDRv6 should return true for valid IPv6 CIDR")
	}
	if CIDRv6("192.168.1.0/24") {
		t.Error("CIDRv6 should return false for IPv4 CIDR")
	}
}
