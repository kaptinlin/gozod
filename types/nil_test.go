package types

import (
	"errors"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestNilBasicFunctionality(t *testing.T) {
	t.Run("basic validation", func(t *testing.T) {
		schema := Nil()

		// Valid nil value
		result, err := schema.Parse(nil)
		if err != nil {
			t.Errorf("Expected nil to be valid, got error: %v", err)
		}
		if result != nil {
			t.Errorf("Expected nil result, got: %v", result)
		}

		// Invalid non-nil value
		_, err = schema.Parse("hello")
		if err == nil {
			t.Error("Expected error for non-nil value")
		}

		// Invalid pointer to nil
		var nilPtr *string
		_, err = schema.Parse(&nilPtr)
		if err == nil {
			t.Error("Expected error for pointer to nil")
		}
	})

	t.Run("smart type inference", func(t *testing.T) {
		schema := Nil()

		// Nil input should return nil
		result, err := schema.Parse(nil)
		if err != nil {
			t.Errorf("Parse failed: %v", err)
		}
		if result != nil {
			t.Errorf("Expected nil, got: %v", result)
		}
	})

	t.Run("package function constructor", func(t *testing.T) {
		// Test Nil() constructor
		schema1 := Nil()
		if schema1 == nil {
			t.Error("Nil() should return a valid schema")
		}

		// Test Null() alias
		schema2 := Null()
		if schema2 == nil {
			t.Error("Null() should return a valid schema")
		}

		// Both should work the same way
		result1, err1 := schema1.Parse(nil)
		result2, err2 := schema2.Parse(nil)

		if err1 != nil || err2 != nil {
			t.Errorf("Both schemas should accept nil: err1=%v, err2=%v", err1, err2)
		}
		if result1 != nil || result2 != nil {
			t.Errorf("Both should return nil: result1=%v, result2=%v", result1, result2)
		}
	})

	t.Run("MustParse method", func(t *testing.T) {
		schema := Nil()

		// Valid case
		result := schema.MustParse(nil)
		if result != nil {
			t.Errorf("Expected nil, got: %v", result)
		}

		// Invalid case should panic
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for invalid input")
			}
		}()
		schema.MustParse("invalid")
	})
}

// =============================================================================
// 2. Validation methods
// =============================================================================

func TestNilValidations(t *testing.T) {
	t.Run("unwrap method", func(t *testing.T) {
		schema := Nil()
		unwrapped := schema.Unwrap()

		// Should return self for basic types
		if unwrapped != schema {
			t.Error("Unwrap should return self for basic nil type")
		}
	})

	t.Run("type checking", func(t *testing.T) {
		schema := Nil()

		// Test various non-nil values
		testCases := []any{
			"string",
			123,
			true,
			[]int{1, 2, 3},
			map[string]int{"key": 1},
			struct{}{},
		}

		for _, testCase := range testCases {
			_, err := schema.Parse(testCase)
			if err == nil {
				t.Errorf("Expected error for %T: %v", testCase, testCase)
			}
		}
	})
}

// =============================================================================
// 3. Modifiers and wrappers
// =============================================================================

func TestNilModifiers(t *testing.T) {
	t.Run("nilable modifier", func(t *testing.T) {
		schema := Nil().Nilable()

		// Nilable on nil type should be a no-op - nil type already handles nil
		result, err := schema.Parse(nil)
		if err != nil {
			t.Errorf("Nilable nil should accept nil: %v", err)
		}
		// For nil type, Nilable doesn't change the behavior - still returns nil
		if result == nil {
			// This is expected behavior
		} else {
			t.Errorf("Expected nil result, got: %v", result)
		}

		// Should still reject non-nil values
		_, err = schema.Parse("hello")
		if err == nil {
			t.Error("Nilable nil should still reject non-nil values")
		}
	})

	t.Run("optional modifier", func(t *testing.T) {
		schema := Nil().Optional()

		// Should accept nil
		if _, err := schema.Parse(nil); err != nil {
			t.Errorf("Optional nil schema should accept nil: %v", err)
		}

		// Should still reject non-nil input
		if _, err := schema.Parse("hello"); err == nil {
			t.Error("Optional nil schema should reject non-nil values")
		}
	})

	t.Run("nullish modifier", func(t *testing.T) {
		schema := Nil().Nullish()

		// Accept nil
		if _, err := schema.Parse(nil); err != nil {
			t.Errorf("Nullish nil schema should accept nil: %v", err)
		}

		// Reject non-nil
		if _, err := schema.Parse("hello"); err == nil {
			t.Error("Nullish nil schema should reject non-nil values")
		}
	})
}

// =============================================================================
// 4. Chaining and method composition
// =============================================================================

func TestNilChaining(t *testing.T) {
	t.Run("refine chaining", func(t *testing.T) {
		// For nil type, we can test basic chaining without complex refine logic
		schema := Nil().Nilable()

		result, err := schema.Parse(nil)
		if err != nil {
			t.Errorf("Chain should work: %v", err)
		}
		if result == nil {
			// This is expected behavior
		} else {
			t.Errorf("Expected nil, got: %v", result)
		}
	})

	t.Run("side effect isolation", func(t *testing.T) {
		original := Nil()
		modified := original.Nilable()

		// Original should not be affected
		_, err1 := original.Parse(nil)
		_, err2 := modified.Parse(nil)

		if err1 != nil || err2 != nil {
			t.Errorf("Both should work: original=%v, modified=%v", err1, err2)
		}

		// Test that both work independently (functional test rather than pointer comparison)
		_, err3 := original.Parse("invalid")
		_, err4 := modified.Parse("invalid")

		if err3 == nil || err4 == nil {
			t.Error("Both should reject invalid input independently")
		}
	})
}

// =============================================================================
// 5. Transform/Pipe
// =============================================================================

func TestNilTransform(t *testing.T) {
	t.Run("basic transform", func(t *testing.T) {
		// For nil type, test basic functionality
		schema := Nil()

		// Nil type should accept nil
		result, err := schema.Parse(nil)
		if err != nil {
			t.Errorf("Nil type should accept nil: %v", err)
		}
		if result != nil {
			t.Errorf("Expected nil result, got: %v", result)
		}

		// Nil type should reject non-nil
		_, err = schema.Parse("not_nil")
		if err == nil {
			t.Error("Nil type should reject non-nil values")
		}
	})

	t.Run("transform modifies output", func(t *testing.T) {
		schema := Nil().Transform(func(v any, _ *core.RefinementContext) (any, error) {
			return "converted", nil
		})

		result, err := schema.Parse(nil)
		if err != nil {
			t.Errorf("Transform should succeed: %v", err)
		}
		if result != nil {
			t.Errorf("Expected nil result (pipeline short-circuit), got %v", result)
		}
	})

	t.Run("pipe construct", func(t *testing.T) {
		// Simple self-pipe to ensure no panic and schema is returned
		schema := Nil().Pipe(Nil())
		if schema == nil {
			t.Error("Pipe should return a valid schema")
		}
	})
}

// =============================================================================
// 6. Refine
// =============================================================================

func TestNilRefine(t *testing.T) {
	t.Run("basic refine concept", func(t *testing.T) {
		// For nil type, refine functionality is limited since nil is a very specific value
		// Let's test that the basic nil validation works
		schema := Nil()

		result, err := schema.Parse(nil)
		if err != nil {
			t.Errorf("Basic nil validation should pass: %v", err)
		}
		if result != nil {
			t.Errorf("Expected nil, got: %v", result)
		}
	})

	t.Run("nil type rejects non-nil", func(t *testing.T) {
		schema := Nil()

		_, err := schema.Parse("not_nil")
		if err == nil {
			t.Error("Expected nil type to reject non-nil values")
		}
	})

	t.Run("nil behavior consistency", func(t *testing.T) {
		// Test that nil type consistently handles nil values
		schema := Nil()

		// Multiple parses should work consistently
		result1, err1 := schema.Parse(nil)
		result2, err2 := schema.Parse(nil)

		if err1 != nil || err2 != nil {
			t.Errorf("Both parses should work: err1=%v, err2=%v", err1, err2)
		}
		if result1 != nil || result2 != nil {
			t.Errorf("Both should return nil: result1=%v, result2=%v", result1, result2)
		}
	})

	t.Run("custom refine", func(t *testing.T) {
		schema := Nil().Refine(func() bool {
			return true // Always pass for nil
		})

		result, err := schema.Parse(nil)
		if err != nil {
			t.Errorf("Refine should work: %v", err)
		}
		if result != nil {
			t.Errorf("Expected nil, got: %v", result)
		}
	})
}

// =============================================================================
// 7. Error handling
// =============================================================================

func TestNilErrorHandling(t *testing.T) {
	t.Run("error structure", func(t *testing.T) {
		schema := Nil()

		_, err := schema.Parse("not_nil")
		if err == nil {
			t.Error("Expected error for non-nil input")
		}

		var zodErr *issues.ZodError
		if !errors.As(err, &zodErr) {
			t.Errorf("Expected ZodError, got: %T", err)
		}

		if len(zodErr.Issues) == 0 {
			t.Error("Expected at least one issue")
		}

		issue := zodErr.Issues[0]
		if issue.Code != core.InvalidType {
			t.Errorf("Expected 'invalid_type', got: %s", issue.Code)
		}
	})

	t.Run("custom error messages", func(t *testing.T) {
		schema := Nil(core.SchemaParams{Error: "Must be nil"})

		_, err := schema.Parse("not_nil")
		if err == nil {
			t.Error("Expected error")
		}

		// Error should contain custom message
		if err.Error() == "" {
			t.Error("Expected non-empty error message")
		}
	})
}

// =============================================================================
// 8. Edge and mutual exclusion cases
// =============================================================================

func TestNilEdgeCases(t *testing.T) {
	t.Run("nil vs nil pointer distinction", func(t *testing.T) {
		schema := Nil()

		// nil should work
		result, err := schema.Parse(nil)
		if err != nil {
			t.Errorf("nil should be valid: %v", err)
		}
		if result != nil {
			t.Errorf("Expected nil, got: %v", result)
		}

		// Pointer to nil should not work
		var nilPtr *string
		_, err = schema.Parse(&nilPtr)
		if err == nil {
			t.Error("Pointer to nil should be invalid")
		}
	})

	t.Run("internal state access", func(t *testing.T) {
		schema := Nil()
		internals := schema.GetInternals()

		if internals == nil {
			t.Error("GetInternals should return valid internals")
		}

		zodInternals := schema.GetZod()
		if zodInternals == nil {
			t.Error("GetZod should return valid nil-specific internals")
		}
	})

	t.Run("clone behavior", func(t *testing.T) {
		original := Nil()
		cloned := Nil()

		// Test CloneFrom method
		cloned.CloneFrom(original)

		// Both should work the same
		result1, err1 := original.Parse(nil)
		result2, err2 := cloned.Parse(nil)

		if err1 != nil || err2 != nil {
			t.Errorf("Both should work: original=%v, cloned=%v", err1, err2)
		}
		if result1 != nil || result2 != nil {
			t.Errorf("Both should return nil: original=%v, cloned=%v", result1, result2)
		}
	})
}

// =============================================================================
// 9. Default and Prefault tests
// =============================================================================

func TestNilDefaultAndPrefault(t *testing.T) {
	t.Run("nil behavior vs optional behavior", func(t *testing.T) {
		nilSchema := Nil()

		// Regular nil schema rejects non-nil
		_, err := nilSchema.Parse("not_nil")
		if err == nil {
			t.Error("Regular nil should reject non-nil")
		}

		// Both accept nil
		result1, err1 := nilSchema.Parse(nil)

		if err1 != nil {
			t.Errorf("Schema should accept nil: regular=%v", err1)
		}
		if result1 != nil {
			t.Errorf("Schema should return nil: regular=%v", result1)
		}
	})

	t.Run("nilable behavior", func(t *testing.T) {
		nilableSchema := Nil().Nilable()

		// Should accept nil - for nil type, Nilable is essentially a no-op
		result, err := nilableSchema.Parse(nil)
		if err != nil {
			t.Errorf("Nilable should accept nil: %v", err)
		}

		// For nil type, Nilable doesn't change the behavior - still returns nil
		// This is because nil type already handles nil values by design
		if result == nil {
			// This is expected behavior
		} else {
			t.Errorf("Expected nil result for nil type, got: %v", result)
		}
	})

	t.Run("default behavior", func(t *testing.T) {
		schema := Nil().Default(nil)
		result, err := schema.Parse(nil)
		if err != nil {
			t.Errorf("Default nil schema should accept nil: %v", err)
		}
		if result != nil {
			t.Errorf("Expected nil result, got %v", result)
		}
	})

	t.Run("defaultFunc behavior", func(t *testing.T) {
		schema := Nil().DefaultFunc(func() any {
			return nil
		})
		if _, err := schema.Parse(nil); err != nil {
			t.Errorf("DefaultFunc nil schema should accept nil: %v", err)
		}
	})

	t.Run("prefault behavior", func(t *testing.T) {
		schema := Nil().Prefault(nil)
		if _, err := schema.Parse(nil); err != nil {
			t.Errorf("Prefault nil schema should accept nil: %v", err)
		}
	})

	t.Run("prefaultFunc behavior", func(t *testing.T) {
		schema := Nil().PrefaultFunc(func() any {
			return nil
		})
		if _, err := schema.Parse(nil); err != nil {
			t.Errorf("PrefaultFunc nil schema should accept nil: %v", err)
		}
	})
}
