// Package main provides AST analysis functionality for the gozodgen tool.
// This module analyzes Go source files to identify structs that require
// code generation and extracts their field information using the existing
// pkg/tags infrastructure. It uses go/types for accurate type checking
// instead of the deprecated ast.Package.
package main

import (
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

// Static error variables to comply with err113
var (
	ErrInvalidRuleFormat = fmt.Errorf("invalid rule format")
	ErrRuleRequiresParam = fmt.Errorf("rule requires a parameter")
	ErrEmptyRuleName     = fmt.Errorf("empty rule name")
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

// GenerationInfo contains information about a struct that needs code generation
type GenerationInfo struct {
	Name        string                // Struct name
	Package     string                // Package name
	Fields      []tagparser.FieldInfo // Field information from tagparser
	Imports     []string              // Required imports
	HasGenerate bool                  // Whether struct has //go:generate gozodgen directive
	FilePath    string                // Source file path
	StructType  reflect.Type          // Struct type for validation
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

// AnalyzePackage analyzes all Go files in a package directory
func (a *StructAnalyzer) AnalyzePackage(pkgPath string) ([]*GenerationInfo, error) {
	// Parse all Go files in the package
	astPkgs, err := parser.ParseDir(a.fset, pkgPath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse package %s: %w", pkgPath, err)
	}

	var allStructs []*GenerationInfo

	// Process each package (usually just one)
	for pkgName, astPkg := range astPkgs {
		// Skip test packages
		if strings.HasSuffix(pkgName, "_test") {
			continue
		}

		// Collect all AST files for type checking
		var files []*ast.File
		for _, file := range astPkg.Files {
			files = append(files, file)
		}

		// Create type checker configuration
		conf := types.Config{
			Importer: importer.Default(),
		}

		// Type check the package
		typesPkg, err := conf.Check(pkgName, a.fset, files, a.info)
		if err != nil {
			// Continue even if there are type errors, as we might still be able to analyze structs
			fmt.Printf("Warning: type checking failed for package %s: %v\n", pkgName, err)
		}

		// Store the types package
		if typesPkg != nil {
			a.packages[pkgName] = typesPkg
		}

		// Analyze each file in the package
		for fileName, file := range astPkg.Files {
			structs, err := a.analyzeFile(fileName, file, pkgName)
			if err != nil {
				return nil, fmt.Errorf("failed to analyze file %s: %w", fileName, err)
			}
			allStructs = append(allStructs, structs...)
		}
	}

	return allStructs, nil
}

// analyzeFile analyzes a single Go file for structs requiring generation
func (a *StructAnalyzer) analyzeFile(fileName string, file *ast.File, pkgName string) ([]*GenerationInfo, error) {
	var structs []*GenerationInfo

	// Extract imports for later use
	imports := a.extractImports(file)

	// Look for struct declarations
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		// Check if this declaration has //go:generate gozodgen directive
		hasGenerate := a.hasGenerateDirective(genDecl.Doc)

		// Process each type specification
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			// Check if it's a struct type
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// Analyze the struct
			structInfo, err := a.analyzeStruct(typeSpec.Name.Name, structType, pkgName, fileName, imports, hasGenerate)
			if err != nil {
				return nil, fmt.Errorf("failed to analyze struct %s: %w", typeSpec.Name.Name, err)
			}

			// Only include structs that have gozod tags or generate directive
			if structInfo != nil && (len(structInfo.Fields) > 0 || hasGenerate) {
				structs = append(structs, structInfo)
			}
		}
	}

	return structs, nil
}

// analyzeStruct analyzes a single struct declaration
func (a *StructAnalyzer) analyzeStruct(name string, structType *ast.StructType, pkgName, fileName string, imports []string, hasGenerate bool) (*GenerationInfo, error) {
	// Parse struct fields directly from AST
	fields, err := a.parseStructFields(structType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse struct fields: %w", err)
	}

	// Filter fields that have gozod tags or should be included
	var gozodFields []tagparser.FieldInfo
	for _, field := range fields {
		// Include fields that have gozod tags (even if empty), have rules, are required, or struct has generate directive
		// We use the Optional field temporarily to store the hasAnyGozodTag flag
		hasGozodTag := field.Optional
		if len(field.Rules) > 0 || field.Required || hasGozodTag || hasGenerate {
			// Reset the Optional field to its proper value for pointer types
			field.Optional = field.Type != nil && field.Type.Kind() == reflect.Ptr && !field.Required
			gozodFields = append(gozodFields, field)
		}
	}

	return &GenerationInfo{
		Name:        name,
		Package:     pkgName,
		Fields:      gozodFields,
		Imports:     imports,
		HasGenerate: hasGenerate,
		FilePath:    fileName,
		StructType:  nil, // We'll set this when needed
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

// hasGenerateDirective checks if comments contain //go:generate gozodgen
func (a *StructAnalyzer) hasGenerateDirective(comments *ast.CommentGroup) bool {
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

// parseStructFields parses struct fields from AST and extracts tag information
func (a *StructAnalyzer) parseStructFields(structType *ast.StructType) ([]tagparser.FieldInfo, error) {
	var fields []tagparser.FieldInfo

	for _, field := range structType.Fields.List {
		// Skip anonymous fields for now
		if len(field.Names) == 0 {
			continue
		}

		for _, name := range field.Names {
			// Skip unexported fields
			if !name.IsExported() {
				continue
			}

			fieldInfo := tagparser.FieldInfo{
				Name:     name.Name,
				Type:     a.getReflectType(field.Type),
				TypeName: a.getTypeNameFromAST(field.Type), // Add AST type name
				JsonName: a.extractJSONName(field),
			}

			// Parse gozod tag if present
			hasAnyGozodTag := false
			if field.Tag != nil {
				tagValue := strings.Trim(field.Tag.Value, "`")
				// Check if gozod tag exists (even if empty)
				if strings.Contains(tagValue, "gozod:") {
					hasAnyGozodTag = true
					gozodTag := a.extractTagValue(tagValue, "gozod")
					fieldInfo.GozodTag = gozodTag // Store the raw gozod tag value (could be empty)
					if gozodTag != "" {
						rules, err := a.parseTagRules(gozodTag)
						if err != nil {
							return nil, fmt.Errorf("failed to parse gozod tag for field %s: %w", name.Name, err)
						}
						fieldInfo.Rules = rules

						// Set required flag
						for _, rule := range rules {
							if rule.Name == "required" {
								fieldInfo.Required = true
							}
						}
					}
				}
			}

			// Store whether field has any gozod tag for filtering purposes
			fieldInfo.Optional = hasAnyGozodTag // Reuse Optional field temporarily as a flag

			fields = append(fields, fieldInfo)
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

// typesToReflectType converts go/types.Type to reflect.Type
func (a *StructAnalyzer) typesToReflectType(t types.Type) reflect.Type {
	switch typ := t.(type) {
	case *types.Basic:
		switch typ.Kind() {
		case types.String:
			return reflect.TypeOf("")
		case types.Int:
			return reflect.TypeOf(0)
		case types.Int8:
			return reflect.TypeOf(int8(0))
		case types.Int16:
			return reflect.TypeOf(int16(0))
		case types.Int32:
			return reflect.TypeOf(int32(0))
		case types.Int64:
			return reflect.TypeOf(int64(0))
		case types.Uint:
			return reflect.TypeOf(uint(0))
		case types.Uint8:
			return reflect.TypeOf(uint8(0))
		case types.Uint16:
			return reflect.TypeOf(uint16(0))
		case types.Uint32:
			return reflect.TypeOf(uint32(0))
		case types.Uint64:
			return reflect.TypeOf(uint64(0))
		case types.Float32:
			return reflect.TypeOf(float32(0))
		case types.Float64:
			return reflect.TypeOf(float64(0))
		case types.Complex64:
			return reflect.TypeOf(complex64(0))
		case types.Complex128:
			return reflect.TypeOf(complex128(0))
		case types.Bool:
			return reflect.TypeOf(false)
		case types.Invalid, types.Uintptr, types.UnsafePointer:
			return reflect.TypeOf((*interface{})(nil)).Elem()
		case types.UntypedBool, types.UntypedInt, types.UntypedRune, types.UntypedFloat, types.UntypedComplex, types.UntypedString, types.UntypedNil:
			return reflect.TypeOf((*interface{})(nil)).Elem()
		default:
			return reflect.TypeOf((*interface{})(nil)).Elem()
		}
	case *types.Pointer:
		baseType := a.typesToReflectType(typ.Elem())
		return reflect.PointerTo(baseType)
	case *types.Slice:
		elemType := a.typesToReflectType(typ.Elem())
		return reflect.SliceOf(elemType)
	case *types.Array:
		elemType := a.typesToReflectType(typ.Elem())
		return reflect.SliceOf(elemType) // Treat arrays as slices for simplicity
	case *types.Map:
		keyType := a.typesToReflectType(typ.Key())
		valueType := a.typesToReflectType(typ.Elem())
		return reflect.MapOf(keyType, valueType)
	case *types.Named:
		// Handle named types like time.Time
		obj := typ.Obj()
		if obj != nil && obj.Pkg() != nil {
			pkgPath := obj.Pkg().Path()
			typeName := obj.Name()
			if pkgPath == "time" && typeName == "Time" {
				return reflect.TypeOf((*timeType)(nil)).Elem()
			}
		}
		// For other named types, try to get the underlying type
		return a.typesToReflectType(typ.Underlying())
	case *types.Interface:
		return reflect.TypeOf((*interface{})(nil)).Elem()
	default:
		return reflect.TypeOf((*interface{})(nil)).Elem()
	}
}

// getReflectTypeFromAST is the fallback AST-based type inference
func (a *StructAnalyzer) getReflectTypeFromAST(expr ast.Expr) reflect.Type {
	switch t := expr.(type) {
	case *ast.Ident:
		// Basic types
		switch t.Name {
		case "string":
			return reflect.TypeOf("")
		case "int":
			return reflect.TypeOf(0)
		case "int32":
			return reflect.TypeOf(int32(0))
		case "int64":
			return reflect.TypeOf(int64(0))
		case "float32":
			return reflect.TypeOf(float32(0))
		case "float64":
			return reflect.TypeOf(float64(0))
		case "bool":
			return reflect.TypeOf(false)
		default:
			// Unknown type - default to interface{}
			return reflect.TypeOf((*interface{})(nil)).Elem()
		}
	case *ast.StarExpr:
		// Pointer type
		baseType := a.getReflectTypeFromAST(t.X)
		return reflect.PointerTo(baseType)
	case *ast.ArrayType:
		// Slice or array type - treat both as slice for simplicity
		elemType := a.getReflectTypeFromAST(t.Elt)
		return reflect.SliceOf(elemType)
	case *ast.MapType:
		// Map type
		keyType := a.getReflectTypeFromAST(t.Key)
		valueType := a.getReflectTypeFromAST(t.Value)
		return reflect.MapOf(keyType, valueType)
	case *ast.SelectorExpr:
		// Qualified type (e.g., time.Time)
		if ident, ok := t.X.(*ast.Ident); ok {
			if ident.Name == "time" && t.Sel.Name == "Time" {
				// Return a special marker for time.Time
				return reflect.TypeOf((*timeType)(nil)).Elem()
			}
		}
		return reflect.TypeOf((*interface{})(nil)).Elem()
	default:
		// Unknown type - default to interface{}
		return reflect.TypeOf((*interface{})(nil)).Elem()
	}
}

// NeedsGeneration checks if a struct needs code generation
func (a *StructAnalyzer) NeedsGeneration(info *GenerationInfo) bool {
	// Check if struct has fields with validation rules
	for _, field := range info.Fields {
		if len(field.Rules) > 0 || field.Required {
			return true
		}
	}

	// Or if struct has //go:generate directive
	if info.HasGenerate {
		return true
	}

	return false
}

// extractJSONName extracts JSON field name from struct tag
func (a *StructAnalyzer) extractJSONName(field *ast.Field) string {
	if field.Tag == nil {
		if len(field.Names) > 0 {
			return field.Names[0].Name
		}
		return ""
	}

	tagValue := strings.Trim(field.Tag.Value, "`")
	jsonTag := a.extractTagValue(tagValue, "json")
	if jsonTag == "" {
		if len(field.Names) > 0 {
			return field.Names[0].Name
		}
		return ""
	}

	// Extract name before comma (ignore omitempty, etc.)
	parts := strings.Split(jsonTag, ",")
	jsonName := strings.TrimSpace(parts[0])
	if jsonName == "" || jsonName == "-" {
		if len(field.Names) > 0 {
			return field.Names[0].Name
		}
		return ""
	}

	return jsonName
}

// extractTagValue extracts the value for a specific tag name from a tag string
func (a *StructAnalyzer) extractTagValue(tagString, tagName string) string {
	// Parse the entire tag string
	tag := reflect.StructTag(tagString)
	return tag.Get(tagName)
}

// parseTagRules parses gozod tag rules with proper handling of complex JSON values
func (a *StructAnalyzer) parseTagRules(tagValue string) ([]tagparser.TagRule, error) {
	if tagValue == "" {
		return nil, nil
	}

	// Smart split that handles JSON arrays and objects
	parts := a.smartSplitTagRules(tagValue)
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
				return nil, fmt.Errorf("%w: %s", ErrInvalidRuleFormat, part)
			}

			rule.Name = strings.TrimSpace(ruleParts[0])
			paramValue := strings.TrimSpace(ruleParts[1])

			if paramValue == "" {
				return nil, fmt.Errorf("%w: %s", ErrRuleRequiresParam, rule.Name)
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
			return nil, ErrEmptyRuleName
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// getTypeNameFromAST extracts the type name string from AST expression for circular reference detection
func (a *StructAnalyzer) getTypeNameFromAST(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		// Pointer type - get the base type with * prefix
		baseTypeName := a.getTypeNameFromAST(t.X)
		return "*" + baseTypeName
	case *ast.ArrayType:
		// Slice or array type - treat both as slice for simplicity
		elemTypeName := a.getTypeNameFromAST(t.Elt)
		return "[]" + elemTypeName
	case *ast.MapType:
		// Map type
		keyTypeName := a.getTypeNameFromAST(t.Key)
		valueTypeName := a.getTypeNameFromAST(t.Value)
		return "map[" + keyTypeName + "]" + valueTypeName
	case *ast.SelectorExpr:
		// Qualified type (e.g., time.Time)
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name + "." + t.Sel.Name
		}
		return t.Sel.Name
	default:
		return "unknown"
	}
}

// smartSplitTagRules splits tag rules by comma while respecting JSON arrays and objects
func (a *StructAnalyzer) smartSplitTagRules(tagValue string) []string {
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
