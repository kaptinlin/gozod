package engine

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
)

// =============================================================================
// BENCHMARK TESTS FOR PARSER OPTIMIZATIONS
// =============================================================================

// BenchmarkParsePrimitive measures the performance of the optimized ParsePrimitive function
func BenchmarkParsePrimitive(b *testing.B) {
	internals := &core.ZodTypeInternals{
		Type:    "string",
		Checks:  []core.ZodCheck{},
		Nilable: false,
		Bag:     map[string]any{},
	}

	ctx := core.NewParseContext()

	testCases := []struct {
		name  string
		input any
	}{
		{"string_value", "test"},
		{"string_pointer", stringPtr("test")},
		{"int_value", 123},
		{"int_pointer", intPtr(123)},
		{"bool_value", true},
		{"bool_pointer", boolPtr(true)},
		{"invalid_type", 3.14},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = ParsePrimitive[string](tc.input, internals, core.ZodTypeString, nil, ctx)
			}
		})
	}
}

// BenchmarkParseType measures the performance of the optimized ParseType function
func BenchmarkParseType(b *testing.B) {
	internals := &core.ZodTypeInternals{
		Type:    "string",
		Checks:  []core.ZodCheck{},
		Nilable: false,
		Bag:     map[string]any{},
	}

	ctx := core.NewParseContext()

	stringChecker := func(input any) (string, bool) {
		if s, ok := input.(string); ok {
			return s, true
		}
		return "", false
	}

	stringPtrChecker := func(input any) (*string, bool) {
		if s, ok := input.(*string); ok {
			return s, true
		}
		return nil, false
	}

	testCases := []struct {
		name  string
		input any
	}{
		{"string_value", "test"},
		{"string_pointer", stringPtr("test")},
		{"nil_input", nil},
		{"invalid_type", 123},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = ParseType[string](tc.input, internals, core.ZodTypeString, stringChecker, stringPtrChecker, nil, ctx)
			}
		})
	}
}

// BenchmarkParseEngineComparison compares the full Parse function performance
func BenchmarkParseEngineComparison(b *testing.B) {
	schema := newMockStringSchema()
	ctx := core.NewParseContext()

	testCases := []struct {
		name  string
		input any
	}{
		{"valid_string", "test"},
		{"empty_string", ""},
		{"long_string", generateLongString(1000)},
		{"invalid_type", 123},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = Parse[any, any](schema, tc.input, ctx)
			}
		})
	}
}

// BenchmarkCoercionPathPerformance measures coercion performance improvements
func BenchmarkCoercionPathPerformance(b *testing.B) {
	internalsWithCoercion := &core.ZodTypeInternals{
		Type:    "string",
		Checks:  []core.ZodCheck{},
		Nilable: false,
		Bag:     map[string]any{"coerce": true},
	}

	internalsWithoutCoercion := &core.ZodTypeInternals{
		Type:    "string",
		Checks:  []core.ZodCheck{},
		Nilable: false,
		Bag:     map[string]any{},
	}

	ctx := core.NewParseContext()

	b.Run("with_coercion", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = ParsePrimitive[string](123, internalsWithCoercion, core.ZodTypeString, nil, ctx)
		}
	})

	b.Run("without_coercion", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = ParsePrimitive[string]("test", internalsWithoutCoercion, core.ZodTypeString, nil, ctx)
		}
	})
}

// BenchmarkNilHandlingPerformance measures nil handling optimization performance
func BenchmarkNilHandlingPerformance(b *testing.B) {
	internals := &core.ZodTypeInternals{
		Type:    "string",
		Checks:  []core.ZodCheck{},
		Nilable: true,
		Bag:     map[string]any{},
	}

	ctx := core.NewParseContext()

	testCases := []struct {
		name  string
		input any
	}{
		{"nil_value", nil},
		{"nil_string_pointer", (*string)(nil)},
		{"nil_int_pointer", (*int)(nil)},
		{"valid_pointer", stringPtr("test")},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = ParsePrimitive[string](tc.input, internals, core.ZodTypeString, nil, ctx)
			}
		})
	}
}

// =============================================================================
// HELPER FUNCTIONS FOR BENCHMARKS
// =============================================================================

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func generateLongString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = 'a' + byte(i%26)
	}
	return string(result)
}
