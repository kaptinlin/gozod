package validate_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/gozod/pkg/validate"
)

func TestRegex(t *testing.T) {
	t.Parallel()

	pattern := regexp.MustCompile(`^gozod-[0-9]+$`)

	tests := []struct {
		name    string
		value   any
		pattern *regexp.Regexp
		want    bool
	}{
		{name: "matches pattern", value: "gozod-42", pattern: pattern, want: true},
		{name: "rejects mismatch", value: "zod-42", pattern: pattern, want: false},
		{name: "rejects non string", value: 42, pattern: pattern, want: false},
		{name: "rejects nil pattern", value: "gozod-42", pattern: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, validate.Regex(tt.value, tt.pattern))
		})
	}
}

func TestStringCasePredicates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{name: "lowercase accepts symbols", got: validate.Lowercase("gozod-123"), want: true},
		{name: "lowercase rejects uppercase", got: validate.Lowercase("goZod"), want: false},
		{name: "uppercase accepts symbols", got: validate.Uppercase("GOZOD-123"), want: true},
		{name: "uppercase rejects lowercase", got: validate.Uppercase("GOzod"), want: false},
		{name: "lowercase rejects non string", got: validate.Lowercase(123), want: false},
		{name: "uppercase rejects non string", got: validate.Uppercase(123), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.got)
		})
	}
}

func TestStringContainsPredicates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{name: "includes substring", got: validate.Includes("gozod", "zo"), want: true},
		{name: "includes rejects missing substring", got: validate.Includes("gozod", "ts"), want: false},
		{name: "includes rejects non string", got: validate.Includes(123, "2"), want: false},
		{name: "starts with prefix", got: validate.StartsWith("gozod", "go"), want: true},
		{name: "starts with rejects missing prefix", got: validate.StartsWith("gozod", "zo"), want: false},
		{name: "starts with rejects non string", got: validate.StartsWith(123, "1"), want: false},
		{name: "ends with suffix", got: validate.EndsWith("gozod", "zod"), want: true},
		{name: "ends with rejects missing suffix", got: validate.EndsWith("gozod", "go"), want: false},
		{name: "ends with rejects non string", got: validate.EndsWith(123, "3"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.got)
		})
	}
}
