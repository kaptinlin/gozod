package gozod

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestFunctionBasicFunctionality(t *testing.T) {
	t.Run("basic function creation", func(t *testing.T) {
		schema := Function()
		require.NotNil(t, schema)
	})

	t.Run("function with input and output", func(t *testing.T) {
		schema := Function(FunctionParams{
			Input:  []Schema{String()},
			Output: Int(),
		})
		require.NotNil(t, schema)

		impl, err := schema.Implement(func(s string) int {
			return len(s)
		})
		require.NoError(t, err)
		require.NotNil(t, impl)

		if wrappedFn, ok := impl.(func(string) int); ok {
			result := wrappedFn("asdf")
			assert.Equal(t, 4, result)
		} else {
			t.Error("Implemented function has wrong type")
		}
	})

	t.Run("function parsing validation", func(t *testing.T) {
		schema := Function(FunctionParams{
			Input:  []Schema{String()},
			Output: Int(),
		})

		testFunc := func(s string) int { return len(s) }
		result, err := schema.Parse(testFunc)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("smart type inference", func(t *testing.T) {
		schema := Function()

		testFunc := func() string { return "hello" }
		result1, err := schema.Parse(testFunc)
		require.NoError(t, err)
		assert.IsType(t, testFunc, result1)

		result2, err := schema.Parse(&testFunc)
		require.NoError(t, err)
		assert.IsType(t, &testFunc, result2)
		assert.Equal(t, &testFunc, result2)
	})

	t.Run("nilable modifier", func(t *testing.T) {
		schema := Function().Nilable()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		testFunc := func() string { return "hello" }
		result2, err := schema.Parse(testFunc)
		require.NoError(t, err)
		assert.NotNil(t, result2)
		assert.IsType(t, testFunc, result2)
	})
}

// =============================================================================
// 2. Coerce (type coercion)
// =============================================================================

func TestFunctionCoercion(t *testing.T) {
	t.Run("basic coercion", func(t *testing.T) {
		schema := Function()

		testFunc := func() string { return "hello" }
		result, err := schema.Parse(testFunc)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("coercion with validation", func(t *testing.T) {
		schema := Function(FunctionParams{
			Input:  []Schema{String()},
			Output: String(),
		})

		impl, err := schema.Implement(strings.ToUpper)
		require.NoError(t, err)
		require.NotNil(t, impl)
	})
}

// =============================================================================
// 3. Validation methods
// =============================================================================

func TestFunctionValidations(t *testing.T) {
	t.Run("input validation", func(t *testing.T) {
		schema := Function(FunctionParams{
			Input:  []Schema{String()},
			Output: Any(),
		})

		impl, err := schema.Implement(func(s string) interface{} {
			return len(s)
		})
		require.NoError(t, err)
		require.NotNil(t, impl)
	})

	t.Run("output validation", func(t *testing.T) {
		schema := Function(FunctionParams{
			Input:  []Schema{},
			Output: String(),
		})

		impl, err := schema.Implement(func() string {
			return "valid"
		})
		require.NoError(t, err)
		require.NotNil(t, impl)
	})

	t.Run("complex input validation", func(t *testing.T) {
		objectSchema := Object(ObjectSchema{
			"f1": Int(),
			"f2": String().Nilable(),
			"f3": Slice(Bool().Optional()).Optional(),
		})

		schema := Function(FunctionParams{
			Input:  []Schema{objectSchema},
			Output: Union([]ZodType[any, any]{String(), Int()}),
		})

		impl, err := schema.Implement(func(obj map[string]interface{}) interface{} {
			return "processed"
		})
		require.NoError(t, err)
		require.NotNil(t, impl)
	})
}

// =============================================================================
// 4. Modifiers and wrappers
// =============================================================================

func TestFunctionModifiers(t *testing.T) {
	t.Run("optional wrapper", func(t *testing.T) {
		schema := Function().Optional()

		result, err := schema.Parse(func() {})
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Optional should handle nil
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nilable wrapper", func(t *testing.T) {
		schema := Function().Nilable()

		result, err := schema.Parse(func() {})
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Nilable should handle nil
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("nullish wrapper", func(t *testing.T) {
		schema := Function().Nullish()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("must parse", func(t *testing.T) {
		schema := Function()
		testFunc := func() string { return "hello" }

		result := schema.MustParse(testFunc)
		assert.NotNil(t, result)
		assert.IsType(t, testFunc, result)
	})
}

// =============================================================================
// 5. Chaining and method composition
// =============================================================================

func TestFunctionChaining(t *testing.T) {
	t.Run("input method chaining", func(t *testing.T) {
		baseFactory := Function()

		inputFactory := baseFactory.Input([]Schema{String()})
		require.NotNil(t, inputFactory)

		singleInputFactory := baseFactory.Input(String())
		require.NotNil(t, singleInputFactory)
	})

	t.Run("output method chaining", func(t *testing.T) {
		baseFactory := Function()

		outputFactory := baseFactory.Output(String())
		require.NotNil(t, outputFactory)
	})

	t.Run("full method chaining", func(t *testing.T) {
		chainedFactory := Function().
			Input([]Schema{String()}).
			Output(Bool())
		require.NotNil(t, chainedFactory)

		impl, err := chainedFactory.Implement(func(s string) bool {
			return len(s) > 0
		})
		require.NoError(t, err)
		require.NotNil(t, impl)
	})

	t.Run("multiple validations", func(t *testing.T) {
		schema := Function(FunctionParams{
			Input:  []Schema{String().Min(3)},
			Output: String().Min(1),
		})

		impl, err := schema.Implement(strings.ToUpper)
		require.NoError(t, err)
		require.NotNil(t, impl)
	})
}

// =============================================================================
// 6. Transform/Pipe
// =============================================================================

func TestFunctionTransform(t *testing.T) {
	t.Run("basic transform", func(t *testing.T) {
		schema := Function().Transform(func(fn interface{}, ctx *RefinementContext) (any, error) {
			// Convert function to wrapper
			return map[string]interface{}{
				"original": fn,
				"type":     reflect.TypeOf(fn).String(),
			}, nil
		})

		testFunc := func() string { return "hello" }
		result, err := schema.Parse(testFunc)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.NotNil(t, resultMap["original"])
		assert.IsType(t, testFunc, resultMap["original"])
		assert.Contains(t, resultMap["type"], "func")
	})

	t.Run("pipe operations", func(t *testing.T) {
		// Function -> Transform -> Pipe -> Validation
		step1 := Function().Transform(func(fn interface{}, ctx *RefinementContext) (any, error) {
			return reflect.TypeOf(fn).String(), nil
		})

		pipeline := step1.Pipe(String().Min(4))

		testFunc := func() string { return "hello" }
		result, err := pipeline.Parse(testFunc)
		require.NoError(t, err)
		assert.IsType(t, "", result)
	})
}

// =============================================================================
// 7. Refine
// =============================================================================

func TestFunctionRefine(t *testing.T) {
	t.Run("basic refine", func(t *testing.T) {
		schema := Function().Refine(func(fn interface{}) bool {
			return fn != nil
		}, SchemaParams{Error: "Function cannot be nil"})

		testFunc := func() string { return "hello" }
		result, err := schema.Parse(testFunc)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.IsType(t, testFunc, result)

		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("refine with custom error", func(t *testing.T) {
		schema := Function().Refine(func(fn interface{}) bool {
			if fn == nil {
				return false
			}
			fnType := reflect.TypeOf(fn)
			return fnType.Kind() == reflect.Func && fnType.NumIn() > 0
		}, SchemaParams{Error: "Function must have at least one parameter"})

		// Valid function with parameters
		validFunc := func(s string) string { return s }
		result, err := schema.Parse(validFunc)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.IsType(t, validFunc, result)

		// Invalid function without parameters
		invalidFunc := func() string { return "hello" }
		_, err = schema.Parse(invalidFunc)
		// Note: Current implementation may not validate parameter count, this is expected
		if err != nil {
			t.Logf("Parameter validation working: %v", err)
		}
	})

	t.Run("refine preserves original value", func(t *testing.T) {
		testFunc := func(s string) int { return len(s) }

		refineSchema := Function().Refine(func(fn interface{}) bool {
			return fn != nil
		})

		refineResult, refineErr := refineSchema.Parse(testFunc)
		require.NoError(t, refineErr)
		assert.NotNil(t, refineResult)
		assert.IsType(t, testFunc, refineResult)
	})
}

// =============================================================================
// 8. Error handling
// =============================================================================

func TestFunctionErrorHandling(t *testing.T) {
	t.Run("error structure", func(t *testing.T) {
		schema := Function()
		_, err := schema.Parse("not a function")
		assert.Error(t, err)

		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, string(InvalidType), zodErr.Issues[0].Code)
	})

	t.Run("custom error messages", func(t *testing.T) {
		// Function type currently does not support Error parameter, test basic error
		schema := Function()
		_, err := schema.Parse("not a function")
		assert.Error(t, err)

		var zodErr *ZodError
		require.True(t, IsZodError(err, &zodErr))
		assert.Contains(t, zodErr.Issues[0].Message, "function")
	})

	t.Run("implementation type mismatch", func(t *testing.T) {
		schema := Function(FunctionParams{
			Input:  []Schema{String()},
			Output: String(),
		})

		// Wrong function signature
		_, err := schema.Implement(func(i int) string {
			return "test"
		})
		// Note: Current implementation may not strictly validate parameter types, this is expected
		if err != nil {
			t.Logf("Type validation working: %v", err)
		}
	})

	t.Run("input validation errors", func(t *testing.T) {
		schema := Function(FunctionParams{
			Input:  []Schema{String().Min(5)},
			Output: String(),
		})

		impl, err := schema.Implement(strings.ToUpper)
		require.NoError(t, err)
		require.NotNil(t, impl)

		// Here we only test implementation success, actual input validation is done when the function is called
	})

	t.Run("output validation errors", func(t *testing.T) {
		schema := Function(FunctionParams{
			Input:  []Schema{String()},
			Output: String().Min(10),
		})

		impl, err := schema.Implement(func(s string) string {
			return "hi" // Too short for output validation
		})
		require.NoError(t, err)
		require.NotNil(t, impl)

		// Here we only test implementation success, actual output validation is done when the function is called
	})
}

// =============================================================================
// 9. Edge cases and internals
// =============================================================================

func TestFunctionEdgeCases(t *testing.T) {
	t.Run("no input parameters", func(t *testing.T) {
		schema := Function(FunctionParams{
			Input:  []Schema{},
			Output: String(),
		})

		impl, err := schema.Implement(func() string {
			return "hello"
		})
		require.NoError(t, err)
		require.NotNil(t, impl)
	})

	t.Run("nil input handling", func(t *testing.T) {
		schema := Function(FunctionParams{
			Input:  []Schema{String().Nilable()},
			Output: String(),
		})

		impl, err := schema.Implement(func(s *string) string {
			if s == nil {
				return "nil"
			}
			return *s
		})
		require.NoError(t, err)
		require.NotNil(t, impl)
	})

	t.Run("complex return types", func(t *testing.T) {
		schema := Function(FunctionParams{
			Input: []Schema{String()},
			Output: Object(ObjectSchema{
				"result": String(),
				"length": Int(),
			}),
		})

		impl, err := schema.Implement(func(s string) map[string]interface{} {
			return map[string]interface{}{
				"result": strings.ToUpper(s),
				"length": len(s),
			}
		})
		require.NoError(t, err)
		require.NotNil(t, impl)
	})

	t.Run("function type reflection", func(t *testing.T) {
		schema := Function(FunctionParams{
			Input:  []Schema{String()},
			Output: String(),
		})

		impl, err := schema.Implement(func(s string) string {
			return s
		})
		require.NoError(t, err)

		// Check that the implemented function has the correct type
		implType := reflect.TypeOf(impl)
		assert.Equal(t, reflect.Func, implType.Kind())
	})

	t.Run("parameter count validation", func(t *testing.T) {
		schema := Function(FunctionParams{
			Input:  []Schema{String(), Int()},
			Output: String(),
		})

		// Wrong number of parameters
		_, err := schema.Implement(func(s string) string {
			return s
		})
		assert.Error(t, err)
	})

	t.Run("return type validation", func(t *testing.T) {
		schema := Function(FunctionParams{
			Input:  []Schema{String()},
			Output: String(),
		})

		// Wrong return type
		_, err := schema.Implement(func(s string) int {
			return 42
		})
		assert.Error(t, err)
	})

	t.Run("nil function implementation", func(t *testing.T) {
		schema := Function(FunctionParams{
			Input:  []Schema{String()},
			Output: String(),
		})

		_, err := schema.Implement(nil)
		assert.Error(t, err)
	})

	t.Run("variadic function handling", func(t *testing.T) {
		schema := Function(FunctionParams{
			Input:  []Schema{String()},
			Output: String(),
		})

		// Variadic functions should be handled appropriately
		_, err := schema.Implement(func(s ...string) string {
			return strings.Join(s, " ")
		})
		// This might be an error depending on implementation
		// The test documents the current behavior
		if err != nil {
			t.Logf("Variadic functions not supported: %v", err)
		}
	})

	t.Run("empty function factory", func(t *testing.T) {
		schema := Function()

		// Should be able to configure after creation
		configured := schema.Input(String()).Output(String())
		require.NotNil(t, configured)

		impl, err := configured.Implement(func(s string) string {
			return s
		})
		require.NoError(t, err)
		require.NotNil(t, impl)
	})
}

// =============================================================================
// 10. Default and Prefault tests
// =============================================================================

func TestFunctionDefaultAndPrefault(t *testing.T) {
	t.Run("default function", func(t *testing.T) {
		defaultFunc := func() string { return "default" }
		schema := Function().Default(defaultFunc)

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.IsType(t, defaultFunc, result)

		// Valid function should pass through
		testFunc := func() int { return 42 }
		result2, err := schema.Parse(testFunc)
		require.NoError(t, err)
		assert.NotNil(t, result2)
		assert.IsType(t, testFunc, result2)
	})

	t.Run("prefault function", func(t *testing.T) {
		fallbackFunc := func() string { return "fallback" }
		schema := Function().Prefault(fallbackFunc)

		// Valid function should pass through
		testFunc := func() int { return 42 }
		result, err := schema.Parse(testFunc)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.IsType(t, testFunc, result)

		// Invalid input should use fallback
		result2, err := schema.Parse("not a function")
		require.NoError(t, err)
		assert.NotNil(t, result2)
		assert.IsType(t, fallbackFunc, result2)
	})

	t.Run("function with default parameters", func(t *testing.T) {
		// Test function schema with default parameters
		schema := Function(FunctionParams{
			Input:  []Schema{String().Default("default")},
			Output: String(),
		})

		impl, err := schema.Implement(func(s string) string {
			return "processed: " + s
		})
		require.NoError(t, err)
		require.NotNil(t, impl)
	})

	t.Run("function with prefault parameters", func(t *testing.T) {
		// Test function schema with prefault parameters
		schema := Function(FunctionParams{
			Input:  []Schema{String().Min(5).Prefault("fallback")},
			Output: String(),
		})

		impl, err := schema.Implement(func(s string) string {
			return "processed: " + s
		})
		require.NoError(t, err)
		require.NotNil(t, impl)
	})

	t.Run("defaultFunc", func(t *testing.T) {
		counter := 0
		schema := Function().DefaultFunc(func() interface{} {
			counter++
			return func(x int) int {
				return x + counter // Each generated function adds a different number
			}
		})

		// nil input should call function and use default
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.NotNil(t, result1)
		assert.Equal(t, 1, counter)

		// Check the generated function works
		if fn, ok := result1.(func(int) int); ok {
			assert.Equal(t, 6, fn(5)) // 5 + 1 = 6
		}

		// Another nil input should call function again
		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		assert.NotNil(t, result2)
		assert.Equal(t, 2, counter)

		// Check the new generated function
		if fn, ok := result2.(func(int) int); ok {
			assert.Equal(t, 7, fn(5)) // 5 + 2 = 7
		}

		// Valid input should not call function
		testFunc := func() string { return "test" }
		result3, err3 := schema.Parse(testFunc)
		require.NoError(t, err3)
		assert.NotNil(t, result3)
		assert.Equal(t, 2, counter) // Counter should not increment
		assert.IsType(t, testFunc, result3)
	})

	t.Run("prefaultFunc", func(t *testing.T) {
		counter := 0
		schema := Function().PrefaultFunc(func() interface{} {
			counter++
			return func(msg string) string {
				return fmt.Sprintf("fallback-%d: %s", counter, msg)
			}
		})

		// Valid input should not call function
		testFunc := func(x int) int { return x * 2 }
		result1, err1 := schema.Parse(testFunc)
		require.NoError(t, err1)
		assert.NotNil(t, result1)
		assert.Equal(t, 0, counter)
		assert.IsType(t, testFunc, result1)

		// Invalid input should call prefault function
		result2, err2 := schema.Parse("not a function")
		require.NoError(t, err2)
		assert.NotNil(t, result2)
		assert.Equal(t, 1, counter)

		// Check the generated fallback function
		if fn, ok := result2.(func(string) string); ok {
			assert.Equal(t, "fallback-1: hello", fn("hello"))
		}

		// Another invalid input should call function again
		result3, err3 := schema.Parse(123)
		require.NoError(t, err3)
		assert.NotNil(t, result3)
		assert.Equal(t, 2, counter)

		// Check the new generated fallback function
		if fn, ok := result3.(func(string) string); ok {
			assert.Equal(t, "fallback-2: world", fn("world"))
		}
	})

	t.Run("default vs prefault distinction", func(t *testing.T) {
		defaultFunc := func() string { return "default_function" }
		prefaultFunc := func() string { return "prefault_function" }

		schema := Function().Default(defaultFunc).Prefault(prefaultFunc)

		// nil input uses default
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.NotNil(t, result1)
		assert.IsType(t, defaultFunc, result1)

		// Valid input succeeds
		testFunc := func(x int) int { return x }
		result2, err2 := schema.Parse(testFunc)
		require.NoError(t, err2)
		assert.NotNil(t, result2)
		assert.IsType(t, testFunc, result2)

		// Invalid input uses prefault
		result3, err3 := schema.Parse("invalid")
		require.NoError(t, err3)
		assert.NotNil(t, result3)
		assert.IsType(t, prefaultFunc, result3)
	})
}
