package tagparser_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/pkg/tagparser"
)

func TestTagParser_ParseTagString(t *testing.T) {
	parser := tagparser.New()

	tests := []struct {
		name     string
		tag      string
		expected []tagparser.TagRule
	}{
		{
			name:     "empty tag",
			tag:      "",
			expected: []tagparser.TagRule{},
		},
		{
			name: "simple rule without params",
			tag:  "required",
			expected: []tagparser.TagRule{
				{Name: "required", Params: nil},
			},
		},
		{
			name: "rule with single param",
			tag:  "min=2",
			expected: []tagparser.TagRule{
				{Name: "min", Params: []string{"2"}},
			},
		},
		{
			name: "multiple rules",
			tag:  "required,min=2,max=50",
			expected: []tagparser.TagRule{
				{Name: "required", Params: nil},
				{Name: "min", Params: []string{"2"}},
				{Name: "max", Params: []string{"50"}},
			},
		},
		{
			name: "rule with multiple params",
			tag:  "enum=red green blue",
			expected: []tagparser.TagRule{
				{Name: "enum", Params: []string{"red", "green", "blue"}},
			},
		},
		{
			name: "rule with quoted param",
			tag:  "regex='^[A-Za-z0-9_]+$'",
			expected: []tagparser.TagRule{
				{Name: "regex", Params: []string{"^[A-Za-z0-9_]+$"}},
			},
		},
		{
			name: "escaped comma in param",
			tag:  "custom='hello\\, world'",
			expected: []tagparser.TagRule{
				{Name: "custom", Params: []string{"hello, world"}},
			},
		},
		{
			name: "complex mix of rules",
			tag:  "required,min=2,max=50,regex='^[A-Z]+$',enum=ACTIVE INACTIVE",
			expected: []tagparser.TagRule{
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
	parser := tagparser.New()

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
	parser := tagparser.NewWithTagName("validate")

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
	parser := tagparser.New()

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
	parser := tagparser.New()

	// Test with non-struct type
	stringType := reflect.TypeFor[string]()
	fields, err := parser.ParseStructTags(stringType)

	require.NoError(t, err)

	assert.Len(t, fields, 0)
	assert.NotNil(t, fields)
}

func TestTagParser_ParseStructTagsStrict(t *testing.T) {
	parser := tagparser.New()

	t.Run("struct input", func(t *testing.T) {
		type TestStruct struct {
			Name string `gozod:"required"`
		}

		fields, err := parser.ParseStructTagsStrict(reflect.TypeFor[TestStruct]())
		require.NoError(t, err)
		require.Len(t, fields, 1)
		assert.Equal(t, "Name", fields[0].Name)
	})

	t.Run("non struct input", func(t *testing.T) {
		fields, err := parser.ParseStructTagsStrict(reflect.TypeFor[string]())
		require.ErrorIs(t, err, tagparser.ErrTypeMustBeStruct)
		assert.Nil(t, fields)
	})
}

func TestNewWithTagName_DefaultFallback(t *testing.T) {
	parser := tagparser.NewWithTagName("  ")

	type TestStruct struct {
		Name string `gozod:"required"`
	}

	fields, err := parser.ParseStructTags(reflect.TypeFor[TestStruct]())
	require.NoError(t, err)
	require.Len(t, fields, 1)
	assert.True(t, fields[0].Required)
}

func findField(fields []tagparser.FieldInfo, name string) *tagparser.FieldInfo {
	for i := range fields {
		if fields[i].Name == name {
			return &fields[i]
		}
	}
	return nil
}
