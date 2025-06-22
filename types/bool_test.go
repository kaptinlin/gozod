package types

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/kaptinlin/gozod/pkg/coerce"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestBoolBasicFunctionality(t *testing.T) {
	t.Run("constructors", func(t *testing.T) {
		// Test Bool constructor
		schema := Bool()
		require.NotNil(t, schema)
		assert.Equal(t, "bool", schema.internals.Def.Type)
	})

	t.Run("basic validation", func(t *testing.T) {
		schema := Bool()

		// Valid boolean values
		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		result, err = schema.Parse(false)
		require.NoError(t, err)
		assert.Equal(t, false, result)

		// Invalid types
		invalidInputs := []any{
			"not a boolean", 123, 3.14, []bool{true}, nil,
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err)
		}
	})

	t.Run("custom error", func(t *testing.T) {
		customError := "core.Custom boolean validation error"
		schema := Bool(core.SchemaParams{Error: customError})
		require.NotNil(t, schema)
		assert.Equal(t, "bool", schema.internals.Def.Type)
	})
}

// =============================================================================
// 2. Coerce (type coercion)
// =============================================================================

func TestBoolCoercion(t *testing.T) {
	t.Run("string coercion", func(t *testing.T) {
		schema := CoercedBool()

		// Test coerce.ToBool directly
		if result, err := coerce.ToBool("true"); err == nil {
			t.Logf("coerce.ToBool('true') = %v", result)
		} else {
			t.Logf("coerce.ToBool('true') failed")
		}

		// Test engine.ShouldCoerce function directly
		// We need to get access to the engine.ShouldCoerce function - it's in parse.go but not exported
		// For now, let's manually check the Bag

		// Compare with String coercion - does it work?
		stringSchema := CoercedString()
		stringResult, stringErr := stringSchema.Parse(123)
		if stringErr != nil {
			t.Logf("String coercion failed: %v", stringErr)
		} else {
			t.Logf("String coercion works: %v", stringResult)
		}

		// Truthy strings
		truthyStrings := []string{"true", "TRUE", "True", "1", "yes", "YES", "on", "ON", "y", "Y"}
		for _, input := range truthyStrings {
			result, err := schema.Parse(input)
			require.NoError(t, err)
			assert.True(t, result.(bool), "Expected true for: %s", input)
		}

		// Falsy strings
		falsyStrings := []string{"false", "FALSE", "False", "0", "no", "NO", "off", "OFF", "n", "N", "", "  "}
		for _, input := range falsyStrings {
			result, err := schema.Parse(input)
			require.NoError(t, err)
			assert.False(t, result.(bool), "Expected false for: %s", input)
		}

		// Invalid strings
		invalidStrings := []string{"invalid", "maybe", "2", "true1", "false0"}
		for _, input := range invalidStrings {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for: %s", input)
		}
	})

	t.Run("numeric coercion", func(t *testing.T) {
		schema := CoercedBool()

		// Zero values should be false
		zeroValues := []any{
			int(0), int8(0), int16(0), int32(0), int64(0),
			uint(0), uint8(0), uint16(0), uint32(0), uint64(0),
			float32(0.0), float64(0.0),
		}
		for _, input := range zeroValues {
			result, err := schema.Parse(input)
			require.NoError(t, err)
			assert.False(t, result.(bool))
		}

		// Non-zero values should be true
		nonZeroValues := []any{
			int(1), int(-1), int8(42), float32(1.0), float64(3.14),
		}
		for _, input := range nonZeroValues {
			result, err := schema.Parse(input)
			require.NoError(t, err)
			assert.True(t, result.(bool))
		}
	})

	t.Run("pointer coercion", func(t *testing.T) {
		schema := CoercedBool()

		trueVal := true
		falseVal := false
		oneInt := 1
		zeroInt := 0

		tests := []struct {
			name     string
			input    any
			expected bool
		}{
			{"bool_ptr_true", &trueVal, true},
			{"bool_ptr_false", &falseVal, false},
			{"int_ptr_one", &oneInt, true},
			{"int_ptr_zero", &zeroInt, false},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				result, err := schema.Parse(test.input)
				require.NoError(t, err)
				assert.Equal(t, test.expected, result)
			})
		}

		// Test nil pointer separately
		t.Run("bool_ptr_nil", func(t *testing.T) {
			_, err := schema.Parse((*bool)(nil))
			assert.Error(t, err)
		})

		t.Run("int_ptr_nil", func(t *testing.T) {
			_, err := schema.Parse((*int)(nil))
			assert.Error(t, err)
		})
	})
}

// =============================================================================
// 3. Validation methods
// =============================================================================

func TestBoolValidationMethods(t *testing.T) {
	t.Run("refine validation", func(t *testing.T) {
		schema := Bool().Refine(func(val bool) bool {
			return val == true // Only allow true values
		})

		// Valid case
		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Invalid case
		_, err = schema.Parse(false)
		assert.Error(t, err)
	})

	t.Run("multiple refinements", func(t *testing.T) {
		schema := Bool().
			Refine(func(val bool) bool { return true }).
			Refine(func(val bool) bool { return val == true })

		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		_, err = schema.Parse(false)
		assert.Error(t, err)
	})
}

// =============================================================================
// 4. Modifiers and wrappers
// =============================================================================

func TestBoolModifiers(t *testing.T) {
	t.Run("optional wrapper", func(t *testing.T) {
		schema := Bool().Optional()

		// Valid boolean
		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// False value
		result, err = schema.Parse(false)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("nilable wrapper", func(t *testing.T) {
		schema := Bool().Nilable()

		// Valid boolean
		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Nil value
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nullish wrapper", func(t *testing.T) {
		schema := Bool().Nullish()

		// Valid boolean
		result, err := schema.Parse(false)
		require.NoError(t, err)
		assert.Equal(t, false, result)

		// Nil value
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("default wrapper", func(t *testing.T) {
		defaultValue := true
		schema := Bool().Default(defaultValue)

		// Valid boolean (should not use default)
		result, err := schema.Parse(false)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("defaultFunc wrapper", func(t *testing.T) {
		defaultFn := func() bool {
			return true
		}
		schema := Bool().DefaultFunc(defaultFn)

		require.NotNil(t, schema)
	})
}

// =============================================================================
// 5. Chaining and method composition
// =============================================================================

func TestBoolChaining(t *testing.T) {
	t.Run("method chaining", func(t *testing.T) {
		schema := Bool().
			Refine(func(val bool) bool { return val == true }).
			Optional()

		require.NotNil(t, schema)

		// Test chained functionality
		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("coercion with chaining", func(t *testing.T) {
		schema := CoercedBool().
			Refine(func(val bool) bool { return val == true }).
			Optional()

		// Test string coercion with refinement
		result, err := schema.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Test invalid string
		_, err = schema.Parse("false")
		assert.Error(t, err)
	})
}

// =============================================================================
// 6. Transform/Pipe
// =============================================================================

func TestBoolTransform(t *testing.T) {
	t.Run("transform method", func(t *testing.T) {
		schema := Bool().Transform(func(val bool, ctx *core.RefinementContext) (any, error) {
			if val {
				return "TRUE", nil
			}
			return "FALSE", nil
		})

		// Transform true to "TRUE"
		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, "TRUE", result)

		// Transform false to "FALSE"
		result, err = schema.Parse(false)
		require.NoError(t, err)
		assert.Equal(t, "FALSE", result)
	})

	t.Run("transform chaining", func(t *testing.T) {
		schema := Bool().
			Transform(func(val bool, ctx *core.RefinementContext) (any, error) {
				return val, nil
			}).
			TransformAny(func(val any, ctx *core.RefinementContext) (any, error) {
				if b, ok := val.(bool); ok && b {
					return "SUCCESS", nil
				}
				return "FAILURE", nil
			})

		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, "SUCCESS", result)

		result, err = schema.Parse(false)
		require.NoError(t, err)
		assert.Equal(t, "FAILURE", result)
	})
}

// =============================================================================
// 7. Refine
// =============================================================================

func TestBoolRefine(t *testing.T) {
	t.Run("simple refinement", func(t *testing.T) {
		schema := Bool().Refine(func(val bool) bool {
			return val == true
		})

		// Valid case
		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Invalid case
		_, err = schema.Parse(false)
		assert.Error(t, err)
	})

	t.Run("complex refinement", func(t *testing.T) {
		schema := Bool().Refine(func(val bool) bool {
			// Complex logic: only allow true on certain conditions
			return val == true
		})

		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})
}

// =============================================================================
// 8. Error handling
// =============================================================================

func TestBoolErrorHandling(t *testing.T) {
	t.Run("type error", func(t *testing.T) {
		schema := Bool()
		_, err := schema.Parse("not a boolean")
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		assert.NotEmpty(t, zodErr.Issues)
	})

	t.Run("refinement error", func(t *testing.T) {
		schema := Bool().Refine(func(val bool) bool {
			return val == true
		})

		_, err := schema.Parse(false)
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
	})

	t.Run("coercion error", func(t *testing.T) {
		schema := CoercedBool()
		_, err := schema.Parse("invalid")
		assert.Error(t, err)
	})
}

// =============================================================================
// 9. Edge and mutual exclusion cases
// =============================================================================

func TestBoolEdgeCases(t *testing.T) {
	t.Run("interface with bool value", func(t *testing.T) {
		schema := Bool()
		var input any = true

		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("complex type rejection", func(t *testing.T) {
		schema := Bool()

		complexTypes := []any{
			make(chan bool),
			func() bool { return true },
			struct{ B bool }{B: true},
			[]any{true},
			map[any]any{"key": true},
		}

		for _, input := range complexTypes {
			_, err := schema.Parse(input)
			assert.Error(t, err)
		}
	})

	t.Run("coercion edge cases", func(t *testing.T) {
		schema := CoercedBool()

		tests := []struct {
			name     string
			input    string
			expected bool
		}{
			{"whitespace_only", "   ", false},
			{"tab_and_newline", "\t\n", false},
			{"mixed_case_true", "TrUe", true},
			{"mixed_case_false", "FaLsE", false},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				result, err := schema.Parse(test.input)
				require.NoError(t, err)
				assert.Equal(t, test.expected, result)
			})
		}
	})
}

// =============================================================================
// 10. Default and Prefault tests
// =============================================================================

func TestBoolDefaultAndPrefault(t *testing.T) {
	t.Run("default value", func(t *testing.T) {
		schema := Bool().Default(true)

		// Valid boolean should not use default
		result, err := schema.Parse(false)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("defaultFunc", func(t *testing.T) {
		defaultFn := func() bool {
			return true
		}
		schema := Bool().DefaultFunc(defaultFn)

		require.NotNil(t, schema)
		// Note: DefaultFunc is not tested here as it's used internally
	})

	t.Run("defaultFunc complete test", func(t *testing.T) {
		counter := 0
		schema := Bool().DefaultFunc(func() bool {
			counter++
			return counter%2 == 1 // Alternates between true and false
		})

		// nil input should call function and use default
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, true, result1)
		assert.Equal(t, 1, counter, "Function should be called once for nil input")

		// Another nil input should call function again
		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		assert.Equal(t, false, result2)
		assert.Equal(t, 2, counter, "Function should be called twice for second nil input")

		// Valid input should not call function
		result3, err3 := schema.Parse(true)
		require.NoError(t, err3)
		assert.Equal(t, true, result3)
		assert.Equal(t, 2, counter, "Function should not be called for valid input")
	})

	t.Run("prefault", func(t *testing.T) {
		schema := Bool().Refine(func(val bool) bool {
			return val == true
		}).Prefault(true)

		// Valid case
		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Invalid case should use prefault
		result, err = schema.Parse(false)
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("prefaultFunc", func(t *testing.T) {
		counter := 0
		schema := Bool().Refine(func(val bool) bool {
			return val == true // Only allow true values
		}).PrefaultFunc(func() bool {
			counter++
			return true // Always return true as fallback
		})

		// Valid input should not call prefault function
		result1, err1 := schema.Parse(true)
		require.NoError(t, err1)
		assert.Equal(t, true, result1)
		assert.Equal(t, 0, counter, "Function should not be called for valid input")

		// Invalid input should call prefault function
		result2, err2 := schema.Parse(false)
		require.NoError(t, err2)
		assert.Equal(t, true, result2)
		assert.Equal(t, 1, counter, "Function should be called once for invalid input")

		// Another invalid input should call function again
		result3, err3 := schema.Parse(false)
		require.NoError(t, err3)
		assert.Equal(t, true, result3)
		assert.Equal(t, 2, counter, "Function should increment counter for each invalid input")

		// Valid input still doesn't call function
		result4, err4 := schema.Parse(true)
		require.NoError(t, err4)
		assert.Equal(t, true, result4)
		assert.Equal(t, 2, counter, "Counter should remain unchanged for valid input")
	})

	t.Run("default vs prefault distinction", func(t *testing.T) {
		schema := Bool().
			Refine(func(val bool) bool {
				return val == true // Only allow true values
			}).
			Default(false).
			Prefault(true)

		// nil input uses default
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, false, result1)

		// Valid input succeeds
		result2, err2 := schema.Parse(true)
		require.NoError(t, err2)
		assert.Equal(t, true, result2)

		// Invalid input uses prefault
		result3, err3 := schema.Parse(false)
		require.NoError(t, err3)
		assert.Equal(t, true, result3)
	})
}
