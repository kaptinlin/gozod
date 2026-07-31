package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/kaptinlin/gozod/pkg/tagparser"
)

var errUnsupportedFieldType = errors.New("unsupported field type")

// timeType is a marker type for time.Time detection.
type timeType struct{}

// StructAnalyzer analyzes Go source files to find structs requiring code generation.
type StructAnalyzer struct {
	fset         *token.FileSet
	info         *types.Info
	ruleTagName  string // struct tag used for validation rules (default "gozod")
	fieldNameTag string // struct tag used for field names (default "json")
}

// GenerationInfo contains information about a struct that needs code generation.
type GenerationInfo struct {
	Name        string                // Struct name
	Package     string                // Package name
	Fields      []tagparser.FieldInfo // Field information from tagparser
	Imports     []string              // Required imports
	HasGenerate bool                  // Whether struct has //go:generate gozodgen directive
	FilePath    string                // Source file path
}

// NewStructAnalyzer creates a new AST analyzer instance.
func NewStructAnalyzer() (*StructAnalyzer, error) {
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	return &StructAnalyzer{
		fset:         token.NewFileSet(),
		info:         info,
		ruleTagName:  defaultRuleTag,
		fieldNameTag: defaultFieldNameTag,
	}, nil
}

// AnalyzePackage analyzes all Go files in a package directory.
func (a *StructAnalyzer) AnalyzePackage(pkgPath string) ([]*GenerationInfo, error) {
	loaded, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo |
			packages.NeedImports | packages.NeedDeps,
		Dir:   pkgPath,
		Tests: false,
	}, ".")
	if err != nil {
		return nil, fmt.Errorf("load package %s: %w", pkgPath, err)
	}
	if len(loaded) != 1 {
		return nil, fmt.Errorf("load package %s: expected one package, got %d", pkgPath, len(loaded))
	}
	pkg := loaded[0]
	if len(pkg.Errors) > 0 {
		messages := make([]string, len(pkg.Errors))
		for i, loadErr := range pkg.Errors {
			messages[i] = loadErr.Error()
		}
		return nil, fmt.Errorf("load package %s: %s", pkgPath, strings.Join(messages, "; "))
	}

	a.fset = pkg.Fset
	a.info = pkg.TypesInfo

	allStructs := make([]*GenerationInfo, 0)
	for _, file := range pkg.Syntax {
		fileName := a.fset.Position(file.Pos()).Filename
		structs, err := a.analyzeFile(fileName, file, pkg.Name)
		if err != nil {
			return nil, fmt.Errorf("analyze file %s: %w", fileName, err)
		}
		allStructs = append(allStructs, structs...)
	}
	return allStructs, nil
}

// analyzeFile analyzes a single Go file for structs requiring generation.
func (a *StructAnalyzer) analyzeFile(fileName string, file *ast.File, pkgName string) ([]*GenerationInfo, error) {
	var structs []*GenerationInfo
	imports := a.extractImports(file)

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		hasGenerate := hasGenerateDirective(genDecl.Doc)

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			info, err := a.analyzeStruct(typeSpec.Name.Name, structType, pkgName, fileName, imports, hasGenerate)
			if err != nil {
				return nil, fmt.Errorf("analyze struct %s: %w", typeSpec.Name.Name, err)
			}
			if info != nil && (len(info.Fields) > 0 || hasGenerate) {
				structs = append(structs, info)
			}
		}
	}

	return structs, nil
}

// analyzeStruct analyzes a single struct declaration.
func (a *StructAnalyzer) analyzeStruct(name string, structType *ast.StructType, pkgName, fileName string, imports []string, hasGenerate bool) (*GenerationInfo, error) {
	fields, err := a.parseStructFields(structType, hasGenerate)
	if err != nil {
		return nil, fmt.Errorf("parse struct fields: %w", err)
	}

	return &GenerationInfo{
		Name:        name,
		Package:     pkgName,
		Fields:      fields,
		Imports:     imports,
		HasGenerate: hasGenerate,
		FilePath:    fileName,
	}, nil
}

// extractImports extracts import statements from a file.
func (a *StructAnalyzer) extractImports(file *ast.File) []string {
	imports := make([]string, 0, len(file.Imports))

	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		imports = append(imports, importPath)
	}

	return imports
}

// hasGenerateDirective reports whether comments contain //go:generate gozodgen.
func hasGenerateDirective(comments *ast.CommentGroup) bool {
	if comments == nil {
		return false
	}

	for _, comment := range comments.List {
		text := strings.TrimSpace(comment.Text)
		if strings.HasPrefix(text, "//go:generate") && strings.Contains(text, "gozodgen") {
			return true
		}
	}

	return false
}

// parseStructFields parses struct fields from AST and extracts tag information.
func (a *StructAnalyzer) parseStructFields(structType *ast.StructType, hasGenerate bool) ([]tagparser.FieldInfo, error) {
	var fields []tagparser.FieldInfo

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue
		}

		for _, name := range field.Names {
			if !name.IsExported() {
				continue
			}

			fieldKey, skip := a.extractFieldKey(field, name.Name)
			if skip {
				continue
			}

			info := tagparser.FieldInfo{
				Name:     name.Name,
				TypeName: getTypeNameFromAST(field.Type),
				FieldKey: fieldKey,
			}

			hasGozodTag, err := a.applyGoZodTag(&info, field)
			if err != nil {
				return nil, fmt.Errorf("parse gozod tag for field %s: %w", name.Name, err)
			}
			if !hasGozodTag && !hasGenerate {
				continue
			}

			fieldType, err := a.getReflectType(field.Type)
			if err != nil {
				return nil, fmt.Errorf("resolve type for field %s: %w", name.Name, err)
			}
			info.Type = fieldType

			fields = append(fields, info)
		}
	}

	return fields, nil
}

// getReflectType converts a type-checked AST expression to reflect.Type.
func (a *StructAnalyzer) getReflectType(expr ast.Expr) (reflect.Type, error) {
	if a.info == nil {
		return nil, errors.New("type information is unavailable")
	}
	typeAndValue, ok := a.info.Types[expr]
	if !ok || typeAndValue.Type == nil {
		return nil, fmt.Errorf("type information is unavailable at %s", a.fset.Position(expr.Pos()))
	}
	fieldType, err := a.typesToReflectType(typeAndValue.Type)
	if err != nil {
		return nil, fmt.Errorf("%w %s", err, types.TypeString(typeAndValue.Type, nil))
	}
	return fieldType, nil
}

// typesToReflectType converts go/types.Type to reflect.Type.
func (a *StructAnalyzer) typesToReflectType(t types.Type) (reflect.Type, error) {
	t = types.Unalias(t)
	switch typ := t.(type) {
	case *types.Basic:
		if converted := basicKindToReflectType(typ.Kind()); converted != nil {
			return converted, nil
		}
		return nil, errUnsupportedFieldType
	case *types.Pointer:
		elem, err := a.typesToReflectType(typ.Elem())
		if err != nil {
			return nil, err
		}
		return reflect.PointerTo(elem), nil
	case *types.Slice:
		elem, err := a.typesToReflectType(typ.Elem())
		if err != nil {
			return nil, err
		}
		return reflect.SliceOf(elem), nil
	case *types.Array:
		elem, err := a.typesToReflectType(typ.Elem())
		if err != nil {
			return nil, err
		}
		return reflect.ArrayOf(int(typ.Len()), elem), nil
	case *types.Map:
		key, err := a.typesToReflectType(typ.Key())
		if err != nil {
			return nil, err
		}
		if key.Kind() != reflect.String {
			return nil, errUnsupportedFieldType
		}
		elem, err := a.typesToReflectType(typ.Elem())
		if err != nil {
			return nil, err
		}
		return reflect.MapOf(key, elem), nil
	case *types.Named:
		obj := typ.Obj()
		if obj != nil && obj.Pkg() != nil && obj.Pkg().Path() == "time" && obj.Name() == "Time" {
			return reflect.TypeFor[timeType](), nil
		}
		if _, ok := typ.Underlying().(*types.Struct); ok {
			return reflect.TypeFor[any](), nil
		}
		return a.typesToReflectType(typ.Underlying())
	case *types.Interface:
		if typ.Empty() {
			return reflect.TypeFor[any](), nil
		}
		return nil, errUnsupportedFieldType
	default:
		return nil, errUnsupportedFieldType
	}
}

// basicKindToReflectType maps go/types basic kinds to reflect.Type.
func basicKindToReflectType(kind types.BasicKind) reflect.Type {
	switch kind {
	case types.String:
		return reflect.TypeFor[string]()
	case types.Int:
		return reflect.TypeFor[int]()
	case types.Int8:
		return reflect.TypeFor[int8]()
	case types.Int16:
		return reflect.TypeFor[int16]()
	case types.Int32:
		return reflect.TypeFor[int32]()
	case types.Int64:
		return reflect.TypeFor[int64]()
	case types.Uint:
		return reflect.TypeFor[uint]()
	case types.Uint8:
		return reflect.TypeFor[uint8]()
	case types.Uint16:
		return reflect.TypeFor[uint16]()
	case types.Uint32:
		return reflect.TypeFor[uint32]()
	case types.Uint64:
		return reflect.TypeFor[uint64]()
	case types.Float32:
		return reflect.TypeFor[float32]()
	case types.Float64:
		return reflect.TypeFor[float64]()
	case types.Complex64:
		return reflect.TypeFor[complex64]()
	case types.Complex128:
		return reflect.TypeFor[complex128]()
	case types.Bool:
		return reflect.TypeFor[bool]()
	default:
		return nil
	}
}

// NeedsGeneration reports whether a struct needs code generation.
func NeedsGeneration(info *GenerationInfo) bool {
	if info.HasGenerate {
		return true
	}
	for _, field := range info.Fields {
		if field.HasSchemaSpec() {
			return true
		}
	}
	return false
}

func (a *StructAnalyzer) applyGoZodTag(info *tagparser.FieldInfo, field *ast.Field) (bool, error) {
	if field.Tag == nil {
		return false, nil
	}

	tagValue := strings.Trim(field.Tag.Value, "`")
	ruleTag, ok := lookupTagValue(tagValue, a.ruleTagName)
	if !ok {
		return false, nil
	}

	info.GoZodTag = ruleTag
	if info.GoZodTag == "" {
		return true, nil
	}

	rules, err := a.parseTagRules(info.GoZodTag)
	if err != nil {
		return true, err
	}
	info.Rules = rules
	info.Required = info.HasRule("required")
	return true, nil
}

// extractFieldKey extracts the schema key from the configured field-name tag.
func (a *StructAnalyzer) extractFieldKey(field *ast.Field, fallbackName string) (string, bool) {
	if field.Tag == nil {
		return fallbackName, false
	}

	tagValue := strings.Trim(field.Tag.Value, "`")
	fieldNameTagValue := extractTagValue(tagValue, a.fieldNameTag)
	if fieldNameTagValue == "" {
		return fallbackName, false
	}

	name, _, _ := strings.Cut(fieldNameTagValue, ",")
	name = strings.TrimSpace(name)
	if name == "-" {
		return "", true
	}
	if name == "" {
		return fallbackName, false
	}
	return name, false
}

// extractTagValue returns the value for a specific tag key.
func extractTagValue(tagString, tagName string) string {
	return reflect.StructTag(tagString).Get(tagName)
}

func lookupTagValue(tagString, tagName string) (string, bool) {
	return reflect.StructTag(tagString).Lookup(tagName)
}

// parseTagRules parses validation tag rules with proper handling of complex JSON values.
func (a *StructAnalyzer) parseTagRules(tagValue string) ([]tagparser.TagRule, error) {
	return tagparser.New().ParseTagString(tagValue)
}

// getTypeNameFromAST extracts the type name string from an AST expression.
func getTypeNameFromAST(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + getTypeNameFromAST(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + getTypeNameFromAST(t.Elt)
		}
		var length strings.Builder
		if err := format.Node(&length, token.NewFileSet(), t.Len); err != nil {
			return "unknown"
		}
		return "[" + length.String() + "]" + getTypeNameFromAST(t.Elt)
	case *ast.MapType:
		return "map[" + getTypeNameFromAST(t.Key) + "]" + getTypeNameFromAST(t.Value)
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name + "." + t.Sel.Name
		}
		return t.Sel.Name
	default:
		return "unknown"
	}
}
