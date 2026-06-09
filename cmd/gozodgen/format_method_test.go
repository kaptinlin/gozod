package main

import (
	"go/ast"
	"reflect"
	"testing"

	"github.com/kaptinlin/gozod/pkg/tagparser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStructAnalyzer_ExtractFieldName_Format(t *testing.T) {
	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err)
	analyzer.format = "yaml"

	field := &ast.Field{
		Names: []*ast.Ident{{Name: "UserName"}},
		Tag:   &ast.BasicLit{Value: "`json:\"userName\" yaml:\"user_name\"`"},
	}
	assert.Equal(t, "user_name", analyzer.extractFieldName(field))

	// A field without the chosen format tag falls back to the Go field name.
	noYAML := &ast.Field{
		Names: []*ast.Ident{{Name: "UserName"}},
		Tag:   &ast.BasicLit{Value: "`json:\"userName\"`"},
	}
	assert.Equal(t, "UserName", analyzer.extractFieldName(noYAML))
}

func TestFileWriter_MethodName(t *testing.T) {
	writer, err := NewFileWriter("", "main", "_gen.go", true, false)
	require.NoError(t, err)
	writer.methodName = "Validate"

	content, err := writer.generateCode(&GenerationInfo{
		Name:    "User",
		Package: "main",
		Fields: []tagparser.FieldInfo{{
			Name:     "Name",
			JSONName: "name",
			Type:     reflect.TypeFor[string](),
			Rules:    []tagparser.TagRule{{Name: "required"}},
		}},
	})
	require.NoError(t, err)
	assert.Contains(t, content, ") Validate() *gozod.ZodStruct[User, User]")
	assert.NotContains(t, content, ") Schema() *gozod.ZodStruct")
}

func TestIsExportedIdent(t *testing.T) {
	valid := []string{"Schema", "Validate", "GozodSchema", "Schema2", "A_b"}
	for _, s := range valid {
		assert.True(t, isExportedIdent(s), "%q should be valid", s)
	}

	invalid := []string{"", "schema", "1Schema", "Sch ema", "Schema-2", "_Schema"}
	for _, s := range invalid {
		assert.False(t, isExportedIdent(s), "%q should be invalid", s)
	}
}
