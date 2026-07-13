package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/pkg/tagparser"
)

func TestFileWriterRejectsInvalidFieldPlan(t *testing.T) {
	t.Parallel()

	writer, err := NewFileWriter("", "main", "_gen.go", true, false)
	require.NoError(t, err)
	field := &tagparser.FieldInfo{
		Name:     "Age",
		Type:     reflect.TypeFor[int](),
		GoZodTag: "mystery=1",
		Rules: []tagparser.TagRule{
			{Name: "mystery", Params: []string{"1"}},
		},
	}

	code, err := writer.generateFieldSchemaCode(field, "User")

	assert.Empty(t, code)
	assert.True(t, errors.Is(err, tagparser.ErrUnknownRule))
	assert.ErrorContains(t, err, "Age")
	assert.ErrorContains(t, err, "mystery=1")
}

func TestFileWriterRejectsUnknownFieldType(t *testing.T) {
	t.Parallel()

	writer, err := NewFileWriter("", "main", "_gen.go", true, false)
	require.NoError(t, err)
	field := &tagparser.FieldInfo{Name: "Mystery", TypeName: "unknown"}

	code, err := writer.generateFieldSchemaCode(field, "User")

	assert.Empty(t, code)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "Mystery")
	assert.ErrorContains(t, err, "unknown")
}

func TestFileWriterUsesCoercionConstructors(t *testing.T) {
	t.Parallel()

	writer, err := NewFileWriter("", "main", "_gen.go", true, false)
	require.NoError(t, err)

	tests := []struct {
		name      string
		fieldType reflect.Type
		want      string
	}{
		{name: "string", fieldType: reflect.TypeFor[string](), want: "coerce.String()"},
		{name: "bool", fieldType: reflect.TypeFor[bool](), want: "coerce.Bool()"},
		{name: "signed integer", fieldType: reflect.TypeFor[int](), want: "coerce.Int()"},
		{name: "unsigned integer", fieldType: reflect.TypeFor[uint32](), want: "coerce.Uint32()"},
		{name: "float", fieldType: reflect.TypeFor[float64](), want: "coerce.Float64()"},
		{name: "time", fieldType: reflect.TypeFor[time.Time](), want: "coerce.Time()"},
		{name: "string pointer", fieldType: reflect.TypeFor[*string](), want: "coerce.StringPtr().Optional()"},
		{name: "integer pointer", fieldType: reflect.TypeFor[*int16](), want: "coerce.Int16Ptr().Optional()"},
		{name: "time pointer", fieldType: reflect.TypeFor[*time.Time](), want: "coerce.TimePtr().Optional()"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			field := &tagparser.FieldInfo{
				Name:     "Value",
				Type:     test.fieldType,
				Required: true,
				Rules: []tagparser.TagRule{
					{Name: "required"},
					{Name: "coerce"},
				},
			}

			code, err := writer.generateFieldSchemaCode(field, "User")

			require.NoError(t, err)
			assert.Equal(t, test.want, code)
		})
	}
}

func TestWriteGeneratedCodePreservesExistingFileWhenAtomicReplaceCannotStart(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "user.go")
	outputPath := filepath.Join(dir, "user_gen.go")
	require.NoError(t, os.WriteFile(sourcePath, []byte("package main\n"), 0o600))
	require.NoError(t, os.WriteFile(outputPath, []byte("old generated content\n"), 0o600))
	require.NoError(t, os.Chmod(dir, 0o500))
	t.Cleanup(func() { require.NoError(t, os.Chmod(dir, 0o700)) })

	writer, err := NewFileWriter("", "main", "_gen.go", false, false)
	require.NoError(t, err)
	err = writer.WriteGeneratedCode(&GenerationInfo{
		Name:     "User",
		Package:  "main",
		FilePath: sourcePath,
	})

	assert.Error(t, err)
	content, readErr := os.ReadFile(outputPath)
	require.NoError(t, readErr)
	assert.Equal(t, "old generated content\n", string(content))
}

func TestRenderValidatorChainRejectsInvalidOperand(t *testing.T) {
	t.Parallel()

	code, err := renderValidatorChain(tagparser.RulePlan{
		Name:    "regex",
		Op:      tagparser.RuleRegex,
		Family:  tagparser.FieldFamilyString,
		Operand: "not-compiled",
	}, reflect.TypeFor[string]())

	assert.Empty(t, code)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "regex")
}

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
					Type: reflect.TypeFor[string](),
					Rules: []tagparser.TagRule{
						{Name: "required"},
						{Name: "uuid"},
					},
				},
				{
					Name: "Email",
					Type: reflect.TypeFor[string](),
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
					Type: reflect.TypeFor[string](),
					Rules: []tagparser.TagRule{
						{Name: "regex", Params: []string{"^[A-Z0-9]+$"}},
					},
				},
			},
			expectedImports:   []string{"github.com/kaptinlin/gozod", "regexp"},
			unexpectedImports: []string{"github.com/kaptinlin/gozod/core"},
		},
		{
			name: "string formats use gozod constructors only",
			fields: []tagparser.FieldInfo{
				{
					Name: "Website",
					Type: reflect.TypeFor[string](),
					Rules: []tagparser.TagRule{
						{Name: "url"},
					},
				},
				{
					Name: "Address",
					Type: reflect.TypeFor[string](),
					Rules: []tagparser.TagRule{
						{Name: "ipv4"},
					},
				},
			},
			expectedImports:   []string{"github.com/kaptinlin/gozod"},
			unexpectedImports: []string{"net/url", "net"},
		},
		{
			name: "time fields require time import",
			fields: []tagparser.FieldInfo{
				{
					Name: "CreatedAt",
					Type: reflect.TypeFor[struct{}](), // Mock time.Time
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
					tt.fields[i].Type = reflect.TypeFor[struct{}]() // We'll mock this
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
				found := slices.Contains(imports, expectedImport)
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
				Type: reflect.TypeFor[string](),
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
				Type: reflect.TypeFor[string](),
				Rules: []tagparser.TagRule{
					{Name: "required"},
					{Name: "email"},
				},
			},
			structName:     "User",
			expectedSchema: "gozod.Email()",
			expectError:    false,
		},
		{
			name: "cidrv4 format field",
			field: tagparser.FieldInfo{
				Name: "Network",
				Type: reflect.TypeFor[string](),
				Rules: []tagparser.TagRule{
					{Name: "required"},
					{Name: "cidrv4"},
					{Name: "min", Params: []string{"9"}},
				},
			},
			structName:     "NetworkConfig",
			expectedSchema: "gozod.CIDRv4().Min(9)",
			expectError:    false,
		},
		{
			name: "integer field with range",
			field: tagparser.FieldInfo{
				Name: "Age",
				Type: reflect.TypeFor[int](),
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
				Type: reflect.TypeFor[float64](),
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
				Type: reflect.TypeFor[string](),
				Rules: []tagparser.TagRule{
					{Name: "enum", Params: []string{"active", "inactive"}},
					{Name: "default", Params: []string{"active"}},
				},
			},
			structName:     "User",
			expectedSchema: `gozod.Enum("active", "inactive").Optional().Default("active")`,
			expectError:    false,
		},
		{
			name: "optional field with default keeps default outermost",
			field: tagparser.FieldInfo{
				Name: "Nickname",
				Type: reflect.TypeFor[string](),
				Rules: []tagparser.TagRule{
					{Name: "default", Params: []string{"guest"}},
				},
			},
			structName:     "User",
			expectedSchema: `gozod.String().Optional().Default("guest")`,
			expectError:    false,
		},
		{
			name: "optional pointer field",
			field: tagparser.FieldInfo{
				Name:     "Description",
				Type:     reflect.TypeFor[*string](),
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

			schema, err := writer.generateFieldSchemaCode(&tt.field, tt.structName)
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

func TestFileWriter_StringFormatRulesUseRootConstructors(t *testing.T) {
	tests := []struct {
		rule        string
		constructor string
	}{
		{rule: "email", constructor: "Email"},
		{rule: "url", constructor: "URL"},
		{rule: "uuid", constructor: "UUID"},
		{rule: "ipv4", constructor: "IPv4"},
		{rule: "ipv6", constructor: "IPv6"},
		{rule: "cidrv4", constructor: "CIDRv4"},
		{rule: "cidrv6", constructor: "CIDRv6"},
		{rule: "cuid", constructor: "CUID"},
		{rule: "cuid2", constructor: "CUID2"},
		{rule: "jwt", constructor: "JWT"},
		{rule: "iso_datetime", constructor: "IsoDateTime"},
		{rule: "iso_date", constructor: "IsoDate"},
		{rule: "iso_time", constructor: "IsoTime"},
		{rule: "iso_duration", constructor: "IsoDuration"},
	}

	for _, tt := range tests {
		t.Run(tt.rule, func(t *testing.T) {
			writer, err := NewFileWriter("", "main", "_gen.go", true, false)
			require.NoError(t, err)

			field := tagparser.FieldInfo{
				Name: "Token",
				Type: reflect.TypeFor[string](),
				Rules: []tagparser.TagRule{
					{Name: "required"},
					{Name: tt.rule},
					{Name: "min", Params: []string{"3"}},
				},
			}

			got, err := writer.generateFieldSchemaCode(&field, "Credentials")
			require.NoError(t, err)
			assert.Contains(t, got, "gozod."+tt.constructor+"().Min(3)")
			assert.NotContains(t, got, "gozod.String()."+tt.constructor+"()")
		})
	}
}

func TestFileWriter_WriteGeneratedCode(t *testing.T) {
	helper := NewTestHelper(t)
	sourcePath := helper.CreateGoFile("user.go", `package main
type User struct {
	Name string `+"`gozod:\"required\"`"+`
}
`)

	writer, err := NewFileWriter(helper.GetTempDir(), "main", "_gen.go", false, true)
	require.NoError(t, err)

	err = writer.WriteGeneratedCode(&GenerationInfo{
		Name:     "User",
		Package:  "main",
		FilePath: sourcePath,
		Fields: []tagparser.FieldInfo{{
			Name:     "Name",
			FieldKey: "name",
			Type:     reflect.TypeFor[string](),
			Rules:    []tagparser.TagRule{{Name: "required"}},
		}},
	})
	require.NoError(t, err)
	helper.AssertFileExists("user_gen.go")

	content := helper.ReadGeneratedFile("user_gen.go")
	assert.NotContains(t, content, "Generated at:")
}

func TestFileWriter_GenerateCodeUsesInfoPackage(t *testing.T) {
	writer, err := NewFileWriter("", "", "_gen.go", true, false)
	require.NoError(t, err)

	content, err := writer.generateCode(&GenerationInfo{
		Name:    "User",
		Package: "models",
		Fields: []tagparser.FieldInfo{{
			Name:     "Name",
			FieldKey: "name",
			Type:     reflect.TypeFor[string](),
			Rules:    []tagparser.TagRule{{Name: "required"}},
		}},
	})
	require.NoError(t, err)
	assert.Contains(t, content, "package models")
}

func TestFileWriter_GenerateCodeReturnsTemplateError(t *testing.T) {
	writer := &FileWriter{templates: template.New("broken"), packageName: "main"}

	_, err := writer.generateCode(&GenerationInfo{
		Name:    "User",
		Package: "main",
		Fields: []tagparser.FieldInfo{{
			Name:     "Name",
			FieldKey: "name",
			Type:     reflect.TypeFor[string](),
			Rules:    []tagparser.TagRule{{Name: "required"}},
		}},
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "execute template")
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
						FieldKey: "id",
						Type:     reflect.TypeFor[string](),
						Rules: []tagparser.TagRule{
							{Name: "required"},
							{Name: "uuid"},
						},
					},
					{
						Name:     "Name",
						FieldKey: "name",
						Type:     reflect.TypeFor[string](),
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
				`"id":   gozod.UUID()`,
				`"name": gozod.String().Min(2)`,
				"gozod.Struct[User](gozod.StructSchema{",
			},
			unexpectedContent: []string{
				"github.com/kaptinlin/gozod/core",
				"Generated at:",
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
				Type:  reflect.TypeFor[string](),
				Rules: tt.rules,
			}

			// For numeric rules test, use int type
			if strings.Contains(tt.name, "numeric") {
				field.Type = reflect.TypeFor[int]()
			}

			result, err := writer.generateFieldSchemaCode(&field, "TestStruct")
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

func TestBasicTypeConstructor(t *testing.T) {
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
		{name: "non-basic type", typeName: "CustomType", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := basicTypeConstructor(tt.typeName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCircularReferenceHandling(t *testing.T) {
	tests := []struct {
		name       string
		typeName   string
		structName string
		expected   string
		wantErr    bool
	}{
		{
			name:       "self reference",
			typeName:   "Node",
			structName: "Node",
			expected:   "gozod.LazyTyped[*Node](func() any { return gozod.MustFromStruct[Node]() })",
		},
		{
			name:       "pointer self reference",
			typeName:   "*Node",
			structName: "Node",
			expected:   "gozod.LazyTyped[*Node](func() any { return gozod.MustFromStruct[Node]() })",
		},
		{
			name:       "slice self reference",
			typeName:   "[]*Node",
			structName: "Node",
			expected:   "gozod.Slice[*Node](gozod.LazyTyped[*Node](func() any { return gozod.MustFromStruct[Node]() }))",
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
		{
			name:       "explicit unknown uses Any",
			typeName:   "unknown",
			structName: "Node",
			expected:   "gozod.Any()",
		},
		{
			name:       "malformed map fails",
			typeName:   "map[string]",
			structName: "Node",
			wantErr:    true,
		},
		{
			name:       "map self reference",
			typeName:   "map[string]*Node",
			structName: "Node",
			expected:   "gozod.Record[string, *Node](gozod.String(), gozod.LazyTyped[*Node](func() any { return gozod.MustFromStruct[Node]() }))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := baseConstructor(tt.typeName, tt.structName, defaultFieldNameTag)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if !strings.Contains(result, tt.expected) {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestBaseConstructor_FieldNameTag(t *testing.T) {
	result, err := baseConstructor("Profile", "User", "yaml")
	require.NoError(t, err)
	assert.Equal(t, `gozod.MustFromStruct[Profile](gozod.WithFieldNameTag("yaml"))`, result)

	self, err := baseConstructor("*Node", "Node", "yaml")
	require.NoError(t, err)
	assert.Contains(t, self, `gozod.MustFromStruct[Node](gozod.WithFieldNameTag("yaml"))`)
}

func TestGenerateValidatorChain(t *testing.T) {
	tests := []struct {
		name      string
		rule      tagparser.TagRule
		fieldType reflect.Type
		expected  string
		wantErr   bool
	}{
		// String validators
		{name: "trim", rule: tagparser.TagRule{Name: "trim"}, fieldType: reflect.TypeFor[string](), expected: ".Trim()"},
		{name: "lowercase", rule: tagparser.TagRule{Name: "lowercase"}, fieldType: reflect.TypeFor[string](), expected: ".ToLowerCase()"},
		{name: "uppercase", rule: tagparser.TagRule{Name: "uppercase"}, fieldType: reflect.TypeFor[string](), expected: ".ToUpperCase()"},
		{name: "nilable", rule: tagparser.TagRule{Name: "nilable"}, fieldType: reflect.TypeFor[string](), expected: ".Nilable()"},
		{name: "email chain", rule: tagparser.TagRule{Name: "email"}, fieldType: reflect.TypeFor[string](), expected: ".Email()"},
		{name: "url chain", rule: tagparser.TagRule{Name: "url"}, fieldType: reflect.TypeFor[string](), expected: ".URL()"},
		{name: "ipv4 chain", rule: tagparser.TagRule{Name: "ipv4"}, fieldType: reflect.TypeFor[string](), expected: ".IPv4()"},
		{name: "ipv6 chain", rule: tagparser.TagRule{Name: "ipv6"}, fieldType: reflect.TypeFor[string](), expected: ".IPv6()"},
		{name: "cidrv4 chain", rule: tagparser.TagRule{Name: "cidrv4"}, fieldType: reflect.TypeFor[string](), expected: ".CIDRv4()"},
		{name: "regex", rule: tagparser.TagRule{Name: "regex", Params: []string{"^[A-Z]+$"}}, fieldType: reflect.TypeFor[string](), expected: `.Regex(regexp.MustCompile("^[A-Z]+$"))`},
		{name: "regex with quotes", rule: tagparser.TagRule{Name: "regex", Params: []string{"^\"[A-Z]+\"$"}}, fieldType: reflect.TypeFor[string](), expected: `.Regex(regexp.MustCompile("^\"[A-Z]+\"$"))`},
		{name: "includes", rule: tagparser.TagRule{Name: "includes", Params: []string{"PROD"}}, fieldType: reflect.TypeFor[string](), expected: `.Includes("PROD")`},
		{name: "starts with", rule: tagparser.TagRule{Name: "startswith", Params: []string{"ID-"}}, fieldType: reflect.TypeFor[string](), expected: `.StartsWith("ID-")`},

		// Numeric validators
		{name: "gte", rule: tagparser.TagRule{Name: "gte", Params: []string{"0"}}, fieldType: reflect.TypeFor[int](), expected: ".Gte(0)"},
		{name: "lt", rule: tagparser.TagRule{Name: "lt", Params: []string{"100"}}, fieldType: reflect.TypeFor[int](), expected: ".Lt(100)"},
		{name: "multiple of", rule: tagparser.TagRule{Name: "multipleof", Params: []string{"5"}}, fieldType: reflect.TypeFor[int](), expected: ".MultipleOf(5)"},

		// Prefault
		{name: "prefault string", rule: tagparser.TagRule{Name: "prefault", Params: []string{"test"}}, fieldType: reflect.TypeFor[string](), expected: `.Prefault("test")`},
		{name: "prefault int", rule: tagparser.TagRule{Name: "prefault", Params: []string{"42"}}, fieldType: reflect.TypeFor[int](), expected: ".Prefault(42)"},

		// Required (returns empty)
		{name: "required", rule: tagparser.TagRule{Name: "required"}, fieldType: reflect.TypeFor[string](), expected: ""},

		// Time (returns empty)
		{name: "time on string fails", rule: tagparser.TagRule{Name: "time"}, fieldType: reflect.TypeFor[string](), wantErr: true},
		{name: "enum method returns empty", rule: tagparser.TagRule{Name: "enum", Params: []string{"active"}}, fieldType: reflect.TypeFor[string](), expected: ""},
		{name: "unknown rule fails", rule: tagparser.TagRule{Name: "unknown"}, fieldType: reflect.TypeFor[string](), wantErr: true},

		// Function names cannot be bound safely from struct-tag strings.
		{name: "refine fails", rule: tagparser.TagRule{Name: "refine", Params: []string{"myValidator"}}, fieldType: reflect.TypeFor[string](), wantErr: true},
		{name: "check fails", rule: tagparser.TagRule{Name: "check", Params: []string{"customCheck"}}, fieldType: reflect.TypeFor[string](), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, err := tagparser.CompileFieldPlan(&tagparser.FieldInfo{
				Name:  "Field",
				Type:  tt.fieldType,
				Rules: []tagparser.TagRule{tt.rule},
			})
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if len(plan.Operations) == 0 {
				assert.Empty(t, tt.expected)
				return
			}
			result, err := renderValidatorChain(plan.Operations[0], tt.fieldType)
			assert.NoError(t, err)
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
		{name: "acronym", input: "APIResponse", expected: "a"}, // All-caps prefix → first letter only
		{name: "empty", input: "", expected: "x"},
		{name: "generic", input: "Response[T any]", expected: "r"},
		{name: "reserved type", input: "Type", expected: "t"},           // Not exactly "type"
		{name: "reserved interface", input: "Interface", expected: "i"}, // Not exactly "interface"
		{name: "reserved struct", input: "Struct", expected: "s"},       // Not exactly "struct"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := receiverName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
