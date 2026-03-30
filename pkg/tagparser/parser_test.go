package tagparser

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagParser_ParseTagString(t *testing.T) {
	parser := New()

	tests := []struct {
		name     string
		tag      string
		expected []TagRule
	}{
		{
			name:     "empty tag",
			tag:      "",
			expected: []TagRule{},
		},
		{
			name: "simple rule without params",
			tag:  "required",
			expected: []TagRule{
				{Name: "required", Params: nil},
			},
		},
		{
			name: "rule with single param",
			tag:  "min=2",
			expected: []TagRule{
				{Name: "min", Params: []string{"2"}},
			},
		},
		{
			name: "multiple rules",
			tag:  "required,min=2,max=50",
			expected: []TagRule{
				{Name: "required", Params: nil},
				{Name: "min", Params: []string{"2"}},
				{Name: "max", Params: []string{"50"}},
			},
		},
		{
			name: "rule with multiple params",
			tag:  "enum=red green blue",
			expected: []TagRule{
				{Name: "enum", Params: []string{"red", "green", "blue"}},
			},
		},
		{
			name: "rule with quoted param",
			tag:  "regex='^[A-Za-z0-9_]+$'",
			expected: []TagRule{
				{Name: "regex", Params: []string{"^[A-Za-z0-9_]+$"}},
			},
		},
		{
			name: "escaped comma in param",
			tag:  "custom='hello\\, world'",
			expected: []TagRule{
				{Name: "custom", Params: []string{"hello, world"}},
			},
		},
		{
			name: "complex mix of rules",
			tag:  "required,min=2,max=50,regex='^[A-Z]+$',enum=ACTIVE INACTIVE",
			expected: []TagRule{
				{Name: "required", Params: nil},
				{Name: "min", Params: []string{"2"}},
				{Name: "max", Params: []string{"50"}},
				{Name: "regex", Params: []string{"^[A-Z]+$"}},
				{Name: "enum", Params: []string{"ACTIVE", "INACTIVE"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseTagString(tt.tag)
			require.NoError(t, err)

			require.Len(t, result, len(tt.expected))

			for i, rule := range result {
				expected := tt.expected[i]
				assert.Equal(t, expected.Name, rule.Name)

				require.Len(t, rule.Params, len(expected.Params))

				for j, param := range rule.Params {
					assert.Equal(t, expected.Params[j], param)
				}
			}
		})
	}
}

func TestTagParser_ParseStructTags(t *testing.T) {
	parser := New()

	type TestStruct struct {
		Name          string  `gozod:"required,min=2,max=50" json:"name"`
		Email         string  `gozod:"required,email" json:"email"`
		Age           int     `gozod:"required,min=18,max=120" json:"age"`
		OptionalField *string `gozod:"max=100" json:"optional_field"`
		IgnoredField  string  `gozod:"-" json:"ignored"`
		NoTagField    string  `json:"no_tag"`
	}

	structType := reflect.TypeFor[TestStruct]()
	fields, err := parser.ParseStructTags(structType)

	require.NoError(t, err)

	// Should have 5 fields (Name, Email, Age, OptionalField, NoTagField)
	// unexportedField and IgnoredField should be excluded
	assert.Len(t, fields, 5)

	// Verify Name field
	nameField := findField(fields, "Name")
	require.NotNil(t, nameField, "Name field not found")
	assert.Equal(t, "name", nameField.JSONName)
	assert.True(t, nameField.Required, "Name field should be required")
	assert.False(t, nameField.Optional, "Name field should not be optional")
	assert.Len(t, nameField.Rules, 3)

	// Verify OptionalField
	optField := findField(fields, "OptionalField")
	require.NotNil(t, optField, "OptionalField not found")
	assert.False(t, optField.Required, "OptionalField should not be required")
	assert.True(t, optField.Optional, "OptionalField should be optional (pointer type)")
	assert.Len(t, optField.Rules, 1)

	// Verify NoTagField (field without gozod tag)
	noTagField := findField(fields, "NoTagField")
	require.NotNil(t, noTagField, "NoTagField not found")
	assert.False(t, noTagField.Required, "NoTagField should not be required")
	assert.False(t, noTagField.Optional, "NoTagField should not be optional (no pointer, no rules)")
	assert.Len(t, noTagField.Rules, 0)
}

func TestTagParser_CustomTagName(t *testing.T) {
	parser := NewWithTagName("validate")

	type TestStruct struct {
		Name  string `validate:"required,min=2"`
		Email string `gozod:"required,email"` // Should be ignored with custom tag name
	}

	structType := reflect.TypeFor[TestStruct]()
	fields, err := parser.ParseStructTags(structType)

	require.NoError(t, err)

	// Should have 2 fields, but only Name should have rules
	assert.Len(t, fields, 2)

	nameField := findField(fields, "Name")
	require.NotNil(t, nameField, "Name field not found")
	assert.Len(t, nameField.Rules, 2)
	assert.True(t, nameField.Required, "Name field should be required")

	emailField := findField(fields, "Email")
	require.NotNil(t, emailField, "Email field not found")
	assert.Len(t, emailField.Rules, 0, "Email field has rules, expected 0 (wrong tag name)")
	assert.False(t, emailField.Required, "Email field should not be required (wrong tag name)")
}

func TestTagParser_PointerToStruct(t *testing.T) {
	parser := New()

	type TestStruct struct {
		Name string `gozod:"required"`
	}

	// Test with pointer to struct
	structType := reflect.TypeFor[*TestStruct]()
	fields, err := parser.ParseStructTags(structType)

	require.NoError(t, err)

	assert.Len(t, fields, 1)

	nameField := findField(fields, "Name")
	require.NotNil(t, nameField, "Name field not found")
	assert.True(t, nameField.Required, "Name field should be required")
}

func TestTagParser_NonStructType(t *testing.T) {
	parser := New()

	// Test with non-struct type
	stringType := reflect.TypeFor[string]()
	fields, err := parser.ParseStructTags(stringType)

	require.NoError(t, err)

	assert.Len(t, fields, 0)
}

func TestJSONFieldName(t *testing.T) {
	tests := []struct {
		name     string
		field    reflect.StructField
		expected string
	}{
		{
			name: "no json tag",
			field: reflect.StructField{
				Name: "TestField",
				Tag:  "",
			},
			expected: "TestField",
		},
		{
			name: "json tag with name",
			field: reflect.StructField{
				Name: "TestField",
				Tag:  `json:"test_field"`,
			},
			expected: "test_field",
		},
		{
			name: "json tag with omitempty",
			field: reflect.StructField{
				Name: "TestField",
				Tag:  `json:"test_field,omitempty"`,
			},
			expected: "test_field",
		},
		{
			name: "json tag with dash (skip)",
			field: reflect.StructField{
				Name: "TestField",
				Tag:  `json:"-"`,
			},
			expected: "TestField",
		},
		{
			name: "json tag omitempty only",
			field: reflect.StructField{
				Name: "TestField",
				Tag:  `json:",omitempty"`,
			},
			expected: "TestField",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jsonFieldName(tt.field)
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

// Helper function to find a field by name
func findField(fields []FieldInfo, name string) *FieldInfo {
	for i := range fields {
		if fields[i].Name == name {
			return &fields[i]
		}
	}
	return nil
}
