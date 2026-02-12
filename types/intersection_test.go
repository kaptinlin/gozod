package types

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntersection_BasicFunctionality(t *testing.T) {
	t.Run("valid inputs", func(t *testing.T) {
		schema := Intersection(String(), String())
		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("conflicting constraints", func(t *testing.T) {
		schema := Intersection(String().Min(5), String().Max(3))
		_, err := schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("Parse and MustParse", func(t *testing.T) {
		schema := Intersection(Bool(), Bool())

		result, err := schema.Parse(true)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		assert.Equal(t, false, schema.MustParse(false))

		assert.Panics(t, func() {
			schema.MustParse("invalid")
		})
	})

	t.Run("custom error message", func(t *testing.T) {
		schema := Intersection(String(), String(), core.SchemaParams{Error: "Expected intersection match"})
		require.NotNil(t, schema)
		_, err := schema.Parse(123)
		assert.Error(t, err)
	})
}

func TestIntersection_TypeSafety(t *testing.T) {
	t.Run("returns merged type", func(t *testing.T) {
		schema := Intersection(String(), String())
		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)
		assert.IsType(t, "", result)
	})

	t.Run("incompatible types", func(t *testing.T) {
		schema := Intersection(String(), Int())
		_, err := schema.Parse("test")
		assert.Error(t, err)
		_, err = schema.Parse(123)
		assert.Error(t, err)
	})

	t.Run("MustParse type safety", func(t *testing.T) {
		schema := Intersection(Bool(), Bool())
		result := schema.MustParse(true)
		assert.IsType(t, true, result)
		assert.Equal(t, true, result)
	})
}

func TestIntersection_Modifiers(t *testing.T) {
	t.Run("Optional", func(t *testing.T) {
		schema := Intersection(String(), String()).Optional()

		result, err := schema.Parse("test")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "test", *result)

		result, err = schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Nilable", func(t *testing.T) {
		schema := Intersection(String(), String()).Nilable()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.Nil(t, result)

		result, err = schema.Parse("test")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "test", *result)
	})

	t.Run("Default", func(t *testing.T) {
		schema := Intersection(String(), String()).Default("default")
		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})

	t.Run("Prefault", func(t *testing.T) {
		schema := Intersection(String(), String()).Prefault("prefault")
		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})
}

func TestIntersection_Chaining(t *testing.T) {
	t.Run("default then optional", func(t *testing.T) {
		schema := Intersection(String(), String()).Default("default").Optional()
		result, err := schema.Parse("test")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "test", *result)
	})

	t.Run("nilable then default", func(t *testing.T) {
		schema := Intersection(Bool(), Bool()).Nilable().Default(true)
		result, err := schema.Parse(false)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, false, *result)
	})

	t.Run("default and prefault", func(t *testing.T) {
		schema := Intersection(String(), String()).Default("default").Prefault("prefault")
		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)
	})
}

func TestIntersection_DefaultAndPrefault(t *testing.T) {
	t.Run("Default has higher priority than Prefault", func(t *testing.T) {
		schema := Intersection(String(), String()).Default("default-value").Prefault("prefault-value").Optional()
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "default-value", *result)
	})

	t.Run("Default bypasses validation", func(t *testing.T) {
		schema := Intersection(String(), String()).Default(123).Optional()
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 123, *result)
	})

	t.Run("Prefault goes through full validation", func(t *testing.T) {
		schema := Intersection(String(), String()).Prefault("valid-string").Optional()
		result, err := schema.Parse(nil)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "valid-string", *result)
	})

	t.Run("Prefault only triggers on nil input", func(t *testing.T) {
		schema := Intersection(String(), String()).Prefault("prefault-value")
		_, err := schema.Parse(123)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid input: expected string")
	})

	t.Run("DefaultFunc and PrefaultFunc", func(t *testing.T) {
		defaultCalled := false
		prefaultCalled := false
		schema := Intersection(String(), String()).
			DefaultFunc(func() any {
				defaultCalled = true
				return "default-func"
			}).
			PrefaultFunc(func() any {
				prefaultCalled = true
				return "prefault-func"
			}).
			Optional()

		result, err := schema.Parse(nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "default-func", *result)
		assert.True(t, defaultCalled, "DefaultFunc should be called")
		assert.False(t, prefaultCalled, "PrefaultFunc should not be called when Default is present")
	})

	t.Run("Prefault validation failure", func(t *testing.T) {
		schema := Intersection(String(), String()).Prefault(123).Optional()
		result, err := schema.Parse(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected string, received number")
		var zero *any
		assert.Equal(t, zero, result)
	})
}

func TestIntersection_Refine(t *testing.T) {
	t.Run("refine validation", func(t *testing.T) {
		schema := Intersection(String().Min(3), String().Max(10)).Refine(func(v any) bool {
			s, ok := v.(string)
			return ok && len(s) > 3 && s != "test"
		})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("refine with custom error", func(t *testing.T) {
		schema := Intersection(String(), String()).Refine(func(v any) bool {
			s, ok := v.(string)
			return ok && len(s) > 3
		}, core.SchemaParams{Error: "Must be a valid intersection"})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("refine with complex intersection", func(t *testing.T) {
		schema := Intersection(String().Min(3), String().Max(10)).Refine(func(v any) bool {
			s, ok := v.(string)
			return ok && s != "forbidden"
		})

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse("forbidden")
		assert.Error(t, err)

		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})
}

func TestIntersection_RefineAny(t *testing.T) {
	t.Run("flexible validation", func(t *testing.T) {
		schema := Intersection(String(), String()).RefineAny(func(v any) bool {
			s, ok := v.(string)
			return ok && len(s) >= 4
		})

		result, err := schema.Parse("test")
		require.NoError(t, err)
		assert.Equal(t, "test", result)

		result, err = schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)

		_, err = schema.Parse("hi")
		assert.Error(t, err)
	})

	t.Run("type checking", func(t *testing.T) {
		schema := Intersection(Int(), Int()).RefineAny(func(v any) bool {
			n, ok := v.(int)
			return ok && n%2 == 0
		})

		result, err := schema.Parse(4)
		require.NoError(t, err)
		assert.Equal(t, 4, result)

		_, err = schema.Parse(3)
		assert.Error(t, err)
	})
}

func TestIntersection_TypeSpecificMethods(t *testing.T) {
	t.Run("Left returns left schema", func(t *testing.T) {
		schema := Intersection(String(), Int())
		assert.NotNil(t, schema.Left())
	})

	t.Run("Right returns right schema", func(t *testing.T) {
		schema := Intersection(String(), Int())
		assert.NotNil(t, schema.Right())
	})
}

func TestIntersection_ErrorHandling(t *testing.T) {
	t.Run("left schema error", func(t *testing.T) {
		schema := Intersection(String().Min(10), String())
		_, err := schema.Parse("short")
		assert.Error(t, err)
	})

	t.Run("right schema error", func(t *testing.T) {
		schema := Intersection(String(), String().Min(10))
		_, err := schema.Parse("short")
		assert.Error(t, err)
	})

	t.Run("both schemas error", func(t *testing.T) {
		schema := Intersection(String().Min(10), String().Max(3))
		_, err := schema.Parse("medium")
		assert.Error(t, err)
	})

	t.Run("merge conflict", func(t *testing.T) {
		schema := Intersection(String(), Int())
		_, err := schema.Parse("test")
		assert.Error(t, err)
	})

	t.Run("custom error message", func(t *testing.T) {
		schema := Intersection(String(), String(), core.SchemaParams{Error: "Expected intersection match"})
		_, err := schema.Parse(123)
		assert.Error(t, err)
	})
}

func TestIntersection_EdgeCases(t *testing.T) {
	t.Run("two structs", func(t *testing.T) {
		type structA struct{ A string }
		type structB struct{ B int }
		schema := Intersection(Struct[structA](), Struct[structB]())

		result, err := schema.Parse(map[string]any{"A": "hello", "B": 123})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"A": "hello", "B": 123}, result)

		_, err = schema.Parse(map[string]any{"A": "hello"})
		assert.Error(t, err)
	})

	t.Run("overlapping fields", func(t *testing.T) {
		type structC struct{ Overlap string }
		type structD struct{ Overlap string }
		schema := Intersection(Struct[structC](), Struct[structD]())

		input := map[string]any{"Overlap": "value"}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, input, result)
	})

	t.Run("nil value", func(t *testing.T) {
		schema := Intersection(String(), String())
		_, err := schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("map input on string schema", func(t *testing.T) {
		schema := Intersection(String(), String())
		_, err := schema.Parse(map[string]any{"name": "test", "age": "30"})
		assert.Error(t, err)
	})
}

func TestIntersection_NonOptional(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		schema := Intersection(String().Min(2), String().Max(5)).Optional().NonOptional()

		result, err := schema.Parse("abc")
		require.NoError(t, err)
		assert.Equal(t, "abc", result)
		assert.IsType(t, "", result)

		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("chained optional and non-optional", func(t *testing.T) {
		schema := Intersection(String(), String()).Optional().NonOptional().Optional().NonOptional()

		result, err := schema.Parse("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
		assert.IsType(t, "", result)

		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("on already non-optional", func(t *testing.T) {
		schema := Intersection(Int(), Int()).NonOptional()

		result, err := schema.Parse(123)
		require.NoError(t, err)
		assert.Equal(t, 123, result)
		assert.IsType(t, 0, result)

		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})

	t.Run("with struct intersection", func(t *testing.T) {
		type Name struct {
			Name string `json:"name"`
		}
		type Age struct {
			Age int `json:"age"`
		}
		schema := Intersection(Struct[Name](), Struct[Age]()).Optional().NonOptional()

		input := map[string]any{"name": "test", "age": 30}
		result, err := schema.Parse(input)
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"name": "test", "age": 30}, result)

		_, err = schema.Parse(nil)
		assert.Error(t, err)
	})
}
