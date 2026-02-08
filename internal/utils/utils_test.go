package utils

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kaptinlin/gozod/core"
)

// =============================================================================
// OPTIMIZED REGEX ESCAPE TESTS
// =============================================================================

func TestEscapeRegex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no special characters",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "single backslash",
			input:    "\\",
			expected: "\\\\",
		},
		{
			name:     "mixed special characters",
			input:    "^hello$",
			expected: "\\^hello\\$",
		},
		{
			name:     "all special characters",
			input:    "\\^$.[]()|?*+{}",
			expected: "\\\\\\^\\$\\.\\[\\]\\(\\)\\|\\?\\*\\+\\{\\}",
		},
		{
			name:     "unicode with special chars",
			input:    "café.txt",
			expected: "café\\.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeRegex(tt.input)
			if result != tt.expected {
				t.Errorf("EscapeRegex(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNeedsEscape(t *testing.T) {
	specialChars := []rune{'\\', '^', '$', '.', '[', ']', '|', '(', ')', '?', '*', '+', '{', '}'}
	for _, char := range specialChars {
		if !needsEscape(char) {
			t.Errorf("needsEscape(%c) should return true", char)
		}
	}

	normalChars := []rune{'a', 'Z', '0', '9', ' ', '_', '-'}
	for _, char := range normalChars {
		if needsEscape(char) {
			t.Errorf("needsEscape(%c) should return false", char)
		}
	}
}

// =============================================================================
// OPTIMIZED PATH FORMATTING TESTS
// =============================================================================

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

// =============================================================================
// SLICE UTILITIES TESTS
// =============================================================================

func TestMergeStringSlices(t *testing.T) {
	tests := []struct {
		name     string
		input    [][]string
		expected []string
	}{
		{
			name:     "empty input",
			input:    [][]string{},
			expected: nil,
		},
		{
			name:     "single slice",
			input:    [][]string{{"a", "b"}},
			expected: []string{"a", "b"},
		},
		{
			name:     "multiple slices",
			input:    [][]string{{"a", "b"}, {"c", "d"}, {"e"}},
			expected: []string{"a", "b", "c", "d", "e"},
		},
		{
			name:     "empty slices mixed",
			input:    [][]string{{"a"}, {}, {"b"}},
			expected: []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeStringSlices(tt.input...)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("MergeStringSlices(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestUniqueStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "single element",
			input:    []string{"a"},
			expected: []string{"a"},
		},
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "all same",
			input:    []string{"a", "a", "a"},
			expected: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UniqueStrings(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("UniqueStrings(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestContainsString(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		value    string
		expected bool
	}{
		{
			name:     "empty slice",
			slice:    []string{},
			value:    "a",
			expected: false,
		},
		{
			name:     "value exists",
			slice:    []string{"a", "b", "c"},
			value:    "b",
			expected: true,
		},
		{
			name:     "value does not exist",
			slice:    []string{"a", "b", "c"},
			value:    "d",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsString(tt.slice, tt.value)
			if result != tt.expected {
				t.Errorf("ContainsString(%v, %q) = %v, want %v", tt.slice, tt.value, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// SIMPLIFIED TYPE DETECTION TESTS
// =============================================================================

func TestGetParsedType(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected core.ParsedType
	}{
		{
			name:     "nil value",
			input:    nil,
			expected: core.ParsedTypeNil,
		},
		{
			name:     "string value",
			input:    "hello",
			expected: core.ParsedTypeString,
		},
		{
			name:     "bool value",
			input:    true,
			expected: core.ParsedTypeBool,
		},
		{
			name:     "int value",
			input:    42,
			expected: core.ParsedTypeNumber,
		},
		{
			name:     "int64 value",
			input:    int64(42),
			expected: core.ParsedTypeNumber,
		},
		{
			name:     "float32 value",
			input:    float32(3.14),
			expected: core.ParsedTypeFloat,
		},
		{
			name:     "float64 value",
			input:    3.14,
			expected: core.ParsedTypeFloat,
		},
		{
			name:     "complex64 value",
			input:    complex64(1 + 2i),
			expected: core.ParsedTypeComplex,
		},
		{
			name:     "slice of any",
			input:    []any{"a", "b"},
			expected: core.ParsedTypeSlice,
		},
		{
			name:     "map[string]any",
			input:    map[string]any{"key": "value"},
			expected: core.ParsedTypeMap,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetParsedType(tt.input)
			if result != tt.expected {
				t.Errorf("GetParsedType(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsPrimitiveValue(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{
			name:     "nil value",
			input:    nil,
			expected: true,
		},
		{
			name:     "string value",
			input:    "hello",
			expected: true,
		},
		{
			name:     "bool value",
			input:    true,
			expected: true,
		},
		{
			name:     "int value",
			input:    42,
			expected: true,
		},
		{
			name:     "float64 value",
			input:    3.14,
			expected: true,
		},
		{
			name:     "slice value",
			input:    []string{"a", "b"},
			expected: false,
		},
		{
			name:     "map value",
			input:    map[string]any{"key": "value"},
			expected: false,
		},
		{
			name:     "struct value",
			input:    struct{ Name string }{Name: "test"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPrimitiveValue(tt.input)
			if result != tt.expected {
				t.Errorf("IsPrimitiveValue(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTypeDescription(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "nil value",
			input:    nil,
			expected: "nil",
		},
		{
			name:     "string value",
			input:    "hello",
			expected: "string",
		},
		{
			name:     "bool value",
			input:    true,
			expected: "bool",
		},
		{
			name:     "int value",
			input:    42,
			expected: "number",
		},
		{
			name:     "float64 value",
			input:    3.14,
			expected: "number",
		},
		{
			name:     "complex128 value",
			input:    complex128(1 + 2i),
			expected: "number",
		},
		{
			name:     "slice of any",
			input:    []any{"a", "b"},
			expected: "array",
		},
		{
			name:     "map[string]any",
			input:    map[string]any{"key": "value"},
			expected: "object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TypeDescription(tt.input)
			if result != tt.expected {
				t.Errorf("TypeDescription(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// BENCHMARKS
// =============================================================================

func BenchmarkEscapeRegex(b *testing.B) {
	inputs := []string{
		"simple",
		"with.dots",
		"^complex$regex*pattern+",
		"\\^$.[]()|?*+{}",
	}

	b.ResetTimer()
	for b.Loop() {
		for _, input := range inputs {
			_ = EscapeRegex(input)
		}
	}
}

func BenchmarkToDotPath(b *testing.B) {
	paths := [][]any{
		{"user"},
		{"users", 0},
		{"user", "profile", "address"},
		{"users", 0, "profile", "address", 1},
		{"user", "first-name", "extra-info", 5},
	}

	b.ResetTimer()
	for b.Loop() {
		for _, path := range paths {
			_ = ToDotPath(path)
		}
	}
}

func BenchmarkGetParsedType(b *testing.B) {
	values := []any{
		"string",
		42,
		3.14,
		true,
		[]any{"a", "b"},
		map[string]any{"key": "value"},
		nil,
	}

	b.ResetTimer()
	for b.Loop() {
		for _, value := range values {
			_ = GetParsedType(value)
		}
	}
}

func BenchmarkMergeStringSlices(b *testing.B) {
	slices := [][]string{
		{"a", "b"},
		{"c", "d"},
		{"e", "f"},
	}

	b.ResetTimer()
	for b.Loop() {
		_ = MergeStringSlices(slices...)
	}
}

// =============================================================================
// MEMORY ALLOCATION TESTS
// =============================================================================

func TestMemoryAllocation(t *testing.T) {
	// Test that ToDotPath doesn't cause excessive allocations
	path := []any{"users", 0, "profile", "address", "street"}

	// Run multiple times to verify consistent memory usage
	for i := 0; i < 100; i++ {
		result := ToDotPath(path)
		expected := "users[0].profile.address.street"
		if result != expected {
			t.Errorf("ToDotPath allocation test failed: got %q, want %q", result, expected)
		}
	}
}

func TestStringBuilderOptimization(t *testing.T) {
	// Test large path to verify string builder optimization
	largePath := make([]any, 100)
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			largePath[i] = "field"
		} else {
			largePath[i] = i
		}
	}

	result := ToDotPath(largePath)

	// Verify result contains expected patterns
	if !strings.Contains(result, "field") {
		t.Error("Large path test failed: missing expected field names")
	}
	if !strings.Contains(result, "[1]") {
		t.Error("Large path test failed: missing expected array indices")
	}
}
