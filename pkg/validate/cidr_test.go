package validate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/gozod/pkg/validate"
)

func TestCIDR(t *testing.T) {
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
		{"invalid prefix", "192.168.1.0/33", 4, false},
		{"invalid prefix ipv6", "fe80::/129", 6, false},

		// Version mismatch
		{"ipv4 as ipv6", "192.168.1.0/24", 6, false},
		{"ipv6 as ipv4", "fe80::/64", 4, false},

		// Using generic version check
		{"ipv4 generic", "192.168.1.0/24", 0, true},
		{"ipv6 generic", "fe80::/64", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validate.CIDR(tt.input, tt.version)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCIDRv4v6Wrappers(t *testing.T) {
	assert.True(t, validate.CIDRv4("192.168.1.0/24"), "CIDRv4 should return true for valid IPv4 CIDR")
	assert.False(t, validate.CIDRv4("fe80::/64"), "CIDRv4 should return false for IPv6 CIDR")

	assert.True(t, validate.CIDRv6("fe80::/64"), "CIDRv6 should return true for valid IPv6 CIDR")
	assert.False(t, validate.CIDRv6("192.168.1.0/24"), "CIDRv6 should return false for IPv4 CIDR")
}
