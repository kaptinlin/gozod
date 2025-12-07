package slicex

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// SLICE CONVERSION TESTS
// =============================================================================

func TestToAny(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		got, err := ToAny(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("[]any returns same slice", func(t *testing.T) {
		input := []any{1, 2, 3}
		got, err := ToAny(input)
		require.NoError(t, err)
		assert.Len(t, got, 3)
	})

	t.Run("[]string converts correctly", func(t *testing.T) {
		input := []string{"a", "b", "c"}
		got, err := ToAny(input)
		require.NoError(t, err)
		assert.Len(t, got, 3)
		assert.Equal(t, "a", got[0])
	})

	t.Run("[]int converts correctly", func(t *testing.T) {
		input := []int{1, 2, 3}
		got, err := ToAny(input)
		require.NoError(t, err)
		assert.Len(t, got, 3)
		assert.Equal(t, 1, got[0])
	})

	t.Run("[]float64 converts correctly", func(t *testing.T) {
		input := []float64{1.1, 2.2, 3.3}
		got, err := ToAny(input)
		require.NoError(t, err)
		assert.Len(t, got, 3)
		assert.Equal(t, 1.1, got[0])
	})

	t.Run("[]bool converts correctly", func(t *testing.T) {
		input := []bool{true, false, true}
		got, err := ToAny(input)
		require.NoError(t, err)
		assert.Len(t, got, 3)
		assert.True(t, got[0].(bool))
	})

	t.Run("string converts to single element slice", func(t *testing.T) {
		input := "hello"
		got, err := ToAny(input)
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

	t.Run("[]any to []string works", func(t *testing.T) {
		input := []any{"a", "b", "c"}
		got, err := ToTyped[string](input)
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b", "c"}, got)
	})

	t.Run("[]any to []int works", func(t *testing.T) {
		input := []any{1, 2, 3}
		got, err := ToTyped[int](input)
		require.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, got)
	})
}

func TestToStrings(t *testing.T) {
	t.Run("converts slice to strings", func(t *testing.T) {
		input := []any{1, "hello", true}
		got, err := ToStrings(input)
		require.NoError(t, err)
		assert.Len(t, got, 3)
		assert.Equal(t, "1", got[0])
		assert.Equal(t, "hello", got[1])
		assert.Equal(t, "true", got[2])
	})
}

func TestFromReflect(t *testing.T) {
	t.Run("invalid value returns error", func(t *testing.T) {
		_, err := FromReflect(reflect.Value{})
		assert.ErrorIs(t, err, ErrInvalidReflectValue)
	})

	t.Run("slice converts correctly", func(t *testing.T) {
		input := []int{1, 2, 3}
		got, err := FromReflect(reflect.ValueOf(input))
		require.NoError(t, err)
		assert.Len(t, got, 3)
	})

	t.Run("array converts correctly", func(t *testing.T) {
		input := [3]int{1, 2, 3}
		got, err := FromReflect(reflect.ValueOf(input))
		require.NoError(t, err)
		assert.Len(t, got, 3)
	})

	t.Run("string converts to chars", func(t *testing.T) {
		input := "abc"
		got, err := FromReflect(reflect.ValueOf(input))
		require.NoError(t, err)
		assert.Len(t, got, 3)
		assert.Equal(t, "a", got[0])
	})

	t.Run("non-slice returns error", func(t *testing.T) {
		_, err := FromReflect(reflect.ValueOf(123))
		assert.Error(t, err)
	})
}

// =============================================================================
// SLICE EXTRACTION TESTS
// =============================================================================

func TestExtract(t *testing.T) {
	t.Run("valid slice extracts correctly", func(t *testing.T) {
		input := []int{1, 2, 3}
		got, ok := Extract(input)
		assert.True(t, ok)
		assert.Len(t, got, 3)
	})

	t.Run("invalid input returns false", func(t *testing.T) {
		_, ok := Extract(123)
		assert.False(t, ok)
	})
}

func TestExtractArray(t *testing.T) {
	t.Run("array extracts correctly", func(t *testing.T) {
		input := [3]int{1, 2, 3}
		got, ok := ExtractArray(input)
		assert.True(t, ok)
		assert.Len(t, got, 3)
	})

	t.Run("slice returns false", func(t *testing.T) {
		input := []int{1, 2, 3}
		_, ok := ExtractArray(input)
		assert.False(t, ok)
	})

	t.Run("nil returns false", func(t *testing.T) {
		_, ok := ExtractArray(nil)
		assert.False(t, ok)
	})
}

func TestExtractSlice(t *testing.T) {
	t.Run("slice extracts correctly", func(t *testing.T) {
		input := []int{1, 2, 3}
		got, ok := ExtractSlice(input)
		assert.True(t, ok)
		assert.Len(t, got, 3)
	})

	t.Run("array returns false", func(t *testing.T) {
		input := [3]int{1, 2, 3}
		_, ok := ExtractSlice(input)
		assert.False(t, ok)
	})
}

// =============================================================================
// SLICE MERGE TESTS
// =============================================================================

func TestMerge(t *testing.T) {
	t.Run("both nil returns nil", func(t *testing.T) {
		got, err := Merge(nil, nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("first nil returns second", func(t *testing.T) {
		second := []int{1, 2}
		got, err := Merge(nil, second)
		require.NoError(t, err)
		assert.Equal(t, second, got)
	})

	t.Run("second nil returns first", func(t *testing.T) {
		first := []int{1, 2}
		got, err := Merge(first, nil)
		require.NoError(t, err)
		assert.Equal(t, first, got)
	})

	t.Run("merges two slices", func(t *testing.T) {
		first := []int{1, 2}
		second := []int{3, 4}
		got, err := Merge(first, second)
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
		gotSlice := got.([]any)
		assert.Len(t, gotSlice, 3)
	})

	t.Run("appends to existing slice", func(t *testing.T) {
		slice := []int{1, 2}
		got, err := Append(slice, 3, 4)
		require.NoError(t, err)
		gotSlice := got.([]int)
		assert.Len(t, gotSlice, 4)
	})
}

func TestPrepend(t *testing.T) {
	t.Run("nil slice returns elements", func(t *testing.T) {
		got, err := Prepend(nil, 1, 2, 3)
		require.NoError(t, err)
		gotSlice := got.([]any)
		assert.Len(t, gotSlice, 3)
	})

	t.Run("prepends to existing slice", func(t *testing.T) {
		slice := []int{3, 4}
		got, err := Prepend(slice, 1, 2)
		require.NoError(t, err)
		gotSlice := got.([]int)
		assert.Len(t, gotSlice, 4)
		assert.Equal(t, 1, gotSlice[0])
		assert.Equal(t, 2, gotSlice[1])
	})
}

// =============================================================================
// SLICE UTILITY TESTS
// =============================================================================

func TestLength(t *testing.T) {
	t.Run("nil returns 0", func(t *testing.T) {
		got, err := Length(nil)
		require.NoError(t, err)
		assert.Equal(t, 0, got)
	})

	t.Run("slice returns correct length", func(t *testing.T) {
		got, err := Length([]int{1, 2, 3})
		require.NoError(t, err)
		assert.Equal(t, 3, got)
	})

	t.Run("array returns correct length", func(t *testing.T) {
		got, err := Length([5]int{})
		require.NoError(t, err)
		assert.Equal(t, 5, got)
	})

	t.Run("string returns correct length", func(t *testing.T) {
		got, err := Length("hello")
		require.NoError(t, err)
		assert.Equal(t, 5, got)
	})

	t.Run("non-slice returns error", func(t *testing.T) {
		_, err := Length(123)
		assert.ErrorIs(t, err, ErrNotSliceArrayOrString)
	})
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  bool
	}{
		{"nil is empty", nil, true},
		{"empty slice is empty", []int{}, true},
		{"non-empty slice is not empty", []int{1}, false},
		{"empty string is empty", "", true},
		{"non-empty string is not empty", "a", false},
		{"non-slice is empty (error)", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsEmpty(tt.input))
		})
	}
}

func TestContains(t *testing.T) {
	t.Run("finds existing element", func(t *testing.T) {
		assert.True(t, Contains([]int{1, 2, 3}, 2))
	})

	t.Run("does not find missing element", func(t *testing.T) {
		assert.False(t, Contains([]int{1, 2, 3}, 5))
	})

	t.Run("handles nil slice", func(t *testing.T) {
		assert.False(t, Contains(nil, 1))
	})
}

func TestIndexOf(t *testing.T) {
	t.Run("finds index of existing element", func(t *testing.T) {
		assert.Equal(t, 1, IndexOf([]int{1, 2, 3}, 2))
	})

	t.Run("returns -1 for missing element", func(t *testing.T) {
		assert.Equal(t, -1, IndexOf([]int{1, 2, 3}, 5))
	})

	t.Run("returns -1 for nil slice", func(t *testing.T) {
		assert.Equal(t, -1, IndexOf(nil, 1))
	})
}

func TestReverse(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		got, err := Reverse(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("reverses slice correctly", func(t *testing.T) {
		input := []int{1, 2, 3, 4}
		got, err := Reverse(input)
		require.NoError(t, err)
		gotSlice := got.([]int)
		assert.Equal(t, []int{4, 3, 2, 1}, gotSlice)
	})

	t.Run("single element stays same", func(t *testing.T) {
		input := []int{1}
		got, err := Reverse(input)
		require.NoError(t, err)
		gotSlice := got.([]int)
		assert.Equal(t, []int{1}, gotSlice)
	})
}

func TestUnique(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		got, err := Unique(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("removes duplicates", func(t *testing.T) {
		input := []int{1, 2, 2, 3, 3, 3}
		got, err := Unique(input)
		require.NoError(t, err)
		gotSlice := got.([]int)
		assert.Len(t, gotSlice, 3)
	})

	t.Run("preserves order", func(t *testing.T) {
		input := []string{"c", "a", "b", "a", "c"}
		got, err := Unique(input)
		require.NoError(t, err)
		gotSlice := got.([]string)
		assert.Equal(t, []string{"c", "a", "b"}, gotSlice)
	})
}

func TestFilter(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		got, err := Filter(nil, func(any) bool { return true })
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("filters elements correctly", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		got, err := Filter(input, func(v any) bool {
			return v.(int) > 2
		})
		require.NoError(t, err)
		gotSlice := got.([]int)
		assert.Len(t, gotSlice, 3)
	})
}

func TestMap(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		got, err := Map(nil, func(any) any { return nil })
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("maps elements correctly", func(t *testing.T) {
		input := []int{1, 2, 3}
		got, err := Map(input, func(v any) any {
			return v.(int) * 2
		})
		require.NoError(t, err)
		gotSlice := got.([]int)
		assert.Equal(t, []int{2, 4, 6}, gotSlice)
	})
}

func TestJoin(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		separator string
		want      string
	}{
		{"empty slice", []int{}, ",", ""},
		{"single element", []int{1}, ",", "1"},
		{"multiple elements", []int{1, 2, 3}, ",", "1,2,3"},
		{"string slice", []string{"a", "b", "c"}, "-", "a-b-c"},
		{"nil values", []any{"a", nil, "b"}, ",", "a,,b"},
		{"invalid input", 123, ",", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Join(tt.input, tt.separator))
		})
	}
}

func TestStringToChars(t *testing.T) {
	t.Run("converts string to char slice", func(t *testing.T) {
		got := StringToChars("abc")
		assert.Len(t, got, 3)
		assert.Equal(t, "a", got[0])
		assert.Equal(t, "b", got[1])
		assert.Equal(t, "c", got[2])
	})

	t.Run("handles unicode", func(t *testing.T) {
		got := StringToChars("你好")
		assert.Len(t, got, 2)
	})

	t.Run("empty string", func(t *testing.T) {
		got := StringToChars("")
		assert.Empty(t, got)
	})
}
