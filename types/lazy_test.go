package types

import (
	"sync"
	"testing"
	"time"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Basic functionality tests
// =============================================================================

func TestLazy_BasicFunctionality(t *testing.T) {
	t.Run("basic lazy evaluation", func(t *testing.T) {
		schema := LazyAny(func() any {
			return String().Min(3)
		})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("type-safe lazy evaluation", func(t *testing.T) {
		// New type-safe Lazy API
		schema := Lazy[*ZodString[string]](func() *ZodString[string] {
			return String().Min(3)
		})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("deferred evaluation caching", func(t *testing.T) {
		evaluationCount := 0
		schema := LazyAny(func() any {
			evaluationCount++
			return String().Min(3)
		})

		// Schema created but not evaluated yet
		assert.Equal(t, 0, evaluationCount)

		// First parse triggers evaluation
		_, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, 1, evaluationCount)

		// Second parse uses cached schema
		_, err = schema.Parse("world")
		require.NoError(t, err)
		assert.Equal(t, 1, evaluationCount, "Schema should be cached")
	})

	t.Run("type-safe deferred evaluation caching", func(t *testing.T) {
		evaluationCount := 0
		schema := Lazy[*ZodString[string]](func() *ZodString[string] {
			evaluationCount++
			return String().Min(3)
		})

		// Schema created but not evaluated yet
		assert.Equal(t, 0, evaluationCount)

		// First parse triggers evaluation
		_, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, 1, evaluationCount)

		// Second parse uses cached schema
		_, err = schema.Parse("world")
		require.NoError(t, err)
		assert.Equal(t, 1, evaluationCount, "Schema should be cached")
	})

	t.Run("Parse and MustParse methods", func(t *testing.T) {
		schema := LazyAny(func() any {
			return String()
		})

		// Test Parse method
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Test MustParse method
		mustResult := schema.MustParse("world")
		assert.Equal(t, "world", mustResult)

		// Test panic on invalid input
		assert.Panics(t, func() {
			schema.MustParse(123)
		})
	})

	t.Run("type-safe Parse and MustParse methods", func(t *testing.T) {
		schema := Lazy[*ZodString[string]](func() *ZodString[string] {
			return String()
		})

		// Test Parse method
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Test MustParse method
		mustResult := schema.MustParse("world")
		assert.Equal(t, "world", mustResult)

		// Test panic on invalid input
		assert.Panics(t, func() {
			schema.MustParse(123)
		})
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected a valid lazy value"
		schema := LazyAny(func() any {
			return String()
		}, core.SchemaParams{Error: customError})

		require.NotNil(t, schema)
		assert.Equal(t, core.ZodTypeLazy, schema.internals.Def.Type)

		_, err := schema.Parse(123)
		assert.Error(t, err)
	})

	t.Run("type-safe custom error message", func(t *testing.T) {
		customError := "Expected a valid type-safe lazy value"
		schema := Lazy[*ZodString[string]](func() *ZodString[string] {
			return String()
		}, core.SchemaParams{Error: customError})

		require.NotNil(t, schema)

		_, err := schema.Parse(123)
		assert.Error(t, err)
	})

	t.Run("invalid getter function", func(t *testing.T) {
		schema := LazyAny(func() any {
			return nil // This should cause an error
		})

		_, err := schema.Parse("test")
		assert.Error(t, err)
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestLazy_TypeSafety(t *testing.T) {
	t.Run("LazyAny returns any type", func(t *testing.T) {
		schema := LazyAny(func() any {
			return String()
		})
		require.NotNil(t, schema)

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
		assert.IsType(t, "", result) // Should be string from inner schema
	})

	t.Run("type-safe Lazy with compile-time type checking", func(t *testing.T) {
		// String schema
		stringSchema := Lazy[*ZodString[string]](func() *ZodString[string] {
			return String().Min(3)
		})
		require.NotNil(t, stringSchema)

		result, err := stringSchema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
		assert.IsType(t, "", result)

		// Bool schema
		boolSchema := Lazy[*ZodBool[bool]](func() *ZodBool[bool] {
			return Bool()
		})
		require.NotNil(t, boolSchema)

		result2, err := boolSchema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result2)
		assert.IsType(t, true, result2)
	})

	t.Run("LazyPtr returns *any type", func(t *testing.T) {
		schema := LazyPtr(func() any {
			return String()
		})
		require.NotNil(t, schema)

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("type inference with assignment", func(t *testing.T) {
		// Type-inference friendly API
		lazySchema := LazyAny(func() any {
			return String()
		})
		ptrSchema := LazyPtr(func() any {
			return String()
		})

		// Test any type
		result1, err1 := lazySchema.Parse("hello")
		require.NoError(t, err1)
		assert.IsType(t, "", result1)

		// Test *any type
		result2, err2 := ptrSchema.Parse("hello")
		require.NoError(t, err2)
		assert.NotNil(t, result2)
	})

	t.Run("type-safe inference with assignment", func(t *testing.T) {
		// Type-safe lazy schema
		typedSchema := Lazy[*ZodString[string]](func() *ZodString[string] {
			return String().Min(2)
		})

		result, err := typedSchema.Parse("hello")
		require.NoError(t, err)
		assert.IsType(t, "", result)
		assert.Equal(t, "hello", result)
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		schema := LazyAny(func() any {
			return String()
		})

		result := schema.MustParse("hello")
		assert.IsType(t, "", result)
		assert.Equal(t, "hello", result)
	})

	t.Run("type-safe MustParse", func(t *testing.T) {
		schema := Lazy[*ZodBool[bool]](func() *ZodBool[bool] {
			return Bool()
		})

		result := schema.MustParse(true)
		assert.IsType(t, true, result)
		assert.Equal(t, true, result)
	})

	t.Run("delegation to inner schema", func(t *testing.T) {
		// Test that lazy schema properly delegates to inner schema
		schema := LazyAny(func() any {
			return Bool()
		})

		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.IsType(t, true, result)
		assert.Equal(t, true, result)

		_, err = schema.Parse("not a boolean")
		assert.Error(t, err)
	})

	t.Run("type-safe delegation to inner schema", func(t *testing.T) {
		schema := Lazy[*ZodBool[bool]](func() *ZodBool[bool] {
			return Bool()
		})

		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.IsType(t, true, result)
		assert.Equal(t, true, result)

		_, err = schema.Parse("not a boolean")
		assert.Error(t, err)
	})
}

// =============================================================================
// Modifier methods tests
// =============================================================================

func TestLazy_Modifiers(t *testing.T) {
	t.Run("Optional allows nil", func(t *testing.T) {
		schema := LazyAny(func() any {
			return String()
		})
		optionalSchema := schema.Optional()

		// Type check: ensure it returns *ZodLazy[*any]
		var _ = optionalSchema

		// Test with valid input
		result, err := optionalSchema.Parse("hello")
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Test with nil (should work with optional)
		result, err = optionalSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("type-safe Optional allows nil", func(t *testing.T) {
		schema := Lazy[*ZodString[string]](func() *ZodString[string] {
			return String()
		})
		optionalSchema := schema.Optional()

		// Type check: ensure it returns *ZodLazy[*any]
		var _ = optionalSchema

		// Test with valid input
		result, err := optionalSchema.Parse("hello")
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Test with nil (should work with optional)
		result, err = optionalSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable allows nil", func(t *testing.T) {
		schema := LazyAny(func() any {
			return String()
		})
		nilableSchema := schema.Nilable()

		var _ = nilableSchema

		// Test nil handling
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("type-safe Nilable allows nil", func(t *testing.T) {
		schema := Lazy[*ZodString[string]](func() *ZodString[string] {
			return String()
		})
		nilableSchema := schema.Nilable()

		var _ = nilableSchema

		// Test nil handling
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default preserves current type", func(t *testing.T) {
		defaultValue := "default"

		// any maintains any
		schema := LazyAny(func() any {
			return String()
		})
		defaultSchema := schema.Default(defaultValue)
		var _ = defaultSchema

		// *any maintains *any
		ptrSchema := LazyPtr(func() any {
			return String()
		})
		defaultPtrSchema := ptrSchema.Default(defaultValue)
		var _ = defaultPtrSchema
	})

	t.Run("type-safe Default preserves type", func(t *testing.T) {
		defaultValue := "default"

		schema := Lazy[*ZodString[string]](func() *ZodString[string] {
			return String()
		})
		defaultSchema := schema.Default(defaultValue)
		var _ = defaultSchema

		// Test behavior
		result, err := defaultSchema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("Prefault preserves current type", func(t *testing.T) {
		prefaultValue := "prefault"

		// any maintains any
		schema := LazyAny(func() any {
			return String()
		})
		prefaultSchema := schema.Prefault(prefaultValue)
		var _ = prefaultSchema

		// *any maintains *any
		ptrSchema := LazyPtr(func() any {
			return String()
		})
		prefaultPtrSchema := ptrSchema.Prefault(prefaultValue)
		var _ = prefaultPtrSchema
	})
}

// =============================================================================
// Chaining tests
// =============================================================================

func TestLazy_Chaining(t *testing.T) {
	t.Run("type evolution through chaining", func(t *testing.T) {
		defaultValue := "default"

		// Chain with type evolution
		schema := LazyAny(func() any {
			return String()
		}). // *ZodLazy[any]
			Default(defaultValue). // *ZodLazy[any] (maintains type)
			Optional()             // *ZodLazy[*any] (type conversion)

		var _ = schema

		// Test final behavior
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("type-safe chaining", func(t *testing.T) {
		defaultValue := "default"

		schema := Lazy[*ZodString[string]](func() *ZodString[string] {
			return String().Min(2)
		}).
			Default(defaultValue) // Preserves ZodLazyTyped type

		var _ = schema

		// Test behavior
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("complex chaining", func(t *testing.T) {
		schema := LazyPtr(func() any {
			return String()
		}). // *ZodLazy[*any]
			Nilable().      // *ZodLazy[*any] (maintains type)
			Default("test") // *ZodLazy[*any] (maintains type)

		var _ = schema

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("default and prefault chaining", func(t *testing.T) {
		defaultValue := "default"
		prefaultValue := "prefault"

		schema := LazyAny(func() any {
			return String()
		}).
			Default(defaultValue).
			Prefault(prefaultValue)

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})
}

// =============================================================================
// Default and prefault tests
// =============================================================================

func TestLazy_DefaultAndPrefault(t *testing.T) {
	// Test 1: Default has higher priority than Prefault
	t.Run("Default priority over Prefault", func(t *testing.T) {
		schema := LazyAny(func() any {
			return String().Min(3)
		}).Default("default_value").Prefault("prefault_value")

		// When input is nil, Default should take precedence
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "default_value", result)
	})

	// Test 2: Default short-circuit mechanism
	t.Run("Default short-circuit bypasses validation", func(t *testing.T) {
		schema := LazyAny(func() any {
			return String().Min(10) // Require at least 10 characters
		}).Default("short") // Default value violates constraint

		// Default should bypass validation even if it violates constraints
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "short", result)
	})

	// Test 3: Prefault requires full validation
	t.Run("Prefault requires full validation", func(t *testing.T) {
		schema := LazyAny(func() any {
			return String().Min(10) // Require at least 10 characters
		}).Prefault("short") // Prefault value violates constraint

		// Prefault should fail validation if it violates constraints
		_, err := schema.Parse(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Too small")
	})

	// Test 4: Prefault only triggers on nil input
	t.Run("Prefault only triggers on nil input", func(t *testing.T) {
		schema := LazyAny(func() any {
			return String().Min(10) // Require at least 10 characters
		}).Prefault("prefault_value")

		// Non-nil input that fails validation should not trigger Prefault
		_, err := schema.Parse("short") // This input is too short
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Too small")
	})

	// Test 5: DefaultFunc and PrefaultFunc behavior
	t.Run("DefaultFunc and PrefaultFunc behavior", func(t *testing.T) {
		defaultCalled := false
		prefaultCalled := false

		schema := LazyAny(func() any {
			return String().Min(3)
		}).DefaultFunc(func() any {
			defaultCalled = true
			return "default_func"
		}).PrefaultFunc(func() any {
			prefaultCalled = true
			return "prefault_func"
		})

		// DefaultFunc should be called and take precedence
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "default_func", result)
		assert.True(t, defaultCalled)
		assert.False(t, prefaultCalled) // PrefaultFunc should not be called
	})

	// Test 6: Error handling for Prefault validation failure
	t.Run("Prefault validation failure returns error", func(t *testing.T) {
		schema := LazyAny(func() any {
			return String().Min(5)
		}).Prefault("bad") // Too short

		// Should return validation error, not attempt fallback
		_, err := schema.Parse(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Too small")
	})
}

// =============================================================================
// Refine tests
// =============================================================================

func TestLazy_Refine(t *testing.T) {
	t.Run("refine validate", func(t *testing.T) {
		// Only accept values that pass inner schema and custom validation
		schema := LazyAny(func() any {
			return String()
		}).Refine(func(val any) bool {
			if str, ok := val.(string); ok {
				return len(str) > 3
			}
			return false
		})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("type-safe refine validate", func(t *testing.T) {
		schema := Lazy[*ZodString[string]](func() *ZodString[string] {
			return String()
		}).Refine(func(val any) bool {
			if str, ok := val.(string); ok {
				return len(str) > 3
			}
			return false
		})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		errorMessage := "String must be longer than 3 characters"
		schema := LazyAny(func() any {
			return String()
		}).Refine(func(val any) bool {
			if str, ok := val.(string); ok {
				return len(str) > 3
			}
			return false
		}, core.SchemaParams{Error: errorMessage})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("refine delegation to inner schema", func(t *testing.T) {
		schema := LazyAny(func() any {
			return Bool()
		}).Refine(func(val any) bool {
			if b, ok := val.(bool); ok {
				return b == true // Only accept true
			}
			return false
		})

		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		_, err = schema.Parse(false)
		assert.Error(t, err)

		_, err = schema.Parse("not a boolean")
		assert.Error(t, err)
	})
}

func TestLazy_RefineAny(t *testing.T) {
	t.Run("refineAny lazy schema", func(t *testing.T) {
		// Only accept strings longer than 3 characters
		schema := LazyAny(func() any {
			return String()
		}).RefineAny(func(v any) bool {
			if str, ok := v.(string); ok {
				return len(str) > 3
			}
			return false
		})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("type-safe refineAny", func(t *testing.T) {
		schema := Lazy[*ZodBool[bool]](func() *ZodBool[bool] {
			return Bool()
		}).Refine(func(v any) bool {
			if b, ok := v.(bool); ok {
				return b == true // Only accept true
			}
			return false
		})

		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		_, err = schema.Parse(false)
		assert.Error(t, err)
	})

	t.Run("refineAny with complex validation", func(t *testing.T) {
		schema := LazyAny(func() any {
			return String()
		}).RefineAny(func(v any) bool {
			// Accept only non-empty strings
			if str, ok := v.(string); ok {
				return len(str) > 0
			}
			return false
		})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse("")
		assert.Error(t, err)
	})
}

// =============================================================================
// Coercion tests (Complex types - Lazy does not support coercion)
// =============================================================================

func TestLazy_Coercion(t *testing.T) {
	t.Run("lazy delegates coercion to inner schema", func(t *testing.T) {
		// String schema supports coercion
		stringSchema := LazyAny(func() any {
			return String()
		})

		// String schema supports coercion, so this should work
		result, canCoerce := stringSchema.Coerce(123) // number to string
		assert.True(t, canCoerce)
		assert.Equal(t, "123", result)

		// Bool schema supports coercion
		boolSchema := LazyAny(func() any {
			return Bool()
		})

		result2, canCoerce2 := boolSchema.Coerce("true")
		assert.True(t, canCoerce2)
		assert.Equal(t, true, result2)
	})

	t.Run("lazy coercion with non-coercible input", func(t *testing.T) {
		schema := LazyAny(func() any {
			return Bool()
		})

		// Test with input that cannot be coerced to bool
		result, canCoerce := schema.Coerce("invalid-bool")
		assert.False(t, canCoerce)
		assert.Equal(t, "invalid-bool", result)
	})

	t.Run("lazy coercion with nil getter", func(t *testing.T) {
		schema := LazyAny(func() any {
			return nil
		})

		result, canCoerce := schema.Coerce("test")
		assert.False(t, canCoerce)
		assert.Equal(t, "test", result)
	})
}

// =============================================================================
// Error handling tests
// =============================================================================

func TestLazy_ErrorHandling(t *testing.T) {
	t.Run("propagates inner schema errors", func(t *testing.T) {
		schema := LazyAny(func() any {
			return String().Min(5)
		})

		_, err := schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("type-safe error propagation", func(t *testing.T) {
		schema := Lazy[*ZodString[string]](func() *ZodString[string] {
			return String().Min(5)
		})

		_, err := schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("handles getter function errors", func(t *testing.T) {
		schema := LazyAny(func() any {
			return nil // This should cause an error
		})

		_, err := schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("nil handling without modifiers", func(t *testing.T) {
		schema := LazyAny(func() any {
			return String()
		})

		_, err := schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("type-safe nil handling", func(t *testing.T) {
		schema := Lazy[*ZodString[string]](func() *ZodString[string] {
			return String()
		})

		_, err := schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected a valid lazy value"
		schema := LazyAny(func() any {
			return String()
		}, core.SchemaParams{Error: customError})

		_, err := schema.Parse(123)
		assert.Error(t, err)
	})
}

// =============================================================================
// Edge case tests
// =============================================================================

func TestLazy_EdgeCases(t *testing.T) {
	t.Run("recursive schema evaluation", func(t *testing.T) {
		TreeNodeSchema := LazyAny(func() any {
			// In a real implementation, this would be a proper object schema
			// For this test, we'll use a simple string schema
			return String()
		})

		result, err := TreeNodeSchema.Parse("root")
		require.NoError(t, err)
		assert.Equal(t, "root", result)
	})

	t.Run("type-safe recursive schema", func(t *testing.T) {
		RecursiveSchema := Lazy[*ZodString[string]](func() *ZodString[string] {
			return String().Min(2)
		})

		result, err := RecursiveSchema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("nil handling with nilable", func(t *testing.T) {
		schema := LazyAny(func() any {
			return String()
		}).Nilable()

		// Test nil input
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid value
		result, err = schema.Parse("test")
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("multiple lazy schemas independence", func(t *testing.T) {
		schema1 := LazyAny(func() any {
			return String().Min(3)
		})

		schema2 := LazyAny(func() any {
			return Bool()
		})

		// Both schemas should work independently
		result1, err1 := schema1.Parse("hello")
		require.NoError(t, err1)
		assert.Equal(t, "hello", result1)

		result2, err2 := schema2.Parse(true)
		require.NoError(t, err2)
		assert.Equal(t, true, result2)

		// Cross-validation should fail appropriately
		_, err := schema1.Parse(42)
		assert.Error(t, err)

		_, err = schema2.Parse("hello")
		assert.Error(t, err)
	})

	t.Run("type-safe schema independence", func(t *testing.T) {
		stringSchema := Lazy[*ZodString[string]](func() *ZodString[string] {
			return String().Min(3)
		})

		boolSchema := Lazy[*ZodBool[bool]](func() *ZodBool[bool] {
			return Bool()
		})

		// Both schemas should work independently
		result1, err1 := stringSchema.Parse("hello")
		require.NoError(t, err1)
		assert.Equal(t, "hello", result1)

		result2, err2 := boolSchema.Parse(true)
		require.NoError(t, err2)
		assert.Equal(t, true, result2)

		// Cross-validation should fail appropriately
		_, err := stringSchema.Parse(42)
		assert.Error(t, err)

		_, err = boolSchema.Parse("hello")
		assert.Error(t, err)
	})

	t.Run("empty context", func(t *testing.T) {
		schema := LazyAny(func() any {
			return String()
		})

		// Parse with empty context slice
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("lazy with various inner schemas", func(t *testing.T) {
		testCases := []struct {
			name   string
			schema func() any
			input  any
			valid  bool
		}{
			{"string schema", func() any { return String() }, "hello", true},
			{"bool schema", func() any { return Bool() }, true, true},
			{"invalid for string", func() any { return String() }, 123, false},
			{"invalid for bool", func() any { return Bool() }, "not bool", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				schema := LazyAny(tc.schema)
				_, err := schema.Parse(tc.input)
				if tc.valid {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
				}
			})
		}
	})
}

// =============================================================================
// Type-specific methods tests
// =============================================================================

func TestLazy_TypeSpecificMethods(t *testing.T) {
	t.Run("Unwrap returns inner schema", func(t *testing.T) {
		innerSchema := String().Min(3)
		lazySchema := LazyAny(func() any {
			return innerSchema
		})

		// Force evaluation by parsing once
		_, _ = lazySchema.Parse("hello")

		// Now unwrap should return the inner schema
		unwrapped := lazySchema.Unwrap()
		assert.NotNil(t, unwrapped)

		// Unwrapped schema should work the same as inner schema
		result, err := unwrapped.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = unwrapped.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("type-safe GetInnerSchema returns original type", func(t *testing.T) {
		originalString := String().Min(3)
		schema := Lazy[*ZodString[string]](func() *ZodString[string] {
			return originalString
		})

		// InnerSchema returns the original typed schema
		innerSchema := schema.GetInnerSchema()
		assert.Equal(t, originalString, innerSchema)

		// Inner schema should work independently
		result, err := innerSchema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = innerSchema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("Unwrap before evaluation", func(t *testing.T) {
		evaluationCount := 0
		lazySchema := LazyAny(func() any {
			evaluationCount++
			return String()
		})

		// Unwrap should trigger evaluation
		unwrapped := lazySchema.Unwrap()
		assert.NotNil(t, unwrapped)
		assert.Equal(t, 1, evaluationCount)

		// Second unwrap should use cached schema
		unwrapped2 := lazySchema.Unwrap()
		assert.NotNil(t, unwrapped2)
		assert.Equal(t, 1, evaluationCount, "Schema should be cached")
	})

	t.Run("type-safe schema caching", func(t *testing.T) {
		evaluationCount := 0
		schema := Lazy[*ZodString[string]](func() *ZodString[string] {
			evaluationCount++
			return String()
		})

		// InnerSchema calls the getter function each time, so it's not cached
		// This is different from the internal schema caching in Parse operations
		inner1 := schema.GetInnerSchema()
		assert.NotNil(t, inner1)
		assert.Equal(t, 1, evaluationCount)

		// Second access calls getter again (this is expected behavior)
		inner2 := schema.GetInnerSchema()
		assert.NotNil(t, inner2)
		assert.Equal(t, 2, evaluationCount, "InnerSchema calls getter each time")

		// Parse operations also trigger getter calls in the current implementation
		_, err := schema.Parse("hello")
		require.NoError(t, err)
		// Parse triggers getter function call
		assert.Equal(t, 3, evaluationCount, "Parse operations also call getter")

		// However, the internal caching mechanism works at the wrapper level
		// The ZodLazyTyped doesn't cache between GetInnerSchema and Parse operations
		// This is the expected behavior for type-safe schemas
	})

	t.Run("Coerce delegation", func(t *testing.T) {
		// Test with schema that supports coercion
		lazySchema := LazyAny(func() any {
			return Bool() // Bool supports coercion
		})

		// Test coercion delegation
		result, canCoerce := lazySchema.Coerce("true")
		assert.True(t, canCoerce)
		assert.Equal(t, true, result)

		// Test with non-coercible input
		result2, canCoerce2 := lazySchema.Coerce("invalid")
		assert.False(t, canCoerce2)
		assert.Equal(t, "invalid", result2)
	})

	t.Run("type-safe schema with specialized methods", func(t *testing.T) {
		schema := Lazy[*ZodString[string]](func() *ZodString[string] {
			return String().Min(3).Max(20).Email()
		})

		// Access inner schema and test its specialized methods
		innerSchema := schema.GetInnerSchema()

		// Test that inner schema works correctly
		result, err := innerSchema.Parse("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", result)

		// Test validation
		_, err = innerSchema.Parse("invalid-email")
		assert.Error(t, err)

		_, err = innerSchema.Parse("x@y") // Too short
		assert.Error(t, err)
	})

	t.Run("Coerce with nil getter", func(t *testing.T) {
		lazySchema := LazyAny(func() any {
			return nil
		})

		result, canCoerce := lazySchema.Coerce("test")
		assert.False(t, canCoerce)
		assert.Equal(t, "test", result)
	})
}

// =============================================================================
// Concurrent thread-safety tests (sync.Once)
// =============================================================================

func TestLazy_ConcurrentAccess(t *testing.T) {
	t.Run("concurrent getInnerType calls are thread-safe", func(t *testing.T) {
		evaluationCount := 0
		var mu sync.Mutex

		schema := LazyAny(func() any {
			mu.Lock()
			evaluationCount++
			mu.Unlock()
			// Simulate some work
			time.Sleep(time.Millisecond)
			return String().Min(3)
		})

		// Launch multiple goroutines that all try to parse concurrently
		const numGoroutines = 100
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		results := make([]any, numGoroutines)
		errs := make([]error, numGoroutines)

		for i := range numGoroutines {
			go func(idx int) {
				defer wg.Done()
				results[idx], errs[idx] = schema.Parse("hello")
			}(i)
		}

		wg.Wait()

		// All parses should succeed
		for i := range numGoroutines {
			require.NoError(t, errs[i], "Goroutine %d failed", i)
			assert.Equal(t, "hello", results[i], "Goroutine %d got wrong result", i)
		}

		// The getter function should have been called exactly once due to sync.Once
		mu.Lock()
		count := evaluationCount
		mu.Unlock()
		assert.Equal(t, 1, count, "Getter should be called exactly once with sync.Once")
	})

	t.Run("concurrent access with type-safe Lazy", func(t *testing.T) {
		evaluationCount := 0
		var mu sync.Mutex

		schema := Lazy[*ZodString[string]](func() *ZodString[string] {
			mu.Lock()
			evaluationCount++
			mu.Unlock()
			time.Sleep(time.Millisecond)
			return String().Min(2)
		})

		const numGoroutines = 50
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := range numGoroutines {
			go func(idx int) {
				defer wg.Done()
				result, err := schema.Parse("test")
				require.NoError(t, err)
				assert.Equal(t, "test", result)
			}(i)
		}

		wg.Wait()

		// Verify getter was called only once
		mu.Lock()
		count := evaluationCount
		mu.Unlock()
		assert.Equal(t, 1, count, "Type-safe Lazy getter should be called exactly once")
	})

	t.Run("concurrent Unwrap calls are thread-safe", func(t *testing.T) {
		evaluationCount := 0
		var mu sync.Mutex

		schema := LazyAny(func() any {
			mu.Lock()
			evaluationCount++
			mu.Unlock()
			time.Sleep(time.Millisecond)
			return Bool()
		})

		const numGoroutines = 50
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for range numGoroutines {
			go func() {
				defer wg.Done()
				unwrapped := schema.Unwrap()
				assert.NotNil(t, unwrapped)
			}()
		}

		wg.Wait()

		mu.Lock()
		count := evaluationCount
		mu.Unlock()
		assert.Equal(t, 1, count, "Unwrap should trigger getter exactly once")
	})

	t.Run("cloned lazy schema has independent sync.Once", func(t *testing.T) {
		originalEvalCount := 0
		var mu sync.Mutex

		original := LazyAny(func() any {
			mu.Lock()
			originalEvalCount++
			mu.Unlock()
			return String()
		})

		// Force evaluation on original
		_, err := original.Parse("test")
		require.NoError(t, err)

		mu.Lock()
		assert.Equal(t, 1, originalEvalCount)
		mu.Unlock()

		// Clone via Optional (creates new instance)
		cloned := original.Optional()

		// Parse on cloned should work without re-evaluating original's getter
		// because cloned inherits cached state
		result, err := cloned.Parse("hello")
		require.NoError(t, err)
		assert.NotNil(t, result)

		mu.Lock()
		// Original getter should still be 1 (cloned uses cached inner type)
		assert.Equal(t, 1, originalEvalCount, "Cloned schema should use cached inner type")
		mu.Unlock()
	})
}

func TestLazy_NonOptional(t *testing.T) {
	t.Run("non-optional on ZodLazy", func(t *testing.T) {
		schema := LazyAny(func() any {
			return String()
		}).Optional().NonOptional()

		// valid value should pass
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// nil should fail
		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("non-optional on ZodLazyTyped", func(t *testing.T) {
		schema := Lazy[*ZodString[string]](func() *ZodString[string] {
			return String()
		}).Optional().NonOptional()

		// valid value should pass
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// nil should fail
		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("chained non-optional", func(t *testing.T) {
		schema := LazyAny(func() any {
			return Int()
		}).Optional().NonOptional().Optional().NonOptional()

		// valid value should pass
		result, err := schema.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, 123, result)

		// nil should fail
		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("non-optional with recursive struct", func(t *testing.T) {
		type Node struct {
			Value string `json:"value"`
			Child *Node  `json:"child,omitempty"`
		}

		var nodeSchema *ZodLazy[any]
		nodeSchema = LazyAny(func() any {
			return Struct[Node](core.StructSchema{
				"value": String(),
				"child": nodeSchema.Optional(),
			})
		})

		schema := nodeSchema.Optional().NonOptional()

		// valid node should pass
		input := Node{Value: "parent", Child: &Node{Value: "child"}}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)

		// nil should fail
		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})
}
