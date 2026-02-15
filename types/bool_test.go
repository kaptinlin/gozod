package types

import (
	"fmt"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Basic functionality tests
// =============================================================================

func TestBool_BasicFunctionality(t *testing.T) {
	t.Run("valid boolean inputs", func(t *testing.T) {
		schema := Bool()

		for _, input := range []bool{true, false} {
			t.Run(fmt.Sprintf("%t", input), func(t *testing.T) {
				result, err := schema.Parse(input)
				require.NoError(t, err)
				assert.Equal(t, input, result)
			})
		}
	})

	t.Run("invalid type inputs", func(t *testing.T) {
		schema := Bool()

		for _, input := range []any{
			"not a boolean", 123, 3.14, []bool{true}, nil,
		} {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})

	t.Run("Parse and MustParse methods", func(t *testing.T) {
		schema := Bool()

		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		mustResult := schema.MustParse(false)
		assert.Equal(t, false, mustResult)

		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected a boolean value"
		schema := Bool(core.SchemaParams{Error: customError})

		require.NotNil(t, schema)
		assert.Equal(t, core.ZodTypeBool, schema.internals.Def.Type)

		_, err := schema.Parse("invalid")
		assert.Error(t, err)
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestBool_TypeSafety(t *testing.T) {
	t.Run("Bool returns bool type", func(t *testing.T) {
		schema := Bool()

		for _, input := range []bool{true, false} {
			result, err := schema.Parse(input)
			require.NoError(t, err)
			assert.Equal(t, input, result)
			assert.IsType(t, bool(false), result)
		}
	})

	t.Run("BoolPtr returns *bool type", func(t *testing.T) {
		schema := BoolPtr()

		for _, input := range []bool{true, false} {
			result, err := schema.Parse(input)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, input, *result)
			assert.IsType(t, (*bool)(nil), result)
		}
	})

	t.Run("type inference with assignment", func(t *testing.T) {
		boolSchema := Bool()
		ptrSchema := BoolPtr()

		result1, err1 := boolSchema.Parse(true)
		require.NoError(t, err1)
		assert.IsType(t, bool(false), result1)
		assert.Equal(t, true, result1)

		result2, err2 := ptrSchema.Parse(true)
		require.NoError(t, err2)
		assert.IsType(t, (*bool)(nil), result2)
		require.NotNil(t, result2)
		assert.Equal(t, true, *result2)
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		boolSchema := Bool()
		result := boolSchema.MustParse(true)
		assert.IsType(t, bool(false), result)
		assert.Equal(t, true, result)

		ptrSchema := BoolPtr()
		ptrResult := ptrSchema.MustParse(true)
		assert.IsType(t, (*bool)(nil), ptrResult)
		require.NotNil(t, ptrResult)
		assert.Equal(t, true, *ptrResult)
	})

	t.Run("generic type constraint verification", func(t *testing.T) {
		// Compile-time type checking
		var _ = Bool()
		var _ = BoolPtr()
		var _ = Bool().Optional()
		var _ = Bool().Default(true)

		boolSchema := Bool()
		result, err := boolSchema.Parse(true)
		require.NoError(t, err)
		assert.IsType(t, bool(false), result)

		ptrSchema := BoolPtr()
		ptrResult, err := ptrSchema.Parse(true)
		require.NoError(t, err)
		assert.IsType(t, (*bool)(nil), ptrResult)
	})
}

// =============================================================================
// Modifier methods tests
// =============================================================================

func TestBool_Modifiers(t *testing.T) {
	t.Run("Optional always returns *bool", func(t *testing.T) {
		// From bool to *bool via Optional
		boolSchema := Bool()
		optionalSchema := boolSchema.Optional()

		// Type check: ensure it returns *ZodBool[*bool]
		var _ = optionalSchema

		// Functionality test
		result, err := optionalSchema.Parse(true)
		require.NoError(t, err)
		assert.IsType(t, (*bool)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)

		// From *bool to *bool via Optional (maintains type)
		ptrSchema := BoolPtr()
		optionalPtrSchema := ptrSchema.Optional()
		var _ = optionalPtrSchema
	})

	t.Run("Nilable always returns *bool", func(t *testing.T) {
		boolSchema := Bool()
		nilableSchema := boolSchema.Nilable()

		var _ = nilableSchema

		// Test nil handling
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nullish combines optional and nilable", func(t *testing.T) {
		boolSchema := Bool()
		nullishSchema := boolSchema.Nullish()

		var _ = nullishSchema

		// Test nil handling
		result, err := nullishSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid value
		result, err = nullishSchema.Parse(true)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)
	})

	t.Run("Default preserves current type", func(t *testing.T) {
		// bool maintains bool
		boolSchema := Bool()
		defaultBoolSchema := boolSchema.Default(true)
		var _ = defaultBoolSchema

		// *bool maintains *bool
		ptrSchema := BoolPtr()
		defaultPtrSchema := ptrSchema.Default(false)
		var _ = defaultPtrSchema

		// Test behavior
		result, err := defaultBoolSchema.Parse(false)
		require.NoError(t, err)
		assert.IsType(t, bool(false), result)
		assert.Equal(t, false, result)
	})

	t.Run("Prefault preserves current type", func(t *testing.T) {
		// bool maintains bool
		boolSchema := Bool()
		prefaultBoolSchema := boolSchema.Prefault(true)
		var _ = prefaultBoolSchema

		// *bool maintains *bool
		ptrSchema := BoolPtr()
		prefaultPtrSchema := ptrSchema.Prefault(false)
		var _ = prefaultPtrSchema

		// Test behavior
		result, err := prefaultBoolSchema.Parse(false)
		require.NoError(t, err)
		assert.IsType(t, bool(false), result)
		assert.Equal(t, false, result)
	})

	t.Run("modifier immutability", func(t *testing.T) {
		originalSchema := Bool()
		modifiedSchema := originalSchema.Optional()

		// Original should not be affected by modifier
		_, err1 := originalSchema.Parse(nil)
		assert.Error(t, err1, "Original schema should reject nil")

		// Modified schema should have new behavior
		result2, err2 := modifiedSchema.Parse(nil)
		require.NoError(t, err2)
		assert.Nil(t, result2)
	})

	t.Run("Default and Prefault functionality", func(t *testing.T) {
		t.Run("default value with Bool", func(t *testing.T) {
			schema := Bool().Default(true)

			// Valid input should override default
			result, err := schema.Parse(false)
			require.NoError(t, err)
			assert.Equal(t, false, result)
			assert.IsType(t, bool(false), result)

			// Nil input should use default
			result, err = schema.Parse(nil)
			require.NoError(t, err)
			assert.Equal(t, true, result)
		})

		t.Run("default value with BoolPtr", func(t *testing.T) {
			schema := BoolPtr().Default(true)

			// Valid input should override default
			result, err := schema.Parse(false)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, false, *result)

			// Nil input should use default
			result, err = schema.Parse(nil)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, true, *result)
		})

		// Test Default priority over Prefault
		t.Run("Default priority over Prefault", func(t *testing.T) {
			schema := Bool().Default(true).Prefault(false)

			// Nil input should use Default (higher priority), not Prefault
			result, err := schema.Parse(nil)
			require.NoError(t, err)
			assert.Equal(t, true, result) // Default value, not Prefault value
		})

		// Test Default short-circuit bypasses validation
		t.Run("Default short-circuit bypasses validation", func(t *testing.T) {
			// Create a schema with validation that would reject the default value
			schema := Bool().Refine(func(b bool) bool {
				return b == false // Only allow false values
			}, "Must be false").Default(true) // Default bypasses validation

			// Nil input should use default and bypass validation
			result, err := schema.Parse(nil)
			require.NoError(t, err)
			assert.Equal(t, true, result)

			// Valid input should still go through validation
			result, err = schema.Parse(false)
			require.NoError(t, err)
			assert.Equal(t, false, result)

			// Invalid input should fail validation
			_, err = schema.Parse(true)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "Must be false")
		})

		t.Run("prefault value", func(t *testing.T) {
			schema := Bool().Prefault(true)

			// Valid input should override prefault
			result, err := schema.Parse(false)
			require.NoError(t, err)
			assert.Equal(t, false, result)
			assert.IsType(t, bool(false), result)

			// Nil input should use prefault
			result, err = schema.Parse(nil)
			require.NoError(t, err)
			assert.Equal(t, true, result)
		})

		// Test Prefault only triggers on nil input
		t.Run("Prefault only triggers on nil input", func(t *testing.T) {
			schema := Bool().Refine(func(b bool) bool {
				return b == false // Only allow false values
			}, "Must be false").Prefault(false)

			// Valid input should pass through
			result, err := schema.Parse(false)
			require.NoError(t, err)
			assert.Equal(t, false, result)

			// Invalid non-nil input should error (NOT use prefault)
			_, err = schema.Parse(true)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "Must be false")

			// Nil input should trigger prefault
			result, err = schema.Parse(nil)
			require.NoError(t, err)
			assert.Equal(t, false, result)
		})

		// Test Prefault goes through full validation
		t.Run("Prefault goes through full validation", func(t *testing.T) {
			schema := Bool().Refine(func(b bool) bool {
				return b == false // Only allow false values
			}, "Must be false").Prefault(true) // This should fail validation

			// Nil input should trigger prefault, but prefault should fail validation
			_, err := schema.Parse(nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "Must be false")
		})

		t.Run("DefaultFunc and PrefaultFunc", func(t *testing.T) {
			defaultCalled := false
			prefaultCalled := false
			schema := Bool().
				DefaultFunc(func() bool {
					defaultCalled = true
					return true
				}).
				PrefaultFunc(func() bool {
					prefaultCalled = true
					return false
				})

			// Valid input should not call any function
			result, err := schema.Parse(true)
			require.NoError(t, err)
			assert.Equal(t, true, result)
			assert.False(t, defaultCalled)
			assert.False(t, prefaultCalled)

			// Nil input should call DefaultFunc (higher priority)
			result, err = schema.Parse(nil)
			require.NoError(t, err)
			assert.Equal(t, true, result)
			assert.True(t, defaultCalled)
			assert.False(t, prefaultCalled) // Should not be called due to Default priority
		})

		// Test Transform interactions
		t.Run("Default bypasses Transform (short-circuit)", func(t *testing.T) {
			transformCalled := false
			schema := Bool().Default(true).Transform(func(b bool, ctx *core.RefinementContext) (any, error) {
				transformCalled = true
				return !b, nil // Invert the boolean
			})

			// Valid input should go through transform
			result, err := schema.Parse(false)
			require.NoError(t, err)
			assert.Equal(t, true, result) // false inverted to true
			assert.True(t, transformCalled)

			// Reset flag
			transformCalled = false

			// Nil input should use default and bypass transform
			result, err = schema.Parse(nil)
			require.NoError(t, err)
			assert.Equal(t, true, result)    // Default value, not transformed
			assert.False(t, transformCalled) // Transform should not be called
		})

		t.Run("Prefault goes through Transform (full pipeline)", func(t *testing.T) {
			transformCalled := false
			schema := Bool().Prefault(false).Transform(func(b bool, ctx *core.RefinementContext) (any, error) {
				transformCalled = true
				return !b, nil // Invert the boolean
			})

			// Valid input should go through transform
			result, err := schema.Parse(true)
			require.NoError(t, err)
			assert.Equal(t, false, result) // true inverted to false
			assert.True(t, transformCalled)

			// Reset flag
			transformCalled = false

			// Nil input should use prefault and go through transform
			result, err = schema.Parse(nil)
			require.NoError(t, err)
			assert.Equal(t, true, result)   // false (prefault) inverted to true
			assert.True(t, transformCalled) // Transform should be called
		})

		// Test Prefault error handling
		t.Run("Prefault error handling", func(t *testing.T) {
			schema := Bool().PrefaultFunc(func() bool {
				return true // This should pass Bool validation
			})

			// Nil input should use prefault function
			result, err := schema.Parse(nil)
			require.NoError(t, err)
			assert.Equal(t, true, result)
		})

		// Test BoolPtr behavior
		t.Run("BoolPtr Prefault only on nil input", func(t *testing.T) {
			schema := BoolPtr().Refine(func(b *bool) bool {
				return b != nil && *b == false // Only allow false values
			}, "Must be false").Prefault(false)

			// Valid input should pass through
			falseVal := false
			result, err := schema.Parse(&falseVal)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, false, *result)

			// Invalid non-nil input should error (NOT use prefault)
			trueVal := true
			_, err = schema.Parse(&trueVal)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "Must be false")

			// Nil input should trigger prefault
			result, err = schema.Parse(nil)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, false, *result)
		})

		t.Run("chaining with other modifiers", func(t *testing.T) {
			schema := Bool().Default(true).Optional()

			// Test type evolution
			var _ = schema

			result, err := schema.Parse(false)
			require.NoError(t, err)
			assert.IsType(t, (*bool)(nil), result)
			require.NotNil(t, result)
			assert.Equal(t, false, *result)
		})
	})

	t.Run("Refine and RefineAny", func(t *testing.T) {
		t.Run("refine validate", func(t *testing.T) {
			// Only accept true values
			schema := Bool().Refine(func(b bool) bool {
				return b == true
			})

			result, err := schema.Parse(true)
			require.NoError(t, err)
			assert.Equal(t, true, result)

			_, err = schema.Parse(false)
			assert.Error(t, err)
		})

		t.Run("refine with custom error message", func(t *testing.T) {
			errorMessage := "Must be true"
			schema := BoolPtr().Refine(func(b *bool) bool {
				return b != nil && *b == true
			}, core.SchemaParams{Error: errorMessage})

			result, err := schema.Parse(true)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, true, *result)

			_, err = schema.Parse(false)
			assert.Error(t, err)
		})

		t.Run("always failing refine", func(t *testing.T) {
			schema := Bool().Refine(func(b bool) bool {
				return false // Always fail
			})

			_, err := schema.Parse(false)
			assert.Error(t, err)
		})

		t.Run("refine pointer allows nil", func(t *testing.T) {
			schema := BoolPtr().Nilable().Refine(func(b *bool) bool {
				// Accept nil or true
				return b == nil || *b
			})

			// Expect nil to be accepted
			result, err := schema.Parse(nil)
			require.NoError(t, err)
			assert.Nil(t, result)

			// true should pass
			result, err = schema.Parse(true)
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, true, *result)

			// false should fail (refine returns false)
			_, err = schema.Parse(false)
			assert.Error(t, err)
		})

		t.Run("refine pointer rejects nil when not nilable", func(t *testing.T) {
			schema := BoolPtr().Refine(func(b *bool) bool {
				return b != nil && *b // Require non-nil and true
			})

			// nil should error because schema not nilable and Refine expects non-nil
			_, err := schema.Parse(nil)
			assert.Error(t, err)

			// true passes
			result, err := schema.Parse(true)
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, *result)
		})

		t.Run("type-safe refine with automatic conversion", func(t *testing.T) {
			// Test that Refine automatically converts values to the correct type T
			boolSchema := Bool().Refine(func(b bool) bool {
				return b // Only accept true values (false fails)
			})

			ptrSchema := BoolPtr().Refine(func(b *bool) bool {
				return b != nil && *b // Only accept non-nil true pointers
			})

			// Test bool schema
			result1, err1 := boolSchema.Parse(true)
			require.NoError(t, err1)
			assert.Equal(t, true, result1)

			_, err1 = boolSchema.Parse(false)
			assert.Error(t, err1)

			// Test pointer schema
			result2, err2 := ptrSchema.Parse(true)
			require.NoError(t, err2)
			assert.NotNil(t, result2)
			assert.Equal(t, true, *result2)

			_, err2 = ptrSchema.Parse(false)
			assert.Error(t, err2)
		})

		t.Run("refineAny bool schema", func(t *testing.T) {
			// Only accept true values via RefineAny on Bool() schema
			schema := Bool().RefineAny(func(v any) bool {
				b, ok := v.(bool)
				return ok && b
			})

			// true passes
			result, err := schema.Parse(true)
			require.NoError(t, err)
			assert.Equal(t, true, result)

			// false fails
			_, err = schema.Parse(false)
			assert.Error(t, err)
		})

		t.Run("refineAny pointer schema", func(t *testing.T) {
			// BoolPtr().RefineAny sees underlying bool value
			schema := BoolPtr().RefineAny(func(v any) bool {
				b, ok := v.(bool)
				return ok && b // accept only true
			})

			result, err := schema.Parse(true)
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, true, *result)

			_, err = schema.Parse(false)
			assert.Error(t, err)
		})

		t.Run("refineAny pointer schema nilable", func(t *testing.T) {
			// Nil input should bypass checks and be accepted because schema is Nilable()
			schema := BoolPtr().Nilable().RefineAny(func(v any) bool {
				// Never called for nil input, but keep return true for completeness
				return true
			})

			// nil passes
			result, err := schema.Parse(nil)
			require.NoError(t, err)
			assert.Nil(t, result)

			// true still passes
			result, err = schema.Parse(true)
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, *result)
		})

		t.Run("refineAny vs refine comparison", func(t *testing.T) {
			// RefineAny: Raw access, manual type handling
			refineAnySchema := Bool().RefineAny(func(v any) bool {
				if b, ok := v.(bool); ok {
					return b == true
				}
				return false
			})

			// Refine: Type-safe, automatic conversion
			refineSchema := Bool().Refine(func(b bool) bool {
				return b == true
			})

			// Both should behave the same for valid inputs
			result1, err1 := refineAnySchema.Parse(true)
			result2, err2 := refineSchema.Parse(true)

			require.NoError(t, err1)
			require.NoError(t, err2)
			assert.Equal(t, result1, result2)

			// Both should fail for false
			_, err1 = refineAnySchema.Parse(false)
			_, err2 = refineSchema.Parse(false)

			assert.Error(t, err1)
			assert.Error(t, err2)
		})
	})

	t.Run("Overwrite", func(t *testing.T) {
		t.Run("basic overwrite functionality", func(t *testing.T) {
			schema := Bool().Overwrite(func(b bool) bool { return !b })
			// true -> false
			res, err := schema.Parse(true)
			require.NoError(t, err)
			assert.False(t, res)
			// false -> true
			res, err = schema.Parse(false)
			require.NoError(t, err)
			assert.True(t, res)
		})

		t.Run("overwrite preserves type", func(t *testing.T) {
			schema := Bool().Overwrite(func(b bool) bool { return !b })
			res, err := schema.Parse(true)
			require.NoError(t, err)
			assert.IsType(t, bool(false), res)
		})

		t.Run("overwrite chaining with validation", func(t *testing.T) {
			schema := Bool().
				Overwrite(func(b bool) bool { return !b }).
				Refine(func(b bool) bool { return b }, "must be true after inversion")
			_, err := schema.Parse(true) // -> false, should fail
			assert.Error(t, err)
			res, err := schema.Parse(false) // -> true, should pass
			require.NoError(t, err)
			assert.True(t, res)
		})

		t.Run("multiple overwrite calls", func(t *testing.T) {
			schema := Bool().
				Overwrite(func(b bool) bool { return !b }).
				Overwrite(func(b bool) bool { return !b }) // invert twice
			res, err := schema.Parse(true)
			require.NoError(t, err)
			assert.True(t, res)
		})

		t.Run("overwrite strict type checking", func(t *testing.T) {
			schema := Bool().Overwrite(func(b bool) bool { return !b })
			_, err := schema.Parse(1) // int input
			assert.Error(t, err)
			_, err = schema.Parse("true") // string input
			assert.Error(t, err)
		})
	})

	t.Run("Check", func(t *testing.T) {
		t.Run("adds multiple issues for invalid input", func(t *testing.T) {
			schema := Bool().Check(func(value bool, p *core.ParsePayload) {
				if !value {
					p.AddIssueWithMessage("must be true")
				}
			})
			_, err := schema.Parse(false)
			require.Error(t, err)
			var zErr *issues.ZodError
			assert.True(t, issues.IsZodError(err, &zErr))
			assert.Len(t, zErr.Issues, 1)
		})

		t.Run("succeeds for valid input", func(t *testing.T) {
			schema := Bool().Check(func(value bool, p *core.ParsePayload) {})
			res, err := schema.Parse(true)
			require.NoError(t, err)
			assert.True(t, res)
		})

		t.Run("works with pointer types", func(t *testing.T) {
			schema := BoolPtr().Check(func(value *bool, p *core.ParsePayload) {
				if value != nil && !*value {
					p.AddIssueWithMessage("pointer must be true")
				}
			})
			f := false
			_, err := schema.Parse(&f)
			assert.Error(t, err)
			tVal := true
			res, err := schema.Parse(&tVal)
			require.NoError(t, err)
			assert.True(t, *res)
		})
	})

	t.Run("NonOptional", func(t *testing.T) {
		// --- Basic NonOptional behaviour ---
		schema := Bool().NonOptional()

		// Valid bool input
		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)
		assert.IsType(t, bool(false), result)

		// Invalid nil input
		_, err = schema.Parse(nil)
		assert.Error(t, err)
		var zErr *issues.ZodError
		if ok := issues.IsZodError(err, &zErr); ok {
			require.Greater(t, len(zErr.Issues), 0)
			assert.Equal(t, core.InvalidType, zErr.Issues[0].Code)
			assert.Equal(t, core.ZodTypeNonOptional, zErr.Issues[0].Expected)
		}

		// --- Optional().NonOptional() chain ---
		chained := Bool().Optional().NonOptional()
		var _ = chained // compile-time type assertion

		res, err := chained.Parse(false)
		require.NoError(t, err)
		assert.Equal(t, false, res)

		_, err = chained.Parse(nil)
		assert.Error(t, err)

		// --- Object embedding ---
		objSchema := Object(map[string]core.ZodSchema{
			"ok": Bool().Optional().NonOptional(),
		})

		// Valid case
		_, err = objSchema.Parse(map[string]any{"ok": true})
		require.NoError(t, err)

		// Nil field should error
		_, err = objSchema.Parse(map[string]any{"ok": nil})
		assert.Error(t, err)

		// Missing key should error as nonoptional (invalid_type)
		_, err = objSchema.Parse(map[string]any{})
		assert.Error(t, err)

		// --- BoolPtr().NonOptional() ---
		val := true
		ptrSchema := BoolPtr().NonOptional()

		res2, err := ptrSchema.Parse(&val)
		require.NoError(t, err)
		assert.IsType(t, bool(false), res2)
		assert.Equal(t, true, res2)

		res3, err := ptrSchema.Parse(false)
		require.NoError(t, err)
		assert.Equal(t, false, res3)

		_, err = ptrSchema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("IsOptional and IsNilable", func(t *testing.T) {
		t.Run("basic schema - not optional, not nilable", func(t *testing.T) {
			schema := Bool()

			assert.False(t, schema.IsOptional(), "Basic bool schema should not be optional")
			assert.False(t, schema.IsNilable(), "Basic bool schema should not be nilable")
		})

		t.Run("optional schema - is optional, not nilable", func(t *testing.T) {
			schema := Bool().Optional()

			assert.True(t, schema.IsOptional(), "Optional bool schema should be optional")
			assert.False(t, schema.IsNilable(), "Optional bool schema should not be nilable")
		})

		t.Run("nilable schema - not optional, is nilable", func(t *testing.T) {
			schema := Bool().Nilable()

			assert.False(t, schema.IsOptional(), "Nilable bool schema should not be optional")
			assert.True(t, schema.IsNilable(), "Nilable bool schema should be nilable")
		})

		t.Run("nullish schema - is optional and nilable", func(t *testing.T) {
			schema := Bool().Nullish()

			assert.True(t, schema.IsOptional(), "Nullish bool schema should be optional")
			assert.True(t, schema.IsNilable(), "Nullish bool schema should be nilable")
		})

		t.Run("chained modifiers", func(t *testing.T) {
			// Optional then Nilable
			schema1 := Bool().Optional().Nilable()
			assert.True(t, schema1.IsOptional(), "Optional().Nilable() should be optional")
			assert.True(t, schema1.IsNilable(), "Optional().Nilable() should be nilable")

			// Nilable then Optional
			schema2 := Bool().Nilable().Optional()
			assert.True(t, schema2.IsOptional(), "Nilable().Optional() should be optional")
			assert.True(t, schema2.IsNilable(), "Nilable().Optional() should be nilable")
		})

		t.Run("nonoptional modifier resets optional flag", func(t *testing.T) {
			schema := Bool().Optional().NonOptional()

			assert.False(t, schema.IsOptional(), "Optional().NonOptional() should not be optional")
			assert.False(t, schema.IsNilable(), "Optional().NonOptional() should not be nilable")
		})

		t.Run("pointer types", func(t *testing.T) {
			// BoolPtr basic
			ptrSchema := BoolPtr()
			assert.False(t, ptrSchema.IsOptional(), "BoolPtr schema should not be optional")
			assert.False(t, ptrSchema.IsNilable(), "BoolPtr schema should not be nilable")

			// BoolPtr with modifiers
			optionalPtrSchema := BoolPtr().Optional()
			assert.True(t, optionalPtrSchema.IsOptional(), "BoolPtr().Optional() should be optional")
			assert.False(t, optionalPtrSchema.IsNilable(), "BoolPtr().Optional() should not be nilable")

			nilablePtrSchema := BoolPtr().Nilable()
			assert.False(t, nilablePtrSchema.IsOptional(), "BoolPtr().Nilable() should not be optional")
			assert.True(t, nilablePtrSchema.IsNilable(), "BoolPtr().Nilable() should be nilable")
		})

		t.Run("consistency with Internals", func(t *testing.T) {
			// Test basic bool schema
			basicSchema := Bool()
			assert.Equal(t, basicSchema.Internals().IsOptional(), basicSchema.IsOptional(),
				"Basic schema: IsOptional() should match Internals().IsOptional()")
			assert.Equal(t, basicSchema.Internals().IsNilable(), basicSchema.IsNilable(),
				"Basic schema: IsNilable() should match Internals().IsNilable()")

			// Test optional bool schema
			optionalSchema := Bool().Optional()
			assert.Equal(t, optionalSchema.Internals().IsOptional(), optionalSchema.IsOptional(),
				"Optional schema: IsOptional() should match Internals().IsOptional()")
			assert.Equal(t, optionalSchema.Internals().IsNilable(), optionalSchema.IsNilable(),
				"Optional schema: IsNilable() should match Internals().IsNilable()")

			// Test nilable bool schema
			nilableSchema := Bool().Nilable()
			assert.Equal(t, nilableSchema.Internals().IsOptional(), nilableSchema.IsOptional(),
				"Nilable schema: IsOptional() should match Internals().IsOptional()")
			assert.Equal(t, nilableSchema.Internals().IsNilable(), nilableSchema.IsNilable(),
				"Nilable schema: IsNilable() should match Internals().IsNilable()")

			// Test nullish bool schema
			nullishSchema := Bool().Nullish()
			assert.Equal(t, nullishSchema.Internals().IsOptional(), nullishSchema.IsOptional(),
				"Nullish schema: IsOptional() should match Internals().IsOptional()")
			assert.Equal(t, nullishSchema.Internals().IsNilable(), nullishSchema.IsNilable(),
				"Nullish schema: IsNilable() should match Internals().IsNilable()")

			// Test nonoptional bool schema
			nonoptionalSchema := Bool().Optional().NonOptional()
			assert.Equal(t, nonoptionalSchema.Internals().IsOptional(), nonoptionalSchema.IsOptional(),
				"NonOptional schema: IsOptional() should match Internals().IsOptional()")
			assert.Equal(t, nonoptionalSchema.Internals().IsNilable(), nonoptionalSchema.IsNilable(),
				"NonOptional schema: IsNilable() should match Internals().IsNilable()")
		})
	})
}

// =============================================================================
// Chaining tests
// =============================================================================

func TestBool_Chaining(t *testing.T) {
	t.Run("type evolution through chaining", func(t *testing.T) {
		// Chain with type evolution: bool → bool → *bool
		schema := Bool(). // *ZodBool[bool]
					Default(false). // *ZodBool[bool] (maintains type)
					Optional()      // *ZodBool[*bool] (type conversion)

		var _ = schema

		// Test final behavior
		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.IsType(t, (*bool)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)
	})

	t.Run("complex chaining", func(t *testing.T) {
		schema := BoolPtr(). // *ZodBool[*bool]
					Nilable().    // *ZodBool[*bool] (maintains type)
					Default(true) // *ZodBool[*bool] (maintains type)

		var _ = schema

		result, err := schema.Parse(false)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, false, *result)
	})

	t.Run("default and prefault chaining", func(t *testing.T) {
		schema := Bool().
			Default(true).
			Prefault(false)

		var _ = schema

		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("order independence verification", func(t *testing.T) {
		// Different chaining orders should produce equivalent results
		schema1 := Bool().Default(true).Optional()
		schema2 := Bool().Optional().Default(true)

		var _ = schema1
		var _ = schema2

		// Both should handle the same inputs similarly
		result1, err1 := schema1.Parse(false)
		result2, err2 := schema2.Parse(false)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.IsType(t, (*bool)(nil), result1)
		assert.IsType(t, (*bool)(nil), result2)
	})
}

// =============================================================================
// Edge case & pointer identity tests
// =============================================================================

func TestBool_EdgeCases(t *testing.T) {
	t.Run("nil handling with *bool", func(t *testing.T) {
		schema := BoolPtr().Nilable()

		// Test nil input
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid boolean
		result, err = schema.Parse(true)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)
	})

	t.Run("empty context", func(t *testing.T) {
		schema := Bool()

		// Parse with empty context slice
		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("performance critical paths", func(t *testing.T) {
		schema := Bool()

		// Test that fast paths work correctly
		t.Run("direct bool input fast path", func(t *testing.T) {
			result, err := schema.Parse(true)
			require.NoError(t, err)
			assert.Equal(t, true, result)
		})

		t.Run("false value fast path", func(t *testing.T) {
			result, err := schema.Parse(false)
			require.NoError(t, err)
			assert.Equal(t, false, result)
		})
	})

	t.Run("API compatibility patterns", func(t *testing.T) {
		// Test that the API matches expected TypeScript Zod patterns

		// Basic usage patterns
		schema1 := Bool()               // z.boolean()
		schema2 := Bool().Optional()    // z.boolean().optional()
		schema3 := Bool().Default(true) // z.boolean().default(true)
		schema4 := CoercedBool()        // z.coerce.boolean()

		// Verify these work as expected
		result, err := schema1.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		result2, err := schema2.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result2)

		result3, err := schema3.Parse(false)
		require.NoError(t, err)
		assert.Equal(t, false, result3)

		result4, err := schema4.Parse("true")
		require.NoError(t, err)
		assert.Equal(t, true, result4)
	})

	t.Run("memory efficiency verification", func(t *testing.T) {
		// Create multiple schemas to verify shared state
		schema1 := Bool()
		schema2 := schema1.Default(true)
		schema3 := schema2.Optional()

		// All should work independently
		result1, err1 := schema1.Parse(false)
		require.NoError(t, err1)
		assert.Equal(t, false, result1)

		result2, err2 := schema2.Parse(false)
		require.NoError(t, err2)
		assert.Equal(t, false, result2)

		result3, err3 := schema3.Parse(false)
		require.NoError(t, err3)
		assert.NotNil(t, result3)
		assert.Equal(t, false, *result3)
	})

	t.Run("transform and pipe integration", func(t *testing.T) {
		schema := Bool()

		transform := schema.Transform(func(b bool, ctx *core.RefinementContext) (any, error) {
			if b {
				return "yes", nil
			}
			return "no", nil
		})

		result, err := transform.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, "yes", result)

		// Test extractBool helper function
		boolVal := true
		ptrVal := &boolVal

		extracted1 := extractBool(boolVal)
		assert.Equal(t, true, extracted1)

		extracted2 := extractBool(ptrVal)
		assert.Equal(t, true, extracted2)

		nilPtr := (*bool)(nil)
		extracted3 := extractBool(nilPtr)
		assert.Equal(t, false, extracted3)
	})

	t.Run("pointer identity preservation", func(t *testing.T) {
		t.Run("Bool Optional preserves pointer identity", func(t *testing.T) {
			schema := Bool().Optional()

			originalBool := true
			originalPtr := &originalBool

			result, err := schema.Parse(originalPtr)
			require.NoError(t, err)

			// Result should be the same pointer
			assert.True(t, result == originalPtr, "Pointer identity should be preserved")
			assert.Equal(t, true, *result)
		})

		t.Run("Bool Nilable preserves pointer identity", func(t *testing.T) {
			schema := Bool().Nilable()

			originalBool := false
			originalPtr := &originalBool

			result, err := schema.Parse(originalPtr)
			require.NoError(t, err)

			// Result should be the same pointer
			assert.True(t, result == originalPtr, "Pointer identity should be preserved")
			assert.Equal(t, false, *result)
		})

		t.Run("BoolPtr Optional preserves pointer identity", func(t *testing.T) {
			schema := BoolPtr().Optional()

			originalBool := true
			originalPtr := &originalBool

			result, err := schema.Parse(originalPtr)
			require.NoError(t, err)

			// Result should be the same pointer
			assert.True(t, result == originalPtr, "BoolPtr Optional should preserve pointer identity")
			assert.Equal(t, true, *result)
		})

		t.Run("BoolPtr Nilable preserves pointer identity", func(t *testing.T) {
			schema := BoolPtr().Nilable()

			originalBool := false
			originalPtr := &originalBool

			result, err := schema.Parse(originalPtr)
			require.NoError(t, err)

			// Result should be the same pointer
			assert.True(t, result == originalPtr, "BoolPtr Nilable should preserve pointer identity")
			assert.Equal(t, false, *result)
		})

		t.Run("Bool Nullish preserves pointer identity", func(t *testing.T) {
			schema := Bool().Nullish()

			originalBool := true
			originalPtr := &originalBool

			result, err := schema.Parse(originalPtr)
			require.NoError(t, err)

			// Result should be the same pointer
			assert.True(t, result == originalPtr, "Bool Nullish should preserve pointer identity")
			assert.Equal(t, true, *result)
		})

		t.Run("Optional handles nil consistently", func(t *testing.T) {
			schema := Bool().Optional()

			result, err := schema.Parse(nil)
			require.NoError(t, err)
			assert.Nil(t, result)
		})

		t.Run("Nilable handles nil consistently", func(t *testing.T) {
			schema := Bool().Nilable()

			result, err := schema.Parse(nil)
			require.NoError(t, err)
			assert.Nil(t, result)
		})

		t.Run("Default().Optional() chaining preserves pointer identity", func(t *testing.T) {
			schema := Bool().Default(false).Optional()

			originalBool := true
			originalPtr := &originalBool

			result, err := schema.Parse(originalPtr)
			require.NoError(t, err)

			// Result should be the same pointer
			assert.True(t, result == originalPtr, "Default().Optional() should preserve pointer identity")
			assert.Equal(t, true, *result)
		})

		t.Run("Refine with Optional preserves pointer identity", func(t *testing.T) {
			schema := Bool().Refine(func(b bool) bool {
				return true // Always pass
			}).Optional()

			originalBool := true
			originalPtr := &originalBool

			result, err := schema.Parse(originalPtr)
			require.NoError(t, err)

			// Result should be the same pointer
			assert.True(t, result == originalPtr, "Refine().Optional() should preserve pointer identity")
			assert.Equal(t, true, *result)
		})

		t.Run("Multiple pointer identity tests", func(t *testing.T) {
			schema := Bool().Optional()

			testCases := []bool{true, false}

			for _, boolVal := range testCases {
				t.Run(fmt.Sprintf("bool_%t", boolVal), func(t *testing.T) {
					originalPtr := &boolVal

					result, err := schema.Parse(originalPtr)
					require.NoError(t, err)

					assert.True(t, result == originalPtr, "Pointer identity should be preserved for %t", boolVal)
					assert.Equal(t, boolVal, *result)
				})
			}
		})
	})
}

// =============================================================================
// StrictParse and MustStrictParse tests
// =============================================================================

func TestBool_StrictParse(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		schema := Bool()

		// Test StrictParse with exact type match
		result, err := schema.StrictParse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)
		assert.IsType(t, true, result)

		// Test StrictParse with false value
		falseResult, err := schema.StrictParse(false)
		require.NoError(t, err)
		assert.Equal(t, false, falseResult)
	})

	t.Run("with validation constraints", func(t *testing.T) {
		schema := Bool().Refine(func(b bool) bool {
			return b == true // Only allow true values
		}, "Must be true")

		// Valid case
		result, err := schema.StrictParse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Invalid case - false not allowed
		_, err = schema.StrictParse(false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Must be true")
	})

	t.Run("with pointer types", func(t *testing.T) {
		schema := BoolPtr()
		boolVal := true

		// Test with valid pointer input
		result, err := schema.StrictParse(&boolVal)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)
		assert.IsType(t, (*bool)(nil), result)
	})

	t.Run("with default values", func(t *testing.T) {
		schema := BoolPtr().Default(true)
		var nilPtr *bool = nil

		// Test with nil input (should use default)
		result, err := schema.StrictParse(nilPtr)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)
	})

	t.Run("with prefault values", func(t *testing.T) {
		schema := BoolPtr().Refine(func(b *bool) bool {
			return b != nil && *b == true // Only allow true values
		}, "Must be true").Prefault(false)
		falseVal := false

		// Test with validation failure (should NOT use prefault, should return error)
		_, err := schema.StrictParse(&falseVal)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Must be true")

		// Test with nil input (should use prefault, but fail validation)
		_, err2 := schema.StrictParse(nil)
		require.Error(t, err2)
		assert.Contains(t, err2.Error(), "Must be true") // Prefault fails validation
	})
}

func TestBool_MustStrictParse(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		schema := Bool()

		// Test MustStrictParse with valid input
		result := schema.MustStrictParse(true)
		assert.Equal(t, true, result)
		assert.IsType(t, true, result)

		// Test MustStrictParse with false value
		falseResult := schema.MustStrictParse(false)
		assert.Equal(t, false, falseResult)
	})

	t.Run("panic behavior", func(t *testing.T) {
		schema := Bool().Refine(func(b bool) bool {
			return b == true // Only allow true values
		}, "Must be true")

		// Test panic with validation failure
		assert.Panics(t, func() {
			schema.MustStrictParse(false) // Should panic
		})
	})

	t.Run("with pointer types", func(t *testing.T) {
		schema := BoolPtr()
		boolVal := false

		// Test with valid pointer input
		result := schema.MustStrictParse(&boolVal)
		require.NotNil(t, result)
		assert.Equal(t, false, *result)
		assert.IsType(t, (*bool)(nil), result)
	})

	t.Run("with default values", func(t *testing.T) {
		schema := BoolPtr().Default(false)
		var nilPtr *bool = nil

		// Test with nil input (should use default)
		result := schema.MustStrictParse(nilPtr)
		require.NotNil(t, result)
		assert.Equal(t, false, *result)
	})
}
