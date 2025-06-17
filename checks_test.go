package gozod

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// CONSTRUCTOR TESTS
// =============================================================================

func TestCheckConstructors(t *testing.T) {
	t.Run("creates check instances with proper internals", func(t *testing.T) {
		tests := []struct {
			name  string
			check func() ZodCheck
		}{
			{"LessThan", func() ZodCheck { return NewZodCheckLessThan(10, false) }},
			{"GreaterThan", func() ZodCheck { return NewZodCheckGreaterThan(5, false) }},
			{"MultipleOf", func() ZodCheck { return NewZodCheckMultipleOf(3) }},
			{"NumberFormat", func() ZodCheck { return NewZodCheckNumberFormat(NumberFormatInt32) }},
			{"MaxLength", func() ZodCheck { return NewZodCheckMaxLength(10) }},
			{"MinLength", func() ZodCheck { return NewZodCheckMinLength(1) }},
			{"Regex", func() ZodCheck { return NewZodCheckRegex(regexp.MustCompile(`^\d+$`)) }},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				check := tt.check()
				require.NotNil(t, check)

				internals := check.GetZod()
				require.NotNil(t, internals)
				require.NotNil(t, internals.Def)
				require.NotNil(t, internals.Check)
				assert.NotEmpty(t, internals.Def.Check)
			})
		}
	})

	t.Run("applies schema parameters correctly", func(t *testing.T) {
		customErrorMap := ZodErrorMap(func(issue ZodRawIssue) string {
			return "Custom validation error"
		})

		params := SchemaParams{
			Error: &customErrorMap,
			Abort: true,
		}

		check := NewZodCheckMaxLength(10, params)
		require.NotNil(t, check)

		internals := check.GetZod()
		assert.NotNil(t, internals.Def.Error)
		assert.True(t, internals.Def.Abort)
	})
}

// =============================================================================
// NUMERIC VALIDATION TESTS
// =============================================================================

func TestNumericChecks_LessThan(t *testing.T) {
	t.Run("validates values less than maximum", func(t *testing.T) {
		check := NewZodCheckLessThan(10, false)
		payload := NewParsePayload(5)

		internals := check.GetZod()
		internals.Check(payload)

		assert.Empty(t, payload.Issues, "Should accept values less than maximum")
	})

	t.Run("validates inclusive maximum", func(t *testing.T) {
		check := NewZodCheckLessThan(10, true)
		payload := NewParsePayload(10)

		internals := check.GetZod()
		internals.Check(payload)

		assert.Empty(t, payload.Issues, "Should accept equal value when inclusive")
	})

	t.Run("rejects values greater than maximum", func(t *testing.T) {
		check := NewZodCheckLessThan(10, false)
		payload := NewParsePayload(15)

		internals := check.GetZod()
		internals.Check(payload)

		require.Len(t, payload.Issues, 1)
		assert.Equal(t, string(TooBig), payload.Issues[0].Code)
	})

	t.Run("rejects equal value when exclusive", func(t *testing.T) {
		check := NewZodCheckLessThan(10, false)
		payload := NewParsePayload(10)

		internals := check.GetZod()
		internals.Check(payload)

		require.Len(t, payload.Issues, 1)
		assert.Equal(t, string(TooBig), payload.Issues[0].Code)
	})
}

func TestNumericChecks_GreaterThan(t *testing.T) {
	t.Run("validates values greater than minimum", func(t *testing.T) {
		check := NewZodCheckGreaterThan(5, false)
		payload := NewParsePayload(10)

		internals := check.GetZod()
		internals.Check(payload)

		assert.Empty(t, payload.Issues, "Should accept values greater than minimum")
	})

	t.Run("validates inclusive minimum", func(t *testing.T) {
		check := NewZodCheckGreaterThan(5, true)
		payload := NewParsePayload(5)

		internals := check.GetZod()
		internals.Check(payload)

		assert.Empty(t, payload.Issues, "Should accept equal value when inclusive")
	})

	t.Run("rejects values less than minimum", func(t *testing.T) {
		check := NewZodCheckGreaterThan(5, false)
		payload := NewParsePayload(3)

		internals := check.GetZod()
		internals.Check(payload)

		require.Len(t, payload.Issues, 1)
		assert.Equal(t, string(TooSmall), payload.Issues[0].Code)
	})

	t.Run("rejects equal value when exclusive", func(t *testing.T) {
		check := NewZodCheckGreaterThan(5, false)
		payload := NewParsePayload(5)

		internals := check.GetZod()
		internals.Check(payload)

		require.Len(t, payload.Issues, 1)
		assert.Equal(t, string(TooSmall), payload.Issues[0].Code)
	})
}

func TestNumericChecks_MultipleOf(t *testing.T) {
	t.Run("validates multiples correctly", func(t *testing.T) {
		tests := []struct {
			name    string
			divisor interface{}
			value   interface{}
			isValid bool
		}{
			{"integer multiple", 3, 9, true},
			{"non-multiple", 3, 10, false},
			{"zero value", 5, 0, true},
			{"float multiple", 2.5, 5.0, true},
			{"float non-multiple", 2.5, 6.0, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				check := NewZodCheckMultipleOf(tt.divisor)
				payload := NewParsePayload(tt.value)

				internals := check.GetZod()
				internals.Check(payload)

				if tt.isValid {
					assert.Empty(t, payload.Issues, "Should be valid multiple")
				} else {
					require.Len(t, payload.Issues, 1)
					assert.Equal(t, string(NotMultipleOf), payload.Issues[0].Code)
				}
			})
		}
	})
}

func TestNumericChecks_NumberFormat(t *testing.T) {
	t.Run("validates integer formats", func(t *testing.T) {
		check := NewZodCheckNumberFormat(NumberFormatInt32)

		tests := []struct {
			name    string
			value   interface{}
			isValid bool
		}{
			{"valid integer", float64(100), true},
			{"invalid float", 100.5, false},
			{"string input", "100", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload := NewParsePayload(tt.value)
				internals := check.GetZod()
				internals.Check(payload)

				if tt.isValid {
					assert.Empty(t, payload.Issues, "Should accept valid integer")
				} else {
					assert.NotEmpty(t, payload.Issues, "Should reject invalid input")
				}
			})
		}
	})

	t.Run("validates safe integer range", func(t *testing.T) {
		check := NewZodCheckNumberFormat(NumberFormatSafeint)
		payload := NewParsePayload(float64(9007199254740992)) // MAX_SAFE_INTEGER + 1

		internals := check.GetZod()
		internals.Check(payload)

		require.Len(t, payload.Issues, 1)
		assert.Equal(t, string(TooBig), payload.Issues[0].Code)
	})
}

//////////////////////////////////
/////    ZodCheckMaxLength    /////
//////////////////////////////////

func TestStringChecks_Length(t *testing.T) {
	t.Run("MaxLength validates string length constraints", func(t *testing.T) {
		tests := []struct {
			name    string
			maxLen  int
			input   string
			isValid bool
		}{
			{"within limit", 10, "hello", true},
			{"at limit", 5, "hello", true},
			{"exceeds limit", 3, "hello", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				check := NewZodCheckMaxLength(tt.maxLen)
				payload := NewParsePayload(tt.input)

				internals := check.GetZod()
				internals.Check(payload)

				if tt.isValid {
					assert.Empty(t, payload.Issues, "Should accept string within length limit")
				} else {
					require.Len(t, payload.Issues, 1)
					assert.Equal(t, string(TooBig), payload.Issues[0].Code)
				}
			})
		}
	})

	t.Run("MinLength validates minimum string length", func(t *testing.T) {
		tests := []struct {
			name    string
			minLen  int
			input   string
			isValid bool
		}{
			{"above minimum", 3, "hello", true},
			{"at minimum", 5, "hello", true},
			{"below minimum", 10, "hello", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				check := NewZodCheckMinLength(tt.minLen)
				payload := NewParsePayload(tt.input)

				internals := check.GetZod()
				internals.Check(payload)

				if tt.isValid {
					assert.Empty(t, payload.Issues, "Should accept string above minimum length")
				} else {
					require.Len(t, payload.Issues, 1)
					assert.Equal(t, string(TooSmall), payload.Issues[0].Code)
				}
			})
		}
	})

	t.Run("LengthEquals validates exact string length", func(t *testing.T) {
		check := NewZodCheckLengthEquals(5)

		tests := []struct {
			name    string
			input   string
			isValid bool
			expCode string
		}{
			{"exact length", "hello", true, ""},
			{"too long", "hello world", false, string(TooBig)},
			{"too short", "hi", false, string(TooSmall)},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload := NewParsePayload(tt.input)
				internals := check.GetZod()
				internals.Check(payload)

				if tt.isValid {
					assert.Empty(t, payload.Issues, "Should accept exact length")
				} else {
					require.Len(t, payload.Issues, 1)
					assert.Equal(t, tt.expCode, payload.Issues[0].Code)
				}
			})
		}
	})

	t.Run("skips check for nil values", func(t *testing.T) {
		check := NewZodCheckMaxLength(5)
		internals := check.GetZod()

		// Test when condition
		payload := NewParsePayload(nil)
		assert.False(t, internals.When(payload), "Should skip check for nil values")
	})
}

func TestStringChecks_Format(t *testing.T) {
	t.Run("validates string format with regex patterns", func(t *testing.T) {
		pattern := regexp.MustCompile(`^[a-z]+$`)
		check := NewZodCheckStringFormat(StringFormatLowercase, pattern)

		tests := []struct {
			name    string
			input   interface{}
			isValid bool
		}{
			{"matching pattern", "hello", true},
			{"non-matching pattern", "Hello", false},
			{"non-string input", 123, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload := NewParsePayload(tt.input)
				internals := check.GetZod()
				internals.Check(payload)

				if tt.isValid {
					assert.Empty(t, payload.Issues, "Should accept matching pattern")
				} else {
					require.Len(t, payload.Issues, 1)
					if _, ok := tt.input.(string); ok {
						assert.Equal(t, string(InvalidFormat), payload.Issues[0].Code)
					} else {
						assert.Equal(t, string(InvalidType), payload.Issues[0].Code)
					}
				}
			})
		}
	})

	t.Run("LowerCase validates lowercase strings", func(t *testing.T) {
		check := NewZodCheckLowerCase()

		tests := []struct {
			name    string
			input   string
			isValid bool
		}{
			{"lowercase string", "hello", true},
			{"mixed case string", "Hello", false},
			{"uppercase string", "HELLO", false},
			{"empty string", "", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload := NewParsePayload(tt.input)
				internals := check.GetZod()
				internals.Check(payload)

				if tt.isValid {
					assert.Empty(t, payload.Issues, "Should accept lowercase string")
				} else {
					require.Len(t, payload.Issues, 1)
					assert.Equal(t, string(InvalidFormat), payload.Issues[0].Code)
				}
			})
		}
	})

	t.Run("UpperCase validates uppercase strings", func(t *testing.T) {
		check := NewZodCheckUpperCase()

		tests := []struct {
			name    string
			input   string
			isValid bool
		}{
			{"uppercase string", "HELLO", true},
			{"mixed case string", "Hello", false},
			{"lowercase string", "hello", false},
			{"empty string", "", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload := NewParsePayload(tt.input)
				internals := check.GetZod()
				internals.Check(payload)

				if tt.isValid {
					assert.Empty(t, payload.Issues, "Should accept uppercase string")
				} else {
					require.Len(t, payload.Issues, 1)
					assert.Equal(t, string(InvalidFormat), payload.Issues[0].Code)
				}
			})
		}
	})
}

func TestStringChecks_Pattern(t *testing.T) {
	t.Run("Regex validates string patterns", func(t *testing.T) {
		pattern := regexp.MustCompile(`^\d+$`)
		check := NewZodCheckRegex(pattern)

		tests := []struct {
			name    string
			input   interface{}
			isValid bool
		}{
			{"matching regex", "12345", true},
			{"non-matching regex", "abc123", false},
			{"non-string input", 123, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload := NewParsePayload(tt.input)
				internals := check.GetZod()
				internals.Check(payload)

				if tt.isValid {
					assert.Empty(t, payload.Issues, "Should accept matching regex")
				} else {
					require.Len(t, payload.Issues, 1)
					if _, ok := tt.input.(string); ok {
						assert.Equal(t, string(InvalidFormat), payload.Issues[0].Code)
					} else {
						assert.Equal(t, string(InvalidType), payload.Issues[0].Code)
					}
				}
			})
		}
	})

	t.Run("Includes validates substring presence", func(t *testing.T) {
		check := NewZodCheckIncludes("world", nil)

		tests := []struct {
			name    string
			input   interface{}
			isValid bool
		}{
			{"contains substring", "hello world", true},
			{"missing substring", "hello there", false},
			{"non-string input", 123, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload := NewParsePayload(tt.input)
				internals := check.GetZod()
				internals.Check(payload)

				if tt.isValid {
					assert.Empty(t, payload.Issues, "Should accept string containing substring")
				} else {
					require.Len(t, payload.Issues, 1)
					if _, ok := tt.input.(string); ok {
						assert.Equal(t, string(InvalidFormat), payload.Issues[0].Code)
					} else {
						assert.Equal(t, string(InvalidType), payload.Issues[0].Code)
					}
				}
			})
		}
	})

	t.Run("Includes respects position parameter", func(t *testing.T) {
		position := 6
		check := NewZodCheckIncludes("world", &position)

		tests := []struct {
			name    string
			input   string
			isValid bool
		}{
			{"substring found at position", "hello world", true},
			{"substring not found at position", "world hello", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload := NewParsePayload(tt.input)
				internals := check.GetZod()
				internals.Check(payload)

				if tt.isValid {
					assert.Empty(t, payload.Issues, "Should find substring at correct position")
				} else {
					require.Len(t, payload.Issues, 1)
					assert.Equal(t, string(InvalidFormat), payload.Issues[0].Code)
				}
			})
		}
	})

	t.Run("StartsWith validates string prefix", func(t *testing.T) {
		check := NewZodCheckStartsWith("hello")

		tests := []struct {
			name    string
			input   interface{}
			isValid bool
		}{
			{"starts with prefix", "hello world", true},
			{"missing prefix", "world hello", false},
			{"non-string input", 123, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload := NewParsePayload(tt.input)
				internals := check.GetZod()
				internals.Check(payload)

				if tt.isValid {
					assert.Empty(t, payload.Issues, "Should accept string with correct prefix")
				} else {
					require.Len(t, payload.Issues, 1)
					if _, ok := tt.input.(string); ok {
						assert.Equal(t, string(InvalidFormat), payload.Issues[0].Code)
					} else {
						assert.Equal(t, string(InvalidType), payload.Issues[0].Code)
					}
				}
			})
		}
	})

	t.Run("EndsWith validates string suffix", func(t *testing.T) {
		check := NewZodCheckEndsWith("world")

		tests := []struct {
			name    string
			input   interface{}
			isValid bool
		}{
			{"ends with suffix", "hello world", true},
			{"missing suffix", "world hello", false},
			{"non-string input", 123, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload := NewParsePayload(tt.input)
				internals := check.GetZod()
				internals.Check(payload)

				if tt.isValid {
					assert.Empty(t, payload.Issues, "Should accept string with correct suffix")
				} else {
					require.Len(t, payload.Issues, 1)
					if _, ok := tt.input.(string); ok {
						assert.Equal(t, string(InvalidFormat), payload.Issues[0].Code)
					} else {
						assert.Equal(t, string(InvalidType), payload.Issues[0].Code)
					}
				}
			})
		}
	})
}

// =============================================================================
// SIZE VALIDATION TESTS
// =============================================================================

func TestSizeChecks_Collections(t *testing.T) {
	t.Run("MaxSize validates collection size constraints", func(t *testing.T) {
		check := NewZodCheckMaxSize(2)

		tests := []struct {
			name    string
			input   interface{}
			isValid bool
		}{
			{"small map", map[string]interface{}{"a": 1}, true},
			{"at limit", map[string]interface{}{"a": 1, "b": 2}, true},
			{"exceeds limit", map[string]interface{}{"a": 1, "b": 2, "c": 3}, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload := NewParsePayload(tt.input)
				internals := check.GetZod()
				internals.Check(payload)

				if tt.isValid {
					assert.Empty(t, payload.Issues, "Should accept collection within size limit")
				} else {
					require.Len(t, payload.Issues, 1)
					assert.Equal(t, string(TooBig), payload.Issues[0].Code)
				}
			})
		}
	})

	t.Run("MinSize validates minimum collection size", func(t *testing.T) {
		check := NewZodCheckMinSize(2)

		tests := []struct {
			name    string
			input   interface{}
			isValid bool
		}{
			{"above minimum", map[string]interface{}{"a": 1, "b": 2, "c": 3}, true},
			{"at minimum", map[string]interface{}{"a": 1, "b": 2}, true},
			{"below minimum", map[string]interface{}{"a": 1}, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload := NewParsePayload(tt.input)
				internals := check.GetZod()
				internals.Check(payload)

				if tt.isValid {
					assert.Empty(t, payload.Issues, "Should accept collection above minimum size")
				} else {
					require.Len(t, payload.Issues, 1)
					assert.Equal(t, string(TooSmall), payload.Issues[0].Code)
				}
			})
		}
	})

	t.Run("SizeEquals validates exact collection size", func(t *testing.T) {
		check := NewZodCheckSizeEquals(2)

		tests := []struct {
			name    string
			input   interface{}
			isValid bool
			expCode string
		}{
			{"exact size", map[string]interface{}{"a": 1, "b": 2}, true, ""},
			{"too large", map[string]interface{}{"a": 1, "b": 2, "c": 3}, false, string(TooBig)},
			{"too small", map[string]interface{}{"a": 1}, false, string(TooSmall)},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload := NewParsePayload(tt.input)
				internals := check.GetZod()
				internals.Check(payload)

				if tt.isValid {
					assert.Empty(t, payload.Issues, "Should accept exact size")
				} else {
					require.Len(t, payload.Issues, 1)
					assert.Equal(t, tt.expCode, payload.Issues[0].Code)
				}
			})
		}
	})
}

// =============================================================================
// TRANSFORM AND CUSTOM VALIDATION TESTS
// =============================================================================

func TestTransformChecks_Overwrite(t *testing.T) {
	t.Run("transforms values using provided function", func(t *testing.T) {
		transform := func(value interface{}) interface{} {
			if str, ok := value.(string); ok {
				return str + "_transformed"
			}
			return value
		}

		check := NewZodCheckOverwrite(transform)
		payload := NewParsePayload("hello")

		internals := check.GetZod()
		internals.Check(payload)

		assert.Equal(t, "hello_transformed", payload.Value, "Should transform value")
		assert.Empty(t, payload.Issues, "Should not have issues for transform")
	})

	t.Run("preserves non-matching types", func(t *testing.T) {
		transform := func(value interface{}) interface{} {
			if str, ok := value.(string); ok {
				return str + "_transformed"
			}
			return value
		}

		check := NewZodCheckOverwrite(transform)
		payload := NewParsePayload(42)

		internals := check.GetZod()
		internals.Check(payload)

		assert.Equal(t, 42, payload.Value, "Should preserve non-string values")
		assert.Empty(t, payload.Issues, "Should not have issues")
	})
}

func TestPropertyChecks_ObjectValidation(t *testing.T) {
	t.Run("validates when property exists", func(t *testing.T) {
		check := NewZodCheckProperty("name", nil)
		payload := NewParsePayload(map[string]interface{}{
			"name": "John",
			"age":  30,
		})

		internals := check.GetZod()
		internals.Check(payload)

		assert.Empty(t, payload.Issues, "Should not have issues when property exists")
	})

	t.Run("fails when property is missing", func(t *testing.T) {
		check := NewZodCheckProperty("email", nil)
		payload := NewParsePayload(map[string]interface{}{
			"name": "John",
			"age":  30,
		})

		internals := check.GetZod()
		internals.Check(payload)

		require.Len(t, payload.Issues, 1)
		assert.Equal(t, string(InvalidType), payload.Issues[0].Code)
	})

	t.Run("fails for non-object input", func(t *testing.T) {
		check := NewZodCheckProperty("name", nil)
		payload := NewParsePayload("not an object")

		internals := check.GetZod()
		internals.Check(payload)

		require.Len(t, payload.Issues, 1)
		assert.Equal(t, string(InvalidType), payload.Issues[0].Code)
	})
}

// =============================================================================
// UTILITY FUNCTION TESTS
// =============================================================================

func TestUtilityFunctions_TypeDetection(t *testing.T) {
	t.Run("getNumericOrigin detects numeric types correctly", func(t *testing.T) {
		tests := []struct {
			name     string
			input    interface{}
			expected string
		}{
			{"integer", 42, "number"},
			{"float", 3.14, "number"},
			{"int32", int32(100), "number"},
			{"int64", int64(1000), "number"},
			{"uint", uint(50), "number"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := getNumericOrigin(tt.input)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("compareNumeric compares values correctly", func(t *testing.T) {
		tests := []struct {
			name     string
			a, b     interface{}
			expected int
		}{
			{"less than", 5, 10, -1},
			{"equal", 5, 5, 0},
			{"greater than", 10, 5, 1},
			{"float comparison", 3.14, 3.15, -1},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := compareNumeric(tt.a, tt.b)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

func TestUtilityFunctions_LengthDetection(t *testing.T) {
	t.Run("hasLength detects length property correctly", func(t *testing.T) {
		tests := []struct {
			name     string
			input    interface{}
			expected bool
		}{
			{"string", "hello", true},
			{"slice", []int{1, 2, 3}, true},
			{"array", [3]int{1, 2, 3}, true},
			{"nil", nil, false},
			{"number", 42, false},
			{"map", map[string]int{"key": 1}, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := hasLength(tt.input)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("getLength returns correct length", func(t *testing.T) {
		tests := []struct {
			name     string
			input    interface{}
			expected int
		}{
			{"string", "hello", 5},
			{"slice", []int{1, 2, 3}, 3},
			{"array", [4]string{"a", "b", "c", "d"}, 4},
			{"empty string", "", 0},
			{"empty slice", []int{}, 0},
			{"nil", nil, 0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := getLength(tt.input)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

func TestUtilityFunctions_SizeDetection(t *testing.T) {
	t.Run("hasSize detects size property correctly", func(t *testing.T) {
		tests := []struct {
			name     string
			input    interface{}
			expected bool
		}{
			{"map", map[string]interface{}{"a": 1}, true},
			{"slice", []int{1, 2, 3}, true},
			{"array", [3]int{1, 2, 3}, true},
			{"string", "hello", false},
			{"nil", nil, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := hasSize(tt.input)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("getSize returns correct size", func(t *testing.T) {
		tests := []struct {
			name     string
			input    interface{}
			expected int
		}{
			{"map with items", map[string]interface{}{"a": 1, "b": 2}, 2},
			{"empty map", map[string]interface{}{}, 0},
			{"slice with items", []int{1, 2, 3}, 3},
			{"empty slice", []int{}, 0},
			{"array", [4]string{"a", "b", "c", "d"}, 4},
			{"non-sizeable", "hello", 0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := getSize(tt.input)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

// =============================================================================
// PARAMETER APPLICATION TESTS
// =============================================================================

func TestParameterApplication_SchemaParams(t *testing.T) {
	t.Run("applies error parameter correctly", func(t *testing.T) {
		def := &ZodCheckDef{Check: "test_check"}

		customErrorMap := ZodErrorMap(func(ZodRawIssue) string {
			return "Custom error message"
		})

		params := SchemaParams{Error: &customErrorMap}
		ApplySchemaParams(def, params)

		require.NotNil(t, def.Error)
		assert.False(t, def.Abort)
	})

	t.Run("applies abort parameter correctly", func(t *testing.T) {
		def := &ZodCheckDef{Check: "test_check"}

		params := SchemaParams{Abort: true}
		ApplySchemaParams(def, params)

		assert.Nil(t, def.Error)
		assert.True(t, def.Abort)
	})

	t.Run("applies multiple parameters correctly", func(t *testing.T) {
		def := &ZodCheckDef{Check: "test_check"}

		customErrorMap := ZodErrorMap(func(ZodRawIssue) string {
			return "Multi-param error"
		})

		params := SchemaParams{
			Error: &customErrorMap,
			Abort: true,
		}
		ApplySchemaParams(def, params)

		require.NotNil(t, def.Error)
		assert.True(t, def.Abort)
	})

	t.Run("handles empty parameters gracefully", func(t *testing.T) {
		def := &ZodCheckDef{Check: "test_check"}
		originalError := def.Error
		originalAbort := def.Abort

		ApplySchemaParams(def)

		assert.Equal(t, originalError, def.Error)
		assert.Equal(t, originalAbort, def.Abort)
	})
}

// =============================================================================
// ERROR HANDLING AND EDGE CASES
// =============================================================================

func TestErrorHandling_IssueGeneration(t *testing.T) {
	t.Run("generates correct error codes for different checks", func(t *testing.T) {
		testCases := []struct {
			name     string
			check    ZodCheck
			input    interface{}
			expected string
		}{
			{
				name:     "LessThan generates TooBig",
				check:    NewZodCheckLessThan(5, false),
				input:    10,
				expected: string(TooBig),
			},
			{
				name:     "GreaterThan generates TooSmall",
				check:    NewZodCheckGreaterThan(10, false),
				input:    5,
				expected: string(TooSmall),
			},
			{
				name:     "MultipleOf generates NotMultipleOf",
				check:    NewZodCheckMultipleOf(3),
				input:    10,
				expected: string(NotMultipleOf),
			},
			{
				name:     "StringFormat generates InvalidFormat",
				check:    NewZodCheckStringFormat(StringFormatLowercase, regexp.MustCompile(`^[a-z]+$`)),
				input:    "Hello",
				expected: string(InvalidFormat),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				payload := NewParsePayload(tc.input)
				internals := tc.check.GetZod()
				internals.Check(payload)

				require.Len(t, payload.Issues, 1)
				assert.Equal(t, tc.expected, payload.Issues[0].Code)
			})
		}
	})
}

func TestEdgeCases_BoundaryValues(t *testing.T) {
	t.Run("handles numeric boundary values correctly", func(t *testing.T) {
		// Test integer overflow scenarios
		check := NewZodCheckLessThan(1000, false)

		tests := []struct {
			name  string
			input interface{}
			valid bool
		}{
			{"exactly at boundary", 999, true},
			{"one above boundary", 1000, false},
			{"far above boundary", 2000, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload := NewParsePayload(tt.input)
				internals := check.GetZod()
				internals.Check(payload)

				if tt.valid {
					assert.Empty(t, payload.Issues)
				} else {
					assert.NotEmpty(t, payload.Issues)
				}
			})
		}
	})

	t.Run("handles string boundary cases correctly", func(t *testing.T) {
		check := NewZodCheckMaxLength(0)

		tests := []struct {
			name  string
			input string
			valid bool
		}{
			{"empty string", "", true},
			{"single character", "a", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				payload := NewParsePayload(tt.input)
				internals := check.GetZod()
				internals.Check(payload)

				if tt.valid {
					assert.Empty(t, payload.Issues)
				} else {
					assert.NotEmpty(t, payload.Issues)
				}
			})
		}
	})

	t.Run("handles position edge cases in includes check", func(t *testing.T) {
		tests := []struct {
			name      string
			position  *int
			input     string
			substring string
			valid     bool
		}{
			{"position at start", func() *int { p := 0; return &p }(), "hello world", "hello", true},
			{"position beyond string", func() *int { p := 20; return &p }(), "hello", "world", false},
			{"negative position", func() *int { p := -1; return &p }(), "hello world", "hello", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				check := NewZodCheckIncludes(tt.substring, tt.position)
				payload := NewParsePayload(tt.input)
				internals := check.GetZod()
				internals.Check(payload)

				if tt.valid {
					assert.Empty(t, payload.Issues)
				} else {
					assert.NotEmpty(t, payload.Issues)
				}
			})
		}
	})
}

// =============================================================================
// INTERFACE COMPLIANCE TESTS
// =============================================================================

func TestInterfaceCompliance_ZodCheck(t *testing.T) {
	t.Run("all checks implement ZodCheck interface", func(t *testing.T) {
		checks := []ZodCheck{
			NewZodCheckLessThan(10, false),
			NewZodCheckGreaterThan(5, false),
			NewZodCheckMultipleOf(3),
			NewZodCheckNumberFormat(NumberFormatInt32),
			NewZodCheckMaxLength(10),
			NewZodCheckMinLength(1),
			NewZodCheckStringFormat(StringFormatLowercase, regexp.MustCompile(`^[a-z]+$`)),
			NewZodCheckRegex(regexp.MustCompile(`^\d+$`)),
			NewZodCheckIncludes("test", nil),
			NewZodCheckStartsWith("hello"),
			NewZodCheckEndsWith("world"),
			NewZodCheckOverwrite(func(v interface{}) interface{} { return v }),
			NewZodCheckLowerCase(),
			NewZodCheckUpperCase(),
			NewZodCheckLengthEquals(5),
			NewZodCheckMaxSize(10),
			NewZodCheckMinSize(1),
			NewZodCheckSizeEquals(5),
			NewZodCheckProperty("test", nil),
		}

		for i, check := range checks {
			t.Run(fmt.Sprintf("check_%d", i), func(t *testing.T) {
				internals := check.GetZod()
				require.NotNil(t, internals, "Check should return valid internals")
				require.NotNil(t, internals.Def, "Check should have valid definition")
				require.NotNil(t, internals.Check, "Check should have valid check function")
				assert.NotEmpty(t, internals.Def.Check, "Check should have non-empty check type")
			})
		}
	})
}

// =============================================================================
// COMPREHENSIVE INTEGRATION TESTS
// =============================================================================

func TestIntegration_CheckChaining(t *testing.T) {
	t.Run("validates complex validation scenarios", func(t *testing.T) {
		// This would typically be tested at the schema level,
		// but we can test check creation and basic functionality
		checks := []ZodCheck{
			NewZodCheckMinLength(3),
			NewZodCheckMaxLength(10),
			NewZodCheckStartsWith("test"),
		}

		input := "test_value"
		for _, check := range checks {
			payload := NewParsePayload(input)
			internals := check.GetZod()
			internals.Check(payload)

			// All checks should pass for this input
			assert.Empty(t, payload.Issues, "All checks should pass for valid input")
		}
	})
}

// =============================================================================
// FILE TYPE VALIDATION TESTS (TODO)
// =============================================================================

func TestFileChecks_MimeType(t *testing.T) {
	t.Run("TODO: MIME type validation needs File interface implementation", func(t *testing.T) {
		t.Skip("MIME type validation requires proper File interface - TODO for future implementation")

		// TODO: Implement once File interface is properly defined
		// This test should validate:
		// - Correct MIME type detection from file headers
		// - Validation against allowed MIME types list
		// - Error handling for invalid file types
		// - Support for multipart.FileHeader and os.File
	})
}

// =============================================================================
// CUSTOM VALIDATION TESTS (TODO)
// =============================================================================

func TestCustomValidation_RefineFunctions(t *testing.T) {
	t.Run("TODO: Custom refine function validation", func(t *testing.T) {
		t.Skip("Custom validation with refine functions - TODO for future implementation")

		// TODO: Implement comprehensive custom validation tests
		// This should test:
		// - RefineFn[T] with different type parameters
		// - Custom error messages and paths
		// - Integration with schema-level validation
		// - Performance optimization for frequently used custom checks
	})
}

// =============================================================================
// BENCHMARK TESTS
// =============================================================================

func BenchmarkCheck_NumericValidation(b *testing.B) {
	check := NewZodCheckLessThan(1000, false)
	internals := check.GetZod()
	input := 500

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		payload := NewParsePayload(input)
		internals.Check(payload)
		if len(payload.Issues) > 0 {
			b.Fatal("Unexpected validation failure")
		}
	}
}

func BenchmarkCheck_StringValidation(b *testing.B) {
	pattern := regexp.MustCompile(`^[a-z]+$`)
	check := NewZodCheckStringFormat(StringFormatLowercase, pattern)
	internals := check.GetZod()
	input := "hello"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		payload := NewParsePayload(input)
		internals.Check(payload)
		if len(payload.Issues) > 0 {
			b.Fatal("Unexpected validation failure")
		}
	}
}

func BenchmarkCheck_RegexValidation(b *testing.B) {
	pattern := regexp.MustCompile(`^\d{3}-\d{3}-\d{4}$`)
	check := NewZodCheckRegex(pattern)
	internals := check.GetZod()
	input := "123-456-7890"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		payload := NewParsePayload(input)
		internals.Check(payload)
		if len(payload.Issues) > 0 {
			b.Fatal("Unexpected validation failure")
		}
	}
}

func BenchmarkCheck_TransformValidation(b *testing.B) {
	transform := func(value interface{}) interface{} {
		if str, ok := value.(string); ok {
			return str + "_suffix"
		}
		return value
	}
	check := NewZodCheckOverwrite(transform)
	internals := check.GetZod()
	input := "test"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		payload := NewParsePayload(input)
		internals.Check(payload)
		// Transform should never generate issues
		if len(payload.Issues) > 0 {
			b.Fatal("Unexpected transform failure")
		}
	}
}

func BenchmarkCheck_ValidationFailure(b *testing.B) {
	check := NewZodCheckLessThan(10, false)
	internals := check.GetZod()
	input := 20 // This will fail validation

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		payload := NewParsePayload(input)
		internals.Check(payload)
		// Should generate exactly one issue
		if len(payload.Issues) != 1 {
			b.Fatal("Expected exactly one validation issue")
		}
	}
}
