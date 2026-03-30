package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/core"
)

func TestToErrorMap(t *testing.T) {
	t.Run("string input", func(t *testing.T) {
		m, ok := ToErrorMap("custom error")
		require.True(t, ok, "expected ok=true for string input")
		require.NotNil(t, m, "expected non-nil error map")
		got := (*m)(core.ZodRawIssue{})
		assert.Equal(t, "custom error", got)
	})

	t.Run("ZodErrorMap input", func(t *testing.T) {
		fn := core.ZodErrorMap(func(core.ZodRawIssue) string {
			return "mapped"
		})
		m, ok := ToErrorMap(fn)
		require.True(t, ok, "expected ok=true")
		assert.Equal(t, "mapped", (*m)(core.ZodRawIssue{}))
	})

	t.Run("*ZodErrorMap input", func(t *testing.T) {
		fn := core.ZodErrorMap(func(core.ZodRawIssue) string {
			return "ptr"
		})
		m, ok := ToErrorMap(&fn)
		require.True(t, ok, "expected ok=true")
		assert.Equal(t, "ptr", (*m)(core.ZodRawIssue{}))
	})

	t.Run("func input", func(t *testing.T) {
		fn := func(core.ZodRawIssue) string { return "func" }
		m, ok := ToErrorMap(fn)
		require.True(t, ok, "expected ok=true")
		assert.Equal(t, "func", (*m)(core.ZodRawIssue{}))
	})

	t.Run("unsupported input", func(t *testing.T) {
		m, ok := ToErrorMap(42)
		assert.False(t, ok, "expected ok=false for int input")
		assert.Nil(t, m, "expected nil map")
	})
}

func TestFirstParam(t *testing.T) {
	assert.Nil(t, FirstParam())
	assert.Equal(t, "hello", FirstParam("hello"))
	assert.Equal(t, 1, FirstParam(1, 2, 3))
}

func TestOriginFromValue(t *testing.T) {
	tests := []struct {
		input any
		want  string
	}{
		{42, "number"},
		{3.14, "number"},
		{"hello", "string"},
		{[]int{1, 2}, "array"},
		{map[string]int{"a": 1}, "object"},
		{struct{}{}, "unknown"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, OriginFromValue(tt.input))
	}
}

func TestNumericOrigin(t *testing.T) {
	tests := []struct {
		input any
		want  string
	}{
		{nil, "nil"},
		{42, "integer"},
		{int8(1), "integer"},
		{int64(100), "integer"},
		{uint(5), "integer"},
		{uint32(10), "integer"},
		{3.14, "number"},
		{float32(1.5), "number"},
		{"not a number", "string"},
		{true, "unknown"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, NumericOrigin(tt.input))
	}
}

func TestSizableOrigin(t *testing.T) {
	tests := []struct {
		input any
		want  string
	}{
		{nil, "nil"},
		{"hello", "string"},
		{[]int{1, 2}, "slice"},
		{[2]int{1, 2}, "array"},
		{map[string]int{"a": 1}, "map"},
		{struct{ Name string }{"x"}, "struct"},
		{42, "unknown"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, SizableOrigin(tt.input))
	}
}

func TestLengthableOrigin(t *testing.T) {
	tests := []struct {
		input any
		want  string
	}{
		{nil, "nil"},
		{"hello", "string"},
		{[]int{1}, "slice"},
		{[1]int{1}, "array"},
		{42, "unknown"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, LengthableOrigin(tt.input))
	}
}

func TestCompareValues(t *testing.T) {
	t.Run("both nil", func(t *testing.T) {
		assert.Equal(t, 0, CompareValues(nil, nil))
	})
	t.Run("first nil", func(t *testing.T) {
		assert.Equal(t, -1, CompareValues(nil, 1))
	})
	t.Run("second nil", func(t *testing.T) {
		assert.Equal(t, 1, CompareValues(1, nil))
	})
	t.Run("int equal", func(t *testing.T) {
		assert.Equal(t, 0, CompareValues(5, 5))
	})
	t.Run("int less", func(t *testing.T) {
		assert.Equal(t, -1, CompareValues(3, 5))
	})
	t.Run("int greater", func(t *testing.T) {
		assert.Equal(t, 1, CompareValues(5, 3))
	})
	t.Run("int64", func(t *testing.T) {
		assert.Equal(t, -1, CompareValues(int64(1), int64(2)))
	})
	t.Run("float64", func(t *testing.T) {
		assert.Equal(t, -1, CompareValues(1.5, 2.5))
	})
	t.Run("float32", func(t *testing.T) {
		a, b := float32(3.0), float32(1.0)
		assert.Equal(t, 1, CompareValues(a, b))
	})
	t.Run("string", func(t *testing.T) {
		assert.Equal(t, -1, CompareValues("a", "b"))
	})
	t.Run("mismatched types", func(t *testing.T) {
		assert.Equal(t, 0, CompareValues(1, "a"))
	})
	t.Run("pointer deref", func(t *testing.T) {
		a, b := 10, 5
		assert.Equal(t, 1, CompareValues(&a, &b))
	})
}

func TestNormalizeParams(t *testing.T) {
	t.Run("no args", func(t *testing.T) {
		p := NormalizeParams()
		require.NotNil(t, p)
		assert.Nil(t, p.Error)
	})
	t.Run("nil arg", func(t *testing.T) {
		p := NormalizeParams(nil)
		require.NotNil(t, p)
		assert.Nil(t, p.Error)
	})
	t.Run("string arg", func(t *testing.T) {
		p := NormalizeParams("err msg")
		assert.Equal(t, "err msg", p.Error)
	})
	t.Run("SchemaParams value", func(t *testing.T) {
		sp := core.SchemaParams{Error: "val"}
		p := NormalizeParams(sp)
		assert.Equal(t, "val", p.Error)
	})
	t.Run("*SchemaParams", func(t *testing.T) {
		sp := &core.SchemaParams{Error: "ptr"}
		p := NormalizeParams(sp)
		assert.Equal(t, "ptr", p.Error)
		// verify copy semantics
		sp.Error = "changed"
		assert.Equal(t, "ptr", p.Error, "expected copy, not reference")
	})
	t.Run("nil *SchemaParams", func(t *testing.T) {
		var sp *core.SchemaParams
		p := NormalizeParams(sp)
		require.NotNil(t, p)
		assert.Nil(t, p.Error)
	})
	t.Run("unsupported type", func(t *testing.T) {
		p := NormalizeParams(42)
		require.NotNil(t, p)
		assert.Nil(t, p.Error)
	})
}

func TestNormalizeCustomParams(t *testing.T) {
	t.Run("no args", func(t *testing.T) {
		p := NormalizeCustomParams()
		require.NotNil(t, p)
		assert.Nil(t, p.Error)
	})
	t.Run("nil arg", func(t *testing.T) {
		p := NormalizeCustomParams(nil)
		require.NotNil(t, p)
		assert.Nil(t, p.Error)
	})
	t.Run("string arg", func(t *testing.T) {
		p := NormalizeCustomParams("err")
		assert.Equal(t, "err", p.Error)
	})
	t.Run("CustomParams value", func(t *testing.T) {
		cp := core.CustomParams{Error: "val"}
		p := NormalizeCustomParams(cp)
		assert.Equal(t, "val", p.Error)
	})
	t.Run("*CustomParams", func(t *testing.T) {
		cp := &core.CustomParams{Error: "ptr"}
		p := NormalizeCustomParams(cp)
		assert.Equal(t, "ptr", p.Error)
	})
	t.Run("nil *CustomParams", func(t *testing.T) {
		var cp *core.CustomParams
		p := NormalizeCustomParams(cp)
		require.NotNil(t, p)
		assert.Nil(t, p.Error)
	})
	t.Run("any as error", func(t *testing.T) {
		p := NormalizeCustomParams(42)
		assert.Equal(t, 42, p.Error)
	})
}

func TestApplySchemaParams(t *testing.T) {
	t.Run("nil params", func(t *testing.T) {
		def := &core.ZodTypeDef{}
		ApplySchemaParams(def, nil)
		assert.Nil(t, def.Error)
	})
	t.Run("with error string", func(t *testing.T) {
		def := &core.ZodTypeDef{}
		params := &core.SchemaParams{Error: "bad"}
		ApplySchemaParams(def, params)
		require.NotNil(t, def.Error, "expected non-nil error map")
		got := (*def.Error)(core.ZodRawIssue{})
		assert.Equal(t, "bad", got)
	})
	t.Run("nil error in params", func(t *testing.T) {
		def := &core.ZodTypeDef{}
		params := &core.SchemaParams{}
		ApplySchemaParams(def, params)
		assert.Nil(t, def.Error)
	})
}

func TestToDotPath(t *testing.T) {
	tests := []struct {
		name  string
		input []any
		want  string
	}{
		{"empty path", []any{}, ""},
		{"single string", []any{"user"}, "user"},
		{"string then int", []any{"users", 0}, "users[0]"},
		{"string then string", []any{"user", "name"}, "user.name"},
		{"special chars need brackets", []any{"user", "first-name"}, `user["first-name"]`},
		{"mixed path", []any{"users", 0, "profile", "address", 1}, "users[0].profile.address[1]"},
		{"string with spaces", []any{"user", "full name"}, `user["full name"]`},
		{"starts with digit", []any{"user", "123name"}, `user["123name"]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ToDotPath(tt.input))
		})
	}
}

func TestNeedsBracketNotation(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"name", false},
		{"firstName", false},
		{"first_name", false},
		{"name123", false},
		{"first-name", true},
		{"first name", true},
		{"user.name", true},
		{"123name", true},
		{"", false},
	}
	for _, tt := range tests {
		got := needsBracketNotation(tt.input)
		assert.Equal(t, tt.want, got)
	}
}

func TestIsIdentChar(t *testing.T) {
	valid := []rune{'a', 'z', 'A', 'Z', '0', '9', '_'}
	for _, c := range valid {
		assert.True(t, isIdentChar(c))
	}
	invalid := []rune{'-', ' ', '.', '!', '@', '[', '中'}
	for _, c := range invalid {
		assert.False(t, isIdentChar(c))
	}
}

func TestFormatErrorPath(t *testing.T) {
	path := []any{"users", 0, "name"}

	dot := FormatErrorPath(path, "dot")
	assert.Equal(t, "users[0].name", dot)

	bracket := FormatErrorPath(path, "bracket")
	assert.Equal(t, `["users"][0]["name"]`, bracket)

	// default style falls back to dot
	def := FormatErrorPath(path, "")
	assert.Equal(t, dot, def)

	// empty path
	assert.Equal(t, "", FormatErrorPath(nil, "dot"))
	assert.Equal(t, "", FormatErrorPath(nil, "bracket"))
}
