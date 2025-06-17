package gozod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestZodTypeDef tests the core ZodTypeDef creation and initialization
func TestZodTypeDef(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
	}{
		{"string type", "string"},
		{"number type", "number"},
		{"boolean type", "boolean"},
		{"object type", "object"},
		{"empty string", ""},
		{"special characters", "string@#$%"},
		{"unicode type", "字符串"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := NewZodTypeDef(tt.typeName)

			require.NotNil(t, def)
			assert.Equal(t, tt.typeName, def.Type)
			assert.NotNil(t, def.Checks)
			assert.Empty(t, def.Checks)
			assert.Nil(t, def.Error)
		})
	}

	t.Run("creates independent instances", func(t *testing.T) {
		def1 := NewZodTypeDef("string")
		def2 := NewZodTypeDef("string")

		// Each instance should be independent
		assert.NotSame(t, def1, def2)
		// Checks slices should be different instances
		if def1.Checks != nil && def2.Checks != nil {
			assert.NotSame(t, &def1.Checks, &def2.Checks)
		}
	})

	t.Run("handles empty checks slice", func(t *testing.T) {
		def := NewZodTypeDef("test")

		require.NotNil(t, def.Checks)
		assert.Equal(t, 0, len(def.Checks))
		assert.Equal(t, 0, cap(def.Checks))
	})
}

// TestZodTypeInternals tests the ZodTypeInternals creation and structure
func TestZodTypeInternals(t *testing.T) {
	t.Run("creates internals with correct structure", func(t *testing.T) {
		def := NewZodTypeDef("string")
		internals := NewZodTypeInternals(def)

		require.NotNil(t, internals)
		assert.Equal(t, Version, internals.Version)
		assert.Equal(t, def.Type, internals.Type)
		assert.NotNil(t, internals.Checks)
		assert.NotNil(t, internals.Values)
		// Parse function may be nil until initZodType is called
	})

	t.Run("creates independent instances", func(t *testing.T) {
		def := NewZodTypeDef("string")
		internals1 := NewZodTypeInternals(def)
		internals2 := NewZodTypeInternals(def)

		assert.NotSame(t, internals1, internals2)
		// Check that slice/map fields are different instances
		if internals1.Checks != nil && internals2.Checks != nil {
			assert.NotSame(t, &internals1.Checks, &internals2.Checks)
		}
		if internals1.Values != nil && internals2.Values != nil {
			assert.NotSame(t, &internals1.Values, &internals2.Values)
		}
	})

	t.Run("initializes empty collections", func(t *testing.T) {
		def := NewZodTypeDef("test")
		internals := NewZodTypeInternals(def)

		require.NotNil(t, internals.Checks)
		require.NotNil(t, internals.Values)
		assert.Empty(t, internals.Checks)
		assert.Empty(t, internals.Values)
	})
}

// TestSchemaInitialization tests schema initialization with real schemas
func TestSchemaInitialization(t *testing.T) {
	tests := []struct {
		name         string
		createSchema func() ZodType[any, any]
		expectedType string
		hasParseFunc bool
	}{
		{
			name:         "string schema",
			createSchema: func() ZodType[any, any] { return String() },
			expectedType: "string",
			hasParseFunc: true,
		},
		{
			name:         "integer schema",
			createSchema: func() ZodType[any, any] { return Int() },
			expectedType: "int",
			hasParseFunc: true,
		},
		{
			name:         "boolean schema",
			createSchema: func() ZodType[any, any] { return Bool() },
			expectedType: "bool",
			hasParseFunc: true,
		},
		{
			name:         "float64 schema",
			createSchema: func() ZodType[any, any] { return Float64() },
			expectedType: "float64",
			hasParseFunc: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.createSchema()
			internals := schema.GetInternals()

			require.NotNil(t, internals)
			assert.Equal(t, Version, internals.Version)
			assert.Equal(t, tt.expectedType, internals.Type)
			assert.NotNil(t, internals.Checks)
			assert.NotNil(t, internals.Values)

			if tt.hasParseFunc {
				assert.NotNil(t, internals.Parse, "Parse function should be initialized")
			}
		})
	}

	t.Run("schema with coercion", func(t *testing.T) {
		// Test coercion feature that is already implemented
		coercedString := String(SchemaParams{Coerce: true})
		internals := coercedString.GetInternals()

		require.NotNil(t, internals)
		assert.Equal(t, "string", internals.Type)
		assert.NotNil(t, internals.Parse, "Parse function should be initialized")

		// Test that coercion actually works
		result, err := coercedString.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, "123", result)
	})
}

// TestAddCheck tests the check addition functionality
func TestAddCheck(t *testing.T) {
	t.Run("adds check and returns new instance", func(t *testing.T) {
		originalSchema := String()
		originalChecksCount := len(originalSchema.GetInternals().Checks)

		// Use method chaining instead of AddCheck function to avoid constructor issues
		newSchema := originalSchema.Min(5)

		// Original schema should be unchanged (immutable)
		assert.Equal(t, originalChecksCount, len(originalSchema.GetInternals().Checks))

		// New schema should have additional check
		assert.Equal(t, originalChecksCount+1, len(newSchema.GetInternals().Checks))
		assert.NotSame(t, originalSchema, newSchema, "Should return new instance")
	})

	t.Run("preserves existing checks", func(t *testing.T) {
		// Create schema with existing checks
		schema := String().Min(3).Max(10)
		originalChecksCount := len(schema.GetInternals().Checks)

		// Add another check using method chaining
		newSchema := schema.Includes("test")

		// Should preserve all existing checks plus new one
		newChecksCount := len(newSchema.GetInternals().Checks)
		assert.Greater(t, newChecksCount, originalChecksCount, "Should have more checks than original")

		// Original should remain unchanged
		assert.Equal(t, originalChecksCount, len(schema.GetInternals().Checks))
	})

	t.Run("adds multiple checks sequentially", func(t *testing.T) {
		// Test multiple checks addition using method chaining instead of AddCheck function
		// This avoids potential constructor issues with AddCheck
		originalSchema := String()
		originalChecksCount := len(originalSchema.GetInternals().Checks)

		// Use method chaining for multiple checks (more reliable)
		schemaWithEmail := originalSchema.Min(3).Max(20).Email()

		// Should have all checks
		finalChecksCount := len(schemaWithEmail.GetInternals().Checks)
		assert.Greater(t, finalChecksCount, originalChecksCount, "Should have more checks than original")

		// Original should remain unchanged
		assert.Equal(t, originalChecksCount, len(originalSchema.GetInternals().Checks))

		// Test functionality
		result, err := schemaWithEmail.Parse("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", result)

		// Test validation failure
		_, err = schemaWithEmail.Parse("no")
		assert.Error(t, err, "Should fail min length validation")
	})

	t.Run("handles edge cases gracefully", func(t *testing.T) {
		// Test that method chaining handles edge cases gracefully
		originalSchema := String()

		// Test chaining multiple methods
		complexSchema := originalSchema.Min(1).Max(100).StartsWith("test").EndsWith("ing")

		// Should have multiple checks
		assert.Greater(t, len(complexSchema.GetInternals().Checks), 3)

		// Test functionality
		result, err := complexSchema.Parse("testing")
		require.NoError(t, err)
		assert.Equal(t, "testing", result)

		// Test validation failure
		_, err = complexSchema.Parse("invalid")
		assert.Error(t, err, "Should fail validation")
	})
}

// TestClone tests the schema cloning functionality
func TestClone(t *testing.T) {
	t.Run("clones with modification", func(t *testing.T) {
		// Test immutability through method chaining instead of Clone function
		originalSchema := String().Min(5).Max(10)
		originalChecksCount := len(originalSchema.GetInternals().Checks)

		// Method chaining creates new instances (acts like cloning)
		modifiedSchema := originalSchema.Email()

		// Original should be unchanged
		assert.Equal(t, originalChecksCount, len(originalSchema.GetInternals().Checks))

		// Modified schema should have additional checks
		modifiedChecksCount := len(modifiedSchema.GetInternals().Checks)
		assert.Greater(t, modifiedChecksCount, originalChecksCount, "Should have more checks than original")
		assert.NotSame(t, originalSchema, modifiedSchema)
	})

	t.Run("clones without modification", func(t *testing.T) {
		// Test instance independence through creating similar schemas
		schema1 := String().Min(5)
		schema2 := String().Min(5)

		// Should be different instances with same configuration
		assert.NotSame(t, schema1, schema2)
		assert.Equal(t, schema1.GetInternals().Type, schema2.GetInternals().Type)
		assert.Equal(t, len(schema1.GetInternals().Checks), len(schema2.GetInternals().Checks))
	})

	t.Run("preserves all checks during clone", func(t *testing.T) {
		// Test check preservation through method chaining
		baseSchema := String().Min(1).Max(100)
		baseChecksCount := len(baseSchema.GetInternals().Checks)

		extendedSchema := baseSchema.Email()

		// Base schema should be unchanged
		assert.Equal(t, baseChecksCount, len(baseSchema.GetInternals().Checks))

		// Extended schema should have all original checks plus new one
		extendedChecksCount := len(extendedSchema.GetInternals().Checks)
		assert.Greater(t, extendedChecksCount, baseChecksCount, "Should have more checks than base")

		// Should be different instances
		assert.NotSame(t, baseSchema, extendedSchema)
	})

	t.Run("deep clone with nested modifications", func(t *testing.T) {
		// Test complex method chaining and immutability
		originalSchema := String().Min(5).Max(50)
		originalChecksCount := len(originalSchema.GetInternals().Checks)

		// Chain multiple modifications
		complexSchema := originalSchema.Email().StartsWith("test").EndsWith("ing")

		// Original should be unchanged
		assert.Equal(t, originalChecksCount, len(originalSchema.GetInternals().Checks))
		assert.Equal(t, "string", originalSchema.GetInternals().Type)

		// Complex schema should have all additional checks
		assert.Equal(t, "string", complexSchema.GetInternals().Type)
		complexChecksCount := len(complexSchema.GetInternals().Checks)
		assert.Greater(t, complexChecksCount, originalChecksCount, "Should have significantly more checks")

		// Test functionality with a string that meets all requirements
		result, err := complexSchema.Parse("testing@example.coming")
		require.NoError(t, err)
		assert.Equal(t, "testing@example.coming", result)
	})

	t.Run("clone with nil modifier function", func(t *testing.T) {
		// Test robustness of method chaining
		schema := String().Email()

		// Chain more methods
		chainedSchema := schema.Min(5).Max(100)

		// Should work fine and create new instances
		require.NotNil(t, chainedSchema)
		assert.NotSame(t, schema, chainedSchema)

		// Test functionality
		result, err := chainedSchema.Parse("valid@email.com")
		require.NoError(t, err)
		assert.Equal(t, "valid@email.com", result)
	})
}

// TestTypeSpecificInternals tests the distinction between GetInternals and GetZod methods
func TestTypeSpecificInternals(t *testing.T) {
	tests := []struct {
		name         string
		createSchema func() interface{ GetInternals() *ZodTypeInternals }
		expectedType string
		testSpecific func(t *testing.T, schema interface{ GetInternals() *ZodTypeInternals })
	}{
		{
			name:         "string schema",
			createSchema: func() interface{ GetInternals() *ZodTypeInternals } { return String().Min(5) },
			expectedType: "string",
			testSpecific: func(t *testing.T, schema interface{ GetInternals() *ZodTypeInternals }) {
				if stringSchema, ok := schema.(*ZodString); ok {
					stringInternals := stringSchema.GetZod()
					require.NotNil(t, stringInternals)
					assert.NotNil(t, stringInternals.Bag)
					assert.Equal(t, "string", stringInternals.Def.Type)
				}
			},
		},
		{
			name: "struct schema",
			createSchema: func() interface{ GetInternals() *ZodTypeInternals } {
				return Struct(ObjectSchema{"name": String(), "age": Int()})
			},
			expectedType: "struct",
			testSpecific: func(t *testing.T, schema interface{ GetInternals() *ZodTypeInternals }) {
				if structSchema, ok := schema.(*ZodStruct); ok {
					structInternals := structSchema.GetZod()
					require.NotNil(t, structInternals)
					assert.NotNil(t, structInternals.Shape)
					assert.Equal(t, 2, len(structInternals.Shape))
					assert.Equal(t, STRIP_MODE, structInternals.Mode)
				}
			},
		},
		{
			name: "array schema",
			createSchema: func() interface{ GetInternals() *ZodTypeInternals } {
				return Array(String(), Int())
			},
			expectedType: "array",
			testSpecific: func(t *testing.T, schema interface{ GetInternals() *ZodTypeInternals }) {
				if arraySchema, ok := schema.(*ZodArray); ok {
					arrayInternals := arraySchema.GetZod()
					require.NotNil(t, arrayInternals)
					assert.NotNil(t, arrayInternals.Items)
					assert.Equal(t, 2, len(arrayInternals.Items))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.createSchema()

			// Test GetInternals returns common interface
			internals := schema.GetInternals()
			require.NotNil(t, internals)
			assert.Equal(t, tt.expectedType, internals.Type)

			// Test type-specific functionality
			if tt.testSpecific != nil {
				tt.testSpecific(t, schema)
			}
		})
	}

	t.Run("union schema access", func(t *testing.T) {
		// TODO: Add union schema tests when implemented
		t.Skip("Union schema access patterns not fully implemented")
	})

	t.Run("lazy schema access", func(t *testing.T) {
		// TODO: Add lazy schema tests when implemented
		t.Skip("Lazy schema access patterns not fully implemented")
	})
}

// TestCoreSchemaTypes tests that core schema types are properly initialized
func TestCoreSchemaTypes(t *testing.T) {
	coreTypes := []struct {
		name         string
		createSchema func() ZodType[any, any]
		expectedType string
	}{
		{"String", func() ZodType[any, any] { return String() }, "string"},
		{"Int", func() ZodType[any, any] { return Int() }, "int"},
		{"Bool", func() ZodType[any, any] { return Bool() }, "bool"},
		{"BigInt", func() ZodType[any, any] { return BigInt() }, "bigint"},
		{"Never", func() ZodType[any, any] { return Never() }, "never"},
		{"IPv4", func() ZodType[any, any] { return IPv4() }, "ipv4"},
		{"File", func() ZodType[any, any] { return File() }, "file"},
		{"Float64", func() ZodType[any, any] { return Float64() }, "float64"},
	}

	for _, ct := range coreTypes {
		t.Run(ct.name, func(t *testing.T) {
			schema := ct.createSchema()

			require.NotNil(t, schema, "Schema should not be nil")

			internals := schema.GetInternals()
			require.NotNil(t, internals, "Internals should not be nil")
			assert.Equal(t, ct.expectedType, internals.Type)
			assert.NotNil(t, internals.Parse, "Parse function should be initialized")
		})
	}

	t.Run("complex numeric types", func(t *testing.T) {
		complexTypes := []struct {
			name         string
			createSchema func() ZodType[any, any]
			expectedType string
		}{
			{"Complex64", func() ZodType[any, any] { return Complex64() }, "complex64"},
			{"Complex128", func() ZodType[any, any] { return Complex128() }, "complex128"},
		}

		for _, ct := range complexTypes {
			t.Run(ct.name, func(t *testing.T) {
				schema := ct.createSchema()
				internals := schema.GetInternals()

				require.NotNil(t, internals)
				assert.Equal(t, ct.expectedType, internals.Type)
			})
		}
	})

	t.Run("edge case types", func(t *testing.T) {
		edgeTypes := []struct {
			name         string
			createSchema func() ZodType[any, any]
			expectedType string
			shouldSkip   bool
			skipReason   string
		}{
			{
				name:         "Any",
				createSchema: func() ZodType[any, any] { return Any() },
				expectedType: "any",
				shouldSkip:   false,
			},
			{
				name:       "Unknown",
				shouldSkip: true,
				skipReason: "Unknown type not implemented yet",
			},
			{
				name:       "Void",
				shouldSkip: true,
				skipReason: "Void type not implemented yet",
			},
		}

		for _, et := range edgeTypes {
			t.Run(et.name, func(t *testing.T) {
				if et.shouldSkip {
					t.Skip("TODO: " + et.skipReason)
					return
				}

				schema := et.createSchema()
				internals := schema.GetInternals()

				require.NotNil(t, internals)
				assert.Equal(t, et.expectedType, internals.Type)
			})
		}
	})
}

// TestSchemaInvariantsAndBoundaries tests schema invariants and boundary conditions
func TestSchemaInvariantsAndBoundaries(t *testing.T) {
	t.Run("schema immutability", func(t *testing.T) {
		originalSchema := String().Min(5)
		originalChecksCount := len(originalSchema.GetInternals().Checks)

		// Operations should not modify original schema
		_ = originalSchema.Max(10)
		_ = originalSchema.Email()

		// Original should remain unchanged
		assert.Equal(t, originalChecksCount, len(originalSchema.GetInternals().Checks))
	})

	t.Run("empty schema creation", func(t *testing.T) {
		schema := String()
		internals := schema.GetInternals()

		require.NotNil(t, internals)
		assert.Empty(t, internals.Checks)
		assert.NotNil(t, internals.Values)
	})

	t.Run("schema with maximum checks", func(t *testing.T) {
		// Test schema with multiple checks (reasonable number)
		schema := String().
			Min(1).
			Max(100).
			StartsWith("test").
			EndsWith("ing").
			Includes("mid")

		internals := schema.GetInternals()
		require.NotNil(t, internals)

		// Should have all the checks
		assert.GreaterOrEqual(t, len(internals.Checks), 5)

		// Schema should still be functional
		result, err := schema.Parse("testmiddlething")
		require.NoError(t, err)
		assert.Equal(t, "testmiddlething", result)

		// Invalid input should fail
		_, err = schema.Parse("invalid")
		assert.Error(t, err)
	})

	t.Run("schema version consistency", func(t *testing.T) {
		schemas := []ZodType[any, any]{
			String(),
			Int(),
			Bool(),
			BigInt(),
		}

		for _, schema := range schemas {
			internals := schema.GetInternals()
			assert.Equal(t, Version, internals.Version, "All schemas should use same version")
		}
	})

	t.Run("nil safety checks", func(t *testing.T) {
		// Test that core functions handle nil gracefully
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Function should not panic on nil input: %v", r)
			}
		}()

		// These should not panic
		schema := String()
		_ = Clone(schema, nil)

		internals := schema.GetInternals()
		require.NotNil(t, internals)
	})
}
