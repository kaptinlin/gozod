package reflectx

import (
	"errors"
	"math/big"
	"testing"

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
			if got := IsNil(tt.v); got != tt.want {
				t.Errorf("IsNil() = %v, want %v", got, tt.want)
			}
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
			if got := IsBool(tt.v); got != tt.want {
				t.Errorf("IsBool() = %v, want %v", got, tt.want)
			}
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
			if got := IsString(tt.v); got != tt.want {
				t.Errorf("IsString() = %v, want %v", got, tt.want)
			}
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
			if got := IsNumeric(tt.v); got != tt.want {
				t.Errorf("IsNumeric() = %v, want %v", got, tt.want)
			}
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
			if got := IsArray(tt.v); got != tt.want {
				t.Errorf("IsArray() = %v, want %v", got, tt.want)
			}
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
			if got := IsSlice(tt.v); got != tt.want {
				t.Errorf("IsSlice() = %v, want %v", got, tt.want)
			}
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
			if got := IsMap(tt.v); got != tt.want {
				t.Errorf("IsMap() = %v, want %v", got, tt.want)
			}
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
			if got := IsStruct(tt.v); got != tt.want {
				t.Errorf("IsStruct() = %v, want %v", got, tt.want)
			}
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
			if got := ParsedType(tt.v); got != tt.want {
				t.Errorf("ParsedType() = %v, want %v", got, tt.want)
			}
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
			if got := ParsedCategory(tt.v); got != tt.want {
				t.Errorf("ParsedCategory() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Pointer utilities
// ---------------------------------------------------------------------------

func TestDeref(t *testing.T) {
	x := 42
	tests := []struct {
		name    string
		v       any
		wantVal any
		wantOK  bool
	}{
		{"nil", nil, nil, false},
		{"concrete int", 42, 42, true},
		{"pointer", &x, 42, true},
		{"nil pointer", (*int)(nil), nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := Deref(tt.v)
			if ok != tt.wantOK {
				t.Errorf("Deref() ok = %v, want %v", ok, tt.wantOK)
			}
			if val != tt.wantVal {
				t.Errorf("Deref() val = %v, want %v", val, tt.wantVal)
			}
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
			if got := DerefAll(tt.v); got != tt.want {
				t.Errorf("DerefAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Value extraction
// ---------------------------------------------------------------------------

func TestExtractString(t *testing.T) {
	tests := []struct {
		name    string
		v       any
		wantStr string
		wantOK  bool
	}{
		{"string", "hello", "hello", true},
		{"empty string", "", "", true},
		{"int", 42, "", false},
		{"nil", nil, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, ok := ExtractString(tt.v)
			if ok != tt.wantOK || s != tt.wantStr {
				t.Errorf("ExtractString() = (%q, %v), want (%q, %v)", s, ok, tt.wantStr, tt.wantOK)
			}
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
			if got := HasLength(tt.v); got != tt.want {
				t.Errorf("HasLength() = %v, want %v", got, tt.want)
			}
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
			if got := HasSize(tt.v); got != tt.want {
				t.Errorf("HasSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLength(t *testing.T) {
	tests := []struct {
		name    string
		v       any
		wantLen int
		wantOK  bool
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
			l, ok := Length(tt.v)
			if ok != tt.wantOK || l != tt.wantLen {
				t.Errorf("Length() = (%d, %v), want (%d, %v)", l, ok, tt.wantLen, tt.wantOK)
			}
		})
	}
}

func TestSize(t *testing.T) {
	tests := []struct {
		name     string
		v        any
		wantSize int
		wantOK   bool
	}{
		{"map", map[string]int{"a": 1, "b": 2}, 2, true},
		{"slice", []int{1, 2}, 2, true},
		{"array", [3]int{}, 3, true},
		{"string", "hi", 0, false},
		{"nil", nil, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, ok := Size(tt.v)
			if ok != tt.wantOK || s != tt.wantSize {
				t.Errorf("Size() = (%d, %v), want (%d, %v)", s, ok, tt.wantSize, tt.wantOK)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Generic conversion
// ---------------------------------------------------------------------------

func TestConvertToGeneric(t *testing.T) {
	t.Run("same type", func(t *testing.T) {
		got, err := ConvertToGeneric[int](42)
		if err != nil || got != 42 {
			t.Errorf("ConvertToGeneric[int](42) = (%v, %v), want (42, nil)", got, err)
		}
	})

	t.Run("numeric conversion", func(t *testing.T) {
		got, err := ConvertToGeneric[float64](42)
		if err != nil || got != 42.0 {
			t.Errorf("ConvertToGeneric[float64](42) = (%v, %v), want (42.0, nil)", got, err)
		}
	})

	t.Run("nil input", func(t *testing.T) {
		_, err := ConvertToGeneric[int](nil)
		if !errors.Is(err, ErrNilValue) {
			t.Errorf("ConvertToGeneric[int](nil) error = %v, want ErrNilValue", err)
		}
	})

	t.Run("unsupported conversion", func(t *testing.T) {
		_, err := ConvertToGeneric[int]("hello")
		if !errors.Is(err, ErrUnsupportedConversion) {
			t.Errorf("ConvertToGeneric[int](string) error = %v, want ErrUnsupportedConversion", err)
		}
	})
}
