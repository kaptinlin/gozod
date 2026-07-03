package main

import (
	"go/ast"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/gozod/pkg/tagparser"
)

func TestStructAnalyzer_ExtractFieldKey_FieldNameTag(t *testing.T) {
	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err)
	analyzer.fieldNameTag = "yaml"

	field := &ast.Field{
		Names: []*ast.Ident{{Name: "UserName"}},
		Tag:   &ast.BasicLit{Value: "`json:\"userName\" yaml:\"user_name\"`"},
	}
	fieldKey, skip := analyzer.extractFieldKey(field, "UserName")
	assert.False(t, skip)
	assert.Equal(t, "user_name", fieldKey)

	// A field without the chosen field-name tag falls back to the Go field name.
	noYAML := &ast.Field{
		Names: []*ast.Ident{{Name: "UserName"}},
		Tag:   &ast.BasicLit{Value: "`json:\"userName\"`"},
	}
	fieldKey, skip = analyzer.extractFieldKey(noYAML, "UserName")
	assert.False(t, skip)
	assert.Equal(t, "UserName", fieldKey)

	ignored := &ast.Field{
		Names: []*ast.Ident{{Name: "Secret"}},
		Tag:   &ast.BasicLit{Value: "`yaml:\"-\"`"},
	}
	_, skip = analyzer.extractFieldKey(ignored, "Secret")
	assert.True(t, skip)
}

func TestStructAnalyzer_ApplyRuleTagName(t *testing.T) {
	analyzer, err := NewStructAnalyzer()
	require.NoError(t, err)
	analyzer.ruleTagName = "validate"

	astField := &ast.Field{
		Names: []*ast.Ident{{Name: "Name"}},
		Tag:   &ast.BasicLit{Value: "`json:\"name\" validate:\"required,min=2\" gozod:\"max=99\"`"},
	}
	info := tagparser.FieldInfo{
		Name:     "Name",
		Type:     reflect.TypeFor[string](),
		FieldKey: "name",
	}

	hasTag, err := analyzer.applyGoZodTag(&info, astField)
	require.NoError(t, err)
	require.True(t, hasTag)
	assert.Equal(t, "required,min=2", info.GoZodTag)
	assert.True(t, info.Required)
	require.Len(t, info.Rules, 2)
	assert.Equal(t, "required", info.Rules[0].Name)
	assert.Equal(t, "min", info.Rules[1].Name)
	assert.Equal(t, []string{"2"}, info.Rules[1].Params)
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
			FieldKey: "name",
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
