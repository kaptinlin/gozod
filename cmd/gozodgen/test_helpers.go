package main

import (
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHelper provides utilities for testing the code generator
type TestHelper struct {
	t       *testing.T
	tempDir string
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	tempDir := t.TempDir()
	return &TestHelper{
		t:       t,
		tempDir: tempDir,
	}
}

// CreateGoFile creates a Go source file in the temp directory
func (h *TestHelper) CreateGoFile(filename, content string) string {
	filePath := filepath.Join(h.tempDir, filename)

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		h.t.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	// Format the content
	formatted, err := format.Source([]byte(content))
	if err != nil {
		// If formatting fails, use original content
		formatted = []byte(content)
	}

	if err := os.WriteFile(filePath, formatted, 0600); err != nil {
		h.t.Fatalf("Failed to write file %s: %v", filePath, err)
	}

	return filePath
}

// ReadGeneratedFile reads a generated file and returns its content
func (h *TestHelper) ReadGeneratedFile(filename string) string {
	filePath := filepath.Join(h.tempDir, filename)
	content, err := os.ReadFile(filePath) // #nosec G304 - Test helper with trusted temp directory
	if err != nil {
		h.t.Fatalf("Failed to read generated file %s: %v", filePath, err)
	}
	return string(content)
}

// AssertFileExists checks if a file exists in the temp directory
func (h *TestHelper) AssertFileExists(filename string) {
	filePath := filepath.Join(h.tempDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		h.t.Errorf("Expected file %s to exist", filePath)
	}
}

// AssertFileNotExists checks if a file does not exist in the temp directory
func (h *TestHelper) AssertFileNotExists(filename string) {
	filePath := filepath.Join(h.tempDir, filename)
	if _, err := os.Stat(filePath); err == nil {
		h.t.Errorf("Expected file %s to not exist", filePath)
	}
}

// AssertCodeContains checks if generated code contains expected patterns
func (h *TestHelper) AssertCodeContains(code string, expected ...string) {
	for _, exp := range expected {
		if !strings.Contains(code, exp) {
			h.t.Errorf("Generated code does not contain expected pattern: %s\nGenerated code:\n%s", exp, code)
		}
	}
}

// AssertCodeNotContains checks if generated code does not contain unwanted patterns
func (h *TestHelper) AssertCodeNotContains(code string, unwanted ...string) {
	for _, unw := range unwanted {
		if strings.Contains(code, unw) {
			h.t.Errorf("Generated code contains unwanted pattern: %s\nGenerated code:\n%s", unw, code)
		}
	}
}

// GetTempDir returns the temporary directory path
func (h *TestHelper) GetTempDir() string {
	return h.tempDir
}

// CreatePackageStructure creates a basic Go package structure
func (h *TestHelper) CreatePackageStructure(packageName string) string {
	pkgDir := filepath.Join(h.tempDir, packageName)
	if err := os.MkdirAll(pkgDir, 0750); err != nil {
		h.t.Fatalf("Failed to create package directory %s: %v", pkgDir, err)
	}
	return pkgDir
}

// ListGeneratedFiles returns all generated files in the temp directory
func (h *TestHelper) ListGeneratedFiles() []string {
	var files []string
	err := filepath.WalkDir(h.tempDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".go") {
			relPath, _ := filepath.Rel(h.tempDir, path)
			files = append(files, relPath)
		}
		return nil
	})
	if err != nil {
		h.t.Fatalf("Failed to list generated files: %v", err)
	}
	return files
}

// TestStruct represents common test struct patterns
type TestStruct struct {
	Name    string
	Package string
	Fields  []TestField
}

// TestField represents a test struct field
type TestField struct {
	Name     string
	Type     string
	JsonTag  string
	GozodTag string
}

// GenerateStructCode generates Go struct code from TestStruct
func (ts TestStruct) GenerateStructCode() string {
	var sb strings.Builder

	if ts.Package != "" {
		sb.WriteString(fmt.Sprintf("package %s\n\n", ts.Package))
	}

	sb.WriteString(fmt.Sprintf("type %s struct {\n", ts.Name))
	for _, field := range ts.Fields {
		sb.WriteString(fmt.Sprintf("\t%s %s", field.Name, field.Type))

		var tags []string
		if field.JsonTag != "" {
			tags = append(tags, fmt.Sprintf(`json:"%s"`, field.JsonTag))
		}
		if field.GozodTag != "" {
			tags = append(tags, fmt.Sprintf(`gozod:"%s"`, field.GozodTag))
		}

		if len(tags) > 0 {
			sb.WriteString(fmt.Sprintf(" `%s`", strings.Join(tags, " ")))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("}\n")

	return sb.String()
}

// CommonTestStructs provides commonly used test structures
var CommonTestStructs = map[string]TestStruct{
	"SimpleUser": {
		Name:    "User",
		Package: "main",
		Fields: []TestField{
			{Name: "ID", Type: "string", JsonTag: "id", GozodTag: "required,uuid"},
			{Name: "Name", Type: "string", JsonTag: "name", GozodTag: "required,min=2,max=50"},
			{Name: "Email", Type: "string", JsonTag: "email", GozodTag: "required,email"},
			{Name: "Age", Type: "int", JsonTag: "age", GozodTag: "required,min=18,max=120"},
		},
	},
	"ComplexProduct": {
		Name:    "Product",
		Package: "main",
		Fields: []TestField{
			{Name: "ID", Type: "string", JsonTag: "id", GozodTag: "required,uuid"},
			{Name: "Name", Type: "string", JsonTag: "name", GozodTag: "required,min=1,max=200"},
			{Name: "Price", Type: "float64", JsonTag: "price", GozodTag: "required,gt=0.0"},
			{Name: "Tags", Type: "[]string", JsonTag: "tags", GozodTag: "min=0,max=10"},
			{Name: "Active", Type: "*bool", JsonTag: "active", GozodTag: "default=true"},
		},
	},
	"CircularRef": {
		Name:    "Node",
		Package: "main",
		Fields: []TestField{
			{Name: "Value", Type: "int", JsonTag: "value", GozodTag: "required"},
			{Name: "Next", Type: "*Node", JsonTag: "next", GozodTag: ""},
			{Name: "Children", Type: "[]*Node", JsonTag: "children", GozodTag: ""},
		},
	},
}

// AssertValidGoCode checks if the generated code is valid Go code
func (h *TestHelper) AssertValidGoCode(code string) {
	_, err := format.Source([]byte(code))
	if err != nil {
		h.t.Errorf("Generated code is not valid Go: %v\nCode:\n%s", err, code)
	}
}

// AssertImportsCorrect checks if the generated imports are correct
func (h *TestHelper) AssertImportsCorrect(code string, expectedImports []string, unexpectedImports []string) {
	for _, imp := range expectedImports {
		if !strings.Contains(code, fmt.Sprintf(`"%s"`, imp)) {
			h.t.Errorf("Generated code should import %s\nCode:\n%s", imp, code)
		}
	}

	for _, imp := range unexpectedImports {
		if strings.Contains(code, fmt.Sprintf(`"%s"`, imp)) {
			h.t.Errorf("Generated code should not import %s\nCode:\n%s", imp, code)
		}
	}
}
