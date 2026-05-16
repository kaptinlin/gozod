package locales

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
)

func TestFormatMessageEn_CoreIssueMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		issue core.ZodRawIssue
		want  string
	}{
		{
			name: "invalid stringbool displays boolean",
			issue: core.ZodRawIssue{
				Code:       core.InvalidType,
				Input:      "maybe",
				Properties: map[string]any{"expected": "stringbool"},
			},
			want: "Invalid input: expected boolean, received string",
		},
		{
			name: "invalid complex displays number",
			issue: core.ZodRawIssue{
				Code:       core.InvalidType,
				Input:      "3+4i",
				Properties: map[string]any{"expected": "complex128"},
			},
			want: "Invalid input: expected number, received string",
		},
		{
			name: "invalid single value",
			issue: core.ZodRawIssue{
				Code:       core.InvalidValue,
				Properties: map[string]any{"values": []any{"draft"}},
			},
			want: `Invalid input: expected "draft"`,
		},
		{
			name: "invalid multiple values",
			issue: core.ZodRawIssue{
				Code:       core.InvalidValue,
				Properties: map[string]any{"values": []any{"draft", "published"}},
			},
			want: `Invalid option: expected one of "draft"|"published"`,
		},
		{
			name: "unrecognized single key",
			issue: core.ZodRawIssue{
				Code:       core.UnrecognizedKeys,
				Properties: map[string]any{"keys": []string{"extra"}},
			},
			want: `Unrecognized key: "extra"`,
		},
		{
			name: "unrecognized multiple keys",
			issue: core.ZodRawIssue{
				Code:       core.UnrecognizedKeys,
				Properties: map[string]any{"keys": []string{"first", "second"}},
			},
			want: `Unrecognized keys: "first", "second"`,
		},
		{
			name: "invalid key names origin",
			issue: core.ZodRawIssue{
				Code:       core.InvalidKey,
				Properties: map[string]any{"origin": "record"},
			},
			want: "Invalid key in record",
		},
		{
			name: "invalid element formats nested error",
			issue: core.ZodRawIssue{
				Code: core.InvalidElement,
				Properties: map[string]any{"element_error": core.ZodRawIssue{
					Code:       core.InvalidValue,
					Properties: map[string]any{"values": []any{"red"}},
				}},
			},
			want: `Invalid input: expected "red"`,
		},
		{
			name: "invalid element names origin",
			issue: core.ZodRawIssue{
				Code:       core.InvalidElement,
				Properties: map[string]any{"origin": "set"},
			},
			want: "Invalid value in set",
		},
		{
			name: "missing required names field",
			issue: core.ZodRawIssue{
				Code:       core.MissingRequired,
				Properties: map[string]any{"field_name": "email", "field_type": "field"},
			},
			want: "Missing required field: email",
		},
		{
			name: "type conversion names source and target",
			issue: core.ZodRawIssue{
				Code:       core.TypeConversion,
				Properties: map[string]any{"from_type": "string", "to_type": "int"},
			},
			want: "Type conversion failed: cannot convert string to int",
		},
		{
			name: "invalid schema includes reason",
			issue: core.ZodRawIssue{
				Code:       core.InvalidSchema,
				Properties: map[string]any{"reason": "duplicate discriminator"},
			},
			want: "Invalid schema: duplicate discriminator",
		},
		{
			name: "invalid discriminator names field",
			issue: core.ZodRawIssue{
				Code:       core.InvalidDiscriminator,
				Properties: map[string]any{"field": "kind"},
			},
			want: "Invalid or missing discriminator field: kind",
		},
		{
			name: "incompatible types names conflict",
			issue: core.ZodRawIssue{
				Code:       core.IncompatibleTypes,
				Properties: map[string]any{"conflict_type": "objects"},
			},
			want: "Cannot merge objects: incompatible types",
		},
		{
			name: "custom uses message",
			issue: core.ZodRawIssue{
				Code:       core.Custom,
				Properties: map[string]any{"message": "Use a stronger password"},
			},
			want: "Use a stronger password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, FormatMessageEn(tt.issue))
		})
	}
}

func TestFormatMessageEn_SizeConstraintMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		issue core.ZodRawIssue
		want  string
	}{
		{
			name:  "too small without threshold",
			issue: core.ZodRawIssue{Code: core.TooSmall},
			want:  "Too small",
		},
		{
			name:  "too big without threshold",
			issue: core.ZodRawIssue{Code: core.TooBig},
			want:  "Too big",
		},
		{
			name: "string minimum",
			issue: core.ZodRawIssue{
				Code:       core.TooSmall,
				Properties: map[string]any{"origin": "string", "minimum": 3, "inclusive": true},
			},
			want: "Too small: expected string to have at least 3 characters",
		},
		{
			name: "array maximum exclusive",
			issue: core.ZodRawIssue{
				Code:       core.TooBig,
				Properties: map[string]any{"origin": "array", "maximum": 5, "inclusive": false},
			},
			want: "Too big: expected array to have less than 5 items",
		},
		{
			name: "file minimum",
			issue: core.ZodRawIssue{
				Code:       core.TooSmall,
				Properties: map[string]any{"origin": "file", "minimum": 1024},
			},
			want: "File size must be at least 1024 bytes",
		},
		{
			name: "number maximum",
			issue: core.ZodRawIssue{
				Code:       core.TooBig,
				Properties: map[string]any{"origin": "number", "maximum": 10, "inclusive": true},
			},
			want: "Too big: expected number to be at most 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, FormatMessageEn(tt.issue))
		})
	}
}

func TestFormatMessageEn_StringFormatMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		issue core.ZodRawIssue
		want  string
	}{
		{
			name: "missing format",
			issue: core.ZodRawIssue{
				Code: core.InvalidFormat,
			},
			want: "Invalid format",
		},
		{
			name: "starts with prefix",
			issue: core.ZodRawIssue{
				Code:       core.InvalidFormat,
				Properties: map[string]any{"format": "starts_with", "prefix": "go"},
			},
			want: `Invalid string: must start with "go"`,
		},
		{
			name: "ends with suffix",
			issue: core.ZodRawIssue{
				Code:       core.InvalidFormat,
				Properties: map[string]any{"format": "ends_with", "suffix": ".org"},
			},
			want: `Invalid string: must end with ".org"`,
		},
		{
			name: "includes substring",
			issue: core.ZodRawIssue{
				Code:       core.InvalidFormat,
				Properties: map[string]any{"format": "includes", "includes": "@"},
			},
			want: `Invalid string: must include "@"`,
		},
		{
			name: "regex pattern",
			issue: core.ZodRawIssue{
				Code:       core.InvalidFormat,
				Properties: map[string]any{"format": "regex", "pattern": "^[a-z]+$"},
			},
			want: "Invalid string: must match pattern ^[a-z]+$",
		},
		{
			name: "known format noun",
			issue: core.ZodRawIssue{
				Code:       core.InvalidFormat,
				Properties: map[string]any{"format": "email"},
			},
			want: "Invalid email address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, FormatMessageEn(tt.issue))
		})
	}
}

func TestENReturnsEnglishLocaleConfig(t *testing.T) {
	t.Parallel()

	config := EN()
	require.NotNil(t, config.LocaleError)
	assert.Equal(t, "Nil pointer encountered", config.LocaleError(core.ZodRawIssue{Code: core.NilPointer}))
}
