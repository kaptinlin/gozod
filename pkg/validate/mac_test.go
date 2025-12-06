package validate_test

import (
	"testing"

	"github.com/kaptinlin/gozod/pkg/validate"
)

// TestMAC tests MAC address validation with default colon delimiter
func TestMAC(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  bool
	}{
		// Valid MAC addresses with uppercase
		{"valid uppercase with colon", "00:1A:2B:3C:4D:5E", true},
		{"valid uppercase all F", "FF:FF:FF:FF:FF:FF", true},

		// Valid MAC addresses with lowercase
		{"valid lowercase with colon", "00:1a:2b:3c:4d:5e", true},
		{"valid lowercase all f", "ff:ff:ff:ff:ff:ff", true},

		// Invalid - wrong delimiter
		{"wrong delimiter hyphen", "00-1A-2B-3C-4D-5E", false},
		{"wrong delimiter dot", "00.1A.2B.3C.4D.5E", false},

		// Invalid - wrong format
		{"too short", "00:1A:2B:3C:4D", false},
		{"too long", "00:1A:2B:3C:4D:5E:6F", false},
		{"invalid char G", "00:1A:2G:3C:4D:5E", false},
		{"no delimiters", "001A2B3C4D5E", false},

		// Invalid - wrong type
		{"nil value", nil, false},
		{"int value", 123, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validate.MAC(tt.input)
			if got != tt.want {
				t.Errorf("MAC(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestMACWithOptions tests MAC address validation with custom delimiters
func TestMACWithOptions(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		delimiter string
		want      bool
	}{
		// Hyphen delimiter
		{"hyphen valid uppercase", "00-1A-2B-3C-4D-5E", "-", true},
		{"hyphen valid lowercase", "00-1a-2b-3c-4d-5e", "-", true},
		{"hyphen wrong delimiter", "00:1A:2B:3C:4D:5E", "-", false},

		// Dot delimiter
		{"dot valid uppercase", "00.1A.2B.3C.4D.5E", ".", true},
		{"dot valid lowercase", "00.1a.2b.3c.4d.5e", ".", true},
		{"dot wrong delimiter", "00:1A:2B:3C:4D:5E", ".", false},

		// Custom delimiter (space)
		{"space delimiter valid", "00 1A 2B 3C 4D 5E", " ", true},
		{"space delimiter invalid", "00:1A:2B:3C:4D:5E", " ", false},

		// Empty delimiter defaults to colon
		{"empty delimiter defaults", "00:1A:2B:3C:4D:5E", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := validate.MACOptions{Delimiter: tt.delimiter}
			got := validate.MACWithOptions(tt.input, opts)
			if got != tt.want {
				t.Errorf("MACWithOptions(%v, delimiter=%q) = %v, want %v",
					tt.input, tt.delimiter, got, tt.want)
			}
		})
	}
}

// TestMACCaseSensitivity verifies that both uppercase and lowercase are accepted
// but the regex uses two separate branches (aligned with Zod implementation)
func TestMACCaseSensitivity(t *testing.T) {
	testCases := []string{
		"00:1A:2B:3C:4D:5E", // uppercase
		"00:1a:2b:3c:4d:5e", // lowercase
		"AA:BB:CC:DD:EE:FF", // all uppercase
		"aa:bb:cc:dd:ee:ff", // all lowercase
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			if !validate.MAC(tc) {
				t.Errorf("MAC(%q) should be valid", tc)
			}
		})
	}
}
