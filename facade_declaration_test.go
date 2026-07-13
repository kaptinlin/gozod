package gozod

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestPrimitiveFacadeHasNoRebindableCallables(t *testing.T) {
	assertNoUnexpectedExportedVars(t, "constructors_primitives.go", map[string]struct{}{
		"PrecisionMinute":      {},
		"PrecisionSecond":      {},
		"PrecisionDecisecond":  {},
		"PrecisionCentisecond": {},
		"PrecisionMillisecond": {},
		"PrecisionMicrosecond": {},
		"PrecisionNanosecond":  {},
	})
}

func TestCompositeFacadeHasNoRebindableCallables(t *testing.T) {
	assertNoUnexpectedExportedVars(t, "constructors_composite.go", nil)
}

func TestSpecialFacadeHasNoRebindableCallables(t *testing.T) {
	assertNoUnexpectedExportedVars(t, "constructors_special.go", nil)
}

func TestConfigFacadeHasNoRebindableCallables(t *testing.T) {
	assertNoUnexpectedExportedVars(t, "core.go", nil)
}

func TestErrorFacadeHasNoRebindableCallables(t *testing.T) {
	assertNoUnexpectedExportedVars(t, "errors.go", map[string]struct{}{
		"ErrNilDiscriminatedUnionOption": {},
		"ErrMissingDiscriminatorValues":  {},
		"ErrDuplicateDiscriminator":      {},
		"ErrNoValidDiscriminators":       {},
	})
}

func TestJSONSchemaFacadeHasNoRebindableCallables(t *testing.T) {
	assertNoUnexpectedExportedVars(t, "jsonschema.go", map[string]struct{}{
		"ErrUnsupportedInputType":          {},
		"ErrCircularReference":             {},
		"ErrUnrepresentableType":           {},
		"ErrSchemaNotObjectOrStruct":       {},
		"ErrSliceElementNotSchema":         {},
		"ErrArrayItemNotSchema":            {},
		"ErrUnhandledArrayLike":            {},
		"ErrUnionInvalid":                  {},
		"ErrUnionNoMembers":                {},
		"ErrIntersectionInvalid":           {},
		"ErrInvalidEnumSchema":             {},
		"ErrEnumExtractValues":             {},
		"ErrLiteralNoValuesMethod":         {},
		"ErrLiteralUnexpectedReturnValues": {},
		"ErrExpectedDiscriminatedUnion":    {},
		"ErrExpectedRecord":                {},
		"ErrRecordValueNotSchema":          {},
		"ErrInvalidRegistrySchemaID":       {},
		"ErrMapNoMethods":                  {},
		"ErrMapKeyNotSchema":               {},
		"ErrMapValueNotSchema":             {},
		"ErrInvalidJSONSchemaOption":       {},
		"ErrUnsupportedJSONSchemaTarget":   {},
		"ErrUnsupportedJSONSchemaType":     {},
		"ErrUnsupportedJSONSchemaKeyword":  {},
		"ErrInvalidJSONSchema":             {},
		"ErrJSONSchemaCircularRef":         {},
		"ErrJSONSchemaPatternCompile":      {},
		"ErrJSONSchemaIfThenElse":          {},
		"ErrJSONSchemaPatternProperties":   {},
		"ErrJSONSchemaDynamicRef":          {},
		"ErrJSONSchemaUnevaluatedProps":    {},
		"ErrJSONSchemaUnevaluatedItems":    {},
		"ErrJSONSchemaDependentSchemas":    {},
		"ErrJSONSchemaPropertyNames":       {},
		"ErrJSONSchemaContains":            {},
	})
}

func assertNoUnexpectedExportedVars(t *testing.T, filename string, allowed map[string]struct{}) {
	t.Helper()

	file, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", filename, err)
	}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}
		for _, spec := range gen.Specs {
			for _, name := range spec.(*ast.ValueSpec).Names {
				if !name.IsExported() {
					continue
				}
				if _, ok := allowed[name.Name]; !ok {
					t.Errorf("%s exposes rebindable callable %s", filename, name.Name)
				}
			}
		}
	}
}
