package reflectx

import (
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
)

// ---------------------------------------------------------------------------
// Type checking: IsNil
// ---------------------------------------------------------------------------

func TestIsNil(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want bool
	}{
		{"untyped nil", nil, true},
		{"string", "hello", false},
		{"nil pointer", (*int)(nil), true},
		{"non-nil pointer", new(int), false},
		{"nil slice", []int(nil), true},
		{"empty slice", []int{}, false},
		{"nil map", map[string]int(nil), true},
		{"empty map", map[string]int{}, false},
		{"nil chan", (chan int)(nil), true},
		{"nil func", (func())(nil), true},
		{"int", 42, false},
		{"bool", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsNil(tt.v))
		})
	}
}

// ---------------------------------------------------------------------------
// Primitive type checks
// ---------------------------------------------------------------------------

func TestIsBool(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want bool
	}{
		{"true", true, true},
		{"false", false, true},
		{"int", 1, false},
		{"string", "true", false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsBool(tt.v))
		})
	}
}

func TestIsString(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want bool
	}{
		{"string", "hello", true},
		{"empty string", "", true},
		{"int", 123, false},
		{"nil", nil, false},
		{"bool", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsString(tt.v))
		})
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want bool
	}{
		{"int", 123, true},
		{"int8", int8(1), true},
		{"int64", int64(1), true},
		{"uint", uint(1), true},
		{"uint64", uint64(1), true},
		{"float32", float32(1.0), true},
		{"float64", 1.0, true},
		{"complex128", 1 + 2i, true},
		{"*big.Int", big.NewInt(1), true},
		{"big.Int value", *big.NewInt(1), true},
		{"string", "123", false},
		{"bool", true, false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsNumeric(tt.v))
		})
	}
}

// ---------------------------------------------------------------------------
// Kind-based type checks
// ---------------------------------------------------------------------------

func TestIsArray(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want bool
	}{
		{"array", [3]int{1, 2, 3}, true},
		{"slice", []int{1, 2, 3}, false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsArray(tt.v))
		})
	}
}

func TestIsSlice(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want bool
	}{
		{"slice", []int{1, 2, 3}, true},
		{"nil slice", []int(nil), true},
		{"array", [3]int{1, 2, 3}, false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsSlice(tt.v))
		})
	}
}

func TestIsMap(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want bool
	}{
		{"map", map[string]int{"a": 1}, true},
		{"nil map", map[string]int(nil), true},
		{"slice", []int{1}, false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsMap(tt.v))
		})
	}
}

func TestIsStruct(t *testing.T) {
	type s struct{ X int }
	tests := []struct {
		name string
		v    any
		want bool
	}{
		{"struct", s{1}, true},
		{"pointer to struct", &s{1}, false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsStruct(tt.v))
		})
	}
}

// ---------------------------------------------------------------------------
// Type parsing
// ---------------------------------------------------------------------------

func TestParsedType(t *testing.T) {
	n := 42
	tests := []struct {
		name string
		v    any
		want core.ParsedType
	}{
		{"nil", nil, core.ParsedTypeNil},
		{"bool", true, core.ParsedTypeBool},
		{"string", "hello", core.ParsedTypeString},
		{"int", 123, core.ParsedTypeNumber},
		{"uint", uint(1), core.ParsedTypeNumber},
		{"float64", 1.5, core.ParsedTypeFloat},
		{"complex128", 1 + 2i, core.ParsedTypeComplex},
		{"*big.Int", big.NewInt(99), core.ParsedTypeBigint},
		{"slice", []int{1, 2}, core.ParsedTypeSlice},
		{"array", [2]int{1, 2}, core.ParsedTypeArray},
		{"map", map[string]int{"a": 1}, core.ParsedTypeMap},
		{"struct", struct{ X int }{1}, core.ParsedTypeStruct},
		{"func", func() {}, core.ParsedTypeFunction},
		{"pointer deref", &n, core.ParsedTypeNumber},
		{"nil pointer", (*int)(nil), core.ParsedTypeNil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ParsedType(tt.v))
		})
	}
}

func TestParsedCategory(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want string
	}{
		{"nil", nil, "nil"},
		{"string", "hello", "string"},
		{"bool", true, "bool"},
		{"int", 123, "number"},
		{"float64", 1.5, "number"},
		{"complex128", 1 + 2i, "number"},
		{"*big.Int", big.NewInt(1), "number"},
		{"slice", []int{1}, "array"},
		{"array", [2]int{1, 2}, "array"},
		{"map", map[string]int{"a": 1}, "object"},
		{"struct", struct{ X int }{1}, "object"},
		{"func", func() {}, "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ParsedCategory(tt.v))
		})
	}
}

// ---------------------------------------------------------------------------
// Pointer utilities
// ---------------------------------------------------------------------------

func TestDeref(t *testing.T) {
	x := 42
	tests := []struct {
		name string
		v    any
		want any
		ok   bool
	}{
		{"nil", nil, nil, false},
		{"concrete int", 42, 42, true},
		{"pointer", &x, 42, true},
		{"nil pointer", (*int)(nil), nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := Deref(tt.v)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, val)
		})
	}
}

func TestDerefAll(t *testing.T) {
	x := 42
	px := &x
	ppx := &px
	tests := []struct {
		name string
		v    any
		want any
	}{
		{"nil", nil, nil},
		{"concrete", 42, 42},
		{"single pointer", &x, 42},
		{"double pointer", ppx, 42},
		{"nil pointer", (*int)(nil), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, DerefAll(tt.v))
		})
	}
	_ = ppx
}

// ---------------------------------------------------------------------------
// Value extraction
// ---------------------------------------------------------------------------

func TestStringVal(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want string
		ok   bool
	}{
		{"string", "hello", "hello", true},
		{"empty string", "", "", true},
		{"int", 42, "", false},
		{"nil", nil, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, ok := StringVal(tt.v)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, s)
		})
	}
}

func TestHasLength(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want bool
	}{
		{"string", "hi", true},
		{"slice", []int{1}, true},
		{"array", [2]int{}, true},
		{"map", map[string]int{}, false},
		{"int", 42, false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, HasLength(tt.v))
		})
	}
}

func TestHasSize(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want bool
	}{
		{"map", map[string]int{"a": 1}, true},
		{"slice", []int{1}, true},
		{"array", [2]int{}, true},
		{"chan", make(chan int), true},
		{"string", "hi", false},
		{"int", 42, false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, HasSize(tt.v))
		})
	}
}

func TestLength(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want int
		ok   bool
	}{
		{"string", "hello", 5, true},
		{"slice", []int{1, 2, 3}, 3, true},
		{"array", [2]int{1, 2}, 2, true},
		{"empty slice", []int{}, 0, true},
		{"map", map[string]int{"a": 1}, 0, false},
		{"nil", nil, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, ok := Length(tt.v)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, n)
		})
	}
}

func TestSize(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want int
		ok   bool
	}{
		{"map", map[string]int{"a": 1, "b": 2}, 2, true},
		{"slice", []int{1, 2}, 2, true},
		{"array", [3]int{}, 3, true},
		{"string", "hi", 0, false},
		{"nil", nil, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, ok := Size(tt.v)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, n)
		})
	}
}

// ---------------------------------------------------------------------------
// Generic conversion
// ---------------------------------------------------------------------------

func TestConvert(t *testing.T) {
	t.Run("same type", func(t *testing.T) {
		got, err := Convert[int](42)
		require.NoError(t, err)
		assert.Equal(t, 42, got)
	})

	t.Run("numeric conversion", func(t *testing.T) {
		got, err := Convert[float64](42)
		require.NoError(t, err)
		assert.Equal(t, 42.0, got)
	})

	t.Run("nil input", func(t *testing.T) {
		_, err := Convert[int](nil)
		assert.True(t, errors.Is(err, ErrNil))
	})

	t.Run("unsupported", func(t *testing.T) {
		_, err := Convert[int]("hello")
		assert.True(t, errors.Is(err, ErrUnsupported))
	})
}
