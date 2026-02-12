package types

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnion_BasicFunctionality(t *testing.T) {
	t.Run("valid inputs", func(t *testing.T) {
		schema := Union([]any{String(), Int()})

		got, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", got)

		got, err = schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, got)

		_, err = schema.Parse(true)
		assert.Error(t, err)
	})

	t.Run("multiple types", func(t *testing.T) {
		schema := Union([]any{String(), Int(), Bool()})

		tests := []struct {
			name  string
			input any
			want  any
			ok    bool
		}{
			{"string", "test", "test", true},
			{"int", 123, 123, true},
			{"bool", true, true, true},
			{"float rejected", 3.14, nil, false},
			{"slice rejected", []string{}, nil, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := schema.Parse(tt.input)
				if tt.ok {
					require.NoError(t, err)
					assert.Equal(t, tt.want, got)
				} else {
					assert.Error(t, err)
				}
			})
		}
	})
	t.Run("Parse and MustParse", func(t *testing.T) {
		schema := Union([]any{String(), Bool()})

		got, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", got)

		got = schema.MustParse(true)
		assert.Equal(t, true, got)

		assert.Panics(t, func() { schema.MustParse(123) })
	})

	t.Run("custom error message", func(t *testing.T) {
		schema := Union([]any{String(), Bool()},
			core.SchemaParams{Error: "Expected string or boolean"})

		require.NotNil(t, schema)
		assert.Equal(t, core.ZodTypeUnion, schema.internals.Def.Type)

		_, err := schema.Parse(123)
		assert.Error(t, err)
	})

	t.Run("empty union", func(t *testing.T) {
		schema := Union([]any{})
		_, err := schema.Parse("anything")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no union options provided")
	})
}

func TestUnion_TypeSafety(t *testing.T) {
	t.Run("preserves input type", func(t *testing.T) {
		schema := Union([]any{String(), Int()})

		got, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", got)
		assert.IsType(t, "", got)

		got, err = schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, got)
		assert.IsType(t, 0, got)
	})

	t.Run("UnionOf variadic", func(t *testing.T) {
		schema := UnionOf(String(), Int(), Bool())

		for _, input := range []any{"hello", 123, true} {
			got, err := schema.Parse(input)
			require.NoError(t, err)
			assert.Equal(t, input, got)
		}
	})
	t.Run("Options returns members", func(t *testing.T) {
		schema := Union([]any{String(), Int()})
		assert.Len(t, schema.Options(), 2)
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		schema := Union([]any{String(), Bool()})

		got := schema.MustParse("test")
		assert.IsType(t, "", got)
		assert.Equal(t, "test", got)

		got = schema.MustParse(false)
		assert.IsType(t, false, got)
		assert.Equal(t, false, got)
	})
}

func TestUnion_Modifiers(t *testing.T) {
	t.Run("Optional", func(t *testing.T) {
		schema := Union([]any{String(), Int()}).Optional()

		got, err := schema.Parse("hello")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "hello", *got)

		got, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("Nilable", func(t *testing.T) {
		schema := Union([]any{String(), Int()}).Nilable()

		got, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, got)

		got, err = schema.Parse("test")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "test", *got)
	})

	t.Run("Nullish", func(t *testing.T) {
		schema := Union([]any{String(), Int()}).Nullish()

		got, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, got)

		got, err = schema.Parse(42)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, 42, *got)
	})
	t.Run("Default preserves type", func(t *testing.T) {
		schema := Union([]any{String(), Int()}).Default("default")

		got, err := schema.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, 123, got)
		assert.IsType(t, 0, got)
	})
}

func TestUnion_Chaining(t *testing.T) {
	t.Run("Default then Optional", func(t *testing.T) {
		schema := Union([]any{String(), Int()}).
			Default("fallback").
			Optional()

		got, err := schema.Parse(42)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, 42, *got)

		// nil triggers Default short-circuit
		got, err = schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "fallback", *got)
	})

	t.Run("Default and Prefault", func(t *testing.T) {
		schema := Union([]any{String(), Int()}).
			Default("default").
			Prefault("prefault")

		got, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", got)
	})

	t.Run("modifier immutability", func(t *testing.T) {
		original := Union([]any{String(), Int()})
		modified := original.Optional()

		_, err := original.Parse(nil)
		assert.Error(t, err, "original should reject nil")

		got, err := modified.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
func TestUnion_DefaultAndPrefault(t *testing.T) {
	t.Run("Default over Prefault", func(t *testing.T) {
		schema := Union([]any{String(), Int()}).Default("default_value").Prefault("prefault_value")

		got, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "default_value", got)
	})

	t.Run("Default short-circuits", func(t *testing.T) {
		// Default bypasses validation — even invalid types are returned.
		schema := Union([]any{String(), Int()}).Default([]string{"invalid", "type"})

		got, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, []string{"invalid", "type"}, got)
	})

	t.Run("Prefault validates", func(t *testing.T) {
		schema := Union([]any{String(), Int()}).Prefault("valid_string")

		got, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "valid_string", got)
	})

	t.Run("Prefault only on nil", func(t *testing.T) {
		schema := Union([]any{String(), Int()}).Prefault("prefault_fallback")

		_, err := schema.Parse([]string{"invalid", "type"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid input")
	})

	t.Run("DefaultFunc and PrefaultFunc", func(t *testing.T) {
		defaultCalled := false
		prefaultCalled := false

		schema := Union([]any{String(), Int()}).DefaultFunc(func() any {
			defaultCalled = true
			return "default_func"
		}).PrefaultFunc(func() any {
			prefaultCalled = true
			return "prefault_func"
		})

		got, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, "default_func", got)
		assert.True(t, defaultCalled)
		assert.False(t, prefaultCalled, "PrefaultFunc should not be called when Default is present")
	})

	t.Run("valid Prefault value", func(t *testing.T) {
		schema := Union([]any{String(), Int()}).Prefault(42)

		got, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Equal(t, 42, got)
	})
}
func TestUnion_Refine(t *testing.T) {
	t.Run("type-based validation", func(t *testing.T) {
		schema := Union([]any{String(), Int()}).Refine(func(v any) bool {
			switch val := v.(type) {
			case string:
				return len(val) > 3
			case int:
				return val > 0
			default:
				return false
			}
		})

		got, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", got)

		got, err = schema.Parse(42)
		require.NoError(t, err)
		assert.Equal(t, 42, got)

		_, err = schema.Parse("hi")
		assert.Error(t, err)

		_, err = schema.Parse(-5)
		assert.Error(t, err)
	})

	t.Run("custom error message", func(t *testing.T) {
		schema := Union([]any{String(), Int()}).Refine(func(v any) bool {
			switch val := v.(type) {
			case string:
				return len(val) > 0
			case int:
				return val > 0
			default:
				return false
			}
		}, core.SchemaParams{Error: "Must be non-empty string or positive number"})

		got, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", got)

		_, err = schema.Parse("")
		assert.Error(t, err)

		_, err = schema.Parse(-1)
		assert.Error(t, err)
	})
}

func TestUnion_RefineAny(t *testing.T) {
	t.Run("raw value validation", func(t *testing.T) {
		schema := Union([]any{String(), Bool()}).RefineAny(func(v any) bool {
			return v == "valid" || v == true
		})

		got, err := schema.Parse("valid")
		require.NoError(t, err)
		assert.Equal(t, "valid", got)

		got, err = schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, got)

		_, err = schema.Parse("invalid")
		assert.Error(t, err)

		_, err = schema.Parse(false)
		assert.Error(t, err)
	})
}
func TestUnion_TypeSpecificMethods(t *testing.T) {
	t.Run("Options preserves order", func(t *testing.T) {
		schema := Union([]any{String(), Int(), Bool()})
		opts := schema.Options()
		assert.Len(t, opts, 3)

		_, err := opts[0].ParseAny("test")
		assert.NoError(t, err, "first option should accept strings")

		_, err = opts[1].ParseAny(123)
		assert.NoError(t, err, "second option should accept ints")

		_, err = opts[2].ParseAny(true)
		assert.NoError(t, err, "third option should accept bools")
	})

	t.Run("UnionOf constructor", func(t *testing.T) {
		schema := UnionOf(String(), Int(), Bool())
		assert.Len(t, schema.Options(), 3)

		for _, input := range []any{"hello", 42, true} {
			got, err := schema.Parse(input)
			require.NoError(t, err)
			assert.Equal(t, input, got)
		}
	})
}

func TestUnion_ErrorHandling(t *testing.T) {
	t.Run("all members fail", func(t *testing.T) {
		schema := Union([]any{String(), Bool()})
		_, err := schema.Parse(123.45)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no union member matched")
	})

	t.Run("invalid member panics", func(t *testing.T) {
		assert.Panics(t, func() {
			Union([]any{String(), 123}) // 123 is not a schema
		})
	})
}
func TestUnion_EdgeCases(t *testing.T) {
	t.Run("with discriminated union", func(t *testing.T) {
		dog := Object(core.ObjectSchema{
			"type": Literal("dog"),
			"bark": String(),
		})
		cat := Object(core.ObjectSchema{
			"type": Literal("cat"),
			"meow": String(),
		})

		schema := Union([]any{
			DiscriminatedUnion("type", []any{dog, cat}),
			Int(),
		})

		got, err := schema.Parse(map[string]any{"type": "dog", "bark": "woof"})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"type": "dog", "bark": "woof"}, got)

		got, err = schema.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, 123, got)
	})

	t.Run("nested unions", func(t *testing.T) {
		schema := UnionOf(
			UnionOf(String(), Int()),
			UnionOf(Bool(), Float()),
		)

		got, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", got)

		got, err = schema.Parse(3.14)
		require.NoError(t, err)
		assert.Equal(t, 3.14, got)
	})

	t.Run("coerced member type preference", func(t *testing.T) {
		schema := Union([]any{CoercedString(), Int()})

		// Int takes precedence for int input.
		got, err := schema.Parse(42)
		require.NoError(t, err)
		assert.IsType(t, 0, got)
	})

	t.Run("nil schema member", func(t *testing.T) {
		schema := Union([]any{String(), nil, Int()})
		_, err := schema.Parse(true)
		assert.Error(t, err)
	})
}
func TestUnion_NonOptional(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		schema := UnionOf(String(), Int()).Optional().NonOptional()

		got, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", got)
		assert.IsType(t, "", got)

		got, err = schema.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, 123, got)
		assert.IsType(t, 0, got)

		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("chained toggles", func(t *testing.T) {
		schema := UnionOf(String(), Bool()).Optional().NonOptional().Optional().NonOptional()

		got, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, got)

		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("on non-optional schema", func(t *testing.T) {
		schema := UnionOf(Float(), String()).NonOptional()

		got, err := schema.Parse(3.14)
		require.NoError(t, err)
		assert.Equal(t, 3.14, got)

		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})
}

// Xor tests

func TestZodXor_ExactlyOneMatch(t *testing.T) {
	schema := Xor([]any{String(), Int()})

	got, err := schema.Parse("hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", got)

	got, err = schema.Parse(42)
	require.NoError(t, err)
	assert.Equal(t, 42, got)
}

func TestZodXor_ZeroMatchesFails(t *testing.T) {
	schema := Xor([]any{String(), Int()})
	_, err := schema.Parse(true)
	assert.Error(t, err)
}
func TestZodXor_MultipleMatchesFails(t *testing.T) {
	// String() and Unknown() both match strings — should fail.
	schema := Xor([]any{String(), Unknown()})
	_, err := schema.Parse("hello")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid input")
}

func TestZodXor_WithCustomErrorMessage(t *testing.T) {
	schema := Xor([]any{String(), Int()}, "Expected exactly one of string or number")
	_, err := schema.Parse(true)
	assert.Error(t, err)
}

func TestZodXor_XorOf_VariadicSyntax(t *testing.T) {
	schema := XorOf(String(), Int(), Bool())

	for _, input := range []any{"hello", 42, true} {
		got, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, got)
	}
}

func TestZodXor_StrictParse(t *testing.T) {
	schema := Xor([]any{String(), Int()})
	got, err := schema.StrictParse("hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", got)
}

func TestZodXor_MustParse_Success(t *testing.T) {
	schema := Xor([]any{String(), Int()})
	got := schema.MustParse("hello")
	assert.Equal(t, "hello", got)
}

func TestZodXor_MustParse_Panic(t *testing.T) {
	schema := Xor([]any{String(), Int()})
	assert.Panics(t, func() { schema.MustParse(true) })
}

func TestZodXor_ParseAny(t *testing.T) {
	schema := Xor([]any{String(), Int()})
	got, err := schema.ParseAny("hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", got)
}
func TestZodXor_EnhancedCoverage(t *testing.T) {
	t.Run("complex object schemas", func(t *testing.T) {
		dog := Object(core.ObjectSchema{
			"type": Literal("dog"),
			"bark": String(),
		})
		cat := Object(core.ObjectSchema{
			"type": Literal("cat"),
			"meow": String(),
		})

		schema := Xor([]any{dog, cat})

		got, err := schema.Parse(map[string]any{"type": "dog", "bark": "woof"})
		require.NoError(t, err)
		assert.Equal(t, "dog", got.(map[string]any)["type"])

		got, err = schema.Parse(map[string]any{"type": "cat", "meow": "meow"})
		require.NoError(t, err)
		assert.Equal(t, "cat", got.(map[string]any)["type"])
	})

	t.Run("multiple matches error", func(t *testing.T) {
		schema := Xor([]any{String(), Unknown()})
		_, err := schema.Parse("hello")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid input")
	})

	t.Run("MustParse", func(t *testing.T) {
		schema := Xor([]any{String(), Int()})

		got := schema.MustParse("hello")
		assert.Equal(t, "hello", got)

		assert.Panics(t, func() { schema.MustParse(true) })
	})

	t.Run("MustStrictParse", func(t *testing.T) {
		schema := Xor([]any{String(), Int()})

		got := schema.MustStrictParse("hello")
		assert.Equal(t, "hello", got)

		assert.Panics(t, func() { schema.MustStrictParse(true) })
	})

	t.Run("immutability", func(t *testing.T) {
		s1 := Xor([]any{String(), Int()})
		s2 := Xor([]any{Bool(), Float64()})
		got, err := s1.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", got)

		got, err = s2.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, got)

		_, err = s1.Parse(true)
		require.Error(t, err)

		_, err = s2.Parse("hello")
		require.Error(t, err)
	})

	t.Run("nested xor", func(t *testing.T) {
		inner := XorOf(String(), Bool())
		outer := XorOf(inner, Int())

		for _, input := range []any{"hello", true, 42} {
			got, err := outer.Parse(input)
			require.NoError(t, err)
			assert.Equal(t, input, got)
		}
	})

	t.Run("Options returns members", func(t *testing.T) {
		schema := Xor([]any{String(), Int(), Bool()})
		opts := schema.Options()
		assert.Len(t, opts, 3)

		_, err := opts[0].ParseAny("test")
		assert.NoError(t, err)

		_, err = opts[1].ParseAny(123)
		assert.NoError(t, err)

		_, err = opts[2].ParseAny(true)
		assert.NoError(t, err)
	})
}
