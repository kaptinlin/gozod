package regexes

import (
	"regexp"
	"testing"
)

// TestPrimitivesNegativeNumbers tests that primitive regexes support negative numbers
func TestPrimitivesNegativeNumbers(t *testing.T) {
	tests := []struct {
		name  string
		regex *regexp.Regexp
		input string
		want  bool
	}{
		// Integer tests
		{"integer positive", Integer, "123", true},
		{"integer negative", Integer, "-123", true},
		{"integer zero", Integer, "0", true},
		{"integer with decimal", Integer, "123.45", false},
		{"integer with text", Integer, "123abc", false},

		// Bigint tests
		{"bigint positive", Bigint, "123", true},
		{"bigint positive with n", Bigint, "123n", true},
		{"bigint negative", Bigint, "-123", true},
		{"bigint negative with n", Bigint, "-123n", true},
		{"bigint zero", Bigint, "0", true},
		{"bigint with decimal", Bigint, "123.45", false},

		// Number tests - critical fix: adding $ anchor prevents partial match
		{"number integer", Number, "123", true},
		{"number negative integer", Number, "-123", true},
		{"number decimal", Number, "123.45", true},
		{"number negative decimal", Number, "-123.45", true},
		{"number zero", Number, "0", true},
		{"number zero decimal", Number, "0.0", true},
		{"number partial match", Number, "123abc", false}, // Fixed by adding $
		{"number with text before", Number, "abc123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.regex.MatchString(tt.input)
			if got != tt.want {
				t.Errorf("regex.MatchString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestBooleanNullUndefinedExactMatch tests that special values match exactly
func TestBooleanNullUndefinedExactMatch(t *testing.T) {
	tests := []struct {
		name  string
		regex *regexp.Regexp
		input string
		want  bool
	}{
		// Boolean tests
		{"boolean true lowercase", Boolean, "true", true},
		{"boolean false lowercase", Boolean, "false", true},
		{"boolean true uppercase", Boolean, "TRUE", true},
		{"boolean false uppercase", Boolean, "FALSE", true},
		{"boolean true mixed", Boolean, "True", true},
		{"boolean partial truthy", Boolean, "truthy", false},
		{"boolean partial falsey", Boolean, "falsey", false},

		// Null tests
		{"null lowercase", Null, "null", true},
		{"null uppercase", Null, "NULL", true},
		{"null mixed", Null, "Null", true},
		{"null partial nullable", Null, "nullable", false},

		// Undefined tests
		{"undefined lowercase", Undefined, "undefined", true},
		{"undefined uppercase", Undefined, "UNDEFINED", true},
		{"undefined mixed", Undefined, "Undefined", true},
		{"undefined partial", Undefined, "undefinedValue", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.regex.MatchString(tt.input)
			if got != tt.want {
				t.Errorf("regex.MatchString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
