package types

import (
	"strings"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Basic functionality tests
// =============================================================================

func TestFunction_BasicFunctionality(t *testing.T) {
	t.Run("valid function inputs", func(t *testing.T) {
		schema := Function()

		// Test simple function
		simpleFunc := func(x int) int { return x * 2 }
		result, err := schema.Parse(simpleFunc)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Test function with multiple parameters
		multiParamFunc := func(a string, b int) string { return a }
		result, err = schema.Parse(multiParamFunc)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("invalid type inputs", func(t *testing.T) {
		schema := Function()

		invalidInputs := []any{
			"not a function", 123, 3.14, []int{1, 2, 3}, nil, map[string]int{},
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})

	t.Run("Parse and MustParse methods", func(t *testing.T) {
		schema := Function()
		testFunc := func() string { return "test" }

		// Test Parse method
		result, err := schema.Parse(testFunc)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Test MustParse method
		mustResult := schema.MustParse(testFunc)
		assert.NotNil(t, mustResult)

		// Test panic on invalid input
		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected a function value"
		schema := Function(core.SchemaParams{Error: customError})

		require.NotNil(t, schema)
		assert.Equal(t, core.ZodTypeFunction, schema.internals.Def.Type)

		_, err := schema.Parse("invalid")
		assert.Error(t, err)
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestFunction_TypeSafety(t *testing.T) {
	t.Run("Function returns any type", func(t *testing.T) {
		schema := Function()
		require.NotNil(t, schema)

		testFunc := func(x int) int { return x + 1 }
		result, err := schema.Parse(testFunc)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Verify that the result is callable as a function
		if fn, ok := result.(func(int) int); ok {
			output := fn(5)
			assert.Equal(t, 6, output)
		} else {
			t.Error("Result should be a function")
		}
	})

	t.Run("FunctionPtr returns *any type", func(t *testing.T) {
		schema := FunctionPtr()
		require.NotNil(t, schema)

		testFunc := func(x string) string { return x + "!" }
		result, err := schema.Parse(testFunc)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("type inference with assignment", func(t *testing.T) {
		// Type-inference friendly API
		funcSchema := Function()   // any type
		ptrSchema := FunctionPtr() // *any type

		testFunc := func() bool { return true }

		// Test any type
		result1, err1 := funcSchema.Parse(testFunc)
		require.NoError(t, err1)
		assert.NotNil(t, result1)

		// Test *any type
		result2, err2 := ptrSchema.Parse(testFunc)
		require.NoError(t, err2)
		assert.NotNil(t, result2)
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		schema := Function()
		testFunc := func(x float64) float64 { return x * 2 }

		result := schema.MustParse(testFunc)
		assert.NotNil(t, result)

		// Verify callable
		if fn, ok := result.(func(float64) float64); ok {
			output := fn(3.5)
			assert.Equal(t, 7.0, output)
		} else {
			t.Error("Result should be a function")
		}
	})
}

// =============================================================================
// Modifier methods tests
// =============================================================================

func TestFunction_Modifiers(t *testing.T) {
	t.Run("Optional allows nil", func(t *testing.T) {
		schema := Function()
		optionalSchema := schema.Optional()

		// Type check: ensure it returns *ZodFunction[*any]
		_ = optionalSchema

		// Test with function
		testFunc := func() {}
		result, err := optionalSchema.Parse(testFunc)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Test with nil (should work with optional)
		result, err = optionalSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable allows nil", func(t *testing.T) {
		schema := Function()
		nilableSchema := schema.Nilable()

		_ = nilableSchema

		// Test nil handling
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default preserves current type", func(t *testing.T) {
		defaultFunc := func() string { return "default" }

		// any maintains any
		schema := Function()
		defaultSchema := schema.Default(defaultFunc)
		_ = defaultSchema

		// *any maintains *any
		ptrSchema := FunctionPtr()
		defaultPtrSchema := ptrSchema.Default(defaultFunc)
		_ = defaultPtrSchema
	})

	t.Run("Prefault preserves current type", func(t *testing.T) {
		prefaultFunc := func() int { return 42 }

		// any maintains any
		schema := Function()
		prefaultSchema := schema.Prefault(prefaultFunc)
		var _ = prefaultSchema

		// *any maintains *any
		ptrSchema := FunctionPtr()
		prefaultPtrSchema := ptrSchema.Prefault(prefaultFunc)
		var _ = prefaultPtrSchema
	})
}

// =============================================================================
// Chaining tests
// =============================================================================

func TestFunction_Chaining(t *testing.T) {
	t.Run("type evolution through chaining", func(t *testing.T) {
		defaultFunc := func() {}

		// Chain with type evolution
		schema := Function(). // *ZodFunction[any]
					Default(defaultFunc). // *ZodFunction[any] (maintains type)
					Optional()            // *ZodFunction[*any] (type conversion)

		var _ = schema

		// Test final behavior
		testFunc := func(x int) int { return x * 2 }
		result, err := schema.Parse(testFunc)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("complex chaining", func(t *testing.T) {
		defaultFunc := func() string { return "test" }

		schema := FunctionPtr(). // *ZodFunction[*any]
						Nilable().           // *ZodFunction[*any] (maintains type)
						Default(defaultFunc) // *ZodFunction[*any] (maintains type)

		var _ = schema

		result, err := schema.Parse(func() {})
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("default and prefault chaining", func(t *testing.T) {
		defaultFunc := func() int { return 1 }
		prefaultFunc := func() int { return 2 }

		schema := Function().
			Default(defaultFunc).
			Prefault(prefaultFunc)

		testFunc := func() int { return 3 }
		result, err := schema.Parse(testFunc)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// =============================================================================
// Default and prefault tests
// =============================================================================

func TestFunction_DefaultAndPrefault(t *testing.T) {
	// Test 1: Default has higher priority than Prefault
	t.Run("Default has higher priority than Prefault", func(t *testing.T) {
		defaultFunc := func() string { return "default" }
		prefaultFunc := func() string { return "prefault" }
		schema := Function().Default(defaultFunc).Prefault(prefaultFunc)

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.NotNil(t, result)
		// Should be default function, not prefault
		if fn, ok := result.(func() string); ok {
			assert.Equal(t, "default", fn())
		} else {
			t.Errorf("Expected function type, got %T", result)
		}
	})

	// Test 2: Default short-circuits validation
	t.Run("Default short-circuits validation", func(t *testing.T) {
		// Default value violates refinement but should still work
		defaultFunc := "not-a-function" // Invalid type
		schema := Function().Refine(func(f any) bool {
			return false // Always fail refinement
		}, "Should never pass").Default(defaultFunc)
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "not-a-function", result) // Default bypasses validation
	})

	// Test 3: Prefault goes through full validation
	t.Run("Prefault goes through full validation", func(t *testing.T) {
		// Prefault value passes validation
		prefaultFunc := func() int { return 42 }
		schema := Function().Refine(func(f any) bool {
			// Only allow functions that return int
			if fn, ok := f.(func() int); ok {
				return fn() > 0
			}
			return false
		}, "Function must return positive int").Prefault(prefaultFunc)
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.NotNil(t, result)
		if fn, ok := result.(func() int); ok {
			assert.Equal(t, 42, fn())
		} else {
			t.Errorf("Expected function type, got %T", result)
		}
	})

	// Test 4: Prefault only triggered by nil input
	t.Run("Prefault only triggered by nil input", func(t *testing.T) {
		prefaultFunc := func() string { return "prefault" }
		schema := Function().Prefault(prefaultFunc)

		// Valid input should override prefault
		testFunc := func() string { return "test" }
		result, err := schema.Parse(testFunc)
		require.NoError(t, err)
		assert.NotNil(t, result)
		if fn, ok := result.(func() string); ok {
			assert.Equal(t, "test", fn()) // Should be input, not prefault
		} else {
			t.Errorf("Expected function type, got %T", result)
		}

		// Invalid input should NOT trigger prefault (should return error)
		_, err = schema.Parse("not-a-function")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid input: expected function")
	})

	// Test 5: DefaultFunc and PrefaultFunc behavior
	t.Run("DefaultFunc and PrefaultFunc behavior", func(t *testing.T) {
		defaultCalled := false
		prefaultCalled := false

		defaultFuncProvider := func() any {
			defaultCalled = true
			return func() string { return "default" }
		}

		prefaultFuncProvider := func() any {
			prefaultCalled = true
			return func() string { return "prefault" }
		}

		// Test DefaultFunc priority over PrefaultFunc
		schema1 := Function().DefaultFunc(defaultFuncProvider).PrefaultFunc(prefaultFuncProvider)
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		assert.True(t, defaultCalled)
		assert.False(t, prefaultCalled) // Should not be called due to default priority
		assert.NotNil(t, result1)
		if fn, ok := result1.(func() string); ok {
			assert.Equal(t, "default", fn())
		} else {
			t.Errorf("Expected function type, got %T", result1)
		}

		// Reset flags
		defaultCalled = false
		prefaultCalled = false

		// Test PrefaultFunc alone
		schema2 := Function().PrefaultFunc(prefaultFuncProvider)
		result2, err2 := schema2.Parse(nil)
		require.NoError(t, err2)
		assert.True(t, prefaultCalled)
		assert.NotNil(t, result2)
		if fn, ok := result2.(func() string); ok {
			assert.Equal(t, "prefault", fn())
		} else {
			t.Errorf("Expected function type, got %T", result2)
		}
	})

	// Test 6: Prefault validation failure
	t.Run("Prefault validation failure", func(t *testing.T) {
		// Prefault value violates refinement
		invalidFunc := func() int { return -1 } // Negative value
		schema := Function().Refine(func(f any) bool {
			if fn, ok := f.(func() int); ok {
				return fn() > 0 // Only allow positive returns
			}
			return false
		}, "Function must return positive int").Prefault(invalidFunc)
		_, err := schema.Parse(nil)
		require.Error(t, err) // Prefault should fail validation
		assert.Contains(t, err.Error(), "Function must return positive int")
	})

	// Test 7: FunctionPtr type with Default and Prefault
	t.Run("FunctionPtr type with Default and Prefault", func(t *testing.T) {
		defaultFunc := func() string { return "default" }
		prefaultFunc := func() string { return "prefault" }
		schema := FunctionPtr().Default(defaultFunc).Prefault(prefaultFunc)

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		// Should be default function, not prefault
		if fn, ok := (*result).(func() string); ok {
			assert.Equal(t, "default", fn())
		} else {
			t.Errorf("Expected function type, got %T", *result)
		}
	})
}

// =============================================================================
// Refine tests
// =============================================================================

func TestFunction_Refine(t *testing.T) {
	t.Run("refine validate", func(t *testing.T) {
		// Only accept functions (always true for this type)
		schema := Function().Refine(func(f any) bool {
			return f != nil
		})

		testFunc := func() {}
		result, err := schema.Parse(testFunc)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		errorMessage := "Function must not be nil"
		schema := Function().Refine(func(f any) bool {
			return f != nil
		}, core.SchemaParams{Error: errorMessage})

		testFunc := func() {}
		result, err := schema.Parse(testFunc)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("always failing refine", func(t *testing.T) {
		schema := Function().Refine(func(f any) bool {
			return false // Always fail
		})

		testFunc := func() {}
		_, err := schema.Parse(testFunc)
		assert.Error(t, err)
	})
}

func TestFunction_RefineAny(t *testing.T) {
	t.Run("refineAny function schema", func(t *testing.T) {
		// Only accept non-nil functions
		schema := Function().RefineAny(func(v any) bool {
			return v != nil
		})

		// function passes
		testFunc := func() {}
		result, err := schema.Parse(testFunc)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("refineAny with complex validation", func(t *testing.T) {
		schema := Function().RefineAny(func(v any) bool {
			// Accept only functions with specific signature (example)
			return v != nil
		})

		result, err := schema.Parse(func(x int) int { return x })
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// =============================================================================
// Type-specific methods tests
// =============================================================================

func TestFunction_TypeSpecificMethods(t *testing.T) {
	t.Run("Input method sets input schema", func(t *testing.T) {
		// Use nil for input schema in this simplified test
		schema := Function().Input(nil)

		require.NotNil(t, schema)
		assert.Nil(t, schema.internals.Input)
	})

	t.Run("Output method sets output schema", func(t *testing.T) {
		// Use nil for output schema in this simplified test
		schema := Function().Output(nil)

		require.NotNil(t, schema)
		assert.Nil(t, schema.internals.Output)
	})

	t.Run("Input and Output chaining", func(t *testing.T) {
		// Use nil for both schemas in this simplified test
		schema := Function().
			Input(nil).
			Output(nil)

		require.NotNil(t, schema)
		assert.Nil(t, schema.internals.Input)
		assert.Nil(t, schema.internals.Output)
	})

	t.Run("Implement method wraps function with validation", func(t *testing.T) {
		schema := Function()
		originalFunc := func(x int) int { return x * 2 }

		wrappedFunc, err := schema.Implement(originalFunc)
		require.NoError(t, err)
		assert.NotNil(t, wrappedFunc)

		// Test that wrapped function works
		if fn, ok := wrappedFunc.(func(int) int); ok {
			result := fn(5)
			assert.Equal(t, 10, result)
		} else {
			t.Error("Implement should return a callable function")
		}
	})

	t.Run("Implement with invalid input", func(t *testing.T) {
		schema := Function()

		_, err := schema.Implement("not a function")
		assert.Error(t, err)

		_, err = schema.Implement(123)
		assert.Error(t, err)

		_, err = schema.Implement(nil)
		assert.Error(t, err)
	})

	t.Run("Implement with input/output validation", func(t *testing.T) {
		// This is a simplified test - full validation would require
		// working Slice and String implementations
		schema := Function()

		// Simple function that should work
		simpleFunc := func(x string) int { return len(x) }
		wrappedFunc, err := schema.Implement(simpleFunc)
		require.NoError(t, err)
		assert.NotNil(t, wrappedFunc)
	})
}

// =============================================================================
// Error handling and edge case tests
// =============================================================================

func TestFunction_ErrorHandling(t *testing.T) {
	t.Run("invalid type error", func(t *testing.T) {
		schema := Function()

		_, err := schema.Parse("not a function")
		assert.Error(t, err)

		_, err = schema.Parse(123)
		assert.Error(t, err)

		_, err = schema.Parse([]int{1, 2, 3})
		assert.Error(t, err)
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected a function value"
		schema := Function(core.SchemaParams{Error: customError})

		_, err := schema.Parse("not a function")
		assert.Error(t, err)
	})

	t.Run("nil handling without modifiers", func(t *testing.T) {
		schema := Function()

		_, err := schema.Parse(nil)
		assert.Error(t, err)
	})
}

func TestFunction_EdgeCases(t *testing.T) {
	t.Run("nil handling with nilable", func(t *testing.T) {
		schema := Function().Nilable()

		// Test nil input
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid function
		testFunc := func() {}
		result, err = schema.Parse(testFunc)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("empty context", func(t *testing.T) {
		schema := Function()
		testFunc := func() {}

		// Parse with empty context slice
		result, err := schema.Parse(testFunc)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("function with various signatures", func(t *testing.T) {
		schema := Function()

		testCases := []any{
			func() {},
			func(int) {},
			func() int { return 0 },
			func(int) int { return 0 },
			func(string, bool) (int, error) { return 0, nil },
			func(...int) {},
		}

		for i, testFunc := range testCases {
			result, err := schema.Parse(testFunc)
			require.NoError(t, err, "Test case %d should pass", i)
			assert.NotNil(t, result, "Test case %d should return non-nil", i)
		}
	})

	t.Run("function parameters validation", func(t *testing.T) {
		// Test FunctionParams struct with nil schemas
		schema := Function(FunctionParams{
			Input:  nil,
			Output: nil,
		})

		require.NotNil(t, schema)
		assert.Nil(t, schema.internals.Input)
		assert.Nil(t, schema.internals.Output)
	})
}

// =============================================================================
// OVERWRITE TESTS
// =============================================================================

func TestFunction_Overwrite(t *testing.T) {
	t.Run("basic function transformation", func(t *testing.T) {
		schema := Function().
			Overwrite(func(fn any) any {
				// Wrap the function with logging
				if f, ok := fn.(func(int) int); ok {
					return func(x int) int {
						result := f(x)
						// In real scenarios, you might log here
						return result + 1 // Add 1 to demonstrate transformation
					}
				}
				return fn
			})

		// Test function
		originalFunc := func(x int) int {
			return x * 2
		}

		result, err := schema.Parse(originalFunc)
		require.NoError(t, err)

		// The result should be a function that has been transformed
		if transformedFunc, ok := result.(func(int) int); ok {
			// Original function: 5 * 2 = 10
			// Transformed function: (5 * 2) + 1 = 11
			output := transformedFunc(5)
			assert.Equal(t, 11, output)
		} else {
			t.Fatal("Expected transformed function")
		}
	})

	t.Run("function wrapper transformation", func(t *testing.T) {
		schema := Function().
			Overwrite(func(fn any) any {
				// Wrap string functions with validation
				if f, ok := fn.(func(string) string); ok {
					return func(s string) string {
						if s == "" {
							return "default"
						}
						return f(s)
					}
				}
				return fn
			})

		originalFunc := strings.ToUpper

		result, err := schema.Parse(originalFunc)
		require.NoError(t, err)

		if transformedFunc, ok := result.(func(string) string); ok {
			// Test with empty string
			output1 := transformedFunc("")
			assert.Equal(t, "default", output1)

			// Test with normal string
			output2 := transformedFunc("hello")
			assert.Equal(t, "HELLO", output2)
		} else {
			t.Fatal("Expected transformed function")
		}
	})

	t.Run("function metadata transformation", func(t *testing.T) {
		// Create a function with metadata (using closure)
		createCounter := func(initial int) func() int {
			count := initial
			return func() int {
				count++
				return count
			}
		}

		schema := Function().
			Overwrite(func(fn any) any {
				// Transform counter functions to start from a different value
				if f, ok := fn.(func() int); ok {
					// Reset the counter by creating a new one
					// This simulates transforming function behavior
					newCount := 100 // Start from 100 instead
					return func() int {
						oldResult := f() // Call original once to advance it
						_ = oldResult    // Ignore original result
						newCount++
						return newCount
					}
				}
				return fn
			})

		originalCounter := createCounter(0)

		result, err := schema.Parse(originalCounter)
		require.NoError(t, err)

		if transformedCounter, ok := result.(func() int); ok {
			// Should start from 101 (100 + 1)
			output1 := transformedCounter()
			assert.Equal(t, 101, output1)

			output2 := transformedCounter()
			assert.Equal(t, 102, output2)
		} else {
			t.Fatal("Expected transformed counter function")
		}
	})

	t.Run("pointer type handling", func(t *testing.T) {
		schema := FunctionPtr().
			Overwrite(func(fn *any) *any {
				if fn == nil {
					// Return a default function
					defaultFunc := func() string { return "default" }
					result := any(defaultFunc)
					return &result
				}

				// Transform the function if it exists
				if f, ok := (*fn).(func(string) string); ok {
					transformedFunc := func(s string) string {
						return "transformed_" + f(s)
					}
					result := any(transformedFunc)
					return &result
				}
				return fn
			})

		originalFunc := func(s string) string {
			return s + "_suffix"
		}
		originalAny := any(originalFunc)

		result, err := schema.Parse(&originalAny)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Handle the triple nested pointer structure: *any -> *any -> *any -> func(string) string
		if anyPtr1, ok := (*result).(*any); ok {
			if anyPtr2, ok := (*anyPtr1).(*any); ok {
				if transformedFunc, ok := (*anyPtr2).(func(string) string); ok {
					output := transformedFunc("test")
					assert.Equal(t, "transformed_test_suffix", output)
				} else {
					t.Fatalf("Expected transformed function, got: %T", *anyPtr2)
				}
			} else {
				t.Fatalf("Expected second *any, got: %T", *anyPtr1)
			}
		} else {
			t.Fatalf("Expected first *any, got: %T", *result)
		}
	})

	t.Run("type preservation", func(t *testing.T) {
		schema := Function().
			Overwrite(func(fn any) any {
				return fn // Identity transformation
			})

		testFunc := func(x int) int {
			return x + 1
		}

		result, err := schema.Parse(testFunc)
		require.NoError(t, err)

		// Should preserve the function type
		if resultFunc, ok := result.(func(int) int); ok {
			output := resultFunc(5)
			assert.Equal(t, 6, output)
		} else {
			t.Fatal("Expected function to be preserved")
		}
	})

	t.Run("chaining with validations", func(t *testing.T) {
		schema := Function().
			Overwrite(func(fn any) any {
				// Ensure function returns non-negative values
				if f, ok := fn.(func(int) int); ok {
					return func(x int) int {
						result := f(x)
						if result < 0 {
							return 0
						}
						return result
					}
				}
				return fn
			}).
			Refine(func(fn any) bool {
				// Validate that function is not nil
				return fn != nil
			}, "Function cannot be nil")

		originalFunc := func(x int) int {
			return x - 10 // This could return negative values
		}

		result, err := schema.Parse(originalFunc)
		require.NoError(t, err)

		if transformedFunc, ok := result.(func(int) int); ok {
			// Test with input that would normally return negative
			output := transformedFunc(5) // 5 - 10 = -5, but should be clamped to 0
			assert.Equal(t, 0, output)

			// Test with input that returns positive
			output2 := transformedFunc(15) // 15 - 10 = 5
			assert.Equal(t, 5, output2)
		} else {
			t.Fatal("Expected transformed function")
		}
	})

	t.Run("error handling preservation", func(t *testing.T) {
		schema := Function().
			Overwrite(func(fn any) any {
				return fn // Identity transformation
			})

		// Non-function input should still fail validation
		_, err := schema.Parse("not a function")
		assert.Error(t, err)

		_, err = schema.Parse(123)
		assert.Error(t, err)
	})

	t.Run("function composition", func(t *testing.T) {
		schema := Function().
			Overwrite(func(fn any) any {
				// Compose with a prefix function
				if f, ok := fn.(func(string) string); ok {
					prefix := func(s string) string {
						return "prefix_" + s
					}
					// Compose: prefix(f(input))
					return func(s string) string {
						return prefix(f(s))
					}
				}
				return fn
			})

		originalFunc := strings.ToUpper

		result, err := schema.Parse(originalFunc)
		require.NoError(t, err)

		if composedFunc, ok := result.(func(string) string); ok {
			output := composedFunc("hello")
			assert.Equal(t, "prefix_HELLO", output)
		} else {
			t.Fatal("Expected composed function")
		}
	})
}
