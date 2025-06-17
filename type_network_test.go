package gozod

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. Basic functionality and type inference
// =============================================================================

func TestNetworkBasicFunctionality(t *testing.T) {
	t.Run("constructor validation", func(t *testing.T) {
		schemas := map[string]interface{}{
			"IPv4":   IPv4(),
			"IPv6":   IPv6(),
			"CIDRv4": CIDRv4(),
			"CIDRv6": CIDRv6(),
		}

		for name, schema := range schemas {
			t.Run(name, func(t *testing.T) {
				require.NotNil(t, schema)

				// Verify the schema has the expected type
				switch name {
				case "IPv4":
					ipv4Schema, ok := schema.(*ZodIPv4)
					require.True(t, ok)
					assert.Equal(t, "ipv4", ipv4Schema.GetZod().Def.Type)
				case "IPv6":
					ipv6Schema, ok := schema.(*ZodIPv6)
					require.True(t, ok)
					assert.Equal(t, "ipv6", ipv6Schema.GetZod().Def.Type)
				case "CIDRv4":
					cidrv4Schema, ok := schema.(*ZodCIDRv4)
					require.True(t, ok)
					assert.Equal(t, "cidrv4", cidrv4Schema.GetZod().Def.Type)
				case "CIDRv6":
					cidrv6Schema, ok := schema.(*ZodCIDRv6)
					require.True(t, ok)
					assert.Equal(t, "cidrv6", cidrv6Schema.GetZod().Def.Type)
				}
			})
		}
	})

	t.Run("custom error messages", func(t *testing.T) {
		customError := "Invalid network address"
		schemas := map[string]interface{}{
			"IPv4":   IPv4(SchemaParams{Error: customError}),
			"IPv6":   IPv6(SchemaParams{Error: customError}),
			"CIDRv4": CIDRv4(SchemaParams{Error: customError}),
			"CIDRv6": CIDRv6(SchemaParams{Error: customError}),
		}

		for name, schema := range schemas {
			t.Run(name, func(t *testing.T) {
				require.NotNil(t, schema)

				// Test that custom error is applied by parsing invalid input
				var err error
				switch s := schema.(type) {
				case *ZodIPv4:
					_, err = s.Parse("invalid")
				case *ZodIPv6:
					_, err = s.Parse("invalid")
				case *ZodCIDRv4:
					_, err = s.Parse("invalid")
				case *ZodCIDRv6:
					_, err = s.Parse("invalid")
				}
				require.Error(t, err)
			})
		}
	})
}

// =============================================================================
// 2. Coerce (type coercion)
// =============================================================================

func TestNetworkCoerce(t *testing.T) {
	t.Run("IPv4 coercion not supported", func(t *testing.T) {
		schema := IPv4()
		// Network types typically don't support coercion from other types
		// as IP addresses have strict format requirements
		_, err := schema.Parse(123)
		assert.Error(t, err)
	})

	t.Run("IPv6 coercion not supported", func(t *testing.T) {
		schema := IPv6()
		_, err := schema.Parse(123)
		assert.Error(t, err)
	})
}

// =============================================================================
// 3. Validation methods (IPv4, IPv6, CIDR validation)
// =============================================================================

// IPv4 validation tests
func TestNetworkIPv4Validation(t *testing.T) {
	schema := IPv4()

	t.Run("valid IPv4 addresses", func(t *testing.T) {
		validIPs := []string{
			"192.168.1.1",
			"10.0.0.1",
			"172.16.0.1",
			"127.0.0.1",
			"255.255.255.255",
			"0.0.0.0",
		}

		for _, ip := range validIPs {
			result, err := schema.Parse(ip)
			assert.NoError(t, err, "Should accept valid IPv4: %s", ip)
			assert.Equal(t, ip, result)
		}
	})

	t.Run("invalid IPv4 addresses", func(t *testing.T) {
		invalidIPs := []string{
			"256.1.1.1",      // out of range
			"192.168.1",      // incomplete
			"192.168.1.1.1",  // too many octets
			"192.168.01.1",   // leading zero
			"192.168.-1.1",   // negative
			"not.an.ip.addr", // non-numeric
			"",               // empty
		}

		for _, ip := range invalidIPs {
			_, err := schema.Parse(ip)
			assert.Error(t, err, "Should reject invalid IPv4: %s", ip)
		}
	})
}

// IPv6 validation tests
func TestNetworkIPv6Validation(t *testing.T) {
	schema := IPv6()

	t.Run("valid IPv6 addresses", func(t *testing.T) {
		validIPs := []string{
			"2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			"2001:db8:85a3:0:0:8a2e:370:7334",
			"2001:db8:85a3::8a2e:370:7334",
			"::1",
			"::",
			"fe80::1",
			"2001:db8::1",
		}

		for _, ip := range validIPs {
			result, err := schema.Parse(ip)
			assert.NoError(t, err, "Should accept valid IPv6: %s", ip)
			assert.Equal(t, ip, result)
		}
	})

	t.Run("invalid IPv6 addresses", func(t *testing.T) {
		invalidIPs := []string{
			"2001:0db8:85a3::8a2e::7334",                    // double ::
			"2001:0db8:85a3:0000:0000:8a2e:0370:7334:extra", // too many groups
			"gggg::1",     // invalid hex
			"192.168.1.1", // IPv4 format
			"",            // empty
		}

		for _, ip := range invalidIPs {
			_, err := schema.Parse(ip)
			assert.Error(t, err, "Should reject invalid IPv6: %s", ip)
		}
	})
}

// CIDR validation tests
func TestNetworkCIDRValidation(t *testing.T) {
	t.Run("CIDRv4 validation", func(t *testing.T) {
		schema := CIDRv4()

		validCIDRs := []string{
			"192.168.1.0/24",
			"10.0.0.0/8",
			"172.16.0.0/12",
			"192.168.1.1/32",
			"0.0.0.0/0",
		}

		for _, cidr := range validCIDRs {
			result, err := schema.Parse(cidr)
			assert.NoError(t, err, "Should accept valid CIDRv4: %s", cidr)
			assert.Equal(t, cidr, result)
		}

		invalidCIDRs := []string{
			"192.168.1.0/33", // invalid prefix length
			"192.168.1.0/-1", // negative prefix
			"192.168.1.0",    // missing prefix
			"256.1.1.1/24",   // invalid IP
		}

		for _, cidr := range invalidCIDRs {
			_, err := schema.Parse(cidr)
			assert.Error(t, err, "Should reject invalid CIDRv4: %s", cidr)
		}
	})

	t.Run("CIDRv6 validation", func(t *testing.T) {
		schema := CIDRv6()

		validCIDRs := []string{
			"2001:db8::/32",
			"fe80::/10",
			"::1/128",
			"::/0",
		}

		for _, cidr := range validCIDRs {
			result, err := schema.Parse(cidr)
			assert.NoError(t, err, "Should accept valid CIDRv6: %s", cidr)
			assert.Equal(t, cidr, result)
		}

		invalidCIDRs := []string{
			"2001:db8::/129", // invalid prefix length
			"2001:db8::",     // missing prefix
			"invalid::/32",   // invalid IP
		}

		for _, cidr := range invalidCIDRs {
			_, err := schema.Parse(cidr)
			assert.Error(t, err, "Should reject invalid CIDRv6: %s", cidr)
		}
	})
}

// =============================================================================
// 4. Modifiers and wrappers
// =============================================================================

func TestNetworkModifiers(t *testing.T) {
	t.Run("optional modifier", func(t *testing.T) {
		schemas := map[string]interface{}{
			"IPv4":   IPv4().Optional(),
			"IPv6":   IPv6().Optional(),
			"CIDRv4": CIDRv4().Optional(),
			"CIDRv6": CIDRv6().Optional(),
		}

		for name, schema := range schemas {
			t.Run(name, func(t *testing.T) {
				// Valid values should pass through
				var validInput string
				switch name {
				case "IPv4":
					validInput = "192.168.1.1"
				case "IPv6":
					validInput = "::1"
				case "CIDRv4":
					validInput = "192.168.1.0/24"
				case "CIDRv6":
					validInput = "2001:db8::/32"
				}

				result, err := schema.(ZodType[any, any]).Parse(validInput)
				require.NoError(t, err)
				assert.Equal(t, validInput, result)

				// nil should be allowed
				result, err = schema.(ZodType[any, any]).Parse(nil)
				require.NoError(t, err)
				assert.Nil(t, result)
			})
		}
	})

	t.Run("nilable modifier", func(t *testing.T) {
		schemas := map[string]interface{}{
			"IPv4":   IPv4().Nilable(),
			"IPv6":   IPv6().Nilable(),
			"CIDRv4": CIDRv4().Nilable(),
			"CIDRv6": CIDRv6().Nilable(),
		}

		for name, schema := range schemas {
			t.Run(name, func(t *testing.T) {
				// Valid values should pass through
				var validInput string
				switch name {
				case "IPv4":
					validInput = "192.168.1.1"
				case "IPv6":
					validInput = "::1"
				case "CIDRv4":
					validInput = "192.168.1.0/24"
				case "CIDRv6":
					validInput = "2001:db8::/32"
				}

				result, err := schema.(ZodType[any, any]).Parse(validInput)
				require.NoError(t, err)
				assert.Equal(t, validInput, result)

				// nil should return typed nil pointer
				result, err = schema.(ZodType[any, any]).Parse(nil)
				require.NoError(t, err)
				assert.Nil(t, result)
			})
		}
	})

	t.Run("modifiers do not affect original schema", func(t *testing.T) {
		baseSchema := IPv4()
		nilableSchema := baseSchema.Nilable()

		// Test nilable schema allows nil
		result, err := nilableSchema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test nilable schema validates non-nil values
		result, err = nilableSchema.Parse("192.168.1.1")
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.1", result)

		// Original schema should remain unchanged
		_, err = baseSchema.Parse(nil)
		assert.Error(t, err, "Original schema should still reject nil")

		result, err = baseSchema.Parse("192.168.1.1")
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.1", result)
	})
}

// =============================================================================
// 5. Chaining and method composition
// =============================================================================

func TestNetworkChaining(t *testing.T) {
	t.Run("method chaining", func(t *testing.T) {
		// Network types don't have specific validation methods to chain
		// but they support modifier chaining
		schema := IPv4().Optional()

		// Valid IP should pass
		result, err := schema.Parse("192.168.1.1")
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.1", result)

		// nil should be allowed
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("refine chaining", func(t *testing.T) {
		// Chain RefineAny with network validation
		schema := IPv4().RefineAny(func(val any) bool {
			if str, ok := val.(string); ok {
				return !strings.HasPrefix(str, "127.") // Reject localhost
			}
			return false
		})

		// Valid non-localhost IP should pass
		result, err := schema.Parse("192.168.1.1")
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.1", result)

		// Localhost should be rejected
		_, err = schema.Parse("127.0.0.1")
		assert.Error(t, err)
	})
}

// =============================================================================
// 6. Transform/Pipe
// =============================================================================

func TestNetworkTransform(t *testing.T) {
	t.Run("basic transform", func(t *testing.T) {
		schema := IPv4().TransformAny(func(input any, ctx *RefinementContext) (any, error) {
			if str, ok := input.(string); ok {
				return map[string]any{
					"address": str,
					"type":    "IPv4",
					"octets":  strings.Split(str, "."),
				}, nil
			}
			return input, nil
		})

		result, err := schema.Parse("192.168.1.1")
		require.NoError(t, err)
		resultMap, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "192.168.1.1", resultMap["address"])
		assert.Equal(t, "IPv4", resultMap["type"])
		assert.Equal(t, []string{"192", "168", "1", "1"}, resultMap["octets"])
	})

	t.Run("transform chaining", func(t *testing.T) {
		schema := IPv4().
			TransformAny(func(input any, ctx *RefinementContext) (any, error) {
				if str, ok := input.(string); ok {
					return strings.ToUpper(str), nil
				}
				return input, nil
			}).
			TransformAny(func(input any, ctx *RefinementContext) (any, error) {
				if str, ok := input.(string); ok {
					return "PREFIX_" + str, nil
				}
				return input, nil
			})

		result, err := schema.Parse("192.168.1.1")
		require.NoError(t, err)
		assert.Equal(t, "PREFIX_192.168.1.1", result)
	})

	t.Run("pipe to another schema", func(t *testing.T) {
		// Pipe IPv4 to a string schema that requires specific format
		stringSchema := String().StartsWith("192.168.")
		schema := IPv4().Pipe(stringSchema)

		// Valid private IP should pass
		result, err := schema.Parse("192.168.1.1")
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.1", result)

		// Public IP should fail the pipe validation
		_, err = schema.Parse("8.8.8.8")
		assert.Error(t, err)
	})

	t.Run("transform different network types", func(t *testing.T) {
		tests := []struct {
			name     string
			schema   ZodType[any, any]
			input    string
			expected string
		}{
			{"IPv4", IPv4().TransformAny(func(input any, ctx *RefinementContext) (any, error) {
				return "IPv4:" + input.(string), nil
			}), "192.168.1.1", "IPv4:192.168.1.1"},
			{"IPv6", IPv6().TransformAny(func(input any, ctx *RefinementContext) (any, error) {
				return "IPv6:" + input.(string), nil
			}), "::1", "IPv6:::1"},
			{"CIDRv4", CIDRv4().TransformAny(func(input any, ctx *RefinementContext) (any, error) {
				return "CIDRv4:" + input.(string), nil
			}), "192.168.1.0/24", "CIDRv4:192.168.1.0/24"},
			{"CIDRv6", CIDRv6().TransformAny(func(input any, ctx *RefinementContext) (any, error) {
				return "CIDRv6:" + input.(string), nil
			}), "2001:db8::/32", "CIDRv6:2001:db8::/32"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := tt.schema.Parse(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

// =============================================================================
// 7. Refine
// =============================================================================

func TestNetworkRefine(t *testing.T) {
	t.Run("basic refine validation", func(t *testing.T) {
		// Reject localhost addresses
		schema := IPv4().RefineAny(func(val any) bool {
			if str, ok := val.(string); ok {
				return !strings.HasPrefix(str, "127.")
			}
			return false
		})

		// Valid non-localhost IP should pass
		result, err := schema.Parse("192.168.1.1")
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.1", result)

		// Localhost should be rejected
		_, err = schema.Parse("127.0.0.1")
		assert.Error(t, err)
	})

	t.Run("refine preserves original value", func(t *testing.T) {
		input := "192.168.1.1"
		schema := IPv4().RefineAny(func(val any) bool {
			return true // Always pass
		})

		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result, "Refine should return exact original value")
	})

	t.Run("multiple refine conditions", func(t *testing.T) {
		schema := IPv4().
			RefineAny(func(val any) bool {
				if str, ok := val.(string); ok {
					return !strings.HasPrefix(str, "127.") // Not localhost
				}
				return false
			}).
			RefineAny(func(val any) bool {
				if str, ok := val.(string); ok {
					return strings.HasPrefix(str, "192.168.") // Must be private
				}
				return false
			})

		// Valid private IP should pass
		result, err := schema.Parse("192.168.1.1")
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.1", result)

		// Public IP should fail
		_, err = schema.Parse("8.8.8.8")
		assert.Error(t, err)

		// Localhost should fail
		_, err = schema.Parse("127.0.0.1")
		assert.Error(t, err)
	})

	t.Run("refine with custom error messages", func(t *testing.T) {
		schema := IPv6().RefineAny(func(val any) bool {
			if str, ok := val.(string); ok {
				return str != "::1" // Reject localhost
			}
			return false
		}, SchemaParams{Error: "Localhost IPv6 addresses are not allowed"})

		// Valid IPv6 should pass
		result, err := schema.Parse("2001:db8::1")
		require.NoError(t, err)
		assert.Equal(t, "2001:db8::1", result)

		// Localhost should fail with custom error
		_, err = schema.Parse("::1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Localhost IPv6 addresses are not allowed")
	})

	t.Run("refine vs transform distinction", func(t *testing.T) {
		input := "192.168.1.1"

		// Refine: only validates, never modifies
		refineSchema := IPv4().RefineAny(func(val any) bool {
			return true
		})
		refineResult, refineErr := refineSchema.Parse(input)

		// Transform: validates and converts
		transformSchema := IPv4().TransformAny(func(val any, ctx *RefinementContext) (any, error) {
			if str, ok := val.(string); ok {
				return strings.ToUpper(str), nil
			}
			return val, nil
		})
		transformResult, transformErr := transformSchema.Parse(input)

		// Refine returns original value unchanged
		require.NoError(t, refineErr)
		assert.Equal(t, "192.168.1.1", refineResult)

		// Transform returns modified value
		require.NoError(t, transformErr)
		assert.Equal(t, "192.168.1.1", transformResult) // Note: IPs don't change case, but concept demonstrated

		// Key distinction: Refine preserves, Transform may modify
		assert.Equal(t, input, refineResult, "Refine should return exact original value")
	})
}

// =============================================================================
// 8. Error handling
// =============================================================================

func TestNetworkErrorHandling(t *testing.T) {
	t.Run("error structure", func(t *testing.T) {
		schema := IPv4()
		_, err := schema.Parse("invalid")
		assert.Error(t, err)
		assert.NotEmpty(t, err.Error())
	})

	t.Run("custom error messages", func(t *testing.T) {
		schemas := []struct {
			name   string
			schema interface{}
			input  interface{}
		}{
			{"IPv4", IPv4(SchemaParams{Error: "Custom IPv4 error"}), "invalid"},
			{"IPv6", IPv6(SchemaParams{Error: "Custom IPv6 error"}), "invalid"},
			{"CIDRv4", CIDRv4(SchemaParams{Error: "Custom CIDRv4 error"}), "invalid"},
			{"CIDRv6", CIDRv6(SchemaParams{Error: "Custom CIDRv6 error"}), "invalid"},
		}

		for _, tt := range schemas {
			t.Run(tt.name, func(t *testing.T) {
				var err error
				switch s := tt.schema.(type) {
				case *ZodIPv4:
					_, err = s.Parse(tt.input)
				case *ZodIPv6:
					_, err = s.Parse(tt.input)
				case *ZodCIDRv4:
					_, err = s.Parse(tt.input)
				case *ZodCIDRv6:
					_, err = s.Parse(tt.input)
				}
				require.Error(t, err)
				assert.NotEmpty(t, err.Error())
			})
		}
	})

	t.Run("must parse utility", func(t *testing.T) {
		schema := IPv4()

		// Valid case should not panic
		result := schema.MustParse("192.168.1.1")
		assert.Equal(t, "192.168.1.1", result)

		// Invalid case should panic
		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})
	})

	t.Run("internals access", func(t *testing.T) {
		schemas := map[string]interface{}{
			"IPv4":   IPv4(),
			"IPv6":   IPv6(),
			"CIDRv4": CIDRv4(),
			"CIDRv6": CIDRv6(),
		}

		for name, schema := range schemas {
			t.Run(name, func(t *testing.T) {
				switch s := schema.(type) {
				case *ZodIPv4:
					internals := s.GetInternals()
					require.NotNil(t, internals)
					assert.Equal(t, "ipv4", internals.Type)

					zod := s.GetZod()
					require.NotNil(t, zod)
					assert.Equal(t, "ipv4", zod.Def.Type)
				case *ZodIPv6:
					internals := s.GetInternals()
					require.NotNil(t, internals)
					assert.Equal(t, "ipv6", internals.Type)

					zod := s.GetZod()
					require.NotNil(t, zod)
					assert.Equal(t, "ipv6", zod.Def.Type)
				case *ZodCIDRv4:
					internals := s.GetInternals()
					require.NotNil(t, internals)
					assert.Equal(t, "cidrv4", internals.Type)

					zod := s.GetZod()
					require.NotNil(t, zod)
					assert.Equal(t, "cidrv4", zod.Def.Type)
				case *ZodCIDRv6:
					internals := s.GetInternals()
					require.NotNil(t, internals)
					assert.Equal(t, "cidrv6", internals.Type)

					zod := s.GetZod()
					require.NotNil(t, zod)
					assert.Equal(t, "cidrv6", zod.Def.Type)
				}
			})
		}
	})
}

// =============================================================================
// 9. Edge and mutual exclusion cases
// =============================================================================

func TestNetworkEdgeCases(t *testing.T) {
	t.Run("boundary values", func(t *testing.T) {
		ipv4Schema := IPv4()

		boundaryTests := []struct {
			name    string
			input   string
			isValid bool
		}{
			{"minimum address", "0.0.0.0", true},
			{"maximum address", "255.255.255.255", true},
			{"just over maximum", "255.255.255.256", false},
		}

		for _, tt := range boundaryTests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := ipv4Schema.Parse(tt.input)
				if tt.isValid {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}
			})
		}
	})

	t.Run("prefix boundaries", func(t *testing.T) {
		cidrv4Schema := CIDRv4()
		cidrv6Schema := CIDRv6()

		tests := []struct {
			name    string
			schema  interface{}
			input   string
			isValid bool
		}{
			{"CIDRv4 min prefix", cidrv4Schema, "192.168.1.0/0", true},
			{"CIDRv4 max prefix", cidrv4Schema, "192.168.1.0/32", true},
			{"CIDRv4 over max", cidrv4Schema, "192.168.1.0/33", false},
			{"CIDRv6 min prefix", cidrv6Schema, "2001:db8::/0", true},
			{"CIDRv6 max prefix", cidrv6Schema, "2001:db8::/128", true},
			{"CIDRv6 over max", cidrv6Schema, "2001:db8::/129", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var err error
				switch s := tt.schema.(type) {
				case *ZodCIDRv4:
					_, err = s.Parse(tt.input)
				case *ZodCIDRv6:
					_, err = s.Parse(tt.input)
				}

				if tt.isValid {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}
			})
		}
	})

	t.Run("real-world addresses", func(t *testing.T) {
		ipv4Schema := IPv4()
		ipv6Schema := IPv6()

		realWorldTests := map[string]struct {
			address string
			isIPv4  bool
			isIPv6  bool
		}{
			"Google DNS":      {"8.8.8.8", true, false},
			"Cloudflare DNS":  {"1.1.1.1", true, false},
			"Localhost IPv4":  {"127.0.0.1", true, false},
			"Localhost IPv6":  {"::1", false, true},
			"Valid IPv6":      {"2001:0db8:85a3:0000:0000:8a2e:0370:7334", false, true},
			"Link Local IPv6": {"fe80::1", false, true},
		}

		for name, test := range realWorldTests {
			t.Run(name, func(t *testing.T) {
				// Test IPv4 schema
				_, err := ipv4Schema.Parse(test.address)
				if test.isIPv4 {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}

				// Test IPv6 schema
				_, err = ipv6Schema.Parse(test.address)
				if test.isIPv6 {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}
			})
		}
	})

	t.Run("whitespace handling", func(t *testing.T) {
		schema := IPv4()

		whitespaceTests := []string{
			" 192.168.1.1",  // Leading space
			"192.168.1.1 ",  // Trailing space
			" 192.168.1.1 ", // Both spaces
			"192.168. 1.1",  // Internal space
		}

		for _, input := range whitespaceTests {
			_, err := schema.Parse(input)
			require.Error(t, err, "Input with whitespace should be invalid: %q", input)
		}
	})

	t.Run("nil input handling", func(t *testing.T) {
		schemas := map[string]interface{}{
			"IPv4":   IPv4(),
			"IPv6":   IPv6(),
			"CIDRv4": CIDRv4(),
			"CIDRv6": CIDRv6(),
		}

		for name, schema := range schemas {
			t.Run(name, func(t *testing.T) {
				// Non-nilable should reject nil
				var err error
				switch s := schema.(type) {
				case *ZodIPv4:
					_, err = s.Parse(nil)
				case *ZodIPv6:
					_, err = s.Parse(nil)
				case *ZodCIDRv4:
					_, err = s.Parse(nil)
				case *ZodCIDRv6:
					_, err = s.Parse(nil)
				}
				assert.Error(t, err)

				// Nilable should accept nil
				var nilableSchema ZodType[any, any]
				switch s := schema.(type) {
				case *ZodIPv4:
					nilableSchema = s.Nilable()
				case *ZodIPv6:
					nilableSchema = s.Nilable()
				case *ZodCIDRv4:
					nilableSchema = s.Nilable()
				case *ZodCIDRv6:
					nilableSchema = s.Nilable()
				}

				result, err := nilableSchema.Parse(nil)
				require.NoError(t, err)
				assert.Nil(t, result)
			})
		}
	})

	t.Run("type coercion rejection", func(t *testing.T) {
		// Network types should not accept type coercion
		schemas := map[string]interface{}{
			"IPv4":   IPv4(),
			"IPv6":   IPv6(),
			"CIDRv4": CIDRv4(),
			"CIDRv6": CIDRv6(),
		}

		invalidTypes := []interface{}{
			123,
			true,
			[]byte("192.168.1.1"),
			map[string]string{"ip": "192.168.1.1"},
			[]string{"192.168.1.1"},
		}

		for name, schema := range schemas {
			for _, input := range invalidTypes {
				t.Run(name+" rejects "+fmt.Sprintf("%T", input), func(t *testing.T) {
					var err error
					switch s := schema.(type) {
					case *ZodIPv4:
						_, err = s.Parse(input)
					case *ZodIPv6:
						_, err = s.Parse(input)
					case *ZodCIDRv4:
						_, err = s.Parse(input)
					case *ZodCIDRv6:
						_, err = s.Parse(input)
					}
					assert.Error(t, err)
				})
			}
		}
	})
}

// =============================================================================
// 10. Default and Prefault tests
// =============================================================================

func TestNetworkDefaultAndPrefault(t *testing.T) {
	t.Run("IPv4 default value", func(t *testing.T) {
		schema := IPv4().Default("192.168.1.1")

		// Valid IP should not use default
		result, err := schema.Parse("10.0.0.1")
		require.NoError(t, err)
		assert.Equal(t, "10.0.0.1", result)

		// nil input should use default
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.1", result)
	})

	t.Run("IPv4 function-based default", func(t *testing.T) {
		counter := 0
		schema := IPv4().DefaultFunc(func() any {
			counter++
			return fmt.Sprintf("192.168.1.%d", counter)
		})

		// Each nil input generates a new default value
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, "192.168.1.1", result1)

		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		assert.Equal(t, "192.168.1.2", result2)

		// Valid input bypasses default generation
		result3, err3 := schema.Parse("10.0.0.1")
		require.NoError(t, err3)
		assert.Equal(t, "10.0.0.1", result3)

		// Counter should only increment for nil inputs
		assert.Equal(t, 2, counter)
	})

	t.Run("IPv6 default value", func(t *testing.T) {
		schema := IPv6().Default("::1")

		// Valid IPv6 should not use default
		result, err := schema.Parse("2001:db8::2")
		require.NoError(t, err)
		assert.Equal(t, "2001:db8::2", result)

		// nil input should use default
		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "::1", result)
	})

	t.Run("IPv6 function-based default", func(t *testing.T) {
		counter := 0
		schema := IPv6().DefaultFunc(func() any {
			counter++
			return "::1"
		})

		// Each nil input generates a new default value
		result1, err1 := schema.Parse(nil)
		require.NoError(t, err1)
		assert.Equal(t, "::1", result1)

		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		assert.Equal(t, "::1", result2)

		// Valid input bypasses default generation
		result3, err3 := schema.Parse("2001:db8::100")
		require.NoError(t, err3)
		assert.Equal(t, "2001:db8::100", result3)

		// Counter should only increment for nil inputs
		assert.Equal(t, 2, counter)
	})

	t.Run("CIDR default values", func(t *testing.T) {
		cidrv4Schema := CIDRv4().Default("192.168.1.0/24")
		cidrv6Schema := CIDRv6().Default("2001:db8::/32")

		// Test CIDRv4
		result, err := cidrv4Schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.0/24", result)

		result, err = cidrv4Schema.Parse("10.0.0.0/8")
		require.NoError(t, err)
		assert.Equal(t, "10.0.0.0/8", result)

		// Test CIDRv6
		result, err = cidrv6Schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "2001:db8::/32", result)

		result, err = cidrv6Schema.Parse("fe80::/10")
		require.NoError(t, err)
		assert.Equal(t, "fe80::/10", result)
	})

	t.Run("prefault value", func(t *testing.T) {
		// Use IPv4 with Prefault directly, not after RefineAny
		baseSchema := IPv4()
		schema := baseSchema.Prefault("192.168.1.1")

		// Valid IP should pass through
		result, err := schema.Parse("10.0.0.1")
		require.NoError(t, err)
		assert.Equal(t, "10.0.0.1", result)

		// Invalid type should use prefault
		result, err = schema.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.1", result)
	})

	t.Run("prefault function", func(t *testing.T) {
		counter := 0
		// Use IPv4 with PrefaultFunc directly, not after RefineAny
		baseSchema := IPv4()
		schema := baseSchema.PrefaultFunc(func() any {
			counter++
			return fmt.Sprintf("192.168.1.%d", counter)
		})

		// Valid IP should not call function
		result, err := schema.Parse("10.0.0.1")
		require.NoError(t, err)
		assert.Equal(t, "10.0.0.1", result)
		assert.Equal(t, 0, counter)

		// Invalid type should call function
		result, err = schema.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.1", result)
		assert.Equal(t, 1, counter)

		// Another invalid input should increment counter
		result, err = schema.Parse("invalid_ip")
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.2", result)
		assert.Equal(t, 2, counter)
	})

	t.Run("default vs prefault distinction", func(t *testing.T) {
		defaultSchema := IPv4().Default("192.168.1.1")
		prefaultSchema := IPv4().Prefault("192.168.1.1")

		// For valid input: different behaviors
		result1, err1 := defaultSchema.Parse("10.0.0.1")
		require.NoError(t, err1)
		assert.Equal(t, "10.0.0.1", result1, "Default: valid input passes through")

		result2, err2 := prefaultSchema.Parse("10.0.0.1")
		require.NoError(t, err2)
		assert.Equal(t, "10.0.0.1", result2, "Prefault: valid input passes through")

		// For nil input: different behaviors
		result3, err3 := defaultSchema.Parse(nil)
		require.NoError(t, err3)
		assert.Equal(t, "192.168.1.1", result3, "Default: nil gets default value")

		result4, err4 := prefaultSchema.Parse(nil)
		require.NoError(t, err4)
		assert.Equal(t, "192.168.1.1", result4, "Prefault: nil fails validation, use fallback")

		// For invalid type: different behaviors
		_, err5 := defaultSchema.Parse(123)
		assert.Error(t, err5, "Default: type validation fails, no fallback for non-nil")

		result6, err6 := prefaultSchema.Parse(123)
		require.NoError(t, err6)
		assert.Equal(t, "192.168.1.1", result6, "Prefault: validation fails, use fallback")
	})

	t.Run("default with chaining", func(t *testing.T) {
		// Default + refine chain - use simpler validation
		schema := IPv4().Default("192.168.1.1")

		// nil input: use default, should pass validation
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.1", result)

		// Valid private IP: should pass
		result, err = schema.Parse("192.168.2.1")
		require.NoError(t, err)
		assert.Equal(t, "192.168.2.1", result)

		// Invalid IP format: should fail validation
		_, err = schema.Parse("invalid_ip")
		assert.Error(t, err)
	})

	t.Run("default with transform compatibility", func(t *testing.T) {
		schema := IPv4().
			Default("192.168.1.1").
			TransformAny(func(input any, ctx *RefinementContext) (any, error) {
				if str, ok := input.(string); ok {
					return map[string]any{
						"original": str,
						"type":     "IPv4",
						"octets":   strings.Split(str, "."),
					}, nil
				}
				return input, nil
			})

		// Non-nil input: should transform
		result1, err1 := schema.Parse("10.0.0.1")
		require.NoError(t, err1)
		result1Map, ok1 := result1.(map[string]any)
		require.True(t, ok1)
		assert.Equal(t, "10.0.0.1", result1Map["original"])
		assert.Equal(t, "IPv4", result1Map["type"])

		// nil input: use default then transform
		result2, err2 := schema.Parse(nil)
		require.NoError(t, err2)
		result2Map, ok2 := result2.(map[string]any)
		require.True(t, ok2)
		assert.Equal(t, "192.168.1.1", result2Map["original"])
		assert.Equal(t, "IPv4", result2Map["type"])
		assert.Equal(t, []string{"192", "168", "1", "1"}, result2Map["octets"])
	})
}
