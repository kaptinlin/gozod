package types_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/issues"
	. "github.com/kaptinlin/gozod/types"
)

type countingDiscriminatorSchema struct {
	internals core.ZodTypeInternals
	calls     int
}

func newCountingDiscriminatorSchema(value any) *countingDiscriminatorSchema {
	return &countingDiscriminatorSchema{
		internals: core.ZodTypeInternals{
			Values: map[any]struct{}{value: {}},
		},
	}
}

func (s *countingDiscriminatorSchema) ParseAny(input any, _ ...*core.ParseContext) (any, error) {
	s.calls++
	return input, nil
}

func (s *countingDiscriminatorSchema) Internals() *core.ZodTypeInternals { return &s.internals }
func (s *countingDiscriminatorSchema) IsOptional() bool                  { return false }
func (s *countingDiscriminatorSchema) IsNilable() bool                   { return false }

// =============================================================================
// Basic functionality tests
// =============================================================================

func TestDiscriminatedUnion_BasicFunctionality(t *testing.T) {
	t.Run("valid discriminated union inputs", func(t *testing.T) {
		// Create discriminated union with object schemas
		userSchema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"user"}),
			"name": String(),
		})
		adminSchema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"admin"}),
			"name": String(),
			"role": String(),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{userSchema, adminSchema})

		// Test user type
		result, err := discriminatedUnion.Parse(map[string]any{
			"type": "user",
			"name": "John",
		})
		require.NoError(t, err)
		expected := map[string]any{
			"type": "user",
			"name": "John",
		}
		assert.Equal(t, expected, result)

		// Test admin type
		result, err = discriminatedUnion.Parse(map[string]any{
			"type": "admin",
			"name": "Jane",
			"role": "superuser",
		})
		require.NoError(t, err)
		expected = map[string]any{
			"type": "admin",
			"name": "Jane",
			"role": "superuser",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("invalid discriminated union inputs", func(t *testing.T) {
		userSchema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"user"}),
			"name": String(),
		})
		adminSchema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"admin"}),
			"name": String(),
			"role": String(),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{userSchema, adminSchema})

		// Test invalid discriminator value
		_, err := discriminatedUnion.Parse(map[string]any{
			"type": "guest",
			"name": "Unknown",
		})
		assert.Error(t, err)

		// Test missing discriminator field
		_, err = discriminatedUnion.Parse(map[string]any{
			"name": "John",
		})
		assert.Error(t, err)

		// Test non-object input
		_, err = discriminatedUnion.Parse("not an object")
		assert.Error(t, err)
	})

	t.Run("Parse and MustParse methods", func(t *testing.T) {
		successSchema := Object(core.ObjectSchema{
			"status": LiteralOf([]string{"success"}),
			"data":   String(),
		})
		errorSchema := Object(core.ObjectSchema{
			"status": LiteralOf([]string{"error"}),
			"error":  String(),
		})

		discriminatedUnion := MustDiscriminatedUnion("status", []core.ZodSchema{successSchema, errorSchema})

		// Test Parse method
		result, err := discriminatedUnion.Parse(map[string]any{
			"status": "success",
			"data":   "operation completed",
		})
		require.NoError(t, err)
		expected := map[string]any{
			"status": "success",
			"data":   "operation completed",
		}
		assert.Equal(t, expected, result)

		// Test MustParse method
		mustResult := discriminatedUnion.MustParse(map[string]any{
			"status": "error",
			"error":  "operation failed",
		})
		expected = map[string]any{
			"status": "error",
			"error":  "operation failed",
		}
		assert.Equal(t, expected, mustResult)

		// Test panic on invalid input
		assert.Panics(t, func() {
			discriminatedUnion.MustParse("invalid")
		})
	})

	t.Run("custom error message", func(t *testing.T) {
		customError := "Expected discriminated union match"
		userSchema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"user"}),
			"name": String(),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{userSchema}, core.SchemaParams{Error: customError})

		require.NotNil(t, discriminatedUnion)

		_, err := discriminatedUnion.Parse(map[string]any{
			"type": "invalid",
			"name": "test",
		})
		assert.Error(t, err)
	})
}

// =============================================================================
// Type safety tests
// =============================================================================

func TestDiscriminatedUnion_TypeSafety(t *testing.T) {
	t.Run("discriminated union returns correct type", func(t *testing.T) {
		userSchema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"user"}),
			"id":   Int(),
		})
		postSchema := Object(core.ObjectSchema{
			"type":  LiteralOf([]string{"post"}),
			"title": String(),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{userSchema, postSchema})
		require.NotNil(t, discriminatedUnion)

		result, err := discriminatedUnion.Parse(map[string]any{
			"type": "user",
			"id":   123,
		})
		require.NoError(t, err)
		assert.IsType(t, map[string]any{}, result)

		resultMap := result.(map[string]any)
		assert.Equal(t, "user", resultMap["type"])
		assert.Equal(t, 123, resultMap["id"])
	})

	t.Run("discriminator field access", func(t *testing.T) {
		schema1 := Object(core.ObjectSchema{
			"kind": LiteralOf([]string{"a"}),
			"val":  String(),
		})
		schema2 := Object(core.ObjectSchema{
			"kind": LiteralOf([]string{"b"}),
			"val":  Int(),
		})

		discriminatedUnion := MustDiscriminatedUnion("kind", []core.ZodSchema{schema1, schema2})

		// Test discriminator accessor
		assert.Equal(t, "kind", discriminatedUnion.Discriminator())

		// Test options accessor
		options := discriminatedUnion.Options()
		assert.Len(t, options, 2)
		assert.Equal(t, schema1, options[0])
		assert.Equal(t, schema2, options[1])
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": Bool(),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema})

		result := discriminatedUnion.MustParse(map[string]any{
			"type": "test",
			"data": true,
		})
		assert.IsType(t, map[string]any{}, result)

		resultMap := result.(map[string]any)
		assert.Equal(t, "test", resultMap["type"])
		assert.Equal(t, true, resultMap["data"])
	})
}

// =============================================================================
// Modifier methods tests
// =============================================================================

func TestDiscriminatedUnion_Modifiers(t *testing.T) {
	t.Run("Optional makes discriminated union optional", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String(),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema})
		optionalDiscriminatedUnion := discriminatedUnion.Optional()

		// Test non-nil value - should return pointer
		result, err := optionalDiscriminatedUnion.Parse(map[string]any{
			"type": "test",
			"data": "hello",
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		expected := map[string]any{
			"type": "test",
			"data": "hello",
		}
		assert.Equal(t, expected, *result)

		// Test nil value (should be handled by Optional)
		result, err = optionalDiscriminatedUnion.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable allows nil values", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String(),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema})
		nilableDiscriminatedUnion := discriminatedUnion.Nilable()

		// Test nil handling
		result, err := nilableDiscriminatedUnion.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid value - should return pointer
		result, err = nilableDiscriminatedUnion.Parse(map[string]any{
			"type": "test",
			"data": "hello",
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		expected := map[string]any{
			"type": "test",
			"data": "hello",
		}
		assert.Equal(t, expected, *result)
	})

	t.Run("Default preserves discriminated union type", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String(),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema})
		defaultValue := map[string]any{
			"type": "test",
			"data": "default",
		}
		defaultDiscriminatedUnion := discriminatedUnion.Default(defaultValue)

		// Valid input should override default
		result, err := defaultDiscriminatedUnion.Parse(map[string]any{
			"type": "test",
			"data": "custom",
		})
		require.NoError(t, err)
		expected := map[string]any{
			"type": "test",
			"data": "custom",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Prefault preserves discriminated union type", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String(),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema})
		prefaultValue := map[string]any{
			"type": "test",
			"data": "prefault",
		}
		prefaultDiscriminatedUnion := discriminatedUnion.Prefault(prefaultValue)

		// Valid input should override prefault
		result, err := prefaultDiscriminatedUnion.Parse(map[string]any{
			"type": "test",
			"data": "custom",
		})
		require.NoError(t, err)
		expected := map[string]any{
			"type": "test",
			"data": "custom",
		}
		assert.Equal(t, expected, result)
	})
}

// =============================================================================
// Chaining tests
// =============================================================================

func TestDiscriminatedUnion_Chaining(t *testing.T) {
	t.Run("discriminated union chaining", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String(),
		})

		defaultValue := map[string]any{
			"type": "test",
			"data": "default",
		}

		chainedSchema := MustDiscriminatedUnion("type", []core.ZodSchema{schema}).
			Default(defaultValue).
			Optional()

		// Test final behavior - should return pointer due to Optional
		result, err := chainedSchema.Parse(map[string]any{
			"type": "test",
			"data": "custom",
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		expected := map[string]any{
			"type": "test",
			"data": "custom",
		}
		assert.Equal(t, expected, *result)
	})

	t.Run("complex chaining", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"flag": Bool(),
		})

		chainedSchema := MustDiscriminatedUnion("type", []core.ZodSchema{schema}).
			Nilable().
			Default(map[string]any{
				"type": "test",
				"flag": true,
			})

		result, err := chainedSchema.Parse(map[string]any{
			"type": "test",
			"flag": false,
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		expected := map[string]any{
			"type": "test",
			"flag": false,
		}
		assert.Equal(t, expected, *result)
	})

	t.Run("default and prefault chaining", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String(),
		})

		chainedSchema := MustDiscriminatedUnion("type", []core.ZodSchema{schema}).
			Default(map[string]any{
				"type": "test",
				"data": "default",
			}).
			Prefault(map[string]any{
				"type": "test",
				"data": "prefault",
			})

		result, err := chainedSchema.Parse(map[string]any{
			"type": "test",
			"data": "custom",
		})
		require.NoError(t, err)
		expected := map[string]any{
			"type": "test",
			"data": "custom",
		}
		assert.Equal(t, expected, result)
	})
}

// =============================================================================
// Default and prefault tests
// =============================================================================

func TestDiscriminatedUnion_DefaultAndPrefault(t *testing.T) {
	t.Run("outer Default over Optional follows modifier order", func(t *testing.T) {
		option := Object(core.ObjectSchema{
			"type": Literal("test"),
		})
		defaultValue := map[string]any{"type": "test"}

		parse := func(t *testing.T, strict bool) {
			t.Helper()
			schema := MustDiscriminatedUnion("type", []core.ZodSchema{option}).
				Optional().
				Default(defaultValue)

			var (
				got *any
				err error
			)
			if strict {
				got, err = schema.StrictParse(nil)
			} else {
				got, err = schema.Parse(nil)
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, defaultValue, (*got).(map[string]any))
		}

		t.Run("Parse", func(t *testing.T) { parse(t, false) })
		t.Run("StrictParse", func(t *testing.T) { parse(t, true) })
	})

	t.Run("outer Optional over Default follows modifier order", func(t *testing.T) {
		option := Object(core.ObjectSchema{
			"type": Literal("test"),
		})
		defaultValue := map[string]any{"type": "test"}
		schema := MustDiscriminatedUnion("type", []core.ZodSchema{option}).
			Default(defaultValue).
			Optional()

		t.Run("Parse", func(t *testing.T) {
			got, err := schema.Parse(nil)
			require.NoError(t, err)
			assert.Nil(t, got)
		})
		t.Run("StrictParse", func(t *testing.T) {
			got, err := schema.StrictParse(nil)
			require.NoError(t, err)
			assert.Nil(t, got)
		})
	})

	t.Run("outer Default over inner Prefault", func(t *testing.T) {
		// When both Default and Prefault are set, the outer Default should handle nil first.
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String(),
		})

		defaultValue := map[string]any{
			"type": "test",
			"data": "default",
		}
		prefaultValue := map[string]any{
			"type": "test",
			"data": "prefault",
		}

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema}).
			Prefault(prefaultValue).
			Default(defaultValue)

		t.Run("Parse", func(t *testing.T) {
			result, err := discriminatedUnion.Parse(nil)
			require.NoError(t, err)
			assert.Equal(t, defaultValue, result)
		})
		t.Run("StrictParse", func(t *testing.T) {
			result, err := discriminatedUnion.StrictParse(nil)
			require.NoError(t, err)
			assert.Equal(t, defaultValue, result)
		})
	})

	t.Run("outer Prefault over inner Default", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String().Min(5),
		})
		defaultValue := map[string]any{
			"type": "test",
			"data": "default",
		}
		prefaultValue := map[string]any{
			"type": "test",
			"data": "prefault",
		}
		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema}).
			Default(defaultValue).
			Prefault(prefaultValue)

		t.Run("Parse", func(t *testing.T) {
			result, err := discriminatedUnion.Parse(nil)
			require.NoError(t, err)
			assert.Equal(t, prefaultValue, result)
		})
		t.Run("StrictParse", func(t *testing.T) {
			result, err := discriminatedUnion.StrictParse(nil)
			require.NoError(t, err)
			assert.Equal(t, prefaultValue, result)
		})
	})

	t.Run("Default short-circuits validation", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String().Min(10),
		})
		defaultValue := map[string]any{
			"type": "test",
			"data": "short",
		}
		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema}).Default(defaultValue)

		t.Run("Parse", func(t *testing.T) {
			result, err := discriminatedUnion.Parse(nil)
			require.NoError(t, err)
			assert.Equal(t, defaultValue, result)
		})
		t.Run("StrictParse", func(t *testing.T) {
			result, err := discriminatedUnion.StrictParse(nil)
			require.NoError(t, err)
			assert.Equal(t, defaultValue, result)
		})
	})

	t.Run("Prefault goes through full validation", func(t *testing.T) {
		// Prefault value must pass all discriminated union validation
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String().Min(5), // Require at least 5 characters
		})

		prefaultValue := map[string]any{
			"type": "test",
			"data": "valid", // 5 characters, meets requirement
		}

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema}).Prefault(prefaultValue)

		// Nil input triggers prefault, goes through validation and succeeds
		result, err := discriminatedUnion.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, prefaultValue, result)

		// Non-nil input that fails validation should not trigger prefault
		_, err = discriminatedUnion.Parse(map[string]any{
			"type": "test",
			"data": "hi", // Only 2 characters, fails validation
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Too small: expected string to have at least 5 characters")
	})

	t.Run("Prefault only triggers for nil input", func(t *testing.T) {
		// Non-nil input that fails validation should not trigger Prefault
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String(),
		})

		prefaultValue := map[string]any{
			"type": "test",
			"data": "prefault",
		}
		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema}).Prefault(prefaultValue)

		valid := map[string]any{
			"type": "test",
			"data": "custom",
		}
		for _, strict := range []bool{false, true} {
			name := "Parse"
			if strict {
				name = "StrictParse"
			}
			t.Run(name, func(t *testing.T) {
				var (
					result any
					err    error
				)
				if strict {
					result, err = discriminatedUnion.StrictParse("invalid-union")
				} else {
					result, err = discriminatedUnion.Parse("invalid-union")
				}
				require.Error(t, err)
				assert.Nil(t, result)
				var zodErr *issues.ZodError
				require.True(t, issues.IsZodError(err, &zodErr))
				require.Len(t, zodErr.Issues, 1)
				assert.Equal(t, core.InvalidType, zodErr.Issues[0].Code)
				assert.Equal(t, core.ZodTypeObject, zodErr.Issues[0].Expected)

				if strict {
					result, err = discriminatedUnion.StrictParse(valid)
				} else {
					result, err = discriminatedUnion.Parse(valid)
				}
				require.NoError(t, err)
				assert.Equal(t, valid, result)
			})
		}
	})

	t.Run("DefaultFunc and PrefaultFunc behavior", func(t *testing.T) {
		// Test function call behavior and priority
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String(),
		})

		defaultCalled := false
		prefaultCalled := false

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema}).
			PrefaultFunc(func() any {
				prefaultCalled = true
				return map[string]any{
					"type": "test",
					"data": "prefault",
				}
			}).
			DefaultFunc(func() any {
				defaultCalled = true
				return map[string]any{
					"type": "test",
					"data": "default",
				}
			})

		result, err := discriminatedUnion.Parse(nil)
		require.NoError(t, err)
		expected := map[string]any{
			"type": "test",
			"data": "default",
		}
		assert.Equal(t, expected, result)
		assert.True(t, defaultCalled, "DefaultFunc should be called")
		assert.False(t, prefaultCalled, "PrefaultFunc should not be called when Default is present")
	})

	t.Run("Prefault validation failure", func(t *testing.T) {
		// When Prefault value fails validation, should return error
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String().Min(10), // Require at least 10 characters
		})

		prefaultValue := map[string]any{
			"type": "test",
			"data": "short", // Only 5 characters, fails validation
		}

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema}).Prefault(prefaultValue)

		_, err := discriminatedUnion.Parse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Too small: expected string to have at least 10 characters")
	})
}

// =============================================================================
// Refine tests
// =============================================================================

func TestDiscriminatedUnion_Refine(t *testing.T) {
	t.Run("refine validation", func(t *testing.T) {
		userSchema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"user"}),
			"name": String(),
		})
		adminSchema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"admin"}),
			"name": String(),
			"role": String(),
		})

		// Only accept entries where name is not empty
		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{userSchema, adminSchema}).Refine(func(v any) bool {
			if obj, ok := v.(map[string]any); ok {
				if name, exists := obj["name"]; exists {
					if nameStr, ok := name.(string); ok {
						return len(nameStr) > 0
					}
				}
			}
			return false
		})

		result, err := discriminatedUnion.Parse(map[string]any{
			"type": "user",
			"name": "John",
		})
		require.NoError(t, err)
		expected := map[string]any{
			"type": "user",
			"name": "John",
		}
		assert.Equal(t, expected, result)

		_, err = discriminatedUnion.Parse(map[string]any{
			"type": "user",
			"name": "",
		})
		assert.Error(t, err)
	})

	t.Run("refine with custom error message", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String(),
		})

		errorMessage := "Must be a valid discriminated union"
		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema}).Refine(func(v any) bool {
			if obj, ok := v.(map[string]any); ok {
				if data, exists := obj["data"]; exists {
					if dataStr, ok := data.(string); ok {
						return len(dataStr) > 3
					}
				}
			}
			return false
		}, core.SchemaParams{Error: errorMessage})

		result, err := discriminatedUnion.Parse(map[string]any{
			"type": "test",
			"data": "hello",
		})
		require.NoError(t, err)
		expected := map[string]any{
			"type": "test",
			"data": "hello",
		}
		assert.Equal(t, expected, result)

		_, err = discriminatedUnion.Parse(map[string]any{
			"type": "test",
			"data": "hi",
		})
		assert.Error(t, err)
	})

	t.Run("refine with complex discriminated union", func(t *testing.T) {
		userSchema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"user"}),
			"name": String().Min(3),
		})
		adminSchema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"admin"}),
			"name": String().Max(10),
			"role": String(),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{userSchema, adminSchema}).Refine(func(v any) bool {
			if obj, ok := v.(map[string]any); ok {
				if objType, exists := obj["type"]; exists {
					return objType != "forbidden"
				}
			}
			return false
		})

		// Should pass all validations
		result, err := discriminatedUnion.Parse(map[string]any{
			"type": "user",
			"name": "John",
		})
		require.NoError(t, err)
		expected := map[string]any{
			"type": "user",
			"name": "John",
		}
		assert.Equal(t, expected, result)

		// Should fail refine validation
		_, err = discriminatedUnion.Parse(map[string]any{
			"type": "forbidden",
			"name": "test",
		})
		assert.Error(t, err)

		// Should fail schema validation (name too short for user)
		_, err = discriminatedUnion.Parse(map[string]any{
			"type": "user",
			"name": "Jo",
		})
		assert.Error(t, err)
	})
}

func TestDiscriminatedUnion_RefineAny(t *testing.T) {
	t.Run("refineAny flexible validation", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String(),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema}).RefineAny(func(v any) bool {
			obj, ok := v.(map[string]any)
			if !ok {
				return false
			}
			data, exists := obj["data"]
			if !exists {
				return false
			}
			dataStr, ok := data.(string)
			return ok && len(dataStr) >= 4
		})

		// String with >= 4 chars should pass
		result, err := discriminatedUnion.Parse(map[string]any{
			"type": "test",
			"data": "hello",
		})
		require.NoError(t, err)
		expected := map[string]any{
			"type": "test",
			"data": "hello",
		}
		assert.Equal(t, expected, result)

		// String with < 4 chars should fail
		_, err = discriminatedUnion.Parse(map[string]any{
			"type": "test",
			"data": "hi",
		})
		assert.Error(t, err)
	})

	t.Run("refineAny with type checking", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"num":  Int(),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema}).RefineAny(func(v any) bool {
			obj, ok := v.(map[string]any)
			if !ok {
				return false
			}
			num, exists := obj["num"]
			if !exists {
				return false
			}
			numInt, ok := num.(int)
			return ok && numInt%2 == 0 // Only even numbers
		})

		result, err := discriminatedUnion.Parse(map[string]any{
			"type": "test",
			"num":  4,
		})
		require.NoError(t, err)
		expected := map[string]any{
			"type": "test",
			"num":  4,
		}
		assert.Equal(t, expected, result)

		_, err = discriminatedUnion.Parse(map[string]any{
			"type": "test",
			"num":  3,
		})
		assert.Error(t, err)
	})
}

func TestDiscriminatedUnion_ParentRefinementRunsOnceOnSelectedOutput(t *testing.T) {
	for _, strict := range []bool{false, true} {
		name := "Parse"
		if strict {
			name = "StrictParse"
		}
		t.Run(name, func(t *testing.T) {
			calls := 0
			option := Object(core.ObjectSchema{
				"type": Literal("test"),
				"data": String().Default("parsed"),
			})
			schema := MustDiscriminatedUnion("type", []core.ZodSchema{option}).RefineAny(func(value any) bool {
				calls++
				object, ok := value.(map[string]any)
				return ok && object["data"] == "parsed"
			})
			input := map[string]any{"type": "test"}

			var (
				result any
				err    error
			)
			if strict {
				result, err = schema.StrictParse(input)
			} else {
				result, err = schema.Parse(input)
			}

			require.NoError(t, err)
			assert.Equal(t, map[string]any{"type": "test", "data": "parsed"}, result)
			assert.Equal(t, 1, calls)
		})
	}
}

// =============================================================================
// Type-specific methods tests
// =============================================================================

func TestDiscriminatedUnion_TypeSpecificMethods(t *testing.T) {
	t.Run("Discriminator returns discriminator field name", func(t *testing.T) {
		schema1 := Object(core.ObjectSchema{
			"status": LiteralOf([]string{"success"}),
			"data":   String(),
		})
		schema2 := Object(core.ObjectSchema{
			"status": LiteralOf([]string{"error"}),
			"error":  String(),
		})

		discriminatedUnion := MustDiscriminatedUnion("status", []core.ZodSchema{schema1, schema2})

		assert.Equal(t, "status", discriminatedUnion.Discriminator())
	})

	t.Run("Options returns all schemas", func(t *testing.T) {
		schema1 := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"a"}),
			"val":  String(),
		})
		schema2 := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"b"}),
			"val":  Int(),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema1, schema2})

		options := discriminatedUnion.Options()
		assert.Len(t, options, 2)
		assert.Equal(t, schema1, options[0])
		assert.Equal(t, schema2, options[1])

		options[0] = nil
		assert.Equal(t, schema1, discriminatedUnion.Options()[0])
	})

	t.Run("DiscriminatorMap returns value mapping", func(t *testing.T) {
		schema1 := Object(core.ObjectSchema{
			"kind": LiteralOf([]string{"first"}),
			"data": String(),
		})
		schema2 := Object(core.ObjectSchema{
			"kind": LiteralOf([]string{"second"}),
			"data": Int(),
		})

		discriminatedUnion := MustDiscriminatedUnion("kind", []core.ZodSchema{schema1, schema2})

		discMap := discriminatedUnion.DiscriminatorMap()
		assert.Len(t, discMap, 2)
		assert.Contains(t, discMap, "first")
		assert.Contains(t, discMap, "second")
		assert.Equal(t, schema1, discMap["first"])
		assert.Equal(t, schema2, discMap["second"])
	})
}

// =============================================================================
// Error handling tests
// =============================================================================

func TestDiscriminatedUnion_ErrorHandling(t *testing.T) {
	t.Run("nil option fails construction", func(t *testing.T) {
		valid := Object(core.ObjectSchema{
			"type": Literal("valid"),
		})

		schema, err := DiscriminatedUnion("type", []core.ZodSchema{valid, nil})

		assert.Nil(t, schema)
		require.ErrorIs(t, err, ErrOptionIsNil)
	})

	t.Run("duplicate discriminator returns structured construction error", func(t *testing.T) {
		first := Object(core.ObjectSchema{"type": Literal("duplicate")})
		second := Object(core.ObjectSchema{"type": Literal("duplicate")})

		schema, err := DiscriminatedUnion("type", []core.ZodSchema{first, second})

		assert.Nil(t, schema)
		require.ErrorIs(t, err, ErrDuplicateDiscriminator)
		detail, ok := errors.AsType[*DiscriminatorError](err)
		require.True(t, ok)
		assert.Equal(t, 1, detail.Option)
		assert.Equal(t, "type", detail.Field)
		assert.Equal(t, "duplicate", detail.Value)
	})

	t.Run("missing discriminator capability fails construction", func(t *testing.T) {
		option := Object(core.ObjectSchema{"type": String()})

		schema, err := DiscriminatedUnion("type", []core.ZodSchema{option})

		assert.Nil(t, schema)
		require.ErrorIs(t, err, ErrNoDiscriminatorValues)
		detail, ok := errors.AsType[*DiscriminatorError](err)
		require.True(t, ok)
		assert.Equal(t, 0, detail.Option)
		assert.Equal(t, "type", detail.Field)
	})

	t.Run("empty discriminator values fail construction", func(t *testing.T) {
		option := Object(core.ObjectSchema{"type": LiteralOf([]string{})})

		schema, err := DiscriminatedUnion("type", []core.ZodSchema{option})

		assert.Nil(t, schema)
		require.ErrorIs(t, err, ErrNoDiscriminatorValues)
	})

	t.Run("empty options fail construction", func(t *testing.T) {
		schema, err := DiscriminatedUnion("type", nil)

		assert.Nil(t, schema)
		require.ErrorIs(t, err, ErrNoValidDiscriminators)
	})

	t.Run("Must constructor panics on invalid definition", func(t *testing.T) {
		assert.Panics(t, func() {
			MustDiscriminatedUnion("type", nil)
		})
	})

	t.Run("missing discriminator field", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String(),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema})

		_, err := discriminatedUnion.Parse(map[string]any{
			"data": "hello",
		})
		assert.Error(t, err)
	})

	t.Run("invalid discriminator value", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String(),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema})

		_, err := discriminatedUnion.Parse(map[string]any{
			"type": "invalid",
			"data": "hello",
		})
		assert.Error(t, err)
	})

	t.Run("schema validation error", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String().Min(10),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema})

		_, err := discriminatedUnion.Parse(map[string]any{
			"type": "test",
			"data": "short",
		})
		assert.Error(t, err)
	})

	t.Run("non-object input", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String(),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema})

		// This should fail for non-object input
		_, err := discriminatedUnion.Parse("not an object")
		assert.Error(t, err)

		_, err = discriminatedUnion.Parse(123)
		assert.Error(t, err)

		_, err = discriminatedUnion.Parse([]string{"array"})
		assert.Error(t, err)
	})

	t.Run("custom error message", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String(),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema}, core.SchemaParams{Error: "Expected discriminated union match"})

		_, err := discriminatedUnion.Parse(map[string]any{
			"type": "invalid",
			"data": "hello",
		})
		assert.Error(t, err)
	})
}

func TestDiscriminatedUnion_UnknownDiscriminatorDoesNotRunOptions(t *testing.T) {
	first := newCountingDiscriminatorSchema("first")
	second := newCountingDiscriminatorSchema("second")
	schema := MustDiscriminatedUnion("type", []core.ZodSchema{first, second})

	_, err := schema.Parse(map[string]any{"type": "unknown"})

	require.Error(t, err)
	assert.Equal(t, 0, first.calls)
	assert.Equal(t, 0, second.calls)

	var zodErr *issues.ZodError
	require.True(t, issues.IsZodError(err, &zodErr))
	require.Len(t, zodErr.Issues, 1)
	assert.Equal(t, core.InvalidUnion, zodErr.Issues[0].Code)
	assert.Equal(t, []any{"type"}, zodErr.Issues[0].Path)
	assert.Equal(t, []any{"first", "second"}, zodErr.Issues[0].Values)
}

func TestDiscriminatedUnion_KnownDiscriminatorRunsOnlyIndexedOption(t *testing.T) {
	for _, strict := range []bool{false, true} {
		name := "Parse"
		if strict {
			name = "StrictParse"
		}
		t.Run(name, func(t *testing.T) {
			first := newCountingDiscriminatorSchema("first")
			second := newCountingDiscriminatorSchema("second")
			schema := MustDiscriminatedUnion("type", []core.ZodSchema{first, second})
			input := map[string]any{"type": "second"}

			var (
				result any
				err    error
			)
			if strict {
				result, err = schema.StrictParse(input)
			} else {
				result, err = schema.Parse(input)
			}

			require.NoError(t, err)
			assert.Equal(t, input, result)
			assert.Equal(t, 0, first.calls)
			assert.Equal(t, 1, second.calls)
		})
	}
}

// =============================================================================
// Edge case tests
// =============================================================================

func TestDiscriminatedUnion_EdgeCases(t *testing.T) {
	t.Run("nil handling with discriminated union", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String(),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema}).Nilable()

		// Test nil input
		result, err := discriminatedUnion.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		// Test valid value - should return pointer
		result, err = discriminatedUnion.Parse(map[string]any{
			"type": "test",
			"data": "hello",
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		expected := map[string]any{
			"type": "test",
			"data": "hello",
		}
		assert.Equal(t, expected, *result)
	})

	t.Run("single schema discriminated union", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"only"}),
			"data": String(),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema})

		result, err := discriminatedUnion.Parse(map[string]any{
			"type": "only",
			"data": "test",
		})
		require.NoError(t, err)
		expected := map[string]any{
			"type": "only",
			"data": "test",
		}
		assert.Equal(t, expected, result)

		_, err = discriminatedUnion.Parse(map[string]any{
			"type": "other",
			"data": "test",
		})
		assert.Error(t, err)
	})

	t.Run("discriminated union strict validation", func(t *testing.T) {
		schema1 := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"a"}),
			"data": String(),
		})
		schema2 := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"b"}),
			"data": Int(),
		})

		// Create discriminated union
		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema1, schema2})

		// Test normal discriminated union behavior
		result, err := discriminatedUnion.Parse(map[string]any{
			"type": "a",
			"data": "hello",
		})
		require.NoError(t, err)
		expected := map[string]any{
			"type": "a",
			"data": "hello",
		}
		assert.Equal(t, expected, result)

		// Test strict validation - missing discriminator should fail
		_, err = discriminatedUnion.Parse(map[string]any{
			"data": "hello",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Missing required discriminator field: type")
	})

	t.Run("empty context", func(t *testing.T) {
		schema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"test"}),
			"data": String(),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{schema})

		// Parse with empty context slice
		result, err := discriminatedUnion.Parse(map[string]any{
			"type": "test",
			"data": "hello",
		})
		require.NoError(t, err)
		expected := map[string]any{
			"type": "test",
			"data": "hello",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("complex discriminator values", func(t *testing.T) {
		// Test with different types of discriminator values
		stringSchema := Object(core.ObjectSchema{
			"type": LiteralOf([]string{"string"}),
			"data": String(),
		})
		intSchema := Object(core.ObjectSchema{
			"type": LiteralOf([]int{42}),
			"data": Int(),
		})
		boolSchema := Object(core.ObjectSchema{
			"type": LiteralOf([]bool{true}),
			"data": Bool(),
		})

		discriminatedUnion := MustDiscriminatedUnion("type", []core.ZodSchema{stringSchema, intSchema, boolSchema})

		// Test string discriminator
		result, err := discriminatedUnion.Parse(map[string]any{
			"type": "string",
			"data": "hello",
		})
		require.NoError(t, err)
		expected := map[string]any{
			"type": "string",
			"data": "hello",
		}
		assert.Equal(t, expected, result)

		// Test int discriminator
		result, err = discriminatedUnion.Parse(map[string]any{
			"type": 42,
			"data": 123,
		})
		require.NoError(t, err)
		expected = map[string]any{
			"type": 42,
			"data": 123,
		}
		assert.Equal(t, expected, result)

		// Test bool discriminator
		result, err = discriminatedUnion.Parse(map[string]any{
			"type": true,
			"data": false,
		})
		require.NoError(t, err)
		expected = map[string]any{
			"type": true,
			"data": false,
		}
		assert.Equal(t, expected, result)
	})
}
