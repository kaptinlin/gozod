package tagparser

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJSONFieldName(t *testing.T) {
	tests := []struct {
		name     string
		field    reflect.StructField
		expected JSONField
	}{
		{
			name: "no json tag",
			field: reflect.StructField{
				Name: "TestField",
				Tag:  "",
			},
			expected: JSONField{Name: "TestField"},
		},
		{
			name: "json tag with name",
			field: reflect.StructField{
				Name: "TestField",
				Tag:  `json:"test_field"`,
			},
			expected: JSONField{Name: "test_field"},
		},
		{
			name: "json tag with omitempty",
			field: reflect.StructField{
				Name: "TestField",
				Tag:  `json:"test_field,omitempty"`,
			},
			expected: JSONField{Name: "test_field"},
		},
		{
			name: "json tag with dash (skip)",
			field: reflect.StructField{
				Name: "TestField",
				Tag:  `json:"-"`,
			},
			expected: JSONField{Skip: true},
		},
		{
			name: "json tag omitempty only",
			field: reflect.StructField{
				Name: "TestField",
				Tag:  `json:",omitempty"`,
			},
			expected: JSONField{Name: "TestField"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JSONFieldName(tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasRule(t *testing.T) {
	rules := []TagRule{
		{Name: "required", Params: nil},
		{Name: "min", Params: []string{"2"}},
		{Name: "max", Params: []string{"50"}},
	}

	assert.True(t, hasRule(rules, "required"), "hasRule should find 'required' rule")
	assert.True(t, hasRule(rules, "min"), "hasRule should find 'min' rule")
	assert.False(t, hasRule(rules, "email"), "hasRule should not find 'email' rule")
	assert.False(t, hasRule([]TagRule{}, "required"), "hasRule should return false for empty rules")
}

func TestIsOptional(t *testing.T) {
	requiredField := reflect.StructField{
		Name: "RequiredField",
		Type: reflect.TypeFor[string](),
	}
	assert.False(t, isOptional(requiredField, true, false), "Required field should not be optional")

	pointerField := reflect.StructField{
		Name: "PointerField",
		Type: reflect.TypeFor[*string](),
	}
	assert.True(t, isOptional(pointerField, false, false), "Pointer field should be optional by default")
	assert.False(t, isOptional(pointerField, true, false), "Required pointer field should not be optional")

	nilableField := reflect.StructField{
		Name: "NilableField",
		Type: reflect.TypeFor[string](),
	}
	assert.True(t, isOptional(nilableField, false, true), "Nilable field should be optional")

	regularField := reflect.StructField{
		Name: "RegularField",
		Type: reflect.TypeFor[string](),
	}
	assert.False(t, isOptional(regularField, false, false), "Regular field should not be optional")
}

func TestFieldInfoHelpers(t *testing.T) {
	info := FieldInfo{
		Type:     reflect.TypeFor[string](),
		TypeName: "MyString",
		Rules: []TagRule{
			{Name: "coerce"},
			{Name: "trim"},
			{Name: "regex", Params: []string{"^a$"}},
			{Name: "url"},
			{Name: "ipv4"},
			{Name: "refine"},
			{Name: "uuid"},
			{Name: "enum", Params: []string{"a", "b"}},
		},
	}

	assert.True(t, info.HasRules())
	assert.True(t, info.HasRule("uuid"))
	assert.False(t, info.IsPointerType())
	assert.True(t, info.HasCoerceRule())
	assert.Equal(t, "MyString", info.EffectiveTypeName())
	assert.True(t, info.HasSchemaSpec())
	assert.True(t, info.NeedsOptionalModifier())
	assert.True(t, info.NeedsGeneratedOptional())
	if assert.NotNil(t, info.EnumRule()) {
		assert.Equal(t, []string{"a", "b"}, info.EnumRule().Params)
	}
	assert.Len(t, info.RulesExcept("uuid", "enum"), 6)
	assert.Len(t, info.ValidationRules(), 7)
	assert.Len(t, info.ValidationRulesExcept("uuid", "enum"), 5)
	assert.ElementsMatch(t, []string{
		"strings",
		"regexp",
		"net/url",
		"net",
		"github.com/kaptinlin/gozod/core",
	}, info.RequiredImports())
	assert.True(t, info.NeedsStringsImport())
	assert.True(t, info.NeedsRegexpImport())
	assert.True(t, info.NeedsNetURLImport())
	assert.True(t, info.NeedsNetImport())
	assert.True(t, info.NeedsCoreImport())
	assert.True(t, info.IsUUIDStringField())
	assert.True(t, info.IsEnumStringField())
	assert.False(t, info.NeedsPointerNilable())
	assert.False(t, info.NeedsPointerOptional())

	info.Type = reflect.TypeFor[*string]()
	assert.False(t, info.NeedsOptionalModifier())
	assert.True(t, info.NeedsGeneratedOptional())
	assert.True(t, info.NeedsPointerNilable())
	assert.True(t, info.NeedsPointerOptional())

	info = FieldInfo{
		Type:     reflect.TypeFor[string](),
		Required: true,
		Rules: []TagRule{
			{Name: "required"},
			{Name: "min", Params: []string{"1"}},
			{Name: "coerce"},
		},
	}
	assert.False(t, info.NeedsGeneratedOptional())
	assert.Equal(t, []TagRule{{Name: "min", Params: []string{"1"}}}, info.ValidationRules())
	assert.Equal(t, []TagRule{{Name: "min", Params: []string{"1"}}}, info.ValidationRulesExcept("uuid"))

	timeInfo := FieldInfo{Type: reflect.TypeFor[struct{ When string }]()}
	assert.False(t, timeInfo.UsesTimeImport())
	assert.Empty(t, timeInfo.RequiredImports())

	timeField := FieldInfo{Type: reflect.TypeFor[time.Time]()}
	assert.True(t, timeField.UsesTimeImport())
	assert.Equal(t, []string{"time"}, timeField.RequiredImports())
}
