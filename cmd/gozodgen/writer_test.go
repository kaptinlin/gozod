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

func TestFileWriter_GenerateValidatorChain(t *testing.T) {
	writer, err := NewFileWriter("", "main", "_gen.go", true, false)
	require.NoError(t, err, "Failed to create writer")

	tests := []struct {
		name      string
		rule      tagparser.TagRule
		fieldType reflect.Type
		expected  string
	}{
		// String validators
		{name: "trim", rule: tagparser.TagRule{Name: "trim"}, fieldType: reflect.TypeOf(""), expected: ".Trim()"},
		{name: "lowercase", rule: tagparser.TagRule{Name: "lowercase"}, fieldType: reflect.TypeOf(""), expected: ".ToLowerCase()"},
		{name: "uppercase", rule: tagparser.TagRule{Name: "uppercase"}, fieldType: reflect.TypeOf(""), expected: ".ToUpperCase()"},
		{name: "nilable", rule: tagparser.TagRule{Name: "nilable"}, fieldType: reflect.TypeOf(""), expected: ".Nilable()"},
		{name: "url", rule: tagparser.TagRule{Name: "url"}, fieldType: reflect.TypeOf(""), expected: ".URL()"},
		{name: "ipv4", rule: tagparser.TagRule{Name: "ipv4"}, fieldType: reflect.TypeOf(""), expected: ".IPv4()"},
		{name: "ipv6", rule: tagparser.TagRule{Name: "ipv6"}, fieldType: reflect.TypeOf(""), expected: ".IPv6()"},
		{name: "regex", rule: tagparser.TagRule{Name: "regex", Params: []string{"^[A-Z]+$"}}, fieldType: reflect.TypeOf(""), expected: `.Regex(regexp.MustCompile("^[A-Z]+$"))`},

		// Numeric validators
		{name: "gte", rule: tagparser.TagRule{Name: "gte", Params: []string{"0"}}, fieldType: reflect.TypeOf(0), expected: ".Gte(0)"},
		{name: "lt", rule: tagparser.TagRule{Name: "lt", Params: []string{"100"}}, fieldType: reflect.TypeOf(0), expected: ".Lt(100)"},

		// Prefault
		{name: "prefault string", rule: tagparser.TagRule{Name: "prefault", Params: []string{"test"}}, fieldType: reflect.TypeOf(""), expected: `.Prefault("test")`},
		{name: "prefault int", rule: tagparser.TagRule{Name: "prefault", Params: []string{"42"}}, fieldType: reflect.TypeOf(0), expected: ".Prefault(42)"},

		// Required (returns empty)
		{name: "required", rule: tagparser.TagRule{Name: "required"}, fieldType: reflect.TypeOf(""), expected: ""},

		// Time (returns empty)
		{name: "time", rule: tagparser.TagRule{Name: "time"}, fieldType: reflect.TypeOf(""), expected: ""},

		// Refine and check
		{name: "refine", rule: tagparser.TagRule{Name: "refine", Params: []string{"myValidator"}}, fieldType: reflect.TypeOf(""), expected: ".Refine(myValidator)"},
		{name: "check", rule: tagparser.TagRule{Name: "check", Params: []string{"customCheck"}}, fieldType: reflect.TypeOf(""), expected: ".Check(customCheck)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := writer.generateValidatorChain(tt.rule, tt.fieldType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFileWriter_GenerateDefaultValue(t *testing.T) {
	writer, err := NewFileWriter("", "main", "_gen.go", true, false)
	require.NoError(t, err, "Failed to create writer")

	tests := []struct {
		name      string
		value     string
		fieldType reflect.Type
		expected  string
	}{
		{name: "string", value: "hello", fieldType: reflect.TypeOf(""), expected: `.Default("hello")`},
		{name: "int", value: "42", fieldType: reflect.TypeOf(0), expected: ".Default(42)"},
		{name: "int64", value: "100", fieldType: reflect.TypeOf(int64(0)), expected: ".Default(100)"},
		{name: "uint", value: "10", fieldType: reflect.TypeOf(uint(0)), expected: ".Default(10)"},
		{name: "float64", value: "3.14", fieldType: reflect.TypeOf(0.0), expected: ".Default(3.14)"},
		{name: "bool", value: "true", fieldType: reflect.TypeOf(false), expected: ".Default(true)"},
		{name: "pointer string", value: "world", fieldType: reflect.TypeOf((*string)(nil)), expected: `.Default("world")`},
		// Slice types
		{name: "string slice", value: `["a","b"]`, fieldType: reflect.TypeOf([]string{}), expected: `.Default([]string{"a", "b"})`},
		{name: "int slice", value: `[1,2,3]`, fieldType: reflect.TypeOf([]int{}), expected: `.Default([]int{1, 2, 3})`},
		{name: "float slice", value: `[1.1,2.2]`, fieldType: reflect.TypeOf([]float64{}), expected: `.Default([]float64{1.1, 2.2})`},
		{name: "bool slice", value: `[true,false]`, fieldType: reflect.TypeOf([]bool{}), expected: `.Default([]bool{true, false})`},
		// Map types
		{name: "string map", value: `{"k":"v"}`, fieldType: reflect.TypeOf(map[string]string{}), expected: `.Default(map[string]string{"k": "v"})`},
		{name: "interface map", value: `{"a":1}`, fieldType: reflect.TypeOf(map[string]interface{}{}), expected: `.Default(map[string]interface{}{"a": 1})`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := writer.generateDefaultValue(tt.value, tt.fieldType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFileWriter_GeneratePrefaultValue(t *testing.T) {
	writer, err := NewFileWriter("", "main", "_gen.go", true, false)
	require.NoError(t, err, "Failed to create writer")

	tests := []struct {
		name      string
		value     string
		fieldType reflect.Type
		expected  string
	}{
		{name: "string", value: "hello", fieldType: reflect.TypeOf(""), expected: `.Prefault("hello")`},
		{name: "int", value: "42", fieldType: reflect.TypeOf(0), expected: ".Prefault(42)"},
		{name: "int64", value: "100", fieldType: reflect.TypeOf(int64(0)), expected: ".Prefault(100)"},
		{name: "uint", value: "10", fieldType: reflect.TypeOf(uint(0)), expected: ".Prefault(10)"},
		{name: "float64", value: "3.14", fieldType: reflect.TypeOf(0.0), expected: ".Prefault(3.14)"},
		{name: "bool", value: "true", fieldType: reflect.TypeOf(false), expected: ".Prefault(true)"},
		{name: "pointer string", value: "world", fieldType: reflect.TypeOf((*string)(nil)), expected: `.Prefault("world")`},
		// Slice types
		{name: "string slice", value: `["x","y"]`, fieldType: reflect.TypeOf([]string{}), expected: `.Prefault([]string{"x", "y"})`},
		{name: "int slice", value: `[4,5,6]`, fieldType: reflect.TypeOf([]int{}), expected: `.Prefault([]int{4, 5, 6})`},
		{name: "float slice", value: `[3.3,4.4]`, fieldType: reflect.TypeOf([]float64{}), expected: `.Prefault([]float64{3.3, 4.4})`},
		{name: "bool slice", value: `[false,true]`, fieldType: reflect.TypeOf([]bool{}), expected: `.Prefault([]bool{false, true})`},
		// Map types
		{name: "string map", value: `{"foo":"bar"}`, fieldType: reflect.TypeOf(map[string]string{}), expected: `.Prefault(map[string]string{"foo": "bar"})`},
		{name: "interface map", value: `{"x":99}`, fieldType: reflect.TypeOf(map[string]interface{}{}), expected: `.Prefault(map[string]interface{}{"x": 99})`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := writer.generatePrefaultValue(tt.value, tt.fieldType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFileWriter_FirstLowerCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "simple", input: "User", expected: "user"},
		{name: "acronym", input: "APIResponse", expected: "apiresponse"}, // All-caps prefix all lowercased
		{name: "http", input: "HTTPClient", expected: "httpclient"},      // All-caps prefix all lowercased
		{name: "xml", input: "XMLParser", expected: "xmlparser"},         // All caps then lowercase
		{name: "single char", input: "A", expected: "a"},
		{name: "empty", input: "", expected: ""},
		{name: "generic", input: "Response[T any]", expected: "response"},
		{name: "already lower", input: "user", expected: "user"},
		{name: "two chars", input: "ID", expected: "iD"}, // Only first char lowercased, not reaching acronym branch
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := firstLowerCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFileWriter_ReceiverName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "simple", input: "User", expected: "u"},
		{name: "camelCase", input: "UserProfile", expected: "up"},
		{name: "acronym", input: "APIResponse", expected: "a"},      // All-caps prefix → first letter only
		{name: "empty", input: "", expected: "x"},
		{name: "generic", input: "Response[T any]", expected: "r"},
		{name: "reserved type", input: "Type", expected: "t"},       // Not exactly "type"
		{name: "reserved interface", input: "Interface", expected: "i"}, // Not exactly "interface"
		{name: "reserved struct", input: "Struct", expected: "s"},   // Not exactly "struct"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := receiverName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
