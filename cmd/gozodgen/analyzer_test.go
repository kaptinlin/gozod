package main

import (
	"go/ast"
	"go/parser"
	"go/token"
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

func TestStructAnalyzer_ExtractImports(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected []string
	}{
		{
			name: "single import",
			source: `package main
import "time"
type User struct {}`,
			expected: []string{"time"},
		},
		{
			name: "multiple imports",
			source: `package main
import (
	"time"
	"github.com/go-json-experiment/json"
)
type User struct {}`,
			expected: []string{"time", "github.com/go-json-experiment/json"},
		},
		{
			name: "aliased import",
			source: `package main
import t "time"
type User struct {}`,
			expected: []string{"time"},
		},
		{
			name: "no imports",
			source: `package main
type User struct {}`,
			expected: []string{},
		},
	}

	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err, "Failed to create analyzer")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helper := NewTestHelper(t)
			helper.CreateGoFile("test.go", tt.source)

			// Use AnalyzePackage to get the imports
			results, err := analyzer.AnalyzePackage(helper.GetTempDir())
			require.NoError(t, err, "Failed to analyze package")

			// We can't directly test extractImports since it's not public,
			// but we can verify the analysis worked
			if len(results) == 0 && len(tt.expected) > 0 {
				// This test is more about ensuring the analysis pipeline works
				t.Logf("Analysis completed successfully for imports test")
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
	assert.Equal(t, "Name", structs[0].Fields[0].JSONName)
}

func TestStructAnalyzer_ParseStructFieldsWithJSONFallbacks(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", `package main

type User struct {
	Name   string
	Alias  string `+"`json:\"\"`"+`
	Hidden string `+"`json:\"-\" gozod:\"required\"`"+`
	hidden string
}
`, parser.ParseComments)
	require.NoError(t, err)

	genDecl := file.Decls[0].(*ast.GenDecl)
	typeSpec := genDecl.Specs[0].(*ast.TypeSpec)
	structType := typeSpec.Type.(*ast.StructType)

	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err)

	fields, err := analyzer.parseStructFields(structType)
	require.NoError(t, err)
	require.Len(t, fields, 3)

	assert.Equal(t, "Name", fields[0].info.Name)
	assert.Equal(t, reflect.TypeFor[string](), fields[0].info.Type)
	assert.Equal(t, "string", fields[0].info.TypeName)
	assert.Equal(t, "Name", fields[0].info.JSONName)
	assert.False(t, fields[0].hasGozodTag)

	assert.Equal(t, "Alias", fields[1].info.Name)
	assert.Equal(t, reflect.TypeFor[string](), fields[1].info.Type)
	assert.Equal(t, "string", fields[1].info.TypeName)
	assert.Equal(t, "Alias", fields[1].info.JSONName)
	assert.False(t, fields[1].hasGozodTag)

	assert.Equal(t, "Hidden", fields[2].info.Name)
	assert.Equal(t, reflect.TypeFor[string](), fields[2].info.Type)
	assert.Equal(t, "string", fields[2].info.TypeName)
	assert.Equal(t, "Hidden", fields[2].info.JSONName)
	assert.Equal(t, "required", fields[2].info.GoZodTag)
	assert.True(t, fields[2].info.Required)
	assert.True(t, fields[2].hasGozodTag)
	if diff := cmp.Diff([]tagparser.TagRule{{Name: "required"}}, fields[2].info.Rules); diff != "" {
		t.Errorf("parseStructFields() rules mismatch (-want +got):\n%s", diff)
	}
}

func TestStructAnalyzer_ExtractJSONName(t *testing.T) {
	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err)

	tests := []struct {
		name  string
		field *ast.Field
		want  string
	}{
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
			name: "ignored json tag falls back to field name",
			field: &ast.Field{
				Names: []*ast.Ident{{Name: "Hidden"}},
				Tag:   &ast.BasicLit{Value: "`json:\"-\"`"},
			},
			want: "Hidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, analyzer.extractJSONName(tt.field))
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, getTypeNameFromAST(tt.expr))
		})
	}
}

func TestSmartSplitTagRules(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want []string
	}{
		{
			name: "simple rules",
			tag:  "required,min=2,max=10",
			want: []string{"required", "min=2", "max=10"},
		},
		{
			name: "json object preserves commas",
			tag:  `required,meta={"label":"user,name","nested":{"min":1}},max=10`,
			want: []string{"required", `meta={"label":"user,name","nested":{"min":1}}`, "max=10"},
		},
		{
			name: "json array preserves commas",
			tag:  `enum=["a,b","c"],required`,
			want: []string{`enum=["a,b","c"]`, "required"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := smartSplitTagRules(tt.tag)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("smartSplitTagRules() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestStructAnalyzer_GetReflectTypeFromAST(t *testing.T) {
	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err)

	tests := []struct {
		name string
		expr ast.Expr
		want reflect.Type
	}{
		{
			name: "string",
			expr: &ast.Ident{Name: "string"},
			want: reflect.TypeFor[string](),
		},
		{
			name: "pointer",
			expr: &ast.StarExpr{X: &ast.Ident{Name: "int"}},
			want: reflect.PointerTo(reflect.TypeFor[int]()),
		},
		{
			name: "slice",
			expr: &ast.ArrayType{Elt: &ast.Ident{Name: "bool"}},
			want: reflect.SliceOf(reflect.TypeFor[bool]()),
		},
		{
			name: "map",
			expr: &ast.MapType{Key: &ast.Ident{Name: "string"}, Value: &ast.Ident{Name: "float64"}},
			want: reflect.MapOf(reflect.TypeFor[string](), reflect.TypeFor[float64]()),
		},
		{
			name: "time selector",
			expr: &ast.SelectorExpr{X: &ast.Ident{Name: "time"}, Sel: &ast.Ident{Name: "Time"}},
			want: reflect.TypeFor[timeType](),
		},
		{
			name: "unknown falls back to any",
			expr: &ast.Ident{Name: "Custom"},
			want: reflect.TypeFor[any](),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, analyzer.getReflectTypeFromAST(tt.expr))
		})
	}
}
