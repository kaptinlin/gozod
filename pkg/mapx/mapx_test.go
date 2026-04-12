package mapx_test

import (
	"testing"

	"github.com/kaptinlin/gozod/pkg/mapx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValueOf(t *testing.T) {
	m := map[string]any{"name": "Alice", "age": 25, "active": true}

	t.Run("matching type", func(t *testing.T) {
		v, ok := mapx.ValueOf[string](m, "name")
		assert.Equal(t, "Alice", v)
		assert.True(t, ok)
	})

	t.Run("wrong type", func(t *testing.T) {
		v, ok := mapx.ValueOf[string](m, "age")
		assert.Empty(t, v)
		assert.False(t, ok)
	})

	t.Run("missing key", func(t *testing.T) {
		v, ok := mapx.ValueOf[string](m, "missing")
		assert.Empty(t, v)
		assert.False(t, ok)
	})

	t.Run("nil map", func(t *testing.T) {
		v, ok := mapx.ValueOf[string](nil, "key")
		assert.Empty(t, v)
		assert.False(t, ok)
	})
}

func TestValueOr(t *testing.T) {
	m := map[string]any{"name": "Alice", "age": 25}

	t.Run("matching type", func(t *testing.T) {
		assert.Equal(t, "Alice", mapx.ValueOr(m, "name", "default"))
	})

	t.Run("wrong type", func(t *testing.T) {
		assert.Equal(t, "default", mapx.ValueOr(m, "age", "default"))
	})

	t.Run("missing key", func(t *testing.T) {
		assert.Equal(t, "default", mapx.ValueOr(m, "missing", "default"))
	})

	t.Run("nil map", func(t *testing.T) {
		assert.Equal(t, 42, mapx.ValueOr(nil, "key", 42))
	})
}

func TestGet(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		val, ok := mapx.Get(nil, "key")
		assert.Nil(t, val)
		assert.False(t, ok)
	})

	t.Run("existing key", func(t *testing.T) {
		m := map[string]any{"name": "Alice"}
		val, ok := mapx.Get(m, "name")
		assert.Equal(t, "Alice", val)
		assert.True(t, ok)
	})

	t.Run("non-existing key", func(t *testing.T) {
		m := map[string]any{"name": "Alice"}
		val, ok := mapx.Get(m, "age")
		assert.Nil(t, val)
		assert.False(t, ok)
	})

	t.Run("nil value", func(t *testing.T) {
		m := map[string]any{"name": nil}
		val, ok := mapx.Get(m, "name")
		assert.Nil(t, val)
		assert.True(t, ok)
	})
}

func TestSet(t *testing.T) {
	t.Run("nil map does nothing", func(t *testing.T) {
		var m map[string]any
		mapx.Set(m, "key", "value") // Should not panic
	})

	t.Run("adds value", func(t *testing.T) {
		m := make(map[string]any)
		mapx.Set(m, "name", "Alice")
		assert.Equal(t, "Alice", m["name"])
	})

	t.Run("overwrites", func(t *testing.T) {
		m := map[string]any{"name": "Alice"}
		mapx.Set(m, "name", "Bob")
		assert.Equal(t, "Bob", m["name"])
	})
}

func TestHas(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		assert.False(t, mapx.Has(nil, "key"))
	})

	t.Run("existing key", func(t *testing.T) {
		m := map[string]any{"name": "Alice"}
		assert.True(t, mapx.Has(m, "name"))
	})

	t.Run("non-existing key", func(t *testing.T) {
		m := map[string]any{"name": "Alice"}
		assert.False(t, mapx.Has(m, "age"))
	})

	t.Run("nil value key", func(t *testing.T) {
		m := map[string]any{"name": nil}
		assert.True(t, mapx.Has(m, "name"))
	})
}

func TestCount(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		assert.Equal(t, 0, mapx.Count(nil))
	})

	t.Run("empty map", func(t *testing.T) {
		assert.Equal(t, 0, mapx.Count(map[string]any{}))
	})

	t.Run("with elements", func(t *testing.T) {
		m := map[string]any{"a": 1, "b": 2, "c": 3}
		assert.Equal(t, 3, mapx.Count(m))
	})
}

func TestCopy(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		assert.Nil(t, mapx.Copy(nil))
	})

	t.Run("empty map", func(t *testing.T) {
		got := mapx.Copy(map[string]any{})
		require.NotNil(t, got)
		assert.Empty(t, got)
	})

	t.Run("independent from original", func(t *testing.T) {
		original := map[string]any{"a": 1, "b": 2}
		copied := mapx.Copy(original)

		copied["a"] = 100
		copied["c"] = 3

		assert.Equal(t, 1, original["a"])
		assert.NotContains(t, original, "c")
	})
}

func TestMerge(t *testing.T) {
	t.Run("both nil", func(t *testing.T) {
		assert.Nil(t, mapx.Merge(nil, nil))
	})

	t.Run("first nil", func(t *testing.T) {
		second := map[string]any{"a": 1}
		got := mapx.Merge(nil, second)
		assert.Equal(t, 1, got["a"])

		got["a"] = 100
		assert.Equal(t, 1, second["a"])
	})

	t.Run("second nil", func(t *testing.T) {
		first := map[string]any{"a": 1}
		got := mapx.Merge(first, nil)
		assert.Equal(t, 1, got["a"])
	})

	t.Run("second takes precedence", func(t *testing.T) {
		first := map[string]any{"a": 1, "b": 2}
		second := map[string]any{"b": 20, "c": 3}
		got := mapx.Merge(first, second)

		assert.Equal(t, 1, got["a"])
		assert.Equal(t, 20, got["b"])
		assert.Equal(t, 3, got["c"])
	})
}

func TestString(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		val, ok := mapx.String(nil, "key")
		assert.Empty(t, val)
		assert.False(t, ok)
	})

	t.Run("string value", func(t *testing.T) {
		m := map[string]any{"name": "Alice"}
		val, ok := mapx.String(m, "name")
		assert.Equal(t, "Alice", val)
		assert.True(t, ok)
	})

	t.Run("non-string value", func(t *testing.T) {
		m := map[string]any{"age": 25}
		val, ok := mapx.String(m, "age")
		assert.Empty(t, val)
		assert.False(t, ok)
	})
}

func TestBool(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		val, ok := mapx.Bool(nil, "key")
		assert.False(t, val)
		assert.False(t, ok)
	})

	t.Run("bool value", func(t *testing.T) {
		m := map[string]any{"active": true}
		val, ok := mapx.Bool(m, "active")
		assert.True(t, val)
		assert.True(t, ok)
	})

	t.Run("non-bool value", func(t *testing.T) {
		m := map[string]any{"active": "yes"}
		val, ok := mapx.Bool(m, "active")
		assert.False(t, val)
		assert.False(t, ok)
	})
}

func TestInt(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		val, ok := mapx.IntCoerce(nil, "key")
		assert.Equal(t, 0, val)
		assert.False(t, ok)
	})

	t.Run("int value", func(t *testing.T) {
		m := map[string]any{"age": 25}
		val, ok := mapx.IntCoerce(m, "age")
		assert.Equal(t, 25, val)
		assert.True(t, ok)
	})

	t.Run("int32 value", func(t *testing.T) {
		m := map[string]any{"age": int32(25)}
		val, ok := mapx.IntCoerce(m, "age")
		assert.Equal(t, 25, val)
		assert.True(t, ok)
	})

	t.Run("int64 value", func(t *testing.T) {
		m := map[string]any{"age": int64(25)}
		val, ok := mapx.IntCoerce(m, "age")
		assert.Equal(t, 25, val)
		assert.True(t, ok)
	})

	t.Run("float64 value", func(t *testing.T) {
		m := map[string]any{"age": float64(25)}
		val, ok := mapx.IntCoerce(m, "age")
		assert.Equal(t, 25, val)
		assert.True(t, ok)
	})

	t.Run("string value", func(t *testing.T) {
		m := map[string]any{"age": "25"}
		val, ok := mapx.IntCoerce(m, "age")
		assert.Equal(t, 0, val)
		assert.False(t, ok)
	})
}

func TestFloat64(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		val, ok := mapx.Float64Coerce(nil, "key")
		assert.Equal(t, 0.0, val)
		assert.False(t, ok)
	})

	t.Run("float64 value", func(t *testing.T) {
		m := map[string]any{"price": 19.99}
		val, ok := mapx.Float64Coerce(m, "price")
		assert.Equal(t, 19.99, val)
		assert.True(t, ok)
	})

	t.Run("int value", func(t *testing.T) {
		m := map[string]any{"price": 20}
		val, ok := mapx.Float64Coerce(m, "price")
		assert.Equal(t, 20.0, val)
		assert.True(t, ok)
	})
}

func TestNumericCoerceCompatibilityAliases(t *testing.T) {
	m := map[string]any{
		"age":   int32(25),
		"price": 20,
	}

	intVal, intOK := mapx.Int(m, "age")
	assert.Equal(t, 25, intVal)
	assert.True(t, intOK)

	floatVal, floatOK := mapx.Float64(m, "price")
	assert.Equal(t, 20.0, floatVal)
	assert.True(t, floatOK)
}

func TestMap(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		val, ok := mapx.Map(nil, "key")
		assert.Nil(t, val)
		assert.False(t, ok)
	})

	t.Run("map value", func(t *testing.T) {
		nested := map[string]any{"nested": true}
		m := map[string]any{"config": nested}
		val, ok := mapx.Map(m, "config")
		require.True(t, ok)
		assert.Equal(t, true, val["nested"])
	})

	t.Run("non-map value", func(t *testing.T) {
		m := map[string]any{"config": "string"}
		_, ok := mapx.Map(m, "config")
		assert.False(t, ok)
	})
}

func TestStringOr(t *testing.T) {
	m := map[string]any{"name": "Alice"}

	t.Run("existing key", func(t *testing.T) {
		assert.Equal(t, "Alice", mapx.StringOr(m, "name", "default"))
	})

	t.Run("missing key", func(t *testing.T) {
		assert.Equal(t, "default", mapx.StringOr(m, "missing", "default"))
	})
}

func TestBoolOr(t *testing.T) {
	m := map[string]any{"active": true}

	t.Run("existing key", func(t *testing.T) {
		assert.True(t, mapx.BoolOr(m, "active", false))
	})

	t.Run("missing key", func(t *testing.T) {
		assert.True(t, mapx.BoolOr(m, "missing", true))
	})
}

func TestIntOr(t *testing.T) {
	m := map[string]any{"count": 10}

	t.Run("existing key", func(t *testing.T) {
		assert.Equal(t, 10, mapx.IntCoerceOr(m, "count", 0))
	})

	t.Run("missing key", func(t *testing.T) {
		assert.Equal(t, 42, mapx.IntCoerceOr(m, "missing", 42))
	})
}

func TestNumericCoerceDefaultCompatibilityAliases(t *testing.T) {
	m := map[string]any{
		"count": int64(10),
		"ratio": float32(2.5),
	}

	assert.Equal(t, 10, mapx.IntOr(m, "count", 0))
	assert.Equal(t, 10, mapx.IntCoerceOr(m, "count", 0))
	assert.Equal(t, 2.5, mapx.FloatOr(m, "ratio", 0))
	assert.Equal(t, 2.5, mapx.Float64CoerceOr(m, "ratio", 0))
}

func TestKeys(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		assert.Nil(t, mapx.Keys(nil))
	})

	t.Run("map[string]any", func(t *testing.T) {
		m := map[string]any{"a": 1, "b": 2}
		got := mapx.Keys(m)
		assert.Len(t, got, 2)
	})

	t.Run("map[any]any with string keys", func(t *testing.T) {
		m := map[any]any{"a": 1, "b": 2, 3: "ignored"}
		got := mapx.Keys(m)
		assert.Len(t, got, 2)
	})
}

func TestToGeneric(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		got, err := mapx.ToGeneric(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("map[any]any", func(t *testing.T) {
		m := map[any]any{"a": 1}
		got, err := mapx.ToGeneric(m)
		require.NoError(t, err)
		assert.Equal(t, 1, got["a"])
	})

	t.Run("map[string]any", func(t *testing.T) {
		m := map[string]any{"a": 1}
		got, err := mapx.ToGeneric(m)
		require.NoError(t, err)
		assert.Equal(t, 1, got["a"])
	})

	t.Run("non-map returns error", func(t *testing.T) {
		_, err := mapx.ToGeneric("not a map")
		assert.ErrorIs(t, err, mapx.ErrInputNotMap)
	})
}
