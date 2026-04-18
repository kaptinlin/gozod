package mapx_test

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/pkg/mapx"
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

func TestIntCoerce(t *testing.T) {
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

func TestFloat64Coerce(t *testing.T) {
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

	t.Run("float32 value", func(t *testing.T) {
		m := map[string]any{"price": float32(19.5)}
		val, ok := mapx.Float64Coerce(m, "price")
		assert.Equal(t, 19.5, val)
		assert.True(t, ok)
	})

	t.Run("int value", func(t *testing.T) {
		m := map[string]any{"price": 20}
		val, ok := mapx.Float64Coerce(m, "price")
		assert.Equal(t, 20.0, val)
		assert.True(t, ok)
	})

	t.Run("int32 value", func(t *testing.T) {
		m := map[string]any{"price": int32(21)}
		val, ok := mapx.Float64Coerce(m, "price")
		assert.Equal(t, 21.0, val)
		assert.True(t, ok)
	})

	t.Run("int64 value", func(t *testing.T) {
		m := map[string]any{"price": int64(22)}
		val, ok := mapx.Float64Coerce(m, "price")
		assert.Equal(t, 22.0, val)
		assert.True(t, ok)
	})

	t.Run("string value", func(t *testing.T) {
		m := map[string]any{"price": "19.99"}
		val, ok := mapx.Float64Coerce(m, "price")
		assert.Equal(t, 0.0, val)
		assert.False(t, ok)
	})
}

func TestStrings(t *testing.T) {
	t.Run("string slice value", func(t *testing.T) {
		m := map[string]any{"tags": []string{"a", "b"}}
		val, ok := mapx.Strings(m, "tags")
		require.True(t, ok)
		if diff := cmp.Diff([]string{"a", "b"}, val); diff != "" {
			t.Errorf("Strings() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("non string slice value", func(t *testing.T) {
		m := map[string]any{"tags": []any{"a", "b"}}
		val, ok := mapx.Strings(m, "tags")
		assert.Nil(t, val)
		assert.False(t, ok)
	})
}

func TestAnySlice(t *testing.T) {
	t.Run("any slice value", func(t *testing.T) {
		m := map[string]any{"items": []any{"a", 1, true}}
		val, ok := mapx.AnySlice(m, "items")
		require.True(t, ok)
		if diff := cmp.Diff([]any{"a", 1, true}, val); diff != "" {
			t.Errorf("AnySlice() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("non any slice value", func(t *testing.T) {
		m := map[string]any{"items": []string{"a", "b"}}
		val, ok := mapx.AnySlice(m, "items")
		assert.Nil(t, val)
		assert.False(t, ok)
	})
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
		if diff := cmp.Diff(nested, val); diff != "" {
			t.Errorf("Map() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("non-map value", func(t *testing.T) {
		m := map[string]any{"config": "string"}
		_, ok := mapx.Map(m, "config")
		assert.False(t, ok)
	})
}

func TestAnyOr(t *testing.T) {
	m := map[string]any{"value": []int{1, 2, 3}}

	t.Run("existing key", func(t *testing.T) {
		got := mapx.AnyOr(m, "value", []int{9})
		if diff := cmp.Diff([]int{1, 2, 3}, got); diff != "" {
			t.Errorf("AnyOr() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("missing key", func(t *testing.T) {
		got := mapx.AnyOr(m, "missing", []int{9})
		if diff := cmp.Diff([]int{9}, got); diff != "" {
			t.Errorf("AnyOr() mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestStringsOr(t *testing.T) {
	m := map[string]any{"tags": []string{"go", "zod"}}

	t.Run("existing key", func(t *testing.T) {
		got := mapx.StringsOr(m, "tags", []string{"fallback"})
		if diff := cmp.Diff([]string{"go", "zod"}, got); diff != "" {
			t.Errorf("StringsOr() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("missing key", func(t *testing.T) {
		got := mapx.StringsOr(m, "missing", []string{"fallback"})
		if diff := cmp.Diff([]string{"fallback"}, got); diff != "" {
			t.Errorf("StringsOr() mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestAnySliceOr(t *testing.T) {
	m := map[string]any{"items": []any{"go", 1}}

	t.Run("existing key", func(t *testing.T) {
		got := mapx.AnySliceOr(m, "items", []any{"fallback"})
		if diff := cmp.Diff([]any{"go", 1}, got); diff != "" {
			t.Errorf("AnySliceOr() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("missing key", func(t *testing.T) {
		got := mapx.AnySliceOr(m, "missing", []any{"fallback"})
		if diff := cmp.Diff([]any{"fallback"}, got); diff != "" {
			t.Errorf("AnySliceOr() mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestMapOr(t *testing.T) {
	m := map[string]any{"config": map[string]any{"enabled": true}}

	t.Run("existing key", func(t *testing.T) {
		got := mapx.MapOr(m, "config", map[string]any{"enabled": false})
		want := map[string]any{"enabled": true}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("MapOr() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("missing key", func(t *testing.T) {
		got := mapx.MapOr(m, "missing", map[string]any{"enabled": false})
		want := map[string]any{"enabled": false}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("MapOr() mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestKeysFromStruct(t *testing.T) {
	type config struct {
		Enabled bool   `json:"enabled"`
		Name    string `json:"name,omitempty"`
		Ignored string `json:"-"`
		hidden  string
	}

	got := mapx.Keys(config{Enabled: true, Name: "gozod", Ignored: "skip", hidden: "skip"})
	slices.Sort(got)

	want := []string{"enabled", "name"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Keys() mismatch (-want +got):\n%s", diff)
	}
}

func TestToGenericArbitraryMap(t *testing.T) {
	type key struct{ ID int }

	got, err := mapx.ToGeneric(map[key]string{{ID: 1}: "value"})
	require.NoError(t, err)

	want := map[any]any{key{ID: 1}: "value"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ToGeneric() mismatch (-want +got):\n%s", diff)
	}
}

func TestToGenericStructReturnsError(t *testing.T) {
	type sample struct{ Name string }

	_, err := mapx.ToGeneric(sample{Name: "gozod"})
	assert.ErrorIs(t, err, mapx.ErrInputNotMap)
}

func TestToGenericMapIntAny(t *testing.T) {
	got, err := mapx.ToGeneric(map[int]any{1: "one"})
	require.NoError(t, err)

	want := map[any]any{1: "one"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ToGeneric() mismatch (-want +got):\n%s", diff)
	}
}

func TestToGenericMapAnyAny(t *testing.T) {
	input := map[any]any{"name": "gozod"}
	got, err := mapx.ToGeneric(input)
	require.NoError(t, err)

	if diff := cmp.Diff(input, got); diff != "" {
		t.Errorf("ToGeneric() mismatch (-want +got):\n%s", diff)
	}
}

func TestToGenericMapStringString(t *testing.T) {
	got, err := mapx.ToGeneric(map[string]string{"name": "gozod"})
	require.NoError(t, err)

	want := map[any]any{"name": "gozod"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ToGeneric() mismatch (-want +got):\n%s", diff)
	}
}

func TestToGenericMapStringInt(t *testing.T) {
	got, err := mapx.ToGeneric(map[string]int{"count": 3})
	require.NoError(t, err)

	want := map[any]any{"count": 3}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ToGeneric() mismatch (-want +got):\n%s", diff)
	}
}

func TestToGenericReflectMapStringBool(t *testing.T) {
	type boolMap map[string]bool

	got, err := mapx.ToGeneric(boolMap{"enabled": true})
	require.NoError(t, err)

	want := map[any]any{"enabled": true}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ToGeneric() mismatch (-want +got):\n%s", diff)
	}
}

func TestToGenericReflectMapPointerKey(t *testing.T) {
	type key struct{ ID int }

	k := &key{ID: 1}
	got, err := mapx.ToGeneric(map[*key]string{k: "value"})
	require.NoError(t, err)

	want := map[any]any{k: "value"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ToGeneric() mismatch (-want +got):\n%s", diff)
	}
}

func TestToGenericReflectMapStringNil(t *testing.T) {
	type alias map[string]*int

	got, err := mapx.ToGeneric(alias{"none": nil})
	require.NoError(t, err)

	want := map[any]any{"none": (*int)(nil)}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ToGeneric() mismatch (-want +got):\n%s", diff)
	}
}

func TestKeysNilStructInput(t *testing.T) {
	type config struct{ Enabled bool }
	var input *config
	assert.Nil(t, mapx.Keys(input))
}

func TestKeysNonStructInput(t *testing.T) {
	assert.Nil(t, mapx.Keys("gozod"))
}

func TestKeysMapStringAnyOrderAgnostic(t *testing.T) {
	got := mapx.Keys(map[string]any{"b": 2, "a": 1})
	slices.Sort(got)

	want := []string{"a", "b"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Keys() mismatch (-want +got):\n%s", diff)
	}
}

func TestKeysMapAnyAnyOrderAgnostic(t *testing.T) {
	got := mapx.Keys(map[any]any{"b": 2, "a": 1, 3: "skip"})
	slices.Sort(got)

	want := []string{"a", "b"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Keys() mismatch (-want +got):\n%s", diff)
	}
}

func TestKeysMapAnyAnyWithoutStringKeys(t *testing.T) {
	got := mapx.Keys(map[any]any{1: "one", 2: "two"})
	assert.Empty(t, got)
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

func TestIntCoerceOr(t *testing.T) {
	m := map[string]any{"count": int64(10)}

	t.Run("existing key", func(t *testing.T) {
		assert.Equal(t, 10, mapx.IntCoerceOr(m, "count", 0))
	})

	t.Run("missing key", func(t *testing.T) {
		assert.Equal(t, 42, mapx.IntCoerceOr(m, "missing", 42))
	})
}

func TestFloat64CoerceOr(t *testing.T) {
	m := map[string]any{"ratio": float32(2.5)}

	t.Run("existing key", func(t *testing.T) {
		assert.Equal(t, 2.5, mapx.Float64CoerceOr(m, "ratio", 0))
	})

	t.Run("missing key", func(t *testing.T) {
		assert.Equal(t, 1.5, mapx.Float64CoerceOr(m, "missing", 1.5))
	})
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
