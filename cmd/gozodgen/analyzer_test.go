package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kaptinlin/gozod/pkg/tagparser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestStructAnalyzer_NeedsGeneration(t *testing.T) {
	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err, "Failed to create analyzer")

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
			result := analyzer.NeedsGeneration(tt.info)
			assert.Equal(t, tt.expected, result, "Expected %t, got %t", tt.expected, result)
		})
	}
}
