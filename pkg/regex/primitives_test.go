package regex

import (
	"regexp"
	"testing"
)

func TestPrimitivesNegativeNumbers(t *testing.T) {
	tests := []struct {
		name  string
		re    *regexp.Regexp
		input string
		want  bool
	}{
		{"integer/positive", Integer, "123", true},
		{"integer/negative", Integer, "-123", true},
		{"integer/zero", Integer, "0", true},
		{"integer/decimal", Integer, "123.45", false},
		{"integer/text", Integer, "123abc", false},

		{"bigint/positive", Bigint, "123", true},
		{"bigint/positive_n", Bigint, "123n", true},
		{"bigint/negative", Bigint, "-123", true},
		{"bigint/negative_n", Bigint, "-123n", true},
		{"bigint/zero", Bigint, "0", true},
		{"bigint/decimal", Bigint, "123.45", false},

		{"number/integer", Number, "123", true},
		{"number/negative", Number, "-123", true},
		{"number/decimal", Number, "123.45", true},
		{"number/negative_decimal", Number, "-123.45", true},
		{"number/zero", Number, "0", true},
		{"number/zero_decimal", Number, "0.0", true},
		{"number/trailing_text", Number, "123abc", false},
		{"number/leading_text", Number, "abc123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.re.MatchString(tt.input); got != tt.want {
				t.Errorf("MatchString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestBooleanNullUndefined(t *testing.T) {
	tests := []struct {
		name  string
		re    *regexp.Regexp
		input string
		want  bool
	}{
		{"boolean/true", Boolean, "true", true},
		{"boolean/false", Boolean, "false", true},
		{"boolean/TRUE", Boolean, "TRUE", true},
		{"boolean/FALSE", Boolean, "FALSE", true},
		{"boolean/True", Boolean, "True", true},
		{"boolean/truthy", Boolean, "truthy", false},
		{"boolean/falsey", Boolean, "falsey", false},

		{"null/lowercase", Null, "null", true},
		{"null/uppercase", Null, "NULL", true},
		{"null/mixed", Null, "Null", true},
		{"null/nullable", Null, "nullable", false},

		{"undefined/lowercase", Undefined, "undefined", true},
		{"undefined/uppercase", Undefined, "UNDEFINED", true},
		{"undefined/mixed", Undefined, "Undefined", true},
		{"undefined/partial", Undefined, "undefinedValue", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.re.MatchString(tt.input); got != tt.want {
				t.Errorf("MatchString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
