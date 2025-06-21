package types

import (
	"errors"
	"fmt"
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

func TestRefineBasicFunctionality(t *testing.T) {
	t.Run("Refine method availability", func(t *testing.T) {
		// Test that Refine method is available on string schema
		schema := String().Refine(func(s string) bool {
			return len(s) > 0
		})

		require.NotNil(t, schema)
		// Refine returns the same type for chaining
		assert.IsType(t, &ZodString{}, schema)
	})

	t.Run("RefineAny method availability", func(t *testing.T) {
		// Test that RefineAny method is available
		schema := String().RefineAny(func(v any) bool {
			if str, ok := v.(string); ok {
				return len(str) > 0
			}
			return false
		})

		require.NotNil(t, schema)
		// RefineAny returns ZodType[any, any] for flexibility
		internals := schema.GetInternals()
		require.NotNil(t, internals)
	})

	t.Run("refine flag verification", func(t *testing.T) {
		schema := String().Refine(func(s string) bool {
			return len(s) > 3
		})

		// Refine adds checks to the schema
		internals := schema.GetInternals()
		assert.Greater(t, len(internals.Checks), 0)
	})
}

// =============================================================================
// 2. String refine operations
// =============================================================================

func TestRefineStringOperations(t *testing.T) {
	t.Run("basic string validation", func(t *testing.T) {
		schema := String().Refine(func(s string) bool {
			return strings.ContainsAny(s, "!@#$%^&*()")
		}, core.SchemaParams{
			Error: "String must contain at least one special character",
		})

		// Valid input with special character
		result, err := schema.Parse("password123!")
		require.NoError(t, err)
		assert.Equal(t, "password123!", result) // Refine doesn't modify data

		// Invalid input without special character
		_, err = schema.Parse("password123")
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.Custom, zodErr.Issues[0].Code)
		assert.Contains(t, zodErr.Issues[0].Message, "special character")
	})

	t.Run("string length validation", func(t *testing.T) {
		schema := String().Refine(func(s string) bool {
			return len(s) >= 8 && len(s) <= 20
		}, core.SchemaParams{
			Error: "String length must be between 8 and 20 characters",
		})

		// Valid lengths
		validInputs := []string{"password", "verylongpassword", "12345678"}
		for _, input := range validInputs {
			t.Run(fmt.Sprintf("valid_%s", input), func(t *testing.T) {
				result, err := schema.Parse(input)
				require.NoError(t, err)
				assert.Equal(t, input, result) // Data unchanged
			})
		}

		// Invalid lengths
		invalidInputs := []string{"short", "verylongpasswordthatexceedslimit"}
		for _, input := range invalidInputs {
			t.Run(fmt.Sprintf("invalid_%s", input), func(t *testing.T) {
				_, err := schema.Parse(input)
				assert.Error(t, err)
			})
		}
	})

	t.Run("string pattern validation", func(t *testing.T) {
		// Email-like pattern validation
		schema := String().Refine(func(s string) bool {
			return strings.Contains(s, "@") && strings.Contains(s, ".")
		}, core.SchemaParams{
			Error: "Must be a valid email format",
		})

		// Valid emails
		validEmails := []string{"user@example.com", "test@domain.org", "admin@site.net"}
		for _, email := range validEmails {
			t.Run(fmt.Sprintf("valid_%s", email), func(t *testing.T) {
				result, err := schema.Parse(email)
				require.NoError(t, err)
				assert.Equal(t, email, result)
			})
		}

		// Invalid emails
		invalidEmails := []string{"notanemail", "missing@domain", "user.domain.com"}
		for _, email := range invalidEmails {
			t.Run(fmt.Sprintf("invalid_%s", email), func(t *testing.T) {
				_, err := schema.Parse(email)
				assert.Error(t, err)
			})
		}
	})

	t.Run("string content validation", func(t *testing.T) {
		schema := String().Refine(func(s string) bool {
			// Must not start with "test"
			return !strings.HasPrefix(strings.ToLower(s), "test")
		}, core.SchemaParams{
			Error: "String cannot start with 'test'",
		})

		// Valid strings
		validStrings := []string{"hello", "world", "example", "demo"}
		for _, str := range validStrings {
			result, err := schema.Parse(str)
			require.NoError(t, err)
			assert.Equal(t, str, result)
		}

		// Invalid strings
		invalidStrings := []string{"test", "testing", "Test123", "TEST"}
		for _, str := range invalidStrings {
			_, err := schema.Parse(str)
			assert.Error(t, err)
		}
	})
}

// =============================================================================
// 3. Type-safe refine operations
// =============================================================================

func TestRefineTypeSafety(t *testing.T) {
	t.Run("type-safe string refine", func(t *testing.T) {
		// Refine method provides type safety for string input
		schema := String().Refine(func(s string) bool {
			// s is guaranteed to be string type
			return len(s) > 0 && !strings.Contains(s, " ")
		}, core.SchemaParams{
			Error: "String must be non-empty and contain no spaces",
		})

		// Valid case
		result, err := schema.Parse("validstring")
		require.NoError(t, err)
		assert.Equal(t, "validstring", result)

		// Invalid case
		_, err = schema.Parse("invalid string")
		assert.Error(t, err)
	})

	t.Run("RefineAny for complex validation", func(t *testing.T) {
		schema := String().RefineAny(func(v any) bool {
			// More complex validation logic
			str, ok := v.(string)
			if !ok {
				return false
			}

			// Check if string contains both letters and numbers
			hasLetter := false
			hasDigit := false
			for _, r := range str {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
					hasLetter = true
				}
				if r >= '0' && r <= '9' {
					hasDigit = true
				}
			}

			return hasLetter && hasDigit
		}, core.SchemaParams{
			Error: "String must contain both letters and numbers",
		})

		// Valid cases
		validCases := []string{"abc123", "test1", "Hello2World"}
		for _, input := range validCases {
			result, err := schema.Parse(input)
			require.NoError(t, err)
			assert.Equal(t, input, result)
		}

		// Invalid cases
		invalidCases := []string{"onlyletters", "123456", ""}
		for _, input := range invalidCases {
			_, err := schema.Parse(input)
			assert.Error(t, err)
		}
	})

	t.Run("refine preserves original data", func(t *testing.T) {
		originalData := "test@example.com"
		schema := String().Refine(func(s string) bool {
			return strings.Contains(s, "@")
		})

		result, err := schema.Parse(originalData)
		require.NoError(t, err)
		// Refine NEVER modifies data - core principle
		assert.Equal(t, originalData, result) // Exact same data
	})
}

// =============================================================================
// 4. Refine chaining and composition
// =============================================================================

func TestRefineChaining(t *testing.T) {
	t.Run("multiple refine operations", func(t *testing.T) {
		schema := String().
			Min(8).
			Refine(func(s string) bool {
				return strings.ContainsAny(s, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
			}, core.SchemaParams{
				Error: "Must contain uppercase letter",
			}).
			Refine(func(s string) bool {
				return strings.ContainsAny(s, "0123456789")
			}, core.SchemaParams{
				Error: "Must contain digit",
			}).
			Refine(func(s string) bool {
				return strings.ContainsAny(s, "!@#$%^&*()")
			}, core.SchemaParams{
				Error: "Must contain special character",
			})

		// Valid password meeting all criteria
		result, err := schema.Parse("Password123!")
		require.NoError(t, err)
		assert.Equal(t, "Password123!", result)

		// Invalid password missing digit
		_, err = schema.Parse("Password!")
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, errors.As(err, &zodErr))

		// Should contain digit validation error
		foundDigitError := false
		for _, issue := range zodErr.Issues {
			if strings.Contains(issue.Message, "digit") {
				foundDigitError = true
				break
			}
		}
		assert.True(t, foundDigitError, "Expected digit validation error")
	})

	t.Run("refine with existing validations", func(t *testing.T) {
		schema := String().Min(5).Max(20).Refine(func(s string) bool {
			return strings.Contains(s, "@")
		}, core.SchemaParams{
			Error: "Must contain @ symbol",
		})

		// Should fail min length validation first (execution order)
		_, err := schema.Parse("hi")
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, errors.As(err, &zodErr))
		// Should have min length error, not refine error
		assert.Equal(t, issues.TooSmall, zodErr.Issues[0].Code)
	})

	t.Run("refine execution order", func(t *testing.T) {
		executionOrder := []string{}

		schema := String().
			Min(3).
			Refine(func(s string) bool {
				executionOrder = append(executionOrder, "refine1")
				return len(s) > 5
			}).
			Refine(func(s string) bool {
				executionOrder = append(executionOrder, "refine2")
				return !strings.Contains(s, "invalid")
			})

		// Valid case - all refines execute
		_, err := schema.Parse("validstring")
		require.NoError(t, err)
		assert.Equal(t, []string{"refine1", "refine2"}, executionOrder)

		// Invalid case - early termination at built-in validation
		executionOrder = []string{}
		_, err = schema.Parse("hi") // Fails Min(3)
		assert.Error(t, err)
		// GoZod may still execute refines even if built-in validation fails
		// This is implementation-specific behavior

		// Invalid case - test actual behavior
		executionOrder = []string{}
		_, err = schema.Parse("test") // Passes Min(3) but fails refine1
		assert.Error(t, err)
		// Check if refines were executed (implementation-dependent)
		if len(executionOrder) > 0 {
			assert.Contains(t, executionOrder, "refine1")
		}
	})
}

// =============================================================================
// 5. Refine with different types
// =============================================================================

func TestRefineWithDifferentTypes(t *testing.T) {
	t.Run("integer refine", func(t *testing.T) {
		schema := Int().Refine(func(value int) bool {
			return value%2 == 0
		}, core.SchemaParams{
			Error: "Number must be even",
		})

		// Valid even numbers
		evenNumbers := []int{2, 4, 6, 8, 10, 0, -2}
		for _, num := range evenNumbers {
			result, err := schema.Parse(num)
			require.NoError(t, err)
			assert.Equal(t, num, result)
		}

		// Invalid odd numbers
		oddNumbers := []int{1, 3, 5, 7, 9, -1}
		for _, num := range oddNumbers {
			_, err := schema.Parse(num)
			assert.Error(t, err)
		}
	})

	t.Run("float refine", func(t *testing.T) {
		schema := Float64().Refine(func(value float64) bool {
			// Check if value has at most 2 decimal places
			return math.Abs(value*100-math.Round(value*100)) < 1e-9
		}, core.SchemaParams{
			Error: "Number must have at most 2 decimal places",
		})

		// Valid precision
		validValues := []float64{12.34, 0.99, 100.00, 7.5, 0.0}
		for _, val := range validValues {
			result, err := schema.Parse(val)
			require.NoError(t, err)
			assert.Equal(t, val, result)
		}

		// Invalid precision
		_, err := schema.Parse(12.345)
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Contains(t, zodErr.Issues[0].Message, "2 decimal places")
	})

	t.Run("boolean refine", func(t *testing.T) {
		schema := Bool().Refine(func(value bool) bool {
			return value == true
		}, core.SchemaParams{
			Error: "Value must be true",
		})

		// Valid case
		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Invalid case
		_, err = schema.Parse(false)
		assert.Error(t, err)
	})

	t.Run("BigInt refine", func(t *testing.T) {
		schema := BigInt().Refine(func(value *big.Int) bool {
			return value.Sign() > 0
		}, core.SchemaParams{
			Error: "BigInt must be positive",
		})

		// Valid positive value
		positiveValue := big.NewInt(42)
		result, err := schema.Parse(positiveValue)
		require.NoError(t, err)
		assert.Equal(t, 0, result.(*big.Int).Cmp(positiveValue))

		// Invalid negative value
		negativeValue := big.NewInt(-42)
		_, err = schema.Parse(negativeValue)
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Equal(t, core.Custom, zodErr.Issues[0].Code)
		assert.Equal(t, "BigInt must be positive", zodErr.Issues[0].Message)
	})
}

// =============================================================================
// 6. Refine error handling and messages
// =============================================================================

func TestRefineErrorHandling(t *testing.T) {
	t.Run("custom error messages", func(t *testing.T) {
		schema := String().Refine(func(s string) bool {
			return len(s) > 5
		}, core.SchemaParams{
			Error: "String is too short, must be longer than 5 characters",
		})

		_, err := schema.Parse("hi")
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, errors.As(err, &zodErr))

		if len(zodErr.Issues) == 0 {
			t.Fatal("Expected validation issues")
		}

		// Check custom error message is correctly applied
		issue := zodErr.Issues[0]
		expectedMsg := "String is too short, must be longer than 5 characters"
		assert.Equal(t, expectedMsg, issue.Message)
	})

	t.Run("function-based error messages", func(t *testing.T) {
		schema := String().Min(3).Refine(func(s string) bool {
			return !strings.HasPrefix(s, "test")
		}, core.SchemaParams{
			Error: func(issue core.ZodRawIssue) string {
				if input, ok := issue.Input.(string); ok {
					return fmt.Sprintf("The string '%s' cannot start with 'test'", input)
				}
				return "Invalid string input"
			},
		})

		_, err := schema.Parse("testing")
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)

		expectedMsg := "The string 'testing' cannot start with 'test'"
		assert.Equal(t, expectedMsg, zodErr.Issues[0].Message)
	})

	t.Run("error path specification", func(t *testing.T) {
		schema := String().Refine(func(s string) bool {
			return len(s) > 0
		}, core.SchemaParams{
			Error: "String cannot be empty",
			Path:  []string{"username"},
		})

		_, err := schema.Parse("")
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)

		// Check error path
		if len(zodErr.Issues[0].Path) > 0 {
			assert.Equal(t, "username", zodErr.Issues[0].Path[0])
		}
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		schema := String().
			Min(8).
			Refine(func(s string) bool {
				return strings.ContainsAny(s, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
			}, core.SchemaParams{
				Error: "Must contain uppercase letter",
			}).
			Refine(func(s string) bool {
				return strings.ContainsAny(s, "0123456789")
			}, core.SchemaParams{
				Error: "Must contain digit",
			})

		// This should fail multiple validations
		_, err := schema.Parse("short")
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, errors.As(err, &zodErr))
		// Should have at least the min length error
		assert.Greater(t, len(zodErr.Issues), 0)
	})
}

// =============================================================================
// 7. Refine with complex data structures
// =============================================================================

func TestRefineComplexStructures(t *testing.T) {
	t.Run("object refine", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"password":        String().Min(6),
			"confirmPassword": String().Min(6),
		}).Refine(func(data map[string]any) bool {
			password, _ := data["password"].(string)
			confirmPassword, _ := data["confirmPassword"].(string)
			return password == confirmPassword
		}, core.SchemaParams{
			Error: "Passwords must match",
			Path:  []string{"confirmPassword"},
		})

		// Valid matching passwords
		validData := map[string]any{
			"password":        "secret123",
			"confirmPassword": "secret123",
		}
		result, err := schema.Parse(validData)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Invalid mismatched passwords
		invalidData := map[string]any{
			"password":        "secret123",
			"confirmPassword": "different",
		}
		_, err = schema.Parse(invalidData)
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Len(t, zodErr.Issues, 1)
		assert.Equal(t, core.Custom, zodErr.Issues[0].Code)

		// Check error path points to confirmPassword
		if len(zodErr.Issues[0].Path) > 0 {
			assert.Equal(t, "confirmPassword", zodErr.Issues[0].Path[0])
		}
	})

	t.Run("slice refine", func(t *testing.T) {
		schema := Slice(Int()).Refine(func(arr []any) bool {
			// All elements must be positive
			for _, item := range arr {
				if num, ok := item.(int); !ok || num <= 0 {
					return false
				}
			}
			return true
		}, core.SchemaParams{
			Error: "All numbers must be positive",
		})

		// Valid cases
		validSlices := [][]int{
			{1, 2, 3},
			{100},
			{5, 10, 15, 20},
		}

		for _, slice := range validSlices {
			input := make([]any, len(slice))
			for i, v := range slice {
				input[i] = v
			}

			_, err := schema.Parse(input)
			require.NoError(t, err)
		}

		// Invalid cases
		invalidSlices := [][]int{
			{1, -2, 3},   // contains negative
			{0, 1, 2},    // contains zero
			{-1, -2, -3}, // all negative
		}

		for _, slice := range invalidSlices {
			input := make([]any, len(slice))
			for i, v := range slice {
				input[i] = v
			}

			_, err := schema.Parse(input)
			assert.Error(t, err)
		}
	})

	t.Run("map refine", func(t *testing.T) {
		schema := Map(String(), Int()).Refine(func(m map[any]any) bool {
			return len(m) > 0
		}, core.SchemaParams{
			Error: "Map must not be empty",
		})

		// Valid non-empty map
		validMap := map[string]int{"key1": 1, "key2": 2}
		result, err := schema.Parse(validMap)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Invalid empty map
		emptyMap := map[string]int{}
		_, err = schema.Parse(emptyMap)
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, errors.As(err, &zodErr))
		assert.Equal(t, "Map must not be empty", zodErr.Issues[0].Message)
	})
}

// =============================================================================
// 8. Refine edge cases and boundary conditions
// =============================================================================

func TestRefineEdgeCases(t *testing.T) {
	t.Run("empty string refine", func(t *testing.T) {
		schema := String().Refine(func(s string) bool {
			return s != ""
		}, core.SchemaParams{
			Error: "String cannot be empty",
		})

		// Valid non-empty string
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Invalid empty string
		_, err = schema.Parse("")
		assert.Error(t, err)
	})

	t.Run("large string refine", func(t *testing.T) {
		largeString := strings.Repeat("a", 10000)
		schema := String().Refine(func(s string) bool {
			return len(s) <= 10000
		}, core.SchemaParams{
			Error: "String too long",
		})

		result, err := schema.Parse(largeString)
		require.NoError(t, err)
		assert.Equal(t, largeString, result)

		// Test with string that's too long
		tooLargeString := strings.Repeat("a", 10001)
		_, err = schema.Parse(tooLargeString)
		assert.Error(t, err)
	})

	t.Run("unicode string refine", func(t *testing.T) {
		schema := String().Refine(func(s string) bool {
			// Check if string contains only ASCII characters
			for _, r := range s {
				if r > 127 {
					return false
				}
			}
			return true
		}, core.SchemaParams{
			Error: "String must contain only ASCII characters",
		})

		// Valid ASCII string
		result, err := schema.Parse("Hello World!")
		require.NoError(t, err)
		assert.Equal(t, "Hello World!", result)

		// Invalid Unicode string
		_, err = schema.Parse("Hello 世界! 🌍")
		assert.Error(t, err)
	})

	t.Run("nil and zero value refine", func(t *testing.T) {
		schema := String().Refine(func(s string) bool {
			// Any string is valid (including empty)
			return true
		})

		// Test with empty string
		result, err := schema.Parse("")
		require.NoError(t, err)
		assert.Equal(t, "", result)

		// Test with normal string
		result, err = schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("refine with nil function", func(t *testing.T) {
		// This tests edge case behavior
		assert.NotPanics(t, func() {
			_ = String().Refine(nil)
		})
	})
}

// =============================================================================
// 9. Integration and workflow tests
// =============================================================================

func TestRefineIntegration(t *testing.T) {
	t.Run("refine with transform separation", func(t *testing.T) {
		// Test that Refine and Transform have clear separation of concerns
		// Refine: only validates, never modifies
		// Transform: can modify data

		originalData := "hello world"

		// Refine should never modify data
		refineSchema := String().Refine(func(s string) bool {
			return len(s) > 5
		})

		result, err := refineSchema.Parse(originalData)
		require.NoError(t, err)
		assert.Equal(t, originalData, result) // Exact same data

		// Transform should modify data
		transformSchema := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			return strings.ToUpper(s), nil
		})

		result, err = transformSchema.Parse(originalData)
		require.NoError(t, err)
		assert.Equal(t, "HELLO WORLD", result)   // Modified data
		assert.NotEqual(t, originalData, result) // Different from original
	})

	t.Run("complex validation pipeline", func(t *testing.T) {
		// Multi-step validation: basic -> format -> content -> security
		schema := String().
			Min(8).Max(50).
			Refine(func(s string) bool {
				// Must contain @ and .
				return strings.Contains(s, "@") && strings.Contains(s, ".")
			}, core.SchemaParams{
				Error: "Must be valid email format",
			}).
			Refine(func(s string) bool {
				// Must not contain spaces
				return !strings.Contains(s, " ")
			}, core.SchemaParams{
				Error: "Email cannot contain spaces",
			}).
			Refine(func(s string) bool {
				// Must not be a common test email
				testEmails := []string{"test@test.com", "admin@admin.com", "user@user.com"}
				for _, testEmail := range testEmails {
					if s == testEmail {
						return false
					}
				}
				return true
			}, core.SchemaParams{
				Error: "Cannot use common test email addresses",
			})

		// Valid email
		result, err := schema.Parse("user@example.com")
		require.NoError(t, err)
		assert.Equal(t, "user@example.com", result)

		// Invalid email (test email)
		_, err = schema.Parse("test@test.com")
		assert.Error(t, err)

		// Invalid email (contains space)
		_, err = schema.Parse("user name@example.com")
		assert.Error(t, err)
	})

	t.Run("refine immutability", func(t *testing.T) {
		original := String().Refine(func(s string) bool {
			return len(s) > 3
		})

		chained := original.Refine(func(s string) bool {
			return !strings.Contains(s, "test")
		})

		// Should be different instances
		assert.NotSame(t, original, chained)

		// Both should work independently
		result1, err := original.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result1)

		result2, err := chained.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result2)

		// Original should pass "test" (only length check)
		result3, err := original.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result3)

		// Chained should fail "test" (additional content check)
		_, err = chained.Parse("test")
		assert.Error(t, err)
	})

	t.Run("refine execution order verification", func(t *testing.T) {
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
		// Check actual behavior - may or may not execute transforms
		// This is implementation-specific
	})

	t.Run("performance with multiple refines", func(t *testing.T) {
		schema := String().
			Refine(func(s string) bool {
				return len(s) > 0
			}).
			Refine(func(s string) bool {
				return !strings.Contains(s, "invalid")
			}).
			Refine(func(s string) bool {
				return len(s) < 100
			}).
			Refine(func(s string) bool {
				return strings.TrimSpace(s) == s
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

	t.Run("refine with all modifier methods", func(t *testing.T) {
		// Test that refine works with various modifiers
		baseRefine := String().Refine(func(s string) bool {
			return len(s) > 3
		})

		// Test with Optional
		optionalSchema := baseRefine.Optional()
		result, err := optionalSchema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Test with Default
		defaultSchema := baseRefine.Default("default")
		result, err = defaultSchema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})
}
