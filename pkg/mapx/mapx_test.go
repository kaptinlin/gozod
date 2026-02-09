package mapx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// BASIC PROPERTY OPERATIONS TESTS
// =============================================================================

func TestValueOf(t *testing.T) {
	props := map[string]any{"name": "Alice", "age": 25, "active": true}

	t.Run("matching type returns value true", func(t *testing.T) {
		v, ok := ValueOf[string](props, "name")
		assert.Equal(t, "Alice", v)
		assert.True(t, ok)
	})

	t.Run("wrong type returns zero false", func(t *testing.T) {
		v, ok := ValueOf[string](props, "age")
		assert.Empty(t, v)
		assert.False(t, ok)
	})

	t.Run("missing key returns zero false", func(t *testing.T) {
		v, ok := ValueOf[string](props, "missing")
		assert.Empty(t, v)
		assert.False(t, ok)
	})

	t.Run("nil map returns zero false", func(t *testing.T) {
		v, ok := ValueOf[string](nil, "key")
		assert.Empty(t, v)
		assert.False(t, ok)
	})
}

func TestValueOrDefault(t *testing.T) {
	props := map[string]any{"name": "Alice", "age": 25}

	t.Run("matching type returns value", func(t *testing.T) {
		assert.Equal(t, "Alice", ValueOrDefault(props, "name", "default"))
	})

	t.Run("wrong type returns default", func(t *testing.T) {
		assert.Equal(t, "default", ValueOrDefault(props, "age", "default"))
	})

	t.Run("missing key returns default", func(t *testing.T) {
		assert.Equal(t, "default", ValueOrDefault(props, "missing", "default"))
	})

	t.Run("nil map returns default", func(t *testing.T) {
		assert.Equal(t, 42, ValueOrDefault(nil, "key", 42))
	})
}

func TestGet(t *testing.T) {
	t.Run("nil map returns nil false", func(t *testing.T) {
		val, exists := Get(nil, "key")
		assert.Nil(t, val)
		assert.False(t, exists)
	})

	t.Run("existing key returns value true", func(t *testing.T) {
		props := map[string]any{"name": "Alice"}
		val, exists := Get(props, "name")
		assert.Equal(t, "Alice", val)
		assert.True(t, exists)
	})

	t.Run("non-existing key returns nil false", func(t *testing.T) {
		props := map[string]any{"name": "Alice"}
		val, exists := Get(props, "age")
		assert.Nil(t, val)
		assert.False(t, exists)
	})

	t.Run("nil value returns nil true", func(t *testing.T) {
		props := map[string]any{"name": nil}
		val, exists := Get(props, "name")
		assert.Nil(t, val)
		assert.True(t, exists)
	})
}

func TestSet(t *testing.T) {
	t.Run("set on nil map does nothing", func(t *testing.T) {
		var props map[string]any
		Set(props, "key", "value") // Should not panic
	})

	t.Run("set on existing map adds value", func(t *testing.T) {
		props := make(map[string]any)
		Set(props, "name", "Alice")
		assert.Equal(t, "Alice", props["name"])
	})

	t.Run("set overwrites existing value", func(t *testing.T) {
		props := map[string]any{"name": "Alice"}
		Set(props, "name", "Bob")
		assert.Equal(t, "Bob", props["name"])
	})
}

func TestHas(t *testing.T) {
	t.Run("nil map returns false", func(t *testing.T) {
		assert.False(t, Has(nil, "key"))
	})

	t.Run("existing key returns true", func(t *testing.T) {
		props := map[string]any{"name": "Alice"}
		assert.True(t, Has(props, "name"))
	})

	t.Run("non-existing key returns false", func(t *testing.T) {
		props := map[string]any{"name": "Alice"}
		assert.False(t, Has(props, "age"))
	})

	t.Run("nil value key returns true", func(t *testing.T) {
		props := map[string]any{"name": nil}
		assert.True(t, Has(props, "name"))
	})
}

func TestCount(t *testing.T) {
	t.Run("nil map returns 0", func(t *testing.T) {
		assert.Equal(t, 0, Count(nil))
	})

	t.Run("empty map returns 0", func(t *testing.T) {
		assert.Equal(t, 0, Count(map[string]any{}))
	})

	t.Run("map with elements returns count", func(t *testing.T) {
		props := map[string]any{"a": 1, "b": 2, "c": 3}
		assert.Equal(t, 3, Count(props))
	})
}

func TestCopy(t *testing.T) {
	t.Run("nil map returns nil", func(t *testing.T) {
		assert.Nil(t, Copy(nil))
	})

	t.Run("empty map returns empty map", func(t *testing.T) {
		got := Copy(map[string]any{})
		require.NotNil(t, got)
		assert.Empty(t, got)
	})

	t.Run("copy is independent from original", func(t *testing.T) {
		original := map[string]any{"a": 1, "b": 2}
		copied := Copy(original)

		// Modify copy
		copied["a"] = 100
		copied["c"] = 3

		// Original should be unchanged
		assert.Equal(t, 1, original["a"])
		assert.NotContains(t, original, "c")
	})
}

func TestMerge(t *testing.T) {
	t.Run("both nil returns nil", func(t *testing.T) {
		assert.Nil(t, Merge(nil, nil))
	})

	t.Run("first nil returns copy of second", func(t *testing.T) {
		second := map[string]any{"a": 1}
		got := Merge(nil, second)
		assert.Equal(t, 1, got["a"])

		// Verify it's a copy
		got["a"] = 100
		assert.Equal(t, 1, second["a"])
	})

	t.Run("second nil returns copy of first", func(t *testing.T) {
		first := map[string]any{"a": 1}
		got := Merge(first, nil)
		assert.Equal(t, 1, got["a"])
	})

	t.Run("second takes precedence", func(t *testing.T) {
		first := map[string]any{"a": 1, "b": 2}
		second := map[string]any{"b": 20, "c": 3}
		got := Merge(first, second)

		assert.Equal(t, 1, got["a"])
		assert.Equal(t, 20, got["b"]) // from second
		assert.Equal(t, 3, got["c"])
	})
}

// =============================================================================
// TYPE-SAFE PROPERTY ACCESSORS TESTS
// =============================================================================

func TestGetString(t *testing.T) {
	t.Run("nil map returns empty false", func(t *testing.T) {
		val, ok := GetString(nil, "key")
		assert.Empty(t, val)
		assert.False(t, ok)
	})

	t.Run("string value returns value true", func(t *testing.T) {
		props := map[string]any{"name": "Alice"}
		val, ok := GetString(props, "name")
		assert.Equal(t, "Alice", val)
		assert.True(t, ok)
	})

	t.Run("non-string value returns empty false", func(t *testing.T) {
		props := map[string]any{"age": 25}
		val, ok := GetString(props, "age")
		assert.Empty(t, val)
		assert.False(t, ok)
	})
}

func TestGetBool(t *testing.T) {
	t.Run("nil map returns false false", func(t *testing.T) {
		val, ok := GetBool(nil, "key")
		assert.False(t, val)
		assert.False(t, ok)
	})

	t.Run("bool value returns value true", func(t *testing.T) {
		props := map[string]any{"active": true}
		val, ok := GetBool(props, "active")
		assert.True(t, val)
		assert.True(t, ok)
	})

	t.Run("non-bool value returns false false", func(t *testing.T) {
		props := map[string]any{"active": "yes"}
		val, ok := GetBool(props, "active")
		assert.False(t, val)
		assert.False(t, ok)
	})
}

func TestGetInt(t *testing.T) {
	t.Run("nil map returns 0 false", func(t *testing.T) {
		val, ok := GetInt(nil, "key")
		assert.Equal(t, 0, val)
		assert.False(t, ok)
	})

	t.Run("int value returns value true", func(t *testing.T) {
		props := map[string]any{"age": 25}
		val, ok := GetInt(props, "age")
		assert.Equal(t, 25, val)
		assert.True(t, ok)
	})

	t.Run("int32 value returns converted value true", func(t *testing.T) {
		props := map[string]any{"age": int32(25)}
		val, ok := GetInt(props, "age")
		assert.Equal(t, 25, val)
		assert.True(t, ok)
	})

	t.Run("int64 value returns converted value true", func(t *testing.T) {
		props := map[string]any{"age": int64(25)}
		val, ok := GetInt(props, "age")
		assert.Equal(t, 25, val)
		assert.True(t, ok)
	})

	t.Run("float64 value returns converted value true", func(t *testing.T) {
		props := map[string]any{"age": float64(25)}
		val, ok := GetInt(props, "age")
		assert.Equal(t, 25, val)
		assert.True(t, ok)
	})

	t.Run("string value returns 0 false", func(t *testing.T) {
		props := map[string]any{"age": "25"}
		val, ok := GetInt(props, "age")
		assert.Equal(t, 0, val)
		assert.False(t, ok)
	})
}

func TestGetFloat64(t *testing.T) {
	t.Run("nil map returns 0 false", func(t *testing.T) {
		val, ok := GetFloat64(nil, "key")
		assert.Equal(t, 0.0, val)
		assert.False(t, ok)
	})

	t.Run("float64 value returns value true", func(t *testing.T) {
		props := map[string]any{"price": 19.99}
		val, ok := GetFloat64(props, "price")
		assert.Equal(t, 19.99, val)
		assert.True(t, ok)
	})

	t.Run("int value returns converted value true", func(t *testing.T) {
		props := map[string]any{"price": 20}
		val, ok := GetFloat64(props, "price")
		assert.Equal(t, 20.0, val)
		assert.True(t, ok)
	})
}

func TestGetMap(t *testing.T) {
	t.Run("nil map returns nil false", func(t *testing.T) {
		val, ok := GetMap(nil, "key")
		assert.Nil(t, val)
		assert.False(t, ok)
	})

	t.Run("map value returns value true", func(t *testing.T) {
		nested := map[string]any{"nested": true}
		props := map[string]any{"config": nested}
		val, ok := GetMap(props, "config")
		require.True(t, ok)
		assert.Equal(t, true, val["nested"])
	})

	t.Run("non-map value returns nil false", func(t *testing.T) {
		props := map[string]any{"config": "string"}
		_, ok := GetMap(props, "config")
		assert.False(t, ok)
	})
}

// =============================================================================
// DEFAULT VALUE PROPERTY ACCESSORS TESTS
// =============================================================================

func TestGetStringDefault(t *testing.T) {
	props := map[string]any{"name": "Alice"}

	t.Run("existing key returns value", func(t *testing.T) {
		assert.Equal(t, "Alice", GetStringDefault(props, "name", "default"))
	})

	t.Run("missing key returns default", func(t *testing.T) {
		assert.Equal(t, "default", GetStringDefault(props, "missing", "default"))
	})
}

func TestGetBoolDefault(t *testing.T) {
	props := map[string]any{"active": true}

	t.Run("existing key returns value", func(t *testing.T) {
		assert.True(t, GetBoolDefault(props, "active", false))
	})

	t.Run("missing key returns default", func(t *testing.T) {
		assert.True(t, GetBoolDefault(props, "missing", true))
	})
}

func TestGetIntDefault(t *testing.T) {
	props := map[string]any{"count": 10}

	t.Run("existing key returns value", func(t *testing.T) {
		assert.Equal(t, 10, GetIntDefault(props, "count", 0))
	})

	t.Run("missing key returns default", func(t *testing.T) {
		assert.Equal(t, 42, GetIntDefault(props, "missing", 42))
	})
}

// =============================================================================
// OBJECT OPERATIONS TESTS
// =============================================================================

func TestKeys(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		assert.Nil(t, Keys(nil))
	})

	t.Run("map[string]any returns keys", func(t *testing.T) {
		m := map[string]any{"a": 1, "b": 2}
		got := Keys(m)
		assert.Len(t, got, 2)
	})

	t.Run("map[any]any with string keys returns string keys", func(t *testing.T) {
		m := map[any]any{"a": 1, "b": 2, 3: "ignored"}
		got := Keys(m)
		assert.Len(t, got, 2)
	})
}

// =============================================================================
// MAP CONVERSION TESTS
// =============================================================================

func TestToGeneric(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		got, err := ToGeneric(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("map[any]any returns same map", func(t *testing.T) {
		m := map[any]any{"a": 1}
		got, err := ToGeneric(m)
		require.NoError(t, err)
		assert.Equal(t, 1, got["a"])
	})

	t.Run("map[string]any converts correctly", func(t *testing.T) {
		m := map[string]any{"a": 1}
		got, err := ToGeneric(m)
		require.NoError(t, err)
		assert.Equal(t, 1, got["a"])
	})

	t.Run("non-map returns error", func(t *testing.T) {
		_, err := ToGeneric("not a map")
		assert.ErrorIs(t, err, ErrInputNotMap)
	})
}
