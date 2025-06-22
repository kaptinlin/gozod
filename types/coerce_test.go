package types

import (
	"errors"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestCoerceBasicFunctionality(t *testing.T) {
	t.Run("Coerce namespace availability", func(t *testing.T) {
		// Test that Coerce namespace is properly initialized
		require.NotNil(t, Coerce.String)
		require.NotNil(t, Coerce.Number)
		require.NotNil(t, Coerce.Bool)
		require.NotNil(t, Coerce.BigInt)
		require.NotNil(t, Coerce.Complex64)
		require.NotNil(t, Coerce.Complex128)
	})

	t.Run("factory function equivalence", func(t *testing.T) {
		// Test namespace vs direct function equivalence
		namespaceSchema := Coerce.String()
		functionSchema := Coerce.String()

		testInput := 42
		expected := "42"

		result1, err1 := namespaceSchema.Parse(testInput)
		result2, err2 := functionSchema.Parse(testInput)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.Equal(t, expected, result1)
		assert.Equal(t, expected, result2)
	})

	t.Run("coercion flag verification", func(t *testing.T) {
		schema := Coerce.String()
		internals := schema.GetInternals()

		coerceFlag, exists := internals.Bag["coerce"].(bool)
		require.True(t, exists)
		assert.True(t, coerceFlag)
	})
}

// =============================================================================
// 2. String coercion
// =============================================================================

func TestCoerceString(t *testing.T) {
	t.Run("basic string coercion", func(t *testing.T) {
		schema := Coerce.String()

		tests := []struct {
			name     string
			input    any
			expected string
		}{
			{"int", 42, "42"},
			{"float", 3.14, "3.14"},
			{"bool true", true, "true"},
			{"bool false", false, "false"},
			{"string", "hello", "hello"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := schema.Parse(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("string coercion with validation", func(t *testing.T) {
		schema := Coerce.String().Min(3)

		// Test successful coercion and validation
		result, err := schema.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, "123", result)

		// Test coercion success but validation failure
		_, err = schema.Parse(12)
		assert.Error(t, err)
	})

	t.Run("string coercion flag verification", func(t *testing.T) {
		schema := Coerce.String()
		internals := schema.GetInternals()

		coerceFlag, exists := internals.Bag["coerce"].(bool)
		require.True(t, exists)
		assert.True(t, coerceFlag)
	})
}

// =============================================================================
// 2.1 Additional string coercion edge-cases
// =============================================================================

func TestCoerceStringAdditionalCases(t *testing.T) {
	schema := Coerce.String()

	tests := []struct {
		name     string
		input    any
		expected string
		wantErr  bool
	}{
		{"empty string", "", "", false},
		{"NaN float", math.NaN(), "NaN", false},
		{"+Inf", math.Inf(1), "+Inf", false},
		{"-Inf", math.Inf(-1), "-Inf", false},
		{"bigint", big.NewInt(15), "15", false},
		{"slice unsupported", []string{"item", "another_item"}, "", true},
		{"array empty unsupported", []int{}, "", true},
		{"map object unsupported", map[string]string{"hello": "world!"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := schema.Parse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// 3. Number coercion
// =============================================================================

func TestCoerceNumber(t *testing.T) {
	t.Run("basic number coercion", func(t *testing.T) {
		schema := Coerce.Number()

		tests := []struct {
			name     string
			input    any
			expected float64
		}{
			{"string number", "123", 123.0},
			{"string float", "123.45", 123.45},
			{"int", 42, 42.0},
			{"bool true", true, 1.0},
			{"bool false", false, 0.0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := schema.Parse(tt.input)
				require.NoError(t, err)
				if math.IsNaN(tt.expected) {
					assert.True(t, math.IsNaN(result.(float64)))
				} else {
					assert.Equal(t, tt.expected, result)
				}
			})
		}
	})

	t.Run("number coercion with validation", func(t *testing.T) {
		schema := Coerce.Number().Min(0).Max(100)

		// Test successful coercion and validation
		result, err := schema.Parse("50")
		require.NoError(t, err)
		assert.Equal(t, 50.0, result)

		// Test coercion success but validation failure
		_, err = schema.Parse("-10")
		assert.Error(t, err)
	})

	t.Run("number coercion flag verification", func(t *testing.T) {
		schema := Coerce.Number()
		internals := schema.GetInternals()

		coerceFlag, exists := internals.Bag["coerce"].(bool)
		require.True(t, exists)
		assert.True(t, coerceFlag)
	})
}

// =============================================================================
// 3.1 Additional number coercion edge-cases
// =============================================================================

func TestCoerceNumberAdditionalCases(t *testing.T) {
	schema := Coerce.Number()

	tests := []struct {
		name     string
		input    any
		expected float64
		wantErr  bool
	}{
		{"empty string", "", 0, false},
		{"string negative", "-12", -12, false},
		{"string float", "3.14", 3.14, false},
		{"string NOT_A_NUMBER", "NOT_A_NUMBER", 0, true},
		{"NaN", math.NaN(), math.NaN(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := schema.Parse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if math.IsNaN(tt.expected) {
				assert.True(t, math.IsNaN(result.(float64)))
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// =============================================================================
// 4. Boolean coercion
// =============================================================================

func TestCoerceBool(t *testing.T) {
	schema := Coerce.Bool()

	t.Run("string to boolean", func(t *testing.T) {
		tests := []struct {
			input    string
			expected bool
		}{
			{"true", true},
			{"false", false},
			{"1", true},
			{"0", false},
			{"yes", true},
			{"no", false},
			{"", false},
		}

		for _, tt := range tests {
			result, err := schema.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("number to boolean", func(t *testing.T) {
		tests := []struct {
			input    any
			expected bool
		}{
			{1, true},
			{0, false},
			{42, true},
			{-1, true},
			{0.0, false},
			{3.14, true},
		}

		for _, tt := range tests {
			result, err := schema.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("boolean passthrough", func(t *testing.T) {
		tests := []bool{true, false}

		for _, input := range tests {
			result, err := schema.Parse(input)
			require.NoError(t, err)
			assert.Equal(t, input, result)
		}
	})
}

// =============================================================================
// 5. BigInt coercion
// =============================================================================

func TestCoerceBigInt(t *testing.T) {
	schema := Coerce.BigInt()

	t.Run("integer to BigInt", func(t *testing.T) {
		tests := []struct {
			input    any
			expected string
		}{
			{42, "42"},
			{int64(123), "123"},
			{uint64(456), "456"},
		}

		for _, tt := range tests {
			result, err := schema.Parse(tt.input)
			require.NoError(t, err)
			bigInt, ok := result.(*big.Int)
			require.True(t, ok)
			assert.Equal(t, tt.expected, bigInt.String())
		}
	})

	t.Run("string to BigInt", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"42", "42"},
			{"123456789012345678901234567890", "123456789012345678901234567890"},
			{"-42", "-42"},
		}

		for _, tt := range tests {
			result, err := schema.Parse(tt.input)
			require.NoError(t, err)
			bigInt, ok := result.(*big.Int)
			require.True(t, ok)
			assert.Equal(t, tt.expected, bigInt.String())
		}
	})

	t.Run("boolean to BigInt", func(t *testing.T) {
		tests := []struct {
			input    bool
			expected string
		}{
			{true, "1"},
			{false, "0"},
		}

		for _, tt := range tests {
			result, err := schema.Parse(tt.input)
			require.NoError(t, err)
			bigInt, ok := result.(*big.Int)
			require.True(t, ok)
			assert.Equal(t, tt.expected, bigInt.String())
		}
	})
}

// =============================================================================
// 5.1 Additional bigint coercion edge-cases
// =============================================================================

func TestCoerceBigIntAdditionalCases(t *testing.T) {
	schema := Coerce.BigInt()

	tests := []struct {
		name     string
		input    any
		expected string
		wantErr  bool
	}{
		{"empty string", "", "0", false},
		{"string negative", "-5", "-5", false},
		{"string float", "3.14", "", true},
		{"NaN", math.NaN(), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := schema.Parse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			bigInt, ok := result.(*big.Int)
			require.True(t, ok)
			assert.Equal(t, tt.expected, bigInt.String())
		})
	}
}

// =============================================================================
// 6. Complex number coercion
// =============================================================================

func TestCoerceComplex(t *testing.T) {
	t.Run("Complex64 coercion", func(t *testing.T) {
		schema := Coerce.Complex64()

		tests := []struct {
			name     string
			input    any
			expected complex64
		}{
			{"int", 42, complex64(42 + 0i)},
			{"float", 3.14, complex64(3.14 + 0i)},
			{"complex64", complex64(1 + 2i), complex64(1 + 2i)},
			{"complex128", complex128(3 + 4i), complex64(3 + 4i)},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := schema.Parse(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("Complex128 coercion", func(t *testing.T) {
		schema := Coerce.Complex128()

		tests := []struct {
			name     string
			input    any
			expected complex128
		}{
			{"int", 42, complex128(42 + 0i)},
			{"float", 3.14, complex128(3.14 + 0i)},
			{"complex64", complex64(1 + 2i), complex128(1 + 2i)},
			{"complex128", complex128(3 + 4i), complex128(3 + 4i)},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := schema.Parse(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

// =============================================================================
// 10. Validation integration
// =============================================================================

func TestCoerceValidationIntegration(t *testing.T) {
	t.Run("string coercion with validation", func(t *testing.T) {
		schema := Coerce.String().Min(3).Max(5)

		// Valid after coercion
		result, err := schema.Parse(1234)
		require.NoError(t, err)
		assert.Equal(t, "1234", result)

		// Invalid after coercion (too short)
		_, err = schema.Parse(12)
		assert.Error(t, err)

		// Invalid after coercion (too long)
		_, err = schema.Parse(123456)
		assert.Error(t, err)
	})

	t.Run("boolean coercion with refine", func(t *testing.T) {
		schema := Coerce.Bool().RefineAny(func(val any) bool {
			if b, ok := val.(bool); ok {
				return b == true // Only allow true values
			}
			return false
		}, core.SchemaParams{Error: "Must be true"})

		// Valid after coercion
		result, err := schema.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Invalid after coercion
		_, err = schema.Parse("false")
		assert.Error(t, err)
	})
}

// =============================================================================
// 11. Error handling and edge cases
// =============================================================================

func TestCoerceErrorHandling(t *testing.T) {
	t.Run("unsupported types", func(t *testing.T) {
		schema := Coerce.String()

		unsupportedTypes := []any{
			[]int{1, 2, 3},
			map[string]int{"key": 1},
			struct{}{},
			make(chan int),
			func() {},
		}

		for _, input := range unsupportedTypes {
			_, err := schema.Parse(input)
			assert.Error(t, err)
		}
	})

	t.Run("nil handling", func(t *testing.T) {
		schema := Coerce.String()

		_, err := schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("pointer handling", func(t *testing.T) {
		schema := Coerce.String()

		// Valid pointer
		intVal := 42
		result, err := schema.Parse(&intVal)
		require.NoError(t, err)
		assert.Equal(t, "42", result)

		// Nil pointer
		var nilPtr *int
		result, err = schema.Parse(nilPtr)
		if err != nil {
			assert.Error(t, err)
		} else {
			assert.NotNil(t, result)
		}
	})

	t.Run("special float values", func(t *testing.T) {
		schema := Coerce.String()

		tests := []struct {
			name     string
			input    any
			expected string
		}{
			{"infinity", math.Inf(1), "+Inf"},
			{"negative infinity", math.Inf(-1), "-Inf"},
			{"NaN", math.NaN(), "NaN"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := schema.Parse(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("custom error messages", func(t *testing.T) {
		customError := "core.Custom coercion error"
		schema := Coerce.String(core.SchemaParams{Error: customError})

		// Verify coercion is enabled
		internals := schema.GetInternals()
		coerceFlag, exists := internals.Bag["coerce"].(bool)
		require.True(t, exists)
		assert.True(t, coerceFlag)

		// Verify custom error is preserved
		schemaInternals := schema.GetInternals()
		assert.NotNil(t, schemaInternals.Error)
	})
}

// =============================================================================
// 12. Integration and workflow tests
// =============================================================================

func TestCoerceIntegration(t *testing.T) {
	t.Run("coerce with function schema", func(t *testing.T) {
		// Test that coercion works with function schemas
		functionSchema := Coerce.String()

		result, err := functionSchema.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, "123", result)
	})

	t.Run("complex validation pipeline", func(t *testing.T) {
		schema := Coerce.String().
			Min(5).
			Max(20).
			RefineAny(func(val any) bool {
				if str, ok := val.(string); ok {
					return !strings.Contains(str, "invalid")
				}
				return false
			}, core.SchemaParams{Error: "String cannot contain 'invalid'"})

		// Valid case
		result, err := schema.Parse(123456)
		require.NoError(t, err)
		assert.Equal(t, "123456", result)

		// Invalid case (too short after coercion)
		_, err = schema.Parse(123)
		assert.Error(t, err)

		// Invalid case (contains "invalid")
		_, err = schema.Parse("invalidstring")
		assert.Error(t, err)
	})

	t.Run("transform integration", func(t *testing.T) {
		schema := Coerce.String().TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
			if str, ok := input.(string); ok {
				return strings.ToUpper(str), nil
			}
			return input, nil
		})

		result, err := schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, "42", result)
	})

	t.Run("type safety throughout pipeline", func(t *testing.T) {
		schema := Coerce.String().TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
			// Ensure input is string after coercion
			str, ok := input.(string)
			require.True(t, ok, "Expected string after coercion")
			return str + "_processed", nil
		})

		result, err := schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, "42_processed", result)
	})

	t.Run("error propagation", func(t *testing.T) {
		schema := Coerce.String().Min(10)

		_, err := schema.Parse(123)
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Equal(t, issues.TooSmall, zodErr.Issues[0].Code)
	})

	t.Run("all coerce types working", func(t *testing.T) {
		// Test that all coerce factory functions work
		stringSchema := Coerce.String()
		numberSchema := Coerce.Number()
		boolSchema := Coerce.Bool()
		bigIntSchema := Coerce.BigInt()
		complex64Schema := Coerce.Complex64()
		complex128Schema := Coerce.Complex128()

		// Basic functionality test
		_, err := stringSchema.Parse(42)
		assert.NoError(t, err)

		_, err = numberSchema.Parse("42")
		assert.NoError(t, err)

		_, err = boolSchema.Parse("true")
		assert.NoError(t, err)

		_, err = bigIntSchema.Parse(42)
		assert.NoError(t, err)

		_, err = complex64Schema.Parse(42)
		assert.NoError(t, err)

		_, err = complex128Schema.Parse(42)
		assert.NoError(t, err)
	})

	t.Run("performance smoke test", func(t *testing.T) {
		schema := Coerce.String()
		testInputs := []any{
			42,
			"hello",
			true,
			3.14,
		}

		// Simple performance check
		for i := 0; i < 100; i++ {
			for _, input := range testInputs {
				_, err := schema.Parse(input)
				require.NoError(t, err)
			}
		}
	})
}
