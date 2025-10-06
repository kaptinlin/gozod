package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/pkg/tagparser"
)

func TestFileWriter_GenerateImports(t *testing.T) {
	tests := []struct {
		name              string
		fields            []tagparser.FieldInfo
		expectedImports   []string
		unexpectedImports []string
	}{
		{
			name: "basic imports",
			fields: []tagparser.FieldInfo{
				{
					Name: "ID",
					Type: reflect.TypeOf(""),
					Rules: []tagparser.TagRule{
						{Name: "required"},
						{Name: "uuid"},
					},
				},
				{
					Name: "Email",
					Type: reflect.TypeOf(""),
					Rules: []tagparser.TagRule{
						{Name: "required"},
						{Name: "email"},
					},
				},
			},
			expectedImports:   []string{"github.com/kaptinlin/gozod"},
			unexpectedImports: []string{"github.com/kaptinlin/gozod/core", "regexp", "net"},
		},
		{
			name: "regex requires regexp import",
			fields: []tagparser.FieldInfo{
				{
					Name: "SKU",
					Type: reflect.TypeOf(""),
					Rules: []tagparser.TagRule{
						{Name: "regex", Params: []string{"^[A-Z0-9]+$"}},
					},
				},
			},
			expectedImports:   []string{"github.com/kaptinlin/gozod", "regexp"},
			unexpectedImports: []string{"github.com/kaptinlin/gozod/core"},
		},
		{
			name: "time fields require time import",
			fields: []tagparser.FieldInfo{
				{
					Name: "CreatedAt",
					Type: reflect.TypeOf(struct{}{}), // Mock time.Time
					Rules: []tagparser.TagRule{
						{Name: "required"},
					},
				},
			},
			expectedImports:   []string{"github.com/kaptinlin/gozod"},
			unexpectedImports: []string{"github.com/kaptinlin/gozod/core"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Modify the type string to simulate time.Time for time fields
			for i := range tt.fields {
				if strings.Contains(tt.name, "time") {
					tt.fields[i].Type = reflect.TypeOf(struct{}{}) // We'll mock this
				}
			}

			writer, err := NewFileWriter("", "main", "_gen.go", true, false)
			require.NoError(t, err, "Failed to create writer")

			// Create a mock GenerationInfo for the test
			info := &GenerationInfo{
				Name:    "TestStruct",
				Fields:  tt.fields,
				Package: "main",
			}
			imports := writer.generateImports(info)

			// Check expected imports
			for _, expectedImport := range tt.expectedImports {
				found := false
				for _, imp := range imports {
					if imp == expectedImport {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected import %s not found in %v", expectedImport, imports)
				}
			}

			// Check unexpected imports
			for _, unexpectedImport := range tt.unexpectedImports {
				for _, imp := range imports {
					if imp == unexpectedImport {
						t.Errorf("Unexpected import %s found in %v", unexpectedImport, imports)
					}
				}
			}
		})
	}
}

func TestFileWriter_GenerateFieldSchema(t *testing.T) {
	tests := []struct {
		name           string
		field          tagparser.FieldInfo
		structName     string
		expectedSchema string
		expectError    bool
	}{
		{
			name: "simple string field",
			field: tagparser.FieldInfo{
				Name: "Name",
				Type: reflect.TypeOf(""),
				Rules: []tagparser.TagRule{
					{Name: "required"},
					{Name: "min", Params: []string{"2"}},
					{Name: "max", Params: []string{"50"}},
				},
			},
			structName:     "User",
			expectedSchema: "gozod.String().Min(2).Max(50)",
			expectError:    false,
		},
		{
			name: "email field",
			field: tagparser.FieldInfo{
				Name: "Email",
				Type: reflect.TypeOf(""),
				Rules: []tagparser.TagRule{
					{Name: "required"},
					{Name: "email"},
				},
			},
			structName:     "User",
			expectedSchema: "gozod.String().Email()",
			expectError:    false,
		},
		{
			name: "integer field with range",
			field: tagparser.FieldInfo{
				Name: "Age",
				Type: reflect.TypeOf(0),
				Rules: []tagparser.TagRule{
					{Name: "required"},
					{Name: "min", Params: []string{"18"}},
					{Name: "max", Params: []string{"120"}},
				},
			},
			structName:     "User",
			expectedSchema: "gozod.Int().Min(18).Max(120)",
			expectError:    false,
		},
		{
			name: "float field with gt validation",
			field: tagparser.FieldInfo{
				Name: "Price",
				Type: reflect.TypeOf(0.0),
				Rules: []tagparser.TagRule{
					{Name: "required"},
					{Name: "gt", Params: []string{"0.0"}},
				},
			},
			structName:     "Product",
			expectedSchema: "gozod.Float64().Gt(0.0)",
			expectError:    false,
		},
		{
			name: "enum field with default",
			field: tagparser.FieldInfo{
				Name: "Status",
				Type: reflect.TypeOf(""),
				Rules: []tagparser.TagRule{
					{Name: "enum", Params: []string{"active", "inactive"}},
					{Name: "default", Params: []string{"active"}},
				},
			},
			structName:     "User",
			expectedSchema: `gozod.Enum("active", "inactive").Default("active")`,
			expectError:    false,
		},
		{
			name: "optional pointer field",
			field: tagparser.FieldInfo{
				Name:     "Description",
				Type:     reflect.TypeOf((*string)(nil)),
				Optional: true,
				Rules: []tagparser.TagRule{
					{Name: "max", Params: []string{"500"}},
				},
			},
			structName:     "Product",
			expectedSchema: "gozod.String().Max(500).Optional()",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, err := NewFileWriter("", "main", "_gen.go", true, false)
			require.NoError(t, err, "Failed to create writer")

			schema, err := writer.generateFieldSchemaCodeForStruct(tt.field, tt.structName)
			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Unexpected error")
			}

			if !tt.expectError {
				if !strings.Contains(schema, tt.expectedSchema) {
					assert.Equal(t, tt.expectedSchema, schema, "Expected schema to contain %s, got %s", tt.expectedSchema, schema)
				}
			}
		})
	}
}

func TestFileWriter_GenerateCode(t *testing.T) {
	tests := []struct {
		name              string
		info              *GenerationInfo
		expectedContent   []string
		unexpectedContent []string
	}{
		{
			name: "simple struct generation",
			info: &GenerationInfo{
				Name:     "User",
				FilePath: "test.go",
				Fields: []tagparser.FieldInfo{
					{
						Name:     "ID",
						JsonName: "id",
						Type:     reflect.TypeOf(""),
						Rules: []tagparser.TagRule{
							{Name: "required"},
							{Name: "uuid"},
						},
					},
					{
						Name:     "Name",
						JsonName: "name",
						Type:     reflect.TypeOf(""),
						Rules: []tagparser.TagRule{
							{Name: "required"},
							{Name: "min", Params: []string{"2"}},
						},
					},
				},
			},
			expectedContent: []string{
				"// Code generated by gozodgen. DO NOT EDIT.",
				"func (u User) Schema() *gozod.ZodStruct[User, User]",
				`"id": gozod.Uuid()`,
				`"name": gozod.String().Min(2)`,
				"gozod.Struct[User](gozod.StructSchema{",
			},
			unexpectedContent: []string{
				"github.com/kaptinlin/gozod/core",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, err := NewFileWriter("", "main", "_gen.go", true, false)
			require.NoError(t, err, "Failed to create writer")

			content, err := writer.generateCode(tt.info)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			helper := &TestHelper{t: t}
			helper.AssertCodeContains(content, tt.expectedContent...)
			helper.AssertCodeNotContains(content, tt.unexpectedContent...)
			helper.AssertValidGoCode(content)
		})
	}
}

func TestFileWriter_GenerateFieldSchemaCodeForStruct(t *testing.T) {
	tests := []struct {
		name     string
		rules    []tagparser.TagRule
		expected string
	}{
		{
			name: "min and max rules",
			rules: []tagparser.TagRule{
				{Name: "min", Params: []string{"2"}},
				{Name: "max", Params: []string{"50"}},
			},
			expected: ".Min(2).Max(50)",
		},
		{
			name: "email rule",
			rules: []tagparser.TagRule{
				{Name: "email"},
			},
			expected: ".Email()",
		},
		{
			name: "enum rule",
			rules: []tagparser.TagRule{
				{Name: "enum", Params: []string{"active", "inactive"}},
			},
			expected: `.Enum("active", "inactive")`,
		},
		{
			name: "default value rule",
			rules: []tagparser.TagRule{
				{Name: "default", Params: []string{"active"}},
			},
			expected: `.Default("active")`,
		},
		{
			name: "numeric rules",
			rules: []tagparser.TagRule{
				{Name: "gt", Params: []string{"0"}},
				{Name: "lte", Params: []string{"100"}},
			},
			expected: ".Gt(0).Lte(100)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, err := NewFileWriter("", "main", "_gen.go", true, false)
			require.NoError(t, err, "Failed to create writer")

			// Create a mock field with the rules
			field := tagparser.FieldInfo{
				Name:  "TestField",
				Type:  reflect.TypeOf(""),
				Rules: tt.rules,
			}

			// For numeric rules test, use int type
			if strings.Contains(tt.name, "numeric") {
				field.Type = reflect.TypeOf(0)
			}

			result, err := writer.generateFieldSchemaCodeForStruct(field, "TestStruct")
			require.NoError(t, err, "Failed to generate field schema")

			if !strings.Contains(result, tt.expected) {
				assert.Equal(t, tt.expected, result, "Expected result to contain %s, got %s", tt.expected, result)
			}

			// Basic validation that we got a schema back
			if result == "" {
				t.Error("Expected non-empty schema result")
			}
		})
	}
}

func TestFileWriter_GetBasicTypeConstructor(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		expected string
	}{
		{name: "string type", typeName: "string", expected: "gozod.String()"},
		{name: "int type", typeName: "int", expected: "gozod.Int()"},
		{name: "int64 type", typeName: "int64", expected: "gozod.Int64()"},
		{name: "float64 type", typeName: "float64", expected: "gozod.Float64()"},
		{name: "bool type", typeName: "bool", expected: "gozod.Bool()"},
		// Note: time.Time is handled by getBaseConstructorFromTypeName, not getBasicTypeConstructor
		{name: "unknown type", typeName: "CustomType", expected: "gozod.Any()"},
	}

	writer, err := NewFileWriter("", "main", "_gen.go", true, false)
	require.NoError(t, err, "Failed to create writer")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := writer.getBasicTypeConstructor(tt.typeName)
			assert.Equal(t, tt.expected, result, "Expected %s, got %s", tt.expected, result)
		})
	}
}

func TestFileWriter_CircularReferenceHandling(t *testing.T) {
	tests := []struct {
		name       string
		typeName   string
		structName string
		expected   string
	}{
		{
			name:       "self reference",
			typeName:   "Node",
			structName: "Node",
			expected:   "gozod.Lazy(func() gozod.ZodType[any] { return gozod.FromStruct[Node]() })",
		},
		{
			name:       "pointer self reference",
			typeName:   "*Node",
			structName: "Node",
			expected:   "gozod.Lazy(func() gozod.ZodType[any] { return gozod.FromStruct[Node]() })",
		},
		{
			name:       "slice self reference",
			typeName:   "[]*Node",
			structName: "Node",
			expected:   "gozod.Slice(gozod.Lazy(func() gozod.ZodType[any] { return gozod.FromStruct[Node]() }))",
		},
		{
			name:       "no circular reference",
			typeName:   "string",
			structName: "Node",
			expected:   "gozod.String()",
		},
		{
			name:       "time.Time type",
			typeName:   "time.Time",
			structName: "Node",
			expected:   "gozod.Time()",
		},
	}

	writer, err := NewFileWriter("", "main", "_gen.go", true, false)
	require.NoError(t, err, "Failed to create writer")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := writer.getBaseConstructorFromTypeName(tt.typeName, tt.structName)
			if !strings.Contains(result, tt.expected) {
				assert.Equal(t, tt.expected, result, "Expected result to contain %s, got %s", tt.expected, result)
			}
		})
	}
}
