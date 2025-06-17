package gozod

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestParseBasicFunctionality(t *testing.T) {
	t.Run("Parse function with valid input", func(t *testing.T) {
		schema := String()
		result, err := schema.Parse("test")

		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("Parse function with invalid input", func(t *testing.T) {
		schema := String()
		_, err := schema.Parse(123)

		require.Error(t, err)
		var zodErr *ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Equal(t, string(InvalidType), zodErr.Issues[0].Code)
	})

	t.Run("MustParse with valid input", func(t *testing.T) {
		schema := String()
		result := schema.MustParse("test")

		assert.Equal(t, "test", result)
	})

	t.Run("MustParse panics with invalid input", func(t *testing.T) {
		schema := String()

		assert.Panics(t, func() {
			schema.MustParse(123)
		})
	})
}

// =============================================================================
// 2. ParseContext functionality
// =============================================================================

func TestParseContext(t *testing.T) {
	t.Run("default ParseContext", func(t *testing.T) {
		ctx := NewParseContext()

		require.NotNil(t, ctx)
		assert.False(t, ctx.ReportInput)
		assert.Nil(t, ctx.Error)
	})

	t.Run("custom ParseContext", func(t *testing.T) {
		customErrorMap := ZodErrorMap(func(issue ZodRawIssue) string {
			return "custom error"
		})

		ctx := &ParseContext{
			Error:       customErrorMap,
			ReportInput: true,
		}

		require.NotNil(t, ctx)
		assert.True(t, ctx.ReportInput)
		assert.NotNil(t, ctx.Error)

		// Test custom error map
		issue := ZodRawIssue{Code: "test", Message: "original"}
		result := ctx.Error(issue)
		assert.Equal(t, "custom error", result)
	})

	t.Run("ParseContext with ReportInput", func(t *testing.T) {
		ctx := &ParseContext{ReportInput: true}
		schema := String()

		_, err := schema.Parse(123, ctx)

		require.Error(t, err)
		var zodErr *ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Equal(t, 123, zodErr.Issues[0].Input)
	})
}

// =============================================================================
// 3. ParsePayload functionality
// =============================================================================

func TestParsePayload(t *testing.T) {
	t.Run("basic payload creation", func(t *testing.T) {
		payload := NewParsePayload("test")

		require.NotNil(t, payload)
		assert.Equal(t, "test", payload.Value)
		assert.NotNil(t, payload.Issues)
		assert.Equal(t, 0, len(payload.Issues))
		assert.False(t, payload.HasIssues())
	})

	t.Run("issue management", func(t *testing.T) {
		payload := NewParsePayload("test")

		// Initially no issues
		assert.False(t, payload.HasIssues())

		// Add issue
		issue := ZodRawIssue{Code: "test_error", Message: "Test error"}
		payload.AddIssue(issue)

		assert.True(t, payload.HasIssues())
		assert.Equal(t, 1, len(payload.Issues))
		assert.Equal(t, "test_error", payload.Issues[0].Code)
	})

	t.Run("different value types", func(t *testing.T) {
		testCases := []interface{}{
			"string",
			42,
			true,
			[]int{1, 2, 3},
			map[string]int{"key": 1},
			nil,
		}

		for _, testValue := range testCases {
			payload := NewParsePayload(testValue)
			assert.Equal(t, testValue, payload.Value)
		}
	})
}

// =============================================================================
// 4. Validation chains
// =============================================================================

func TestParseValidationChains(t *testing.T) {
	t.Run("string validation chain", func(t *testing.T) {
		schema := String().Min(3).Max(10)

		// Valid input
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Invalid input (too short)
		_, err = schema.Parse("hi")
		require.Error(t, err)
		var zodErr *ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Equal(t, string(TooSmall), zodErr.Issues[0].Code)

		// Invalid input (too long)
		_, err = schema.Parse("this is too long")
		require.Error(t, err)
		require.True(t, errors.As(err, &zodErr))
		assert.Equal(t, string(TooBig), zodErr.Issues[0].Code)
	})

	t.Run("email validation", func(t *testing.T) {
		schema := String().Email()

		// Valid email
		result, err := schema.Parse("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", result)

		// Invalid email
		_, err = schema.Parse("invalid-email")
		require.Error(t, err)
	})
}

// =============================================================================
// 5. Error handling and ZodError
// =============================================================================

func TestParseErrorHandling(t *testing.T) {
	t.Run("ZodError structure", func(t *testing.T) {
		schema := String()
		_, err := schema.Parse(123)

		require.Error(t, err)
		var zodErr *ZodError
		require.True(t, errors.As(err, &zodErr))

		assert.Equal(t, "ZodError", zodErr.Name)
		require.Len(t, zodErr.Issues, 1)
		assert.Equal(t, string(InvalidType), zodErr.Issues[0].Code)
	})

	t.Run("issue properties", func(t *testing.T) {
		schema := String().Min(5)
		_, err := schema.Parse("hi")

		require.Error(t, err)
		var zodErr *ZodError
		require.True(t, errors.As(err, &zodErr))

		issue := zodErr.Issues[0]
		minimum, hasMinimum := issue.GetMinimum()
		assert.True(t, hasMinimum)
		assert.Equal(t, 5, minimum)
	})

	t.Run("custom error messages", func(t *testing.T) {
		ctx := &ParseContext{
			Error: func(issue ZodRawIssue) string {
				switch issue.Code {
				case string(TooSmall):
					return "Input is too short"
				default:
					return "Validation failed"
				}
			},
		}

		schema := String().Min(5)
		_, err := schema.Parse("hi", ctx)

		require.Error(t, err)
		var zodErr *ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Equal(t, "Input is too short", zodErr.Issues[0].Message)
	})
}

// =============================================================================
// 6. Complex data structures
// =============================================================================

func TestParseComplexStructures(t *testing.T) {
	t.Run("object validation", func(t *testing.T) {
		data := map[string]interface{}{
			"name":  "John",
			"age":   30,
			"email": "john@example.com",
		}

		schema := Object(map[string]ZodType[any, any]{
			"name":  String(),
			"age":   Int(),
			"email": String().Email(),
		})

		result, err := schema.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, data, result)
	})

	t.Run("array validation", func(t *testing.T) {
		data := []string{"hello", "world", "test"}
		schema := Slice(String())

		result, err := schema.Parse(data)
		require.NoError(t, err)
		expected := []interface{}{"hello", "world", "test"}
		assert.Equal(t, expected, result)
	})

	t.Run("nested structures", func(t *testing.T) {
		data := map[string]interface{}{
			"user": map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			"tags": []string{"admin", "user"},
		}

		schema := Object(map[string]ZodType[any, any]{
			"user": Object(map[string]ZodType[any, any]{
				"name": String(),
				"age":  Int(),
			}),
			"tags": Slice(String()),
		})

		result, err := schema.Parse(data)
		require.NoError(t, err)
		expected := map[string]interface{}{
			"user": map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			"tags": []interface{}{"admin", "user"},
		}
		assert.Equal(t, expected, result)
	})
}

// =============================================================================
// 7. Edge cases and special values
// =============================================================================

func TestParseEdgeCases(t *testing.T) {
	t.Run("nil input with nilable schema", func(t *testing.T) {
		schema := String().Nilable()
		result, err := schema.Parse(nil)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("empty string", func(t *testing.T) {
		schema := String()
		result, err := schema.Parse("")

		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("nil ParseContext", func(t *testing.T) {
		schema := String()
		result, err := schema.Parse("test", nil)

		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		schema := String().Min(10).Email()
		_, err := schema.Parse("hi")

		require.Error(t, err)
		var zodErr *ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Greater(t, len(zodErr.Issues), 0)
	})
}

// =============================================================================
// 8. Performance and optimization
// =============================================================================

func TestParsePerformance(t *testing.T) {
	t.Run("simple validation performance", func(t *testing.T) {
		schema := String()

		// This is more of a smoke test than a real performance test
		for i := 0; i < 100; i++ {
			result, err := schema.Parse("test")
			require.NoError(t, err)
			assert.Equal(t, "test", result)
		}
	})

	t.Run("complex validation performance", func(t *testing.T) {
		schema := Object(map[string]ZodType[any, any]{
			"name":  String().Min(1).Max(50),
			"email": String().Email(),
			"age":   Int().Min(0).Max(120),
		})

		data := map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   30,
		}

		// Smoke test for complex validation
		for i := 0; i < 50; i++ {
			result, err := schema.Parse(data)
			require.NoError(t, err)
			assert.Equal(t, data, result)
		}
	})
}

// =============================================================================
// 9. Integration and workflow tests
// =============================================================================

func TestParseIntegration(t *testing.T) {
	t.Run("complete validation workflow", func(t *testing.T) {
		// Clear global config for clean test
		originalConfig := globalConfig
		defer func() { globalConfig = originalConfig }()
		Config(nil)

		schema := String().Min(3).Max(20).Email()

		// Valid case
		result, err := schema.Parse("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", result)

		// Invalid case
		_, err = schema.Parse("ab")
		require.Error(t, err)
		var zodErr *ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Equal(t, string(TooSmall), zodErr.Issues[0].Code)
	})

	t.Run("MustParse in initialization context", func(t *testing.T) {
		schema := String().Email()

		// Should succeed
		result := schema.MustParse("admin@example.com")
		assert.Equal(t, "admin@example.com", result)

		// Should panic
		assert.Panics(t, func() {
			schema.MustParse("invalid-email")
		})
	})

	t.Run("error customization workflow", func(t *testing.T) {
		ctx := &ParseContext{
			Error: func(issue ZodRawIssue) string {
				return "Custom validation error"
			},
			ReportInput: true,
		}

		schema := String().Min(5)
		_, err := schema.Parse("hi", ctx)

		require.Error(t, err)
		var zodErr *ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Equal(t, "Custom validation error", zodErr.Issues[0].Message)
		assert.Equal(t, "hi", zodErr.Issues[0].Input)
	})

	t.Run("schema composition", func(t *testing.T) {
		userSchema := Object(map[string]ZodType[any, any]{
			"name":  String().Min(1),
			"email": String().Email(),
		})

		postSchema := Object(map[string]ZodType[any, any]{
			"title":  String().Min(1),
			"author": userSchema,
		})

		data := map[string]interface{}{
			"title": "Test Post",
			"author": map[string]interface{}{
				"name":  "John",
				"email": "john@example.com",
			},
		}

		result, err := postSchema.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, data, result)
	})
}
