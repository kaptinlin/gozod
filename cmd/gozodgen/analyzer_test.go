package main

import (
	"go/ast"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/pkg/tagparser"
)

func TestStructAnalyzer_AnalyzePackageSkipsTestPackages(t *testing.T) {
	helper := NewTestHelper(t)
	helper.CreateGoFile("user_test.go", `package main_test
type User struct {
	Name string `+"`gozod:\"required\"`"+`
}
`)

	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err)

	structs, err := analyzer.AnalyzePackage(helper.GetTempDir())
	require.NoError(t, err)
	assert.Len(t, structs, 0)
}

func TestStructAnalyzer_AnalyzePackage(t *testing.T) {
	tests := []struct {
		name          string
		sourceFiles   map[string]string
		expectedCount int
		expectError   bool
	}{
		{
			name: "single file with simple struct",
			sourceFiles: map[string]string{
				"user.go": `package main
type User struct {
	Name string ` + "`gozod:\"required,min=2\"`" + `
	Age  int    ` + "`gozod:\"required,min=18\"`" + `
}`,
			},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "multiple structs in one file",
			sourceFiles: map[string]string{
				"models.go": `package main
type User struct {
	Name string ` + "`gozod:\"required\"`" + `
}
type Product struct {
	Name  string  ` + "`gozod:\"required\"`" + `
	Price float64 ` + "`gozod:\"required,gt=0\"`" + `
}`,
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "struct without gozod tags",
			sourceFiles: map[string]string{
				"plain.go": `package main
type User struct {
	Name string
	Age  int
}`,
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "empty package",
			sourceFiles: map[string]string{
				"empty.go": `package main
`,
			},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helper := NewTestHelper(t)

			// Create source files
			for filename, content := range tt.sourceFiles {
				helper.CreateGoFile(filename, content)
			}

			analyzer, err := NewStructAnalyzer()
			require.NoError(t, err, "Failed to create analyzer")

			structs, err := analyzer.AnalyzePackage(helper.GetTempDir())
			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Unexpected error")
			}

			if !assert.Equal(t, tt.expectedCount, len(structs), "Expected %d structs, got %d", tt.expectedCount, len(structs)) {
				for i, s := range structs {
					t.Logf("  Struct %d: %s with %d fields", i, s.Name, len(s.Fields))
				}
			}
		})
	}
}

func TestStructAnalyzer_AnalyzePackageReturnsTagParseError(t *testing.T) {
	helper := NewTestHelper(t)
	helper.CreateGoFile("user.go", `package main
type User struct {
	Name string `+"`gozod:\"min=\"`"+`
}
`)

	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err)

	_, err = analyzer.AnalyzePackage(helper.GetTempDir())
	require.Error(t, err)
	assert.ErrorContains(t, err, "parse gozod tag for field Name")
}

func TestStructAnalyzer_AnalyzePackageIgnoresUnsupportedFieldsOutsideGenerationScope(t *testing.T) {
	helper := NewTestHelper(t)
	helper.CreateGoFile("models.go", `package main

type Runner interface { Run() }

type Plain struct {
	Updates  chan int
	Callback func(int) error
	Runner   Runner
}

type Generated struct {
	Name string `+"`gozod:\"required\"`"+`
}
`)

	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err)

	structs, err := analyzer.AnalyzePackage(helper.GetTempDir())
	require.NoError(t, err)
	require.Len(t, structs, 1)
	assert.Equal(t, "Generated", structs[0].Name)
	require.Len(t, structs[0].Fields, 1)
	assert.Equal(t, "Name", structs[0].Fields[0].Name)
}

func TestStructAnalyzer_AnalyzePackageLoadsModuleLocalImports(t *testing.T) {
	helper := NewTestHelper(t)
	helper.CreateGoFile("go.mod", `module example.test/local

go 1.26.5
`)
	helper.CreateGoFile("domain/code.go", `package domain

type Code string
`)
	helper.CreateGoFile("model/user.go", `package model

import "example.test/local/domain"

type User struct {
	Code domain.Code `+"`gozod:\"min=2\"`"+`
}
`)

	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err)

	structs, err := analyzer.AnalyzePackage(filepath.Join(helper.GetTempDir(), "model"))
	require.NoError(t, err)
	require.Len(t, structs, 1)
	require.Len(t, structs[0].Fields, 1)
	assert.Equal(t, reflect.String, structs[0].Fields[0].Type.Kind())
	assert.Equal(t, "domain.Code", structs[0].Fields[0].TypeName)
}

func TestStructAnalyzer_AnalyzePackageUsesDefaultBuildConstraints(t *testing.T) {
	helper := NewTestHelper(t)
	helper.CreateGoFile("active.go", `package fixture

type User struct {
	Name string `+"`gozod:\"required\"`"+`
}
`)
	helper.CreateGoFile("ignored.go", `//go:build never

package fixture

this is not valid Go syntax
`)

	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err)

	structs, err := analyzer.AnalyzePackage(helper.GetTempDir())
	require.NoError(t, err)
	require.Len(t, structs, 1)
	assert.Equal(t, "User", structs[0].Name)
}

func TestStructAnalyzer_AnalyzePackagePreservesFixedArrayType(t *testing.T) {
	helper := NewTestHelper(t)
	helper.CreateGoFile("shape.go", `package fixture

type Shape struct {
	Fixed   [3]string `+"`gozod:\"required\"`"+`
	Dynamic []string  `+"`gozod:\"required\"`"+`
}
`)

	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err)

	structs, err := analyzer.AnalyzePackage(helper.GetTempDir())
	require.NoError(t, err)
	require.Len(t, structs, 1)
	require.Len(t, structs[0].Fields, 2)

	fixed := structs[0].Fields[0]
	require.Equal(t, reflect.Array, fixed.Type.Kind())
	assert.Equal(t, 3, fixed.Type.Len())
	assert.Equal(t, "[3]string", fixed.TypeName)

	dynamic := structs[0].Fields[1]
	assert.Equal(t, reflect.Slice, dynamic.Type.Kind())
	assert.Equal(t, "[]string", dynamic.TypeName)
}

func TestStructAnalyzer_AnalyzePackageRejectsTypeErrors(t *testing.T) {
	helper := NewTestHelper(t)
	helper.CreateGoFile("go.mod", `module example.test/typeerror

go 1.26.5
`)
	helper.CreateGoFile("model/user.go", `package model

var invalid string = 42

type User struct {
	Name string `+"`gozod:\"required\"`"+`
}
`)

	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err)

	_, err = analyzer.AnalyzePackage(filepath.Join(helper.GetTempDir(), "model"))
	require.Error(t, err)
	assert.ErrorContains(t, err, "model/user.go")
	assert.ErrorContains(t, err, "cannot use 42")
}

func TestStructAnalyzer_AnalyzePackageRejectsInvalidPackages(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		contains []string
	}{
		{
			name: "unresolved import",
			files: map[string]string{
				"user.go": `package fixture

import "example.test/missing"

type User struct {
	Value missing.Value ` + "`gozod:\"required\"`" + `
}
`,
			},
			contains: []string{"user.go", "example.test/missing"},
		},
		{
			name: "undefined type",
			files: map[string]string{
				"user.go": `package fixture

type User struct {
	Value Missing ` + "`gozod:\"required\"`" + `
}
`,
			},
			contains: []string{"user.go", "undefined: Missing"},
		},
		{
			name: "conflicting package",
			files: map[string]string{
				"alpha.go": "package alpha\n",
				"beta.go":  "package beta\n",
			},
			contains: []string{"alpha.go", "beta.go", "found packages"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helper := NewTestHelper(t)
			helper.CreateGoFile("go.mod", `module example.test/invalid

go 1.26.5
`)
			for name, content := range tt.files {
				helper.CreateGoFile(name, content)
			}

			analyzer, err := NewStructAnalyzer()
			require.NoError(t, err)

			_, err = analyzer.AnalyzePackage(helper.GetTempDir())
			require.Error(t, err)
			for _, text := range tt.contains {
				assert.ErrorContains(t, err, text)
			}
		})
	}
}

func TestStructAnalyzer_ParseTagRules(t *testing.T) {
	tests := []struct {
		name        string
		tagValue    string
		expectError bool
		expected    []tagparser.TagRule
	}{
		{
			name:        "simple required tag",
			tagValue:    "required",
			expectError: false,
			expected: []tagparser.TagRule{
				{Name: "required", Params: nil},
			},
		},
		{
			name:        "multiple rules",
			tagValue:    "required,min=2,max=50",
			expectError: false,
			expected: []tagparser.TagRule{
				{Name: "required", Params: nil},
				{Name: "min", Params: []string{"2"}},
				{Name: "max", Params: []string{"50"}},
			},
		},
		{
			name:        "enum with multiple values",
			tagValue:    "required,enum=active inactive pending",
			expectError: false,
			expected: []tagparser.TagRule{
				{Name: "required", Params: nil},
				{Name: "enum", Params: []string{"active", "inactive", "pending"}},
			},
		},
		{
			name:        "regex pattern",
			tagValue:    "regex=^[A-Za-z0-9]+$",
			expectError: false,
			expected: []tagparser.TagRule{
				{Name: "regex", Params: []string{"^[A-Za-z0-9]+$"}},
			},
		},
		{
			name:        "empty tag",
			tagValue:    "",
			expectError: false,
			expected:    nil,
		},
		{
			name:        "blank rule is ignored",
			tagValue:    "required, ,max=50",
			expectError: false,
			expected: []tagparser.TagRule{
				{Name: "required", Params: nil},
				{Name: "max", Params: []string{"50"}},
			},
		},
		{
			name:        "empty parameter errors",
			tagValue:    "min=",
			expectError: true,
		},
		{
			name:        "space separated non-regex parameters",
			tagValue:    "between=1 10",
			expectError: false,
			expected: []tagparser.TagRule{
				{Name: "between", Params: []string{"1", "10"}},
			},
		},
	}

	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err, "Failed to create analyzer")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, err := analyzer.parseTagRules(tt.tagValue)
			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Unexpected error")
			}

			if !assert.Equal(t, len(tt.expected), len(rules), "Expected %d rules, got %d", len(tt.expected), len(rules)) {
				return
			}

			for i, rule := range rules {
				expected := tt.expected[i]
				assert.Equal(t, expected.Name, rule.Name, "Rule %d: expected name %s, got %s", i, expected.Name, rule.Name)
				assert.Equal(t, expected.Params, rule.Params, "Rule %d: expected params %v, got %v", i, expected.Params, rule.Params)
			}
		})
	}
}

func TestStructAnalyzer_CircularReferences(t *testing.T) {
	sourceCode := `package main
type Node struct {
	Value    int     ` + "`gozod:\"required\"`" + `
	Next     *Node   ` + "`gozod:\"\"`" + `
	Children []*Node ` + "`gozod:\"\"`" + `
}`

	helper := NewTestHelper(t)
	helper.CreateGoFile("test.go", sourceCode)

	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err, "Failed to create analyzer")

	structs, err := analyzer.AnalyzePackage(helper.GetTempDir())
	require.NoError(t, err, "Failed to analyze package")

	require.Equal(t, 1, len(structs), "Expected 1 struct, got %d", len(structs))

	node := structs[0]
	assert.Equal(t, "Node", node.Name, "Expected struct name 'Node', got %s", node.Name)

	// Check that circular fields are detected
	hasNext := false
	hasChildren := false
	for _, field := range node.Fields {
		if field.Name == "Next" {
			hasNext = true
			t.Logf("Found Next field with type: %s", field.Type.String())
		}
		if field.Name == "Children" {
			hasChildren = true
			t.Logf("Found Children field with type: %s", field.Type.String())
		}
	}

	assert.True(t, hasNext, "Expected to find Next field")
	assert.True(t, hasChildren, "Expected to find Children field")
}

func TestStructAnalyzer_RealFiles(t *testing.T) {
	// Test with actual testdata files
	testFiles := []string{
		"simple_struct.go",
		"complex_struct.go",
		"circular_struct.go",
		"edge_cases.go",
		"validators.go",
	}

	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err, "Failed to create analyzer")

	for _, fileName := range testFiles {
		t.Run(fileName, func(t *testing.T) {
			filePath := filepath.Join("testdata", fileName)

			// Create a temp package with the test file
			helper := NewTestHelper(t)
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Skipf("Testdata file %s not found: %v", filePath, err)
				return
			}

			// Change package name to main for testing
			contentStr := string(content)
			contentStr = strings.Replace(contentStr, "package testdata", "package main", 1)
			helper.CreateGoFile(fileName, contentStr)

			structs, err := analyzer.AnalyzePackage(helper.GetTempDir())
			assert.NoError(t, err, "Failed to analyze %s", fileName)

			// Basic sanity check - we should find some structs in most files
			switch fileName {
			case "simple_struct.go":
				assert.NotEmpty(t, structs, "Expected to find structs in simple_struct.go")
			case "complex_struct.go":
				assert.GreaterOrEqual(t, len(structs), 2, "Expected at least 2 structs in complex_struct.go, got %d", len(structs))
			case "circular_struct.go":
				assert.GreaterOrEqual(t, len(structs), 3, "Expected at least 3 structs in circular_struct.go, got %d", len(structs))
			}

			// Print struct info for debugging
			t.Logf("Found %d structs in %s:", len(structs), fileName)
			for _, s := range structs {
				t.Logf("  - %s with %d fields", s.Name, len(s.Fields))
			}
		})
	}
}

func TestNeedsGeneration(t *testing.T) {
	tests := []struct {
		name     string
		info     *GenerationInfo
		expected bool
	}{
		{
			name: "struct with gozod tags",
			info: &GenerationInfo{
				Name: "User",
				Fields: []tagparser.FieldInfo{
					{
						Name: "Name",
						Rules: []tagparser.TagRule{
							{Name: "required"},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "struct with generate directive",
			info: &GenerationInfo{
				Name:        "User",
				HasGenerate: true,
			},
			expected: true,
		},
		{
			name: "struct without gozod tags",
			info: &GenerationInfo{
				Name: "User",
				Fields: []tagparser.FieldInfo{
					{
						Name:  "Name",
						Rules: nil,
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NeedsGeneration(tt.info)
			assert.Equal(t, tt.expected, result, "Expected %t, got %t", tt.expected, result)
		})
	}
}

func TestStructAnalyzer_AnalyzePackageWithGenerateDirective(t *testing.T) {
	helper := NewTestHelper(t)
	helper.CreateGoFile("generated.go", `package main

//go:generate gozodgen
type User struct {
	Name string
}
`)

	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err)

	structs, err := analyzer.AnalyzePackage(helper.GetTempDir())
	require.NoError(t, err)
	require.Len(t, structs, 1)

	assert.True(t, structs[0].HasGenerate)
	require.Len(t, structs[0].Fields, 1)
	assert.Equal(t, "Name", structs[0].Fields[0].Name)
	assert.Equal(t, "Name", structs[0].Fields[0].FieldKey)
}

func TestStructAnalyzer_ParseStructFieldsWithFieldNameFallbacks(t *testing.T) {
	helper := NewTestHelper(t)
	helper.CreateGoFile("test.go", `package main

//go:generate gozodgen
type User struct {
	Name   string
	Alias  string `+"`json:\"\"`"+`
	Hidden string `+"`json:\"-\" gozod:\"required\"`"+`
	hidden string
}
`)

	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err)

	structs, err := analyzer.AnalyzePackage(helper.GetTempDir())
	require.NoError(t, err)
	require.Len(t, structs, 1)
	fields := structs[0].Fields
	require.Len(t, fields, 2)

	assert.Equal(t, "Name", fields[0].Name)
	assert.Equal(t, reflect.TypeFor[string](), fields[0].Type)
	assert.Equal(t, "string", fields[0].TypeName)
	assert.Equal(t, "Name", fields[0].FieldKey)

	assert.Equal(t, "Alias", fields[1].Name)
	assert.Equal(t, reflect.TypeFor[string](), fields[1].Type)
	assert.Equal(t, "string", fields[1].TypeName)
	assert.Equal(t, "Alias", fields[1].FieldKey)

	for _, field := range fields {
		assert.NotEqual(t, "Hidden", field.Name)
	}
}

func TestStructAnalyzerRejectsUnsupportedFieldTypes(t *testing.T) {
	tests := []struct {
		name         string
		declarations string
		fieldType    string
		wantType     string
	}{
		{name: "channel", fieldType: "chan int", wantType: "chan int"},
		{name: "function", fieldType: "func(int) error", wantType: "func(int) error"},
		{
			name:         "non-empty interface",
			declarations: "type Runner interface { Run() }\n",
			fieldType:    "Runner",
			wantType:     "Runner",
		},
		{
			name:         "type parameter",
			declarations: "type Box[T any] struct {\n\tValue T `gozod:\"required\"`\n}\n",
			wantType:     "T",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helper := NewTestHelper(t)
			source := "package main\n\n" + tt.declarations
			if tt.fieldType != "" {
				source += "type Box struct {\n\tValue " + tt.fieldType + " `gozod:\"required\"`\n}\n"
			}
			helper.CreateGoFile("box.go", source)

			analyzer, err := NewStructAnalyzer()
			require.NoError(t, err)
			_, err = analyzer.AnalyzePackage(helper.GetTempDir())

			require.Error(t, err)
			assert.ErrorContains(t, err, "Value")
			assert.ErrorContains(t, err, tt.wantType)
			assert.ErrorContains(t, err, "unsupported field type")
		})
	}
}

func TestStructAnalyzer_ExtractFieldKey(t *testing.T) {
	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err)

	tests := []struct {
		name  string
		field *ast.Field
		want  string
		skip  bool
	}{
		{
			name:  "unnamed field without tag uses empty name",
			field: &ast.Field{},
			want:  "",
		},
		{
			name: "without tag uses field name",
			field: &ast.Field{
				Names: []*ast.Ident{{Name: "Name"}},
			},
			want: "Name",
		},
		{
			name: "json alias uses alias",
			field: &ast.Field{
				Names: []*ast.Ident{{Name: "Name"}},
				Tag:   &ast.BasicLit{Value: "`json:\"full_name,omitempty\"`"},
			},
			want: "full_name",
		},
		{
			name: "empty json tag falls back to field name",
			field: &ast.Field{
				Names: []*ast.Ident{{Name: "Alias"}},
				Tag:   &ast.BasicLit{Value: "`json:\"\"`"},
			},
			want: "Alias",
		},
		{
			name: "ignored json tag skips field",
			field: &ast.Field{
				Names: []*ast.Ident{{Name: "Hidden"}},
				Tag:   &ast.BasicLit{Value: "`json:\"-\"`"},
			},
			skip: true,
		},
		{
			name: "ignored json tag without field name skips field",
			field: &ast.Field{
				Tag: &ast.BasicLit{Value: "`json:\"-\"`"},
			},
			skip: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fallbackName := ""
			if len(tt.field.Names) > 0 {
				fallbackName = tt.field.Names[0].Name
			}
			got, skip := analyzer.extractFieldKey(tt.field, fallbackName)
			assert.Equal(t, tt.skip, skip)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractTagValue(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		tagName string
		want    string
	}{
		{
			name:    "extract existing tag",
			tag:     `json:"name,omitempty" gozod:"required,min=2"`,
			tagName: "gozod",
			want:    "required,min=2",
		},
		{
			name:    "missing tag returns empty string",
			tag:     `json:"name,omitempty"`,
			tagName: "gozod",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, extractTagValue(tt.tag, tt.tagName))
		})
	}
}

func TestGetTypeNameFromAST(t *testing.T) {
	tests := []struct {
		name string
		expr ast.Expr
		want string
	}{
		{
			name: "identifier",
			expr: &ast.Ident{Name: "string"},
			want: "string",
		},
		{
			name: "pointer",
			expr: &ast.StarExpr{X: &ast.Ident{Name: "User"}},
			want: "*User",
		},
		{
			name: "slice",
			expr: &ast.ArrayType{Elt: &ast.Ident{Name: "int"}},
			want: "[]int",
		},
		{
			name: "map",
			expr: &ast.MapType{Key: &ast.Ident{Name: "string"}, Value: &ast.Ident{Name: "int"}},
			want: "map[string]int",
		},
		{
			name: "selector",
			expr: &ast.SelectorExpr{X: &ast.Ident{Name: "time"}, Sel: &ast.Ident{Name: "Time"}},
			want: "time.Time",
		},
		{
			name: "selector without identifier prefix",
			expr: &ast.SelectorExpr{X: &ast.CallExpr{}, Sel: &ast.Ident{Name: "Name"}},
			want: "Name",
		},
		{
			name: "unknown expression",
			expr: &ast.CallExpr{},
			want: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, getTypeNameFromAST(tt.expr))
		})
	}
}

func TestStructAnalyzer_ParseTagRulesPreservesStructuredParams(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want []tagparser.TagRule
	}{
		{
			name: "simple rules",
			tag:  "required,min=2,max=10",
			want: []tagparser.TagRule{
				{Name: "required"},
				{Name: "min", Params: []string{"2"}},
				{Name: "max", Params: []string{"10"}},
			},
		},
		{
			name: "json object preserves commas",
			tag:  `required,meta={"label":"user,name","nested":{"min":1}},max=10`,
			want: []tagparser.TagRule{
				{Name: "required"},
				{Name: "meta", Params: []string{`{"label":"user,name","nested":{"min":1}}`}},
				{Name: "max", Params: []string{"10"}},
			},
		},
		{
			name: "json array preserves commas",
			tag:  `enum=["a,b","c"],required`,
			want: []tagparser.TagRule{
				{Name: "enum", Params: []string{`["a,b","c"]`}},
				{Name: "required"},
			},
		},
	}

	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := analyzer.parseTagRules(tt.tag)
			require.NoError(t, err)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("parseTagRules() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
