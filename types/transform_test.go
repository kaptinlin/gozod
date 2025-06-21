package types

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Error definitions for testing transformations
var (
	ErrEmptyString           = errors.New("empty string not allowed")
	ErrTransformationFailed  = errors.New("transformation failed")
	ErrTransformFailed       = errors.New("transform failed")
	ErrExpectedStringInput   = errors.New("expected string input")
	ErrCleanedStringTooShort = errors.New("cleaned string too short")
	ErrExpectedStringType    = errors.New("expected string type")
	ErrContainsInvalidWord   = errors.New("contains invalid word")
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestTransformBasicFunctionality(t *testing.T) {
	t.Run("Transform constructor availability", func(t *testing.T) {
		// Test global Transform function
		schema := Transform(func(input any, ctx *core.RefinementContext) (any, error) {
			if str, ok := input.(string); ok {
				return strings.ToUpper(str), nil
			}
			return input, nil
		})

		require.NotNil(t, schema)
		internals := schema.GetInternals()
		require.NotNil(t, internals)
		assert.Equal(t, "transform", internals.Type)
	})

	t.Run("string Transform method", func(t *testing.T) {
		baseSchema := String()
		transformSchema := baseSchema.Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return len(s), nil
		})

		require.NotNil(t, transformSchema)
		internals := transformSchema.GetInternals()
		require.NotNil(t, internals)
		assert.Equal(t, "pipe", internals.Type) // Transform creates a pipe
	})

	t.Run("transform flag verification", func(t *testing.T) {
		schema := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return s, nil
		})

		// Transform creates a pipe, so we check the pipe structure
		internals := schema.GetInternals()
		assert.Equal(t, "pipe", internals.Type)
	})
}

// =============================================================================
// 2. String transformations
// =============================================================================

func TestTransformStringOperations(t *testing.T) {
	t.Run("string to string transformation", func(t *testing.T) {
		schema := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return strings.ToUpper(s), nil
		})

		tests := []struct {
			input    string
			expected string
		}{
			{"hello", "HELLO"},
			{"world", "WORLD"},
			{"Hello World", "HELLO WORLD"},
			{"", ""},
			{"test123", "TEST123"},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("input_%s", tt.input), func(t *testing.T) {
				result, err := schema.Parse(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("string to int transformation", func(t *testing.T) {
		schema := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return len(s), nil
		})

		tests := []struct {
			input    string
			expected int
		}{
			{"", 0},
			{"a", 1},
			{"hello", 5},
			{"hello world", 11},
			{"unicode: 你好", 15}, // Unicode characters
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("length_%d", tt.expected), func(t *testing.T) {
				result, err := schema.Parse(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("string to bool transformation", func(t *testing.T) {
		schema := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return s == "true" || s == "1" || s == "yes", nil
		})

		truthyInputs := []string{"true", "1", "yes"}
		for _, input := range truthyInputs {
			t.Run(fmt.Sprintf("truthy_%s", input), func(t *testing.T) {
				result, err := schema.Parse(input)
				require.NoError(t, err)
				assert.Equal(t, true, result)
			})
		}

		falsyInputs := []string{"false", "0", "no", "invalid", ""}
		for _, input := range falsyInputs {
			t.Run(fmt.Sprintf("falsy_%s", input), func(t *testing.T) {
				result, err := schema.Parse(input)
				require.NoError(t, err)
				assert.Equal(t, false, result)
			})
		}
	})

	t.Run("string trim transformation", func(t *testing.T) {
		schema := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return strings.TrimSpace(s), nil
		})

		tests := []struct {
			input    string
			expected string
		}{
			{"  hello  ", "hello"},
			{"\thello\t", "hello"},
			{" \n hello \n ", "hello"},
			{"hello", "hello"},
			{"", ""},
			{"  ", ""},
		}

		for _, tt := range tests {
			t.Run("trim_test", func(t *testing.T) {
				result, err := schema.Parse(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

// =============================================================================
// 3. Type-safe transformations
// =============================================================================

func TestTransformTypeSafety(t *testing.T) {
	t.Run("type-safe string transform", func(t *testing.T) {
		// Transform method provides type safety for string input
		schema := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			// s is guaranteed to be string type
			return fmt.Sprintf("processed: %s", s), nil
		})

		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "processed: test", result)
	})

	t.Run("type-safe with context usage", func(t *testing.T) {
		schema := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			// Use context for error reporting
			if len(s) == 0 {
				// Create a proper ZodIssue using the base structure
				issue := core.ZodIssue{
					ZodIssueBase: core.ZodIssueBase{
						Code:    "custom",
						Message: "Empty string not allowed in transform",
					},
				}
				ctx.AddIssue(issue)
				return nil, ErrEmptyString
			}
			return strings.ToUpper(s), nil
		})

		// Valid case
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "HELLO", result)

		// Invalid case with context error
		_, err = schema.Parse("")
		assert.Error(t, err)
	})

	t.Run("transform with complex output type", func(t *testing.T) {
		schema := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return map[string]any{
				"original": s,
				"length":   len(s),
				"upper":    strings.ToUpper(s),
				"words":    strings.Fields(s),
			}, nil
		})

		result, err := schema.Parse("hello world")
		require.NoError(t, err)

		resultMap, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "hello world", resultMap["original"])
		assert.Equal(t, 11, resultMap["length"])
		assert.Equal(t, "HELLO WORLD", resultMap["upper"])
		assert.Equal(t, []string{"hello", "world"}, resultMap["words"])
	})
}

// =============================================================================
// 4. Transform chaining
// =============================================================================

func TestTransformChaining(t *testing.T) {
	t.Run("simple chain transformation", func(t *testing.T) {
		// First transform: trim whitespace
		step1 := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return strings.TrimSpace(s), nil
		})

		// Second transform: get length - use global Transform function
		step2 := step1.Pipe(Transform(func(input any, ctx *core.RefinementContext) (any, error) {
			if str, ok := input.(string); ok {
				return len(str), nil
			}
			return 0, nil
		}))

		result, err := step2.Parse("  hello  ")
		require.NoError(t, err)
		assert.Equal(t, 5, result) // "hello" has length 5
	})

	t.Run("multi-step string processing", func(t *testing.T) {
		// Step 1: trim and lowercase
		step1 := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return strings.ToLower(strings.TrimSpace(s)), nil
		})

		// Step 2: check if valid - use global Transform function
		step2 := step1.Pipe(Transform(func(input any, ctx *core.RefinementContext) (any, error) {
			if str, ok := input.(string); ok {
				return len(str) > 0 && str != "invalid", nil
			}
			return false, nil
		}))

		tests := []struct {
			name     string
			input    string
			expected bool
		}{
			{"valid_input", "  HELLO  ", true},
			{"empty_after_trim", "   ", false},
			{"invalid_keyword", "INVALID", false},
			{"normal_input", "test", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := step2.Parse(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("complex type chain", func(t *testing.T) {
		// String -> length (int) -> even/odd (bool) -> text (string)
		step1 := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return len(s), nil
		})

		step2 := step1.Pipe(Transform(func(input any, ctx *core.RefinementContext) (any, error) {
			if length, ok := input.(int); ok {
				return length%2 == 0, nil
			}
			return false, nil
		}))

		schema := step2.Pipe(Transform(func(input any, ctx *core.RefinementContext) (any, error) {
			if isEven, ok := input.(bool); ok {
				if isEven {
					return "even length", nil
				}
				return "odd length", nil
			}
			return "unknown", nil
		}))

		tests := []struct {
			input    string
			expected string
		}{
			{"ab", "even length"},   // length 2
			{"abc", "odd length"},   // length 3
			{"abcd", "even length"}, // length 4
			{"", "even length"},     // length 0
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("length_%d", len(tt.input)), func(t *testing.T) {
				result, err := schema.Parse(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

// =============================================================================
// 5. Transform with validation integration
// =============================================================================

func TestTransformValidationIntegration(t *testing.T) {
	t.Run("validation before transform", func(t *testing.T) {
		// Validation happens before transformation
		schema := String().Min(3).Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return strings.ToUpper(s), nil
		})

		// Valid case: passes validation, then transforms
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "HELLO", result)

		// Invalid case: fails validation, transform not executed
		_, err = schema.Parse("hi")
		assert.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Equal(t, issues.TooSmall, zodErr.Issues[0].Code)
	})

	t.Run("transform with pipe validation", func(t *testing.T) {
		// Transform then validate the result
		baseTransform := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return len(s), nil
		})

		// Create a validator for the transformed result
		resultValidator := Transform(func(input any, ctx *core.RefinementContext) (any, error) {
			if length, ok := input.(int); ok && length >= 3 {
				return length, nil
			}
			return nil, fmt.Errorf("%w: length must be >= 3", ErrTransformationFailed)
		})

		schema := baseTransform.Pipe(resultValidator)

		// Valid case: string length >= 3
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, 5, result)

		// Invalid case: string length < 3
		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("transform with refine after", func(t *testing.T) {
		// Transform then refine the result
		schema := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return strings.ToUpper(s), nil
		}).RefineAny(func(v any) bool {
			if str, ok := v.(string); ok {
				return !strings.Contains(str, "TEST")
			}
			return false
		}, core.SchemaParams{Error: "Transformed string cannot contain 'TEST'"})

		// Valid case
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "HELLO", result)

		// Invalid case: after transform contains "TEST"
		_, err = schema.Parse("test")
		assert.Error(t, err)
	})
}

// =============================================================================
// 6. Transform error handling
// =============================================================================

func TestTransformErrorHandling(t *testing.T) {
	t.Run("transform function returning error", func(t *testing.T) {
		schema := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			if s == "error" {
				return nil, ErrTransformationFailed
			}
			return strings.ToUpper(s), nil
		})

		// Valid transformation
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "HELLO", result)

		// Transformation error
		_, err = schema.Parse("error")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "transformation failed")
	})

	t.Run("transform with context error reporting", func(t *testing.T) {
		schema := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			if len(s) == 0 {
				issue := core.ZodIssue{
					ZodIssueBase: core.ZodIssueBase{
						Code:    "custom",
						Message: "Empty string not allowed",
					},
				}
				ctx.AddIssue(issue)
				return nil, ErrEmptyString
			}
			return s, nil
		})

		// Valid case
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Error case with context
		_, err = schema.Parse("")
		assert.Error(t, err)
	})

	t.Run("transform preserves original data on error", func(t *testing.T) {
		schema := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			if s == "fail" {
				return nil, ErrTransformFailed
			}
			return strings.ToUpper(s), nil
		})

		// Transform failure should not modify original data
		_, err := schema.Parse("fail")
		assert.Error(t, err)

		// Error should contain information about the failure
		assert.Contains(t, err.Error(), "transform failed")
	})

	t.Run("MustParse with transform", func(t *testing.T) {
		schema := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return strings.ToUpper(s), nil
		})

		// Valid case
		result := schema.MustParse("hello")
		assert.Equal(t, "HELLO", result)

		// Invalid case should panic
		assert.Panics(t, func() {
			schema.MustParse(123) // Wrong type
		})
	})
}

// =============================================================================
// 7. Transform with different types
// =============================================================================

func TestTransformWithDifferentTypes(t *testing.T) {
	t.Run("global Transform function", func(t *testing.T) {
		schema := Transform(func(input any, ctx *core.RefinementContext) (any, error) {
			switch v := input.(type) {
			case string:
				return strings.ToUpper(v), nil
			case int:
				return v * 2, nil
			case bool:
				if v {
					return "TRUE", nil
				}
				return "FALSE", nil
			default:
				return fmt.Sprintf("%v", v), nil
			}
		})

		tests := []struct {
			name     string
			input    any
			expected any
		}{
			{"string_upper", "hello", "HELLO"},
			{"int_double", 21, 42},
			{"bool_true", true, "TRUE"},
			{"bool_false", false, "FALSE"},
			{"other_type", []int{1, 2, 3}, "[1 2 3]"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := schema.Parse(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("TransformAny method", func(t *testing.T) {
		schema := String().TransformAny(func(input any, ctx *core.RefinementContext) (any, error) {
			// TransformAny allows more flexible input handling
			if str, ok := input.(string); ok {
				return map[string]any{
					"original": str,
					"length":   len(str),
					"upper":    strings.ToUpper(str),
				}, nil
			}
			return nil, ErrExpectedStringInput
		})

		result, err := schema.Parse("hello")
		require.NoError(t, err)

		resultMap, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "hello", resultMap["original"])
		assert.Equal(t, 5, resultMap["length"])
		assert.Equal(t, "HELLO", resultMap["upper"])
	})
}

// =============================================================================
// 8. Transform edge cases and boundary conditions
// =============================================================================

func TestTransformEdgeCases(t *testing.T) {
	t.Run("empty string transformations", func(t *testing.T) {
		schema := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			if s == "" {
				return "empty", nil
			}
			return len(s), nil
		})

		// Empty string
		result, err := schema.Parse("")
		require.NoError(t, err)
		assert.Equal(t, "empty", result)

		// Non-empty string
		result, err = schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, 5, result)
	})

	t.Run("large string transformation", func(t *testing.T) {
		largeString := strings.Repeat("a", 10000)
		schema := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return len(s), nil
		})

		result, err := schema.Parse(largeString)
		require.NoError(t, err)
		assert.Equal(t, 10000, result)
	})

	t.Run("unicode string handling", func(t *testing.T) {
		schema := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return map[string]any{
				"bytes":  len(s),
				"runes":  len([]rune(s)),
				"upper":  strings.ToUpper(s),
				"fields": strings.Fields(s),
			}, nil
		})

		unicodeInput := "Hello 世界! 🌍"
		result, err := schema.Parse(unicodeInput)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Greater(t, resultMap["bytes"], resultMap["runes"]) // Bytes > runes for unicode
		assert.Contains(t, resultMap["upper"], "世界")
	})

	t.Run("nil and zero value handling", func(t *testing.T) {
		schema := Transform(func(input any, ctx *core.RefinementContext) (any, error) {
			if input == nil {
				return "nil", nil
			}
			switch v := input.(type) {
			case string:
				if v == "" {
					return "empty string", nil
				}
				return "string: " + v, nil
			case int:
				if v == 0 {
					return "zero int", nil
				}
				return fmt.Sprintf("int: %d", v), nil
			default:
				return fmt.Sprintf("other: %v", v), nil
			}
		})

		tests := []struct {
			name     string
			input    any
			expected string
		}{
			{"nil_value", nil, "nil"},
			{"empty_string", "", "empty string"},
			{"zero_int", 0, "zero int"},
			{"normal_string", "hello", "string: hello"},
			{"normal_int", 42, "int: 42"},
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
// 9. Integration and workflow tests
// =============================================================================

func TestTransformIntegration(t *testing.T) {
	t.Run("complex transformation pipeline", func(t *testing.T) {
		// Multi-step transformation: trim -> lowercase -> validate -> format
		step1 := String().Min(3).Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			// Step 1: trim and lowercase
			cleaned := strings.ToLower(strings.TrimSpace(s))
			return cleaned, nil
		})

		schema := step1.Pipe(Transform(func(input any, ctx *core.RefinementContext) (any, error) {
			// Step 2: validate and format
			if str, ok := input.(string); ok {
				if len(str) < 3 {
					return nil, ErrCleanedStringTooShort
				}
				return fmt.Sprintf("processed: %s", str), nil
			}
			return nil, ErrExpectedStringType
		}))

		// Valid case
		result, err := schema.Parse("  HELLO  ")
		require.NoError(t, err)
		assert.Equal(t, "processed: hello", result)

		// Invalid case (too short after cleaning)
		_, err = schema.Parse("  HI  ")
		assert.Error(t, err)
	})

	t.Run("transform with pipe and refine", func(t *testing.T) {
		// Transform -> Pipe -> Refine workflow
		baseTransform := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return strings.ToUpper(s), nil
		})

		// Add validation to the transformed result
		validator := Transform(func(input any, ctx *core.RefinementContext) (any, error) {
			if str, ok := input.(string); ok {
				if strings.Contains(str, "INVALID") {
					return nil, ErrContainsInvalidWord
				}
				return str, nil
			}
			return input, nil
		})

		schema := baseTransform.Pipe(validator)

		// Valid case
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "HELLO", result)

		// Invalid case
		_, err = schema.Parse("invalid")
		assert.Error(t, err)
	})

	t.Run("transform immutability", func(t *testing.T) {
		original := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return len(s), nil
		})

		chained := original.Pipe(Transform(func(input any, ctx *core.RefinementContext) (any, error) {
			if length, ok := input.(int); ok {
				return length * 2, nil
			}
			return 0, nil
		}))

		// Should be different instances
		assert.NotSame(t, original, chained)

		// Original should return length
		result1, err := original.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, 4, result1)

		// Chained should return length * 2
		result2, err := chained.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, 8, result2)
	})

	t.Run("transform execution order verification", func(t *testing.T) {
		executionOrder := []string{}

		step1 := String().Min(3).Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			executionOrder = append(executionOrder, "transform1")
			return strings.ToUpper(s), nil
		})

		schema := step1.Pipe(Transform(func(input any, ctx *core.RefinementContext) (any, error) {
			executionOrder = append(executionOrder, "transform2")
			if str, ok := input.(string); ok {
				return str + "_PROCESSED", nil
			}
			return input, nil
		}))

		// Valid case - all steps execute
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "HELLO_PROCESSED", result)
		assert.Equal(t, []string{"transform1", "transform2"}, executionOrder)

		// Invalid case - validation fails, transforms don't execute
		executionOrder = []string{}
		_, err = schema.Parse("hi")
		assert.Error(t, err)
		assert.Empty(t, executionOrder) // No transforms executed
	})

	t.Run("performance smoke test", func(t *testing.T) {
		schema := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return strings.ToUpper(s), nil
		})

		testInputs := []string{
			"hello",
			"world",
			"test",
			"performance",
		}

		// Simple performance check
		for i := 0; i < 100; i++ {
			for _, input := range testInputs {
				_, err := schema.Parse(input)
				require.NoError(t, err)
			}
		}
	})

	t.Run("transform with all modifier methods", func(t *testing.T) {
		// Test that transform works with various modifiers
		baseTransform := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return len(s), nil
		})

		// Test with Nilable
		nilableSchema := baseTransform.Nilable()
		result, err := nilableSchema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, 4, result)

		// Test with Pipe
		pipeSchema := baseTransform.Pipe(Transform(func(input any, ctx *core.RefinementContext) (any, error) {
			if length, ok := input.(int); ok {
				return length > 3, nil
			}
			return false, nil
		}))

		result, err = pipeSchema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})
}
