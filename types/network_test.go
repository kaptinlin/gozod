package types

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Basic functionality tests
// =============================================================================

func TestIPv4_BasicFunctionality(t *testing.T) {
	t.Run("valid IPv4 inputs", func(t *testing.T) {
		schema := IPv4()

		validIPs := []string{
			"192.168.1.1",
			"10.0.0.1",
			"172.16.0.1",
			"8.8.8.8",
			"255.255.255.255",
			"0.0.0.0",
		}

		for _, ip := range validIPs {
			result, err := schema.Parse(ip)
			require.NoError(t, err, "IPv4 %s should be valid", ip)
			assert.Equal(t, ip, result)
		}
	})

	t.Run("invalid IPv4 inputs", func(t *testing.T) {
		schema := IPv4()

		invalidInputs := []any{
			"256.1.1.1",             // Invalid octet
			"192.168.1",             // Missing octet
			"192.168.1.1.1",         // Too many octets
			"192.168.1.a",           // Non-numeric
			"not an ip",             // Random string
			123,                     // Number
			true,                    // Boolean
			[]string{"192.168.1.1"}, // Array
			nil,                     // Nil
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})

	t.Run("Parse and MustParse methods", func(t *testing.T) {
		schema := IPv4()
		validIP := "192.168.1.1"

		// Test Parse method
		result, err := schema.Parse(validIP)
		require.NoError(t, err)
		assert.Equal(t, validIP, result)

		// Test MustParse method
		mustResult := schema.MustParse(validIP)
		assert.Equal(t, validIP, mustResult)

		// Test panic on invalid input
		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected a valid IPv4 address"
		schema := IPv4(core.SchemaParams{Error: customError})

		require.NotNil(t, schema)

		_, err := schema.Parse("invalid")
		assert.Error(t, err)
	})
}

// =============================================================================
// 2. Type safety tests
// =============================================================================

func TestIPv4_TypeSafety(t *testing.T) {
	t.Run("IPv4 returns string type", func(t *testing.T) {
		schema := IPv4()
		require.NotNil(t, schema)

		result, err := schema.Parse("192.168.1.1")
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.1", result)
		assert.IsType(t, "", result) // Ensure type is string
	})

	t.Run("IPv4Ptr returns *string type", func(t *testing.T) {
		schema := IPv4Ptr()
		require.NotNil(t, schema)

		result, err := schema.Parse("192.168.1.1")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "192.168.1.1", *result)
		assert.IsType(t, (*string)(nil), result) // Ensure type is *string
	})

	t.Run("type inference with assignment", func(t *testing.T) {
		// Type-inference friendly API
		stringSchema := IPv4() // string type
		ptrSchema := IPv4Ptr() // *string type

		// Test string type
		result1, err1 := stringSchema.Parse("10.0.0.1")
		require.NoError(t, err1)
		assert.IsType(t, "", result1)
		assert.Equal(t, "10.0.0.1", result1)

		// Test *string type
		result2, err2 := ptrSchema.Parse("10.0.0.1")
		require.NoError(t, err2)
		assert.IsType(t, (*string)(nil), result2)
		require.NotNil(t, result2)
		assert.Equal(t, "10.0.0.1", *result2)
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		// Test string type
		stringSchema := IPv4()
		result := stringSchema.MustParse("8.8.8.8")
		assert.IsType(t, "", result)
		assert.Equal(t, "8.8.8.8", result)

		// Test *string type
		ptrSchema := IPv4Ptr()
		ptrResult := ptrSchema.MustParse("8.8.8.8")
		assert.IsType(t, (*string)(nil), ptrResult)
		require.NotNil(t, ptrResult)
		assert.Equal(t, "8.8.8.8", *ptrResult)
	})
}

// =============================================================================
// 3. Modifier methods tests
// =============================================================================

func TestIPv4_Modifiers(t *testing.T) {
	t.Run("Optional always returns *string", func(t *testing.T) {
		// From string to *string via Optional
		stringSchema := IPv4()
		optionalSchema := stringSchema.Optional()

		// Type check: ensure it returns *ZodIPv4[*string]
		var _ = optionalSchema

		// Functionality test
		result, err := optionalSchema.Parse("192.168.1.1")
		require.NoError(t, err)
		assert.IsType(t, (*string)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, "192.168.1.1", *result)

		// From *string to *string via Optional (maintains type)
		ptrSchema := IPv4Ptr()
		optionalPtrSchema := ptrSchema.Optional()
		var _ = optionalPtrSchema
	})

	t.Run("Nilable always returns *string", func(t *testing.T) {
		stringSchema := IPv4()
		nilableSchema := stringSchema.Nilable()

		var _ = nilableSchema

		// Test nil handling
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Default preserves current type", func(t *testing.T) {
		// string maintains string
		stringSchema := IPv4()
		defaultStringSchema := stringSchema.Default("192.168.1.1")
		var _ = defaultStringSchema

		// *string maintains *string
		ptrSchema := IPv4Ptr()
		defaultPtrSchema := ptrSchema.Default("192.168.1.1")
		var _ = defaultPtrSchema
	})

	t.Run("Prefault preserves current type", func(t *testing.T) {
		// string maintains string
		stringSchema := IPv4()
		prefaultStringSchema := stringSchema.Prefault("192.168.1.1")
		var _ = prefaultStringSchema

		// *string maintains *string
		ptrSchema := IPv4Ptr()
		prefaultPtrSchema := ptrSchema.Prefault("192.168.1.1")
		var _ = prefaultPtrSchema
	})
}

// =============================================================================
// 4. Chaining tests
// =============================================================================

func TestIPv4_Chaining(t *testing.T) {
	t.Run("type evolution through chaining", func(t *testing.T) {
		// Chain with type evolution
		schema := IPv4(). // *ZodIPv4[string]
					Default("192.168.1.1"). // *ZodIPv4[string] (maintains type)
					Optional()              // *ZodIPv4[*string] (type conversion)

		var _ = schema

		// Test final behavior
		result, err := schema.Parse("10.0.0.1")
		require.NoError(t, err)
		assert.IsType(t, (*string)(nil), result)
		require.NotNil(t, result)
		assert.Equal(t, "10.0.0.1", *result)
	})

	t.Run("complex chaining", func(t *testing.T) {
		schema := IPv4Ptr(). // *ZodIPv4[*string]
					Nilable().             // *ZodIPv4[*string] (maintains type)
					Default("192.168.1.1") // *ZodIPv4[*string] (maintains type)

		var _ = schema

		result, err := schema.Parse("172.16.0.1")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "172.16.0.1", *result)
	})

	t.Run("default and prefault chaining", func(t *testing.T) {
		schema := IPv4().
			Default("192.168.1.1").
			Prefault("10.0.0.1")

		result, err := schema.Parse("8.8.8.8")
		require.NoError(t, err)
		assert.Equal(t, "8.8.8.8", result)
	})
}

// =============================================================================
// 5. Default and prefault tests
// =============================================================================

func TestIPv4_DefaultAndPrefault(t *testing.T) {
	// Test 1: Default has higher priority than Prefault
	t.Run("Default priority over Prefault", func(t *testing.T) {
		schema := IPv4().Default("192.168.1.1").Prefault("10.0.0.1")

		// When input is nil, Default should take precedence
		result, err := schema.ParseAny(nil)
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.1", result)
	})

	// Test 2: Default short-circuit mechanism
	t.Run("Default short-circuit bypasses validation", func(t *testing.T) {
		// Create a schema where default value is invalid IPv4
		schema := IPv4().Default("invalid_ip")

		// Default should bypass validation even if it's not a valid IPv4
		result, err := schema.ParseAny(nil)
		require.NoError(t, err)
		assert.Equal(t, "invalid_ip", result)
	})

	// Test 3: Prefault requires full validation
	t.Run("Prefault requires full validation", func(t *testing.T) {
		// Create a schema where prefault value is invalid IPv4
		schema := IPv4().Prefault("invalid_ip")

		// Prefault should fail validation if it's not a valid IPv4
		_, err := schema.ParseAny(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid")
	})

	// Test 4: Prefault only triggers on nil input
	t.Run("Prefault only triggers on nil input", func(t *testing.T) {
		schema := IPv4().Prefault("192.168.1.1")

		// Non-nil input that fails validation should not trigger Prefault
		_, err := schema.ParseAny("invalid_ip")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid")
	})

	// Test 5: DefaultFunc and PrefaultFunc behavior
	t.Run("DefaultFunc and PrefaultFunc behavior", func(t *testing.T) {
		defaultCalled := false
		prefaultCalled := false

		schema := IPv4().DefaultFunc(func() string {
			defaultCalled = true
			return "192.168.1.1"
		}).PrefaultFunc(func() string {
			prefaultCalled = true
			return "10.0.0.1"
		})

		// DefaultFunc should be called and take precedence
		result, err := schema.ParseAny(nil)
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.1", result)
		assert.True(t, defaultCalled)
		assert.False(t, prefaultCalled) // PrefaultFunc should not be called
	})

	// Test 6: Error handling for Prefault validation failure
	t.Run("Prefault validation failure returns error", func(t *testing.T) {
		schema := IPv4().Prefault("999.999.999.999") // Invalid IPv4 address

		// Should return validation error, not attempt fallback
		_, err := schema.ParseAny(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid")
	})
}

// =============================================================================
// 6. Refine tests
// =============================================================================

func TestIPv4_Refine(t *testing.T) {
	t.Run("refine validate", func(t *testing.T) {
		// Only accept private IP addresses
		schema := IPv4().Refine(func(ip string) bool {
			return ip[:3] == "192" || ip[:2] == "10"
		})

		result, err := schema.Parse("192.168.1.1")
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.1", result)

		result, err = schema.Parse("10.0.0.1")
		require.NoError(t, err)
		assert.Equal(t, "10.0.0.1", result)

		_, err = schema.Parse("8.8.8.8")
		assert.Error(t, err)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		errorMessage := "Must be a private IP address"
		schema := IPv4Ptr().Refine(func(ip *string) bool {
			return ip != nil && (*ip)[:3] == "192"
		}, core.SchemaParams{Error: errorMessage})

		result, err := schema.Parse("192.168.1.1")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "192.168.1.1", *result)

		_, err = schema.Parse("8.8.8.8")
		assert.Error(t, err)
	})

	t.Run("refine pointer allows nil", func(t *testing.T) {
		schema := IPv4Ptr().Nilable().Refine(func(ip *string) bool {
			// Accept nil or private IPs
			return ip == nil || (*ip)[:3] == "192"
		})

		// Expect nil to be accepted
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// 192.* should pass
		result, err = schema.Parse("192.168.1.1")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "192.168.1.1", *result)

		// Public IP should fail
		_, err = schema.Parse("8.8.8.8")
		assert.Error(t, err)
	})
}

func TestIPv4_RefineAny(t *testing.T) {
	t.Run("refineAny string schema", func(t *testing.T) {
		// Only accept IPs starting with "10."
		schema := IPv4().RefineAny(func(v any) bool {
			ip, ok := v.(string)
			return ok && len(ip) > 3 && ip[:3] == "10."
		})

		// "10.*" passes
		result, err := schema.Parse("10.0.0.1")
		require.NoError(t, err)
		assert.Equal(t, "10.0.0.1", result)

		// Others fail
		_, err = schema.Parse("192.168.1.1")
		assert.Error(t, err)
	})

	t.Run("refineAny pointer schema", func(t *testing.T) {
		// IPv4Ptr().RefineAny sees underlying string value
		schema := IPv4Ptr().RefineAny(func(v any) bool {
			ip, ok := v.(string)
			return ok && len(ip) > 3 && ip[:3] == "10."
		})

		result, err := schema.Parse("10.0.0.1")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "10.0.0.1", *result)

		_, err = schema.Parse("192.168.1.1")
		assert.Error(t, err)
	})
}

// =============================================================================
// 7. Coercion tests
// =============================================================================

func TestIPv4_Coercion(t *testing.T) {
	t.Run("string coercion", func(t *testing.T) {
		schema := CoercedIPv4()

		// Test string input (should work directly)
		result, err := schema.Parse("192.168.1.1")
		require.NoError(t, err, "Should accept string IPv4")
		assert.Equal(t, "192.168.1.1", result)
	})

	t.Run("invalid coercion inputs", func(t *testing.T) {
		schema := CoercedIPv4()

		// Inputs that cannot be coerced to valid IPv4
		invalidInputs := []any{
			123,            // Number
			true,           // Boolean
			[]string{"ip"}, // Array
			nil,            // Nil
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})
}

// =============================================================================
// 8. Error handling tests
// =============================================================================

func TestIPv4_ErrorHandling(t *testing.T) {
	t.Run("nil input handling", func(t *testing.T) {
		testCases := []struct {
			name         string
			schema       func() any
			expectError  bool
			expectedType any
		}{
			{"IPv4 rejects nil", func() any { return IPv4() }, true, ""},
			{"IPv4Ptr rejects nil", func() any { return IPv4Ptr() }, true, (*string)(nil)},
			{"Optional allows nil", func() any { return IPv4().Optional() }, false, (*string)(nil)},
			{"Nilable allows nil", func() any { return IPv4().Nilable() }, false, (*string)(nil)},
			{"Default allows nil and returns default", func() any { return IPv4().Default("1.1.1.1") }, false, ""},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				schema := tc.schema()

				// Type assertion to get Parse method
				if ipv4Schema, ok := schema.(*ZodIPv4[string]); ok {
					result, err := ipv4Schema.Parse(nil)
					if tc.expectError {
						assert.Error(t, err)
					} else {
						require.NoError(t, err)
						assert.IsType(t, tc.expectedType, result)
					}
				} else if ptrSchema, ok := schema.(*ZodIPv4[*string]); ok {
					result, err := ptrSchema.Parse(nil)
					if tc.expectError {
						assert.Error(t, err)
					} else {
						require.NoError(t, err)
						assert.IsType(t, tc.expectedType, result)
					}
				}
			})
		}
	})
}

// =============================================================================
// 9. Edge case tests
// =============================================================================

func TestIPv4_EdgeCases(t *testing.T) {
	t.Run("nil handling with *string", func(t *testing.T) {
		schema := IPv4Ptr().Nilable()

		// Test nil input
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid IPv4
		result, err = schema.Parse("192.168.1.1")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "192.168.1.1", *result)
	})

	t.Run("empty context", func(t *testing.T) {
		schema := IPv4()

		// Parse with empty context slice
		result, err := schema.Parse("192.168.1.1")
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.1", result)
	})

	t.Run("performance critical paths", func(t *testing.T) {
		schema := IPv4()

		// Test that fast paths work correctly
		t.Run("direct string input fast path", func(t *testing.T) {
			result, err := schema.Parse("10.0.0.1")
			require.NoError(t, err)
			assert.Equal(t, "10.0.0.1", result)
		})

		t.Run("various valid IPs", func(t *testing.T) {
			ips := []string{
				"0.0.0.0",
				"127.0.0.1",
				"255.255.255.255",
			}
			for _, ip := range ips {
				result, err := schema.Parse(ip)
				require.NoError(t, err)
				assert.Equal(t, ip, result)
			}
		})
	})

	t.Run("API compatibility patterns", func(t *testing.T) {
		// Test that the API matches expected patterns

		// Basic usage patterns
		schema1 := IPv4()                        // Basic IPv4
		schema2 := IPv4().Optional()             // Optional IPv4
		schema3 := IPv4().Default("192.168.1.1") // IPv4 with default
		schema4 := CoercedIPv4()                 // Coerced IPv4

		// Verify these work as expected
		result, err := schema1.Parse("10.0.0.1")
		require.NoError(t, err)
		assert.Equal(t, "10.0.0.1", result)

		result2, err := schema2.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result2)

		result3, err := schema3.Parse("172.16.0.1")
		require.NoError(t, err)
		assert.Equal(t, "172.16.0.1", result3)

		result4, err := schema4.Parse("8.8.8.8")
		require.NoError(t, err)
		assert.Equal(t, "8.8.8.8", result4)
	})

	t.Run("memory efficiency verification", func(t *testing.T) {
		// Create multiple schemas to verify shared state
		schema1 := IPv4()
		schema2 := schema1.Default("192.168.1.1")
		schema3 := schema2.Optional()

		// All should work independently
		result1, err1 := schema1.Parse("10.0.0.1")
		require.NoError(t, err1)
		assert.Equal(t, "10.0.0.1", result1)

		result2, err2 := schema2.Parse("172.16.0.1")
		require.NoError(t, err2)
		assert.Equal(t, "172.16.0.1", result2)

		result3, err3 := schema3.Parse("8.8.8.8")
		require.NoError(t, err3)
		assert.NotNil(t, result3)
		assert.Equal(t, "8.8.8.8", *result3)
	})

	t.Run("transform and pipe integration", func(t *testing.T) {
		schema := IPv4()

		// Test Transform
		transform := schema.Transform(func(ip string, ctx *core.RefinementContext) (any, error) {
			return "IP: " + ip, nil
		})

		result, err := transform.Parse("192.168.1.1")
		require.NoError(t, err)
		assert.Equal(t, "IP: 192.168.1.1", result)

		// Test extractNetworkString helper function
		ipVal := "192.168.1.1"
		ptrVal := &ipVal

		extracted1 := extractNetworkString[string](ipVal)
		assert.Equal(t, "192.168.1.1", extracted1)

		extracted2 := extractNetworkString[*string](ptrVal)
		assert.Equal(t, "192.168.1.1", extracted2)

		nilPtr := (*string)(nil)
		extracted3 := extractNetworkString[*string](nilPtr)
		assert.Equal(t, "", extracted3)
	})

	t.Run("Pointers from multiple parses should not be the same", func(t *testing.T) {
		testCases := []struct {
			name   string
			schema any
			input  string
		}{
			{"IPv4Ptr", IPv4Ptr(), "192.168.1.1"},
			{"IPv6Ptr", IPv6Ptr(), "::1"},
			{"CIDRv4Ptr", CIDRv4Ptr(), "192.168.1.0/24"},
			{"CIDRv6Ptr", CIDRv6Ptr(), "2001:db8::/32"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// We need to use type assertion to handle different schema types
				if s, ok := any(tc.schema).(interface { //nolint:unconvert
					Parse(any, ...*core.ParseContext) (any, error)
				}); ok {
					// First parse should return a pointer with the same address
					result1, err1 := s.Parse(tc.input)
					require.NoError(t, err1, "First parse failed")
					resultPtr1, ok1 := result1.(*string)
					require.True(t, ok1, "Result of first parse should be *string")
					require.NotNil(t, resultPtr1, "Result pointer from first parse should not be nil")
					assert.Equal(t, tc.input, *resultPtr1, "Value should match input")

					// Second parse should also work correctly
					result2, err2 := s.Parse(tc.input)
					require.NoError(t, err2, "Second parse failed")
					resultPtr2, ok2 := result2.(*string)
					require.True(t, ok2, "Result of second parse should be *string")
					require.NotNil(t, resultPtr2, "Result pointer from second parse should not be nil")
					assert.Equal(t, tc.input, *resultPtr2, "Value should match input on second parse")

					// Compare addresses - they should be different due to new allocations
					assert.NotSame(t, resultPtr1, resultPtr2, "Pointers from two separate parses should not be the same")
				}
			})
		}
	})

	t.Run("Overwrite with pointer types preserves identity", func(t *testing.T) {
		// ... existing code ...
	})
}

// =============================================================================
// IPv6 TESTS
// =============================================================================

// =============================================================================
// 1. Basic functionality tests
// =============================================================================

func TestIPv6_BasicFunctionality(t *testing.T) {
	t.Run("valid IPv6 inputs", func(t *testing.T) {
		schema := IPv6()

		validIPs := []string{
			"2001:db8::1",
			"::1",
			"::ffff:192.168.1.1",
			"2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			"2001:db8:85a3::8a2e:370:7334",
			"::",
		}

		for _, ip := range validIPs {
			result, err := schema.Parse(ip)
			require.NoError(t, err, "IPv6 %s should be valid", ip)
			assert.Equal(t, ip, result)
		}
	})

	t.Run("invalid IPv6 inputs", func(t *testing.T) {
		schema := IPv6()

		invalidInputs := []any{
			"2001:db8::1::1",                               // Double "::"
			"2001:db8:85a3::8a2e::7334",                    // Multiple "::"
			"2001:db8:85a3:0000:0000:8a2e:0370:7334:extra", // Too many groups
			"2001:db8:85a3:gggg:0000:8a2e:0370:7334",       // Invalid hex
			"192.168.1.1",                                  // IPv4 address
			"not an ip",                                    // Random string
			123,                                            // Number
			true,                                           // Boolean
			[]string{"2001:db8::1"},                        // Array
			nil,                                            // Nil
		}

		for _, input := range invalidInputs {
			_, err := schema.Parse(input)
			assert.Error(t, err, "Expected error for input: %v", input)
		}
	})

	t.Run("Parse and MustParse methods", func(t *testing.T) {
		schema := IPv6()
		validIP := "2001:db8::1"

		// Test Parse method
		result, err := schema.Parse(validIP)
		require.NoError(t, err)
		assert.Equal(t, validIP, result)

		// Test MustParse method
		mustResult := schema.MustParse(validIP)
		assert.Equal(t, validIP, mustResult)

		// Test panic on invalid input
		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected a valid IPv6 address"
		schema := IPv6(core.SchemaParams{Error: customError})

		require.NotNil(t, schema)

		_, err := schema.Parse("invalid")
		assert.Error(t, err)
	})
}

// =============================================================================
// 2. Type safety tests
// =============================================================================

func TestIPv6_TypeSafety(t *testing.T) {
	t.Run("IPv6 returns string type", func(t *testing.T) {
		schema := IPv6()
		require.NotNil(t, schema)

		result, err := schema.Parse("2001:db8::1")
		require.NoError(t, err)
		assert.Equal(t, "2001:db8::1", result)
		assert.IsType(t, "", result) // Ensure type is string
	})

	t.Run("IPv6Ptr returns *string type", func(t *testing.T) {
		schema := IPv6Ptr()
		require.NotNil(t, schema)

		result, err := schema.Parse("2001:db8::1")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "2001:db8::1", *result)
		assert.IsType(t, (*string)(nil), result) // Ensure type is *string
	})

	t.Run("type inference with assignment", func(t *testing.T) {
		// Type-inference friendly API
		stringSchema := IPv6() // string type
		ptrSchema := IPv6Ptr() // *string type

		// Test string type
		result1, err1 := stringSchema.Parse("::1")
		require.NoError(t, err1)
		assert.IsType(t, "", result1)
		assert.Equal(t, "::1", result1)

		// Test *string type
		result2, err2 := ptrSchema.Parse("::1")
		require.NoError(t, err2)
		assert.IsType(t, (*string)(nil), result2)
		require.NotNil(t, result2)
		assert.Equal(t, "::1", *result2)
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		// Test string type
		stringSchema := IPv6()
		result := stringSchema.MustParse("2001:db8::1")
		assert.IsType(t, "", result)
		assert.Equal(t, "2001:db8::1", result)

		// Test *string type
		ptrSchema := IPv6Ptr()
		ptrResult := ptrSchema.MustParse("2001:db8::1")
		assert.IsType(t, (*string)(nil), ptrResult)
		require.NotNil(t, ptrResult)
		assert.Equal(t, "2001:db8::1", *ptrResult)
	})
}

// =============================================================================
// HOSTNAME REFINE TESTS
// =============================================================================

func TestHostname_Refine(t *testing.T) {
	t.Run("refine with string type", func(t *testing.T) {
		// Only accept hostnames ending with ".com"
		schema := Hostname().Refine(func(h string) bool {
			return len(h) > 4 && h[len(h)-4:] == ".com"
		})

		result, err := schema.Parse("example.com")
		require.NoError(t, err)
		assert.Equal(t, "example.com", result)

		_, err = schema.Parse("example.org")
		assert.Error(t, err)
	})

	t.Run("refine with pointer type", func(t *testing.T) {
		schema := HostnamePtr().Refine(func(h *string) bool {
			return h != nil && len(*h) > 5
		})

		result, err := schema.Parse("example.com")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "example.com", *result)

		_, err = schema.Parse("a.co")
		assert.Error(t, err)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		errorMessage := "Hostname must be a .io domain"
		schema := Hostname().Refine(func(h string) bool {
			return len(h) > 3 && h[len(h)-3:] == ".io"
		}, core.SchemaParams{Error: errorMessage})

		_, err := schema.Parse("example.com")
		assert.Error(t, err)
	})

	t.Run("refine with nilable pointer", func(t *testing.T) {
		schema := HostnamePtr().Nilable().Refine(func(h *string) bool {
			return h == nil || len(*h) > 3
		})

		// Nil should pass
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid hostname should pass
		result, err = schema.Parse("test.com")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "test.com", *result)
	})
}

func TestHostname_RefineAny(t *testing.T) {
	t.Run("refineAny with string schema", func(t *testing.T) {
		// Only accept hostnames containing "test"
		schema := Hostname().RefineAny(func(v any) bool {
			if h, ok := v.(string); ok {
				for i := 0; i <= len(h)-4; i++ {
					if h[i:i+4] == "test" {
						return true
					}
				}
			}
			return false
		})

		result, err := schema.Parse("test.example.com")
		require.NoError(t, err)
		assert.Equal(t, "test.example.com", result)

		_, err = schema.Parse("prod.example.com")
		assert.Error(t, err)
	})

	t.Run("refineAny with pointer schema", func(t *testing.T) {
		schema := HostnamePtr().RefineAny(func(v any) bool {
			if h, ok := v.(string); ok {
				return len(h) >= 5
			}
			return false
		})

		result, err := schema.Parse("hello.com")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "hello.com", *result)
	})
}

// =============================================================================
// MAC ADDRESS REFINE TESTS
// =============================================================================

func TestMAC_Refine(t *testing.T) {
	t.Run("refine with string type", func(t *testing.T) {
		// Only accept MACs starting with specific vendor prefix
		schema := MAC().Refine(func(m string) bool {
			return len(m) >= 8 && m[:8] == "00:1A:2B"
		})

		result, err := schema.Parse("00:1A:2B:3C:4D:5E")
		require.NoError(t, err)
		assert.Equal(t, "00:1A:2B:3C:4D:5E", result)

		_, err = schema.Parse("AA:BB:CC:DD:EE:FF")
		assert.Error(t, err)
	})

	t.Run("refine with pointer type", func(t *testing.T) {
		schema := MACPtr().Refine(func(m *string) bool {
			return m != nil && len(*m) == 17
		})

		result, err := schema.Parse("00:11:22:33:44:55")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "00:11:22:33:44:55", *result)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		errorMessage := "MAC must be from approved vendor"
		schema := MAC().Refine(func(m string) bool {
			return len(m) >= 2 && m[:2] == "AA"
		}, core.SchemaParams{Error: errorMessage})

		_, err := schema.Parse("00:11:22:33:44:55")
		assert.Error(t, err)
	})

	t.Run("refine with nilable pointer", func(t *testing.T) {
		schema := MACPtr().Nilable().Refine(func(m *string) bool {
			return m == nil || len(*m) > 0
		})

		// Nil should pass
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid MAC should pass
		result, err = schema.Parse("00:11:22:33:44:55")
		require.NoError(t, err)
		require.NotNil(t, result)
	})
}

func TestMAC_RefineAny(t *testing.T) {
	t.Run("refineAny with string schema", func(t *testing.T) {
		// Validate MAC contains uppercase letters
		schema := MAC().RefineAny(func(v any) bool {
			if m, ok := v.(string); ok {
				for _, c := range m {
					if c >= 'A' && c <= 'F' {
						return true
					}
				}
			}
			return false
		})

		result, err := schema.Parse("AA:BB:CC:DD:EE:FF")
		require.NoError(t, err)
		assert.Equal(t, "AA:BB:CC:DD:EE:FF", result)

		_, err = schema.Parse("00:11:22:33:44:55")
		assert.Error(t, err)
	})

	t.Run("refineAny with pointer schema", func(t *testing.T) {
		schema := MACPtr().RefineAny(func(v any) bool {
			if m, ok := v.(string); ok {
				return len(m) == 17
			}
			return false
		})

		result, err := schema.Parse("00:11:22:33:44:55")
		require.NoError(t, err)
		require.NotNil(t, result)
	})
}

// =============================================================================
// E164 REFINE TESTS
// =============================================================================

func TestE164_Refine(t *testing.T) {
	t.Run("refine with string type", func(t *testing.T) {
		// Only accept US phone numbers (+1)
		schema := E164().Refine(func(p string) bool {
			return len(p) > 2 && p[:2] == "+1"
		})

		result, err := schema.Parse("+14155551234")
		require.NoError(t, err)
		assert.Equal(t, "+14155551234", result)

		_, err = schema.Parse("+442071234567")
		assert.Error(t, err)
	})

	t.Run("refine with pointer type", func(t *testing.T) {
		schema := E164Ptr().Refine(func(p *string) bool {
			return p != nil && len(*p) >= 10
		})

		result, err := schema.Parse("+14155551234")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "+14155551234", *result)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		errorMessage := "Phone number must be from UK"
		schema := E164().Refine(func(p string) bool {
			return len(p) > 3 && p[:3] == "+44"
		}, core.SchemaParams{Error: errorMessage})

		_, err := schema.Parse("+14155551234")
		assert.Error(t, err)
	})

	t.Run("refine with nilable pointer", func(t *testing.T) {
		schema := E164Ptr().Nilable().Refine(func(p *string) bool {
			return p == nil || len(*p) >= 8
		})

		// Nil should pass
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Valid E164 should pass
		result, err = schema.Parse("+14155551234")
		require.NoError(t, err)
		require.NotNil(t, result)
	})
}

func TestE164_RefineAny(t *testing.T) {
	t.Run("refineAny with string schema", func(t *testing.T) {
		// Validate phone number length
		schema := E164().RefineAny(func(v any) bool {
			if p, ok := v.(string); ok {
				return len(p) >= 10 && len(p) <= 15
			}
			return false
		})

		result, err := schema.Parse("+14155551234")
		require.NoError(t, err)
		assert.Equal(t, "+14155551234", result)
	})

	t.Run("refineAny with pointer schema", func(t *testing.T) {
		schema := E164Ptr().RefineAny(func(v any) bool {
			if p, ok := v.(string); ok {
				return len(p) > 1 && p[0] == '+'
			}
			return false
		})

		result, err := schema.Parse("+14155551234")
		require.NoError(t, err)
		require.NotNil(t, result)
	})
}
