package utils

import (
	"testing"

	"github.com/kaptinlin/gozod/core"
)

func TestToErrorMap(t *testing.T) {
	t.Run("string input", func(t *testing.T) {
		m, ok := ToErrorMap("custom error")
		if !ok {
			t.Fatal("expected ok=true for string input")
		}
		if m == nil {
			t.Fatal("expected non-nil error map")
		}
		got := (*m)(core.ZodRawIssue{})
		if got != "custom error" {
			t.Errorf("got %q, want %q", got, "custom error")
		}
	})

	t.Run("ZodErrorMap input", func(t *testing.T) {
		fn := core.ZodErrorMap(func(core.ZodRawIssue) string {
			return "mapped"
		})
		m, ok := ToErrorMap(fn)
		if !ok {
			t.Fatal("expected ok=true")
		}
		if (*m)(core.ZodRawIssue{}) != "mapped" {
			t.Error("expected mapped result")
		}
	})

	t.Run("*ZodErrorMap input", func(t *testing.T) {
		fn := core.ZodErrorMap(func(core.ZodRawIssue) string {
			return "ptr"
		})
		m, ok := ToErrorMap(&fn)
		if !ok {
			t.Fatal("expected ok=true")
		}
		if (*m)(core.ZodRawIssue{}) != "ptr" {
			t.Error("expected ptr result")
		}
	})

	t.Run("func input", func(t *testing.T) {
		fn := func(core.ZodRawIssue) string { return "func" }
		m, ok := ToErrorMap(fn)
		if !ok {
			t.Fatal("expected ok=true")
		}
		if (*m)(core.ZodRawIssue{}) != "func" {
			t.Error("expected func result")
		}
	})

	t.Run("unsupported input", func(t *testing.T) {
		m, ok := ToErrorMap(42)
		if ok {
			t.Error("expected ok=false for int input")
		}
		if m != nil {
			t.Error("expected nil map")
		}
	})
}

func TestFirstParam(t *testing.T) {
	if got := FirstParam(); got != nil {
		t.Errorf("FirstParam() = %v, want nil", got)
	}
	if got := FirstParam("hello"); got != "hello" {
		t.Errorf("FirstParam(hello) = %v, want hello", got)
	}
	if got := FirstParam(1, 2, 3); got != 1 {
		t.Errorf("FirstParam(1,2,3) = %v, want 1", got)
	}
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
		if got := OriginFromValue(tt.input); got != tt.want {
			t.Errorf("OriginFromValue(%v) = %q, want %q", tt.input, got, tt.want)
		}
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
		if got := NumericOrigin(tt.input); got != tt.want {
			t.Errorf("NumericOrigin(%v) = %q, want %q", tt.input, got, tt.want)
		}
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
		if got := SizableOrigin(tt.input); got != tt.want {
			t.Errorf("SizableOrigin(%v) = %q, want %q", tt.input, got, tt.want)
		}
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
		if got := LengthableOrigin(tt.input); got != tt.want {
			t.Errorf("LengthableOrigin(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCompareValues(t *testing.T) {
	t.Run("both nil", func(t *testing.T) {
		if got := CompareValues(nil, nil); got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})
	t.Run("first nil", func(t *testing.T) {
		if got := CompareValues(nil, 1); got != -1 {
			t.Errorf("got %d, want -1", got)
		}
	})
	t.Run("second nil", func(t *testing.T) {
		if got := CompareValues(1, nil); got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})
	t.Run("int equal", func(t *testing.T) {
		if got := CompareValues(5, 5); got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})
	t.Run("int less", func(t *testing.T) {
		if got := CompareValues(3, 5); got != -1 {
			t.Errorf("got %d, want -1", got)
		}
	})
	t.Run("int greater", func(t *testing.T) {
		if got := CompareValues(5, 3); got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})
	t.Run("int64", func(t *testing.T) {
		if got := CompareValues(int64(1), int64(2)); got != -1 {
			t.Errorf("got %d, want -1", got)
		}
	})
	t.Run("float64", func(t *testing.T) {
		if got := CompareValues(1.5, 2.5); got != -1 {
			t.Errorf("got %d, want -1", got)
		}
	})
	t.Run("float32", func(t *testing.T) {
		a, b := float32(3.0), float32(1.0)
		if got := CompareValues(a, b); got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})
	t.Run("string", func(t *testing.T) {
		if got := CompareValues("a", "b"); got != -1 {
			t.Errorf("got %d, want -1", got)
		}
	})
	t.Run("mismatched types", func(t *testing.T) {
		if got := CompareValues(1, "a"); got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})
	t.Run("pointer deref", func(t *testing.T) {
		a, b := 10, 5
		if got := CompareValues(&a, &b); got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})
}

func TestNormalizeParams(t *testing.T) {
	t.Run("no args", func(t *testing.T) {
		p := NormalizeParams()
		if p == nil || p.Error != nil {
			t.Error("expected empty SchemaParams")
		}
	})
	t.Run("nil arg", func(t *testing.T) {
		p := NormalizeParams(nil)
		if p == nil || p.Error != nil {
			t.Error("expected empty SchemaParams")
		}
	})
	t.Run("string arg", func(t *testing.T) {
		p := NormalizeParams("err msg")
		if p.Error != "err msg" {
			t.Errorf("got %v, want err msg", p.Error)
		}
	})
	t.Run("SchemaParams value", func(t *testing.T) {
		sp := core.SchemaParams{Error: "val"}
		p := NormalizeParams(sp)
		if p.Error != "val" {
			t.Errorf("got %v, want val", p.Error)
		}
	})
	t.Run("*SchemaParams", func(t *testing.T) {
		sp := &core.SchemaParams{Error: "ptr"}
		p := NormalizeParams(sp)
		if p.Error != "ptr" {
			t.Errorf("got %v, want ptr", p.Error)
		}
		// verify copy semantics
		sp.Error = "changed"
		if p.Error != "ptr" {
			t.Error("expected copy, not reference")
		}
	})
	t.Run("nil *SchemaParams", func(t *testing.T) {
		var sp *core.SchemaParams
		p := NormalizeParams(sp)
		if p == nil || p.Error != nil {
			t.Error("expected empty SchemaParams")
		}
	})
	t.Run("unsupported type", func(t *testing.T) {
		p := NormalizeParams(42)
		if p == nil || p.Error != nil {
			t.Error("expected empty SchemaParams")
		}
	})
}

func TestNormalizeCustomParams(t *testing.T) {
	t.Run("no args", func(t *testing.T) {
		p := NormalizeCustomParams()
		if p == nil || p.Error != nil {
			t.Error("expected empty CustomParams")
		}
	})
	t.Run("nil arg", func(t *testing.T) {
		p := NormalizeCustomParams(nil)
		if p == nil || p.Error != nil {
			t.Error("expected empty CustomParams")
		}
	})
	t.Run("string arg", func(t *testing.T) {
		p := NormalizeCustomParams("err")
		if p.Error != "err" {
			t.Errorf("got %v, want err", p.Error)
		}
	})
	t.Run("CustomParams value", func(t *testing.T) {
		cp := core.CustomParams{Error: "val"}
		p := NormalizeCustomParams(cp)
		if p.Error != "val" {
			t.Errorf("got %v, want val", p.Error)
		}
	})
	t.Run("*CustomParams", func(t *testing.T) {
		cp := &core.CustomParams{Error: "ptr"}
		p := NormalizeCustomParams(cp)
		if p.Error != "ptr" {
			t.Errorf("got %v, want ptr", p.Error)
		}
	})
	t.Run("nil *CustomParams", func(t *testing.T) {
		var cp *core.CustomParams
		p := NormalizeCustomParams(cp)
		if p == nil || p.Error != nil {
			t.Error("expected empty CustomParams")
		}
	})
	t.Run("any as error", func(t *testing.T) {
		p := NormalizeCustomParams(42)
		if p.Error != 42 {
			t.Errorf("got %v, want 42", p.Error)
		}
	})
}

func TestApplySchemaParams(t *testing.T) {
	t.Run("nil params", func(t *testing.T) {
		def := &core.ZodTypeDef{}
		ApplySchemaParams(def, nil)
		if def.Error != nil {
			t.Error("expected nil error")
		}
	})
	t.Run("with error string", func(t *testing.T) {
		def := &core.ZodTypeDef{}
		params := &core.SchemaParams{Error: "bad"}
		ApplySchemaParams(def, params)
		if def.Error == nil {
			t.Fatal("expected non-nil error map")
		}
		got := (*def.Error)(core.ZodRawIssue{})
		if got != "bad" {
			t.Errorf("got %q, want %q", got, "bad")
		}
	})
	t.Run("nil error in params", func(t *testing.T) {
		def := &core.ZodTypeDef{}
		params := &core.SchemaParams{}
		ApplySchemaParams(def, params)
		if def.Error != nil {
			t.Error("expected nil error")
		}
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
			if got := ToDotPath(tt.input); got != tt.want {
				t.Errorf("ToDotPath(%v) = %q, want %q", tt.input, got, tt.want)
			}
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
		if got != tt.want {
			t.Errorf("needsBracketNotation(%q) = %v, want %v",
				tt.input, got, tt.want)
		}
	}
}

func TestIsIdentChar(t *testing.T) {
	valid := []rune{'a', 'z', 'A', 'Z', '0', '9', '_'}
	for _, c := range valid {
		if !isIdentChar(c) {
			t.Errorf("isIdentChar(%q) = false, want true", c)
		}
	}
	invalid := []rune{'-', ' ', '.', '!', '@', '[', '中'}
	for _, c := range invalid {
		if isIdentChar(c) {
			t.Errorf("isIdentChar(%q) = true, want false", c)
		}
	}
}

func TestFormatErrorPath(t *testing.T) {
	path := []any{"users", 0, "name"}

	dot := FormatErrorPath(path, "dot")
	if dot != "users[0].name" {
		t.Errorf("dot style: got %q, want %q", dot, "users[0].name")
	}

	bracket := FormatErrorPath(path, "bracket")
	want := `["users"][0]["name"]`
	if bracket != want {
		t.Errorf("bracket style: got %q, want %q", bracket, want)
	}

	// default style falls back to dot
	def := FormatErrorPath(path, "")
	if def != dot {
		t.Errorf("default style: got %q, want %q", def, dot)
	}

	// empty path
	if got := FormatErrorPath(nil, "dot"); got != "" {
		t.Errorf("empty dot: got %q, want empty", got)
	}
	if got := FormatErrorPath(nil, "bracket"); got != "" {
		t.Errorf("empty bracket: got %q, want empty", got)
	}
}
