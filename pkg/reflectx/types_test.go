package reflectx

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
)

func TestIsNil(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"nil", nil, true},
		{"non-nil string", "hello", false},
		{"nil pointer", (*int)(nil), true},
		{"non-nil pointer", &[]int{1, 2, 3}[0], false},
		{"nil slice", []int(nil), true},
		{"empty slice", []int{}, false},
		{"nil map", map[string]int(nil), true},
		{"empty map", map[string]int{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNil(tt.value)
			if result != tt.expected {
				t.Errorf("IsNil(%v) = %v, expected %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestIsString(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"string", "hello", true},
		{"empty string", "", true},
		{"int", 123, false},
		{"nil", nil, false},
		{"bool", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsString(tt.value)
			if result != tt.expected {
				t.Errorf("IsString(%v) = %v, expected %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"int", 123, true},
		{"int8", int8(123), true},
		{"int16", int16(123), true},
		{"int32", int32(123), true},
		{"int64", int64(123), true},
		{"uint", uint(123), true},
		{"uint8", uint8(123), true},
		{"uint16", uint16(123), true},
		{"uint32", uint32(123), true},
		{"uint64", uint64(123), true},
		{"float32", float32(123.45), true},
		{"float64", float64(123.45), true},
		{"complex64", complex64(1 + 2i), true},
		{"complex128", complex128(1 + 2i), true},
		{"string", "123", false},
		{"bool", true, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNumeric(tt.value)
			if result != tt.expected {
				t.Errorf("IsNumeric(%v) = %v, expected %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestParsedType(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected core.ParsedType
	}{
		{"nil", nil, core.ParsedTypeNil},
		{"bool", true, core.ParsedTypeBool},
		{"string", "hello", core.ParsedTypeString},
		{"int", 123, core.ParsedTypeNumber},
		{"float", 123.45, core.ParsedTypeFloat},
		{"complex", 1 + 2i, core.ParsedTypeComplex},
		{"slice", []int{1, 2, 3}, core.ParsedTypeSlice},
		{"array", [3]int{1, 2, 3}, core.ParsedTypeArray},
		{"map", map[string]int{"a": 1}, core.ParsedTypeMap},
		{"struct", struct{ Name string }{Name: "test"}, core.ParsedTypeStruct},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParsedType(tt.value)
			if result != tt.expected {
				t.Errorf("ParsedType(%v) = %v, expected %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestParsedCategory(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{"nil", nil, "nil"},
		{"string", "hello", "string"},
		{"bool", true, "bool"},
		{"int", 123, "number"},
		{"float", 123.45, "number"},
		{"slice", []int{1, 2, 3}, "array"},
		{"map", map[string]int{"a": 1}, "object"},
		{"struct", struct{ Name string }{Name: "test"}, "object"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParsedCategory(tt.value)
			if result != tt.expected {
				t.Errorf("ParsedCategory(%v) = %v, expected %v", tt.value, result, tt.expected)
			}
		})
	}
}
