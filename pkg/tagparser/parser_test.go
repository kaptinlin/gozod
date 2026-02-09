package tagparser

import (
	"reflect"
	"testing"
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
			if err != nil {
				t.Errorf("ParseTagString() error = %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("ParseTagString() got %d rules, expected %d", len(result), len(tt.expected))
				return
			}

			for i, rule := range result {
				expected := tt.expected[i]
				if rule.Name != expected.Name {
					t.Errorf("Rule[%d].Name = %q, expected %q", i, rule.Name, expected.Name)
				}

				if len(rule.Params) != len(expected.Params) {
					t.Errorf("Rule[%d].Params length = %d, expected %d", i, len(rule.Params), len(expected.Params))
					continue
				}

				for j, param := range rule.Params {
					if param != expected.Params[j] {
						t.Errorf("Rule[%d].Params[%d] = %q, expected %q", i, j, param, expected.Params[j])
					}
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

	structType := reflect.TypeOf(TestStruct{})
	fields, err := parser.ParseStructTags(structType)

	if err != nil {
		t.Fatalf("ParseStructTags() error = %v", err)
	}

	// Should have 5 fields (Name, Email, Age, OptionalField, NoTagField)
	// unexportedField and IgnoredField should be excluded
	if len(fields) != 5 {
		t.Errorf("ParseStructTags() got %d fields, expected 5", len(fields))
	}

	// Verify Name field
	nameField := findField(fields, "Name")
	if nameField == nil {
		t.Error("Name field not found")
	} else {
		if nameField.JsonName != "name" {
			t.Errorf("Name field JsonName = %q, expected 'name'", nameField.JsonName)
		}
		if !nameField.Required {
			t.Error("Name field should be required")
		}
		if nameField.Optional {
			t.Error("Name field should not be optional")
		}
		if len(nameField.Rules) != 3 {
			t.Errorf("Name field has %d rules, expected 3", len(nameField.Rules))
		}
	}

	// Verify OptionalField
	optField := findField(fields, "OptionalField")
	if optField == nil {
		t.Error("OptionalField not found")
	} else {
		if optField.Required {
			t.Error("OptionalField should not be required")
		}
		if !optField.Optional {
			t.Error("OptionalField should be optional (pointer type)")
		}
		if len(optField.Rules) != 1 {
			t.Errorf("OptionalField has %d rules, expected 1", len(optField.Rules))
		}
	}

	// Verify NoTagField (field without gozod tag)
	noTagField := findField(fields, "NoTagField")
	if noTagField == nil {
		t.Error("NoTagField not found")
	} else {
		if noTagField.Required {
			t.Error("NoTagField should not be required")
		}
		if noTagField.Optional {
			t.Error("NoTagField should not be optional (no pointer, no rules)")
		}
		if len(noTagField.Rules) != 0 {
			t.Errorf("NoTagField has %d rules, expected 0", len(noTagField.Rules))
		}
	}
}

func TestTagParser_CustomTagName(t *testing.T) {
	parser := NewWithTagName("validate")

	type TestStruct struct {
		Name  string `validate:"required,min=2"`
		Email string `gozod:"required,email"` // Should be ignored with custom tag name
	}

	structType := reflect.TypeOf(TestStruct{})
	fields, err := parser.ParseStructTags(structType)

	if err != nil {
		t.Fatalf("ParseStructTags() error = %v", err)
	}

	// Should have 2 fields, but only Name should have rules
	if len(fields) != 2 {
		t.Errorf("ParseStructTags() got %d fields, expected 2", len(fields))
	}

	nameField := findField(fields, "Name")
	if nameField == nil {
		t.Error("Name field not found")
	} else {
		if len(nameField.Rules) != 2 {
			t.Errorf("Name field has %d rules, expected 2", len(nameField.Rules))
		}
		if !nameField.Required {
			t.Error("Name field should be required")
		}
	}

	emailField := findField(fields, "Email")
	if emailField == nil {
		t.Error("Email field not found")
	} else {
		if len(emailField.Rules) != 0 {
			t.Errorf("Email field has %d rules, expected 0 (wrong tag name)", len(emailField.Rules))
		}
		if emailField.Required {
			t.Error("Email field should not be required (wrong tag name)")
		}
	}
}

func TestTagParser_PointerToStruct(t *testing.T) {
	parser := New()

	type TestStruct struct {
		Name string `gozod:"required"`
	}

	// Test with pointer to struct
	structType := reflect.TypeOf(&TestStruct{})
	fields, err := parser.ParseStructTags(structType)

	if err != nil {
		t.Fatalf("ParseStructTags() error = %v", err)
	}

	if len(fields) != 1 {
		t.Errorf("ParseStructTags() got %d fields, expected 1", len(fields))
	}

	nameField := findField(fields, "Name")
	if nameField == nil {
		t.Error("Name field not found")
	} else if !nameField.Required {
		t.Error("Name field should be required")
	}
}

func TestTagParser_NonStructType(t *testing.T) {
	parser := New()

	// Test with non-struct type
	stringType := reflect.TypeOf("hello")
	fields, err := parser.ParseStructTags(stringType)

	if err != nil {
		t.Fatalf("ParseStructTags() error = %v", err)
	}

	if len(fields) != 0 {
		t.Errorf("ParseStructTags() got %d fields, expected 0 for non-struct", len(fields))
	}
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
			if result != tt.expected {
				t.Errorf("jsonFieldName() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestHasRule(t *testing.T) {
	rules := []TagRule{
		{Name: "required", Params: nil},
		{Name: "min", Params: []string{"2"}},
		{Name: "max", Params: []string{"50"}},
	}

	if !hasRule(rules, "required") {
		t.Error("hasRule should find 'required' rule")
	}

	if !hasRule(rules, "min") {
		t.Error("hasRule should find 'min' rule")
	}

	if hasRule(rules, "email") {
		t.Error("hasRule should not find 'email' rule")
	}

	if hasRule([]TagRule{}, "required") {
		t.Error("hasRule should return false for empty rules")
	}
}

func TestIsOptional(t *testing.T) {
	requiredField := reflect.StructField{
		Name: "RequiredField",
		Type: reflect.TypeOf(""),
	}
	if isOptional(requiredField, true, false) {
		t.Error("Required field should not be optional")
	}

	pointerField := reflect.StructField{
		Name: "PointerField",
		Type: reflect.TypeOf((*string)(nil)),
	}
	if !isOptional(pointerField, false, false) {
		t.Error("Pointer field should be optional by default")
	}

	if isOptional(pointerField, true, false) {
		t.Error("Required pointer field should not be optional")
	}

	nilableField := reflect.StructField{
		Name: "NilableField",
		Type: reflect.TypeOf(""),
	}
	if !isOptional(nilableField, false, true) {
		t.Error("Nilable field should be optional")
	}

	regularField := reflect.StructField{
		Name: "RegularField",
		Type: reflect.TypeOf(""),
	}
	if isOptional(regularField, false, false) {
		t.Error("Regular field should not be optional")
	}
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
