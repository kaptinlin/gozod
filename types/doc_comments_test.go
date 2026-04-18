package types

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypesPackageDocs(t *testing.T) {
	typesDir, err := filepath.Abs(".")
	require.NoError(t, err)

	entries, err := os.ReadDir(typesDir)
	require.NoError(t, err)

	fset := token.NewFileSet()
	var docFile *ast.File

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".go" || name == "doc_comments_test.go" {
			continue
		}

		file, err := parser.ParseFile(fset, filepath.Join(typesDir, name), nil, parser.ParseComments)
		require.NoError(t, err)

		if name == "doc.go" {
			docFile = file
			continue
		}
		assert.Nilf(t, file.Doc, "%s should not carry the package doc comment", name)
	}

	require.NotNil(t, docFile)
	require.NotNil(t, docFile.Doc)
	assert.Equal(t, "Package types provides public schema implementations for GoZod validation types.", strings.TrimSpace(docFile.Doc.Text()))
}

func TestTargetedExportedDocs(t *testing.T) {
	typesDir, err := filepath.Abs(".")
	require.NoError(t, err)

	assertDocComments(t, filepath.Join(typesDir, "object.go"), map[string]struct{}{
		"ErrNilObjectPointer":   {},
		"ErrPickRefinements":    {},
		"ErrOmitRefinements":    {},
		"ErrExtendRefinements":  {},
		"ErrUnrecognizedKey":    {},
		"ObjectModeStrict":      {},
		"ObjectModeStrip":       {},
		"ObjectModePassthrough": {},
	})
	assertDocComments(t, filepath.Join(typesDir, "enum.go"), map[string]struct{}{
		"ZodEnumDef":       {},
		"ZodEnumInternals": {},
	})
	assertDocComments(t, filepath.Join(typesDir, "struct.go"), map[string]struct{}{
		"ErrFieldNotFoundOrNotSettable": {},
		"ErrCannotAssignToField":        {},
		"ErrValueMustBeStruct":          {},
		"ZodStructDef":                  {},
		"ZodStructInternals":            {},
		"ZodStruct":                     {},
		"Struct":                        {},
		"StructPtr":                     {},
		"FromStructOption":              {},
		"WithTagName":                   {},
		"FromStruct":                    {},
		"FromStructPtr":                 {},
	})
}

func assertDocComments(t *testing.T, filePath string, names map[string]struct{}) {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	require.NoError(t, err)

	seen := map[string]bool{}

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if _, ok := names[s.Name.Name]; ok {
						require.NotNilf(t, d.Doc, "%s should have a doc comment", s.Name.Name)
						assert.Contains(t, d.Doc.Text(), s.Name.Name)
						seen[s.Name.Name] = true
					}
				case *ast.ValueSpec:
					for _, ident := range s.Names {
						if _, ok := names[ident.Name]; !ok {
							continue
						}
						doc := s.Doc
						if doc == nil {
							doc = d.Doc
						}
						require.NotNilf(t, doc, "%s should have a doc comment", ident.Name)
						assert.Contains(t, doc.Text(), ident.Name)
						seen[ident.Name] = true
					}
				}
			}
		case *ast.FuncDecl:
			if _, ok := names[d.Name.Name]; ok {
				require.NotNilf(t, d.Doc, "%s should have a doc comment", d.Name.Name)
				assert.Contains(t, d.Doc.Text(), d.Name.Name)
				seen[d.Name.Name] = true
			}
		}
	}

	for name := range names {
		assert.Truef(t, seen[name], "%s was not found in %s", name, filePath)
	}
}
