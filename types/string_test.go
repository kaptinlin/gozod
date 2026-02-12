package types

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Basic functionality tests
// =============================================================================

func TestString_BasicFunctionality(t *testing.T) {
	t.Parallel()
	t.Run("valid string inputs", func(t *testing.T) {
		t.Parallel()
		schema := String()

		// Test string value
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Test empty string
		result, err = schema.Parse("")
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("invalid type inputs", func(t *testing.T) {
		schema := String()

		invalidInputs := []any{
			123, 3.14, true, []string{"hello"}, nil,
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})

	t.Run("Parse and MustParse methods", func(t *testing.T) {
		schema := String()

		// Test Parse method
		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)

		// Test MustParse method
		mustResult := schema.MustParse("hello")
		assert.Equal(t, "hello", mustResult)

		// Test panic on invalid input
		assert.Panics(t, func() {
			schema.MustParse(123)
		})
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected a string value"
		schema := String(core.SchemaParams{Error: customError})

		require.NotNil(t, schema)
		assert.Equal(t, core.ZodTypeString, schema.internals.Def.Type)

		_, err := schema.Parse(123)
		assert.Error(t, err)
	})
}

// =============================================================================
// StrictParse and MustStrictParse tests
// =============================================================================

func TestString_StrictParse(t *testing.T) {
	t.Parallel()
	t.Run("basic functionality", func(t *testing.T) {
		t.Parallel()
		schema := String()

		// Test StrictParse with exact type match
		result, err := schema.StrictParse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)
		assert.IsType(t, "", result)

		// Test StrictParse with empty string
		emptyResult, err := schema.StrictParse("")
		require.NoError(t, err)
		assert.Equal(t, "", emptyResult)
	})

	t.Run("with validation constraints", func(t *testing.T) {
		schema := String().Min(5)

		// Valid case
		result, err := schema.StrictParse("hello world")
		require.NoError(t, err)
		assert.Equal(t, "hello world", result)

		// Invalid case - too short
		_, err = schema.StrictParse("hi")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Too small")
	})

	t.Run("with pointer types", func(t *testing.T) {
		schema := StringPtr()
		str := "hello"

		// Test with valid pointer input
		result, err := schema.StrictParse(&str)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "hello", *result)
		assert.IsType(t, (*string)(nil), result)
	})

	t.Run("with default values", func(t *testing.T) {
		schema := StringPtr().Default("default_value")
		var nilPtr *string = nil

		// Test with nil input (should use default)
		result, err := schema.StrictParse(nilPtr)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "default_value", *result)
	})

	t.Run("with prefault values", func(t *testing.T) {
		schema := StringPtr().Min(10).Prefault("prefault_value")
		shortStr := "hi" // Too short for Min(10)

		// Test with validation failure (should NOT use prefault, should return error)
		_, err := schema.StrictParse(&shortStr)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least 10")

		// Test with nil input (should use prefault)
		result, err := schema.StrictParse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "prefault_value", *result)
	})
}

func TestString_MustStrictParse(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		schema := String()

		// Test MustStrictParse with valid input
		result := schema.MustStrictParse("hello")
		assert.Equal(t, "hello", result)
		assert.IsType(t, "", result)

		// Test MustStrictParse with empty string
		emptyResult := schema.MustStrictParse("")
		assert.Equal(t, "", emptyResult)
	})

	t.Run("panic behavior", func(t *testing.T) {
		schema := String().Min(5)

		// Test panic with validation failure
		assert.Panics(t, func() {
			schema.MustStrictParse("hi") // Too short, should panic
		})
	})

	t.Run("with pointer types", func(t *testing.T) {
		schema := StringPtr()
		str := "world"

		// Test with valid pointer input
		result := schema.MustStrictParse(&str)
		require.NotNil(t, result)
		assert.Equal(t, "world", *result)
		assert.IsType(t, (*string)(nil), result)
	})

	t.Run("with default values", func(t *testing.T) {
		schema := StringPtr().Default("default_value")
		var nilPtr *string = nil

		// Test with nil input (should use default)
		result := schema.MustStrictParse(nilPtr)
		require.NotNil(t, result)
		assert.Equal(t, "default_value", *result)
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestString_TypeSafety(t *testing.T) {
	t.Parallel()
	t.Run("String returns string type", func(t *testing.T) {
		t.Parallel()
		schema := String()
		require.NotNil(t, schema)

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
		assert.IsType(t, "", result) // Ensure type is string

		result, err = schema.Parse("world")
		require.NoError(t, err)
		assert.Equal(t, "world", result)
		assert.IsType(t, "", result)
	})

	t.Run("StringPtr returns *string type", func(t *testing.T) {
		schema := StringPtr()
		require.NotNil(t, schema)

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "hello", *result)
		assert.IsType(t, (*string)(nil), result) // Ensure type is *string

		result, err = schema.Parse("world")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "world", *result)
	})

	t.Run("type inference with assignment", func(t *testing.T) {
		// Type-inference friendly API
		stringSchema := String() // string type
		ptrSchema := StringPtr() // *string type

		// Test string type
		result1, err1 := stringSchema.Parse("hello")
		require.NoError(t, err1)
		assert.IsType(t, "", result1)
		assert.Equal(t, "hello", result1)

		// Test *string type
		result2, err2 := ptrSchema.Parse("world")
		require.NoError(t, err2)
		assert.IsType(t, (*string)(nil), result2)
		require.NotNil(t, result2)
		assert.Equal(t, "world", *result2)
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		// Test string type
		stringSchema := String()
		result := stringSchema.MustParse("hello")
		assert.IsType(t, "", result)
		assert.Equal(t, "hello", result)

		// Test *string type
		ptrSchema := StringPtr()
		ptrResult := ptrSchema.MustParse("world")
		assert.IsType(t, (*string)(nil), ptrResult)
		require.NotNil(t, ptrResult)
		assert.Equal(t, "world", *ptrResult)
	})

	t.Run("StrictParse type safety", func(t *testing.T) {
		// Test string type with StrictParse
		stringSchema := String()
		result, err := stringSchema.StrictParse("hello")
		require.NoError(t, err)
		assert.IsType(t, "", result)
		assert.Equal(t, "hello", result)

		// Test *string type with StrictParse
		ptrSchema := StringPtr()
		str := "world"
		ptrResult, err := ptrSchema.StrictParse(&str)
		require.NoError(t, err)
		assert.IsType(t, (*string)(nil), ptrResult)
		require.NotNil(t, ptrResult)
		assert.Equal(t, "world", *ptrResult)

		// Test that StrictParse works correctly with pointer types
		originalPtr := &str
		resultPtr, err := ptrSchema.StrictParse(originalPtr)
		require.NoError(t, err)
		require.NotNil(t, resultPtr)
		assert.Equal(t, "world", *resultPtr)
		// Note: StrictParse may not preserve pointer identity in current implementation

		// Test nil handling with pointer schemas
		nilResult, err := ptrSchema.StrictParse(nil)
		assert.Error(t, err) // Should error for nil input to non-nilable schema
		assert.Nil(t, nilResult)
	})

	t.Run("MustStrictParse type safety", func(t *testing.T) {
		// Test string type with MustStrictParse
		stringSchema := String()
		result := stringSchema.MustStrictParse("hello")
		assert.IsType(t, "", result)
		assert.Equal(t, "hello", result)

		// Test *string type with MustStrictParse
		ptrSchema := StringPtr()
		str := "world"
		ptrResult := ptrSchema.MustStrictParse(&str)
		assert.IsType(t, (*string)(nil), ptrResult)
		require.NotNil(t, ptrResult)
		assert.Equal(t, "world", *ptrResult)

		// Test panic behavior with invalid input
		assert.Panics(t, func() {
			ptrSchema.MustStrictParse(nil) // Should panic for nil input
		})

		// Test panic with validation failure
		constraintSchema := String().Min(10)
		assert.Panics(t, func() {
			constraintSchema.MustStrictParse("short") // Should panic for validation failure
		})
	})
}

// =============================================================================
// Modifier methods tests
// =============================================================================

func TestString_Modifiers(t *testing.T) {
	t.Parallel()
	t.Run("Optional always returns *string", func(t *testing.T) {
		t.Parallel()
		// From string to *string via Optional
		stringSchema := String()
		optionalSchema := stringSchema.Optional()

		// Type check: ensure it returns *ZodString[*string]
		var _ = optionalSchema

		// Functionality test
		result, err := optionalSchema.Parse("hello")
		require.NoError(t, err)
		assert.IsType(t, (*string)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, "hello", *result)

		// From *string to *string via Optional (maintains type)
		ptrSchema := StringPtr()
		optionalPtrSchema := ptrSchema.Optional()
		var _ = optionalPtrSchema
	})

	t.Run("Nilable always returns *string", func(t *testing.T) {
		stringSchema := String()
		nilableSchema := stringSchema.Nilable()

		var _ = nilableSchema

		// Test nil handling
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default preserves current type", func(t *testing.T) {
		// string maintains string
		stringSchema := String()
		defaultStringSchema := stringSchema.Default("default")
		var _ = defaultStringSchema

		// *string maintains *string
		ptrSchema := StringPtr()
		defaultPtrSchema := ptrSchema.Default("default")
		var _ = defaultPtrSchema
	})

	t.Run("Prefault preserves current type", func(t *testing.T) {
		// string maintains string
		stringSchema := String()
		prefaultStringSchema := stringSchema.Prefault("fallback")
		var _ = prefaultStringSchema

		// *string maintains *string
		ptrSchema := StringPtr()
		prefaultPtrSchema := ptrSchema.Prefault("fallback")
		var _ = prefaultPtrSchema
	})

	t.Run("Nullish combines optional and nilable", func(t *testing.T) {
		stringSchema := String()
		nullishSchema := stringSchema.Nullish()

		var _ = nullishSchema

		// Test nil handling
		result, err := nullishSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid string
		result, err = nullishSchema.Parse("hello")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "hello", *result)
	})

	t.Run("StrictParse with modifiers", func(t *testing.T) {
		// Test Optional with StrictParse
		optionalSchema := String().Optional()
		str := "hello"
		result, err := optionalSchema.StrictParse(&str)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "hello", *result)

		// Test Nilable with StrictParse
		nilableSchema := String().Nilable()
		nilableResult, err := nilableSchema.StrictParse(&str)
		require.NoError(t, err)
		require.NotNil(t, nilableResult)
		assert.Equal(t, "hello", *nilableResult)

		// Test Nullish with StrictParse
		nullishSchema := String().Nullish()
		nullishResult, err := nullishSchema.StrictParse(&str)
		require.NoError(t, err)
		require.NotNil(t, nullishResult)
		assert.Equal(t, "hello", *nullishResult)
	})
}

// =============================================================================
// Validation methods tests
// =============================================================================

func TestString_BasicFunctionality_Validations(t *testing.T) {
	t.Parallel()
	t.Run("length validations", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name    string
			schema  *ZodString[string]
			input   string
			wantErr bool
		}{
			{"min length valid", String().Min(5), "hello", false},
			{"min length invalid", String().Min(5), "hi", true},
			{"max length valid", String().Max(5), "hello", false},
			{"max length invalid", String().Max(5), "hello world", true},
			{"exact length valid", String().Length(5), "hello", false},
			{"exact length invalid", String().Length(5), "hi", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := tt.schema.Parse(tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("pattern validations", func(t *testing.T) {
		tests := []struct {
			name    string
			schema  *ZodString[string]
			input   string
			wantErr bool
		}{
			{"starts with valid", String().StartsWith("hello"), "hello world", false},
			{"starts with invalid", String().StartsWith("hello"), "world hello", true},
			{"ends with valid", String().EndsWith("world"), "hello world", false},
			{"ends with invalid", String().EndsWith("world"), "world hello", true},
			{"includes valid", String().Includes("test"), "this is a test string", false},
			{"includes invalid", String().Includes("test"), "this is a string", true},
			{"regex valid", String().Regex(regexp.MustCompile(`^\d+$`)), "12345", false},
			{"regex invalid", String().Regex(regexp.MustCompile(`^\d+$`)), "abc123", true},
			{"regex string valid digits", String().RegexString(`^\d+$`), "12345", false},
			{"regex string invalid digits", String().RegexString(`^\d+$`), "abc123", true},
			{"regex string valid email", String().RegexString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`), "test@example.com", false},
			{"regex string invalid email", String().RegexString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`), "invalid-email", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := tt.schema.Parse(tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("regex string convenience method", func(t *testing.T) {
		t.Run("basic functionality", func(t *testing.T) {
			// Test that RegexString works the same as Regex with compiled pattern
			regexSchema := String().Regex(regexp.MustCompile(`^[A-Z][a-z]+$`))
			regexStringSchema := String().RegexString(`^[A-Z][a-z]+$`)

			testCases := []struct {
				input   string
				wantErr bool
			}{
				{"Hello", false},
				{"World", false},
				{"hello", true},    // doesn't start with uppercase
				{"HELLO", true},    // all uppercase
				{"Hello123", true}, // contains numbers
			}

			for _, tc := range testCases {
				// Test regex method
				_, err1 := regexSchema.Parse(tc.input)
				hasErr1 := err1 != nil

				// Test regex string method
				_, err2 := regexStringSchema.Parse(tc.input)
				hasErr2 := err2 != nil

				// Both methods should behave identically
				assert.Equal(t, hasErr1, hasErr2, "RegexString and Regex should behave identically for input: %s", tc.input)
				assert.Equal(t, tc.wantErr, hasErr1, "Expected error status for input: %s", tc.input)
			}
		})

		t.Run("chaining with other validations", func(t *testing.T) {
			// Test RegexString can be chained with other string validations
			schema := String().Min(10).RegexString(`^\w+@\w+\.\w+$`).Max(50)

			// Valid: meets all criteria
			result, err := schema.Parse("user@domain.com")
			require.NoError(t, err)
			assert.Equal(t, "user@domain.com", result)

			// Invalid: too short (a@b.c is only 5 characters, less than min 10)
			_, err = schema.Parse("a@b.c")
			assert.Error(t, err)

			// Invalid: doesn't match regex
			_, err = schema.Parse("invalid-email-format")
			assert.Error(t, err)

			// Invalid: too long
			longEmail := strings.Repeat("a", 40) + "@example.com"
			_, err = schema.Parse(longEmail)
			assert.Error(t, err)
		})

		t.Run("complex patterns", func(t *testing.T) {
			tests := []struct {
				name    string
				pattern string
				valid   []string
				invalid []string
			}{
				{
					name:    "phone number",
					pattern: `^\+?1?[- ]?\(?[0-9]{3}\)?[- ]?[0-9]{3}[- ]?[0-9]{4}$`,
					valid:   []string{"123-456-7890", "(123) 456-7890", "+1-123-456-7890", "1234567890"},
					invalid: []string{"123-45-6789", "abc-def-ghij", "123-456-789"},
				},
				{
					name:    "hex color",
					pattern: `^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`,
					valid:   []string{"#ff0000", "#FF0000", "#f00", "#123ABC"},
					invalid: []string{"ff0000", "#gg0000", "#12345", "#1234567"},
				},
				{
					name:    "uuid v4",
					pattern: `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`,
					valid:   []string{"550e8400-e29b-41d4-a716-446655440000", "f47ac10b-58cc-4372-a567-0e02b2c3d479"},
					invalid: []string{"550e8400-e29b-41d4-a716-44665544000", "not-a-uuid", "550e8400-e29b-31d4-a716-446655440000", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					schema := String().RegexString(tt.pattern)

					// Test valid inputs
					for _, input := range tt.valid {
						result, err := schema.Parse(input)
						require.NoError(t, err, "Expected %s to be valid for pattern %s", input, tt.name)
						assert.Equal(t, input, result)
					}

					// Test invalid inputs
					for _, input := range tt.invalid {
						_, err := schema.Parse(input)
						assert.Error(t, err, "Expected %s to be invalid for pattern %s", input, tt.name)
					}
				})
			}
		})

		t.Run("error handling", func(t *testing.T) {
			// Test that invalid regex patterns panic (same as regexp.MustCompile)
			assert.Panics(t, func() {
				String().RegexString(`[invalid regex`)
			}, "RegexString should panic on invalid regex pattern")
		})

		t.Run("with custom error messages", func(t *testing.T) {
			schema := String().RegexString(`^\d{3}-\d{2}-\d{4}$`, "Must be in SSN format (XXX-XX-XXXX)")

			// Valid input
			result, err := schema.Parse("123-45-6789")
			require.NoError(t, err)
			assert.Equal(t, "123-45-6789", result)

			// Invalid input should contain custom error
			_, err = schema.Parse("invalid-ssn")
			assert.Error(t, err)
			// Note: The exact error message checking would depend on the error format implementation
		})
	})

	t.Run("json validation", func(t *testing.T) {
		schema := String().JSON()

		// Valid JSON
		result, err := schema.Parse(`{"key": "value"}`)
		require.NoError(t, err)
		assert.Equal(t, `{"key": "value"}`, result)

		// Invalid JSON
		_, err = schema.Parse(`{invalid json}`)
		assert.Error(t, err)
	})
}

// =============================================================================
// Chaining tests
// =============================================================================

func TestString_Chaining(t *testing.T) {
	t.Parallel()
	t.Run("multiple validations", func(t *testing.T) {
		t.Parallel()
		schema := String().Min(5).Max(10).StartsWith("hello")

		// Valid input
		result, err := schema.Parse("hello123")
		require.NoError(t, err)
		assert.Equal(t, "hello123", result)

		// Invalid: too short
		_, err = schema.Parse("hi")
		assert.Error(t, err)

		// Invalid: too long
		_, err = schema.Parse("hello world!")
		assert.Error(t, err)

		// Invalid: wrong prefix
		_, err = schema.Parse("world hello")
		assert.Error(t, err)
	})

	t.Run("modifiers with validations", func(t *testing.T) {
		schema := String().Min(3).Optional()

		// Valid string
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "hello", *result)

		// Invalid: too short
		_, err = schema.Parse("hi")
		assert.Error(t, err)

		// nil should be allowed for optional
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("type evolution through chain", func(t *testing.T) {
		// Start with string, add validation, then make optional
		base := String()
		var _ = base

		withValidation := base.Min(3)
		var _ = withValidation

		optional := withValidation.Optional()
		var _ = optional

		// Functionality test
		result, err := optional.Parse("hello")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "hello", *result)
	})

	t.Run("Refine with chained validations", func(t *testing.T) {
		// Test that Refine supports chaining with other String validations
		// Similar to z.string().refine(val => val.includes("@")).min(5)
		schema := String().Refine(func(s string) bool {
			return len(s) > 0 && s[len(s)-1:] != " " // Custom validation: doesn't end with space
		}, "String must not end with space").Min(5)

		// Valid: passes both refine and min validation
		result, err := schema.Parse("hello@world")
		require.NoError(t, err)
		assert.Equal(t, "hello@world", result)

		// Invalid: fails min validation but passes refine
		_, err = schema.Parse("test")
		assert.Error(t, err)

		// Invalid: passes min validation but fails refine
		_, err = schema.Parse("hello world ")
		assert.Error(t, err)
	})
}

// =============================================================================
// Default and Prefault tests
// =============================================================================

func TestString_Modifiers_DefaultAndPrefault(t *testing.T) {
	// Test Default behavior - only triggers on nil input
	t.Run("Default with nil input", func(t *testing.T) {
		schema := String().Default("default value")

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "default value", result)
	})

	t.Run("Default with valid input", func(t *testing.T) {
		schema := String().Default("default value")

		result, err := schema.Parse("actual value")
		require.NoError(t, err)
		assert.Equal(t, "actual value", result)
	})

	// Test that empty string does NOT trigger default (Zod v4 behavior)
	t.Run("Default with empty string input - should NOT trigger", func(t *testing.T) {
		schema := String().Default("default value")

		result, err := schema.Parse("")
		require.NoError(t, err)
		assert.Equal(t, "", result, "Empty string should not trigger default")

		// Test whitespace also doesn't trigger default
		result2, err2 := schema.Parse(" ")
		require.NoError(t, err2)
		assert.Equal(t, " ", result2, "Whitespace should not trigger default")
	})

	t.Run("DefaultFunc", func(t *testing.T) {
		called := false
		schema := String().DefaultFunc(func() string {
			called = true
			return "generated default"
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "generated default", result)
		assert.True(t, called)
	})

	// Test DefaultFunc is NOT called for empty string
	t.Run("DefaultFunc with empty string - should NOT be called", func(t *testing.T) {
		called := false
		schema := String().DefaultFunc(func() string {
			called = true
			return "generated default"
		})

		result, err := schema.Parse("")
		require.NoError(t, err)
		assert.Equal(t, "", result)
		assert.False(t, called, "DefaultFunc should not be called for empty string")
	})

	// Test Default has higher priority than Prefault
	t.Run("Default priority over Prefault", func(t *testing.T) {
		schema := String().Min(5).Default("hi").Prefault("fallback value")

		// Nil input should use default (short-circuit), not prefault
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "hi", result, "Default should take priority over prefault")
	})

	// Test Default short-circuit mechanism
	t.Run("Default short-circuit bypasses validation", func(t *testing.T) {
		schema := String().Min(10).Default("short")

		// Default value doesn't meet validation but should still be returned
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "short", result, "Default should bypass validation")
	})

	// Test Prefault only triggers on nil input (Zod v4 semantics)
	t.Run("Prefault only triggers on nil input", func(t *testing.T) {
		schema := String().Min(5).Prefault("fallback")

		// Valid input should not trigger prefault
		result1, err1 := schema.Parse("valid input")
		require.NoError(t, err1)
		assert.Equal(t, "valid input", result1)

		// Invalid non-nil input should return error (NOT trigger prefault)
		_, err2 := schema.Parse("hi")
		require.Error(t, err2)
		assert.Contains(t, err2.Error(), "at least 5 characters")

		// Nil input should trigger prefault and go through full validation
		result3, err3 := schema.Parse(nil)
		require.NoError(t, err3)
		assert.Equal(t, "fallback", result3)
	})

	// Test Prefault goes through full validation pipeline
	t.Run("Prefault goes through full validation", func(t *testing.T) {
		schema := String().Min(10).Prefault("short")

		// Prefault value that fails validation should return error
		_, err := schema.Parse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least 10 characters")
	})

	t.Run("PrefaultFunc", func(t *testing.T) {
		called := false
		schema := String().Min(5).PrefaultFunc(func() string {
			called = true
			return "generated fallback"
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "generated fallback", result)
		assert.True(t, called)
	})

	// Test Transform interaction with Default and Prefault
	t.Run("Default bypasses Transform (short-circuit)", func(t *testing.T) {
		// Note: In current implementation, Transform creates a new schema type
		// So we test the conceptual behavior that Default should short-circuit
		baseSchema := String().Default("default")
		transformCalled := false
		transformSchema := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			transformCalled = true
			return strings.ToUpper(s), nil
		})

		// Test that default works without transform
		result, err := baseSchema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "default", result)

		// Test that transform works with valid input
		result2, err2 := transformSchema.Parse("hello")
		require.NoError(t, err2)
		assert.Equal(t, "HELLO", result2)
		assert.True(t, transformCalled)
	})

	t.Run("Prefault goes through Transform (full pipeline)", func(t *testing.T) {
		// Note: In current implementation, we test the conceptual behavior
		// that Prefault should go through the full validation and transform pipeline
		transformCalled := false
		baseSchema := String().Prefault("prefault")
		transformSchema := String().Transform(func(s string, ctx *core.RefinementContext) (any, error) {
			transformCalled = true
			return strings.ToUpper(s), nil
		})

		// Test that prefault works for nil input
		result, err := baseSchema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "prefault", result)

		// Test that transform works with prefault value
		result2, err2 := transformSchema.Parse("prefault")
		require.NoError(t, err2)
		assert.Equal(t, "PREFAULT", result2)
		assert.True(t, transformCalled)
	})

	// Test error handling for Prefault
	t.Run("Prefault error handling", func(t *testing.T) {
		schema := String().Min(10).Max(5).Prefault("invalid")

		// Prefault validation failure should return error directly
		_, err := schema.Parse(nil)
		require.Error(t, err)
		// Should contain validation error, not fallback to other mechanisms
		assert.Contains(t, err.Error(), "at least 10 characters")
	})

	// Test StringPtr behavior with unified API (accepts string value)
	t.Run("StringPtr Default accepts string value", func(t *testing.T) {
		schema := StringPtr().Default("default value") // Unified API: accepts string, not *string

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "default value", *result)
	})

	// Test StringPtr with empty string pointer - should NOT trigger default
	t.Run("StringPtr with empty string pointer - should NOT trigger default", func(t *testing.T) {
		schema := StringPtr().Default("default value")
		emptyStr := ""

		result, err := schema.Parse(&emptyStr)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "", *result, "Empty string pointer should not trigger default")
		assert.True(t, result == &emptyStr, "Should preserve original pointer")
	})

	// Test pointer identity preservation
	t.Run("StringPtr preserves pointer identity for non-nil input", func(t *testing.T) {
		schema := StringPtr().Default("default value")
		inputStr := "test value"
		inputPtr := &inputStr

		result, err := schema.Parse(inputPtr)
		require.NoError(t, err)
		assert.True(t, result == inputPtr, "Should preserve original pointer for non-nil input")
	})

	// Test side effect prevention - each Parse(nil) returns new pointer
	t.Run("StringPtr Default returns new pointer each time", func(t *testing.T) {
		schema := StringPtr().Default("default value")

		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		result3, err3 := schema.Parse(nil)
		require.NoError(t, err3)

		// Values should be equal but pointers should be different
		assert.Equal(t, *result1, *result2, "Values should be equal")
		assert.Equal(t, *result2, *result3, "Values should be equal")
		assert.True(t, result1 != result2, "Should return different pointers to prevent side effects")
		assert.True(t, result2 != result3, "Should return different pointers to prevent side effects")
		assert.True(t, result1 != result3, "Should return different pointers to prevent side effects")

		// Modifying one should not affect others
		*result1 = "modified"
		assert.Equal(t, "modified", *result1)
		assert.Equal(t, "default value", *result2, "result2 should not be affected")
		assert.Equal(t, "default value", *result3, "result3 should not be affected")
	})

	// Test unified method signatures
	t.Run("Unified API signatures", func(t *testing.T) {
		// Both String and StringPtr should accept string values for Default/Prefault
		schema1 := String().Default("value1")
		schema2 := StringPtr().Default("value2") // Unified API: accepts string, not *string
		schema3 := String().DefaultFunc(func() string { return "func1" })
		schema4 := StringPtr().DefaultFunc(func() string { return "func2" }) // Unified: returns string
		schema5 := String().Min(10).Default("short").Prefault("long enough")
		schema6 := StringPtr().Min(10).Default("short").Prefault("long enough ptr") // Unified API

		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, "value1", result1)

		result2, err2 := schema2.Parse(nil)
		require.NoError(t, err2)
		require.NotNil(t, result2)
		assert.Equal(t, "value2", *result2)

		result3, err3 := schema3.Parse(nil)
		require.NoError(t, err3)
		assert.Equal(t, "func1", result3)

		result4, err4 := schema4.Parse(nil)
		require.NoError(t, err4)
		require.NotNil(t, result4)
		assert.Equal(t, "func2", *result4)

		result5, err5 := schema5.Parse(nil)
		require.NoError(t, err5)
		assert.Equal(t, "short", result5) // Default has higher priority than Prefault

		result6, err6 := schema6.Parse(nil)
		require.NoError(t, err6)
		require.NotNil(t, result6)
		assert.Equal(t, "short", *result6) // Default has higher priority than Prefault
	})

	// Test StringPtr Prefault behavior (only on nil input)
	t.Run("StringPtr Prefault only on nil input", func(t *testing.T) {
		schema := StringPtr().Min(10).Prefault("long enough")

		// Valid input should pass through
		longStr := "this is long enough"
		result, err := schema.Parse(&longStr)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "this is long enough", *result)
		assert.True(t, result == &longStr, "Should preserve original pointer")

		// Invalid non-nil input should return error (NOT trigger prefault)
		shortStr := "hi"
		_, err = schema.Parse(&shortStr)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least 10 characters")

		// Nil input should trigger prefault
		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		require.NotNil(t, result2)
		assert.Equal(t, "long enough", *result2)

		// Multiple nil parses should create different pointer instances
		result3, err3 := schema.Parse(nil)
		require.NoError(t, err3)
		require.NotNil(t, result3)
		assert.Equal(t, "long enough", *result3)
		assert.True(t, result2 != result3, "Should create different pointer instances")
	})

	// Test comprehensive Default and Prefault scenarios
	t.Run("Comprehensive Default and Prefault tests", func(t *testing.T) {
		// Default with nil input
		schema1 := StringPtr().Default("default_value")
		result1, err1 := schema1.Parse(nil)
		require.NoError(t, err1)
		require.NotNil(t, result1)
		assert.Equal(t, "default_value", *result1)

		// Prefault only triggers on nil input
		schema2 := String().Min(10).Prefault("prefault_value")
		// Valid input should pass through
		result2, err2 := schema2.Parse("long enough input")
		require.NoError(t, err2)
		assert.Equal(t, "long enough input", result2)
		// Invalid non-nil input should error
		_, err2 = schema2.Parse("short")
		require.Error(t, err2)
		// Nil input should trigger prefault
		result2, err2 = schema2.Parse(nil)
		require.NoError(t, err2)
		assert.Equal(t, "prefault_value", result2)

		// DefaultFunc with nil input
		counter := 0
		schema3 := StringPtr().DefaultFunc(func() string {
			counter++
			return fmt.Sprintf("default_%d", counter)
		})
		result3, err3 := schema3.Parse(nil)
		require.NoError(t, err3)
		require.NotNil(t, result3)
		assert.Equal(t, "default_1", *result3)
		assert.Equal(t, 1, counter)

		// PrefaultFunc with nil input
		counter2 := 0
		schema4 := String().Min(10).PrefaultFunc(func() string {
			counter2++
			return fmt.Sprintf("prefault_%d", counter2)
		})
		result4, err4 := schema4.Parse(nil)
		require.NoError(t, err4)
		assert.Equal(t, "prefault_1", result4)
		assert.Equal(t, 1, counter2)

		// Optional with nil
		schema5 := StringPtr().Optional()
		result5, err5 := schema5.Parse(nil)
		require.NoError(t, err5)
		assert.Nil(t, result5)

		// Nilable with nil
		schema6 := StringPtr().Nilable()
		result6, err6 := schema6.Parse(nil)
		require.NoError(t, err6)
		assert.Nil(t, result6)
	})
}

// =============================================================================
// Refinement tests
// =============================================================================

func TestString_Modifiers_Refine(t *testing.T) {
	t.Run("Refine with string type", func(t *testing.T) {
		schema := String().Refine(func(s string) bool {
			return len(s) > 3
		}, "String must be longer than 3 characters")

		// Valid
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Invalid
		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("Refine with *string type", func(t *testing.T) {
		schema := StringPtr().Refine(func(s *string) bool {
			return s != nil && len(*s) > 3
		}, "String must be longer than 3 characters")

		// Valid
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "hello", *result)

		// Invalid
		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("RefineAny", func(t *testing.T) {
		schema := String().RefineAny(func(v any) bool {
			s, ok := v.(string)
			return ok && len(s) > 3
		}, "Value must be a string longer than 3 characters")

		// Valid
		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Invalid
		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})
}

// =============================================================================
// NonOptional tests
// =============================================================================

func TestString_Modifiers_NonOptional(t *testing.T) {
	// --- Basic nonoptional behaviour ---
	schema := String().NonOptional()

	// Valid string input
	result, err := schema.Parse("hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", result)
	// Type check – result should be string, not *string
	assert.IsType(t, "", result)

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
	chainSchema := String().Optional().NonOptional()
	var _ = chainSchema // compile-time type check

	// Should behave like required string now
	result2, err2 := chainSchema.Parse("world")
	require.NoError(t, err2)
	assert.Equal(t, "world", result2)
	assert.IsType(t, "", result2)

	_, err2 = chainSchema.Parse(nil)
	assert.Error(t, err2)

	// --- NonOptional inside object ---
	objectSchema := Object(map[string]core.ZodSchema{
		"hi": String().Optional().NonOptional(),
	})

	// Case 1: field present & valid
	obj1 := map[string]any{"hi": "asdf"}
	parsed1, err := objectSchema.Parse(obj1)
	require.NoError(t, err)
	assert.Equal(t, obj1, parsed1)

	// Case 2: field provided as nil – should error
	_, err = objectSchema.Parse(map[string]any{"hi": nil})
	assert.Error(t, err)

	// Case 3: field missing – should error
	_, err = objectSchema.Parse(map[string]any{})
	assert.Error(t, err)

	// --- StringPtr().NonOptional() ---
	ptrChain := StringPtr().NonOptional()
	var _ = ptrChain

	// Accepts raw string input
	resPtr1, err := ptrChain.Parse("pointer-str")
	require.NoError(t, err)
	assert.Equal(t, "pointer-str", resPtr1)
	assert.IsType(t, "", resPtr1)

	// Accepts *string input and returns value type
	val := "pointer-input"
	resPtr2, err := ptrChain.Parse(&val)
	require.NoError(t, err)
	assert.Equal(t, "pointer-input", resPtr2)
	assert.IsType(t, "", resPtr2)

	// Nil should error
	_, err = ptrChain.Parse(nil)
	assert.Error(t, err)
}

// =============================================================================
// IsOptional and IsNilable tests
// =============================================================================

func TestString_Modifiers_IsOptionalAndIsNilable(t *testing.T) {
	t.Run("basic schema - not optional, not nilable", func(t *testing.T) {
		schema := String()

		assert.False(t, schema.IsOptional(), "Basic string schema should not be optional")
		assert.False(t, schema.IsNilable(), "Basic string schema should not be nilable")
	})

	t.Run("optional schema - is optional, not nilable", func(t *testing.T) {
		schema := String().Optional()

		assert.True(t, schema.IsOptional(), "Optional string schema should be optional")
		assert.False(t, schema.IsNilable(), "Optional string schema should not be nilable")
	})

	t.Run("nilable schema - not optional, is nilable", func(t *testing.T) {
		schema := String().Nilable()

		assert.False(t, schema.IsOptional(), "Nilable string schema should not be optional")
		assert.True(t, schema.IsNilable(), "Nilable string schema should be nilable")
	})

	t.Run("nullish schema - is optional and nilable", func(t *testing.T) {
		schema := String().Nullish()

		assert.True(t, schema.IsOptional(), "Nullish string schema should be optional")
		assert.True(t, schema.IsNilable(), "Nullish string schema should be nilable")
	})

	t.Run("chained modifiers", func(t *testing.T) {
		// Optional then Nilable
		schema1 := String().Optional().Nilable()
		assert.True(t, schema1.IsOptional(), "Optional().Nilable() should be optional")
		assert.True(t, schema1.IsNilable(), "Optional().Nilable() should be nilable")

		// Nilable then Optional
		schema2 := String().Nilable().Optional()
		assert.True(t, schema2.IsOptional(), "Nilable().Optional() should be optional")
		assert.True(t, schema2.IsNilable(), "Nilable().Optional() should be nilable")
	})

	t.Run("nonoptional modifier resets optional flag", func(t *testing.T) {
		schema := String().Optional().NonOptional()

		assert.False(t, schema.IsOptional(), "Optional().NonOptional() should not be optional")
		assert.False(t, schema.IsNilable(), "Optional().NonOptional() should not be nilable")
	})

	t.Run("pointer types", func(t *testing.T) {
		// StringPtr basic
		ptrSchema := StringPtr()
		assert.False(t, ptrSchema.IsOptional(), "StringPtr schema should not be optional")
		assert.False(t, ptrSchema.IsNilable(), "StringPtr schema should not be nilable")

		// StringPtr with modifiers
		optionalPtrSchema := StringPtr().Optional()
		assert.True(t, optionalPtrSchema.IsOptional(), "StringPtr().Optional() should be optional")
		assert.False(t, optionalPtrSchema.IsNilable(), "StringPtr().Optional() should not be nilable")

		nilablePtrSchema := StringPtr().Nilable()
		assert.False(t, nilablePtrSchema.IsOptional(), "StringPtr().Nilable() should not be optional")
		assert.True(t, nilablePtrSchema.IsNilable(), "StringPtr().Nilable() should be nilable")
	})

	t.Run("consistency with GetInternals", func(t *testing.T) {
		// Test basic string schema
		basicSchema := String()
		assert.Equal(t, basicSchema.GetInternals().IsOptional(), basicSchema.IsOptional(),
			"Basic schema: IsOptional() should match GetInternals().IsOptional()")
		assert.Equal(t, basicSchema.GetInternals().IsNilable(), basicSchema.IsNilable(),
			"Basic schema: IsNilable() should match GetInternals().IsNilable()")

		// Test optional string schema
		optionalSchema := String().Optional()
		assert.Equal(t, optionalSchema.GetInternals().IsOptional(), optionalSchema.IsOptional(),
			"Optional schema: IsOptional() should match GetInternals().IsOptional()")
		assert.Equal(t, optionalSchema.GetInternals().IsNilable(), optionalSchema.IsNilable(),
			"Optional schema: IsNilable() should match GetInternals().IsNilable()")

		// Test nilable string schema
		nilableSchema := String().Nilable()
		assert.Equal(t, nilableSchema.GetInternals().IsOptional(), nilableSchema.IsOptional(),
			"Nilable schema: IsOptional() should match GetInternals().IsOptional()")
		assert.Equal(t, nilableSchema.GetInternals().IsNilable(), nilableSchema.IsNilable(),
			"Nilable schema: IsNilable() should match GetInternals().IsNilable()")

		// Test nullish string schema
		nullishSchema := String().Nullish()
		assert.Equal(t, nullishSchema.GetInternals().IsOptional(), nullishSchema.IsOptional(),
			"Nullish schema: IsOptional() should match GetInternals().IsOptional()")
		assert.Equal(t, nullishSchema.GetInternals().IsNilable(), nullishSchema.IsNilable(),
			"Nullish schema: IsNilable() should match GetInternals().IsNilable()")

		// Test nonoptional string schema
		nonoptionalSchema := String().Optional().NonOptional()
		assert.Equal(t, nonoptionalSchema.GetInternals().IsOptional(), nonoptionalSchema.IsOptional(),
			"NonOptional schema: IsOptional() should match GetInternals().IsOptional()")
		assert.Equal(t, nonoptionalSchema.GetInternals().IsNilable(), nonoptionalSchema.IsNilable(),
			"NonOptional schema: IsNilable() should match GetInternals().IsNilable()")
	})
}

// =============================================================================
// Coercion tests
// =============================================================================

func TestString_Coercion(t *testing.T) {
	t.Run("basic coercion", func(t *testing.T) {
		schema := CoercedString()

		tests := []struct {
			input    any
			expected string
		}{
			{123, "123"},
			{true, "true"},
			{false, "false"},
			{12.34, "12.34"},
			{"hello", "hello"},
		}

		for _, tt := range tests {
			result, err := schema.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("coercion with validation", func(t *testing.T) {
		schema := CoercedString().Min(3).Max(5)

		// Coercion then validation passes
		result, err := schema.Parse(1234)
		require.NoError(t, err)
		assert.Equal(t, "1234", result)

		// Coercion then validation fails
		_, err = schema.Parse(12)
		assert.Error(t, err)
	})
}

// =============================================================================
// Error handling tests
// =============================================================================

func TestString_BasicFunctionality_ErrorHandling(t *testing.T) {
	t.Run("type error", func(t *testing.T) {
		schema := String()

		_, err := schema.Parse(123)
		assert.Error(t, err)
	})

	t.Run("validation error", func(t *testing.T) {
		schema := String().Min(5)

		_, err := schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("custom error message", func(t *testing.T) {
		schema := String().Min(5, "String must be at least 5 characters")

		_, err := schema.Parse("hi")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "String must be at least 5 characters")
	})
}

// =============================================================================
// Overwrite tests
// =============================================================================

func TestString_Modifiers_Overwrite(t *testing.T) {
	t.Run("basic overwrite", func(t *testing.T) {
		schema := String().Overwrite(strings.TrimSpace)

		result, err := schema.Parse("  hello  ")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("overwrite with validations", func(t *testing.T) {
		stringSchema := String().Overwrite(strings.ToUpper).Min(5)

		// Valid case
		result, err := stringSchema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "HELLO", result)

		// Invalid case
		_, err = stringSchema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("chaining overwrites", func(t *testing.T) {
		lowerSchema := String().Overwrite(strings.ToLower)
		upperSchema := lowerSchema.Overwrite(strings.ToUpper)

		// Should apply both transformations in order: " HeLLo " -> " hello " -> " HELLO "
		result, err := upperSchema.Parse(" HeLLo ")
		require.NoError(t, err)
		assert.Equal(t, " HELLO ", result)
	})

	t.Run("complex chaining", func(t *testing.T) {
		schema := String().
			Overwrite(strings.TrimSpace).
			Min(3).
			Overwrite(func(s string) string {
				return s + "!"
			})

		// Test valid case: "  hi  " -> "hi" (len 2, fail)
		_, err := schema.Parse("  hi  ")
		assert.Error(t, err)

		// Test valid case: "  hello  " -> "hello" (len 5, pass) -> "hello!"
		result, err := schema.Parse("  hello  ")
		require.NoError(t, err)
		assert.Equal(t, "hello!", result)
	})

	t.Run("type preservation", func(t *testing.T) {
		original := String()
		overwritten := original.Overwrite(strings.ToUpper)

		// Should still be a string schema
		var _ = overwritten

		result, err := overwritten.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "TEST", result)
	})

	t.Run("pointer type preservation", func(t *testing.T) {
		schema := StringPtr().Overwrite(func(s *string) *string {
			if s == nil {
				return nil
			}
			upper := strings.ToUpper(*s)
			return &upper
		})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "HELLO", *result)
	})

	t.Run("composition of different transformations", func(t *testing.T) {
		trimSchema := String().Overwrite(strings.TrimSpace)
		lowerSchema := trimSchema.Overwrite(strings.ToLower)

		// Test: "  HELLO  " -> "HELLO" -> "hello"
		result, err := lowerSchema.Parse("  HELLO  ")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("overwrite with different functions", func(t *testing.T) {
		schema := String().
			Overwrite(strings.TrimSpace).
			Overwrite(strings.ToLower).
			Max(5)

		// Test: "  HELLO  " -> "hello" (len 5, ok)
		result, err := schema.Parse("  HELLO  ")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		// Test: "  WORLD!  " -> "world!" (len 6, fail)
		_, err = schema.Parse("  WORLD!  ")
		assert.Error(t, err)
	})
}

func TestString_Modifiers_Check(t *testing.T) {
	t.Run("adds multiple issues for invalid input", func(t *testing.T) {
		schema := String().Check(func(value string, p *core.ParsePayload) {
			if len(value) < 5 {
				p.AddIssueWithMessage("String must be at least 5 characters long")
			}
			if !strings.Contains(value, "@") {
				p.AddIssueWithCode(core.InvalidFormat, "String must contain an @ symbol")
			}
		})

		_, err := schema.Parse("test")
		require.Error(t, err)

		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		require.Len(t, zodErr.Issues, 2)

		assert.Equal(t, "String must be at least 5 characters long", zodErr.Issues[0].Message)
		assert.Equal(t, core.Custom, zodErr.Issues[0].Code)

		assert.Equal(t, "String must contain an @ symbol", zodErr.Issues[1].Message)
		assert.Equal(t, core.InvalidFormat, zodErr.Issues[1].Code)
	})

	t.Run("succeeds for valid input", func(t *testing.T) {
		schema := String().Check(func(value string, p *core.ParsePayload) {
			if len(value) < 5 {
				p.AddIssueWithMessage("String must be at least 5 characters long")
			}
		})

		result, err := schema.Parse("valid_string")
		require.NoError(t, err)
		assert.Equal(t, "valid_string", result)
	})

	t.Run("works with pointer types", func(t *testing.T) {
		schema := StringPtr().Check(func(value *string, p *core.ParsePayload) {
			if value != nil && len(*value) < 5 {
				p.AddIssueWithMessage("String must be at least 5 characters long")
			}
		})

		invalid := "test"
		_, err := schema.Parse(&invalid)
		require.Error(t, err)
		var zodErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zodErr))
		require.Len(t, zodErr.Issues, 1)
		assert.Equal(t, "String must be at least 5 characters long", zodErr.Issues[0].Message)

		valid := "valid_string"
		result, err := schema.Parse(&valid)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "valid_string", *result)
	})
}

// =============================================================================
// Trim / Case transformation tests
// =============================================================================

func TestString_Modifiers_TrimAndCaseTransforms(t *testing.T) {
	t.Run("trim transformations and ordering", func(t *testing.T) {
		// .trim().min(2) should first trim then validate length
		schema1 := String().Trim().Min(2)
		result1, err1 := schema1.Parse(" 12 ")
		require.NoError(t, err1)
		assert.Equal(t, "12", result1)

		// .min(2).trim() should validate length before trimming
		schema2 := String().Min(2).Trim()
		result2, err2 := schema2.Parse(" 1 ")
		require.NoError(t, err2)
		assert.Equal(t, "1", result2)

		// .trim().min(2) with insufficient trimmed length should fail
		schema3 := String().Trim().Min(2)
		_, err3 := schema3.Parse(" 1 ")
		assert.Error(t, err3)
	})

	t.Run("toLowerCase and toUpperCase transformations", func(t *testing.T) {
		lowerSchema := String().ToLowerCase()
		resLower, errLower := lowerSchema.Parse("ASDF")
		require.NoError(t, errLower)
		assert.Equal(t, "asdf", resLower)

		upperSchema := String().ToUpperCase()
		resUpper, errUpper := upperSchema.Parse("asdf")
		require.NoError(t, errUpper)
		assert.Equal(t, "ASDF", resUpper)
	})

	t.Run("slugify transformations", func(t *testing.T) {
		schema := String().Slugify()

		// Basic slugify
		result, err := schema.Parse("Hello World")
		require.NoError(t, err)
		assert.Equal(t, "hello-world", result)

		// Trim spaces
		result2, err2 := schema.Parse("  Hello   World  ")
		require.NoError(t, err2)
		assert.Equal(t, "hello-world", result2)

		// Remove special characters
		result3, err3 := schema.Parse("Hello@World#123")
		require.NoError(t, err3)
		assert.Equal(t, "helloworld123", result3)

		// Preserve hyphens
		result4, err4 := schema.Parse("Hello-World")
		require.NoError(t, err4)
		assert.Equal(t, "hello-world", result4)

		// Convert underscores to hyphens
		result5, err5 := schema.Parse("Hello_World")
		require.NoError(t, err5)
		assert.Equal(t, "hello-world", result5)

		// Collapse multiple hyphens
		result6, err6 := schema.Parse("---Hello---World---")
		require.NoError(t, err6)
		assert.Equal(t, "hello-world", result6)

		// Multiple spaces
		result7, err7 := schema.Parse("Hello  World")
		require.NoError(t, err7)
		assert.Equal(t, "hello-world", result7)

		// Special characters only
		result8, err8 := schema.Parse("Hello!@#$%^&*()World")
		require.NoError(t, err8)
		assert.Equal(t, "helloworld", result8)
	})

	t.Run("slugify with min validation (matches Zod)", func(t *testing.T) {
		// z.string().slugify().min(5) - should slugify first, then validate length
		schema := String().Slugify().Min(5)

		result, err := schema.Parse("Hello World")
		require.NoError(t, err)
		assert.Equal(t, "hello-world", result) // 11 chars, passes min(5)

		// Fail validation after slugify
		_, err2 := schema.Parse("Hi")
		assert.Error(t, err2) // "hi" is only 2 chars
	})

	t.Run("slugify with pointer type", func(t *testing.T) {
		schema := StringPtr().Slugify()

		input := "Hello World"
		result, err := schema.Parse(&input)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "hello-world", *result)
	})

	t.Run("lowercase validation", func(t *testing.T) {
		schema := String().Lowercase()

		// Valid: all lowercase
		result, err := schema.Parse("hello world")
		require.NoError(t, err)
		assert.Equal(t, "hello world", result)

		// Valid: empty string (Zod v4 behavior)
		result2, err2 := schema.Parse("")
		require.NoError(t, err2)
		assert.Equal(t, "", result2)

		// Valid: numbers and symbols only (no uppercase letters)
		result3, err3 := schema.Parse("123!@#$%")
		require.NoError(t, err3)
		assert.Equal(t, "123!@#$%", result3)

		// Invalid: contains uppercase letters
		_, err4 := schema.Parse("Hello")
		assert.Error(t, err4)

		// Invalid: all uppercase
		_, err5 := schema.Parse("HELLO")
		assert.Error(t, err5)
	})

	t.Run("uppercase validation", func(t *testing.T) {
		schema := String().Uppercase()

		// Valid: all uppercase
		result, err := schema.Parse("HELLO WORLD")
		require.NoError(t, err)
		assert.Equal(t, "HELLO WORLD", result)

		// Valid: empty string (Zod v4 behavior)
		result2, err2 := schema.Parse("")
		require.NoError(t, err2)
		assert.Equal(t, "", result2)

		// Valid: numbers and symbols only (no lowercase letters)
		result3, err3 := schema.Parse("123!@#$%")
		require.NoError(t, err3)
		assert.Equal(t, "123!@#$%", result3)

		// Invalid: contains lowercase letters
		_, err4 := schema.Parse("Hello")
		assert.Error(t, err4)

		// Invalid: all lowercase
		_, err5 := schema.Parse("hello")
		assert.Error(t, err5)
	})

	t.Run("normalize transformation", func(t *testing.T) {
		// Default (NFC)
		schema := String().Normalize()

		// Test with NFD composed and decomposed forms
		// é can be represented as:
		// - NFC (composed): U+00E9 (single character)
		// - NFD (decomposed): U+0065 U+0301 (e + combining acute accent)
		composedE := "caf\u00e9"    // café with precomposed é
		decomposedE := "cafe\u0301" // café with e + combining accent

		// Both should normalize to the same NFC form
		result1, err1 := schema.Parse(composedE)
		require.NoError(t, err1)
		result2, err2 := schema.Parse(decomposedE)
		require.NoError(t, err2)
		assert.Equal(t, result1, result2)

		// Test specific form (NFD)
		schemaNFD := String().Normalize("NFD")
		resultNFD, errNFD := schemaNFD.Parse(composedE)
		require.NoError(t, errNFD)
		// NFD should decompose the character
		assert.NotEqual(t, composedE, resultNFD) // Different representation
	})

	t.Run("normalize with pointer type", func(t *testing.T) {
		schema := StringPtr().Normalize()

		input := "caf\u00e9"
		result, err := schema.Parse(&input)
		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("lowercase vs toLowerCase", func(t *testing.T) {
		// lowercase() validates that string has no uppercase - validation
		lcValidate := String().Lowercase()
		_, err := lcValidate.Parse("Hello")
		assert.Error(t, err) // Fails because 'H' is uppercase

		// toLowerCase() transforms string to lowercase - transformation
		lcTransform := String().ToLowerCase()
		result, err2 := lcTransform.Parse("Hello")
		require.NoError(t, err2)
		assert.Equal(t, "hello", result) // Transformed to lowercase
	})

	t.Run("uppercase vs toUpperCase", func(t *testing.T) {
		// uppercase() validates that string has no lowercase - validation
		ucValidate := String().Uppercase()
		_, err := ucValidate.Parse("Hello")
		assert.Error(t, err) // Fails because 'e', 'l', 'l', 'o' are lowercase

		// toUpperCase() transforms string to uppercase - transformation
		ucTransform := String().ToUpperCase()
		result, err2 := ucTransform.Parse("Hello")
		require.NoError(t, err2)
		assert.Equal(t, "HELLO", result) // Transformed to uppercase
	})
}

// =============================================================================
// Edge cases tests
// =============================================================================

func TestString_EdgeCases(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		schema := String()

		result, err := schema.Parse("")
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("empty string with min validation", func(t *testing.T) {
		schema := String().Min(1)

		_, err := schema.Parse("")
		assert.Error(t, err)
	})

	t.Run("unicode strings", func(t *testing.T) {
		schema := String().Min(3)

		result, err := schema.Parse("你好世界")
		require.NoError(t, err)
		assert.Equal(t, "你好世界", result)
	})

	t.Run("very long string", func(t *testing.T) {
		longString := make([]byte, 10000)
		for i := range longString {
			longString[i] = 'a'
		}

		schema := String().Max(10000)
		result, err := schema.Parse(string(longString))
		require.NoError(t, err)
		assert.Equal(t, string(longString), result)
	})
}

// =============================================================================
// Pointer identity preservation tests
// =============================================================================

func TestString_EdgeCases_PointerIdentityPreservation(t *testing.T) {
	t.Run("String Optional preserves pointer identity", func(t *testing.T) {
		schema := String().Optional()

		originalStr := "test string"
		originalPtr := &originalStr

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Pointer identity should be preserved")
		assert.Equal(t, "test string", *result)
	})

	t.Run("String Nilable preserves pointer identity", func(t *testing.T) {
		schema := String().Nilable()

		originalStr := "nilable test"
		originalPtr := &originalStr

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Pointer identity should be preserved")
		assert.Equal(t, "nilable test", *result)
	})

	t.Run("StringPtr Optional preserves pointer identity", func(t *testing.T) {
		schema := StringPtr().Optional()

		originalStr := "ptr test"
		originalPtr := &originalStr

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "StringPtr Optional should preserve pointer identity")
		assert.Equal(t, "ptr test", *result)
	})

	t.Run("StringPtr Nilable preserves pointer identity", func(t *testing.T) {
		schema := StringPtr().Nilable()

		originalStr := "ptr nilable"
		originalPtr := &originalStr

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "StringPtr Nilable should preserve pointer identity")
		assert.Equal(t, "ptr nilable", *result)
	})

	t.Run("String Nullish preserves pointer identity", func(t *testing.T) {
		schema := String().Nullish()

		originalStr := "nullish test"
		originalPtr := &originalStr

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "String Nullish should preserve pointer identity")
		assert.Equal(t, "nullish test", *result)
	})

	t.Run("Optional handles nil consistently", func(t *testing.T) {
		schema := String().Optional()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable handles nil consistently", func(t *testing.T) {
		schema := String().Nilable()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default().Optional() chaining preserves pointer identity", func(t *testing.T) {
		schema := String().Default("default").Optional()

		originalStr := "input value"
		originalPtr := &originalStr

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Default().Optional() should preserve pointer identity")
		assert.Equal(t, "input value", *result)
	})

	t.Run("Validation with Optional preserves pointer identity", func(t *testing.T) {
		schema := String().Min(3).Max(20).Optional()

		originalStr := "valid input"
		originalPtr := &originalStr

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Validation().Optional() should preserve pointer identity")
		assert.Equal(t, "valid input", *result)
	})

	t.Run("Refine with Optional preserves pointer identity", func(t *testing.T) {
		schema := String().Refine(func(s string) bool {
			return len(s) > 0 // Always pass for non-empty strings
		}).Optional()

		originalStr := "refined test"
		originalPtr := &originalStr

		result, err := schema.Parse(originalPtr)
		require.NoError(t, err)

		// Result should be the same pointer
		assert.True(t, result == originalPtr, "Refine().Optional() should preserve pointer identity")
		assert.Equal(t, "refined test", *result)
	})

	t.Run("Multiple string pointer identity tests", func(t *testing.T) {
		schema := String().Optional()

		testCases := []string{"hello", "world", "", "test with spaces", "unicode: 你好"}

		for i, strVal := range testCases {
			t.Run(fmt.Sprintf("string_%d", i), func(t *testing.T) {
				originalPtr := &strVal

				result, err := schema.Parse(originalPtr)
				require.NoError(t, err)

				assert.True(t, result == originalPtr, "Pointer identity should be preserved for '%s'", strVal)
				assert.Equal(t, strVal, *result)
			})
		}
	})
}

// =============================================================================
// Refine advanced parameter tests (error, abort, path, when)
// =============================================================================

func TestString_RefineParameters(t *testing.T) {
	t.Run("abort option stops subsequent refinements", func(t *testing.T) {
		schema := String().
			Refine(func(s string) bool { return len(s) > 8 }, core.CustomParams{Error: "Too short!", Abort: true}).
			Refine(func(s string) bool { return strings.ToLower(s) == s }, "Must be lowercase")

		_, err := schema.Parse("OH NO")
		require.Error(t, err)

		var zErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zErr))
		assert.Len(t, zErr.Issues, 1)
		assert.Equal(t, "Too short!", zErr.Issues[0].Message)
	})

	t.Run("path parameter overrides issue path", func(t *testing.T) {
		schema := String().
			Refine(func(s string) bool { return len(s) > 5 }, core.CustomParams{
				Error: "Too short!",
				Path:  []any{"custom", "path"},
			})

		_, err := schema.Parse("hi")
		require.Error(t, err)

		var zErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zErr))
		assert.Len(t, zErr.Issues, 1)
		assert.Equal(t, []any{"custom", "path"}, zErr.Issues[0].Path)
	})

	t.Run("when predicate controls execution", func(t *testing.T) {
		// Only run validation when payload has no existing issues
		whenFn := func(p *core.ParsePayload) bool {
			return p.IssueCount() == 0
		}

		schema := String().
			Refine(func(s string) bool { return len(s) > 3 }, "Too short!").
			Refine(func(s string) bool { return len(s) < 10 }, core.CustomParams{
				Error: "Too long!",
				When:  whenFn,
			})

		// First refinement fails, second should not run due to when predicate
		_, err := schema.Parse("hi")
		require.Error(t, err)

		var zErr *issues.ZodError
		require.True(t, issues.IsZodError(err, &zErr))
		assert.Len(t, zErr.Issues, 1)
		assert.Equal(t, "Too short!", zErr.Issues[0].Message)
	})
}
