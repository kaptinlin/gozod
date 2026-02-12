package utils

import (
	"testing"
)

func TestToDotPath(t *testing.T) {
	tests := []struct {
		name     string
		input    []any
		expected string
	}{
		{
			name:     "empty path",
			input:    []any{},
			expected: "",
		},
		{
			name:     "single string",
			input:    []any{"user"},
			expected: "user",
		},
		{
			name:     "string then int",
			input:    []any{"users", 0},
			expected: "users[0]",
		},
		{
			name:     "string then string",
			input:    []any{"user", "name"},
			expected: "user.name",
		},
		{
			name:     "special characters need brackets",
			input:    []any{"user", "first-name"},
			expected: `user["first-name"]`,
		},
		{
			name:     "mixed path types",
			input:    []any{"users", 0, "profile", "address", 1},
			expected: "users[0].profile.address[1]",
		},
		{
			name:     "string with spaces",
			input:    []any{"user", "full name"},
			expected: `user["full name"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToDotPath(tt.input)
			if result != tt.expected {
				t.Errorf("ToDotPath(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNeedsBracketNotation(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"name", false},
		{"firstName", false},
		{"first_name", false},
		{"name123", false},
		{"first-name", true},
		{"first name", true},
		{"user.name", true},
		{"123name", true},
		{"", false},
	}

	for _, tt := range tests {
		result := needsBracketNotation(tt.input)
		if result != tt.expected {
			t.Errorf("needsBracketNotation(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}
