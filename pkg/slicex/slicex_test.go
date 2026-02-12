package slicex

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToAny(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		got, err := ToAny(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("[]any returns same slice", func(t *testing.T) {
		got, err := ToAny([]any{1, 2, 3})
		require.NoError(t, err)
		assert.Len(t, got, 3)
	})

	t.Run("[]string", func(t *testing.T) {
		got, err := ToAny([]string{"a", "b", "c"})
		require.NoError(t, err)
		assert.Len(t, got, 3)
		assert.Equal(t, "a", got[0])
	})

	t.Run("[]int", func(t *testing.T) {
		got, err := ToAny([]int{1, 2, 3})
		require.NoError(t, err)
		assert.Len(t, got, 3)
		assert.Equal(t, 1, got[0])
	})

	t.Run("[]float64", func(t *testing.T) {
		got, err := ToAny([]float64{1.1, 2.2, 3.3})
		require.NoError(t, err)
		assert.Len(t, got, 3)
		assert.Equal(t, 1.1, got[0])
	})

	t.Run("[]bool", func(t *testing.T) {
		got, err := ToAny([]bool{true, false, true})
		require.NoError(t, err)
		assert.Len(t, got, 3)
		assert.True(t, got[0].(bool))
	})

	t.Run("string as single element", func(t *testing.T) {
		got, err := ToAny("hello")
		require.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, "hello", got[0])
	})
}

func TestToTyped(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		got, err := ToTyped[string](nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("[]any to []string", func(t *testing.T) {
		got, err := ToTyped[string]([]any{"a", "b", "c"})
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b", "c"}, got)
	})

	t.Run("[]any to []int", func(t *testing.T) {
		got, err := ToTyped[int]([]any{1, 2, 3})
		require.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, got)
	})
}

func TestToStrings(t *testing.T) {
	t.Run("mixed types", func(t *testing.T) {
		got, err := ToStrings([]any{1, "hello", true})
		require.NoError(t, err)
		assert.Equal(t, []string{"1", "hello", "true"}, got)
	})
}

func TestFromReflect(t *testing.T) {
	t.Run("invalid value returns error", func(t *testing.T) {
		_, err := FromReflect(reflect.Value{})
		assert.ErrorIs(t, err, ErrInvalidReflectValue)
	})

	t.Run("slice", func(t *testing.T) {
		got, err := FromReflect(reflect.ValueOf([]int{1, 2, 3}))
		require.NoError(t, err)
		assert.Len(t, got, 3)
	})

	t.Run("array", func(t *testing.T) {
		got, err := FromReflect(reflect.ValueOf([3]int{1, 2, 3}))
		require.NoError(t, err)
		assert.Len(t, got, 3)
	})

	t.Run("string converts to chars", func(t *testing.T) {
		got, err := FromReflect(reflect.ValueOf("abc"))
		require.NoError(t, err)
		assert.Len(t, got, 3)
		assert.Equal(t, "a", got[0])
	})

	t.Run("non-slice returns error", func(t *testing.T) {
		_, err := FromReflect(reflect.ValueOf(123))
		assert.Error(t, err)
	})
}

func TestExtract(t *testing.T) {
	t.Run("valid slice", func(t *testing.T) {
		got, ok := Extract([]int{1, 2, 3})
		assert.True(t, ok)
		assert.Len(t, got, 3)
	})

	t.Run("invalid input", func(t *testing.T) {
		_, ok := Extract(123)
		assert.False(t, ok)
	})
}

func TestExtractArray(t *testing.T) {
	t.Run("array", func(t *testing.T) {
		got, ok := ExtractArray([3]int{1, 2, 3})
		assert.True(t, ok)
		assert.Len(t, got, 3)
	})

	t.Run("slice returns false", func(t *testing.T) {
		_, ok := ExtractArray([]int{1, 2, 3})
		assert.False(t, ok)
	})

	t.Run("nil returns false", func(t *testing.T) {
		_, ok := ExtractArray(nil)
		assert.False(t, ok)
	})
}

func TestExtractSlice(t *testing.T) {
	t.Run("slice", func(t *testing.T) {
		got, ok := ExtractSlice([]int{1, 2, 3})
		assert.True(t, ok)
		assert.Len(t, got, 3)
	})

	t.Run("array returns false", func(t *testing.T) {
		_, ok := ExtractSlice([3]int{1, 2, 3})
		assert.False(t, ok)
	})
}

func TestMerge(t *testing.T) {
	t.Run("both nil", func(t *testing.T) {
		got, err := Merge(nil, nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("first nil returns second", func(t *testing.T) {
		got, err := Merge(nil, []int{1, 2})
		require.NoError(t, err)
		assert.Equal(t, []int{1, 2}, got)
	})

	t.Run("second nil returns first", func(t *testing.T) {
		got, err := Merge([]int{1, 2}, nil)
		require.NoError(t, err)
		assert.Equal(t, []int{1, 2}, got)
	})

	t.Run("two slices", func(t *testing.T) {
		got, err := Merge([]int{1, 2}, []int{3, 4})
		require.NoError(t, err)
		gotSlice, ok := got.([]int)
		require.True(t, ok)
		assert.Len(t, gotSlice, 4)
	})
}

func TestAppend(t *testing.T) {
	t.Run("nil slice returns elements", func(t *testing.T) {
		got, err := Append(nil, 1, 2, 3)
		require.NoError(t, err)
		assert.Len(t, got.([]any), 3)
	})

	t.Run("appends to existing", func(t *testing.T) {
		got, err := Append([]int{1, 2}, 3, 4)
		require.NoError(t, err)
		assert.Len(t, got.([]int), 4)
	})
}

func TestPrepend(t *testing.T) {
	t.Run("nil slice returns elements", func(t *testing.T) {
		got, err := Prepend(nil, 1, 2, 3)
		require.NoError(t, err)
		assert.Len(t, got.([]any), 3)
	})

	t.Run("prepends to existing", func(t *testing.T) {
		got, err := Prepend([]int{3, 4}, 1, 2)
		require.NoError(t, err)
		gotSlice := got.([]int)
		assert.Len(t, gotSlice, 4)
		assert.Equal(t, 1, gotSlice[0])
		assert.Equal(t, 2, gotSlice[1])
	})
}

func TestLength(t *testing.T) {
	t.Run("nil returns 0", func(t *testing.T) {
		got, err := Length(nil)
		require.NoError(t, err)
		assert.Equal(t, 0, got)
	})

	t.Run("slice", func(t *testing.T) {
		got, err := Length([]int{1, 2, 3})
		require.NoError(t, err)
		assert.Equal(t, 3, got)
	})

	t.Run("array", func(t *testing.T) {
		got, err := Length([5]int{})
		require.NoError(t, err)
		assert.Equal(t, 5, got)
	})

	t.Run("string", func(t *testing.T) {
		got, err := Length("hello")
		require.NoError(t, err)
		assert.Equal(t, 5, got)
	})

	t.Run("non-slice returns error", func(t *testing.T) {
		_, err := Length(123)
		assert.ErrorIs(t, err, ErrNotCollection)
	})
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  bool
	}{
		{name: "nil", input: nil, want: true},
		{name: "empty slice", input: []int{}, want: true},
		{name: "non-empty slice", input: []int{1}, want: false},
		{name: "empty string", input: "", want: true},
		{name: "non-empty string", input: "a", want: false},
		{name: "non-slice (error)", input: 123, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsEmpty(tt.input))
		})
	}
}

func TestContains(t *testing.T) {
	t.Run("finds existing", func(t *testing.T) {
		assert.True(t, Contains([]int{1, 2, 3}, 2))
	})

	t.Run("missing element", func(t *testing.T) {
		assert.False(t, Contains([]int{1, 2, 3}, 5))
	})

	t.Run("nil slice", func(t *testing.T) {
		assert.False(t, Contains(nil, 1))
	})
}

func TestIndexOf(t *testing.T) {
	t.Run("finds index", func(t *testing.T) {
		assert.Equal(t, 1, IndexOf([]int{1, 2, 3}, 2))
	})

	t.Run("missing returns -1", func(t *testing.T) {
		assert.Equal(t, -1, IndexOf([]int{1, 2, 3}, 5))
	})

	t.Run("nil returns -1", func(t *testing.T) {
		assert.Equal(t, -1, IndexOf(nil, 1))
	})
}

func TestReverse(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		got, err := Reverse(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("reverses correctly", func(t *testing.T) {
		got, err := Reverse([]int{1, 2, 3, 4})
		require.NoError(t, err)
		assert.Equal(t, []int{4, 3, 2, 1}, got.([]int))
	})

	t.Run("single element", func(t *testing.T) {
		got, err := Reverse([]int{1})
		require.NoError(t, err)
		assert.Equal(t, []int{1}, got.([]int))
	})
}

func TestUnique(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		got, err := Unique(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("removes duplicates", func(t *testing.T) {
		got, err := Unique([]int{1, 2, 2, 3, 3, 3})
		require.NoError(t, err)
		assert.Len(t, got.([]int), 3)
	})

	t.Run("preserves order", func(t *testing.T) {
		got, err := Unique([]string{"c", "a", "b", "a", "c"})
		require.NoError(t, err)
		assert.Equal(t, []string{"c", "a", "b"}, got.([]string))
	})
}

func TestFilter(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		got, err := Filter(nil, func(any) bool { return true })
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("filters correctly", func(t *testing.T) {
		got, err := Filter([]int{1, 2, 3, 4, 5}, func(v any) bool {
			return v.(int) > 2
		})
		require.NoError(t, err)
		assert.Len(t, got.([]int), 3)
	})
}

func TestMap(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		got, err := Map(nil, func(any) any { return nil })
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("maps correctly", func(t *testing.T) {
		got, err := Map([]int{1, 2, 3}, func(v any) any {
			return v.(int) * 2
		})
		require.NoError(t, err)
		assert.Equal(t, []int{2, 4, 6}, got.([]int))
	})
}

func TestJoin(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		separator string
		want      string
	}{
		{name: "empty slice", input: []int{}, separator: ",", want: ""},
		{name: "single element", input: []int{1}, separator: ",", want: "1"},
		{name: "multiple elements", input: []int{1, 2, 3}, separator: ",", want: "1,2,3"},
		{name: "string slice", input: []string{"a", "b", "c"}, separator: "-", want: "a-b-c"},
		{name: "nil values", input: []any{"a", nil, "b"}, separator: ",", want: "a,,b"},
		{name: "invalid input", input: 123, separator: ",", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Join(tt.input, tt.separator))
		})
	}
}

func TestStringToChars(t *testing.T) {
	t.Run("ascii", func(t *testing.T) {
		got := StringToChars("abc")
		assert.Equal(t, []any{"a", "b", "c"}, got)
	})

	t.Run("unicode", func(t *testing.T) {
		got := StringToChars("你好")
		assert.Len(t, got, 2)
	})

	t.Run("empty string", func(t *testing.T) {
		assert.Empty(t, StringToChars(""))
	})
}
