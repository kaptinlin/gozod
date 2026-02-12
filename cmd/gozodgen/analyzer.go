// Package main provides the gozodgen code generation tool.
package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"reflect"
	"strings"

	"github.com/kaptinlin/gozod/pkg/tagparser"
)

var (
	errInvalidRuleFormat = errors.New("invalid rule format")
	errRuleRequiresParam = errors.New("rule requires a parameter")
	errEmptyRuleName     = errors.New("empty rule name")
)

// timeType is a marker type for time.Time detection
type timeType struct{}

// StructAnalyzer analyzes Go source files to find structs requiring code generation
type StructAnalyzer struct {
	fset     *token.FileSet
	packages map[string]*types.Package
	imports  map[string]string
	info     *types.Info
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

// NewStructAnalyzer creates a new AST analyzer instance
func NewStructAnalyzer() (*StructAnalyzer, error) {
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	return &StructAnalyzer{
		fset:     token.NewFileSet(),
		packages: make(map[string]*types.Package),
		imports:  make(map[string]string),
		info:     info,
	}, nil
}

// AnalyzePackage analyzes all Go files in a package directory.
func (a *StructAnalyzer) AnalyzePackage(pkgPath string) ([]*GenerationInfo, error) {
	astPkgs, err := parser.ParseDir(a.fset, pkgPath, nil, parser.ParseComments) //nolint:deprecated // ParseDir is sufficient for code generation
	if err != nil {
		return nil, fmt.Errorf("parse package %s: %w", pkgPath, err)
	}

	var allStructs []*GenerationInfo

	for pkgName, astPkg := range astPkgs {
		if strings.HasSuffix(pkgName, "_test") {
			continue
		}

		var files []*ast.File
		for _, file := range astPkg.Files {
			files = append(files, file)
		}

		conf := types.Config{Importer: importer.Default()}
		typesPkg, err := conf.Check(pkgName, a.fset, files, a.info)
		if err != nil {
			fmt.Printf("Warning: type checking for package %s: %v\n", pkgName, err)
		}
		if typesPkg != nil {
			a.packages[pkgName] = typesPkg
		}

		for fileName, file := range astPkg.Files {
			structs, err := a.analyzeFile(fileName, file, pkgName)
			if err != nil {
				return nil, fmt.Errorf("analyze file %s: %w", fileName, err)
			}
			allStructs = append(allStructs, structs...)
		}
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

// parsedField pairs a FieldInfo with a flag indicating whether the field had a gozod tag.
type parsedField struct {
	info        tagparser.FieldInfo
	hasGozodTag bool
}

// analyzeStruct analyzes a single struct declaration.
func (a *StructAnalyzer) analyzeStruct(name string, structType *ast.StructType, pkgName, fileName string, imports []string, hasGenerate bool) (*GenerationInfo, error) {
	parsed, err := a.parseStructFields(structType)
	if err != nil {
		return nil, fmt.Errorf("parse struct fields: %w", err)
	}

	var gozodFields []tagparser.FieldInfo
	for _, pf := range parsed {
		if len(pf.info.Rules) > 0 || pf.info.Required || pf.hasGozodTag || hasGenerate {
			pf.info.Optional = pf.info.Type != nil && pf.info.Type.Kind() == reflect.Pointer && !pf.info.Required
			gozodFields = append(gozodFields, pf.info)
		}
	}

	return &GenerationInfo{
		Name:        name,
		Package:     pkgName,
		Fields:      gozodFields,
		Imports:     imports,
		HasGenerate: hasGenerate,
		FilePath:    fileName,
	}, nil
}

// extractImports extracts import statements from a file
func (a *StructAnalyzer) extractImports(file *ast.File) []string {
	imports := make([]string, 0, len(file.Imports))

	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		imports = append(imports, importPath)

		// Store import alias if present
		if imp.Name != nil {
			a.imports[imp.Name.Name] = importPath
		}
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
func (a *StructAnalyzer) parseStructFields(structType *ast.StructType) ([]parsedField, error) {
	var fields []parsedField

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue
		}

		for _, name := range field.Names {
			if !name.IsExported() {
				continue
			}

			info := tagparser.FieldInfo{
				Name:     name.Name,
				Type:     a.getReflectType(field.Type),
				TypeName: getTypeNameFromAST(field.Type),
				JsonName: a.extractJSONName(field),
			}

			hasGozodTag := false
			if field.Tag != nil {
				tagValue := strings.Trim(field.Tag.Value, "`")
				if strings.Contains(tagValue, "gozod:") {
					hasGozodTag = true
					gozodTag := extractTagValue(tagValue, "gozod")
					info.GozodTag = gozodTag
					if gozodTag != "" {
						rules, err := a.parseTagRules(gozodTag)
						if err != nil {
							return nil, fmt.Errorf("parse gozod tag for field %s: %w", name.Name, err)
						}
						info.Rules = rules
						for _, rule := range rules {
							if rule.Name == "required" {
								info.Required = true
							}
						}
					}
				}
			}

			fields = append(fields, parsedField{info: info, hasGozodTag: hasGozodTag})
		}
	}

	return fields, nil
}

// getReflectType converts AST type expression to reflect.Type using go/types
func (a *StructAnalyzer) getReflectType(expr ast.Expr) reflect.Type {
	// Try to get type information from the type checker first
	if a.info != nil {
		if typeAndValue, ok := a.info.Types[expr]; ok {
			return a.typesToReflectType(typeAndValue.Type)
		}
	}

	// Fallback to AST-based type inference if type checker info is not available
	return a.getReflectTypeFromAST(expr)
}

// typesToReflectType converts go/types.Type to reflect.Type.
func (a *StructAnalyzer) typesToReflectType(t types.Type) reflect.Type {
	switch typ := t.(type) {
	case *types.Basic:
		return basicKindToReflectType(typ.Kind())
	case *types.Pointer:
		return reflect.PointerTo(a.typesToReflectType(typ.Elem()))
	case *types.Slice:
		return reflect.SliceOf(a.typesToReflectType(typ.Elem()))
	case *types.Array:
		return reflect.SliceOf(a.typesToReflectType(typ.Elem()))
	case *types.Map:
		return reflect.MapOf(a.typesToReflectType(typ.Key()), a.typesToReflectType(typ.Elem()))
	case *types.Named:
		obj := typ.Obj()
		if obj != nil && obj.Pkg() != nil && obj.Pkg().Path() == "time" && obj.Name() == "Time" {
			return reflect.TypeFor[timeType]()
		}
		return a.typesToReflectType(typ.Underlying())
	case *types.Interface:
		return reflect.TypeFor[any]()
	default:
		return reflect.TypeFor[any]()
	}
}

// basicKindToReflectType maps go/types basic kinds to reflect.Type.
func basicKindToReflectType(kind types.BasicKind) reflect.Type {
	switch kind { //nolint:exhaustive // only concrete types need mapping
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
		return reflect.TypeFor[any]()
	}
}

// getReflectTypeFromAST is the fallback AST-based type inference.
func (a *StructAnalyzer) getReflectTypeFromAST(expr ast.Expr) reflect.Type {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string":
			return reflect.TypeFor[string]()
		case "int":
			return reflect.TypeFor[int]()
		case "int32":
			return reflect.TypeFor[int32]()
		case "int64":
			return reflect.TypeFor[int64]()
		case "float32":
			return reflect.TypeFor[float32]()
		case "float64":
			return reflect.TypeFor[float64]()
		case "bool":
			return reflect.TypeFor[bool]()
		default:
			return reflect.TypeFor[any]()
		}
	case *ast.StarExpr:
		return reflect.PointerTo(a.getReflectTypeFromAST(t.X))
	case *ast.ArrayType:
		return reflect.SliceOf(a.getReflectTypeFromAST(t.Elt))
	case *ast.MapType:
		return reflect.MapOf(a.getReflectTypeFromAST(t.Key), a.getReflectTypeFromAST(t.Value))
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			if ident.Name == "time" && t.Sel.Name == "Time" {
				return reflect.TypeFor[timeType]()
			}
		}
		return reflect.TypeFor[any]()
	default:
		return reflect.TypeFor[any]()
	}
}

// NeedsGeneration reports whether a struct needs code generation.
func NeedsGeneration(info *GenerationInfo) bool {
	if info.HasGenerate {
		return true
	}
	for _, field := range info.Fields {
		if len(field.Rules) > 0 || field.Required {
			return true
		}
	}
	return false
}

// extractJSONName extracts the JSON field name from a struct tag.
func (a *StructAnalyzer) extractJSONName(field *ast.Field) string {
	if field.Tag == nil {
		if len(field.Names) > 0 {
			return field.Names[0].Name
		}
		return ""
	}

	tagValue := strings.Trim(field.Tag.Value, "`")
	jsonTag := extractTagValue(tagValue, "json")
	if jsonTag == "" {
		if len(field.Names) > 0 {
			return field.Names[0].Name
		}
		return ""
	}

	name, _, _ := strings.Cut(jsonTag, ",")
	name = strings.TrimSpace(name)
	if name == "" || name == "-" {
		if len(field.Names) > 0 {
			return field.Names[0].Name
		}
		return ""
	}
	return name
}

// extractTagValue returns the value for a specific tag key.
func extractTagValue(tagString, tagName string) string {
	return reflect.StructTag(tagString).Get(tagName)
}

// parseTagRules parses gozod tag rules with proper handling of complex JSON values
func (a *StructAnalyzer) parseTagRules(tagValue string) ([]tagparser.TagRule, error) {
	if tagValue == "" {
		return nil, nil
	}

	parts := smartSplitTagRules(tagValue)
	rules := make([]tagparser.TagRule, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		rule := tagparser.TagRule{}

		// Check if rule has parameters (contains =)
		if strings.Contains(part, "=") {
			ruleParts := strings.SplitN(part, "=", 2)
			if len(ruleParts) != 2 {
				return nil, fmt.Errorf("%w: %s", errInvalidRuleFormat, part)
			}

			rule.Name = strings.TrimSpace(ruleParts[0])
			paramValue := strings.TrimSpace(ruleParts[1])

			if paramValue == "" {
				return nil, fmt.Errorf("%w: %s", errRuleRequiresParam, rule.Name)
			}

			// Handle complex parameters (JSON arrays/objects) and enum values
			switch {
			case (strings.HasPrefix(paramValue, "[") && strings.HasSuffix(paramValue, "]")) ||
				(strings.HasPrefix(paramValue, "{") && strings.HasSuffix(paramValue, "}")):
				// Complex parameter (JSON array or object)
				rule.Params = []string{paramValue}
			case rule.Name == "enum" && strings.Contains(paramValue, " "):
				// Enum values separated by spaces
				rule.Params = strings.Fields(paramValue)
			case strings.Contains(paramValue, " ") && rule.Name != "regex":
				// Multiple space-separated parameters (but not for regex)
				rule.Params = strings.Fields(paramValue)
			default:
				// Single parameter
				rule.Params = []string{paramValue}
			}
		} else {
			// Simple rule without parameters
			rule.Name = strings.TrimSpace(part)
		}

		if rule.Name == "" {
			return nil, errEmptyRuleName
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// getTypeNameFromAST extracts the type name string from an AST expression.
func getTypeNameFromAST(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + getTypeNameFromAST(t.X)
	case *ast.ArrayType:
		return "[]" + getTypeNameFromAST(t.Elt)
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

// smartSplitTagRules splits tag rules by comma while respecting JSON arrays and objects.
func smartSplitTagRules(tagValue string) []string {
	var parts []string
	var current strings.Builder
	var inQuotes bool
	var inBraces, inBrackets int

	for i, char := range tagValue {
		switch char {
		case '"':
			if i == 0 || tagValue[i-1] != '\\' {
				inQuotes = !inQuotes
			}
			current.WriteRune(char)
		case '[':
			if !inQuotes {
				inBrackets++
			}
			current.WriteRune(char)
		case ']':
			if !inQuotes {
				inBrackets--
			}
			current.WriteRune(char)
		case '{':
			if !inQuotes {
				inBraces++
			}
			current.WriteRune(char)
		case '}':
			if !inQuotes {
				inBraces--
			}
			current.WriteRune(char)
		case ',':
			if !inQuotes && inBraces == 0 && inBrackets == 0 {
				// This is a rule separator
				parts = append(parts, current.String())
				current.Reset()
			} else {
				// This is inside a JSON value
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}

	// Add the final part
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}
