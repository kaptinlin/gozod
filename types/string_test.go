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
	t.Run("valid string inputs", func(t *testing.T) {
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
// Type safety tests
// =============================================================================

func TestString_TypeSafety(t *testing.T) {
	t.Run("String returns string type", func(t *testing.T) {
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
}

// =============================================================================
// Modifier methods tests
// =============================================================================

func TestString_Modifiers(t *testing.T) {
	t.Run("Optional always returns *string", func(t *testing.T) {
		// From string to *string via Optional
		stringSchema := String()
		optionalSchema := stringSchema.Optional()

		// Type check: ensure it returns *ZodString[*string]
		var _ *ZodString[*string] = optionalSchema

		// Functionality test
		result, err := optionalSchema.Parse("hello")
		require.NoError(t, err)
		assert.IsType(t, (*string)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, "hello", *result)

		// From *string to *string via Optional (maintains type)
		ptrSchema := StringPtr()
		optionalPtrSchema := ptrSchema.Optional()
		var _ *ZodString[*string] = optionalPtrSchema
	})

	t.Run("Nilable always returns *string", func(t *testing.T) {
		stringSchema := String()
		nilableSchema := stringSchema.Nilable()

		var _ *ZodString[*string] = nilableSchema

		// Test nil handling
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default preserves current type", func(t *testing.T) {
		// string maintains string
		stringSchema := String()
		defaultStringSchema := stringSchema.Default("default")
		var _ *ZodString[string] = defaultStringSchema

		// *string maintains *string
		ptrSchema := StringPtr()
		defaultPtrSchema := ptrSchema.Default("default")
		var _ *ZodString[*string] = defaultPtrSchema
	})

	t.Run("Prefault preserves current type", func(t *testing.T) {
		// string maintains string
		stringSchema := String()
		prefaultStringSchema := stringSchema.Prefault("fallback")
		var _ *ZodString[string] = prefaultStringSchema

		// *string maintains *string
		ptrSchema := StringPtr()
		prefaultPtrSchema := ptrSchema.Prefault("fallback")
		var _ *ZodString[*string] = prefaultPtrSchema
	})

	t.Run("Nullish combines optional and nilable", func(t *testing.T) {
		stringSchema := String()
		nullishSchema := stringSchema.Nullish()

		var _ *ZodString[*string] = nullishSchema

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
}

// =============================================================================
// Validation methods tests
// =============================================================================

func TestString_BasicFunctionality_Validations(t *testing.T) {
	t.Run("length validations", func(t *testing.T) {
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
	t.Run("multiple validations", func(t *testing.T) {
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
		var _ *ZodString[string] = base

		withValidation := base.Min(3)
		var _ *ZodString[string] = withValidation

		optional := withValidation.Optional()
		var _ *ZodString[*string] = optional

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

	t.Run("Prefault with invalid default", func(t *testing.T) {
		schema := String().Min(5).Default("hi").Prefault("fallback value")

		// Default "hi" fails validation, should use prefault
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "fallback value", result)
	})

	t.Run("PrefaultFunc", func(t *testing.T) {
		called := false
		schema := String().Min(5).Default("hi").PrefaultFunc(func() string {
			called = true
			return "generated fallback"
		})

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "generated fallback", result)
		assert.True(t, called)
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
	var _ *ZodString[string] = chainSchema // compile-time type check

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
	var _ *ZodString[string] = ptrChain

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
		var _ *ZodString[string] = overwritten

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
