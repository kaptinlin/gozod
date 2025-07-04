package engine

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
)

// =============================================================================
// MOCK SCHEMA FOR TESTING
// =============================================================================

// mockStringSchema represents a simple string validation schema for testing
type mockStringSchema struct {
	internals *core.ZodTypeInternals
}

func (m *mockStringSchema) GetInternals() *core.ZodTypeInternals {
	return m.internals
}

func (m *mockStringSchema) Parse(input any, ctx ...*core.ParseContext) (any, error) {
	// Use the engine Parse function
	var context *core.ParseContext = nil
	if len(ctx) > 0 {
		context = ctx[0]
	}

	// Check if schema has internals
	if m.internals == nil {
		rawIssue := issues.CreateCustomIssue("schema has no internals defined", nil, input)
		finalIssue := issues.FinalizeIssue(rawIssue, context, nil)
		return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
	}

	// Check if internals has a parse function
	if m.internals.Parse == nil {
		rawIssue := issues.CreateCustomIssue("schema has no parse function defined", nil, input)
		finalIssue := issues.FinalizeIssue(rawIssue, context, nil)
		return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
	}

	// create parse payload
	payload := core.NewParsePayload(input)

	// call internal parse function
	result := m.internals.Parse(payload, context)

	// check if parse result is nil
	if result == nil {
		rawIssue := issues.CreateCustomIssue("parse function returned nil", nil, input)
		finalIssue := issues.FinalizeIssue(rawIssue, context, nil)
		return nil, issues.NewZodError([]core.ZodIssue{finalIssue})
	}

	// check if there are any issues
	if len(result.GetIssues()) > 0 {
		// ensure all raw issues have correct input field
		resultIssues := result.GetIssues()
		for i := range resultIssues {
			if resultIssues[i].Input == nil {
				resultIssues[i].Input = input
			}
		}
		result.SetIssues(resultIssues)
		finalizedIssues := issues.ConvertRawIssuesToIssues(result.GetIssues(), context)
		return nil, issues.NewZodError(finalizedIssues)
	}

	return result.GetValue(), nil
}

func (m *mockStringSchema) MustParse(input any, ctx ...*core.ParseContext) any {
	// Use the engine MustParse function
	var context *core.ParseContext = nil
	if len(ctx) > 0 {
		context = ctx[0]
	}
	return MustParse[any, any](m, input, context)
}

func (m *mockStringSchema) Nilable() core.ZodType[any, any] {
	// Create a copy with nilable flag
	newInternals := *m.internals
	newInternals.Nilable = true
	return &mockStringSchema{internals: &newInternals}
}

func (m *mockStringSchema) RefineAny(fn func(any) bool, params ...any) core.ZodType[any, any] {
	return m // Simple mock implementation
}

func (m *mockStringSchema) TransformAny(fn func(any, *core.RefinementContext) (any, error)) core.ZodType[any, any] {
	return m // Simple mock implementation
}

func (m *mockStringSchema) Pipe(out core.ZodType[any, any]) core.ZodType[any, any] {
	return out // Simple mock implementation
}

func (m *mockStringSchema) Unwrap() core.ZodType[any, any] {
	return m
}

func newMockStringSchema() *mockStringSchema {
	return &mockStringSchema{
		internals: &core.ZodTypeInternals{
			Type:   "string",
			Checks: []core.ZodCheck{},
			Parse: func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
				// Simple string validation
				if str, ok := payload.GetValue().(string); ok {
					result := core.NewParsePayload(str)
					result.SetPath(payload.GetPath())
					return result
				}
				// Create type error
				rawIssue := issues.CreateInvalidTypeIssue("string", payload.GetValue())
				result := core.NewParsePayload(payload.GetValue())
				result.AddIssue(rawIssue)
				result.SetPath(payload.GetPath())
				return result
			},
		},
	}
}

// =============================================================================
// BASIC PARSING FUNCTIONALITY TESTS
// =============================================================================

func TestParseBasicFunctionality(t *testing.T) {
	t.Run("Parse function with valid input", func(t *testing.T) {
		schema := newMockStringSchema()
		result, err := Parse[any, any](schema, "test", nil)

		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("Parse function with invalid input", func(t *testing.T) {
		schema := newMockStringSchema()
		_, err := Parse[any, any](schema, 123, nil)

		require.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Equal(t, core.InvalidType, zodErr.Issues[0].Code)
	})

	t.Run("MustParse with valid input", func(t *testing.T) {
		schema := newMockStringSchema()
		result := MustParse[any, any](schema, "test", nil)

		assert.Equal(t, "test", result)
	})

	t.Run("MustParse panics with invalid input", func(t *testing.T) {
		schema := newMockStringSchema()

		assert.Panics(t, func() {
			MustParse[any, any](schema, 123, nil)
		})
	})
}

// =============================================================================
// PARSE CONTEXT TESTS
// =============================================================================

func TestParseContextFunctionality(t *testing.T) {
	t.Run("Parse with default context", func(t *testing.T) {
		ctx := core.NewParseContext()
		schema := newMockStringSchema()

		result, err := Parse[any, any](schema, "test", ctx)

		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("Parse with custom context", func(t *testing.T) {
		customErrorMap := core.ZodErrorMap(func(issue core.ZodRawIssue) string {
			return "custom error"
		})

		ctx := &core.ParseContext{
			Error:       customErrorMap,
			ReportInput: true,
		}

		schema := newMockStringSchema()
		_, err := Parse[any, any](schema, 123, ctx)

		require.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, errors.As(err, &zodErr))

		// Note: Custom error map would be applied during finalization
		assert.Equal(t, 123, zodErr.Issues[0].Input)
	})
}

// =============================================================================
// INTERNAL HELPER TESTS
// =============================================================================

func TestInternalHelpers(t *testing.T) {
	t.Run("coercion configuration through bag", func(t *testing.T) {
		// Test coercion behavior indirectly through schema configuration

		// Schema without coercion
		schema1 := &mockStringSchema{
			internals: &core.ZodTypeInternals{
				Type: "string",
				Bag:  map[string]any{},
				Parse: func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
					// Mock parse that respects coercion setting
					if str, ok := payload.GetValue().(string); ok {
						result := core.NewParsePayload(str)
						result.SetPath(payload.GetPath())
						return result
					}
					rawIssue := issues.CreateInvalidTypeIssue("string", payload.GetValue())
					result := core.NewParsePayload(payload.GetValue())
					result.AddIssue(rawIssue)
					result.SetPath(payload.GetPath())
					return result
				},
			},
		}

		// Test that coercion flag can be set in bag
		assert.NotNil(t, schema1.GetInternals().Bag)
		_, exists := schema1.GetInternals().Bag["coerce"]
		assert.False(t, exists)

		// Schema with coercion enabled
		schema2 := &mockStringSchema{
			internals: &core.ZodTypeInternals{
				Type: "string",
				Bag:  map[string]any{"coerce": true},
				Parse: func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
					result := core.NewParsePayload(payload.GetValue())
					result.SetPath(payload.GetPath())
					return result
				},
			},
		}

		// Test that coercion flag is properly set
		coerceFlag, exists := schema2.GetInternals().Bag["coerce"]
		assert.True(t, exists)
		assert.True(t, coerceFlag.(bool))
	})

	t.Run("bag configuration edge cases", func(t *testing.T) {
		// Test nil bag
		schema := &mockStringSchema{
			internals: &core.ZodTypeInternals{
				Type: "string",
				Bag:  nil,
			},
		}
		assert.Nil(t, schema.GetInternals().Bag)

		// Test empty bag
		schema2 := &mockStringSchema{
			internals: &core.ZodTypeInternals{
				Type: "string",
				Bag:  map[string]any{},
			},
		}
		assert.NotNil(t, schema2.GetInternals().Bag)
		assert.Empty(t, schema2.GetInternals().Bag)
	})
}

// =============================================================================
// ERROR HANDLING TESTS
// =============================================================================

func TestParseErrorHandling(t *testing.T) {
	t.Run("Parse returns error for nil internals", func(t *testing.T) {
		schema := &mockStringSchema{internals: nil}
		_, err := Parse[any, any](schema, "test", nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "schema has no internals defined")
	})

	t.Run("Parse returns error for nil parse function", func(t *testing.T) {
		schema := &mockStringSchema{
			internals: &core.ZodTypeInternals{
				Type:  "string",
				Parse: nil,
			},
		}
		_, err := Parse[any, any](schema, "test", nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "schema has no parse function defined")
	})

	t.Run("Parse returns error when parse function returns nil", func(t *testing.T) {
		schema := &mockStringSchema{
			internals: &core.ZodTypeInternals{
				Type: "string",
				Parse: func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
					return nil
				},
			},
		}
		_, err := Parse[any, any](schema, "test", nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "parse function returned nil")
	})

	t.Run("Parse handles type assertion failure", func(t *testing.T) {
		schema := &mockStringSchema{
			internals: &core.ZodTypeInternals{
				Type: "string",
				Parse: func(payload *core.ParsePayload, ctx *core.ParseContext) *core.ParsePayload {
					// Return wrong type
					result := core.NewParsePayload(123) // int instead of string
					result.SetPath(payload.GetPath())
					return result
				},
			},
		}
		_, err := Parse[any, string](schema, "test", nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "type assertion failed")
	})
}

// =============================================================================
// EDGE CASES AND SPECIAL VALUES
// =============================================================================

func TestParseEdgeCases(t *testing.T) {
	t.Run("Parse handles nil input correctly", func(t *testing.T) {
		schema := newMockStringSchema()
		_, err := Parse[any, any](schema, nil, nil)

		// Should return validation error for nil input on non-nilable schema
		require.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Equal(t, core.InvalidType, zodErr.Issues[0].Code)
	})

	t.Run("Parse with nil context works", func(t *testing.T) {
		schema := newMockStringSchema()
		result, err := Parse[any, any](schema, "test", nil)

		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("Parse handles empty string", func(t *testing.T) {
		schema := newMockStringSchema()
		result, err := Parse[any, any](schema, "", nil)

		require.NoError(t, err)
		assert.Equal(t, "", result)
	})
}

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

func TestParseIntegration(t *testing.T) {
	t.Run("complete parsing workflow", func(t *testing.T) {
		schema := newMockStringSchema()
		ctx := core.NewParseContext()

		// Valid case
		result, err := Parse[any, any](schema, "test", ctx)
		require.NoError(t, err)
		assert.Equal(t, "test", result)

		// Invalid case
		_, err = Parse[any, any](schema, 123, ctx)
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Equal(t, core.InvalidType, zodErr.Issues[0].Code)
	})

	t.Run("MustParse in initialization context", func(t *testing.T) {
		schema := newMockStringSchema()

		// Should succeed
		result := MustParse[any, any](schema, "test", nil)
		assert.Equal(t, "test", result)

		// Should panic
		assert.Panics(t, func() {
			MustParse[any, any](schema, 123, nil)
		})
	})
}
